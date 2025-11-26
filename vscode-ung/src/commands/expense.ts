import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';

/**
 * Expense command handlers
 */
export class ExpenseCommands {
    private refreshCallback?: () => void;

    constructor(private cli: UngCli, refreshCallback?: () => void) {
        this.refreshCallback = refreshCallback;
    }

    /**
     * Log a new expense
     */
    async logExpense(): Promise<void> {
        // Get description
        const description = await vscode.window.showInputBox({
            prompt: 'Expense description',
            placeHolder: 'e.g., Adobe Creative Cloud subscription',
            validateInput: (value) => value ? null : 'Description is required'
        });
        if (!description) return;

        // Get amount
        const amountStr = await vscode.window.showInputBox({
            prompt: 'Amount',
            placeHolder: 'e.g., 52.99',
            validateInput: (value) => {
                if (!value) return 'Amount is required';
                const num = parseFloat(value);
                if (isNaN(num) || num <= 0) return 'Please enter a valid positive number';
                return null;
            }
        });
        if (!amountStr) return;

        // Get category
        const categories = [
            { label: 'Software', value: 'software' },
            { label: 'Hardware', value: 'hardware' },
            { label: 'Travel', value: 'travel' },
            { label: 'Meals', value: 'meals' },
            { label: 'Office Supplies', value: 'office_supplies' },
            { label: 'Utilities', value: 'utilities' },
            { label: 'Marketing', value: 'marketing' },
            { label: 'Other', value: 'other' }
        ];
        const categorySelection = await vscode.window.showQuickPick(categories, {
            placeHolder: 'Select category'
        });
        if (!categorySelection) return;

        // Get vendor (optional)
        const vendor = await vscode.window.showInputBox({
            prompt: 'Vendor (optional)',
            placeHolder: 'e.g., Adobe'
        });

        // Get date (optional)
        const today = new Date().toISOString().split('T')[0];
        const dateStr = await vscode.window.showInputBox({
            prompt: 'Date (YYYY-MM-DD, leave empty for today)',
            placeHolder: today,
            validateInput: (value) => {
                if (!value) return null;
                if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return 'Use format YYYY-MM-DD';
                return null;
            }
        });

        // Get notes (optional)
        const notes = await vscode.window.showInputBox({
            prompt: 'Notes (optional)',
            placeHolder: 'Additional notes about this expense'
        });

        // Build command arguments
        const args = [
            'expense', 'add',
            '--description', description,
            '--amount', amountStr,
            '--category', categorySelection.value
        ];

        if (vendor) {
            args.push('--vendor', vendor);
        }
        if (dateStr) {
            args.push('--date', dateStr);
        }
        if (notes) {
            args.push('--notes', notes);
        }

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Adding expense...',
            cancellable: false
        }, async () => {
            const result = await this.cli.exec(args);

            if (result.success) {
                vscode.window.showInformationMessage(`Expense added: ${description} - $${amountStr}`);
                this.refreshCallback?.();
            } else {
                vscode.window.showErrorMessage(`Failed to add expense: ${result.error}`);
            }
        });
    }

    /**
     * Edit an expense
     */
    async editExpense(expenseId?: number): Promise<void> {
        if (!expenseId) {
            vscode.window.showErrorMessage('No expense selected');
            return;
        }

        // Get new description
        const description = await vscode.window.showInputBox({
            prompt: 'New description (leave empty to keep current)',
            placeHolder: 'e.g., Updated expense description'
        });

        // Get new amount
        const amountStr = await vscode.window.showInputBox({
            prompt: 'New amount (leave empty to keep current)',
            placeHolder: 'e.g., 52.99',
            validateInput: (value) => {
                if (!value) return null;
                const num = parseFloat(value);
                if (isNaN(num) || num <= 0) return 'Please enter a valid positive number';
                return null;
            }
        });

        // Get new category
        const categories = [
            { label: '(Keep current)', value: '' },
            { label: 'Software', value: 'software' },
            { label: 'Hardware', value: 'hardware' },
            { label: 'Travel', value: 'travel' },
            { label: 'Meals', value: 'meals' },
            { label: 'Office Supplies', value: 'office_supplies' },
            { label: 'Utilities', value: 'utilities' },
            { label: 'Marketing', value: 'marketing' },
            { label: 'Other', value: 'other' }
        ];
        const categorySelection = await vscode.window.showQuickPick(categories, {
            placeHolder: 'Select new category (or keep current)'
        });

        // Build command arguments
        const args = ['expense', 'edit', String(expenseId)];

        if (description) {
            args.push('--description', description);
        }
        if (amountStr) {
            args.push('--amount', amountStr);
        }
        if (categorySelection?.value) {
            args.push('--category', categorySelection.value);
        }

        // If no changes, show message
        if (args.length === 3) {
            vscode.window.showInformationMessage('No changes made');
            return;
        }

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Updating expense...',
            cancellable: false
        }, async () => {
            const result = await this.cli.exec(args);

            if (result.success) {
                vscode.window.showInformationMessage('Expense updated successfully');
                this.refreshCallback?.();
            } else {
                vscode.window.showErrorMessage(`Failed to update expense: ${result.error}`);
            }
        });
    }

    /**
     * Delete an expense
     */
    async deleteExpense(expenseId?: number): Promise<void> {
        if (!expenseId) {
            vscode.window.showErrorMessage('No expense selected');
            return;
        }

        const confirm = await vscode.window.showWarningMessage(
            `Are you sure you want to delete this expense?`,
            { modal: true },
            'Yes', 'No'
        );

        if (confirm !== 'Yes') return;

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Deleting expense...',
            cancellable: false
        }, async () => {
            const result = await this.cli.exec(['expense', 'delete', String(expenseId)]);

            if (result.success) {
                vscode.window.showInformationMessage('Expense deleted successfully');
                this.refreshCallback?.();
            } else {
                vscode.window.showErrorMessage(`Failed to delete expense: ${result.error}`);
            }
        });
    }

    /**
     * View expense report
     */
    async viewExpenseReport(): Promise<void> {
        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Generating expense report...',
            cancellable: false
        }, async () => {
            const result = await this.cli.getExpenseReport();

            if (result.success && result.stdout) {
                // Show output in a new document
                const doc = await vscode.workspace.openTextDocument({
                    content: result.stdout,
                    language: 'plaintext'
                });
                await vscode.window.showTextDocument(doc);
            } else {
                vscode.window.showErrorMessage(`Failed to generate report: ${result.error}`);
            }
        });
    }
}
