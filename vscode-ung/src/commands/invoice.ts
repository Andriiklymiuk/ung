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
