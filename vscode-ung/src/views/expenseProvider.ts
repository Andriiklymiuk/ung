import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';
import { Formatter } from '../utils/formatting';

/**
 * Expense tree item
 */
export class ExpenseItem extends vscode.TreeItem {
    public readonly expenseDescription: string;

    constructor(
        public readonly itemId: number,
        description: string,
        public readonly amount: number,
        public readonly currency: string,
        public readonly category: string,
        public readonly date: string,
        public readonly collapsibleState: vscode.TreeItemCollapsibleState
    ) {
        super(description, collapsibleState);

        this.id = String(itemId);
        this.expenseDescription = description;
        this.tooltip = `${description}\nAmount: ${Formatter.formatCurrency(amount, currency)}\nCategory: ${category}\nDate: ${Formatter.formatDate(date)}`;
        this.description = `${Formatter.formatCurrency(amount, currency)} • ${category}`;
        this.contextValue = 'expense';
        this.iconPath = new vscode.ThemeIcon('credit-card');
    }
}

/**
 * Expense tree data provider
 */
export class ExpenseProvider implements vscode.TreeDataProvider<ExpenseItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<ExpenseItem | undefined | null | void> = new vscode.EventEmitter<ExpenseItem | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<ExpenseItem | undefined | null | void> = this._onDidChangeTreeData.event;

    constructor(private cli: UngCli) {}

    refresh(): void {
        this._onDidChangeTreeData.fire();
    }

    getTreeItem(element: ExpenseItem): vscode.TreeItem {
        return element;
    }

    async getChildren(element?: ExpenseItem): Promise<ExpenseItem[]> {
        if (element) {
            return [];
        }

        try {
            const result = await this.cli.listExpenses();

            if (!result.success || !result.stdout) {
                return [];
            }

            // Parse the CLI output
            const expenses = this.parseExpenseOutput(result.stdout);
            return expenses;
        } catch (error) {
            vscode.window.showErrorMessage(`Failed to load expenses: ${error}`);
            return [];
        }
    }

    /**
     * Parse expense list output from CLI
     */
    private parseExpenseOutput(output: string): ExpenseItem[] {
        const lines = output.split('\n').filter(line => line.trim());
        const expenses: ExpenseItem[] = [];

        for (let i = 1; i < lines.length; i++) { // Skip header
            const line = lines[i].trim();
            if (!line || line.startsWith('Total:')) continue;

            const parts = line.split(/\s{2,}/); // Split by multiple spaces
            if (parts.length >= 6) {
                const id = parseInt(parts[0]);
                const date = parts[1];
                const description = parts[2];
                const category = parts[3];
                const amountParts = parts[5].split(' ');
                const amount = parseFloat(amountParts[0]);
                const currency = amountParts[1] || 'USD';

                if (!isNaN(id) && !isNaN(amount)) {
                    expenses.push(new ExpenseItem(
                        id,
                        description,
                        amount,
                        currency,
                        category,
                        date,
                        vscode.TreeItemCollapsibleState.None
                    ));
                }
            }
        }

        return expenses;
    }
}
