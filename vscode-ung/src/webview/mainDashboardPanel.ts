import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';

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
                retainContextWhenHidden: true
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
            async message => {
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
            }
        });

        if (confirm2 !== 'RESET') {
            vscode.window.showInformationMessage('Database reset cancelled.');
            return;
        }

        // Execute reset in terminal (the CLI command handles the actual reset)
        const terminal = vscode.window.createTerminal('UNG Database Reset');
        terminal.show();
        terminal.sendText('echo "Resetting database..." && ung database reset');

        vscode.window.showInformationMessage('Database reset initiated. Check the terminal for progress.');
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
                expensesResult
            ] = await Promise.all([
                this.cli.listInvoices(),
                this.cli.listClients(),
                this.cli.listContracts(),
                this.cli.getCurrentSession(),
                this.cli.listExpenses()
            ]);

            // Parse data
            const invoices = this.parseInvoices(invoicesResult.stdout || '');
            const clients = this.parseClients(clientsResult.stdout || '');
            const contracts = this.parseContracts(contractsResult.stdout || '');
            const currentSession = this.parseCurrentSession(currentSessionResult.stdout || '');
            const expenses = this.parseExpenses(expensesResult.stdout || '');

            // Calculate metrics
            const now = new Date();
            const thisMonth = invoices.filter(i => {
                const date = new Date(i.date);
                return date.getMonth() === now.getMonth() && date.getFullYear() === now.getFullYear();
            });

            const monthlyRevenue = thisMonth.reduce((sum, i) => sum + i.amount, 0);
            const paidThisMonth = thisMonth.filter(i => i.status === 'paid').reduce((sum, i) => sum + i.amount, 0);
            const pendingAmount = invoices.filter(i => i.status === 'pending' || i.status === 'sent').reduce((sum, i) => sum + i.amount, 0);
            const overdueAmount = invoices.filter(i => i.status === 'overdue').reduce((sum, i) => sum + i.amount, 0);

            // Recent invoices (last 5)
            const recentInvoices = invoices.slice(0, 5);

            // Overdue invoices
            const overdueInvoices = invoices.filter(i => i.status === 'overdue').slice(0, 5);

            return {
                metrics: {
                    totalClients: clients.length,
                    activeContracts: contracts.filter(c => c.active).length,
                    monthlyRevenue,
                    paidThisMonth,
                    pendingAmount,
                    overdueAmount,
                    totalInvoices: invoices.length,
                    totalExpenses: expenses.reduce((sum, e) => sum + e.amount, 0)
                },
                currentSession,
                recentInvoices,
                overdueInvoices,
                hasActiveSession: currentSession !== null
            };
        } catch (error) {
            return {
                metrics: {
                    totalClients: 0,
                    activeContracts: 0,
                    monthlyRevenue: 0,
                    paidThisMonth: 0,
                    pendingAmount: 0,
                    overdueAmount: 0,
                    totalInvoices: 0,
                    totalExpenses: 0
                },
                currentSession: null,
                recentInvoices: [],
                overdueInvoices: [],
                hasActiveSession: false
            };
        }
    }

    private parseInvoices(output: string): Array<{ id: number; client: string; date: string; amount: number; status: string }> {
        const lines = output.trim().split('\n');
        if (lines.length < 2) return [];

        return lines.slice(1).map(line => {
            const parts = line.split(/\s{2,}/).filter(p => p.trim());
            const amountStr = parts[3]?.replace(/[^0-9.]/g, '') || '0';
            return {
                id: parseInt(parts[0], 10),
                client: parts[1] || 'Unknown',
                date: parts[2] || new Date().toISOString().split('T')[0],
                amount: parseFloat(amountStr),
                status: (parts[4] || 'draft').toLowerCase()
            };
        }).filter(i => !isNaN(i.id));
    }

    private parseClients(output: string): Array<{ id: number; name: string }> {
        const lines = output.trim().split('\n');
        if (lines.length < 2) return [];

        return lines.slice(1).map(line => {
            const parts = line.split(/\s{2,}/).filter(p => p.trim());
            return {
                id: parseInt(parts[0], 10),
                name: parts[1] || 'Unknown'
            };
        }).filter(c => !isNaN(c.id));
    }

    private parseContracts(output: string): Array<{ id: number; active: boolean }> {
        const lines = output.trim().split('\n');
        if (lines.length < 2) return [];

        return lines.slice(1).map(line => {
            const parts = line.split(/\s{2,}/).filter(p => p.trim());
            return {
                id: parseInt(parts[0], 10),
                active: line.includes('✓')
            };
        }).filter(c => !isNaN(c.id));
    }

    private parseCurrentSession(output: string): { project: string; client: string; elapsed: string } | null {
        if (output.includes('No active') || !output.trim()) {
            return null;
        }

        const projectMatch = output.match(/Project:\s*(.+)/);
        const clientMatch = output.match(/Client:\s*(.+)/);
        const elapsedMatch = output.match(/Elapsed:\s*(.+)/);

        return {
            project: projectMatch?.[1]?.trim() || 'No project',
            client: clientMatch?.[1]?.trim() || 'No client',
            elapsed: elapsedMatch?.[1]?.trim() || '0h 0m'
        };
    }

    private parseExpenses(output: string): Array<{ amount: number }> {
        const lines = output.trim().split('\n');
        if (lines.length < 2) return [];

        return lines.slice(1).map(line => {
            const parts = line.split(/\s{2,}/).filter(p => p.trim());
            const amountStr = parts[2]?.replace(/[^0-9.]/g, '') || '0';
            return { amount: parseFloat(amountStr) };
        }).filter(e => !isNaN(e.amount));
    }

    private getHtmlContent(data: DashboardData): string {
        const formatCurrency = (value: number) => `$${value.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

        const getStatusBadge = (status: string) => {
            const colors: Record<string, string> = {
                paid: '#4caf50',
                pending: '#ff9800',
                sent: '#2196f3',
                overdue: '#f44336',
                draft: '#9e9e9e'
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
        :root {
            --bg-primary: var(--vscode-editor-background);
            --bg-secondary: var(--vscode-sideBar-background);
            --bg-tertiary: var(--vscode-input-background);
            --text-primary: var(--vscode-editor-foreground);
            --text-secondary: var(--vscode-descriptionForeground);
            --border: var(--vscode-panel-border);
            --accent: var(--vscode-button-background);
            --accent-hover: var(--vscode-button-hoverBackground);
            --success: #4caf50;
            --warning: #ff9800;
            --danger: #f44336;
            --info: #2196f3;
        }

        * { box-sizing: border-box; margin: 0; padding: 0; }

        body {
            font-family: var(--vscode-font-family);
            background: var(--bg-primary);
            color: var(--text-primary);
            padding: 24px;
            line-height: 1.6;
        }

        .dashboard-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 32px;
            padding-bottom: 16px;
            border-bottom: 1px solid var(--border);
        }

        .logo {
            display: flex;
            align-items: center;
            gap: 12px;
        }

        .logo-icon {
            width: 40px;
            height: 40px;
            background: linear-gradient(135deg, var(--accent), var(--info));
            border-radius: 10px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 20px;
            font-weight: bold;
            color: white;
        }

        .logo h1 {
            font-size: 24px;
            font-weight: 600;
        }

        .logo span {
            font-size: 12px;
            color: var(--text-secondary);
            font-weight: normal;
        }

        .header-actions {
            display: flex;
            gap: 8px;
        }

        button {
            background: var(--accent);
            color: var(--vscode-button-foreground);
            border: none;
            padding: 10px 20px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 13px;
            font-weight: 500;
            transition: all 0.2s;
            display: flex;
            align-items: center;
            gap: 8px;
        }

        button:hover {
            background: var(--accent-hover);
            transform: translateY(-1px);
        }

        button.secondary {
            background: var(--bg-secondary);
            border: 1px solid var(--border);
            color: var(--text-primary);
        }

        button.secondary:hover {
            background: var(--bg-tertiary);
        }

        button.danger {
            background: var(--danger);
        }

        button.danger:hover {
            background: #d32f2f;
        }

        .tracking-banner {
            background: linear-gradient(135deg, var(--success), #2e7d32);
            border-radius: 12px;
            padding: 20px 24px;
            margin-bottom: 24px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            color: white;
        }

        .tracking-banner.inactive {
            background: var(--bg-secondary);
            border: 2px dashed var(--border);
            color: var(--text-primary);
        }

        .tracking-info h3 {
            font-size: 14px;
            opacity: 0.9;
            margin-bottom: 4px;
        }

        .tracking-info .time {
            font-size: 28px;
            font-weight: 700;
        }

        .tracking-info .details {
            font-size: 13px;
            opacity: 0.8;
            margin-top: 4px;
        }

        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 16px;
            margin-bottom: 32px;
        }

        .metric-card {
            background: var(--bg-secondary);
            border: 1px solid var(--border);
            border-radius: 12px;
            padding: 20px;
            transition: all 0.2s;
        }

        .metric-card:hover {
            border-color: var(--accent);
            transform: translateY(-2px);
        }

        .metric-card.clickable {
            cursor: pointer;
        }

        .metric-icon {
            width: 40px;
            height: 40px;
            border-radius: 10px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 18px;
            margin-bottom: 12px;
        }

        .metric-icon.green { background: rgba(76, 175, 80, 0.15); color: var(--success); }
        .metric-icon.blue { background: rgba(33, 150, 243, 0.15); color: var(--info); }
        .metric-icon.orange { background: rgba(255, 152, 0, 0.15); color: var(--warning); }
        .metric-icon.red { background: rgba(244, 67, 54, 0.15); color: var(--danger); }
        .metric-icon.purple { background: rgba(156, 39, 176, 0.15); color: #9c27b0; }

        .metric-label {
            font-size: 12px;
            color: var(--text-secondary);
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-bottom: 4px;
        }

        .metric-value {
            font-size: 24px;
            font-weight: 700;
        }

        .metric-subtitle {
            font-size: 12px;
            color: var(--text-secondary);
            margin-top: 4px;
        }

        .content-grid {
            display: grid;
            grid-template-columns: 2fr 1fr;
            gap: 24px;
        }

        @media (max-width: 900px) {
            .content-grid { grid-template-columns: 1fr; }
        }

        .card {
            background: var(--bg-secondary);
            border: 1px solid var(--border);
            border-radius: 12px;
            overflow: hidden;
        }

        .card-header {
            padding: 16px 20px;
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
            padding: 16px 20px;
        }

        .invoice-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 12px 0;
            border-bottom: 1px solid var(--border);
        }

        .invoice-item:last-child {
            border-bottom: none;
        }

        .invoice-info {
            flex: 1;
        }

        .invoice-client {
            font-weight: 500;
            margin-bottom: 4px;
        }

        .invoice-date {
            font-size: 12px;
            color: var(--text-secondary);
        }

        .invoice-amount {
            font-weight: 600;
            margin-right: 12px;
        }

        .badge {
            padding: 4px 10px;
            border-radius: 12px;
            font-size: 11px;
            font-weight: 500;
            color: white;
            text-transform: uppercase;
        }

        .quick-actions {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 12px;
        }

        .action-btn {
            background: var(--bg-tertiary);
            border: 1px solid var(--border);
            border-radius: 8px;
            padding: 16px;
            text-align: center;
            cursor: pointer;
            transition: all 0.2s;
        }

        .action-btn:hover {
            background: var(--accent);
            border-color: var(--accent);
            color: white;
        }

        .action-btn:hover .action-icon {
            background: rgba(255,255,255,0.2);
        }

        .action-icon {
            width: 36px;
            height: 36px;
            margin: 0 auto 8px;
            background: var(--bg-secondary);
            border-radius: 8px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 16px;
        }

        .action-label {
            font-size: 12px;
            font-weight: 500;
        }

        .empty-state {
            text-align: center;
            padding: 32px;
            color: var(--text-secondary);
        }

        .empty-state .icon {
            font-size: 48px;
            margin-bottom: 12px;
            opacity: 0.5;
        }

        .danger-zone {
            margin-top: 32px;
            padding: 20px;
            background: rgba(244, 67, 54, 0.1);
            border: 1px solid rgba(244, 67, 54, 0.3);
            border-radius: 12px;
        }

        .danger-zone h3 {
            color: var(--danger);
            font-size: 14px;
            margin-bottom: 8px;
        }

        .danger-zone p {
            font-size: 12px;
            color: var(--text-secondary);
            margin-bottom: 12px;
        }

        .footer {
            margin-top: 32px;
            text-align: center;
            color: var(--text-secondary);
            font-size: 12px;
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
        ${data.hasActiveSession && data.currentSession ? `
            <div class="tracking-info">
                <h3>⏱️ Currently Tracking</h3>
                <div class="time">${data.currentSession.elapsed}</div>
                <div class="details">${data.currentSession.project} • ${data.currentSession.client}</div>
            </div>
            <button onclick="stopTracking()" style="background: white; color: #2e7d32;">⏹ Stop Tracking</button>
        ` : `
            <div class="tracking-info">
                <h3>No Active Session</h3>
                <div class="details">Start tracking your work time</div>
            </div>
            <button onclick="startTracking()">▶ Start Tracking</button>
        `}
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
                ${data.recentInvoices.length > 0 ? data.recentInvoices.map(inv => `
                    <div class="invoice-item">
                        <div class="invoice-info">
                            <div class="invoice-client">${inv.client}</div>
                            <div class="invoice-date">${inv.date}</div>
                        </div>
                        <div class="invoice-amount">${formatCurrency(inv.amount)}</div>
                        ${getStatusBadge(inv.status)}
                    </div>
                `).join('') : `
                    <div class="empty-state">
                        <div class="icon">📄</div>
                        <p>No invoices yet</p>
                    </div>
                `}
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

    ${data.overdueInvoices.length > 0 ? `
        <!-- Overdue Alert -->
        <div class="danger-zone" style="margin-top: 24px; background: rgba(244, 67, 54, 0.1);">
            <h3>⚠️ Overdue Invoices (${data.overdueInvoices.length})</h3>
            <p>These invoices are past their due date and require attention.</p>
            ${data.overdueInvoices.map(inv => `
                <div class="invoice-item" style="padding: 8px 0;">
                    <div class="invoice-info">
                        <div class="invoice-client">${inv.client}</div>
                        <div class="invoice-date">Due: ${inv.date}</div>
                    </div>
                    <div class="invoice-amount" style="color: var(--danger);">${formatCurrency(inv.amount)}</div>
                </div>
            `).join('')}
        </div>
    ` : ''}

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
    recentInvoices: Array<{ id: number; client: string; date: string; amount: number; status: string }>;
    overdueInvoices: Array<{ id: number; client: string; date: string; amount: number; status: string }>;
    hasActiveSession: boolean;
}
