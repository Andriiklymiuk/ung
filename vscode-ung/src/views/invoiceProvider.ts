import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';
import { Formatter } from '../utils/formatting';

/**
 * Invoice tree item
 */
export class InvoiceItem extends vscode.TreeItem {
    constructor(
        public readonly itemId: number,
        public readonly invoiceNum: string,
        public readonly amount: number,
        public readonly currency: string,
        public readonly status: string,
        public readonly dueDate: string,
        public readonly collapsibleState: vscode.TreeItemCollapsibleState
    ) {
        super(invoiceNum, collapsibleState);

        this.id = String(itemId);
        this.tooltip = `${invoiceNum} - ${Formatter.formatCurrency(amount, currency)}\nStatus: ${status}\nDue: ${Formatter.formatDate(dueDate)}`;
        this.description = `${Formatter.formatCurrency(amount, currency)} • ${status}`;
        this.contextValue = 'invoice';

        // Set icon based on status
        this.iconPath = this.getStatusIcon(status);
    }

    private getStatusIcon(status: string): vscode.ThemeIcon {
        const statusLower = status.toLowerCase();
        if (statusLower === 'paid') {
            return new vscode.ThemeIcon('check', new vscode.ThemeColor('testing.iconPassed'));
        } else if (statusLower === 'pending') {
            return new vscode.ThemeIcon('clock', new vscode.ThemeColor('testing.iconQueued'));
        } else if (statusLower === 'overdue') {
            return new vscode.ThemeIcon('warning', new vscode.ThemeColor('testing.iconFailed'));
        } else {
            return new vscode.ThemeIcon('file');
        }
    }
}

/**
 * Invoice tree data provider
 */
export class InvoiceProvider implements vscode.TreeDataProvider<InvoiceItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<InvoiceItem | undefined | null | void> = new vscode.EventEmitter<InvoiceItem | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<InvoiceItem | undefined | null | void> = this._onDidChangeTreeData.event;

    constructor(private cli: UngCli) {}

    refresh(): void {
        this._onDidChangeTreeData.fire();
    }

    getTreeItem(element: InvoiceItem): vscode.TreeItem {
        return element;
    }

    async getChildren(element?: InvoiceItem): Promise<InvoiceItem[]> {
        if (element) {
            return [];
        }

        try {
            const result = await this.cli.listInvoices();

            if (!result.success || !result.stdout) {
                return [];
            }

            // Parse the CLI output
            const invoices = this.parseInvoiceOutput(result.stdout);
            return invoices;
        } catch (error) {
            vscode.window.showErrorMessage(`Failed to load invoices: ${error}`);
            return [];
        }
    }

    /**
     * Parse invoice list output from CLI
     */
    private parseInvoiceOutput(output: string): InvoiceItem[] {
        const lines = output.split('\n').filter(line => line.trim());
        const invoices: InvoiceItem[] = [];

        for (let i = 1; i < lines.length; i++) { // Skip header
            const line = lines[i].trim();
            if (!line) continue;

            const parts = line.split(/\s{2,}/); // Split by multiple spaces
            if (parts.length >= 6) {
                const id = parseInt(parts[0]);
                const invoiceNum = parts[1];
                const amountParts = parts[2].split(' ');
                const amount = parseFloat(amountParts[0]);
                const currency = amountParts[1] || 'USD';
                const status = parts[3];
                const dueDate = parts[5];

                if (!isNaN(id) && !isNaN(amount)) {
                    invoices.push(new InvoiceItem(
                        id,
                        invoiceNum,
                        amount,
                        currency,
                        status,
                        dueDate,
                        vscode.TreeItemCollapsibleState.None
                    ));
                }
            }
        }

        return invoices;
    }
}
