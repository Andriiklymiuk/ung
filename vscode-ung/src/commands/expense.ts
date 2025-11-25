import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';

/**
 * Expense command handlers
 */
export class ExpenseCommands {
    constructor(private cli: UngCli, private refreshCallback?: () => void) {}

    /**
     * Log a new expense
     */
    async logExpense(): Promise<void> {
        vscode.window.showInformationMessage(
            'Expense logging is interactive. Please use the CLI: ung expense add'
        );
    }

    /**
     * Edit an expense
     */
    async editExpense(expenseId?: number): Promise<void> {
        if (!expenseId) {
            vscode.window.showErrorMessage('No expense selected');
            return;
        }

        vscode.window.showInformationMessage(
            'Expense editing will be available in a future version.'
        );
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

        vscode.window.showInformationMessage('Expense deletion will be available in a future version.');
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
