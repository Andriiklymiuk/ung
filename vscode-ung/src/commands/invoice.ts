import * as vscode from 'vscode';
import * as os from 'os';
import * as path from 'path';
import { UngCli } from '../cli/ungCli';

/**
 * Invoice command handlers
 */
export class InvoiceCommands {
    constructor(private cli: UngCli, private refreshCallback?: () => void) {}

    /**
     * Create a new invoice
     */
    async createInvoice(): Promise<void> {
        // Get contracts to select from (they have client, rate, and currency info)
        const contractResult = await this.cli.listContracts();
        if (!contractResult.success) {
            vscode.window.showErrorMessage('Failed to fetch contracts');
            return;
        }

        // Parse contracts from CLI output
        const contracts = this.parseContractsFromOutput(contractResult.stdout || '');
        if (contracts.length === 0) {
            vscode.window.showErrorMessage('No contracts found. Create one first with "ung contract add"');
            return;
        }

        // Show contract dropdown with client name and rate info
        const contractItems = contracts.map(c => ({
            label: c.client,
            description: `${c.type} - ${c.ratePrice}`,
            detail: c.name,
            contract: c
        }));

        const selectedContract = await vscode.window.showQuickPick(contractItems, {
            placeHolder: 'Select a contract',
            matchOnDescription: true,
            matchOnDetail: true
        });

        if (!selectedContract) return;

        const contract = selectedContract.contract;

        // Parse currency and amount from ratePrice (e.g., "4500.00 USD" or "29.00 USD/hr")
        let currency = 'USD';
        const rateMatch = contract.ratePrice.match(/^([\d.]+)\s*(\w+)/);
        if (rateMatch) {
            currency = rateMatch[2].replace('/hr', '');
        }

        let amount: number;

        if (contract.type === 'fixed_price') {
            // Use the fixed price directly, just confirm
            amount = parseFloat(rateMatch?.[1] || '0');

            const confirmAmount = await vscode.window.showInputBox({
                prompt: `Invoice amount for ${contract.client}`,
                value: amount.toString(),
                placeHolder: 'e.g., 4500.00',
                validateInput: (value) => {
                    if (!value) return 'Amount is required';
                    if (isNaN(Number(value))) return 'Must be a valid number';
                    if (Number(value) <= 0) return 'Amount must be greater than 0';
                    return null;
                }
            });

            if (!confirmAmount) return;
            amount = Number(confirmAmount);
        } else {
            // Hourly contract - ask for amount
            const amountStr = await vscode.window.showInputBox({
                prompt: `Invoice amount for ${contract.client} (rate: ${contract.ratePrice})`,
                placeHolder: 'e.g., 1500.00',
                validateInput: (value) => {
                    if (!value) return 'Amount is required';
                    if (isNaN(Number(value))) return 'Must be a valid number';
                    if (Number(value) <= 0) return 'Amount must be greater than 0';
                    return null;
                }
            });

            if (!amountStr) return;
            amount = Number(amountStr);
        }

        // Show progress
        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Creating invoice...',
            cancellable: false
        }, async () => {
            const result = await this.cli.createInvoice({
                clientName: contract.client,
                amount,
                currency
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

        // Ask what actions to take
        const actions = await vscode.window.showQuickPick([
            { label: 'Generate Invoice + PDF', value: 'pdf' },
            { label: 'Generate Invoice + PDF + Email', value: 'email' },
            { label: 'Generate Invoice Only', value: 'none' }
        ], {
            placeHolder: 'Select action'
        });

        if (!actions) return;

        let emailApp: string | undefined;
        if (actions.value === 'email') {
            const emailClients = [
                { label: 'Apple Mail', value: 'apple' },
                { label: 'Outlook', value: 'outlook' },
                { label: 'Gmail (Browser)', value: 'gmail' }
            ];

            const selected = await vscode.window.showQuickPick(emailClients, {
                placeHolder: 'Select email client'
            });

            if (!selected) return;
            emailApp = selected.value;
        }

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Generating invoice from time tracking...',
            cancellable: false
        }, async () => {
            const result = await this.cli.generateInvoiceFromTime(clientName, {
                pdf: actions.value === 'pdf' || actions.value === 'email',
                email: actions.value === 'email',
                emailApp
            });

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
     * View invoice details - shows action menu like contracts
     */
    async viewInvoice(invoiceId?: number): Promise<void> {
        if (!invoiceId) {
            vscode.window.showErrorMessage('No invoice selected');
            return;
        }

        // Show action menu like contracts
        const actions = [
            { label: '$(file-pdf) Export to PDF', action: 'pdf' },
            { label: '$(mail) Email Invoice', action: 'email' },
            { label: '$(check) Mark as Paid', action: 'markPaid' },
            { label: '$(send) Mark as Sent', action: 'markSent' },
            { label: '$(list-selection) Change Status...', action: 'changeStatus' },
            { label: '$(copy) Duplicate Invoice', action: 'duplicate' },
            { label: '$(close) Close', action: 'close' }
        ];

        const selected = await vscode.window.showQuickPick(actions, {
            placeHolder: `Invoice ${invoiceId}`,
            title: 'Invoice Actions'
        });

        if (selected) {
            switch (selected.action) {
                case 'pdf':
                    await this.exportInvoice(invoiceId);
                    break;
                case 'email':
                    await this.emailInvoice(invoiceId);
                    break;
                case 'markPaid':
                    await this.markAsPaid(invoiceId);
                    break;
                case 'markSent':
                    await this.markAsSent(invoiceId);
                    break;
                case 'changeStatus':
                    await this.changeInvoiceStatus(invoiceId);
                    break;
                case 'duplicate':
                    await this.duplicateInvoice(invoiceId);
                    break;
            }
        }
    }

    /**
     * Duplicate an existing invoice
     */
    async duplicateInvoice(invoiceId?: number): Promise<void> {
        if (!invoiceId) {
            // Show invoice list to select
            const result = await this.cli.listInvoices();
            if (!result.success || !result.stdout) {
                vscode.window.showErrorMessage('Failed to fetch invoices');
                return;
            }

            const invoices = this.parseInvoicesFromOutput(result.stdout);
            if (invoices.length === 0) {
                vscode.window.showInformationMessage('No invoices found to duplicate');
                return;
            }

            const selected = await vscode.window.showQuickPick(
                invoices.map(inv => ({
                    label: inv.invoiceNum,
                    description: `${inv.client} - ${inv.amount}`,
                    detail: inv.status,
                    id: inv.id
                })),
                { placeHolder: 'Select an invoice to duplicate' }
            );

            if (!selected) return;
            invoiceId = selected.id;
        }

        // Get invoice details
        const invoicesResult = await this.cli.listInvoices();
        if (!invoicesResult.success || !invoicesResult.stdout) {
            vscode.window.showErrorMessage('Failed to fetch invoice details');
            return;
        }

        const invoices = this.parseInvoicesFromOutput(invoicesResult.stdout);
        const originalInvoice = invoices.find(inv => inv.id === invoiceId);

        if (!originalInvoice) {
            vscode.window.showErrorMessage('Invoice not found');
            return;
        }

        // Ask for new amount (default to original)
        const amountStr = await vscode.window.showInputBox({
            prompt: `Amount for new invoice (original: ${originalInvoice.amount})`,
            value: originalInvoice.amount.replace(/[^0-9.]/g, ''),
            validateInput: value => {
                if (!value) return 'Amount is required';
                if (isNaN(Number(value))) return 'Must be a valid number';
                if (Number(value) <= 0) return 'Amount must be greater than 0';
                return null;
            }
        });

        if (!amountStr) return;

        // Parse currency from original
        const currencyMatch = originalInvoice.amount.match(/(USD|EUR|GBP|CHF|PLN)/);
        const currency = currencyMatch ? currencyMatch[1] : 'USD';

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Creating duplicate invoice...',
            cancellable: false
        }, async () => {
            const result = await this.cli.createInvoice({
                clientName: originalInvoice.client,
                amount: Number(amountStr),
                currency
            });

            if (result.success) {
                vscode.window.showInformationMessage(`Invoice duplicated! New invoice created for ${originalInvoice.client}`);
                if (this.refreshCallback) {
                    this.refreshCallback();
                }
            } else {
                vscode.window.showErrorMessage(`Failed to duplicate invoice: ${result.error}`);
            }
        });
    }

    /**
     * Parse invoices from CLI output
     */
    private parseInvoicesFromOutput(output: string): Array<{
        id: number;
        invoiceNum: string;
        client: string;
        date: string;
        amount: string;
        status: string;
    }> {
        const lines = output.trim().split('\n');
        if (lines.length < 2) return [];

        return lines.slice(1).map(line => {
            const parts = line.split(/\s{2,}/).filter(p => p.trim());
            const id = parseInt(parts[0], 10);
            if (isNaN(id)) return null;

            return {
                id,
                invoiceNum: parts[1] || '',
                client: parts[2] || 'Unknown',
                date: parts[3] || '',
                amount: parts[4] || '',
                status: parts[5] || ''
            };
        }).filter((inv): inv is NonNullable<typeof inv> => inv !== null);
    }

    /**
     * Mark invoice as paid
     */
    async markAsPaid(invoiceId?: number): Promise<void> {
        if (!invoiceId) {
            vscode.window.showErrorMessage('No invoice selected');
            return;
        }

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Marking invoice as paid...',
            cancellable: false
        }, async () => {
            const result = await this.cli.markInvoicePaid(invoiceId);

            if (result.success) {
                vscode.window.showInformationMessage('Invoice marked as paid!');
                if (this.refreshCallback) {
                    this.refreshCallback();
                }
            } else {
                vscode.window.showErrorMessage(`Failed to update invoice: ${result.error}`);
            }
        });
    }

    /**
     * Mark invoice as sent
     */
    async markAsSent(invoiceId?: number): Promise<void> {
        if (!invoiceId) {
            vscode.window.showErrorMessage('No invoice selected');
            return;
        }

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Marking invoice as sent...',
            cancellable: false
        }, async () => {
            const result = await this.cli.markInvoiceSent(invoiceId);

            if (result.success) {
                vscode.window.showInformationMessage('Invoice marked as sent!');
                if (this.refreshCallback) {
                    this.refreshCallback();
                }
            } else {
                vscode.window.showErrorMessage(`Failed to update invoice: ${result.error}`);
            }
        });
    }

    /**
     * Change invoice status
     */
    async changeInvoiceStatus(invoiceId?: number): Promise<void> {
        if (!invoiceId) {
            vscode.window.showErrorMessage('No invoice selected');
            return;
        }

        const statuses = [
            { label: '$(clock) Pending', value: 'pending' as const, description: 'Invoice not yet sent' },
            { label: '$(send) Sent', value: 'sent' as const, description: 'Invoice has been sent to client' },
            { label: '$(check) Paid', value: 'paid' as const, description: 'Payment received' },
            { label: '$(warning) Overdue', value: 'overdue' as const, description: 'Past due date' }
        ];

        const selected = await vscode.window.showQuickPick(statuses, {
            placeHolder: 'Select new status',
            title: 'Change Invoice Status'
        });

        if (!selected) return;

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: `Updating invoice status to ${selected.value}...`,
            cancellable: false
        }, async () => {
            const result = await this.cli.updateInvoiceStatus(invoiceId, selected.value);

            if (result.success) {
                vscode.window.showInformationMessage(`Invoice status updated to ${selected.value}!`);
                if (this.refreshCallback) {
                    this.refreshCallback();
                }
            } else {
                vscode.window.showErrorMessage(`Failed to update invoice: ${result.error}`);
            }
        });
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

        let pdfPath: string | undefined;

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Generating invoice PDF...',
            cancellable: false
        }, async () => {
            const result = await this.cli.generateInvoicePDF(invoiceId);

            if (result.success) {
                pdfPath = this.parsePDFPath(result.stdout || '');
            } else {
                vscode.window.showErrorMessage(`Failed to generate PDF: ${result.error}`);
            }
        });

        if (pdfPath) {
            // Auto-open the PDF
            await vscode.env.openExternal(vscode.Uri.file(pdfPath));

            // Also show notification with buttons
            const action = await vscode.window.showInformationMessage(
                `Invoice PDF: ${pdfPath}`,
                'Open Again',
                'Show in Finder'
            );

            if (action === 'Open Again') {
                await vscode.env.openExternal(vscode.Uri.file(pdfPath));
            } else if (action === 'Show in Finder') {
                await vscode.commands.executeCommand('revealFileInOS', vscode.Uri.file(pdfPath));
            }
        } else {
            // Fallback if path parsing fails
            const invoicesDir = path.join(os.homedir(), '.ung', 'invoices');
            const action = await vscode.window.showInformationMessage(
                'Invoice PDF generated!',
                'Open Invoices Folder'
            );
            if (action === 'Open Invoices Folder') {
                await vscode.env.openExternal(vscode.Uri.file(invoicesDir));
            }
        }
    }

    /**
     * Get email client options based on platform
     */
    private getEmailClients(): Array<{ label: string; value: string; description?: string }> {
        const platform = process.platform;

        // Common options for all platforms
        const common = [
            { label: '$(globe) Gmail (Browser)', value: 'gmail', description: 'Opens in web browser' },
            { label: '$(globe) Outlook Web', value: 'outlook-web', description: 'Opens in web browser' }
        ];

        // Platform-specific options
        if (platform === 'darwin') {
            return [
                { label: '$(mail) Apple Mail', value: 'apple', description: 'Default macOS mail app' },
                { label: '$(mail) Outlook', value: 'outlook', description: 'Microsoft Outlook app' },
                ...common
            ];
        } else if (platform === 'win32') {
            return [
                { label: '$(mail) Windows Mail', value: 'windows-mail', description: 'Default Windows mail app' },
                { label: '$(mail) Outlook', value: 'outlook', description: 'Microsoft Outlook app' },
                { label: '$(mail) Thunderbird', value: 'thunderbird', description: 'Mozilla Thunderbird' },
                ...common
            ];
        } else {
            // Linux
            return [
                { label: '$(mail) Thunderbird', value: 'thunderbird', description: 'Mozilla Thunderbird' },
                { label: '$(mail) Evolution', value: 'evolution', description: 'GNOME Evolution' },
                ...common
            ];
        }
    }

    /**
     * Email invoice
     */
    async emailInvoice(invoiceId?: number): Promise<void> {
        if (!invoiceId) {
            vscode.window.showErrorMessage('No invoice selected');
            return;
        }

        // Get platform-specific email clients
        const emailClients = this.getEmailClients();

        const selected = await vscode.window.showQuickPick(emailClients, {
            placeHolder: 'Select email client',
            title: 'Choose Email Application'
        });

        if (!selected) {
            return;
        }

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Preparing invoice email...',
            cancellable: false
        }, async () => {
            const result = await this.cli.emailInvoice(invoiceId, selected.value);

            if (result.success) {
                vscode.window.showInformationMessage('Invoice email prepared!');
            } else {
                vscode.window.showErrorMessage(`Failed to email invoice: ${result.error}`);
            }
        });
    }

    /**
     * Open invoice PDF in external viewer
     */
    async openInvoicePDF(invoiceNum?: string): Promise<void> {
        if (!invoiceNum) {
            invoiceNum = await vscode.window.showInputBox({
                prompt: 'Enter invoice number',
                placeHolder: 'INV-001'
            });
        }

        if (!invoiceNum) {
            return;
        }

        try {
            let pdfPath: string | undefined;

            // Generate PDF first
            await vscode.window.withProgress({
                location: vscode.ProgressLocation.Notification,
                title: `Generating PDF for ${invoiceNum}...`,
                cancellable: false
            }, async () => {
                const result = await this.cli.exec(['invoice', 'pdf', invoiceNum]);
                if (!result.success) {
                    throw new Error(result.error || 'Failed to generate PDF');
                }

                // Parse PDF path from output and open it
                pdfPath = this.parsePDFPath(result.stdout || '');
            });

            if (pdfPath) {
                const uri = vscode.Uri.file(pdfPath);
                await vscode.env.openExternal(uri);
                vscode.window.showInformationMessage(`Opened ${invoiceNum} in external viewer`);
            } else {
                vscode.window.showWarningMessage('Could not locate generated PDF file');
            }
        } catch (error) {
            vscode.window.showErrorMessage(`Failed to open PDF: ${error}`);
        }
    }

    /**
     * Parse PDF path from CLI output
     */
    private parsePDFPath(output: string): string | undefined {
        // Match: "✓ PDF generated successfully: /path/to/file.pdf" or "PDF saved to: /path/to/file.pdf"
        const match = output.match(/PDF (?:generated successfully|saved to):\s*(.+\.pdf)/i);
        return match ? match[1].trim() : undefined;
    }

    /**
     * Parse contracts from CLI output (tabular format)
     */
    private parseContractsFromOutput(output: string): Array<{
        id: number;
        contractNum: string;
        name: string;
        client: string;
        type: string;
        ratePrice: string;
        active: boolean;
    }> {
        const lines = output.trim().split('\n');
        if (lines.length < 2) return [];

        // Skip header line
        const dataLines = lines.slice(1);
        const contracts: Array<{
            id: number;
            contractNum: string;
            name: string;
            client: string;
            type: string;
            ratePrice: string;
            active: boolean;
        }> = [];

        for (const line of dataLines) {
            // Parse: ID  CONTRACT#  NAME  CLIENT  TYPE  RATE/PRICE  ACTIVE
            const parts = line.split(/\s{2,}/).map(p => p.trim()).filter(p => p);
            if (parts.length >= 6) {
                const id = parseInt(parts[0], 10);
                if (!isNaN(id)) {
                    contracts.push({
                        id,
                        contractNum: parts[1],
                        name: parts[2],
                        client: parts[3],
                        type: parts[4],
                        ratePrice: parts[5],
                        active: parts[6] === '✓'
                    });
                }
            }
        }

        return contracts;
    }
}
