import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';
import { Formatter } from '../utils/formatting';

/**
 * Expense tree item types
 */
type ExpenseItemType = 'summary' | 'action' | 'category' | 'expense';

/**
 * Expense tree item
 */
export class ExpenseItem extends vscode.TreeItem {
  public readonly expenseDescription: string;

  constructor(
    public readonly itemType: ExpenseItemType,
    public readonly itemId: number,
    description: string,
    public readonly amount: number,
    public readonly currency: string,
    public readonly category: string,
    public readonly date: string,
    public readonly collapsibleState: vscode.TreeItemCollapsibleState,
    public readonly children?: ExpenseItem[]
  ) {
    super(description, collapsibleState);

    this.id = `${itemType}-${itemId}-${category}`;
    this.expenseDescription = description;

    switch (itemType) {
      case 'summary':
        this.contextValue = 'summary';
        break;
      case 'action':
        this.contextValue = 'action';
        break;
      case 'category':
        this.contextValue = 'category';
        this.iconPath = this.getCategoryIcon(category);
        break;
      case 'expense':
        this.tooltip = this.buildTooltip(description);
        this.description = `${Formatter.formatCurrency(amount, currency)} • ${Formatter.formatDate(date)}`;
        this.contextValue = 'expense';
        this.iconPath = new vscode.ThemeIcon('credit-card');
        break;
    }
  }

  private buildTooltip(description: string): string {
    return `**${description}**\n\nAmount: ${Formatter.formatCurrency(this.amount, this.currency)}\nCategory: ${this.category}\nDate: ${Formatter.formatDate(this.date)}`;
  }

  private getCategoryIcon(category: string): vscode.ThemeIcon {
    const iconMap: Record<string, string> = {
      software: 'symbol-misc',
      office: 'home',
      travel: 'globe',
      equipment: 'tools',
      marketing: 'megaphone',
      utilities: 'plug',
      meals: 'coffee',
      other: 'ellipsis',
    };
    const iconName = iconMap[category.toLowerCase()] || 'tag';
    return new vscode.ThemeIcon(
      iconName,
      new vscode.ThemeColor('charts.purple')
    );
  }
}

/**
 * Expense tree data provider with summary and categories
 */
export class ExpenseProvider implements vscode.TreeDataProvider<ExpenseItem> {
  private _onDidChangeTreeData: vscode.EventEmitter<
    ExpenseItem | undefined | null | undefined
  > = new vscode.EventEmitter<ExpenseItem | undefined | null | undefined>();
  readonly onDidChangeTreeData: vscode.Event<
    ExpenseItem | undefined | null | undefined
  > = this._onDidChangeTreeData.event;

  private categoryItems: Map<string, ExpenseItem> = new Map();

  constructor(private cli: UngCli) {}

  refresh(): void {
    this.categoryItems.clear();
    this._onDidChangeTreeData.fire(undefined);
  }

  getTreeItem(element: ExpenseItem): vscode.TreeItem {
    return element;
  }

  async getChildren(element?: ExpenseItem): Promise<ExpenseItem[]> {
    // Return children of a category
    if (element && element.itemType === 'category' && element.children) {
      return element.children;
    }

    if (element) {
      return [];
    }

    try {
      const result = await this.cli.listExpenses();

      if (!result.success || !result.stdout) {
        return this.getEmptyState();
      }

      // Parse the CLI output
      const expenses = this.parseExpenseOutput(result.stdout);

      if (expenses.length === 0) {
        return this.getEmptyState();
      }

      return this.buildTreeItems(expenses);
    } catch (error) {
      vscode.window.showErrorMessage(`Failed to load expenses: ${error}`);
      return this.getEmptyState();
    }
  }

  /**
   * Build tree items with summary and category grouping
   */
  private buildTreeItems(expenses: ExpenseItem[]): ExpenseItem[] {
    const items: ExpenseItem[] = [];

    // Calculate totals
    const totalAmount = expenses.reduce((sum, e) => sum + e.amount, 0);
    const currency = expenses[0]?.currency || 'USD';

    // Summary
    const summaryItem = new ExpenseItem(
      'summary',
      0,
      `Total: ${Formatter.formatCurrency(totalAmount, currency)}`,
      totalAmount,
      currency,
      '',
      '',
      vscode.TreeItemCollapsibleState.None
    );
    summaryItem.iconPath = new vscode.ThemeIcon(
      'graph',
      new vscode.ThemeColor('charts.red')
    );
    summaryItem.description = `${expenses.length} expense${expenses.length !== 1 ? 's' : ''}`;
    items.push(summaryItem);

    // Quick action
    const addItem = new ExpenseItem(
      'action',
      -1,
      'Log New Expense',
      0,
      currency,
      '',
      '',
      vscode.TreeItemCollapsibleState.None
    );
    addItem.iconPath = new vscode.ThemeIcon(
      'add',
      new vscode.ThemeColor('charts.green')
    );
    addItem.command = {
      command: 'ung.logExpense',
      title: 'Log Expense',
    };
    items.push(addItem);

    // Group by category
    const byCategory = new Map<string, ExpenseItem[]>();
    for (const expense of expenses) {
      const cat = expense.category || 'Other';
      if (!byCategory.has(cat)) {
        byCategory.set(cat, []);
      }
      byCategory.get(cat)!.push(expense);
    }

    // Sort categories by total amount (descending)
    const sortedCategories = Array.from(byCategory.entries()).sort((a, b) => {
      const totalA = a[1].reduce((sum, e) => sum + e.amount, 0);
      const totalB = b[1].reduce((sum, e) => sum + e.amount, 0);
      return totalB - totalA;
    });

    // Create category items
    for (const [category, categoryExpenses] of sortedCategories) {
      const categoryTotal = categoryExpenses.reduce(
        (sum, e) => sum + e.amount,
        0
      );

      const categoryItem = new ExpenseItem(
        'category',
        0,
        category,
        categoryTotal,
        currency,
        category,
        '',
        vscode.TreeItemCollapsibleState.Collapsed,
        categoryExpenses
      );
      categoryItem.description = `${Formatter.formatCurrency(categoryTotal, currency)} (${categoryExpenses.length})`;

      this.categoryItems.set(category, categoryItem);
      items.push(categoryItem);
    }

    return items;
  }

  /**
   * Empty state with helpful message
   */
  private getEmptyState(): ExpenseItem[] {
    const emptyItem = new ExpenseItem(
      'summary',
      0,
      'No expenses recorded',
      0,
      'USD',
      '',
      '',
      vscode.TreeItemCollapsibleState.None
    );
    emptyItem.iconPath = new vscode.ThemeIcon('info');
    emptyItem.description = 'Track your business expenses';

    const addItem = new ExpenseItem(
      'action',
      -1,
      'Log Your First Expense',
      0,
      'USD',
      '',
      '',
      vscode.TreeItemCollapsibleState.None
    );
    addItem.iconPath = new vscode.ThemeIcon(
      'add',
      new vscode.ThemeColor('charts.green')
    );
    addItem.command = {
      command: 'ung.logExpense',
      title: 'Log Expense',
    };

    return [emptyItem, addItem];
  }

  /**
   * Parse expense list output from CLI
   */
  private parseExpenseOutput(output: string): ExpenseItem[] {
    const lines = output.split('\n').filter((line) => line.trim());
    const expenses: ExpenseItem[] = [];

    for (let i = 1; i < lines.length; i++) {
      // Skip header
      const line = lines[i].trim();
      if (!line || line.startsWith('Total:')) continue;

      const parts = line.split(/\s{2,}/); // Split by multiple spaces
      if (parts.length >= 6) {
        const id = parseInt(parts[0], 10);
        const date = parts[1];
        const description = parts[2];
        const category = parts[3];
        const amountParts = parts[5].split(' ');
        const amount = parseFloat(amountParts[0]);
        const currency = amountParts[1] || 'USD';

        if (!Number.isNaN(id) && !Number.isNaN(amount)) {
          expenses.push(
            new ExpenseItem(
              'expense',
              id,
              description,
              amount,
              currency,
              category,
              date,
              vscode.TreeItemCollapsibleState.None
            )
          );
        }
      }
    }

    return expenses;
  }
}
