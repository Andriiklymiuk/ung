import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';

interface Expense {
  id: number;
  description: string;
  amount: number;
  currency: string;
  category: string;
  date: string;
  vendor: string;
}

/**
 * Expense Panel - View and manage expenses
 */
export class ExpensePanel {
  public static currentPanel: ExpensePanel | undefined;
  private static readonly viewType = 'ungExpenses';

  private readonly panel: vscode.WebviewPanel;
  private readonly cli: UngCli;
  private disposables: vscode.Disposable[] = [];
  private expenses: Expense[] = [];

  public static createOrShow(cli: UngCli) {
    const column = vscode.window.activeTextEditor
      ? vscode.window.activeTextEditor.viewColumn
      : undefined;

    if (ExpensePanel.currentPanel) {
      ExpensePanel.currentPanel.panel.reveal(column);
      ExpensePanel.currentPanel.update();
      return;
    }

    const panel = vscode.window.createWebviewPanel(
      ExpensePanel.viewType,
      'Expenses',
      column || vscode.ViewColumn.One,
      {
        enableScripts: true,
        retainContextWhenHidden: true,
      }
    );

    ExpensePanel.currentPanel = new ExpensePanel(panel, cli);
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
          case 'addExpense':
            await vscode.commands.executeCommand('ung.logExpense');
            await this.update();
            break;
          case 'deleteExpense':
            await this.deleteExpense(message.id);
            break;
        }
      },
      null,
      this.disposables
    );
  }

  private async deleteExpense(id: number): Promise<void> {
    const confirm = await vscode.window.showWarningMessage(
      'Delete this expense?',
      { modal: true },
      'Delete'
    );
    if (confirm === 'Delete') {
      const result = await this.cli.exec(['expense', 'delete', id.toString()]);
      if (result.success) {
        vscode.window.showInformationMessage('Expense deleted');
        await this.update();
      } else {
        vscode.window.showErrorMessage('Failed to delete expense');
      }
    }
  }

  private async update() {
    this.expenses = await this.loadExpenses();
    this.panel.webview.html = this.getHtmlContent();
  }

  private async loadExpenses(): Promise<Expense[]> {
    const result = await this.cli.listExpenses();
    if (!result.success || !result.stdout) {
      return [];
    }

    const lines = result.stdout
      .split('\n')
      .filter((l) => l.trim() && !l.includes('─') && !l.includes('ID'));
    if (lines.length < 2) return [];

    const expenses: Expense[] = [];
    for (let i = 1; i < lines.length; i++) {
      const parts = lines[i]
        .split(/\s{2,}/)
        .map((p) => p.trim())
        .filter((p) => p);
      if (parts.length >= 5) {
        const id = parseInt(parts[0], 10);
        if (!Number.isNaN(id)) {
          expenses.push({
            id,
            description: parts[1] || 'Unknown',
            amount: parseFloat(parts[2]?.replace(/[^0-9.-]/g, '') || '0'),
            currency: 'USD',
            category: parts[3] || 'other',
            date: parts[4] || '',
            vendor: parts[5] || '',
          });
        }
      }
    }
    return expenses;
  }

  private formatCurrency(amount: number): string {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
    }).format(amount);
  }

  private getCategoryIcon(category: string): string {
    const icons: Record<string, string> = {
      software: '💻',
      hardware: '🖥️',
      travel: '✈️',
      meals: '🍔',
      office_supplies: '📎',
      utilities: '💡',
      marketing: '📢',
      other: '📦',
    };
    return icons[category] || '📦';
  }

  private getHtmlContent(): string {
    const totalExpenses = this.expenses.reduce((sum, e) => sum + e.amount, 0);

    const expenseRows = this.expenses
      .map(
        (e) => `
        <tr>
          <td>${e.date}</td>
          <td>
            <span class="category-badge">${this.getCategoryIcon(e.category)} ${e.category}</span>
          </td>
          <td>${e.description}</td>
          <td>${e.vendor || '-'}</td>
          <td class="amount">${this.formatCurrency(e.amount)}</td>
          <td>
            <button class="delete-btn" onclick="deleteExpense(${e.id})" title="Delete">×</button>
          </td>
        </tr>
      `
      )
      .join('');

    // Group by category for summary
    const byCategory = this.expenses.reduce(
      (acc, e) => {
        acc[e.category] = (acc[e.category] || 0) + e.amount;
        return acc;
      },
      {} as Record<string, number>
    );

    const categorySummary = Object.entries(byCategory)
      .sort((a, b) => b[1] - a[1])
      .map(
        ([cat, amount]) => `
        <div class="category-item">
          <span>${this.getCategoryIcon(cat)} ${cat}</span>
          <span class="category-amount">${this.formatCurrency(amount)}</span>
        </div>
      `
      )
      .join('');

    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Expenses</title>
    <style>
        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }
        body {
            font-family: var(--vscode-font-family);
            color: var(--vscode-foreground);
            background-color: var(--vscode-editor-background);
            padding: 20px;
            line-height: 1.5;
        }
        .header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
        }
        h1 {
            font-size: 24px;
            font-weight: 600;
        }
        .header-actions {
            display: flex;
            gap: 8px;
        }
        .btn {
            padding: 8px 16px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 13px;
            font-family: var(--vscode-font-family);
            transition: all 0.15s;
        }
        .btn-primary {
            background: var(--vscode-button-background);
            color: var(--vscode-button-foreground);
        }
        .btn-primary:hover {
            background: var(--vscode-button-hoverBackground);
        }
        .btn-secondary {
            background: var(--vscode-input-background);
            color: var(--vscode-foreground);
        }
        .btn-secondary:hover {
            background: var(--vscode-list-hoverBackground);
        }
        .summary-cards {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 16px;
            margin-bottom: 24px;
        }
        .card {
            background: var(--vscode-input-background);
            border-radius: 8px;
            padding: 16px;
        }
        .card-title {
            font-size: 12px;
            color: var(--vscode-descriptionForeground);
            text-transform: uppercase;
            margin-bottom: 8px;
        }
        .card-value {
            font-size: 24px;
            font-weight: 600;
            color: var(--vscode-charts-red, #f44336);
        }
        .category-summary {
            display: flex;
            flex-direction: column;
            gap: 8px;
        }
        .category-item {
            display: flex;
            justify-content: space-between;
            font-size: 13px;
        }
        .category-amount {
            font-weight: 500;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 16px;
        }
        th, td {
            text-align: left;
            padding: 12px;
            border-bottom: 1px solid var(--vscode-panel-border);
        }
        th {
            font-size: 11px;
            text-transform: uppercase;
            color: var(--vscode-descriptionForeground);
            font-weight: 600;
        }
        td.amount {
            font-family: monospace;
            color: var(--vscode-charts-red, #f44336);
        }
        .category-badge {
            font-size: 12px;
            padding: 2px 8px;
            background: var(--vscode-badge-background);
            color: var(--vscode-badge-foreground);
            border-radius: 4px;
        }
        .delete-btn {
            background: none;
            border: none;
            color: var(--vscode-descriptionForeground);
            cursor: pointer;
            font-size: 16px;
            padding: 4px 8px;
            border-radius: 4px;
            opacity: 0.5;
        }
        .delete-btn:hover {
            opacity: 1;
            background: var(--vscode-inputValidation-errorBackground);
            color: var(--vscode-errorForeground);
        }
        .empty-state {
            text-align: center;
            padding: 40px;
            color: var(--vscode-descriptionForeground);
        }
        .empty-state h3 {
            margin-bottom: 8px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Expenses</h1>
        <div class="header-actions">
            <button class="btn btn-secondary" onclick="refresh()">Refresh</button>
            <button class="btn btn-primary" onclick="addExpense()">+ Add Expense</button>
        </div>
    </div>

    <div class="summary-cards">
        <div class="card">
            <div class="card-title">Total Expenses</div>
            <div class="card-value">${this.formatCurrency(totalExpenses)}</div>
        </div>
        <div class="card">
            <div class="card-title">By Category</div>
            <div class="category-summary">
                ${categorySummary || '<span style="color: var(--vscode-descriptionForeground)">No expenses</span>'}
            </div>
        </div>
    </div>

    ${
      this.expenses.length > 0
        ? `
        <table>
            <thead>
                <tr>
                    <th>Date</th>
                    <th>Category</th>
                    <th>Description</th>
                    <th>Vendor</th>
                    <th>Amount</th>
                    <th></th>
                </tr>
            </thead>
            <tbody>
                ${expenseRows}
            </tbody>
        </table>
    `
        : `
        <div class="empty-state">
            <h3>No expenses yet</h3>
            <p>Click "+ Add Expense" to log your first expense</p>
        </div>
    `
    }

    <script>
        const vscode = acquireVsCodeApi();

        function refresh() {
            vscode.postMessage({ command: 'refresh' });
        }

        function addExpense() {
            vscode.postMessage({ command: 'addExpense' });
        }

        function deleteExpense(id) {
            vscode.postMessage({ command: 'deleteExpense', id: id });
        }
    </script>
</body>
</html>`;
  }

  public dispose() {
    ExpensePanel.currentPanel = undefined;
    this.panel.dispose();
    while (this.disposables.length) {
      const d = this.disposables.pop();
      if (d) {
        d.dispose();
      }
    }
  }
}
