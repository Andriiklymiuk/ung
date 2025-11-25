import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';
import { Formatter } from '../utils/formatting';

type InvoiceTreeItem = InvoiceSectionItem | InvoiceSummaryItem | InvoiceItem | InvoiceActionItem;

/**
 * Invoice section for grouping
 */
class InvoiceSectionItem extends vscode.TreeItem {
    constructor(
        public readonly label: string,
        public readonly section: string,
        public readonly icon: string,
        public readonly colorId: string,
        public readonly count: number
    ) {
        super(label, count > 0 ? vscode.TreeItemCollapsibleState.Expanded : vscode.TreeItemCollapsibleState.Collapsed);
        this.description = `${count}`;
        this.iconPath = new vscode.ThemeIcon(icon, new vscode.ThemeColor(colorId));
        this.contextValue = `invoice_section_${section}`;
    }
}

/**
 * Invoice summary item
 */
class InvoiceSummaryItem extends vscode.TreeItem {
    constructor(
        label: string,
        value: string,
        icon: string,
        colorId?: string
    ) {
        super(label, vscode.TreeItemCollapsibleState.None);
        this.description = value;
        this.iconPath = colorId
            ? new vscode.ThemeIcon(icon, new vscode.ThemeColor(colorId))
            : new vscode.ThemeIcon(icon);
        this.contextValue = 'invoice_summary';
    }
}

/**
 * Invoice action item for quick actions
 */
class InvoiceActionItem extends vscode.TreeItem {
    constructor(
        label: string,
        icon: string,
        commandId: string,
        colorId?: string
    ) {
        super(label, vscode.TreeItemCollapsibleState.None);
        this.iconPath = colorId
            ? new vscode.ThemeIcon(icon, new vscode.ThemeColor(colorId))
            : new vscode.ThemeIcon(icon);
        this.command = {
            command: commandId,
            title: label
        };
        this.contextValue = 'invoice_action';
    }
}

/**
 * Invoice tree item
 */
export class InvoiceItem extends vscode.TreeItem {
    constructor(
        public readonly itemId: number,
        public readonly invoiceNum: string,
        public readonly clientName: string,
        public readonly amount: number,
        public readonly currency: string,
        public readonly status: string,
        public readonly dueDate: string,
        public readonly collapsibleState: vscode.TreeItemCollapsibleState
    ) {
        super(invoiceNum, collapsibleState);

        this.id = `invoice_${itemId}`;
        this.tooltip = new vscode.MarkdownString();
        this.tooltip.appendMarkdown(`**${invoiceNum}**\n\n`);
        this.tooltip.appendMarkdown(`- Client: ${clientName}\n`);
        this.tooltip.appendMarkdown(`- Amount: ${Formatter.formatCurrency(amount, currency)}\n`);
        this.tooltip.appendMarkdown(`- Status: ${status}\n`);
        this.tooltip.appendMarkdown(`- Due: ${Formatter.formatDate(dueDate)}\n`);
        this.tooltip.appendMarkdown(`\n*Click to view details*`);

        this.description = `${clientName} • ${Formatter.formatCurrency(amount, currency)}`;
        this.contextValue = 'invoice';

        // Set icon based on status with colors
        this.iconPath = this.getStatusIcon(status);

        // Add view command on click
        this.command = {
            command: 'ung.viewInvoice',
            title: 'View Invoice',
            arguments: [{ itemId }]
        };
    }

    private getStatusIcon(status: string): vscode.ThemeIcon {
        const statusLower = status.toLowerCase();
        switch (statusLower) {
            case 'paid':
                return new vscode.ThemeIcon('pass-filled', new vscode.ThemeColor('testing.iconPassed'));
            case 'pending':
                return new vscode.ThemeIcon('circle-outline', new vscode.ThemeColor('charts.yellow'));
            case 'sent':
                return new vscode.ThemeIcon('mail', new vscode.ThemeColor('charts.blue'));
            case 'overdue':
                return new vscode.ThemeIcon('alert', new vscode.ThemeColor('testing.iconFailed'));
            case 'draft':
                return new vscode.ThemeIcon('edit', new vscode.ThemeColor('charts.gray'));
            default:
                return new vscode.ThemeIcon('file');
        }
    }
}

interface ParsedInvoice {
    id: number;
    invoiceNum: string;
    clientName: string;
    amount: number;
    currency: string;
    status: string;
    dueDate: string;
}

/**
 * Invoice tree data provider with grouping by status
 */
export class InvoiceProvider implements vscode.TreeDataProvider<InvoiceTreeItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<InvoiceTreeItem | undefined | null | void> =
        new vscode.EventEmitter<InvoiceTreeItem | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<InvoiceTreeItem | undefined | null | void> =
        this._onDidChangeTreeData.event;

    private cachedInvoices: ParsedInvoice[] = [];

    constructor(private cli: UngCli) {}

    refresh(): void {
        this.cachedInvoices = [];
        this._onDidChangeTreeData.fire();
    }

    getTreeItem(element: InvoiceTreeItem): vscode.TreeItem {
        return element;
    }

    async getChildren(element?: InvoiceTreeItem): Promise<InvoiceTreeItem[]> {
        if (!element) {
            return this.getRootItems();
        }

        if (element instanceof InvoiceSectionItem) {
            return this.getInvoicesForSection(element.section);
        }

        return [];
    }

    private async getRootItems(): Promise<InvoiceTreeItem[]> {
        await this.loadInvoices();
        const items: InvoiceTreeItem[] = [];

        // Summary section
        const totalAmount = this.cachedInvoices.reduce((sum, inv) => sum + inv.amount, 0);
        const unpaidAmount = this.cachedInvoices
            .filter(inv => ['pending', 'sent', 'overdue'].includes(inv.status.toLowerCase()))
            .reduce((sum, inv) => sum + inv.amount, 0);
        const overdueCount = this.cachedInvoices.filter(inv => inv.status.toLowerCase() === 'overdue').length;

        if (this.cachedInvoices.length > 0) {
            items.push(new InvoiceSummaryItem(
                'Total Invoiced',
                Formatter.formatCurrency(totalAmount, 'USD'),
                'graph-line',
                'charts.green'
            ));

            if (unpaidAmount > 0) {
                items.push(new InvoiceSummaryItem(
                    'Awaiting Payment',
                    Formatter.formatCurrency(unpaidAmount, 'USD'),
                    'clock',
                    overdueCount > 0 ? 'charts.red' : 'charts.yellow'
                ));
            }
        }

        // Quick actions
        items.push(new InvoiceActionItem(
            'Create New Invoice',
            'add',
            'ung.createInvoice',
            'charts.blue'
        ));

        items.push(new InvoiceActionItem(
            'Generate from Time',
            'clock',
            'ung.generateFromTime',
            'charts.purple'
        ));

        // Status sections
        const statusGroups = [
            { key: 'overdue', label: 'Overdue', icon: 'alert', color: 'charts.red' },
            { key: 'pending', label: 'Pending', icon: 'circle-outline', color: 'charts.yellow' },
            { key: 'sent', label: 'Sent', icon: 'mail', color: 'charts.blue' },
            { key: 'paid', label: 'Paid', icon: 'pass-filled', color: 'charts.green' },
            { key: 'draft', label: 'Drafts', icon: 'edit', color: 'charts.gray' },
        ];

        for (const group of statusGroups) {
            const count = this.cachedInvoices.filter(
                inv => inv.status.toLowerCase() === group.key
            ).length;
            if (count > 0 || group.key === 'pending') {
                items.push(new InvoiceSectionItem(
                    group.label,
                    group.key,
                    group.icon,
                    group.color,
                    count
                ));
            }
        }

        return items;
    }

    private async getInvoicesForSection(section: string): Promise<InvoiceItem[]> {
        return this.cachedInvoices
            .filter(inv => inv.status.toLowerCase() === section)
            .sort((a, b) => new Date(b.dueDate).getTime() - new Date(a.dueDate).getTime())
            .map(inv => new InvoiceItem(
                inv.id,
                inv.invoiceNum,
                inv.clientName,
                inv.amount,
                inv.currency,
                inv.status,
                inv.dueDate,
                vscode.TreeItemCollapsibleState.None
            ));
    }

    private async loadInvoices(): Promise<void> {
        if (this.cachedInvoices.length > 0) {
            return;
        }

        try {
            const result = await this.cli.listInvoices();

            if (!result.success || !result.stdout) {
                this.cachedInvoices = [];
                return;
            }

            this.cachedInvoices = this.parseInvoiceOutput(result.stdout);
        } catch (error) {
            vscode.window.showErrorMessage(`Failed to load invoices: ${error}`);
            this.cachedInvoices = [];
        }
    }

    /**
     * Parse invoice list output from CLI
     */
    private parseInvoiceOutput(output: string): ParsedInvoice[] {
        const lines = output.split('\n').filter(line => line.trim());
        const invoices: ParsedInvoice[] = [];

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
                const clientName = parts[4] || 'Unknown';
                const dueDate = parts[5];

                if (!isNaN(id) && !isNaN(amount)) {
                    invoices.push({
                        id,
                        invoiceNum,
                        clientName,
                        amount,
                        currency,
                        status,
                        dueDate
                    });
                }
            }
        }

        return invoices;
    }
}
