import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';

/**
 * Expense command handlers
 */
export class ExpenseCommands {
  private refreshCallback?: () => void;

  constructor(
    private cli: UngCli,
    refreshCallback?: () => void
  ) {
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
      validateInput: (value) => (value ? null : 'Description is required'),
    });
    if (!description) return;

    // Get amount
    const amountStr = await vscode.window.showInputBox({
      prompt: 'Amount',
      placeHolder: 'e.g., 52.99',
      validateInput: (value) => {
        if (!value) return 'Amount is required';
        const num = parseFloat(value);
        if (Number.isNaN(num) || num <= 0)
          return 'Please enter a valid positive number';
        return null;
      },
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
      { label: 'Other', value: 'other' },
    ];
    const categorySelection = await vscode.window.showQuickPick(categories, {
      placeHolder: 'Select category',
    });
    if (!categorySelection) return;

    // Get vendor (optional)
    const vendor = await vscode.window.showInputBox({
      prompt: 'Vendor (optional)',
      placeHolder: 'e.g., Adobe',
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
      },
    });

    // Get notes (optional)
    const notes = await vscode.window.showInputBox({
      prompt: 'Notes (optional)',
      placeHolder: 'Additional notes about this expense',
    });

    // Build command arguments
    const args = [
      'expense',
      'add',
      '--description',
      description,
      '--amount',
      amountStr,
      '--category',
      categorySelection.value,
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

    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Adding expense...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.exec(args);

        if (result.success) {
          vscode.window.showInformationMessage(
            `Expense added: ${description} - $${amountStr}`
          );
          this.refreshCallback?.();
        } else {
          vscode.window.showErrorMessage(
            `Failed to add expense: ${result.error}`
          );
        }
      }
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

    // Get new description
    const description = await vscode.window.showInputBox({
      prompt: 'New description (leave empty to keep current)',
      placeHolder: 'e.g., Updated expense description',
    });

    // Get new amount
    const amountStr = await vscode.window.showInputBox({
      prompt: 'New amount (leave empty to keep current)',
      placeHolder: 'e.g., 52.99',
      validateInput: (value) => {
        if (!value) return null;
        const num = parseFloat(value);
        if (Number.isNaN(num) || num <= 0)
          return 'Please enter a valid positive number';
        return null;
      },
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
      { label: 'Other', value: 'other' },
    ];
    const categorySelection = await vscode.window.showQuickPick(categories, {
      placeHolder: 'Select new category (or keep current)',
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

    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Updating expense...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.exec(args);

        if (result.success) {
          vscode.window.showInformationMessage('Expense updated successfully');
          this.refreshCallback?.();
        } else {
          vscode.window.showErrorMessage(
            `Failed to update expense: ${result.error}`
          );
        }
      }
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

    // Get expense details for confirmation and potential revert
    const expensesResult = await this.cli.listExpenses();
    let expenseData: {
      description: string;
      amount: number;
      currency: string;
      category: string;
      date: string;
      vendor?: string;
    } | null = null;

    if (expensesResult.success && expensesResult.stdout) {
      const expense = this.parseExpenseFromOutput(
        expensesResult.stdout,
        expenseId
      );
      if (expense) {
        expenseData = expense;
      }
    }

    const confirm = await vscode.window.showWarningMessage(
      expenseData
        ? `Are you sure you want to delete "${expenseData.description}"?`
        : `Are you sure you want to delete this expense?`,
      { modal: true },
      'Yes',
      'No'
    );

    if (confirm !== 'Yes') return;

    let deleteSuccess = false;
    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Deleting expense...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.exec([
          'expense',
          'delete',
          String(expenseId),
        ]);

        if (result.success) {
          deleteSuccess = true;
          this.refreshCallback?.();
        } else {
          vscode.window.showErrorMessage(
            `Failed to delete expense: ${result.error}`
          );
        }
      }
    );

    // Show success message with revert option
    if (deleteSuccess && expenseData) {
      const action = await vscode.window.showInformationMessage(
        `Expense "${expenseData.description}" deleted successfully!`,
        'Revert'
      );

      if (action === 'Revert') {
        await this.revertExpenseDeletion(expenseData);
      }
    } else if (deleteSuccess) {
      vscode.window.showInformationMessage('Expense deleted successfully');
    }
  }

  /**
   * Parse a specific expense from CLI output
   */
  private parseExpenseFromOutput(
    output: string,
    expenseId: number
  ): {
    description: string;
    amount: number;
    currency: string;
    category: string;
    date: string;
    vendor?: string;
  } | null {
    const lines = output.split('\n').filter((line) => line.trim());

    for (let i = 1; i < lines.length; i++) {
      const line = lines[i].trim();
      if (!line || line.startsWith('Total:')) continue;

      const parts = line.split(/\s{2,}/);
      if (parts.length >= 6) {
        const id = parseInt(parts[0], 10);
        if (id === expenseId) {
          const date = parts[1];
          const description = parts[2];
          const category = parts[3];
          const vendor = parts[4] !== '-' ? parts[4] : undefined;
          const amountParts = parts[5].split(' ');
          const amount = parseFloat(amountParts[0]);
          const currency = amountParts[1] || 'USD';

          return {
            description,
            amount,
            currency,
            category,
            date,
            vendor,
          };
        }
      }
    }

    return null;
  }

  /**
   * Revert an expense deletion by re-creating the expense
   */
  private async revertExpenseDeletion(expenseData: {
    description: string;
    amount: number;
    currency: string;
    category: string;
    date: string;
    vendor?: string;
  }): Promise<void> {
    const args = [
      'expense',
      'add',
      '--description',
      expenseData.description,
      '--amount',
      expenseData.amount.toString(),
      '--category',
      expenseData.category,
    ];

    if (expenseData.vendor) {
      args.push('--vendor', expenseData.vendor);
    }
    if (expenseData.date) {
      args.push('--date', expenseData.date);
    }

    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Reverting expense deletion...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.exec(args);

        if (result.success) {
          vscode.window.showInformationMessage(
            `Expense "${expenseData.description}" restored successfully!`
          );
          this.refreshCallback?.();
        } else {
          vscode.window.showErrorMessage(
            `Failed to restore expense: ${result.error}`
          );
        }
      }
    );
  }

  /**
   * View expense report
   */
  async viewExpenseReport(): Promise<void> {
    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Generating expense report...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.getExpenseReport();

        if (result.success && result.stdout) {
          // Show output in a new document
          const doc = await vscode.workspace.openTextDocument({
            content: result.stdout,
            language: 'plaintext',
          });
          await vscode.window.showTextDocument(doc);
        } else {
          vscode.window.showErrorMessage(
            `Failed to generate report: ${result.error}`
          );
        }
      }
    );
  }
}
