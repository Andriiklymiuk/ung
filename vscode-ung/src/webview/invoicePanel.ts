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
                :root {
                    --bg-primary: var(--vscode-editor-background);
                    --bg-secondary: var(--vscode-sideBar-background);
                    --bg-tertiary: var(--vscode-input-background);
                    --text-primary: var(--vscode-editor-foreground);
                    --text-secondary: var(--vscode-descriptionForeground);
                    --border: var(--vscode-panel-border);
                    --accent: var(--vscode-button-background);
                    --accent-hover: var(--vscode-button-hoverBackground);
                    --success: #4caf50;
                    --warning: #ff9800;
                    --danger: #f44336;
                    --info: #2196f3;
                }

                * { box-sizing: border-box; margin: 0; padding: 0; }

                body {
                    font-family: var(--vscode-font-family);
                    background: var(--bg-primary);
                    color: var(--text-primary);
                    padding: 24px;
                    line-height: 1.6;
                }

                .container {
                    max-width: 800px;
                    margin: 0 auto;
                }

                .header {
                    display: flex;
                    justify-content: space-between;
                    align-items: flex-start;
                    margin-bottom: 32px;
                    padding-bottom: 20px;
                    border-bottom: 1px solid var(--border);
                    flex-wrap: wrap;
                    gap: 16px;
                }

                .header-info {
                    display: flex;
                    align-items: center;
                    gap: 16px;
                }

                .invoice-icon {
                    width: 56px;
                    height: 56px;
                    background: linear-gradient(135deg, var(--accent), var(--info));
                    border-radius: 12px;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    font-size: 24px;
                    color: white;
                }

                .header-text h1 {
                    font-size: 28px;
                    font-weight: 700;
                    margin-bottom: 4px;
                }

                .header-text .subtitle {
                    color: var(--text-secondary);
                    font-size: 14px;
                }

                .actions {
                    display: flex;
                    gap: 10px;
                    flex-wrap: wrap;
                }

                button {
                    background: var(--accent);
                    color: var(--vscode-button-foreground);
                    border: none;
                    padding: 12px 20px;
                    border-radius: 8px;
                    cursor: pointer;
                    font-size: 13px;
                    font-weight: 500;
                    transition: all 0.2s ease;
                    display: flex;
                    align-items: center;
                    gap: 8px;
                }

                button:hover {
                    background: var(--accent-hover);
                    transform: translateY(-1px);
                }

                button.secondary {
                    background: var(--bg-secondary);
                    border: 1px solid var(--border);
                    color: var(--text-primary);
                }

                button.secondary:hover {
                    background: var(--bg-tertiary);
                }

                .card {
                    background: var(--bg-secondary);
                    border: 1px solid var(--border);
                    border-radius: 12px;
                    padding: 24px;
                    margin-bottom: 20px;
                }

                .card-header {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    margin-bottom: 20px;
                    padding-bottom: 12px;
                    border-bottom: 1px solid var(--border);
                }

                .card-header h2 {
                    font-size: 16px;
                    font-weight: 600;
                    display: flex;
                    align-items: center;
                    gap: 8px;
                }

                .detail-grid {
                    display: grid;
                    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
                    gap: 20px;
                }

                .detail-item {
                    padding: 16px;
                    background: var(--bg-tertiary);
                    border-radius: 8px;
                    transition: all 0.2s ease;
                }

                .detail-item:hover {
                    background: var(--bg-primary);
                }

                .detail-label {
                    font-size: 11px;
                    color: var(--text-secondary);
                    text-transform: uppercase;
                    letter-spacing: 0.5px;
                    margin-bottom: 6px;
                }

                .detail-value {
                    font-size: 18px;
                    font-weight: 600;
                }

                .detail-value.loading {
                    color: var(--text-secondary);
                    font-style: italic;
                    font-weight: normal;
                }

                .badge {
                    display: inline-block;
                    padding: 6px 12px;
                    border-radius: 20px;
                    font-size: 12px;
                    font-weight: 500;
                    text-transform: uppercase;
                    letter-spacing: 0.5px;
                }

                .badge.draft { background: rgba(158, 158, 158, 0.2); color: #9e9e9e; }
                .badge.pending { background: rgba(255, 152, 0, 0.2); color: var(--warning); }
                .badge.sent { background: rgba(33, 150, 243, 0.2); color: var(--info); }
                .badge.paid { background: rgba(76, 175, 80, 0.2); color: var(--success); }
                .badge.overdue { background: rgba(244, 67, 54, 0.2); color: var(--danger); }

                .amount-highlight {
                    font-size: 32px;
                    font-weight: 700;
                    color: var(--success);
                }

                .footer-note {
                    margin-top: 32px;
                    padding: 16px;
                    background: var(--bg-secondary);
                    border-radius: 8px;
                    border-left: 4px solid var(--info);
                }

                .footer-note p {
                    color: var(--text-secondary);
                    font-size: 13px;
                    margin: 0;
                    display: flex;
                    align-items: center;
                    gap: 8px;
                }

                @media (max-width: 600px) {
                    body { padding: 16px; }
                    .header { flex-direction: column; }
                    .actions { width: 100%; }
                    .actions button { flex: 1; justify-content: center; }
                    .detail-grid { grid-template-columns: 1fr; }
                    .amount-highlight { font-size: 24px; }
                }
            </style>
        </head>
        <body>
            <div class="container">
                <div class="header">
                    <div class="header-info">
                        <div class="invoice-icon">📄</div>
                        <div class="header-text">
                            <h1>Invoice #${this.invoiceId}</h1>
                            <span class="subtitle">View and manage invoice details</span>
                        </div>
                    </div>
                    <div class="actions">
                        <button onclick="exportPDF()">
                            <span>📥</span> Export PDF
                        </button>
                        <button class="secondary" onclick="emailInvoice()">
                            <span>📧</span> Send Email
                        </button>
                    </div>
                </div>

                <div class="card">
                    <div class="card-header">
                        <h2>💰 Invoice Details</h2>
                    </div>
                    <div class="detail-grid">
                        <div class="detail-item">
                            <div class="detail-label">Invoice Number</div>
                            <div class="detail-value">#${this.invoiceId}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Amount</div>
                            <div class="detail-value amount-highlight loading">Loading...</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Status</div>
                            <div class="detail-value">
                                <span class="badge pending">Loading...</span>
                            </div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Due Date</div>
                            <div class="detail-value loading">Loading...</div>
                        </div>
                    </div>
                </div>

                <div class="card">
                    <div class="card-header">
                        <h2>👤 Client Information</h2>
                    </div>
                    <div class="detail-grid">
                        <div class="detail-item">
                            <div class="detail-label">Client Name</div>
                            <div class="detail-value loading">Loading...</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Email</div>
                            <div class="detail-value loading">Loading...</div>
                        </div>
                    </div>
                </div>

                <div class="footer-note">
                    <p>
                        <span>💡</span>
                        Detailed invoice editing and line items will be available in a future version.
                    </p>
                </div>
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
