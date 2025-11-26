import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';

/**
 * Statistics Panel - Rich webview with charts and analytics
 */
export class StatisticsPanel {
  public static currentPanel: StatisticsPanel | undefined;
  private static readonly viewType = 'ungStatistics';

  private readonly panel: vscode.WebviewPanel;
  private readonly cli: UngCli;
  private disposables: vscode.Disposable[] = [];

  public static createOrShow(cli: UngCli) {
    const column = vscode.window.activeTextEditor
      ? vscode.window.activeTextEditor.viewColumn
      : undefined;

    if (StatisticsPanel.currentPanel) {
      StatisticsPanel.currentPanel.panel.reveal(column);
      StatisticsPanel.currentPanel.update();
      return;
    }

    const panel = vscode.window.createWebviewPanel(
      StatisticsPanel.viewType,
      'UNG Statistics & Reports',
      column || vscode.ViewColumn.One,
      {
        enableScripts: true,
        retainContextWhenHidden: true,
      }
    );

    StatisticsPanel.currentPanel = new StatisticsPanel(panel, cli);
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
          case 'exportCsv':
            await this.exportToCsv(message.type);
            break;
          case 'changePeriod':
            await this.update(message.period);
            break;
        }
      },
      null,
      this.disposables
    );
  }

  private async update(period: string = 'month') {
    const data = await this.gatherStatistics(period);
    this.panel.webview.html = this.getHtmlContent(data, period);
  }

  private async gatherStatistics(period: string): Promise<StatisticsData> {
    const [invoicesResult, trackingResult, contractsResult, expensesResult] =
      await Promise.all([
        this.cli.listInvoices(),
        this.cli.listTimeEntries(),
        this.cli.listContracts(),
        this.cli.listExpenses(),
      ]);

    // Parse invoices
    const invoices = this.parseInvoices(invoicesResult.stdout || '');
    const timeEntries = this.parseTimeEntries(trackingResult.stdout || '');
    const contracts = this.parseContracts(contractsResult.stdout || '');
    const expenses = this.parseExpenses(expensesResult.stdout || '');

    // Calculate statistics
    const now = new Date();
    const periodStart = this.getPeriodStart(now, period);

    const filteredInvoices = invoices.filter(
      (i) => new Date(i.date) >= periodStart
    );
    const filteredTime = timeEntries.filter(
      (t) => new Date(t.date) >= periodStart
    );
    const filteredExpenses = expenses.filter(
      (e) => new Date(e.date) >= periodStart
    );

    // Revenue calculations
    const totalRevenue = filteredInvoices.reduce((sum, i) => sum + i.amount, 0);
    const paidRevenue = filteredInvoices
      .filter((i) => i.status === 'paid')
      .reduce((sum, i) => sum + i.amount, 0);
    const pendingRevenue = filteredInvoices
      .filter((i) => i.status === 'pending' || i.status === 'sent')
      .reduce((sum, i) => sum + i.amount, 0);
    const overdueRevenue = filteredInvoices
      .filter((i) => i.status === 'overdue')
      .reduce((sum, i) => sum + i.amount, 0);

    // Time calculations
    const totalHours = filteredTime.reduce((sum, t) => sum + t.hours, 0);
    const billableHours = filteredTime
      .filter((t) => t.billable)
      .reduce((sum, t) => sum + t.hours, 0);

    // Expense calculations
    const totalExpenses = filteredExpenses.reduce(
      (sum, e) => sum + e.amount,
      0
    );

    // Daily/Weekly breakdown for charts
    const revenueByDay = this.groupByDay(
      filteredInvoices,
      'amount',
      periodStart
    );
    const hoursByDay = this.groupByDay(filteredTime, 'hours', periodStart);
    const revenueByClient = this.groupByField(
      filteredInvoices,
      'client',
      'amount'
    );
    const hoursByClient = this.groupByField(filteredTime, 'client', 'hours');

    // Invoice status breakdown
    const invoicesByStatus = {
      draft: filteredInvoices.filter((i) => i.status === 'draft').length,
      pending: filteredInvoices.filter((i) => i.status === 'pending').length,
      sent: filteredInvoices.filter((i) => i.status === 'sent').length,
      paid: filteredInvoices.filter((i) => i.status === 'paid').length,
      overdue: filteredInvoices.filter((i) => i.status === 'overdue').length,
    };

    return {
      period,
      periodStart: periodStart.toISOString(),
      revenue: {
        total: totalRevenue,
        paid: paidRevenue,
        pending: pendingRevenue,
        overdue: overdueRevenue,
      },
      time: {
        totalHours,
        billableHours,
        avgPerDay: totalHours / this.getDaysDiff(periodStart, now),
      },
      expenses: {
        total: totalExpenses,
      },
      netIncome: paidRevenue - totalExpenses,
      charts: {
        revenueByDay,
        hoursByDay,
        revenueByClient,
        hoursByClient,
        invoicesByStatus,
      },
      activeContracts: contracts.filter((c) => c.active).length,
      totalInvoices: filteredInvoices.length,
      avgInvoiceValue:
        filteredInvoices.length > 0
          ? totalRevenue / filteredInvoices.length
          : 0,
    };
  }

  private parseInvoices(output: string): Array<{
    id: number;
    date: string;
    client: string;
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
          date: parts[2] || new Date().toISOString(),
          client: parts[1] || 'Unknown',
          amount: parseFloat(amountStr),
          status: (parts[4] || 'draft').toLowerCase(),
        };
      })
      .filter((i) => !Number.isNaN(i.id));
  }

  private parseTimeEntries(
    output: string
  ): Array<{ date: string; hours: number; client: string; billable: boolean }> {
    const lines = output.trim().split('\n');
    if (lines.length < 2) return [];

    return lines
      .slice(1)
      .map((line) => {
        const parts = line.split(/\s{2,}/).filter((p) => p.trim());
        const hoursStr = parts[2]?.replace(/[^0-9.]/g, '') || '0';
        return {
          date: parts[1] || new Date().toISOString(),
          hours: parseFloat(hoursStr),
          client: parts[3] || 'Unknown',
          billable: !line.toLowerCase().includes('non-billable'),
        };
      })
      .filter((t) => !Number.isNaN(t.hours));
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

  private parseExpenses(
    output: string
  ): Array<{ date: string; amount: number; category: string }> {
    const lines = output.trim().split('\n');
    if (lines.length < 2) return [];

    return lines
      .slice(1)
      .map((line) => {
        const parts = line.split(/\s{2,}/).filter((p) => p.trim());
        const amountStr = parts[2]?.replace(/[^0-9.]/g, '') || '0';
        return {
          date: parts[1] || new Date().toISOString(),
          amount: parseFloat(amountStr),
          category: parts[3] || 'Other',
        };
      })
      .filter((e) => !Number.isNaN(e.amount));
  }

  private getPeriodStart(now: Date, period: string): Date {
    const start = new Date(now);
    switch (period) {
      case 'week':
        start.setDate(start.getDate() - 7);
        break;
      case 'month':
        start.setMonth(start.getMonth() - 1);
        break;
      case 'quarter':
        start.setMonth(start.getMonth() - 3);
        break;
      case 'year':
        start.setFullYear(start.getFullYear() - 1);
        break;
      default:
        start.setMonth(start.getMonth() - 1);
    }
    return start;
  }

  private getDaysDiff(start: Date, end: Date): number {
    const diff = Math.abs(end.getTime() - start.getTime());
    return Math.ceil(diff / (1000 * 60 * 60 * 24)) || 1;
  }

  private groupByDay(
    items: Array<{ date: string; [key: string]: string | number | boolean }>,
    valueKey: string,
    periodStart: Date
  ): Array<{ date: string; value: number }> {
    const grouped: { [key: string]: number } = {};
    const now = new Date();

    // Initialize all days in period
    const current = new Date(periodStart);
    while (current <= now) {
      const key = current.toISOString().split('T')[0];
      grouped[key] = 0;
      current.setDate(current.getDate() + 1);
    }

    // Sum values by day
    for (const item of items) {
      const key = new Date(item.date).toISOString().split('T')[0];
      if (grouped[key] !== undefined) {
        const value = item[valueKey];
        grouped[key] += typeof value === 'number' ? value : 0;
      }
    }

    return Object.entries(grouped)
      .map(([date, value]) => ({ date, value }))
      .slice(-14); // Last 14 days for chart
  }

  private groupByField(
    items: Array<{ [key: string]: string | number | boolean }>,
    groupKey: string,
    valueKey: string
  ): Array<{ label: string; value: number }> {
    const grouped: { [key: string]: number } = {};

    for (const item of items) {
      const key = String(item[groupKey] || 'Unknown');
      const value = item[valueKey];
      grouped[key] =
        (grouped[key] || 0) + (typeof value === 'number' ? value : 0);
    }

    return Object.entries(grouped)
      .map(([label, value]) => ({ label, value }))
      .sort((a, b) => b.value - a.value)
      .slice(0, 5); // Top 5
  }

  private async exportToCsv(type: string) {
    let data = '';
    let filename = '';

    switch (type) {
      case 'invoices': {
        const invoicesResult = await this.cli.listInvoices();
        data = invoicesResult.stdout || '';
        filename = 'invoices.csv';
        break;
      }
      case 'time': {
        const timeResult = await this.cli.listTimeEntries();
        data = timeResult.stdout || '';
        filename = 'time-entries.csv';
        break;
      }
      case 'expenses': {
        const expensesResult = await this.cli.listExpenses();
        data = expensesResult.stdout || '';
        filename = 'expenses.csv';
        break;
      }
    }

    if (data) {
      const uri = await vscode.window.showSaveDialog({
        defaultUri: vscode.Uri.file(filename),
        filters: { 'CSV Files': ['csv'] },
      });

      if (uri) {
        await vscode.workspace.fs.writeFile(uri, Buffer.from(data, 'utf-8'));
        vscode.window.showInformationMessage(`Exported to ${uri.fsPath}`);
      }
    }
  }

  private getHtmlContent(data: StatisticsData, period: string): string {
    const formatCurrency = (value: number) =>
      `$${value.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
    const formatHours = (value: number) => value.toFixed(1);

    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>UNG Statistics</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        :root {
            --bg-primary: var(--vscode-editor-background);
            --bg-secondary: var(--vscode-sideBar-background);
            --text-primary: var(--vscode-editor-foreground);
            --text-secondary: var(--vscode-descriptionForeground);
            --border: var(--vscode-panel-border);
            --accent: var(--vscode-button-background);
            --success: #4caf50;
            --warning: #ff9800;
            --danger: #f44336;
        }

        * { box-sizing: border-box; margin: 0; padding: 0; }

        body {
            font-family: var(--vscode-font-family);
            background: var(--bg-primary);
            color: var(--text-primary);
            padding: 20px;
            line-height: 1.5;
        }

        .header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 24px;
            flex-wrap: wrap;
            gap: 16px;
        }

        h1 {
            font-size: 24px;
            font-weight: 600;
        }

        .controls {
            display: flex;
            gap: 8px;
            flex-wrap: wrap;
        }

        button, select {
            background: var(--vscode-button-background);
            color: var(--vscode-button-foreground);
            border: none;
            padding: 8px 16px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 13px;
        }

        button:hover, select:hover {
            background: var(--vscode-button-hoverBackground);
        }

        select {
            background: var(--bg-secondary);
            color: var(--text-primary);
            border: 1px solid var(--border);
        }

        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 16px;
            margin-bottom: 24px;
        }

        .stat-card {
            background: var(--bg-secondary);
            border: 1px solid var(--border);
            border-radius: 8px;
            padding: 20px;
        }

        .stat-card.highlight {
            border-left: 4px solid var(--accent);
        }

        .stat-card.success { border-left-color: var(--success); }
        .stat-card.warning { border-left-color: var(--warning); }
        .stat-card.danger { border-left-color: var(--danger); }

        .stat-label {
            font-size: 12px;
            color: var(--text-secondary);
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-bottom: 4px;
        }

        .stat-value {
            font-size: 28px;
            font-weight: 600;
        }

        .stat-value.small { font-size: 20px; }

        .stat-subtitle {
            font-size: 12px;
            color: var(--text-secondary);
            margin-top: 4px;
        }

        .charts-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
            gap: 24px;
            margin-bottom: 24px;
        }

        .chart-card {
            background: var(--bg-secondary);
            border: 1px solid var(--border);
            border-radius: 8px;
            padding: 20px;
        }

        .chart-title {
            font-size: 14px;
            font-weight: 600;
            margin-bottom: 16px;
        }

        .chart-container {
            position: relative;
            height: 200px;
        }

        .export-section {
            background: var(--bg-secondary);
            border: 1px solid var(--border);
            border-radius: 8px;
            padding: 20px;
        }

        .export-section h3 {
            font-size: 14px;
            margin-bottom: 12px;
        }

        .export-buttons {
            display: flex;
            gap: 8px;
            flex-wrap: wrap;
        }

        .export-buttons button {
            background: transparent;
            border: 1px solid var(--border);
            color: var(--text-primary);
        }

        .export-buttons button:hover {
            background: var(--bg-primary);
        }

        @media (max-width: 600px) {
            .charts-grid { grid-template-columns: 1fr; }
            .stat-value { font-size: 22px; }
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Statistics & Reports</h1>
        <div class="controls">
            <select id="periodSelect" onchange="changePeriod(this.value)">
                <option value="week" ${period === 'week' ? 'selected' : ''}>Last 7 Days</option>
                <option value="month" ${period === 'month' ? 'selected' : ''}>Last 30 Days</option>
                <option value="quarter" ${period === 'quarter' ? 'selected' : ''}>Last 3 Months</option>
                <option value="year" ${period === 'year' ? 'selected' : ''}>Last Year</option>
            </select>
            <button onclick="refresh()">Refresh</button>
        </div>
    </div>

    <div class="stats-grid">
        <div class="stat-card highlight success">
            <div class="stat-label">Total Revenue</div>
            <div class="stat-value">${formatCurrency(data.revenue.total)}</div>
            <div class="stat-subtitle">${data.totalInvoices} invoices</div>
        </div>
        <div class="stat-card highlight">
            <div class="stat-label">Paid</div>
            <div class="stat-value">${formatCurrency(data.revenue.paid)}</div>
        </div>
        <div class="stat-card highlight warning">
            <div class="stat-label">Pending</div>
            <div class="stat-value">${formatCurrency(data.revenue.pending)}</div>
        </div>
        <div class="stat-card highlight danger">
            <div class="stat-label">Overdue</div>
            <div class="stat-value">${formatCurrency(data.revenue.overdue)}</div>
        </div>
    </div>

    <div class="stats-grid">
        <div class="stat-card">
            <div class="stat-label">Hours Tracked</div>
            <div class="stat-value">${formatHours(data.time.totalHours)}h</div>
            <div class="stat-subtitle">${formatHours(data.time.billableHours)}h billable</div>
        </div>
        <div class="stat-card">
            <div class="stat-label">Avg Hours/Day</div>
            <div class="stat-value small">${formatHours(data.time.avgPerDay)}h</div>
        </div>
        <div class="stat-card">
            <div class="stat-label">Total Expenses</div>
            <div class="stat-value small">${formatCurrency(data.expenses.total)}</div>
        </div>
        <div class="stat-card highlight ${data.netIncome >= 0 ? 'success' : 'danger'}">
            <div class="stat-label">Net Income</div>
            <div class="stat-value">${formatCurrency(data.netIncome)}</div>
            <div class="stat-subtitle">Revenue - Expenses</div>
        </div>
    </div>

    <div class="stats-grid">
        <div class="stat-card">
            <div class="stat-label">Active Contracts</div>
            <div class="stat-value small">${data.activeContracts}</div>
        </div>
        <div class="stat-card">
            <div class="stat-label">Avg Invoice Value</div>
            <div class="stat-value small">${formatCurrency(data.avgInvoiceValue)}</div>
        </div>
    </div>

    <div class="charts-grid">
        <div class="chart-card">
            <div class="chart-title">Revenue Trend (Last 14 Days)</div>
            <div class="chart-container">
                <canvas id="revenueChart"></canvas>
            </div>
        </div>
        <div class="chart-card">
            <div class="chart-title">Hours Trend (Last 14 Days)</div>
            <div class="chart-container">
                <canvas id="hoursChart"></canvas>
            </div>
        </div>
        <div class="chart-card">
            <div class="chart-title">Revenue by Client</div>
            <div class="chart-container">
                <canvas id="clientRevenueChart"></canvas>
            </div>
        </div>
        <div class="chart-card">
            <div class="chart-title">Invoice Status</div>
            <div class="chart-container">
                <canvas id="statusChart"></canvas>
            </div>
        </div>
    </div>

    <div class="export-section">
        <h3>Export Data</h3>
        <div class="export-buttons">
            <button onclick="exportCsv('invoices')">Export Invoices</button>
            <button onclick="exportCsv('time')">Export Time Entries</button>
            <button onclick="exportCsv('expenses')">Export Expenses</button>
        </div>
    </div>

    <script>
        const vscode = acquireVsCodeApi();

        function refresh() {
            vscode.postMessage({ command: 'refresh' });
        }

        function changePeriod(period) {
            vscode.postMessage({ command: 'changePeriod', period });
        }

        function exportCsv(type) {
            vscode.postMessage({ command: 'exportCsv', type });
        }

        // Chart data
        const revenueByDay = ${JSON.stringify(data.charts.revenueByDay)};
        const hoursByDay = ${JSON.stringify(data.charts.hoursByDay)};
        const revenueByClient = ${JSON.stringify(data.charts.revenueByClient)};
        const invoicesByStatus = ${JSON.stringify(data.charts.invoicesByStatus)};

        // Chart.js configuration
        Chart.defaults.color = getComputedStyle(document.body).getPropertyValue('--text-secondary').trim() || '#888';
        Chart.defaults.borderColor = getComputedStyle(document.body).getPropertyValue('--border').trim() || '#333';

        // Revenue Chart
        new Chart(document.getElementById('revenueChart'), {
            type: 'line',
            data: {
                labels: revenueByDay.map(d => d.date.slice(5)),
                datasets: [{
                    label: 'Revenue',
                    data: revenueByDay.map(d => d.value),
                    borderColor: '#4caf50',
                    backgroundColor: 'rgba(76, 175, 80, 0.1)',
                    fill: true,
                    tension: 0.3
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: { legend: { display: false } },
                scales: {
                    y: { beginAtZero: true }
                }
            }
        });

        // Hours Chart
        new Chart(document.getElementById('hoursChart'), {
            type: 'bar',
            data: {
                labels: hoursByDay.map(d => d.date.slice(5)),
                datasets: [{
                    label: 'Hours',
                    data: hoursByDay.map(d => d.value),
                    backgroundColor: '#2196f3'
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: { legend: { display: false } },
                scales: {
                    y: { beginAtZero: true }
                }
            }
        });

        // Client Revenue Chart
        new Chart(document.getElementById('clientRevenueChart'), {
            type: 'doughnut',
            data: {
                labels: revenueByClient.map(d => d.label),
                datasets: [{
                    data: revenueByClient.map(d => d.value),
                    backgroundColor: ['#4caf50', '#2196f3', '#ff9800', '#9c27b0', '#f44336']
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'right',
                        labels: { boxWidth: 12 }
                    }
                }
            }
        });

        // Status Chart
        new Chart(document.getElementById('statusChart'), {
            type: 'pie',
            data: {
                labels: ['Draft', 'Pending', 'Sent', 'Paid', 'Overdue'],
                datasets: [{
                    data: [
                        invoicesByStatus.draft,
                        invoicesByStatus.pending,
                        invoicesByStatus.sent,
                        invoicesByStatus.paid,
                        invoicesByStatus.overdue
                    ],
                    backgroundColor: ['#9e9e9e', '#ff9800', '#2196f3', '#4caf50', '#f44336']
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'right',
                        labels: { boxWidth: 12 }
                    }
                }
            }
        });
    </script>
</body>
</html>`;
  }

  public dispose() {
    StatisticsPanel.currentPanel = undefined;

    this.panel.dispose();

    while (this.disposables.length) {
      const x = this.disposables.pop();
      if (x) {
        x.dispose();
      }
    }
  }
}

interface StatisticsData {
  period: string;
  periodStart: string;
  revenue: {
    total: number;
    paid: number;
    pending: number;
    overdue: number;
  };
  time: {
    totalHours: number;
    billableHours: number;
    avgPerDay: number;
  };
  expenses: {
    total: number;
  };
  netIncome: number;
  charts: {
    revenueByDay: Array<{ date: string; value: number }>;
    hoursByDay: Array<{ date: string; value: number }>;
    revenueByClient: Array<{ label: string; value: number }>;
    hoursByClient: Array<{ label: string; value: number }>;
    invoicesByStatus: {
      draft: number;
      pending: number;
      sent: number;
      paid: number;
      overdue: number;
    };
  };
  activeContracts: number;
  totalInvoices: number;
  avgInvoiceValue: number;
}
