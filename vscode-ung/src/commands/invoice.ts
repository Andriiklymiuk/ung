import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';
import { Config } from '../utils/config';
import { Formatter } from '../utils/formatting';

/**
 * Invoice command handlers
 */
export class InvoiceCommands {
    constructor(private cli: UngCli, private refreshCallback?: () => void) {}

    /**
     * Create a new invoice
     */
    async createInvoice(): Promise<void> {
        // Get company ID (assume 1 for now)
        const companyId = 1;

        // Get client list to select from
        const clientResult = await this.cli.listClients();
        if (!clientResult.success) {
            vscode.window.showErrorMessage('Failed to fetch clients');
            return;
        }

        // For simplicity, ask for client ID
        const clientIdStr = await vscode.window.showInputBox({
            prompt: 'Client ID',
            placeHolder: 'Enter client ID (use "ung client ls" to see list)',
            validateInput: (value) => {
                if (!value) return 'Client ID is required';
                if (isNaN(Number(value))) return 'Must be a number';
                return null;
            }
        });

        if (!clientIdStr) return;
        const clientId = Number(clientIdStr);

        const amountStr = await vscode.window.showInputBox({
            prompt: 'Invoice Amount',
            placeHolder: 'e.g., 1500.00',
            validateInput: (value) => {
                if (!value) return 'Amount is required';
                if (isNaN(Number(value))) return 'Must be a valid number';
                if (Number(value) <= 0) return 'Amount must be greater than 0';
                return null;
            }
        });

        if (!amountStr) return;
        const amount = Number(amountStr);

        const currency = await vscode.window.showQuickPick(
            ['USD', 'EUR', 'GBP', 'UAH', 'CAD'],
            {
                placeHolder: 'Select currency',
                canPickMany: false
            }
        ) || Config.getDefaultCurrency();

        const description = await vscode.window.showInputBox({
            prompt: 'Description (optional)',
            placeHolder: 'e.g., Development services'
        });

        const dueDateStr = await vscode.window.showInputBox({
            prompt: 'Due Date (optional)',
            placeHolder: 'YYYY-MM-DD (leave empty for 30 days from now)'
        });

        // Show progress
        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Creating invoice...',
            cancellable: false
        }, async () => {
            const result = await this.cli.createInvoice({
                companyId,
                clientId,
                amount,
                currency,
                description,
                dueDate: dueDateStr
            });

            if (result.success) {
                vscode.window.showInformationMessage('Invoice created successfully!');
                if (this.refreshCallback) {
                    this.refreshCallback();
                }
            } else {
                vscode.window.showErrorMessage(`Failed to create invoice: ${result.error}`);
            }
        });
    }

    /**
     * Generate invoice from time tracking
     */
    async generateFromTime(): Promise<void> {
        const clientName = await vscode.window.showInputBox({
            prompt: 'Client Name',
            placeHolder: 'e.g., acme',
            validateInput: (value) => value ? null : 'Client name is required'
        });

        if (!clientName) return;

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Generating invoice from time tracking...',
            cancellable: false
        }, async () => {
            const result = await this.cli.generateInvoiceFromTime(clientName);

            if (result.success) {
                vscode.window.showInformationMessage('Invoice generated from time tracking!');
                if (this.refreshCallback) {
                    this.refreshCallback();
                }
            } else {
                vscode.window.showErrorMessage(`Failed to generate invoice: ${result.error}`);
            }
        });
    }

    /**
     * View invoice details
     */
    async viewInvoice(invoiceId?: number): Promise<void> {
        if (!invoiceId) {
            vscode.window.showErrorMessage('No invoice selected');
            return;
        }

        vscode.window.showInformationMessage(`Invoice ID: ${invoiceId}. Detailed view coming soon!`);
    }

    /**
     * Edit an invoice
     */
    async editInvoice(invoiceId?: number): Promise<void> {
        if (!invoiceId) {
            vscode.window.showErrorMessage('No invoice selected');
            return;
        }

        vscode.window.showInformationMessage(
            'Invoice editing will be available in a future version.'
        );
    }

    /**
     * Delete an invoice
     */
    async deleteInvoice(invoiceId?: number): Promise<void> {
        if (!invoiceId) {
            vscode.window.showErrorMessage('No invoice selected');
            return;
        }

        const confirm = await vscode.window.showWarningMessage(
            `Are you sure you want to delete this invoice?`,
            { modal: true },
            'Yes', 'No'
        );

        if (confirm !== 'Yes') return;

        vscode.window.showInformationMessage('Invoice deletion will be available in a future version.');
    }

    /**
     * Export invoice to PDF
     */
    async exportInvoice(invoiceId?: number): Promise<void> {
        if (!invoiceId) {
            vscode.window.showErrorMessage('No invoice selected');
            return;
        }

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Generating invoice PDF...',
            cancellable: false
        }, async () => {
            const result = await this.cli.generateInvoicePDF(invoiceId);

            if (result.success) {
                vscode.window.showInformationMessage('Invoice PDF generated successfully!');
            } else {
                vscode.window.showErrorMessage(`Failed to generate PDF: ${result.error}`);
            }
        });
    }

    /**
     * Email invoice
     */
    async emailInvoice(invoiceId?: number): Promise<void> {
        if (!invoiceId) {
            vscode.window.showErrorMessage('No invoice selected');
            return;
        }

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Preparing invoice email...',
            cancellable: false
        }, async () => {
            const result = await this.cli.emailInvoice(invoiceId);

            if (result.success) {
                vscode.window.showInformationMessage('Invoice email prepared!');
            } else {
                vscode.window.showErrorMessage(`Failed to email invoice: ${result.error}`);
            }
        });
    }
}
