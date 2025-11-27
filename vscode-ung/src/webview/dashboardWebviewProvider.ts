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
 * Dashboard webview provider for the sidebar
 * Shows a professional dashboard with metrics, quick actions, and navigation
 */
export class DashboardWebviewProvider implements vscode.WebviewViewProvider {
  public static readonly viewType = 'ungDashboard';

  private _view?: vscode.WebviewView;
  private _extensionUri: vscode.Uri;
  private _metrics: DashboardMetrics | null = null;
  private _activeTracking: {
    project: string;
    client: string;
    duration: string;
  } | null = null;
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
          vscode.commands.executeCommand('ungInvoices.focus');
          break;
        case 'openClients':
          vscode.commands.executeCommand('ungClients.focus');
          break;
        case 'openContracts':
          vscode.commands.executeCommand('ungContracts.focus');
          break;
        case 'openExpenses':
          vscode.commands.executeCommand('ungExpenses.focus');
          break;
        case 'openTracking':
          vscode.commands.executeCommand('ungTracking.focus');
          break;
        case 'openStatistics':
          vscode.commands.executeCommand('ung.openStatistics');
          break;
        case 'openTemplateEditor':
          vscode.commands.executeCommand('ung.openTemplateEditor');
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

    // Load all data in parallel
    await Promise.all([
      this._loadMetrics(),
      this._checkActiveTracking(),
      this._checkSetupStatus(),
    ]);

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

        for (const line of lines) {
          if (line.includes('Project:')) {
            project = line.split(':')[1]?.trim() || 'Unknown';
          } else if (line.includes('Client:')) {
            client = line.split(':')[1]?.trim() || '';
          } else if (line.includes('Duration:') || line.includes('Elapsed:')) {
            duration = line.split(':').slice(1).join(':').trim() || '0:00';
          }
        }

        this._activeTracking = { project, client, duration };
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

        /* Header */
        .header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 20px;
            padding-bottom: 16px;
            border-bottom: 1px solid var(--vscode-panel-border);
        }

        .header-left {
            display: flex;
            align-items: center;
            gap: 12px;
        }

        .logo {
            width: 36px;
            height: 36px;
            background: linear-gradient(135deg, var(--vscode-textLink-foreground), var(--vscode-textLink-activeForeground, var(--vscode-textLink-foreground)));
            border-radius: 8px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 16px;
            font-weight: 600;
            color: white;
        }

        .header-title {
            font-size: 16px;
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

        /* Active Tracking Banner */
        .tracking-banner {
            display: flex;
            align-items: center;
            gap: 12px;
            padding: 12px 14px;
            background: linear-gradient(135deg, color-mix(in srgb, var(--vscode-charts-green, #4caf50) 20%, var(--vscode-sideBar-background)), var(--vscode-sideBar-background));
            border: 1px solid var(--vscode-charts-green, #4caf50);
            border-radius: 10px;
            margin-bottom: 16px;
        }

        .tracking-indicator {
            width: 10px;
            height: 10px;
            background-color: var(--vscode-charts-green, #4caf50);
            border-radius: 50%;
            animation: pulse 2s infinite;
        }

        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }

        .tracking-info {
            flex: 1;
        }

        .tracking-project {
            font-weight: 600;
            font-size: 13px;
        }

        .tracking-meta {
            font-size: 11px;
            color: var(--vscode-descriptionForeground);
        }

        .tracking-duration {
            font-size: 18px;
            font-weight: 600;
            font-family: monospace;
            color: var(--vscode-charts-green, #4caf50);
        }

        .stop-btn {
            background-color: var(--vscode-inputValidation-errorBackground, #5a1d1d);
            border: 1px solid var(--vscode-inputValidation-errorBorder, #be1100);
            color: var(--vscode-errorForeground, #f48771);
            padding: 6px 12px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 12px;
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
            gap: 10px;
            margin-bottom: 20px;
        }

        .metric-card {
            background-color: var(--vscode-input-background);
            border-radius: 8px;
            padding: 14px;
            transition: background-color 0.15s;
        }

        .metric-card:hover {
            background-color: var(--vscode-list-hoverBackground);
        }

        .metric-card.wide {
            grid-column: span 2;
        }

        .metric-icon {
            font-size: 16px;
            margin-bottom: 6px;
        }

        .metric-value {
            font-size: 20px;
            font-weight: 600;
            margin-bottom: 2px;
        }

        .metric-label {
            font-size: 11px;
            color: var(--vscode-descriptionForeground);
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }

        .metric-value.green { color: var(--vscode-charts-green, #4caf50); }
        .metric-value.blue { color: var(--vscode-charts-blue, #2196f3); }
        .metric-value.orange { color: var(--vscode-charts-orange, #ff9800); }
        .metric-value.red { color: var(--vscode-charts-red, #f44336); }

        /* Section */
        .section {
            margin-bottom: 20px;
        }

        .section-header {
            display: flex;
            align-items: center;
            gap: 8px;
            margin-bottom: 12px;
        }

        .section-title {
            font-size: 11px;
            font-weight: 600;
            color: var(--vscode-descriptionForeground);
            text-transform: uppercase;
            letter-spacing: 0.8px;
        }

        /* Quick Actions */
        .quick-actions {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 8px;
        }

        .action-btn {
            display: flex;
            align-items: center;
            gap: 10px;
            padding: 12px;
            background-color: var(--vscode-input-background);
            border: 1px solid transparent;
            border-radius: 8px;
            cursor: pointer;
            font-size: 12px;
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
            font-size: 16px;
            width: 24px;
            text-align: center;
        }

        .action-label {
            font-weight: 500;
        }

        /* Navigation */
        .nav-grid {
            display: grid;
            grid-template-columns: 1fr 1fr 1fr;
            gap: 8px;
        }

        .nav-item {
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 6px;
            padding: 14px 8px;
            background-color: var(--vscode-input-background);
            border: 1px solid transparent;
            border-radius: 8px;
            cursor: pointer;
            font-size: 11px;
            font-family: var(--vscode-font-family);
            color: var(--vscode-foreground);
            transition: all 0.15s;
        }

        .nav-item:hover {
            background-color: var(--vscode-list-hoverBackground);
            border-color: var(--vscode-focusBorder);
            transform: translateY(-2px);
        }

        .nav-icon {
            font-size: 20px;
        }

        .nav-label {
            font-weight: 500;
        }

        /* Setup Section */
        .setup-section {
            background-color: color-mix(in srgb, var(--vscode-textLink-foreground) 8%, var(--vscode-input-background));
            border: 1px solid color-mix(in srgb, var(--vscode-textLink-foreground) 30%, transparent);
            border-radius: 10px;
            padding: 16px;
            margin-bottom: 20px;
        }

        .setup-title {
            font-size: 13px;
            font-weight: 600;
            margin-bottom: 12px;
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .setup-steps {
            display: flex;
            flex-direction: column;
            gap: 8px;
        }

        .setup-step {
            display: flex;
            align-items: center;
            gap: 10px;
            padding: 10px 12px;
            background-color: var(--vscode-sideBar-background);
            border-radius: 6px;
            cursor: pointer;
            transition: all 0.15s;
        }

        .setup-step:hover {
            background-color: var(--vscode-list-hoverBackground);
        }

        .setup-step.completed {
            opacity: 0.6;
        }

        .step-icon {
            width: 24px;
            height: 24px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 12px;
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
            font-size: 12px;
            font-weight: 500;
        }

        .step-desc {
            font-size: 11px;
            color: var(--vscode-descriptionForeground);
        }

        /* Divider */
        .divider {
            height: 1px;
            background: linear-gradient(to right, transparent, var(--vscode-panel-border), transparent);
            margin: 20px 0;
        }

        /* Tools Grid */
        .tools-grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 8px;
        }

        .tool-item {
            display: flex;
            align-items: center;
            gap: 10px;
            padding: 10px 12px;
            background-color: var(--vscode-input-background);
            border-radius: 6px;
            cursor: pointer;
            font-size: 12px;
            transition: all 0.15s;
        }

        .tool-item:hover {
            background-color: var(--vscode-list-hoverBackground);
        }

        .tool-icon {
            font-size: 14px;
            opacity: 0.8;
        }
    </style>
</head>
<body>
    <div class="container">
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

        ${needsSetup ? this._getSetupSectionHtml() : ''}

        <!-- Metrics -->
        <div class="section">
            <div class="section-header">
                <span class="section-title">This Month</span>
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

        <!-- Navigation -->
        <div class="section">
            <div class="section-header">
                <span class="section-title">Navigate</span>
            </div>
            <div class="nav-grid">
                <button class="nav-item" data-command="openInvoices">
                    <span class="nav-icon">📄</span>
                    <span class="nav-label">Invoices</span>
                </button>
                <button class="nav-item" data-command="openClients">
                    <span class="nav-icon">👥</span>
                    <span class="nav-label">Clients</span>
                </button>
                <button class="nav-item" data-command="openContracts">
                    <span class="nav-icon">📝</span>
                    <span class="nav-label">Contracts</span>
                </button>
                <button class="nav-item" data-command="openExpenses">
                    <span class="nav-icon">💳</span>
                    <span class="nav-label">Expenses</span>
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
                    <span>Template Editor</span>
                </button>
                <button class="tool-item" data-command="openStatistics">
                    <span class="tool-icon">📈</span>
                    <span>Statistics</span>
                </button>
            </div>
        </div>
    </div>

    <script>
        const vscode = acquireVsCodeApi();

        document.querySelectorAll('[data-command]').forEach(btn => {
            btn.addEventListener('click', () => {
                const command = btn.getAttribute('data-command');
                vscode.postMessage({ command });
            });
        });
    </script>
</body>
</html>`;
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
            <div class="tracking-duration">${this._activeTracking.duration}</div>
            <button class="stop-btn" data-command="stopTracking">Stop</button>
        </div>
    `;
  }

  private _getSetupSectionHtml(): string {
    return `
        <div class="setup-section">
            <div class="setup-title">
                <span>🚀</span>
                <span>Quick Setup</span>
            </div>
            <div class="setup-steps">
                <div class="setup-step ${this._setupStatus.hasCompany ? 'completed' : ''}" data-command="createCompany">
                    <div class="step-icon ${this._setupStatus.hasCompany ? 'done' : 'pending'}">
                        ${this._setupStatus.hasCompany ? '✓' : '1'}
                    </div>
                    <div class="step-content">
                        <div class="step-label">${this._setupStatus.hasCompany ? 'Company Added' : 'Add Your Company'}</div>
                        <div class="step-desc">Your business details for invoices</div>
                    </div>
                </div>
                <div class="setup-step ${this._setupStatus.hasClient ? 'completed' : ''}" data-command="createClient">
                    <div class="step-icon ${this._setupStatus.hasClient ? 'done' : 'pending'}">
                        ${this._setupStatus.hasClient ? '✓' : '2'}
                    </div>
                    <div class="step-content">
                        <div class="step-label">${this._setupStatus.hasClient ? 'Client Added' : 'Add Your First Client'}</div>
                        <div class="step-desc">Who you're working with</div>
                    </div>
                </div>
                <div class="setup-step ${this._setupStatus.hasContract ? 'completed' : ''}" data-command="createContract">
                    <div class="step-icon ${this._setupStatus.hasContract ? 'done' : 'pending'}">
                        ${this._setupStatus.hasContract ? '✓' : '3'}
                    </div>
                    <div class="step-content">
                        <div class="step-label">${this._setupStatus.hasContract ? 'Contract Created' : 'Create a Contract'}</div>
                        <div class="step-desc">Define rates and terms</div>
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
      currency: 'USD',
    };

    return `
        <div class="metric-card">
            <div class="metric-icon">💰</div>
            <div class="metric-value green">${this._formatCurrency(m.totalMonthlyRevenue, m.currency)}</div>
            <div class="metric-label">Revenue</div>
        </div>
        <div class="metric-card">
            <div class="metric-icon">⏱️</div>
            <div class="metric-value blue">${m.projectedHours.toFixed(1)}h</div>
            <div class="metric-label">Hours</div>
        </div>
        <div class="metric-card">
            <div class="metric-icon">📄</div>
            <div class="metric-value ${m.pendingInvoices > 0 ? 'orange' : ''}">${m.pendingInvoices}</div>
            <div class="metric-label">Pending</div>
        </div>
        <div class="metric-card">
            <div class="metric-icon">⚠️</div>
            <div class="metric-value ${m.unpaidAmount > 0 ? 'red' : ''}">${this._formatCurrency(m.unpaidAmount, m.currency)}</div>
            <div class="metric-label">Unpaid</div>
        </div>
    `;
  }

  private _getQuickActionsHtml(): string {
    if (this._activeTracking) {
      return `
            <button class="action-btn" data-command="createInvoice">
                <span class="action-icon">📄</span>
                <span class="action-label">Create Invoice</span>
            </button>
            <button class="action-btn" data-command="logExpense">
                <span class="action-icon">💳</span>
                <span class="action-label">Log Expense</span>
            </button>
        `;
    }

    return `
        <button class="action-btn primary" data-command="startTracking">
            <span class="action-icon">▶️</span>
            <span class="action-label">Start Time Tracking</span>
        </button>
        <button class="action-btn" data-command="createInvoice">
            <span class="action-icon">📄</span>
            <span class="action-label">Create Invoice</span>
        </button>
        <button class="action-btn" data-command="logExpense">
            <span class="action-icon">💳</span>
            <span class="action-label">Log Expense</span>
        </button>
    `;
  }
}
