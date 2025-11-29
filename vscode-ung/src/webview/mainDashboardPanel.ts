import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';

/**
 * Main Dashboard Panel - Professional overview with quick actions
 */
export class MainDashboardPanel {
  public static currentPanel: MainDashboardPanel | undefined;
  private static readonly viewType = 'ungMainDashboard';

  private readonly panel: vscode.WebviewPanel;
  private readonly cli: UngCli;
  private disposables: vscode.Disposable[] = [];

  public static createOrShow(cli: UngCli) {
    const column = vscode.window.activeTextEditor
      ? vscode.window.activeTextEditor.viewColumn
      : undefined;

    if (MainDashboardPanel.currentPanel) {
      MainDashboardPanel.currentPanel.panel.reveal(column);
      MainDashboardPanel.currentPanel.update();
      return;
    }

    const panel = vscode.window.createWebviewPanel(
      MainDashboardPanel.viewType,
      'UNG Dashboard',
      column || vscode.ViewColumn.One,
      {
        enableScripts: true,
        retainContextWhenHidden: true,
      }
    );

    MainDashboardPanel.currentPanel = new MainDashboardPanel(panel, cli);
  }

  private constructor(panel: vscode.WebviewPanel, cli: UngCli) {
    this.panel = panel;
    this.cli = cli;

    this.update();

    this.panel.onDidDispose(() => this.dispose(), null, this.disposables);

    this.panel.webview.onDidReceiveMessage(
      async (message) => {
        switch (message.command) {
          case 'refresh':
            await this.update();
            break;
          case 'executeCommand':
            await vscode.commands.executeCommand(message.commandId);
            break;
          case 'openStatistics':
            await vscode.commands.executeCommand('ung.showStatistics');
            break;
          case 'startTracking':
            await vscode.commands.executeCommand('ung.startTracking');
            break;
          case 'stopTracking':
            await vscode.commands.executeCommand('ung.stopTracking');
            break;
          case 'createInvoice':
            await vscode.commands.executeCommand('ung.createInvoice');
            break;
          case 'viewInvoices':
            await vscode.commands.executeCommand('ung.viewInvoices');
            break;
          case 'resetDatabase':
            await this.handleDatabaseReset();
            break;
        }
      },
      null,
      this.disposables
    );
  }

  private async handleDatabaseReset() {
    // First confirmation
    const confirm1 = await vscode.window.showWarningMessage(
      '⚠️ Are you sure you want to reset the database? This will DELETE ALL DATA!',
      { modal: true },
      'Yes, I understand'
    );

    if (confirm1 !== 'Yes, I understand') {
      vscode.window.showInformationMessage('Database reset cancelled.');
      return;
    }

    // Second confirmation - require typing RESET
    const confirm2 = await vscode.window.showInputBox({
      prompt: 'Type RESET to confirm database deletion',
      placeHolder: 'RESET',
      validateInput: (value) => {
        if (value !== 'RESET') {
          return 'Please type RESET exactly to confirm';
        }
        return null;
      },
    });

    if (confirm2 !== 'RESET') {
      vscode.window.showInformationMessage('Database reset cancelled.');
      return;
    }

    // Execute reset in terminal (the CLI command handles the actual reset)
    const terminal = vscode.window.createTerminal('UNG Database Reset');
    terminal.show();
    terminal.sendText('echo "Resetting database..." && ung database reset');

    vscode.window.showInformationMessage(
      'Database reset initiated. Check the terminal for progress.'
    );
  }

  private async update() {
    const data = await this.gatherDashboardData();
    this.panel.webview.html = this.getHtmlContent(data);
  }

  private async gatherDashboardData(): Promise<DashboardData> {
    try {
      const [
        invoicesResult,
        clientsResult,
        contractsResult,
        currentSessionResult,
        expensesResult,
      ] = await Promise.all([
        this.cli.listInvoices(),
        this.cli.listClients(),
        this.cli.listContracts(),
        this.cli.getCurrentSession(),
        this.cli.listExpenses(),
      ]);

      // Parse data
      const invoices = this.parseInvoices(invoicesResult.stdout || '');
      const clients = this.parseClients(clientsResult.stdout || '');
      const contracts = this.parseContracts(contractsResult.stdout || '');
      const currentSession = this.parseCurrentSession(
        currentSessionResult.stdout || ''
      );
      const expenses = this.parseExpenses(expensesResult.stdout || '');

      // Calculate metrics
      const now = new Date();
      const thisMonth = invoices.filter((i) => {
        const date = new Date(i.date);
        return (
          date.getMonth() === now.getMonth() &&
          date.getFullYear() === now.getFullYear()
        );
      });

      const monthlyRevenue = thisMonth.reduce((sum, i) => sum + i.amount, 0);
      const paidThisMonth = thisMonth
        .filter((i) => i.status === 'paid')
        .reduce((sum, i) => sum + i.amount, 0);
      const pendingAmount = invoices
        .filter((i) => i.status === 'pending' || i.status === 'sent')
        .reduce((sum, i) => sum + i.amount, 0);
      const overdueAmount = invoices
        .filter((i) => i.status === 'overdue')
        .reduce((sum, i) => sum + i.amount, 0);

      // Recent invoices (last 5)
      const recentInvoices = invoices.slice(0, 5);

      // Overdue invoices
      const overdueInvoices = invoices
        .filter((i) => i.status === 'overdue')
        .slice(0, 5);

      return {
        metrics: {
          totalClients: clients.length,
          activeContracts: contracts.filter((c) => c.active).length,
          monthlyRevenue,
          paidThisMonth,
          pendingAmount,
          overdueAmount,
          totalInvoices: invoices.length,
          totalExpenses: expenses.reduce((sum, e) => sum + e.amount, 0),
        },
        currentSession,
        recentInvoices,
        overdueInvoices,
        hasActiveSession: currentSession !== null,
      };
    } catch (_error) {
      return {
        metrics: {
          totalClients: 0,
          activeContracts: 0,
          monthlyRevenue: 0,
          paidThisMonth: 0,
          pendingAmount: 0,
          overdueAmount: 0,
          totalInvoices: 0,
          totalExpenses: 0,
        },
        currentSession: null,
        recentInvoices: [],
        overdueInvoices: [],
        hasActiveSession: false,
      };
    }
  }

  private parseInvoices(output: string): Array<{
    id: number;
    client: string;
    date: string;
    amount: number;
    status: string;
  }> {
    const lines = output.trim().split('\n');
    if (lines.length < 2) return [];

    return lines
      .slice(1)
      .map((line) => {
        const parts = line.split(/\s{2,}/).filter((p) => p.trim());
        const amountStr = parts[3]?.replace(/[^0-9.]/g, '') || '0';
        return {
          id: parseInt(parts[0], 10),
          client: parts[1] || 'Unknown',
          date: parts[2] || new Date().toISOString().split('T')[0],
          amount: parseFloat(amountStr),
          status: (parts[4] || 'draft').toLowerCase(),
        };
      })
      .filter((i) => !Number.isNaN(i.id));
  }

  private parseClients(output: string): Array<{ id: number; name: string }> {
    const lines = output.trim().split('\n');
    if (lines.length < 2) return [];

    return lines
      .slice(1)
      .map((line) => {
        const parts = line.split(/\s{2,}/).filter((p) => p.trim());
        return {
          id: parseInt(parts[0], 10),
          name: parts[1] || 'Unknown',
        };
      })
      .filter((c) => !Number.isNaN(c.id));
  }

  private parseContracts(
    output: string
  ): Array<{ id: number; active: boolean }> {
    const lines = output.trim().split('\n');
    if (lines.length < 2) return [];

    return lines
      .slice(1)
      .map((line) => {
        const parts = line.split(/\s{2,}/).filter((p) => p.trim());
        return {
          id: parseInt(parts[0], 10),
          active: line.includes('✓'),
        };
      })
      .filter((c) => !Number.isNaN(c.id));
  }

  private parseCurrentSession(
    output: string
  ): { project: string; client: string; elapsed: string } | null {
    if (output.includes('No active') || !output.trim()) {
      return null;
    }

    const projectMatch = output.match(/Project:\s*(.+)/);
    const clientMatch = output.match(/Client:\s*(.+)/);
    const elapsedMatch = output.match(/Elapsed:\s*(.+)/);

    return {
      project: projectMatch?.[1]?.trim() || 'No project',
      client: clientMatch?.[1]?.trim() || 'No client',
      elapsed: elapsedMatch?.[1]?.trim() || '0h 0m',
    };
  }

  private parseExpenses(output: string): Array<{ amount: number }> {
    const lines = output.trim().split('\n');
    if (lines.length < 2) return [];

    return lines
      .slice(1)
      .map((line) => {
        const parts = line.split(/\s{2,}/).filter((p) => p.trim());
        const amountStr = parts[2]?.replace(/[^0-9.]/g, '') || '0';
        return { amount: parseFloat(amountStr) };
      })
      .filter((e) => !Number.isNaN(e.amount));
  }

  private getHtmlContent(data: DashboardData): string {
    const formatCurrency = (value: number) =>
      `$${value.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

    const getStatusBadge = (status: string) => {
      const colors: Record<string, string> = {
        paid: '#4caf50',
        pending: '#ff9800',
        sent: '#2196f3',
        overdue: '#f44336',
        draft: '#9e9e9e',
      };
      const color = colors[status] || '#9e9e9e';
      return `<span class="badge" style="background: ${color}">${status}</span>`;
    };

    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>UNG Dashboard</title>
    <style>
        /* ==============================================
           UNG Design System - Main Dashboard Panel
           Aligned with macOS DesignTokens.swift
           ============================================== */

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

            /* VSCode Theme Integration */
            --bg-primary: var(--vscode-editor-background);
            --bg-secondary: var(--vscode-sideBar-background);
            --bg-tertiary: var(--vscode-input-background);
            --text-primary: var(--vscode-editor-foreground);
            --text-secondary: var(--vscode-descriptionForeground);
            --border: var(--vscode-panel-border);

            /* Spacing - 8pt grid system */
            --space-xxs: 4px;
            --space-xs: 8px;
            --space-sm: 12px;
            --space-md: 16px;
            --space-lg: 24px;
            --space-xl: 32px;

            /* Border Radius */
            --radius-xs: 4px;
            --radius-sm: 8px;
            --radius-md: 12px;
            --radius-lg: 16px;
            --radius-full: 9999px;

            /* Transitions - Snappy micro interactions */
            --transition-micro: 0.1s cubic-bezier(0.4, 0, 0.2, 1);
            --transition-quick: 0.15s cubic-bezier(0.4, 0, 0.2, 1);
            --transition-standard: 0.25s cubic-bezier(0.4, 0, 0.2, 1);
            --transition-bounce: 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);
        }

        * { box-sizing: border-box; margin: 0; padding: 0; }

        body {
            font-family: var(--vscode-font-family);
            background: var(--bg-primary);
            color: var(--text-primary);
            padding: var(--space-lg);
            line-height: 1.5;
            font-size: 13px;
        }

        .dashboard-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: var(--space-xl);
            padding-bottom: var(--space-md);
            border-bottom: 1px solid var(--border);
        }

        .logo {
            display: flex;
            align-items: center;
            gap: var(--space-sm);
        }

        .logo-icon {
            width: 40px;
            height: 40px;
            background: linear-gradient(135deg, var(--ung-brand), var(--ung-brand-dark));
            border-radius: var(--radius-md);
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 18px;
            font-weight: 700;
            color: white;
            transition: transform var(--transition-bounce);
        }

        .logo-icon:hover {
            transform: scale(1.05) rotate(-3deg);
        }

        .logo h1 {
            font-size: 22px;
            font-weight: 600;
        }

        .logo span {
            font-size: 12px;
            color: var(--text-secondary);
            font-weight: normal;
        }

        .header-actions {
            display: flex;
            gap: var(--space-xs);
        }

        button {
            background: var(--ung-brand);
            color: white;
            border: none;
            padding: var(--space-sm) var(--space-md);
            border-radius: var(--radius-sm);
            cursor: pointer;
            font-size: 13px;
            font-weight: 500;
            transition: all var(--transition-quick);
            display: flex;
            align-items: center;
            gap: var(--space-xs);
        }

        button:hover {
            background: var(--ung-brand-dark);
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(51, 115, 232, 0.25);
        }

        button:active {
            transform: translateY(0) scale(0.98);
        }

        button.secondary {
            background: var(--bg-secondary);
            border: 1px solid var(--border);
            color: var(--text-primary);
        }

        button.secondary:hover {
            background: var(--bg-tertiary);
            border-color: var(--ung-brand);
            box-shadow: none;
        }

        button.danger {
            background: var(--ung-error);
        }

        button.danger:hover {
            background: #D14545;
            box-shadow: 0 4px 12px rgba(230, 90, 90, 0.25);
        }

        /* Active Tracking Banner */
        .tracking-banner {
            background: linear-gradient(135deg, var(--ung-success), #2A8F4A);
            border-radius: var(--radius-md);
            padding: var(--space-lg);
            margin-bottom: var(--space-lg);
            display: flex;
            justify-content: space-between;
            align-items: center;
            color: white;
            position: relative;
            overflow: hidden;
        }

        .tracking-banner::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: linear-gradient(90deg, transparent, rgba(255,255,255,0.1), transparent);
            animation: shimmer 2s infinite;
        }

        @keyframes shimmer {
            0% { transform: translateX(-100%); }
            100% { transform: translateX(100%); }
        }

        .tracking-banner.inactive {
            background: var(--bg-secondary);
            border: 2px dashed var(--border);
            color: var(--text-primary);
        }

        .tracking-banner.inactive::before {
            display: none;
        }

        .tracking-info {
            position: relative;
            z-index: 1;
        }

        .tracking-info h3 {
            font-size: 13px;
            opacity: 0.9;
            margin-bottom: var(--space-xxs);
            font-weight: 500;
        }

        .tracking-info .time {
            font-size: 32px;
            font-weight: 700;
            font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, monospace;
            letter-spacing: -1px;
        }

        .tracking-info .details {
            font-size: 13px;
            opacity: 0.85;
            margin-top: var(--space-xxs);
        }

        /* Metrics Grid */
        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
            gap: var(--space-md);
            margin-bottom: var(--space-xl);
        }

        .metric-card {
            background: var(--bg-secondary);
            border: 1px solid var(--border);
            border-radius: var(--radius-md);
            padding: var(--space-lg);
            transition: all var(--transition-quick);
            position: relative;
            overflow: hidden;
        }

        .metric-card::after {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            width: 3px;
            height: 100%;
            background: var(--ung-brand);
            opacity: 0;
            transition: opacity var(--transition-micro);
        }

        .metric-card:hover {
            border-color: var(--ung-brand);
            transform: translateY(-2px);
            box-shadow: 0 8px 24px rgba(0,0,0,0.1);
        }

        .metric-card:hover::after {
            opacity: 1;
        }

        .metric-card.clickable {
            cursor: pointer;
        }

        .metric-card:active {
            transform: translateY(0) scale(0.99);
        }

        .metric-icon {
            width: 40px;
            height: 40px;
            border-radius: var(--radius-sm);
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 18px;
            margin-bottom: var(--space-sm);
            transition: transform var(--transition-bounce);
        }

        .metric-card:hover .metric-icon {
            transform: scale(1.1);
        }

        .metric-icon.green { background: var(--ung-success-light); color: var(--ung-success); }
        .metric-icon.blue { background: var(--ung-info-light); color: var(--ung-info); }
        .metric-icon.orange { background: var(--ung-warning-light); color: var(--ung-warning); }
        .metric-icon.red { background: var(--ung-error-light); color: var(--ung-error); }
        .metric-icon.purple { background: var(--ung-purple-light); color: var(--ung-purple); }

        .metric-label {
            font-size: 11px;
            color: var(--text-secondary);
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-bottom: var(--space-xxs);
            font-weight: 500;
        }

        .metric-value {
            font-size: 26px;
            font-weight: 700;
            transition: color var(--transition-micro);
        }

        .metric-card:hover .metric-value {
            color: var(--ung-brand);
        }

        .metric-subtitle {
            font-size: 12px;
            color: var(--text-secondary);
            margin-top: var(--space-xxs);
        }

        /* Content Grid */
        .content-grid {
            display: grid;
            grid-template-columns: 2fr 1fr;
            gap: var(--space-lg);
        }

        @media (max-width: 900px) {
            .content-grid { grid-template-columns: 1fr; }
        }

        .card {
            background: var(--bg-secondary);
            border: 1px solid var(--border);
            border-radius: var(--radius-md);
            overflow: hidden;
            transition: border-color var(--transition-quick);
        }

        .card:hover {
            border-color: rgba(51, 115, 232, 0.3);
        }

        .card-header {
            padding: var(--space-md) var(--space-lg);
            border-bottom: 1px solid var(--border);
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .card-header h3 {
            font-size: 14px;
            font-weight: 600;
        }

        .card-body {
            padding: var(--space-md) var(--space-lg);
        }

        /* Invoice Items */
        .invoice-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: var(--space-sm) 0;
            border-bottom: 1px solid var(--border);
            transition: all var(--transition-quick);
            margin: 0 calc(-1 * var(--space-lg));
            padding-left: var(--space-lg);
            padding-right: var(--space-lg);
        }

        .invoice-item:hover {
            background: var(--bg-tertiary);
        }

        .invoice-item:last-child {
            border-bottom: none;
        }

        .invoice-info {
            flex: 1;
        }

        .invoice-client {
            font-weight: 500;
            margin-bottom: var(--space-xxs);
            transition: color var(--transition-micro);
        }

        .invoice-item:hover .invoice-client {
            color: var(--ung-brand);
        }

        .invoice-date {
            font-size: 12px;
            color: var(--text-secondary);
        }

        .invoice-amount {
            font-weight: 600;
            margin-right: var(--space-sm);
            font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, monospace;
        }

        .badge {
            padding: var(--space-xxs) var(--space-sm);
            border-radius: var(--radius-xs);
            font-size: 10px;
            font-weight: 600;
            color: white;
            text-transform: uppercase;
            letter-spacing: 0.3px;
        }

        /* Quick Actions */
        .quick-actions {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: var(--space-sm);
        }

        .action-btn {
            background: var(--bg-tertiary);
            border: 1px solid var(--border);
            border-radius: var(--radius-sm);
            padding: var(--space-md);
            text-align: center;
            cursor: pointer;
            transition: all var(--transition-quick);
            position: relative;
            overflow: hidden;
        }

        .action-btn::before {
            content: '';
            position: absolute;
            top: 50%;
            left: 50%;
            width: 0;
            height: 0;
            background: rgba(51, 115, 232, 0.1);
            border-radius: 50%;
            transition: all var(--transition-standard);
            transform: translate(-50%, -50%);
        }

        .action-btn:hover {
            background: var(--ung-brand);
            border-color: var(--ung-brand);
            color: white;
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(51, 115, 232, 0.25);
        }

        .action-btn:hover::before {
            width: 200%;
            height: 200%;
        }

        .action-btn:active {
            transform: translateY(0) scale(0.98);
        }

        .action-btn:hover .action-icon {
            background: rgba(255,255,255,0.2);
            transform: scale(1.1);
        }

        .action-icon {
            width: 36px;
            height: 36px;
            margin: 0 auto var(--space-xs);
            background: var(--bg-secondary);
            border-radius: var(--radius-sm);
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 16px;
            transition: all var(--transition-bounce);
            position: relative;
            z-index: 1;
        }

        .action-label {
            font-size: 12px;
            font-weight: 500;
            position: relative;
            z-index: 1;
        }

        /* Empty State */
        .empty-state {
            text-align: center;
            padding: var(--space-xl);
            color: var(--text-secondary);
        }

        .empty-state .icon {
            font-size: 48px;
            margin-bottom: var(--space-sm);
            opacity: 0.5;
        }

        /* Danger Zone */
        .danger-zone {
            margin-top: var(--space-xl);
            padding: var(--space-lg);
            background: var(--ung-error-light);
            border: 1px solid rgba(230, 90, 90, 0.3);
            border-radius: var(--radius-md);
            transition: border-color var(--transition-quick);
        }

        .danger-zone:hover {
            border-color: var(--ung-error);
        }

        .danger-zone h3 {
            color: var(--ung-error);
            font-size: 14px;
            margin-bottom: var(--space-xs);
        }

        .danger-zone p {
            font-size: 12px;
            color: var(--text-secondary);
            margin-bottom: var(--space-sm);
        }

        /* Footer */
        .footer {
            margin-top: var(--space-xl);
            text-align: center;
            color: var(--text-secondary);
            font-size: 12px;
        }

        .footer a {
            color: var(--ung-brand);
            text-decoration: none;
            transition: color var(--transition-micro);
        }

        .footer a:hover {
            color: var(--ung-brand-dark);
            text-decoration: underline;
        }

        /* Focus States for Accessibility */
        button:focus-visible,
        .action-btn:focus-visible,
        .metric-card:focus-visible {
            outline: 2px solid var(--ung-brand);
            outline-offset: 2px;
        }

        /* Reduced Motion Preference */
        @media (prefers-reduced-motion: reduce) {
            *,
            *::before,
            *::after {
                animation-duration: 0.01ms !important;
                animation-iteration-count: 1 !important;
                transition-duration: 0.01ms !important;
            }
        }
    </style>
</head>
<body>
    <div class="dashboard-header">
        <div class="logo">
            <div class="logo-icon">U</div>
            <div>
                <h1>UNG Dashboard</h1>
                <span>Business Management</span>
            </div>
        </div>
        <div class="header-actions">
            <button class="secondary" onclick="refresh()">↻ Refresh</button>
            <button onclick="openStatistics()">📊 Statistics</button>
        </div>
    </div>

    <!-- Time Tracking Banner -->
    <div class="tracking-banner ${data.hasActiveSession ? '' : 'inactive'}">
        ${
          data.hasActiveSession && data.currentSession
            ? `
            <div class="tracking-info">
                <h3>⏱️ Currently Tracking</h3>
                <div class="time">${data.currentSession.elapsed}</div>
                <div class="details">${data.currentSession.project} • ${data.currentSession.client}</div>
            </div>
            <button onclick="stopTracking()" style="background: white; color: #2e7d32;">⏹ Stop Tracking</button>
        `
            : `
            <div class="tracking-info">
                <h3>No Active Session</h3>
                <div class="details">Start tracking your work time</div>
            </div>
            <button onclick="startTracking()">▶ Start Tracking</button>
        `
        }
    </div>

    <!-- Metrics Grid -->
    <div class="metrics-grid">
        <div class="metric-card clickable" onclick="viewInvoices()">
            <div class="metric-icon green">💰</div>
            <div class="metric-label">Monthly Revenue</div>
            <div class="metric-value">${formatCurrency(data.metrics.monthlyRevenue)}</div>
            <div class="metric-subtitle">${formatCurrency(data.metrics.paidThisMonth)} paid</div>
        </div>
        <div class="metric-card clickable" onclick="viewInvoices()">
            <div class="metric-icon orange">⏳</div>
            <div class="metric-label">Pending</div>
            <div class="metric-value">${formatCurrency(data.metrics.pendingAmount)}</div>
            <div class="metric-subtitle">Awaiting payment</div>
        </div>
        <div class="metric-card ${data.metrics.overdueAmount > 0 ? 'clickable' : ''}" ${data.metrics.overdueAmount > 0 ? 'onclick="viewInvoices()"' : ''}>
            <div class="metric-icon red">⚠️</div>
            <div class="metric-label">Overdue</div>
            <div class="metric-value">${formatCurrency(data.metrics.overdueAmount)}</div>
            <div class="metric-subtitle">${data.overdueInvoices.length} invoices</div>
        </div>
        <div class="metric-card">
            <div class="metric-icon blue">👥</div>
            <div class="metric-label">Clients</div>
            <div class="metric-value">${data.metrics.totalClients}</div>
            <div class="metric-subtitle">${data.metrics.activeContracts} active contracts</div>
        </div>
    </div>

    <!-- Content Grid -->
    <div class="content-grid">
        <!-- Recent Invoices -->
        <div class="card">
            <div class="card-header">
                <h3>Recent Invoices</h3>
                <button class="secondary" onclick="viewInvoices()">View All</button>
            </div>
            <div class="card-body">
                ${
                  data.recentInvoices.length > 0
                    ? data.recentInvoices
                        .map(
                          (inv) => `
                    <div class="invoice-item">
                        <div class="invoice-info">
                            <div class="invoice-client">${inv.client}</div>
                            <div class="invoice-date">${inv.date}</div>
                        </div>
                        <div class="invoice-amount">${formatCurrency(inv.amount)}</div>
                        ${getStatusBadge(inv.status)}
                    </div>
                `
                        )
                        .join('')
                    : `
                    <div class="empty-state">
                        <div class="icon">📄</div>
                        <p>No invoices yet</p>
                    </div>
                `
                }
            </div>
        </div>

        <!-- Quick Actions -->
        <div class="card">
            <div class="card-header">
                <h3>Quick Actions</h3>
            </div>
            <div class="card-body">
                <div class="quick-actions">
                    <div class="action-btn" onclick="createInvoice()">
                        <div class="action-icon">📝</div>
                        <div class="action-label">New Invoice</div>
                    </div>
                    <div class="action-btn" onclick="startTracking()">
                        <div class="action-icon">⏱️</div>
                        <div class="action-label">Track Time</div>
                    </div>
                    <div class="action-btn" onclick="executeCommand('ung.addExpense')">
                        <div class="action-icon">💸</div>
                        <div class="action-label">Add Expense</div>
                    </div>
                    <div class="action-btn" onclick="executeCommand('ung.addClient')">
                        <div class="action-icon">👤</div>
                        <div class="action-label">Add Client</div>
                    </div>
                    <div class="action-btn" onclick="openStatistics()">
                        <div class="action-icon">📊</div>
                        <div class="action-label">Reports</div>
                    </div>
                    <div class="action-btn" onclick="executeCommand('ung.createBackup')">
                        <div class="action-icon">💾</div>
                        <div class="action-label">Backup</div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    ${
      data.overdueInvoices.length > 0
        ? `
        <!-- Overdue Alert -->
        <div class="danger-zone" style="margin-top: 24px; background: rgba(244, 67, 54, 0.1);">
            <h3>⚠️ Overdue Invoices (${data.overdueInvoices.length})</h3>
            <p>These invoices are past their due date and require attention.</p>
            ${data.overdueInvoices
              .map(
                (inv) => `
                <div class="invoice-item" style="padding: 8px 0;">
                    <div class="invoice-info">
                        <div class="invoice-client">${inv.client}</div>
                        <div class="invoice-date">Due: ${inv.date}</div>
                    </div>
                    <div class="invoice-amount" style="color: var(--danger);">${formatCurrency(inv.amount)}</div>
                </div>
            `
              )
              .join('')}
        </div>
    `
        : ''
    }

    <!-- Danger Zone -->
    <div class="danger-zone">
        <h3>⚠️ Danger Zone</h3>
        <p>Reset the database to delete all data. This action cannot be undone.</p>
        <button class="danger" onclick="resetDatabase()">🗑️ Reset Database</button>
    </div>

    <div class="footer">
        <p>UNG Business Management • <a href="#" onclick="executeCommand('ung.openDocs')">Documentation</a></p>
    </div>

    <script>
        const vscode = acquireVsCodeApi();

        function refresh() {
            vscode.postMessage({ command: 'refresh' });
        }

        function openStatistics() {
            vscode.postMessage({ command: 'openStatistics' });
        }

        function startTracking() {
            vscode.postMessage({ command: 'startTracking' });
        }

        function stopTracking() {
            vscode.postMessage({ command: 'stopTracking' });
        }

        function createInvoice() {
            vscode.postMessage({ command: 'createInvoice' });
        }

        function viewInvoices() {
            vscode.postMessage({ command: 'viewInvoices' });
        }

        function resetDatabase() {
            vscode.postMessage({ command: 'resetDatabase' });
        }

        function executeCommand(commandId) {
            vscode.postMessage({ command: 'executeCommand', commandId });
        }
    </script>
</body>
</html>`;
  }

  public dispose() {
    MainDashboardPanel.currentPanel = undefined;

    this.panel.dispose();

    while (this.disposables.length) {
      const x = this.disposables.pop();
      if (x) {
        x.dispose();
      }
    }
  }
}

interface DashboardData {
  metrics: {
    totalClients: number;
    activeContracts: number;
    monthlyRevenue: number;
    paidThisMonth: number;
    pendingAmount: number;
    overdueAmount: number;
    totalInvoices: number;
    totalExpenses: number;
  };
  currentSession: { project: string; client: string; elapsed: string } | null;
  recentInvoices: Array<{
    id: number;
    client: string;
    date: string;
    amount: number;
    status: string;
  }>;
  overdueInvoices: Array<{
    id: number;
    client: string;
    date: string;
    amount: number;
    status: string;
  }>;
  hasActiveSession: boolean;
}
