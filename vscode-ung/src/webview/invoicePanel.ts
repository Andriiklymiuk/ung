import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';

/**
 * Invoice data structure
 */
interface InvoiceData {
  id: number;
  invoiceNum: string;
  amount: number;
  currency: string;
  status: string;
  dueDate: string;
  clientName: string;
  clientEmail: string;
  description?: string;
}

/**
 * Invoice detail webview panel
 */
export class InvoicePanel {
  public static currentPanel: InvoicePanel | undefined;
  private readonly panel: vscode.WebviewPanel;
  private disposables: vscode.Disposable[] = [];
  private invoiceData: InvoiceData | null = null;
  private refreshCallback?: () => void;

  private constructor(
    panel: vscode.WebviewPanel,
    private cli: UngCli,
    private invoiceId: number,
    refreshCallback?: () => void
  ) {
    this.panel = panel;
    this.refreshCallback = refreshCallback;

    // Set the webview's HTML content
    this.update();

    // Handle messages from the webview
    this.panel.webview.onDidReceiveMessage(
      (message) => {
        switch (message.command) {
          case 'export':
            this.exportPDF();
            break;
          case 'email':
            this.emailInvoice();
            break;
          case 'edit':
            this.editInvoice(message.field, message.value);
            break;
          case 'markPaid':
            this.markAsPaid();
            break;
          case 'markSent':
            this.markAsSent();
            break;
          case 'delete':
            this.deleteInvoice();
            break;
          case 'duplicate':
            this.duplicateInvoice();
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
  public static createOrShow(
    cli: UngCli,
    invoiceId: number,
    refreshCallback?: () => void
  ) {
    const column = vscode.window.activeTextEditor
      ? vscode.window.activeTextEditor.viewColumn
      : undefined;

    // If we already have a panel, show it
    if (InvoicePanel.currentPanel) {
      InvoicePanel.currentPanel.panel.reveal(column);
      InvoicePanel.currentPanel.invoiceId = invoiceId;
      InvoicePanel.currentPanel.refreshCallback = refreshCallback;
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
        retainContextWhenHidden: true,
      }
    );

    InvoicePanel.currentPanel = new InvoicePanel(
      panel,
      cli,
      invoiceId,
      refreshCallback
    );
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
      vscode.window.showInformationMessage(
        'Invoice PDF generated successfully!'
      );
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
      { label: 'Gmail (Browser)', value: 'gmail' },
    ];

    const selected = await vscode.window.showQuickPick(emailClients, {
      placeHolder: 'Select email client',
    });

    if (!selected) {
      return;
    }

    const result = await this.cli.emailInvoice(this.invoiceId, selected.value);
    if (result.success) {
      vscode.window.showInformationMessage('Invoice email prepared!');
    } else {
      vscode.window.showErrorMessage(
        `Failed to email invoice: ${result.error}`
      );
    }
  }

  /**
   * Edit invoice field
   */
  private async editInvoice(field: string, currentValue: string) {
    let newValue: string | undefined;
    const updates: { amount?: number; dueDate?: string; description?: string } =
      {};

    switch (field) {
      case 'amount': {
        newValue = await vscode.window.showInputBox({
          prompt: 'Enter new amount',
          value: currentValue,
          validateInput: (v) => {
            const num = parseFloat(v);
            return Number.isNaN(num) || num <= 0
              ? 'Please enter a valid positive number'
              : null;
          },
        });
        if (newValue) updates.amount = parseFloat(newValue);
        break;
      }
      case 'dueDate': {
        newValue = await vscode.window.showInputBox({
          prompt: 'Enter new due date (YYYY-MM-DD)',
          value: currentValue,
          validateInput: (v) => {
            const dateRegex = /^\d{4}-\d{2}-\d{2}$/;
            return dateRegex.test(v) ? null : 'Please use format YYYY-MM-DD';
          },
        });
        if (newValue) updates.dueDate = newValue;
        break;
      }
      case 'description': {
        newValue = await vscode.window.showInputBox({
          prompt: 'Enter invoice description',
          value: currentValue,
        });
        if (newValue !== undefined) updates.description = newValue;
        break;
      }
    }

    if (Object.keys(updates).length === 0) return;

    const result = await this.cli.editInvoice(this.invoiceId, updates);
    if (result.success) {
      vscode.window.showInformationMessage('Invoice updated successfully!');
      this.update();
      if (this.refreshCallback) this.refreshCallback();
    } else {
      vscode.window.showErrorMessage(
        `Failed to update invoice: ${result.error}`
      );
    }
  }

  /**
   * Mark invoice as paid
   */
  private async markAsPaid() {
    const result = await this.cli.markInvoicePaid(this.invoiceId);
    if (result.success) {
      vscode.window.showInformationMessage('Invoice marked as paid!');
      this.update();
      if (this.refreshCallback) this.refreshCallback();
    } else {
      vscode.window.showErrorMessage(`Failed to mark as paid: ${result.error}`);
    }
  }

  /**
   * Mark invoice as sent
   */
  private async markAsSent() {
    const result = await this.cli.markInvoiceSent(this.invoiceId);
    if (result.success) {
      vscode.window.showInformationMessage('Invoice marked as sent!');
      this.update();
      if (this.refreshCallback) this.refreshCallback();
    } else {
      vscode.window.showErrorMessage(`Failed to mark as sent: ${result.error}`);
    }
  }

  /**
   * Delete invoice
   */
  private async deleteInvoice() {
    const confirm = await vscode.window.showWarningMessage(
      `Are you sure you want to delete Invoice #${this.invoiceId}? This action cannot be undone.`,
      { modal: true },
      'Delete'
    );

    if (confirm !== 'Delete') return;

    const result = await this.cli.deleteInvoice(this.invoiceId);
    if (result.success) {
      vscode.window.showInformationMessage('Invoice deleted successfully!');
      if (this.refreshCallback) this.refreshCallback();
      this.dispose();
    } else {
      vscode.window.showErrorMessage(
        `Failed to delete invoice: ${result.error}`
      );
    }
  }

  /**
   * Duplicate invoice (create new invoice with same details)
   */
  private async duplicateInvoice() {
    if (!this.invoiceData) {
      vscode.window.showErrorMessage('Invoice data not loaded');
      return;
    }

    const result = await this.cli.createInvoice({
      clientName: this.invoiceData.clientName,
      amount: this.invoiceData.amount,
      currency: this.invoiceData.currency,
    });

    if (result.success) {
      vscode.window.showInformationMessage(
        'Invoice duplicated! A new draft invoice has been created.'
      );
      if (this.refreshCallback) this.refreshCallback();
    } else {
      vscode.window.showErrorMessage(
        `Failed to duplicate invoice: ${result.error}`
      );
    }
  }

  /**
   * Fetch invoice data from CLI
   */
  private async fetchInvoiceData(): Promise<InvoiceData | null> {
    const result = await this.cli.listInvoices();
    if (!result.success || !result.stdout) return null;

    const lines = result.stdout.trim().split('\n');
    if (lines.length < 2) return null;

    // Skip header and find our invoice
    for (let i = 1; i < lines.length; i++) {
      const parts = lines[i]
        .split(/\s{2,}/)
        .map((p) => p.trim())
        .filter((p) => p);
      if (parts.length >= 5) {
        const id = parseInt(parts[0], 10);
        if (id === this.invoiceId) {
          // Parse amount and currency
          const amountStr = parts[2];
          const amountMatch = amountStr.match(
            /([\d,.]+)\s*(USD|EUR|GBP|CHF|PLN)?/i
          );
          const amount = amountMatch
            ? parseFloat(amountMatch[1].replace(',', ''))
            : 0;
          const currency = amountMatch?.[2] || 'USD';

          return {
            id,
            invoiceNum: parts[1],
            amount,
            currency,
            status: parts[3],
            dueDate: parts[5] || 'Not set',
            clientName: parts[4],
            clientEmail: '', // Will be fetched separately if needed
          };
        }
      }
    }

    return null;
  }

  /**
   * Get HTML content for webview
   */
  private async getHtmlForWebview(): Promise<string> {
    // Fetch invoice data
    this.invoiceData = await this.fetchInvoiceData();

    const data = this.invoiceData;
    const invoiceNum = data?.invoiceNum || `#${this.invoiceId}`;
    const amount = data
      ? `${data.amount.toLocaleString('en-US', { minimumFractionDigits: 2 })} ${data.currency}`
      : 'N/A';
    const status = data?.status || 'Unknown';
    const statusLower = status.toLowerCase();
    const dueDate = data?.dueDate || 'Not set';
    const clientName = data?.clientName || 'Unknown';

    return `<!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <title>Invoice ${invoiceNum}</title>
            <style>
                /* ==============================================
                   UNG Design System - Invoice Panel
                   Aligned with macOS DesignTokens.swift
                   ============================================== */
                :root {
                    /* Brand Colors */
                    --ung-brand: #3373E8;
                    --ung-brand-light: rgba(51, 115, 232, 0.15);
                    --ung-brand-dark: #2660CC;

                    /* Semantic Colors */
                    --ung-success: #33A756;
                    --ung-success-light: rgba(51, 167, 86, 0.12);
                    --ung-warning: #F29932;
                    --ung-warning-light: rgba(242, 153, 50, 0.12);
                    --ung-error: #E65A5A;
                    --ung-error-light: rgba(230, 90, 90, 0.12);
                    --ung-info: #3373E8;
                    --ung-info-light: rgba(51, 115, 232, 0.12);

                    /* VSCode Integration */
                    --bg-primary: var(--vscode-editor-background);
                    --bg-secondary: var(--vscode-sideBar-background);
                    --bg-tertiary: var(--vscode-input-background);
                    --text-primary: var(--vscode-editor-foreground);
                    --text-secondary: var(--vscode-descriptionForeground);
                    --border: var(--vscode-panel-border);

                    /* Spacing */
                    --space-xs: 8px;
                    --space-sm: 12px;
                    --space-md: 16px;
                    --space-lg: 24px;
                    --space-xl: 32px;

                    /* Border Radius */
                    --radius-sm: 8px;
                    --radius-md: 12px;
                    --radius-full: 9999px;

                    /* Transitions */
                    --transition-micro: 0.1s cubic-bezier(0.4, 0, 0.2, 1);
                    --transition-quick: 0.15s cubic-bezier(0.4, 0, 0.2, 1);
                    --transition-bounce: 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);
                }

                * { box-sizing: border-box; margin: 0; padding: 0; }

                body {
                    font-family: var(--vscode-font-family);
                    background: var(--bg-primary);
                    color: var(--text-primary);
                    padding: var(--space-lg);
                    line-height: 1.5;
                    font-size: 13px;
                }

                .container {
                    max-width: 800px;
                    margin: 0 auto;
                }

                .header {
                    display: flex;
                    justify-content: space-between;
                    align-items: flex-start;
                    margin-bottom: var(--space-xl);
                    padding-bottom: var(--space-lg);
                    border-bottom: 1px solid var(--border);
                    flex-wrap: wrap;
                    gap: var(--space-md);
                }

                .header-info {
                    display: flex;
                    align-items: center;
                    gap: var(--space-md);
                }

                .invoice-icon {
                    width: 56px;
                    height: 56px;
                    background: linear-gradient(135deg, var(--ung-brand), var(--ung-brand-dark));
                    border-radius: var(--radius-md);
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    font-size: 24px;
                    color: white;
                    transition: transform var(--transition-bounce);
                }

                .invoice-icon:hover {
                    transform: scale(1.05) rotate(-3deg);
                }

                .header-text h1 {
                    font-size: 26px;
                    font-weight: 700;
                    margin-bottom: 4px;
                }

                .header-text .subtitle {
                    color: var(--text-secondary);
                    font-size: 13px;
                }

                .actions {
                    display: flex;
                    gap: var(--space-xs);
                    flex-wrap: wrap;
                }

                button {
                    background: var(--ung-brand);
                    color: white;
                    border: none;
                    padding: var(--space-sm) var(--space-md);
                    border-radius: var(--radius-sm);
                    cursor: pointer;
                    font-size: 13px;
                    font-weight: 500;
                    transition: all var(--transition-quick);
                    display: flex;
                    align-items: center;
                    gap: 6px;
                }

                button:hover {
                    background: var(--ung-brand-dark);
                    transform: translateY(-1px);
                    box-shadow: 0 4px 12px rgba(51, 115, 232, 0.25);
                }

                button:active {
                    transform: translateY(0) scale(0.98);
                }

                button.secondary {
                    background: var(--bg-secondary);
                    border: 1px solid var(--border);
                    color: var(--text-primary);
                }

                button.secondary:hover {
                    background: var(--bg-tertiary);
                    border-color: var(--ung-brand);
                    box-shadow: none;
                }

                button.success {
                    background: var(--ung-success);
                }

                button.success:hover {
                    background: #2A8F4A;
                    box-shadow: 0 4px 12px rgba(51, 167, 86, 0.25);
                }

                button.danger {
                    background: var(--ung-error);
                }

                button.danger:hover {
                    background: #D14545;
                    box-shadow: 0 4px 12px rgba(230, 90, 90, 0.25);
                }

                .card {
                    background: var(--bg-secondary);
                    border: 1px solid var(--border);
                    border-radius: var(--radius-md);
                    padding: var(--space-lg);
                    margin-bottom: var(--space-lg);
                    transition: border-color var(--transition-quick);
                }

                .card:hover {
                    border-color: rgba(51, 115, 232, 0.3);
                }

                .card-header {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    margin-bottom: var(--space-lg);
                    padding-bottom: var(--space-sm);
                    border-bottom: 1px solid var(--border);
                }

                .card-header h2 {
                    font-size: 15px;
                    font-weight: 600;
                    display: flex;
                    align-items: center;
                    gap: var(--space-xs);
                }

                .detail-grid {
                    display: grid;
                    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
                    gap: var(--space-md);
                }

                .detail-item {
                    padding: var(--space-md);
                    background: var(--bg-tertiary);
                    border-radius: var(--radius-sm);
                    transition: all var(--transition-quick);
                    position: relative;
                    border: 1px solid transparent;
                }

                .detail-item:hover {
                    transform: translateY(-2px);
                    box-shadow: 0 4px 12px rgba(0,0,0,0.08);
                }

                .detail-item.editable {
                    cursor: pointer;
                }

                .detail-item.editable:hover {
                    background: var(--bg-primary);
                    border-color: var(--ung-brand);
                }

                .detail-item.editable::after {
                    content: '✏️';
                    position: absolute;
                    top: var(--space-xs);
                    right: var(--space-xs);
                    font-size: 12px;
                    opacity: 0;
                    transition: opacity var(--transition-quick);
                }

                .detail-item.editable:hover::after {
                    opacity: 1;
                }

                .detail-label {
                    font-size: 11px;
                    color: var(--text-secondary);
                    text-transform: uppercase;
                    letter-spacing: 0.5px;
                    margin-bottom: var(--space-xs);
                    font-weight: 500;
                }

                .detail-value {
                    font-size: 17px;
                    font-weight: 600;
                    transition: color var(--transition-quick);
                }

                .detail-item:hover .detail-value {
                    color: var(--ung-brand);
                }

                .badge {
                    display: inline-block;
                    padding: 6px 14px;
                    border-radius: var(--radius-full);
                    font-size: 11px;
                    font-weight: 600;
                    text-transform: uppercase;
                    letter-spacing: 0.3px;
                }

                .badge.draft { background: rgba(128, 128, 128, 0.15); color: #808080; }
                .badge.pending { background: var(--ung-warning-light); color: var(--ung-warning); }
                .badge.sent { background: var(--ung-info-light); color: var(--ung-info); }
                .badge.paid { background: var(--ung-success-light); color: var(--ung-success); }
                .badge.overdue { background: var(--ung-error-light); color: var(--ung-error); }

                .amount-highlight {
                    font-size: 28px;
                    font-weight: 700;
                    color: var(--ung-success);
                    font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, monospace;
                }

                .quick-actions {
                    display: flex;
                    gap: var(--space-sm);
                    flex-wrap: wrap;
                    margin-top: var(--space-lg);
                    padding-top: var(--space-lg);
                    border-top: 1px solid var(--border);
                    align-items: center;
                }

                .action-buttons {
                    display: flex;
                    gap: var(--space-xs);
                    flex-wrap: wrap;
                }

                .danger-zone {
                    margin-top: var(--space-xl);
                    padding: var(--space-lg);
                    background: var(--ung-error-light);
                    border: 1px solid rgba(230, 90, 90, 0.3);
                    border-radius: var(--radius-md);
                    transition: border-color var(--transition-quick);
                }

                .danger-zone:hover {
                    border-color: var(--ung-error);
                }

                .danger-zone h3 {
                    color: var(--ung-error);
                    margin-bottom: var(--space-sm);
                    font-size: 14px;
                }

                /* Focus States */
                button:focus-visible {
                    outline: 2px solid var(--ung-brand);
                    outline-offset: 2px;
                }

                /* Reduced Motion */
                @media (prefers-reduced-motion: reduce) {
                    *, *::before, *::after {
                        animation-duration: 0.01ms !important;
                        transition-duration: 0.01ms !important;
                    }
                }

                @media (max-width: 600px) {
                    body { padding: var(--space-md); }
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
                            <h1>Invoice ${invoiceNum}</h1>
                            <span class="subtitle">View and manage invoice details</span>
                        </div>
                    </div>
                    <div class="actions">
                        <button onclick="exportPDF()">
                            📥 Export PDF
                        </button>
                        <button class="secondary" onclick="emailInvoice()">
                            📧 Send Email
                        </button>
                        <button class="secondary" onclick="duplicateInvoice()">
                            📋 Duplicate
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
                            <div class="detail-value">${invoiceNum}</div>
                        </div>
                        <div class="detail-item editable" onclick="editField('amount', '${data?.amount || 0}')">
                            <div class="detail-label">Amount (click to edit)</div>
                            <div class="detail-value amount-highlight">${amount}</div>
                        </div>
                        <div class="detail-item">
                            <div class="detail-label">Status</div>
                            <div class="detail-value">
                                <span class="badge ${statusLower}">${status}</span>
                            </div>
                        </div>
                        <div class="detail-item editable" onclick="editField('dueDate', '${dueDate}')">
                            <div class="detail-label">Due Date (click to edit)</div>
                            <div class="detail-value">${dueDate}</div>
                        </div>
                    </div>

                    <div class="quick-actions">
                        <span style="color: var(--text-secondary); margin-right: 10px;">Quick Actions:</span>
                        <div class="action-buttons">
                            ${
                              statusLower !== 'paid'
                                ? `<button class="success" onclick="markPaid()">✅ Mark as Paid</button>`
                                : ''
                            }
                            ${
                              statusLower === 'pending' ||
                              statusLower === 'draft'
                                ? `<button class="secondary" onclick="markSent()">📤 Mark as Sent</button>`
                                : ''
                            }
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
                            <div class="detail-value">${clientName}</div>
                        </div>
                    </div>
                </div>

                <div class="danger-zone">
                    <h3>⚠️ Danger Zone</h3>
                    <p style="margin-bottom: 12px; color: var(--text-secondary); font-size: 13px;">
                        Deleting an invoice is permanent and cannot be undone.
                    </p>
                    <button class="danger" onclick="deleteInvoice()">
                        🗑️ Delete Invoice
                    </button>
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

                function editField(field, currentValue) {
                    vscode.postMessage({ command: 'edit', field, value: currentValue });
                }

                function markPaid() {
                    vscode.postMessage({ command: 'markPaid' });
                }

                function markSent() {
                    vscode.postMessage({ command: 'markSent' });
                }

                function deleteInvoice() {
                    vscode.postMessage({ command: 'delete' });
                }

                function duplicateInvoice() {
                    vscode.postMessage({ command: 'duplicate' });
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
