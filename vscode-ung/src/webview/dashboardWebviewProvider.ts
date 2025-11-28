import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';

/**
 * Dashboard metrics interface
 */
interface DashboardMetrics {
  totalMonthlyRevenue: number;
  hourlyRevenue: number;
  retainerRevenue: number;
  projectedHours: number;
  averageHourlyRate: number;
  activeContracts: number;
  totalClients: number;
  pendingInvoices: number;
  unpaidAmount: number;
  currency: string;
}

/**
 * Counts for navigation badges
 */
interface EntityCounts {
  invoices: number;
  clients: number;
  contracts: number;
  expenses: number;
}

/**
 * Recent contract info
 */
interface RecentContract {
  id: number;
  name: string;
  client: string;
  type: string;
  rate: string;
}

/**
 * Recent invoice info
 */
interface RecentInvoice {
  id: number;
  invoiceNum: string;
  client: string;
  amount: string;
  status: string;
}

/**
 * Dashboard webview provider for the sidebar
 * Shows a professional dashboard with metrics, quick actions, and navigation
 */
export class DashboardWebviewProvider implements vscode.WebviewViewProvider {
  public static readonly viewType = 'ungDashboard';

  private _view?: vscode.WebviewView;
  private _extensionUri: vscode.Uri;
  private _metrics: DashboardMetrics | null = null;
  private _counts: EntityCounts = {
    invoices: 0,
    clients: 0,
    contracts: 0,
    expenses: 0,
  };
  private _todayHours: number = 0;
  private _isLoading: boolean = true;
  private _activeTracking: {
    project: string;
    client: string;
    duration: string;
    startTime: number; // Unix timestamp in seconds for live timer
  } | null = null;
  private _recentContracts: RecentContract[] = [];
  private _recentInvoices: RecentInvoice[] = [];
  private _setupStatus: {
    hasCompany: boolean;
    hasClient: boolean;
    hasContract: boolean;
  } = {
    hasCompany: false,
    hasClient: false,
    hasContract: false,
  };

  constructor(
    extensionUri: vscode.Uri,
    private readonly cli: UngCli
  ) {
    this._extensionUri = extensionUri;
  }

  public resolveWebviewView(
    webviewView: vscode.WebviewView,
    _context: vscode.WebviewViewResolveContext,
    _token: vscode.CancellationToken
  ) {
    this._view = webviewView;

    webviewView.webview.options = {
      enableScripts: true,
      localResourceRoots: [this._extensionUri],
    };

    // Set initial HTML with loading state
    this._isLoading = true;
    webviewView.webview.html = this._getHtmlForWebview();

    // Handle messages from the webview
    webviewView.webview.onDidReceiveMessage(async (message) => {
      switch (message.command) {
        case 'startTracking':
          vscode.commands.executeCommand('ung.startTracking');
          break;
        case 'stopTracking':
          vscode.commands.executeCommand('ung.stopTracking');
          break;
        case 'createInvoice':
          vscode.commands.executeCommand('ung.createInvoice');
          break;
        case 'logExpense':
          vscode.commands.executeCommand('ung.logExpense');
          break;
        case 'createClient':
          vscode.commands.executeCommand('ung.createClient');
          break;
        case 'createContract':
          vscode.commands.executeCommand('ung.createContract');
          break;
        case 'createCompany':
          vscode.commands.executeCommand('ung.createCompany');
          break;
        case 'openInvoices':
          vscode.commands.executeCommand('ung.searchInvoices');
          break;
        case 'openClients':
          vscode.commands.executeCommand('ung.searchClients');
          break;
        case 'openContracts':
          vscode.commands.executeCommand('ung.searchContracts');
          break;
        case 'openExpenses':
          vscode.commands.executeCommand('ung.viewExpenseReport');
          break;
        case 'openTracking':
          vscode.commands.executeCommand('ung.viewActiveSession');
          break;
        case 'openStatistics':
          vscode.commands.executeCommand('ung.openStatistics');
          break;
        case 'openTemplateEditor':
          vscode.commands.executeCommand('ung.openTemplateEditor');
          break;
        case 'openPomodoro':
          vscode.commands.executeCommand('ung.startPomodoro');
          break;
        case 'refresh':
          await this.refresh();
          break;
      }
    });

    // Load data
    this.refresh();
  }

  public async refresh(): Promise<void> {
    if (!this._view) {
      return;
    }

    this._isLoading = true;
    this._view.webview.html = this._getHtmlForWebview();

    // Load all data in parallel for speed
    await Promise.all([
      this._loadMetrics(),
      this._checkActiveTracking(),
      this._checkSetupStatus(),
      this._loadCounts(),
      this._loadTodayHours(),
      this._loadRecentContracts(),
      this._loadRecentInvoices(),
    ]);

    this._isLoading = false;
    this._view.webview.html = this._getHtmlForWebview();
  }

  private async _loadMetrics(): Promise<void> {
    try {
      const result = await this.cli.exec(['dashboard']);
      if (result.success && result.stdout) {
        this._metrics = this._parseDashboardOutput(result.stdout);
      }
    } catch {
      this._metrics = null;
    }
  }

  private async _loadCounts(): Promise<void> {
    try {
      const [invoiceResult, clientResult, contractResult, expenseResult] =
        await Promise.all([
          this.cli.exec(['invoice', 'list']),
          this.cli.exec(['client', 'list']),
          this.cli.exec(['contract', 'list']),
          this.cli.exec(['expense', 'list']),
        ]);

      this._counts = {
        invoices: this._countLines(invoiceResult.stdout),
        clients: this._countLines(clientResult.stdout),
        contracts: this._countLines(contractResult.stdout),
        expenses: this._countLines(expenseResult.stdout),
      };
    } catch {
      this._counts = { invoices: 0, clients: 0, contracts: 0, expenses: 0 };
    }
  }

  private _countLines(output: string | undefined): number {
    if (!output) return 0;
    const lines = output
      .split('\n')
      .filter((l) => l.trim() && !l.includes('─') && !l.includes('ID'));
    return Math.max(0, lines.length - 1); // Subtract header
  }

  private async _loadTodayHours(): Promise<void> {
    try {
      const result = await this.cli.exec(['track', 'today']);
      if (result.success && result.stdout) {
        const match = result.stdout.match(/(\d+\.?\d*)\s*h/i);
        if (match) {
          this._todayHours = parseFloat(match[1]);
        } else {
          // Try to parse duration format like "2:30"
          const durationMatch = result.stdout.match(/(\d+):(\d+)/);
          if (durationMatch) {
            this._todayHours =
              parseInt(durationMatch[1], 10) +
              parseInt(durationMatch[2], 10) / 60;
          }
        }
      }
    } catch {
      this._todayHours = 0;
    }
  }

  private async _checkActiveTracking(): Promise<void> {
    try {
      const result = await this.cli.exec(['track', 'now']);
      if (
        result.success &&
        result.stdout &&
        !result.stdout.includes('No active')
      ) {
        const lines = result.stdout.split('\n');
        let project = 'Unknown';
        let client = '';
        let duration = '0:00';
        let startTime = Math.floor(Date.now() / 1000); // Default to now

        for (const line of lines) {
          if (line.includes('Project:')) {
            project = line.split(':')[1]?.trim() || 'Unknown';
          } else if (line.includes('Client:')) {
            client = line.split(':')[1]?.trim() || '';
          } else if (line.includes('Duration:') || line.includes('Elapsed:')) {
            duration = line.split(':').slice(1).join(':').trim() || '0:00';
            // Calculate start time from duration
            const durationParts = duration.match(/(\d+):(\d+)(?::(\d+))?/);
            if (durationParts) {
              const hours = parseInt(durationParts[1], 10) || 0;
              const minutes = parseInt(durationParts[2], 10) || 0;
              const seconds = parseInt(durationParts[3], 10) || 0;
              const elapsedSeconds = hours * 3600 + minutes * 60 + seconds;
              startTime = Math.floor(Date.now() / 1000) - elapsedSeconds;
            }
          } else if (line.includes('Started:') || line.includes('Start:')) {
            // Try to parse start time if provided
            const timeStr = line.split(':').slice(1).join(':').trim();
            const parsed = Date.parse(timeStr);
            if (!Number.isNaN(parsed)) {
              startTime = Math.floor(parsed / 1000);
            }
          }
        }

        this._activeTracking = { project, client, duration, startTime };
      } else {
        this._activeTracking = null;
      }
    } catch {
      this._activeTracking = null;
    }
  }

  private async _checkSetupStatus(): Promise<void> {
    try {
      const [companyResult, clientResult, contractResult] = await Promise.all([
        this.cli.exec(['company', 'list']),
        this.cli.exec(['client', 'list']),
        this.cli.exec(['contract', 'list']),
      ]);

      this._setupStatus = {
        hasCompany: !!(
          companyResult.success &&
          companyResult.stdout &&
          !companyResult.stdout.includes('No company')
        ),
        hasClient: !!(
          clientResult.success &&
          clientResult.stdout &&
          clientResult.stdout.split('\n').length > 2
        ),
        hasContract: !!(
          contractResult.success &&
          contractResult.stdout &&
          contractResult.stdout.split('\n').length > 2
        ),
      };
    } catch {
      this._setupStatus = {
        hasCompany: false,
        hasClient: false,
        hasContract: false,
      };
    }
  }

  private async _loadRecentContracts(): Promise<void> {
    try {
      const result = await this.cli.exec(['contract', 'list']);
      if (result.success && result.stdout) {
        this._recentContracts = this._parseContractsFromOutput(result.stdout);
      } else {
        this._recentContracts = [];
      }
    } catch {
      this._recentContracts = [];
    }
  }

  private _parseContractsFromOutput(output: string): RecentContract[] {
    const lines = output
      .split('\n')
      .filter((l) => l.trim() && !l.includes('─') && !l.includes('ID'));
    if (lines.length < 2) return [];

    const contracts: RecentContract[] = [];
    // Skip header line
    for (let i = 1; i < lines.length && contracts.length < 3; i++) {
      const parts = lines[i]
        .split(/\s{2,}/)
        .map((p) => p.trim())
        .filter((p) => p);
      if (parts.length >= 6) {
        const id = parseInt(parts[0], 10);
        if (!Number.isNaN(id)) {
          contracts.push({
            id,
            name: parts[2] || 'Unnamed',
            client: parts[3] || 'Unknown',
            type: parts[4] || 'hourly',
            rate: parts[5] || '-',
          });
        }
      }
    }
    return contracts;
  }

  private async _loadRecentInvoices(): Promise<void> {
    try {
      const result = await this.cli.exec(['invoice', 'list']);
      if (result.success && result.stdout) {
        this._recentInvoices = this._parseInvoicesFromOutput(result.stdout);
      } else {
        this._recentInvoices = [];
      }
    } catch {
      this._recentInvoices = [];
    }
  }

  private _parseInvoicesFromOutput(output: string): RecentInvoice[] {
    const lines = output
      .split('\n')
      .filter((l) => l.trim() && !l.includes('─') && !l.includes('ID'));
    if (lines.length < 2) return [];

    const invoices: RecentInvoice[] = [];
    // Skip header line
    for (let i = 1; i < lines.length && invoices.length < 3; i++) {
      const parts = lines[i]
        .split(/\s{2,}/)
        .map((p) => p.trim())
        .filter((p) => p);
      if (parts.length >= 5) {
        const id = parseInt(parts[0], 10);
        if (!Number.isNaN(id)) {
          invoices.push({
            id,
            invoiceNum: parts[1] || `INV-${id}`,
            client: parts[2] || 'Unknown',
            amount: parts[3] || '$0',
            status: parts[4] || 'draft',
          });
        }
      }
    }
    return invoices;
  }

  private _parseDashboardOutput(output: string): DashboardMetrics {
    const metrics: DashboardMetrics = {
      totalMonthlyRevenue: 0,
      hourlyRevenue: 0,
      retainerRevenue: 0,
      projectedHours: 0,
      averageHourlyRate: 0,
      activeContracts: 0,
      totalClients: 0,
      pendingInvoices: 0,
      unpaidAmount: 0,
      currency: 'USD',
    };

    const lines = output.split('\n');

    for (const line of lines) {
      if (line.includes('Total Monthly Revenue') || line.includes('TOTAL')) {
        const match = line.match(/\$?([0-9,]+\.?[0-9]*)/);
        if (match) {
          metrics.totalMonthlyRevenue = parseFloat(match[1].replace(/,/g, ''));
        }
      }
      if (line.includes('Hourly Contracts')) {
        const match = line.match(/\$?([0-9,]+\.?[0-9]*)/);
        if (match) {
          metrics.hourlyRevenue = parseFloat(match[1].replace(/,/g, ''));
        }
      }
      if (line.includes('Retainer Contracts')) {
        const match = line.match(/\$?([0-9,]+\.?[0-9]*)/);
        if (match) {
          metrics.retainerRevenue = parseFloat(match[1].replace(/,/g, ''));
        }
      }
      if (line.includes('Projected Hours')) {
        const match = line.match(/([0-9,]+\.?[0-9]*)/);
        if (match) {
          metrics.projectedHours = parseFloat(match[1].replace(/,/g, ''));
        }
      }
      if (line.includes('Average Rate')) {
        const match = line.match(/\$?([0-9,]+\.?[0-9]*)/);
        if (match) {
          metrics.averageHourlyRate = parseFloat(match[1].replace(/,/g, ''));
        }
      }
      if (line.includes('Total Clients') || line.includes('Clients:')) {
        const match = line.match(/:\s*(\d+)/);
        if (match) {
          metrics.totalClients = parseInt(match[1], 10);
        }
      }
      if (line.includes('Active Contracts') || line.includes('Contracts:')) {
        const match = line.match(/:\s*(\d+)/);
        if (match) {
          metrics.activeContracts = parseInt(match[1], 10);
        }
      }
      if (line.includes('Pending Invoices')) {
        const match = line.match(/:\s*(\d+)/);
        if (match) {
          metrics.pendingInvoices = parseInt(match[1], 10);
        }
      }
      if (line.includes('Unpaid Amount')) {
        const match = line.match(/\$?([0-9,]+\.?[0-9]*)/);
        if (match) {
          metrics.unpaidAmount = parseFloat(match[1].replace(/,/g, ''));
        }
      }
    }

    return metrics;
  }

  private _formatCurrency(amount: number, currency: string = 'USD'): string {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency,
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(amount);
  }

  private _getHtmlForWebview(): string {
    const needsSetup =
      !this._setupStatus.hasCompany ||
      !this._setupStatus.hasClient ||
      !this._setupStatus.hasContract;

    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>UNG Dashboard</title>
    <style>
        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            font-family: var(--vscode-font-family);
            color: var(--vscode-foreground);
            background-color: var(--vscode-sideBar-background);
            padding: 16px;
            line-height: 1.5;
        }

        .container {
            max-width: 100%;
        }

        /* Loading State */
        .loading-overlay {
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            padding: 40px 20px;
        }

        .spinner {
            width: 32px;
            height: 32px;
            border: 3px solid var(--vscode-input-background);
            border-top: 3px solid var(--vscode-textLink-foreground);
            border-radius: 50%;
            animation: spin 1s linear infinite;
        }

        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }

        .loading-text {
            margin-top: 12px;
            color: var(--vscode-descriptionForeground);
            font-size: 12px;
        }

        /* Header */
        .header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 16px;
            padding-bottom: 12px;
            border-bottom: 1px solid var(--vscode-panel-border);
        }

        .header-left {
            display: flex;
            align-items: center;
            gap: 10px;
        }

        .logo {
            width: 32px;
            height: 32px;
            background: linear-gradient(135deg, var(--vscode-textLink-foreground), var(--vscode-textLink-activeForeground, var(--vscode-textLink-foreground)));
            border-radius: 8px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 14px;
            font-weight: 600;
            color: white;
        }

        .header-title {
            font-size: 15px;
            font-weight: 600;
        }

        .refresh-btn {
            background: none;
            border: none;
            color: var(--vscode-textLink-foreground);
            cursor: pointer;
            padding: 6px;
            border-radius: 4px;
            font-size: 14px;
            transition: background-color 0.15s;
        }

        .refresh-btn:hover {
            background-color: var(--vscode-list-hoverBackground);
        }

        .refresh-btn.spinning {
            animation: spin 1s linear infinite;
        }

        /* Alert Banner */
        .alert-banner {
            display: flex;
            align-items: center;
            gap: 10px;
            padding: 10px 12px;
            background-color: color-mix(in srgb, var(--vscode-charts-red, #f44336) 15%, var(--vscode-sideBar-background));
            border: 1px solid var(--vscode-charts-red, #f44336);
            border-radius: 8px;
            margin-bottom: 14px;
            font-size: 12px;
            cursor: pointer;
            transition: background-color 0.15s;
        }

        .alert-banner:hover {
            background-color: color-mix(in srgb, var(--vscode-charts-red, #f44336) 25%, var(--vscode-sideBar-background));
        }

        .alert-icon {
            font-size: 16px;
        }

        .alert-text {
            flex: 1;
        }

        .alert-value {
            font-weight: 600;
            color: var(--vscode-charts-red, #f44336);
        }

        /* Active Tracking Banner */
        .tracking-banner {
            display: flex;
            align-items: center;
            gap: 10px;
            padding: 10px 12px;
            background: linear-gradient(135deg, color-mix(in srgb, var(--vscode-charts-green, #4caf50) 15%, var(--vscode-sideBar-background)), var(--vscode-sideBar-background));
            border: 1px solid var(--vscode-charts-green, #4caf50);
            border-radius: 8px;
            margin-bottom: 14px;
        }

        .tracking-indicator {
            width: 8px;
            height: 8px;
            background-color: var(--vscode-charts-green, #4caf50);
            border-radius: 50%;
            animation: pulse 2s infinite;
        }

        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.4; }
        }

        .tracking-info {
            flex: 1;
            min-width: 0;
        }

        .tracking-project {
            font-weight: 600;
            font-size: 12px;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .tracking-meta {
            font-size: 10px;
            color: var(--vscode-descriptionForeground);
        }

        .tracking-duration {
            font-size: 16px;
            font-weight: 600;
            font-family: monospace;
            color: var(--vscode-charts-green, #4caf50);
        }

        .stop-btn {
            background-color: var(--vscode-inputValidation-errorBackground, #5a1d1d);
            border: 1px solid var(--vscode-inputValidation-errorBorder, #be1100);
            color: var(--vscode-errorForeground, #f48771);
            padding: 5px 10px;
            border-radius: 5px;
            cursor: pointer;
            font-size: 11px;
            font-weight: 500;
            transition: all 0.15s;
        }

        .stop-btn:hover {
            background-color: var(--vscode-inputValidation-errorBorder, #be1100);
            color: white;
        }

        /* Metrics Grid */
        .metrics-grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 8px;
            margin-bottom: 16px;
        }

        .metric-card {
            background-color: var(--vscode-input-background);
            border-radius: 8px;
            padding: 12px;
            cursor: pointer;
            transition: all 0.15s;
            border: 1px solid transparent;
        }

        .metric-card:hover {
            background-color: var(--vscode-list-hoverBackground);
            border-color: var(--vscode-focusBorder);
        }

        .metric-header {
            display: flex;
            align-items: center;
            gap: 6px;
            margin-bottom: 4px;
        }

        .metric-icon {
            font-size: 14px;
        }

        .metric-label {
            font-size: 10px;
            color: var(--vscode-descriptionForeground);
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }

        .metric-value {
            font-size: 18px;
            font-weight: 600;
        }

        .metric-value.green { color: var(--vscode-charts-green, #4caf50); }
        .metric-value.blue { color: var(--vscode-charts-blue, #2196f3); }
        .metric-value.orange { color: var(--vscode-charts-orange, #ff9800); }
        .metric-value.red { color: var(--vscode-charts-red, #f44336); }
        .metric-value.purple { color: var(--vscode-charts-purple, #9c27b0); }

        /* Section */
        .section {
            margin-bottom: 16px;
        }

        .section-header {
            display: flex;
            align-items: center;
            gap: 6px;
            margin-bottom: 10px;
        }

        .section-title {
            font-size: 10px;
            font-weight: 600;
            color: var(--vscode-descriptionForeground);
            text-transform: uppercase;
            letter-spacing: 0.8px;
        }

        /* Quick Actions */
        .quick-actions {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 6px;
        }

        .action-btn {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 10px;
            background-color: var(--vscode-input-background);
            border: 1px solid transparent;
            border-radius: 6px;
            cursor: pointer;
            font-size: 11px;
            font-family: var(--vscode-font-family);
            color: var(--vscode-foreground);
            transition: all 0.15s;
            text-align: left;
        }

        .action-btn:hover {
            background-color: var(--vscode-list-hoverBackground);
            border-color: var(--vscode-focusBorder);
        }

        .action-btn.primary {
            background-color: var(--vscode-button-background);
            color: var(--vscode-button-foreground);
            grid-column: span 2;
        }

        .action-btn.primary:hover {
            background-color: var(--vscode-button-hoverBackground);
        }

        .action-icon {
            font-size: 14px;
            width: 20px;
            text-align: center;
        }

        .action-label {
            font-weight: 500;
        }

        /* Navigation */
        .nav-grid {
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 6px;
        }

        .nav-item {
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 4px;
            padding: 12px 6px;
            background-color: var(--vscode-input-background);
            border: 1px solid transparent;
            border-radius: 8px;
            cursor: pointer;
            font-size: 10px;
            font-family: var(--vscode-font-family);
            color: var(--vscode-foreground);
            transition: all 0.15s;
            position: relative;
        }

        .nav-item:hover {
            background-color: var(--vscode-list-hoverBackground);
            border-color: var(--vscode-focusBorder);
            transform: translateY(-1px);
        }

        .nav-icon {
            font-size: 18px;
        }

        .nav-label {
            font-weight: 500;
        }

        .nav-badge {
            position: absolute;
            top: 6px;
            right: 6px;
            min-width: 16px;
            height: 16px;
            padding: 0 4px;
            background-color: var(--vscode-badge-background);
            color: var(--vscode-badge-foreground);
            border-radius: 8px;
            font-size: 9px;
            font-weight: 600;
            display: flex;
            align-items: center;
            justify-content: center;
        }

        /* Setup Section */
        .setup-section {
            background-color: color-mix(in srgb, var(--vscode-textLink-foreground) 8%, var(--vscode-input-background));
            border: 1px solid color-mix(in srgb, var(--vscode-textLink-foreground) 25%, transparent);
            border-radius: 8px;
            padding: 12px;
            margin-bottom: 16px;
        }

        .setup-title {
            font-size: 12px;
            font-weight: 600;
            margin-bottom: 10px;
            display: flex;
            align-items: center;
            gap: 6px;
        }

        .setup-steps {
            display: flex;
            flex-direction: column;
            gap: 6px;
        }

        .setup-step {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 8px 10px;
            background-color: var(--vscode-sideBar-background);
            border-radius: 6px;
            cursor: pointer;
            transition: all 0.15s;
        }

        .setup-step:hover {
            background-color: var(--vscode-list-hoverBackground);
        }

        .setup-step.completed {
            opacity: 0.5;
        }

        .step-icon {
            width: 20px;
            height: 20px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 10px;
            flex-shrink: 0;
        }

        .step-icon.pending {
            background-color: var(--vscode-input-background);
            border: 2px solid var(--vscode-textLink-foreground);
            color: var(--vscode-textLink-foreground);
        }

        .step-icon.done {
            background-color: var(--vscode-charts-green, #4caf50);
            color: white;
        }

        .step-content {
            flex: 1;
        }

        .step-label {
            font-size: 11px;
            font-weight: 500;
        }

        .step-desc {
            font-size: 10px;
            color: var(--vscode-descriptionForeground);
        }

        /* Divider */
        .divider {
            height: 1px;
            background: linear-gradient(to right, transparent, var(--vscode-panel-border), transparent);
            margin: 14px 0;
        }

        /* Tools Grid - improved styling */
        .tools-grid {
            display: flex;
            gap: 8px;
        }

        .tool-item {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 6px;
            padding: 10px 14px;
            background: linear-gradient(135deg, var(--vscode-input-background) 0%, color-mix(in srgb, var(--vscode-input-background) 90%, var(--vscode-textLink-foreground)) 100%);
            border-radius: 8px;
            cursor: pointer;
            font-size: 11px;
            font-weight: 500;
            transition: all 0.2s;
            border: 1px solid color-mix(in srgb, var(--vscode-panel-border) 50%, transparent);
            flex: 1;
        }

        .tool-item:hover {
            background: linear-gradient(135deg, var(--vscode-list-hoverBackground) 0%, color-mix(in srgb, var(--vscode-list-hoverBackground) 80%, var(--vscode-textLink-foreground)) 100%);
            border-color: var(--vscode-focusBorder);
            transform: translateY(-1px);
            box-shadow: 0 2px 8px rgba(0,0,0,0.15);
        }

        .tool-icon {
            font-size: 14px;
        }

        /* Recent Items List */
        .recent-list {
            display: flex;
            flex-direction: column;
            gap: 4px;
        }

        .recent-item {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 8px 10px;
            background-color: var(--vscode-input-background);
            border-radius: 6px;
            cursor: pointer;
            transition: all 0.15s;
            border: 1px solid transparent;
        }

        .recent-item:hover {
            background-color: var(--vscode-list-hoverBackground);
            border-color: var(--vscode-focusBorder);
        }

        .recent-item-icon {
            font-size: 12px;
            width: 20px;
            text-align: center;
        }

        .recent-item-content {
            flex: 1;
            min-width: 0;
        }

        .recent-item-title {
            font-size: 11px;
            font-weight: 500;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .recent-item-subtitle {
            font-size: 10px;
            color: var(--vscode-descriptionForeground);
        }

        .recent-item-badge {
            font-size: 9px;
            padding: 2px 6px;
            border-radius: 4px;
            font-weight: 500;
        }

        .recent-item-badge.paid {
            background-color: color-mix(in srgb, var(--vscode-charts-green, #4caf50) 20%, transparent);
            color: var(--vscode-charts-green, #4caf50);
        }

        .recent-item-badge.pending, .recent-item-badge.sent {
            background-color: color-mix(in srgb, var(--vscode-charts-orange, #ff9800) 20%, transparent);
            color: var(--vscode-charts-orange, #ff9800);
        }

        .recent-item-badge.draft {
            background-color: color-mix(in srgb, var(--vscode-descriptionForeground) 20%, transparent);
            color: var(--vscode-descriptionForeground);
        }

        .recent-item-badge.overdue {
            background-color: color-mix(in srgb, var(--vscode-charts-red, #f44336) 20%, transparent);
            color: var(--vscode-charts-red, #f44336);
        }

        /* Empty State */
        .empty-state {
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 8px;
            padding: 16px;
            background-color: color-mix(in srgb, var(--vscode-input-background) 50%, transparent);
            border: 1px dashed var(--vscode-panel-border);
            border-radius: 8px;
            text-align: center;
        }

        .empty-state-icon {
            font-size: 24px;
            opacity: 0.6;
        }

        .empty-state-text {
            font-size: 11px;
            color: var(--vscode-descriptionForeground);
        }

        .empty-state-action {
            font-size: 11px;
            color: var(--vscode-textLink-foreground);
            cursor: pointer;
            padding: 4px 10px;
            background-color: var(--vscode-button-background);
            color: var(--vscode-button-foreground);
            border-radius: 4px;
            border: none;
            font-weight: 500;
            margin-top: 4px;
        }

        .empty-state-action:hover {
            background-color: var(--vscode-button-hoverBackground);
        }

        /* Section with view all link */
        .section-header-with-link {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 10px;
        }

        .section-link {
            font-size: 10px;
            color: var(--vscode-textLink-foreground);
            cursor: pointer;
            border: none;
            background: none;
            padding: 2px 6px;
            border-radius: 4px;
        }

        .section-link:hover {
            background-color: var(--vscode-list-hoverBackground);
        }
    </style>
</head>
<body>
    <div class="container">
        ${this._isLoading ? this._getLoadingHtml() : this._getContentHtml(needsSetup)}
    </div>

    <script>
        (function() {
            const vscode = acquireVsCodeApi();

            // Use event delegation - attach to document body
            // This handles all current and future elements with data-command
            document.body.addEventListener('click', function(e) {
                const target = e.target.closest('[data-command]');
                if (target) {
                    e.preventDefault();
                    e.stopPropagation();
                    const command = target.getAttribute('data-command');
                    if (command) {
                        vscode.postMessage({ command: command });
                    }
                }
            });

            // Live timer for active tracking
            const timerEl = document.getElementById('live-timer');
            const startTimeAttr = timerEl ? timerEl.getAttribute('data-start-time') : null;

            if (timerEl && startTimeAttr) {
                const startTime = parseInt(startTimeAttr, 10);

                function updateTimer() {
                    const now = Math.floor(Date.now() / 1000);
                    const elapsed = now - startTime;

                    const hours = Math.floor(elapsed / 3600);
                    const minutes = Math.floor((elapsed % 3600) / 60);
                    const seconds = elapsed % 60;

                    timerEl.textContent = hours.toString().padStart(1, '0') + ':' +
                                         minutes.toString().padStart(2, '0') + ':' +
                                         seconds.toString().padStart(2, '0');
                }

                // Update immediately and then every second
                updateTimer();
                setInterval(updateTimer, 1000);
            }
        })();
    </script>
</body>
</html>`;
  }

  private _getLoadingHtml(): string {
    return `
        <div class="header">
            <div class="header-left">
                <div class="logo">U</div>
                <span class="header-title">Dashboard</span>
            </div>
        </div>
        <div class="loading-overlay">
            <div class="spinner"></div>
            <div class="loading-text">Loading dashboard...</div>
        </div>
    `;
  }

  private _getContentHtml(needsSetup: boolean): string {
    const m = this._metrics;
    const hasUnpaid = m && m.unpaidAmount > 0;

    return `
        <!-- Header -->
        <div class="header">
            <div class="header-left">
                <div class="logo">U</div>
                <span class="header-title">Dashboard</span>
            </div>
            <button class="refresh-btn" data-command="refresh" title="Refresh">
                🔄
            </button>
        </div>

        ${this._activeTracking ? this._getActiveTrackingHtml() : ''}

        ${hasUnpaid ? this._getAlertBannerHtml() : ''}

        ${needsSetup ? this._getSetupSectionHtml() : ''}

        <!-- Metrics -->
        <div class="section">
            <div class="section-header">
                <span class="section-title">Overview</span>
            </div>
            <div class="metrics-grid">
                ${this._getMetricsHtml()}
            </div>
        </div>

        <!-- Quick Actions -->
        <div class="section">
            <div class="section-header">
                <span class="section-title">Quick Actions</span>
            </div>
            <div class="quick-actions">
                ${this._getQuickActionsHtml()}
            </div>
        </div>

        <div class="divider"></div>

        <!-- Recent Contracts -->
        <div class="section">
            <div class="section-header-with-link">
                <span class="section-title">Contracts</span>
                ${this._recentContracts.length > 0 ? '<button class="section-link" data-command="openContracts">View All</button>' : ''}
            </div>
            ${this._getRecentContractsHtml()}
        </div>

        <!-- Recent Invoices -->
        <div class="section">
            <div class="section-header-with-link">
                <span class="section-title">Invoices</span>
                ${this._recentInvoices.length > 0 ? '<button class="section-link" data-command="openInvoices">View All</button>' : ''}
            </div>
            ${this._getRecentInvoicesHtml()}
        </div>

        <div class="divider"></div>

        <!-- Navigation -->
        <div class="section">
            <div class="section-header">
                <span class="section-title">Navigate</span>
            </div>
            <div class="nav-grid">
                <button class="nav-item" data-command="openInvoices">
                    <span class="nav-icon">📄</span>
                    <span class="nav-label">Invoices</span>
                    ${this._counts.invoices > 0 ? `<span class="nav-badge">${this._counts.invoices}</span>` : ''}
                </button>
                <button class="nav-item" data-command="openClients">
                    <span class="nav-icon">👥</span>
                    <span class="nav-label">Clients</span>
                    ${this._counts.clients > 0 ? `<span class="nav-badge">${this._counts.clients}</span>` : ''}
                </button>
                <button class="nav-item" data-command="openContracts">
                    <span class="nav-icon">📝</span>
                    <span class="nav-label">Contracts</span>
                    ${this._counts.contracts > 0 ? `<span class="nav-badge">${this._counts.contracts}</span>` : ''}
                </button>
                <button class="nav-item" data-command="openExpenses">
                    <span class="nav-icon">💳</span>
                    <span class="nav-label">Expenses</span>
                    ${this._counts.expenses > 0 ? `<span class="nav-badge">${this._counts.expenses}</span>` : ''}
                </button>
                <button class="nav-item" data-command="openTracking">
                    <span class="nav-icon">⏱️</span>
                    <span class="nav-label">Time</span>
                </button>
                <button class="nav-item" data-command="openStatistics">
                    <span class="nav-icon">📊</span>
                    <span class="nav-label">Reports</span>
                </button>
            </div>
        </div>

        <div class="divider"></div>

        <!-- Tools -->
        <div class="section">
            <div class="section-header">
                <span class="section-title">Tools</span>
            </div>
            <div class="tools-grid">
                <button class="tool-item" data-command="openTemplateEditor">
                    <span class="tool-icon">🎨</span>
                    <span>Templates</span>
                </button>
                <button class="tool-item" data-command="openPomodoro">
                    <span class="tool-icon">🍅</span>
                    <span>Pomodoro</span>
                </button>
            </div>
        </div>
    `;
  }

  private _getAlertBannerHtml(): string {
    const m = this._metrics;
    if (!m || m.unpaidAmount <= 0) return '';

    return `
        <div class="alert-banner" data-command="openInvoices">
            <span class="alert-icon">⚠️</span>
            <span class="alert-text">
                <span class="alert-value">${this._formatCurrency(m.unpaidAmount, m.currency)}</span> unpaid
            </span>
        </div>
    `;
  }

  private _getActiveTrackingHtml(): string {
    if (!this._activeTracking) return '';

    return `
        <div class="tracking-banner">
            <div class="tracking-indicator"></div>
            <div class="tracking-info">
                <div class="tracking-project">${this._activeTracking.project}</div>
                ${this._activeTracking.client ? `<div class="tracking-meta">${this._activeTracking.client}</div>` : ''}
            </div>
            <div class="tracking-duration" id="live-timer" data-start-time="${this._activeTracking.startTime}">${this._activeTracking.duration}</div>
            <button class="stop-btn" data-command="stopTracking">Stop</button>
        </div>
    `;
  }

  private _getSetupSectionHtml(): string {
    return `
        <div class="setup-section">
            <div class="setup-title">
                <span>🚀</span>
                <span>Get Started</span>
            </div>
            <div class="setup-steps">
                <div class="setup-step ${this._setupStatus.hasCompany ? 'completed' : ''}" data-command="createCompany">
                    <div class="step-icon ${this._setupStatus.hasCompany ? 'done' : 'pending'}">
                        ${this._setupStatus.hasCompany ? '✓' : '1'}
                    </div>
                    <div class="step-content">
                        <div class="step-label">${this._setupStatus.hasCompany ? 'Company Added' : 'Add Company'}</div>
                        <div class="step-desc">Your business details</div>
                    </div>
                </div>
                <div class="setup-step ${this._setupStatus.hasClient ? 'completed' : ''}" data-command="createClient">
                    <div class="step-icon ${this._setupStatus.hasClient ? 'done' : 'pending'}">
                        ${this._setupStatus.hasClient ? '✓' : '2'}
                    </div>
                    <div class="step-content">
                        <div class="step-label">${this._setupStatus.hasClient ? 'Client Added' : 'Add Client'}</div>
                        <div class="step-desc">Who you work with</div>
                    </div>
                </div>
                <div class="setup-step ${this._setupStatus.hasContract ? 'completed' : ''}" data-command="createContract">
                    <div class="step-icon ${this._setupStatus.hasContract ? 'done' : 'pending'}">
                        ${this._setupStatus.hasContract ? '✓' : '3'}
                    </div>
                    <div class="step-content">
                        <div class="step-label">${this._setupStatus.hasContract ? 'Contract Created' : 'Create Contract'}</div>
                        <div class="step-desc">Rates and terms</div>
                    </div>
                </div>
            </div>
        </div>
    `;
  }

  private _getMetricsHtml(): string {
    const m = this._metrics || {
      totalMonthlyRevenue: 0,
      projectedHours: 0,
      pendingInvoices: 0,
      unpaidAmount: 0,
      totalClients: 0,
      activeContracts: 0,
      averageHourlyRate: 0,
      currency: 'USD',
    };

    return `
        <div class="metric-card" data-command="openStatistics">
            <div class="metric-header">
                <span class="metric-icon">💰</span>
                <span class="metric-label">Revenue</span>
            </div>
            <div class="metric-value green">${this._formatCurrency(m.totalMonthlyRevenue, m.currency)}</div>
        </div>
        <div class="metric-card" data-command="openTracking">
            <div class="metric-header">
                <span class="metric-icon">📅</span>
                <span class="metric-label">Today</span>
            </div>
            <div class="metric-value blue">${this._todayHours.toFixed(1)}h</div>
        </div>
        <div class="metric-card" data-command="openInvoices">
            <div class="metric-header">
                <span class="metric-icon">📄</span>
                <span class="metric-label">Pending</span>
            </div>
            <div class="metric-value ${m.pendingInvoices > 0 ? 'orange' : ''}">${m.pendingInvoices}</div>
        </div>
        <div class="metric-card" data-command="openStatistics">
            <div class="metric-header">
                <span class="metric-icon">💵</span>
                <span class="metric-label">Avg Rate</span>
            </div>
            <div class="metric-value purple">${m.averageHourlyRate > 0 ? `$${m.averageHourlyRate.toFixed(0)}/h` : '-'}</div>
        </div>
    `;
  }

  private _getQuickActionsHtml(): string {
    if (this._activeTracking) {
      return `
            <button class="action-btn" data-command="createInvoice">
                <span class="action-icon">📄</span>
                <span class="action-label">New Invoice</span>
            </button>
            <button class="action-btn" data-command="logExpense">
                <span class="action-icon">💳</span>
                <span class="action-label">Log Expense</span>
            </button>
            <button class="action-btn" data-command="createClient">
                <span class="action-icon">👤</span>
                <span class="action-label">Add Client</span>
            </button>
            <button class="action-btn" data-command="createContract">
                <span class="action-icon">📝</span>
                <span class="action-label">New Contract</span>
            </button>
        `;
    }

    return `
        <button class="action-btn primary" data-command="startTracking">
            <span class="action-icon">▶️</span>
            <span class="action-label">Start Tracking</span>
        </button>
        <button class="action-btn" data-command="createInvoice">
            <span class="action-icon">📄</span>
            <span class="action-label">New Invoice</span>
        </button>
        <button class="action-btn" data-command="logExpense">
            <span class="action-icon">💳</span>
            <span class="action-label">Log Expense</span>
        </button>
    `;
  }

  private _getRecentContractsHtml(): string {
    if (this._recentContracts.length === 0) {
      // Show helpful empty state like CLI
      if (!this._setupStatus.hasClient) {
        return `
            <div class="empty-state">
                <div class="empty-state-icon">👥</div>
                <div class="empty-state-text">Add a client first to create contracts</div>
                <button class="empty-state-action" data-command="createClient">Add Client</button>
            </div>
        `;
      }
      return `
            <div class="empty-state">
                <div class="empty-state-icon">📝</div>
                <div class="empty-state-text">No contracts yet. Create one to start tracking work.</div>
                <button class="empty-state-action" data-command="createContract">Create Contract</button>
            </div>
        `;
    }

    const items = this._recentContracts
      .map(
        (c) => `
            <div class="recent-item" data-command="openContracts">
                <span class="recent-item-icon">📝</span>
                <div class="recent-item-content">
                    <div class="recent-item-title">${c.client}</div>
                    <div class="recent-item-subtitle">${c.type} • ${c.rate}</div>
                </div>
            </div>
        `
      )
      .join('');

    return `<div class="recent-list">${items}</div>`;
  }

  private _getRecentInvoicesHtml(): string {
    if (this._recentInvoices.length === 0) {
      // Show helpful empty state like CLI
      if (!this._setupStatus.hasContract) {
        return `
            <div class="empty-state">
                <div class="empty-state-icon">📝</div>
                <div class="empty-state-text">Create a contract first to generate invoices</div>
                <button class="empty-state-action" data-command="createContract">Create Contract</button>
            </div>
        `;
      }
      return `
            <div class="empty-state">
                <div class="empty-state-icon">📄</div>
                <div class="empty-state-text">No invoices yet. Create one when ready to bill.</div>
                <button class="empty-state-action" data-command="createInvoice">Create Invoice</button>
            </div>
        `;
    }

    const items = this._recentInvoices
      .map((inv) => {
        const statusClass = inv.status.toLowerCase().replace(/\s+/g, '-');
        return `
            <div class="recent-item" data-command="openInvoices">
                <span class="recent-item-icon">📄</span>
                <div class="recent-item-content">
                    <div class="recent-item-title">${inv.invoiceNum} - ${inv.client}</div>
                    <div class="recent-item-subtitle">${inv.amount}</div>
                </div>
                <span class="recent-item-badge ${statusClass}">${inv.status}</span>
            </div>
        `;
      })
      .join('');

    return `<div class="recent-list">${items}</div>`;
  }
}
