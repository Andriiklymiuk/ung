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
 * Recent time session info
 */
interface RecentSession {
  id: number;
  project: string;
  client: string;
  duration: string;
  date: string;
}

/**
 * Recent expense info
 */
interface RecentExpense {
  id: number;
  description: string;
  amount: string;
  category: string;
  date: string;
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
  private _isLoading: boolean = true;
  private _activeTracking: {
    project: string;
    client: string;
    duration: string;
    startTime: number; // Unix timestamp in seconds for live timer
  } | null = null;
  private _recentContracts: RecentContract[] = [];
  private _recentInvoices: RecentInvoice[] = [];
  private _recentSessions: RecentSession[] = [];
  private _recentExpenses: RecentExpense[] = [];
  private _weeklyHours: number = 0;
  private _weeklyTarget: number = 40;
  private _trackingStreak: number = 0;
  private _secureMode: boolean = false;
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

  public toggleSecureMode(): void {
    this._secureMode = !this._secureMode;
    if (this._view) {
      this._view.webview.html = this._getHtmlForWebview();
    }
  }

  public isSecureMode(): boolean {
    return this._secureMode;
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
          // Check prerequisites: company → client → contract
          if (!this._setupStatus.hasCompany) {
            const choice = await vscode.window.showWarningMessage(
              'Add a company first to start tracking',
              'Add Company'
            );
            if (choice === 'Add Company') {
              vscode.commands.executeCommand('ung.createCompany');
            }
          } else if (!this._setupStatus.hasClient) {
            const choice = await vscode.window.showWarningMessage(
              'Add a client first to start tracking',
              'Add Client'
            );
            if (choice === 'Add Client') {
              vscode.commands.executeCommand('ung.createClient');
            }
          } else if (!this._setupStatus.hasContract) {
            const choice = await vscode.window.showWarningMessage(
              'Create a contract first to start tracking',
              'Create Contract'
            );
            if (choice === 'Create Contract') {
              vscode.commands.executeCommand('ung.createContract');
            }
          } else {
            vscode.commands.executeCommand('ung.startTracking');
          }
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
        case 'setWeeklyTarget':
          await this._promptSetWeeklyTarget();
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
        case 'viewInvoice':
          if (message.invoiceId) {
            vscode.commands.executeCommand('ung.viewInvoice', {
              itemId: message.invoiceId,
            });
          }
          break;
        case 'emailInvoice':
          if (message.invoiceId) {
            vscode.commands.executeCommand(
              'ung.emailInvoice',
              message.invoiceId
            );
          }
          break;
        case 'deleteExpense':
          if (message.expenseId) {
            const confirm = await vscode.window.showWarningMessage(
              'Delete this expense?',
              { modal: true },
              'Delete'
            );
            if (confirm === 'Delete') {
              await vscode.commands.executeCommand('ung.deleteExpense', {
                itemId: message.expenseId,
              });
              await this.refresh();
            }
          }
          break;
        case 'refresh':
          await this.refresh();
          break;
        case 'toggleSecureMode':
          this._secureMode = !this._secureMode;
          if (this._view) {
            this._view.webview.html = this._getHtmlForWebview();
          }
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
      this._loadRecentContracts(),
      this._loadRecentInvoices(),
      this._loadRecentSessions(),
      this._loadRecentExpenses(),
      this._loadWeeklyHours(),
      this._loadWeeklyTarget(),
      this._loadTrackingStreak(),
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
    // Filter out table borders, headers, and "No data" messages
    const lines = output
      .split('\n')
      .filter(
        (l) =>
          l.trim() &&
          !l.includes('─') &&
          !l.includes('ID') &&
          !l.toLowerCase().startsWith('no ')
      );
    return Math.max(0, lines.length - 1); // Subtract header
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

      // Check for actual data rows (not just headers/empty tables)
      const companyLines = (companyResult.stdout || '')
        .split('\n')
        .filter((l) => l.trim() && !l.includes('─') && !l.includes('ID'));
      const clientLines = (clientResult.stdout || '')
        .split('\n')
        .filter((l) => l.trim() && !l.includes('─') && !l.includes('ID'));
      const contractLines = (contractResult.stdout || '')
        .split('\n')
        .filter((l) => l.trim() && !l.includes('─') && !l.includes('ID'));

      this._setupStatus = {
        hasCompany:
          companyResult.success &&
          companyLines.length > 1 &&
          !companyResult.stdout?.includes('No company'),
        hasClient: clientResult.success && clientLines.length > 1,
        hasContract: contractResult.success && contractLines.length > 1,
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

  private async _loadRecentSessions(): Promise<void> {
    try {
      const result = await this.cli.exec(['track', 'list', '--limit', '3']);
      if (result.success && result.stdout) {
        this._recentSessions = this._parseSessionsFromOutput(result.stdout);
      } else {
        this._recentSessions = [];
      }
    } catch {
      this._recentSessions = [];
    }
  }

  private _parseSessionsFromOutput(output: string): RecentSession[] {
    const lines = output
      .split('\n')
      .filter((l) => l.trim() && !l.includes('─') && !l.includes('ID'));
    if (lines.length < 2) return [];

    const sessions: RecentSession[] = [];
    // Skip header line
    for (let i = 1; i < lines.length && sessions.length < 3; i++) {
      const parts = lines[i]
        .split(/\s{2,}/)
        .map((p) => p.trim())
        .filter((p) => p);
      if (parts.length >= 4) {
        const id = parseInt(parts[0], 10);
        if (!Number.isNaN(id)) {
          sessions.push({
            id,
            project: parts[1] || 'Unknown',
            client: parts[2] || '',
            duration: parts[3] || '0:00',
            date: parts[4] || 'Today',
          });
        }
      }
    }
    return sessions;
  }

  private async _loadRecentExpenses(): Promise<void> {
    try {
      const result = await this.cli.exec(['expense', 'list']);
      if (result.success && result.stdout) {
        this._recentExpenses = this._parseExpensesFromOutput(result.stdout);
      } else {
        this._recentExpenses = [];
      }
    } catch {
      this._recentExpenses = [];
    }
  }

  private _parseExpensesFromOutput(output: string): RecentExpense[] {
    const lines = output
      .split('\n')
      .filter((l) => l.trim() && !l.includes('─') && !l.includes('ID'));
    if (lines.length < 2) return [];

    const expenses: RecentExpense[] = [];
    // Skip header line
    for (let i = 1; i < lines.length && expenses.length < 3; i++) {
      const parts = lines[i]
        .split(/\s{2,}/)
        .map((p) => p.trim())
        .filter((p) => p);
      if (parts.length >= 4) {
        const id = parseInt(parts[0], 10);
        if (!Number.isNaN(id)) {
          expenses.push({
            id,
            description: parts[1] || 'Unknown',
            amount: parts[2] || '$0',
            category: parts[3] || 'other',
            date: parts[4] || '',
          });
        }
      }
    }
    return expenses;
  }

  private async _loadWeeklyHours(): Promise<void> {
    try {
      // Get this week's sessions and sum up hours
      const result = await this.cli.exec(['track', 'list', '--week']);
      if (result.success && result.stdout) {
        this._weeklyHours = this._parseWeeklyHours(result.stdout);
      } else {
        this._weeklyHours = 0;
      }
    } catch {
      this._weeklyHours = 0;
    }
  }

  private _parseWeeklyHours(output: string): number {
    const lines = output
      .split('\n')
      .filter((l) => l.trim() && !l.includes('─') && !l.includes('ID'));
    if (lines.length < 2) return 0;

    let totalMinutes = 0;
    // Skip header line, sum durations
    for (let i = 1; i < lines.length; i++) {
      const parts = lines[i]
        .split(/\s{2,}/)
        .map((p) => p.trim())
        .filter((p) => p);
      if (parts.length >= 4) {
        const duration = parts[3] || '0:00';
        const match = duration.match(/(\d+):(\d+)(?::(\d+))?/);
        if (match) {
          const hours = parseInt(match[1], 10) || 0;
          const minutes = parseInt(match[2], 10) || 0;
          totalMinutes += hours * 60 + minutes;
        }
      }
    }
    return totalMinutes / 60; // Return as decimal hours
  }

  private async _loadWeeklyTarget(): Promise<void> {
    try {
      const result = await this.cli.exec([
        'settings',
        'get',
        'weekly_hours_target',
      ]);
      if (result.success && result.stdout) {
        const target = parseFloat(result.stdout.trim());
        if (!Number.isNaN(target) && target > 0) {
          this._weeklyTarget = target;
        }
      }
    } catch {
      // Keep default of 40
    }
  }

  private async _loadTrackingStreak(): Promise<void> {
    try {
      const result = await this.cli.exec(['settings', 'streak']);
      if (result.success && result.stdout) {
        // Parse "X day streak" from output
        const match = result.stdout.match(/(\d+)\s*day/);
        if (match) {
          this._trackingStreak = parseInt(match[1], 10);
        }
      }
    } catch {
      this._trackingStreak = 0;
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
    if (this._secureMode) {
      return '****';
    }
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency,
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(amount);
  }

  private _maskDuration(duration: string): string {
    return this._secureMode ? '**:**' : duration;
  }

  private _maskAmount(amount: string): string {
    return this._secureMode ? '****' : amount;
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
        /* ==============================================
           UNG Design System - VSCode Extension
           Aligned with macOS DesignTokens.swift
           ============================================== */

        /* Design Tokens - Colors matching macOS app */
        :root {
            /* Brand Colors - Professional blue palette */
            --ung-brand: #3373E8;
            --ung-brand-light: rgba(51, 115, 232, 0.15);
            --ung-brand-dark: #2660CC;

            /* Semantic Colors - Matching macOS DesignTokens */
            --ung-success: #33A756;
            --ung-success-light: rgba(51, 167, 86, 0.12);
            --ung-warning: #F29932;
            --ung-warning-light: rgba(242, 153, 50, 0.12);
            --ung-error: #E65A5A;
            --ung-error-light: rgba(230, 90, 90, 0.12);
            --ung-info: #3373E8;
            --ung-info-light: rgba(51, 115, 232, 0.12);

            /* Accent Colors */
            --ung-purple: #8C59B2;
            --ung-purple-light: rgba(140, 89, 178, 0.12);
            --ung-indigo: #5966BF;

            /* Spacing - 8pt grid system */
            --space-xxs: 4px;
            --space-xs: 8px;
            --space-sm: 12px;
            --space-md: 16px;
            --space-lg: 24px;

            /* Border Radius */
            --radius-xs: 4px;
            --radius-sm: 8px;
            --radius-md: 12px;
            --radius-lg: 16px;
            --radius-full: 9999px;

            /* Transitions */
            --transition-quick: 0.15s ease-out;
            --transition-standard: 0.25s ease-in-out;
        }

        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            font-family: var(--vscode-font-family);
            color: var(--vscode-foreground);
            background-color: var(--vscode-sideBar-background);
            padding: var(--space-sm);
            line-height: 1.5;
            font-size: 13px;
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
            padding: 48px var(--space-md);
        }

        .spinner {
            width: 28px;
            height: 28px;
            border: 2px solid var(--vscode-input-background);
            border-top: 2px solid var(--ung-brand);
            border-radius: var(--radius-full);
            animation: spin 0.8s linear infinite;
        }

        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }

        .loading-text {
            margin-top: var(--space-sm);
            color: var(--vscode-descriptionForeground);
            font-size: 12px;
        }

        /* Alert Banner - Unpaid invoices warning */
        .alert-banner {
            display: flex;
            align-items: center;
            gap: var(--space-xs);
            padding: var(--space-xs) var(--space-sm);
            background-color: var(--ung-error-light);
            border: 1px solid var(--ung-error);
            border-radius: var(--radius-sm);
            margin-bottom: var(--space-sm);
            font-size: 12px;
            cursor: pointer;
            transition: all var(--transition-quick);
        }

        .alert-banner:hover {
            background-color: rgba(230, 90, 90, 0.2);
        }

        .alert-icon {
            font-size: 14px;
        }

        .alert-text {
            flex: 1;
        }

        .alert-value {
            font-weight: 600;
            color: var(--ung-error);
        }

        /* Active Tracking Banner */
        .tracking-banner {
            display: flex;
            align-items: center;
            gap: var(--space-xs);
            padding: var(--space-sm);
            background-color: var(--ung-success-light);
            border: 1px solid var(--ung-success);
            border-radius: var(--radius-sm);
            margin-bottom: var(--space-sm);
        }

        .tracking-indicator {
            width: 8px;
            height: 8px;
            background-color: var(--ung-success);
            border-radius: var(--radius-full);
            animation: pulse 2s infinite;
            flex-shrink: 0;
        }

        @keyframes pulse {
            0%, 100% { opacity: 1; transform: scale(1); }
            50% { opacity: 0.5; transform: scale(0.9); }
        }

        .tracking-info {
            flex: 1;
            min-width: 0;
        }

        .tracking-project {
            font-weight: 600;
            font-size: 13px;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .tracking-meta {
            font-size: 11px;
            color: var(--vscode-descriptionForeground);
        }

        .tracking-duration {
            font-size: 18px;
            font-weight: 700;
            font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, monospace;
            color: var(--ung-success);
            letter-spacing: -0.5px;
        }

        .stop-btn {
            background-color: var(--ung-error);
            border: none;
            color: white;
            padding: 6px 12px;
            border-radius: var(--radius-sm);
            cursor: pointer;
            font-size: 12px;
            font-weight: 500;
            transition: all var(--transition-quick);
        }

        .stop-btn:hover {
            background-color: #D14545;
            transform: scale(1.02);
        }

        /* Revenue Bar */
        .revenue-bar {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: var(--space-sm);
            background-color: var(--vscode-input-background);
            border-radius: var(--radius-sm);
            margin-bottom: var(--space-xs);
            cursor: pointer;
            transition: all var(--transition-quick);
            border: 1px solid transparent;
        }

        .revenue-bar:hover {
            background-color: var(--vscode-list-hoverBackground);
            border-color: var(--ung-brand);
        }

        .revenue-label {
            font-size: 11px;
            color: var(--vscode-descriptionForeground);
            text-transform: uppercase;
            letter-spacing: 0.5px;
            font-weight: 500;
        }

        .revenue-value {
            font-size: 16px;
            font-weight: 700;
            color: var(--ung-success);
        }

        /* Section */
        .section {
            margin-bottom: var(--space-md);
        }

        .section-header {
            display: flex;
            align-items: center;
            gap: 6px;
            margin-bottom: var(--space-xs);
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
            gap: var(--space-xs);
        }

        .action-btn {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: var(--space-xs);
            padding: var(--space-sm);
            background-color: var(--vscode-input-background);
            border: 1px solid transparent;
            border-radius: var(--radius-sm);
            cursor: pointer;
            font-size: 12px;
            font-family: var(--vscode-font-family);
            color: var(--vscode-foreground);
            transition: all var(--transition-quick);
            text-align: center;
            font-weight: 500;
        }

        .action-btn:hover {
            background-color: var(--vscode-list-hoverBackground);
            border-color: var(--ung-brand);
            transform: translateY(-1px);
        }

        .action-btn.primary {
            background-color: var(--ung-brand);
            color: white;
            grid-column: span 2;
        }

        .action-btn.primary:hover {
            background-color: var(--ung-brand-dark);
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
            gap: var(--space-xs);
        }

        .nav-grid-compact {
            display: flex;
            gap: var(--space-xs);
        }

        .nav-grid-compact .nav-item {
            flex: 1;
            padding: var(--space-sm) var(--space-xs);
        }

        .nav-item {
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: var(--space-xxs);
            padding: var(--space-sm) var(--space-xs);
            background-color: var(--vscode-input-background);
            border: 1px solid transparent;
            border-radius: var(--radius-sm);
            cursor: pointer;
            font-size: 11px;
            font-family: var(--vscode-font-family);
            color: var(--vscode-foreground);
            transition: all var(--transition-quick);
            position: relative;
        }

        .nav-item:hover {
            background-color: var(--vscode-list-hoverBackground);
            border-color: var(--ung-brand);
            transform: translateY(-1px);
        }

        .nav-icon {
            font-size: 16px;
        }

        .nav-label {
            font-weight: 500;
        }

        .nav-badge {
            position: absolute;
            top: var(--space-xxs);
            right: var(--space-xxs);
            min-width: 18px;
            height: 18px;
            padding: 0 5px;
            background-color: var(--ung-brand);
            color: white;
            border-radius: var(--radius-full);
            font-size: 10px;
            font-weight: 600;
            display: flex;
            align-items: center;
            justify-content: center;
        }

        /* Setup Section */
        .setup-section {
            background-color: var(--ung-brand-light);
            border: 1px solid rgba(51, 115, 232, 0.3);
            border-radius: var(--radius-md);
            padding: var(--space-md);
            margin-bottom: var(--space-md);
        }

        .setup-title {
            font-size: 13px;
            font-weight: 600;
            margin-bottom: var(--space-sm);
            display: flex;
            align-items: center;
            gap: var(--space-xs);
        }

        .setup-steps {
            display: flex;
            flex-direction: column;
            gap: var(--space-xs);
        }

        .setup-step {
            display: flex;
            align-items: center;
            gap: var(--space-sm);
            padding: var(--space-sm);
            background-color: var(--vscode-sideBar-background);
            border-radius: var(--radius-sm);
            cursor: pointer;
            transition: all var(--transition-quick);
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
            border-radius: var(--radius-full);
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 11px;
            font-weight: 600;
            flex-shrink: 0;
        }

        .step-icon.pending {
            background-color: transparent;
            border: 2px solid var(--ung-brand);
            color: var(--ung-brand);
        }

        .step-icon.done {
            background-color: var(--ung-success);
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
            background: var(--vscode-panel-border);
            margin: var(--space-md) 0;
            opacity: 0.5;
        }

        /* Tools Grid */
        .tools-grid {
            display: flex;
            gap: var(--space-xs);
        }

        .tool-item {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: var(--space-xs);
            padding: var(--space-sm) var(--space-md);
            background-color: var(--vscode-input-background);
            border-radius: var(--radius-sm);
            cursor: pointer;
            font-size: 12px;
            font-weight: 500;
            font-family: var(--vscode-font-family);
            color: var(--vscode-foreground);
            transition: all var(--transition-quick);
            border: 1px solid transparent;
            flex: 1;
        }

        .tool-item:hover {
            background-color: var(--vscode-list-hoverBackground);
            border-color: var(--ung-brand);
            transform: translateY(-1px);
        }

        .tool-icon {
            font-size: 14px;
        }

        /* Recent Items List */
        .recent-list {
            display: flex;
            flex-direction: column;
            gap: var(--space-xxs);
        }

        .recent-item {
            display: flex;
            align-items: center;
            gap: var(--space-xs);
            padding: var(--space-xs) var(--space-sm);
            background-color: var(--vscode-input-background);
            border-radius: var(--radius-sm);
            cursor: pointer;
            transition: all var(--transition-quick);
            border: 1px solid transparent;
        }

        .recent-item:hover {
            background-color: var(--vscode-list-hoverBackground);
            border-color: var(--ung-brand);
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
            font-size: 12px;
            font-weight: 500;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .recent-item-subtitle {
            font-size: 11px;
            color: var(--vscode-descriptionForeground);
        }

        .recent-item-badge {
            font-size: 10px;
            padding: 2px 8px;
            border-radius: var(--radius-xs);
            font-weight: 500;
            text-transform: capitalize;
        }

        .recent-item-badge.paid {
            background-color: var(--ung-success-light);
            color: var(--ung-success);
        }

        .recent-item-badge.pending, .recent-item-badge.sent {
            background-color: var(--ung-warning-light);
            color: var(--ung-warning);
        }

        .recent-item-badge.draft {
            background-color: rgba(128, 128, 128, 0.15);
            color: var(--vscode-descriptionForeground);
        }

        .recent-item-badge.overdue {
            background-color: var(--ung-error-light);
            color: var(--ung-error);
        }

        /* Session duration */
        .session-duration {
            font-size: 11px;
            font-weight: 600;
            font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, monospace;
            color: var(--ung-info);
            padding: 3px 8px;
            background: var(--ung-info-light);
            border-radius: var(--radius-xs);
        }

        /* Expense amount */
        .expense-amount {
            font-size: 11px;
            font-weight: 600;
            font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, monospace;
            color: var(--ung-warning);
            padding: 3px 8px;
            background: var(--ung-warning-light);
            border-radius: var(--radius-xs);
        }

        /* Delete button */
        .delete-btn {
            display: none;
            align-items: center;
            justify-content: center;
            width: 20px;
            height: 20px;
            border: none;
            background: var(--ung-error-light);
            color: var(--ung-error);
            border-radius: var(--radius-xs);
            cursor: pointer;
            font-size: 11px;
            font-weight: bold;
            margin-left: var(--space-xxs);
            transition: all var(--transition-quick);
        }

        .recent-item:hover .delete-btn {
            display: flex;
        }

        .delete-btn:hover {
            background: var(--ung-error);
            color: white;
        }

        /* Weekly Progress Widget */
        .weekly-progress {
            padding: var(--space-sm);
            background-color: var(--vscode-input-background);
            border-radius: var(--radius-sm);
            margin-bottom: var(--space-xs);
            cursor: pointer;
            transition: all var(--transition-quick);
            border: 1px solid transparent;
        }

        .weekly-progress:hover {
            background-color: var(--vscode-list-hoverBackground);
            border-color: var(--ung-brand);
        }

        .weekly-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: var(--space-xs);
        }

        .weekly-label {
            font-size: 11px;
            color: var(--vscode-descriptionForeground);
            text-transform: uppercase;
            letter-spacing: 0.5px;
            font-weight: 500;
        }

        .weekly-hours {
            font-size: 14px;
            font-weight: 700;
            color: var(--ung-info);
        }

        .weekly-bar {
            height: 4px;
            background-color: rgba(128, 128, 128, 0.2);
            border-radius: var(--radius-full);
            overflow: hidden;
        }

        .weekly-bar-fill {
            height: 100%;
            background: linear-gradient(90deg, var(--ung-info), var(--ung-success));
            border-radius: var(--radius-full);
            transition: width 0.4s ease-out;
        }

        .weekly-target {
            margin-top: var(--space-xxs);
            text-align: right;
        }

        .target-text {
            font-size: 10px;
            color: var(--vscode-descriptionForeground);
            cursor: pointer;
            opacity: 0.7;
            transition: all var(--transition-quick);
        }

        .target-text:hover {
            opacity: 1;
            color: var(--ung-brand);
        }

        /* Streak Badge */
        .streak-badge {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            padding: 2px 6px;
            background: linear-gradient(135deg, var(--ung-warning), #FFD93D);
            color: #1a1a1a;
            font-size: 9px;
            font-weight: 700;
            border-radius: var(--radius-full);
            margin-left: var(--space-xs);
            text-transform: none;
            letter-spacing: 0;
        }

        /* Invoice Action Buttons */
        .invoice-actions {
            display: flex;
            gap: var(--space-xxs);
            margin-left: auto;
            margin-right: var(--space-xs);
        }

        .invoice-action-btn {
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 4px 10px;
            border: none;
            background: rgba(128, 128, 128, 0.1);
            border-radius: var(--radius-xs);
            cursor: pointer;
            color: var(--vscode-descriptionForeground);
            transition: all var(--transition-quick);
            font-size: 11px;
            font-family: var(--vscode-font-family);
            font-weight: 500;
        }

        .invoice-action-btn:hover {
            background: var(--vscode-list-hoverBackground);
            color: var(--vscode-foreground);
        }

        .invoice-action-btn.send:hover {
            background: var(--ung-brand);
            color: white;
        }

        .invoice-item .recent-item-content {
            cursor: pointer;
            flex: 1;
        }

        .invoice-item .recent-item-content:hover .recent-item-title {
            color: var(--ung-brand);
        }

        /* Empty State */
        .empty-state {
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: var(--space-xs);
            padding: var(--space-md);
            background-color: rgba(128, 128, 128, 0.05);
            border: 1px dashed var(--vscode-panel-border);
            border-radius: var(--radius-md);
            text-align: center;
        }

        .empty-state.compact {
            padding: var(--space-sm);
            gap: var(--space-xxs);
        }

        .empty-state.compact .empty-state-icon {
            font-size: 16px;
        }

        .empty-state-icon {
            font-size: 20px;
            opacity: 0.5;
        }

        .empty-state-text {
            font-size: 11px;
            color: var(--vscode-descriptionForeground);
        }

        .empty-state-action {
            font-size: 11px;
            cursor: pointer;
            padding: 6px 12px;
            background-color: var(--ung-brand);
            color: white;
            border-radius: var(--radius-sm);
            border: none;
            font-weight: 500;
            margin-top: var(--space-xxs);
            transition: all var(--transition-quick);
        }

        .empty-state-action:hover {
            background-color: var(--ung-brand-dark);
        }

        /* Section with view all link */
        .section-header-with-link {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: var(--space-xs);
        }

        .section-link {
            font-size: 11px;
            color: var(--ung-brand);
            cursor: pointer;
            border: none;
            background: none;
            padding: 2px var(--space-xs);
            border-radius: var(--radius-xs);
            font-weight: 500;
            transition: all var(--transition-quick);
        }

        .section-link:hover {
            background-color: var(--ung-brand-light);
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
                    const invoiceId = target.getAttribute('data-invoice-id');
                    const expenseId = target.getAttribute('data-expense-id');
                    if (command) {
                        const message = { command: command };
                        if (invoiceId) {
                            message.invoiceId = parseInt(invoiceId, 10);
                        }
                        if (expenseId) {
                            message.expenseId = parseInt(expenseId, 10);
                        }
                        vscode.postMessage(message);
                    }
                }
            });

            // Live timer for active tracking
            const timerEl = document.getElementById('live-timer');
            const startTimeAttr = timerEl ? timerEl.getAttribute('data-start-time') : null;
            const isSecure = timerEl ? timerEl.getAttribute('data-secure') === 'true' : false;

            if (timerEl && startTimeAttr && !isSecure) {
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
        <div class="loading-overlay">
            <div class="spinner"></div>
            <div class="loading-text">Loading...</div>
        </div>
    `;
  }

  private _getContentHtml(needsSetup: boolean): string {
    const m = this._metrics;
    const hasUnpaid = m && m.unpaidAmount > 0;
    const revenue = m
      ? this._formatCurrency(m.totalMonthlyRevenue, m.currency)
      : '$0';

    return `
        ${this._activeTracking ? this._getActiveTrackingHtml() : ''}

        ${hasUnpaid ? this._getAlertBannerHtml() : ''}

        ${needsSetup ? this._getSetupSectionHtml() : ''}

        <!-- Revenue + Quick Actions combined -->
        <div class="section">
            <div class="revenue-bar" data-command="openStatistics">
                <span class="revenue-label">Revenue</span>
                <span class="revenue-value">${revenue}</span>
            </div>
            ${this._getWeeklyProgressHtml()}
            <div class="quick-actions">
                ${this._getQuickActionsHtml()}
            </div>
        </div>

        <div class="divider"></div>

        <!-- Recent Time Sessions -->
        <div class="section">
            <div class="section-header-with-link">
                <span class="section-title">Time</span>
                ${this._recentSessions.length > 0 ? '<button class="section-link" data-command="openTracking">All</button>' : ''}
            </div>
            ${this._getRecentSessionsHtml()}
        </div>

        <!-- Recent Contracts -->
        <div class="section">
            <div class="section-header-with-link">
                <span class="section-title">Contracts</span>
                ${this._recentContracts.length > 0 ? '<button class="section-link" data-command="openContracts">All</button>' : ''}
            </div>
            ${this._getRecentContractsHtml()}
        </div>

        <!-- Recent Invoices -->
        <div class="section">
            <div class="section-header-with-link">
                <span class="section-title">Invoices</span>
                ${this._recentInvoices.length > 0 ? '<button class="section-link" data-command="openInvoices">All</button>' : ''}
            </div>
            ${this._getRecentInvoicesHtml()}
        </div>

        <!-- Recent Expenses -->
        <div class="section">
            <div class="section-header-with-link">
                <span class="section-title">Expenses</span>
                ${this._recentExpenses.length > 0 ? '<button class="section-link" data-command="openExpenses">All</button>' : ''}
            </div>
            ${this._getRecentExpensesHtml()}
        </div>

        <div class="divider"></div>

        <!-- Quick Access -->
        <div class="nav-grid nav-grid-compact">
            <button class="nav-item" data-command="openExpenses">
                <span class="nav-label">Expenses</span>
                ${this._counts.expenses > 0 ? `<span class="nav-badge">${this._counts.expenses}</span>` : ''}
            </button>
            <button class="nav-item" data-command="openClients">
                <span class="nav-label">Clients</span>
                ${this._counts.clients > 0 ? `<span class="nav-badge">${this._counts.clients}</span>` : ''}
            </button>
            <button class="nav-item" data-command="openStatistics">
                <span class="nav-label">Reports</span>
            </button>
        </div>
    `;
  }

  private _getAlertBannerHtml(): string {
    const m = this._metrics;
    if (!m || m.unpaidAmount <= 0) return '';

    return `
        <div class="alert-banner" data-command="openInvoices">
            <span class="alert-value">${this._formatCurrency(m.unpaidAmount, m.currency)}</span> unpaid
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
            <div class="tracking-duration" id="live-timer" data-start-time="${this._activeTracking.startTime}" data-secure="${this._secureMode}">${this._secureMode ? '**:**:**' : this._activeTracking.duration}</div>
            <button class="stop-btn" data-command="stopTracking">Stop</button>
        </div>
    `;
  }

  private _getSetupSectionHtml(): string {
    return `
        <div class="setup-section">
            <div class="setup-title">Get Started</div>
            <div class="setup-steps">
                <div class="setup-step ${this._setupStatus.hasCompany ? 'completed' : ''}" data-command="createCompany">
                    <div class="step-icon ${this._setupStatus.hasCompany ? 'done' : 'pending'}">
                        ${this._setupStatus.hasCompany ? '✓' : '1'}
                    </div>
                    <div class="step-content">
                        <div class="step-label">${this._setupStatus.hasCompany ? 'Company Added' : 'Add Your Company'}</div>
                    </div>
                </div>
                <div class="setup-step ${this._setupStatus.hasClient ? 'completed' : ''}" data-command="createClient">
                    <div class="step-icon ${this._setupStatus.hasClient ? 'done' : 'pending'}">
                        ${this._setupStatus.hasClient ? '✓' : '2'}
                    </div>
                    <div class="step-content">
                        <div class="step-label">${this._setupStatus.hasClient ? 'Client Added' : 'Add Client'}</div>
                    </div>
                </div>
                <div class="setup-step ${this._setupStatus.hasContract ? 'completed' : ''}" data-command="createContract">
                    <div class="step-icon ${this._setupStatus.hasContract ? 'done' : 'pending'}">
                        ${this._setupStatus.hasContract ? '✓' : '3'}
                    </div>
                    <div class="step-content">
                        <div class="step-label">${this._setupStatus.hasContract ? 'Contract Created' : 'Create Contract'}</div>
                    </div>
                </div>
            </div>
        </div>
    `;
  }

  private _getQuickActionsHtml(): string {
    if (this._activeTracking) {
      return `
            <button class="action-btn" data-command="createInvoice">
                <span class="action-label">+ Invoice</span>
            </button>
            <button class="action-btn" data-command="logExpense">
                <span class="action-label">+ Expense</span>
            </button>
            <button class="action-btn" data-command="createClient">
                <span class="action-label">+ Client</span>
            </button>
            <button class="action-btn" data-command="createContract">
                <span class="action-label">+ Contract</span>
            </button>
        `;
    }

    return `
        <button class="action-btn primary" data-command="startTracking">
            <span class="action-label">Start Tracking</span>
        </button>
        <button class="action-btn" data-command="createInvoice">
            <span class="action-label">+ Invoice</span>
        </button>
        <button class="action-btn" data-command="logExpense">
            <span class="action-label">+ Expense</span>
        </button>
    `;
  }

  private _getRecentContractsHtml(): string {
    if (this._recentContracts.length === 0) {
      if (!this._setupStatus.hasClient) {
        return `
            <div class="empty-state compact">
                <div class="empty-state-text">Add a client first</div>
                <button class="empty-state-action" data-command="createClient">Add Client</button>
            </div>
        `;
      }
      return `
            <div class="empty-state compact">
                <div class="empty-state-text">No contracts yet</div>
                <button class="empty-state-action" data-command="createContract">+ Contract</button>
            </div>
        `;
    }

    // Check if multiple contracts per client exist
    const clientCounts = new Map<string, number>();
    for (const c of this._recentContracts) {
      clientCounts.set(c.client, (clientCounts.get(c.client) || 0) + 1);
    }

    const items = this._recentContracts
      .map((c) => {
        const hasMultiple = (clientCounts.get(c.client) || 0) > 1;
        const title = hasMultiple ? `${c.name} (${c.client})` : c.client;
        return `
            <div class="recent-item" data-command="openContracts">
                <div class="recent-item-content">
                    <div class="recent-item-title">${title}</div>
                    <div class="recent-item-subtitle">${c.type} ${this._maskAmount(c.rate)}</div>
                </div>
            </div>
        `;
      })
      .join('');

    return `<div class="recent-list">${items}</div>`;
  }

  private _getRecentInvoicesHtml(): string {
    if (this._recentInvoices.length === 0) {
      if (!this._setupStatus.hasContract) {
        return `
            <div class="empty-state compact">
                <div class="empty-state-text">Create a contract first</div>
                <button class="empty-state-action" data-command="createContract">+ Contract</button>
            </div>
        `;
      }
      return `
            <div class="empty-state compact">
                <div class="empty-state-text">No invoices yet</div>
                <button class="empty-state-action" data-command="createInvoice">+ Invoice</button>
            </div>
        `;
    }

    const items = this._recentInvoices
      .map((inv) => {
        const statusClass = inv.status.toLowerCase().replace(/\s+/g, '-');
        return `
            <div class="recent-item invoice-item" data-invoice-id="${inv.id}">
                <div class="recent-item-content" data-command="viewInvoice" data-invoice-id="${inv.id}">
                    <div class="recent-item-title">${inv.invoiceNum} - ${inv.client}</div>
                    <div class="recent-item-subtitle">${this._maskAmount(inv.amount)}</div>
                </div>
                <div class="invoice-actions">
                    <button class="invoice-action-btn view" data-command="viewInvoice" data-invoice-id="${inv.id}" title="View invoice">
                        View
                    </button>
                    <button class="invoice-action-btn send" data-command="emailInvoice" data-invoice-id="${inv.id}" title="Send invoice by email">
                        Send
                    </button>
                </div>
                <span class="recent-item-badge ${statusClass}">${inv.status}</span>
            </div>
        `;
      })
      .join('');

    return `<div class="recent-list">${items}</div>`;
  }

  private _getRecentExpensesHtml(): string {
    if (this._recentExpenses.length === 0) {
      return `
            <div class="empty-state compact">
                <div class="empty-state-text">No expenses yet</div>
                <button class="empty-state-action" data-command="logExpense">+ Expense</button>
            </div>
        `;
    }

    const items = this._recentExpenses
      .map((e) => {
        return `
            <div class="recent-item expense-item" data-expense-id="${e.id}">
                <div class="recent-item-content" data-command="openExpenses">
                    <div class="recent-item-title">${e.description}</div>
                    <div class="recent-item-subtitle">${e.category} - ${e.date}</div>
                </div>
                <span class="expense-amount">${this._maskAmount(e.amount)}</span>
                <button class="delete-btn" data-command="deleteExpense" data-expense-id="${e.id}" title="Delete expense">x</button>
            </div>
        `;
      })
      .join('');

    return `<div class="recent-list">${items}</div>`;
  }

  private _getRecentSessionsHtml(): string {
    if (this._recentSessions.length === 0) {
      return `
            <div class="empty-state compact">
                <div class="empty-state-text">No time tracked yet</div>
            </div>
        `;
    }

    const items = this._recentSessions
      .map((s) => {
        const subtitle = s.client ? `${s.client} - ${s.date}` : s.date;
        return `
            <div class="recent-item" data-command="openTracking">
                <div class="recent-item-content">
                    <div class="recent-item-title">${s.project}</div>
                    <div class="recent-item-subtitle">${subtitle}</div>
                </div>
                <span class="session-duration">${this._maskDuration(s.duration)}</span>
            </div>
        `;
      })
      .join('');

    return `<div class="recent-list">${items}</div>`;
  }

  private async _promptSetWeeklyTarget(): Promise<void> {
    const input = await vscode.window.showInputBox({
      prompt: 'Set your weekly hours target',
      placeHolder: 'e.g., 40',
      value: this._weeklyTarget.toString(),
      validateInput: (value) => {
        const num = parseFloat(value);
        if (Number.isNaN(num) || num <= 0) {
          return 'Please enter a positive number';
        }
        if (num > 168) {
          return 'Cannot exceed 168 hours (7 days × 24 hours)';
        }
        return null;
      },
    });

    if (input) {
      const target = parseFloat(input);
      const result = await this.cli.exec([
        'settings',
        'target',
        target.toString(),
      ]);
      if (result.success) {
        this._weeklyTarget = target;
        vscode.window.showInformationMessage(
          `Weekly target set to ${target} hours`
        );
        await this.refresh();
      } else {
        vscode.window.showErrorMessage('Failed to save weekly target');
      }
    }
  }

  private _getWeeklyProgressHtml(): string {
    if (this._weeklyHours === 0 && this._recentSessions.length === 0) {
      return ''; // Don't show if no tracking data at all
    }

    const hours = Math.floor(this._weeklyHours);
    const minutes = Math.round((this._weeklyHours - hours) * 60);
    const displayTime = this._secureMode
      ? '**h **m'
      : minutes > 0
        ? `${hours}h ${minutes}m`
        : `${hours}h`;

    // Progress towards customizable weekly target
    const progress = Math.min(
      (this._weeklyHours / this._weeklyTarget) * 100,
      100
    );

    // Streak badge (only show if streak > 1)
    const streakBadge =
      this._trackingStreak > 1
        ? `<span class="streak-badge" title="${this._trackingStreak} day streak">${this._trackingStreak}d</span>`
        : '';

    const targetDisplay = this._secureMode ? '**h' : `${this._weeklyTarget}h`;

    return `
        <div class="weekly-progress">
            <div class="weekly-header">
                <span class="weekly-label" data-command="openTracking">This Week ${streakBadge}</span>
                <span class="weekly-hours" data-command="openTracking">${displayTime}</span>
            </div>
            <div class="weekly-bar" data-command="openTracking">
                <div class="weekly-bar-fill" style="width: ${this._secureMode ? 0 : progress}%"></div>
            </div>
            <div class="weekly-target">
                <span class="target-text" data-command="setWeeklyTarget" title="Click to change weekly target">Goal: ${targetDisplay}</span>
            </div>
        </div>
    `;
  }
}
