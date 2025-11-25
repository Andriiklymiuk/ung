import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';

/**
 * Invoice detail webview panel
 */
export class InvoicePanel {
    public static currentPanel: InvoicePanel | undefined;
    private readonly panel: vscode.WebviewPanel;
    private disposables: vscode.Disposable[] = [];

    private constructor(
        panel: vscode.WebviewPanel,
        private cli: UngCli,
        private invoiceId: number
    ) {
        this.panel = panel;

        // Set the webview's HTML content
        this.update();

        // Handle messages from the webview
        this.panel.webview.onDidReceiveMessage(
            message => {
                switch (message.command) {
                    case 'export':
                        this.exportPDF();
                        break;
                    case 'email':
                        this.emailInvoice();
                        break;
                }
            },
            null,
            this.disposables
        );

        // Handle panel disposed
        this.panel.onDidDispose(() => this.dispose(), null, this.disposables);
    }

    /**
     * Create or show invoice panel
     */
    public static createOrShow(cli: UngCli, invoiceId: number) {
        const column = vscode.window.activeTextEditor
            ? vscode.window.activeTextEditor.viewColumn
            : undefined;

        // If we already have a panel, show it
        if (InvoicePanel.currentPanel) {
            InvoicePanel.currentPanel.panel.reveal(column);
            InvoicePanel.currentPanel.invoiceId = invoiceId;
            InvoicePanel.currentPanel.update();
            return;
        }

        // Otherwise, create a new panel
        const panel = vscode.window.createWebviewPanel(
            'ungInvoice',
            `Invoice ${invoiceId}`,
            column || vscode.ViewColumn.One,
            {
                enableScripts: true,
                retainContextWhenHidden: true
            }
        );

        InvoicePanel.currentPanel = new InvoicePanel(panel, cli, invoiceId);
    }

    /**
     * Update webview content
     */
    private async update() {
        this.panel.title = `Invoice ${this.invoiceId}`;
        this.panel.webview.html = await this.getHtmlForWebview();
    }

    /**
     * Export invoice to PDF
     */
    private async exportPDF() {
        const result = await this.cli.generateInvoicePDF(this.invoiceId);
        if (result.success) {
            vscode.window.showInformationMessage('Invoice PDF generated successfully!');
        } else {
            vscode.window.showErrorMessage(`Failed to generate PDF: ${result.error}`);
        }
    }

    /**
     * Email invoice
     */
    private async emailInvoice() {
        // Ask user to select email client
        const emailClients = [
            { label: 'Apple Mail', value: 'apple' },
            { label: 'Outlook', value: 'outlook' },
            { label: 'Gmail (Browser)', value: 'gmail' }
        ];

        const selected = await vscode.window.showQuickPick(emailClients, {
            placeHolder: 'Select email client'
        });

        if (!selected) {
            return;
        }

        const result = await this.cli.emailInvoice(this.invoiceId, selected.value);
        if (result.success) {
            vscode.window.showInformationMessage('Invoice email prepared!');
        } else {
            vscode.window.showErrorMessage(`Failed to email invoice: ${result.error}`);
        }
    }

    /**
     * Get HTML content for webview
     */
    private async getHtmlForWebview(): Promise<string> {
        return `<!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <title>Invoice ${this.invoiceId}</title>
            <style>
                body {
                    font-family: var(--vscode-font-family);
                    color: var(--vscode-foreground);
                    padding: 20px;
                }
                .container {
                    max-width: 800px;
                    margin: 0 auto;
                }
                h1 {
                    color: var(--vscode-titleBar-activeForeground);
                    border-bottom: 2px solid var(--vscode-panel-border);
                    padding-bottom: 10px;
                }
                .invoice-details {
                    margin: 20px 0;
                    padding: 15px;
                    background-color: var(--vscode-editor-background);
                    border: 1px solid var(--vscode-panel-border);
                    border-radius: 4px;
                }
                .detail-row {
                    display: flex;
                    justify-content: space-between;
                    margin: 10px 0;
                }
                .label {
                    font-weight: bold;
                    color: var(--vscode-descriptionForeground);
                }
                .actions {
                    margin-top: 20px;
                    display: flex;
                    gap: 10px;
                }
                button {
                    background-color: var(--vscode-button-background);
                    color: var(--vscode-button-foreground);
                    border: none;
                    padding: 10px 20px;
                    cursor: pointer;
                    border-radius: 2px;
                }
                button:hover {
                    background-color: var(--vscode-button-hoverBackground);
                }
            </style>
        </head>
        <body>
            <div class="container">
                <h1>Invoice ${this.invoiceId}</h1>

                <div class="invoice-details">
                    <div class="detail-row">
                        <span class="label">Invoice Number:</span>
                        <span>Loading...</span>
                    </div>
                    <div class="detail-row">
                        <span class="label">Amount:</span>
                        <span>Loading...</span>
                    </div>
                    <div class="detail-row">
                        <span class="label">Status:</span>
                        <span>Loading...</span>
                    </div>
                    <div class="detail-row">
                        <span class="label">Due Date:</span>
                        <span>Loading...</span>
                    </div>
                </div>

                <div class="actions">
                    <button onclick="exportPDF()">Export to PDF</button>
                    <button onclick="emailInvoice()">Email Invoice</button>
                </div>

                <p style="margin-top: 30px; color: var(--vscode-descriptionForeground);">
                    <em>Note: Detailed invoice editing and line items will be available in a future version.</em>
                </p>
            </div>

            <script>
                const vscode = acquireVsCodeApi();

                function exportPDF() {
                    vscode.postMessage({ command: 'export' });
                }

                function emailInvoice() {
                    vscode.postMessage({ command: 'email' });
                }
            </script>
        </body>
        </html>`;
    }

    /**
     * Dispose resources
     */
    public dispose() {
        InvoicePanel.currentPanel = undefined;

        this.panel.dispose();

        while (this.disposables.length) {
            const disposable = this.disposables.pop();
            if (disposable) {
                disposable.dispose();
            }
        }
    }
}
