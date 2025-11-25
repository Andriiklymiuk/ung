import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';
import { BaseCommand } from '../utils/baseCommand';
import { NotificationManager } from '../utils/notifications';
import { ErrorHandler } from '../utils/errors';

/**
 * Bulk operation item
 */
interface BulkItem {
    id: number;
    label: string;
    description?: string;
}

/**
 * Bulk operations command handlers
 */
export class BulkCommands extends BaseCommand {
    constructor(cli: UngCli, private refreshCallback?: () => void) {
        super(cli, refreshCallback);
    }

    /**
     * Bulk export invoices to PDF
     */
    async bulkExportInvoices(): Promise<void> {
        const invoices = await this.selectInvoices('Select invoices to export');
        if (!invoices || invoices.length === 0) {
            return;
        }

        const confirmed = await this.confirm(
            `Export ${invoices.length} invoice(s) to PDF?`,
            {
                detail: 'This will generate PDF files for all selected invoices.',
                confirmLabel: 'Export',
                cancelLabel: 'Cancel'
            }
        );

        if (!confirmed) {
            return;
        }

        await this.executeBulkOperation(
            invoices,
            async (invoice) => {
                return await this.cli.generateInvoicePDF(invoice.id);
            },
            'Exporting invoices',
            'Export'
        );
    }

    /**
     * Bulk email invoices
     */
    async bulkEmailInvoices(): Promise<void> {
        const invoices = await this.selectInvoices('Select invoices to email');
        if (!invoices || invoices.length === 0) {
            return;
        }

        const confirmed = await this.confirm(
            `Email ${invoices.length} invoice(s)?`,
            {
                detail: 'This will send emails for all selected invoices.',
                confirmLabel: 'Send',
                cancelLabel: 'Cancel'
            }
        );

        if (!confirmed) {
            return;
        }

        await this.executeBulkOperation(
            invoices,
            async (invoice) => {
                return await this.cli.emailInvoice(invoice.id);
            },
            'Emailing invoices',
            'Email'
        );
    }

    /**
     * Bulk mark invoices as paid
     */
    async bulkMarkInvoicesPaid(): Promise<void> {
        const invoices = await this.selectInvoices('Select invoices to mark as paid', 'Pending');
        if (!invoices || invoices.length === 0) {
            return;
        }

        const confirmed = await this.confirm(
            `Mark ${invoices.length} invoice(s) as paid?`,
            {
                detail: 'This action cannot be undone.',
                confirmLabel: 'Mark as Paid',
                cancelLabel: 'Cancel'
            }
        );

        if (!confirmed) {
            return;
        }

        await this.executeBulkOperation(
            invoices,
            async (invoice) => {
                // Note: This assumes there's a CLI command for marking as paid
                // Adjust based on actual CLI API
                return await this.cli.exec(['invoice', 'pay', invoice.id.toString()]);
            },
            'Marking invoices as paid',
            'Mark as Paid'
        );
    }

    /**
     * Bulk delete invoices
     */
    async bulkDeleteInvoices(): Promise<void> {
        const invoices = await this.selectInvoices('Select invoices to delete');
        if (!invoices || invoices.length === 0) {
            return;
        }

        const confirmed = await this.confirm(
            `Delete ${invoices.length} invoice(s)?`,
            {
                detail: 'This action cannot be undone. Are you absolutely sure?',
                confirmLabel: 'Delete',
                cancelLabel: 'Cancel'
            }
        );

        if (!confirmed) {
            return;
        }

        // Double confirmation for delete
        const doubleConfirm = await this.confirm(
            'Final confirmation required',
            {
                detail: `You are about to permanently delete ${invoices.length} invoice(s).`,
                confirmLabel: 'Yes, Delete',
                cancelLabel: 'Cancel'
            }
        );

        if (!doubleConfirm) {
            return;
        }

        await this.executeBulkOperation(
            invoices,
            async (invoice) => {
                return await this.cli.exec(['invoice', 'delete', invoice.id.toString()]);
            },
            'Deleting invoices',
            'Delete'
        );
    }

    /**
     * Bulk delete clients
     */
    async bulkDeleteClients(): Promise<void> {
        const clients = await this.selectClients('Select clients to delete');
        if (!clients || clients.length === 0) {
            return;
        }

        const confirmed = await this.confirm(
            `Delete ${clients.length} client(s)?`,
            {
                detail: 'This will also affect related invoices and contracts.',
                confirmLabel: 'Delete',
                cancelLabel: 'Cancel'
            }
        );

        if (!confirmed) {
            return;
        }

        await this.executeBulkOperation(
            clients,
            async (client) => {
                return await this.cli.deleteClient(client.id);
            },
            'Deleting clients',
            'Delete'
        );
    }

    /**
     * Bulk delete expenses
     */
    async bulkDeleteExpenses(): Promise<void> {
        const expenses = await this.selectExpenses('Select expenses to delete');
        if (!expenses || expenses.length === 0) {
            return;
        }

        const confirmed = await this.confirm(
            `Delete ${expenses.length} expense(s)?`,
            {
                detail: 'This action cannot be undone.',
                confirmLabel: 'Delete',
                cancelLabel: 'Cancel'
            }
        );

        if (!confirmed) {
            return;
        }

        await this.executeBulkOperation(
            expenses,
            async (expense) => {
                return await this.cli.exec(['expense', 'delete', expense.id.toString()]);
            },
            'Deleting expenses',
            'Delete'
        );
    }

    /**
     * Execute bulk operation with progress tracking
     */
    private async executeBulkOperation<T extends BulkItem>(
        items: T[],
        operation: (item: T) => Promise<any>,
        progressTitle: string,
        operationName: string
    ): Promise<void> {
        let completed = 0;
        let failed = 0;
        const errors: string[] = [];

        await vscode.window.withProgress(
            {
                location: vscode.ProgressLocation.Notification,
                title: progressTitle,
                cancellable: true
            },
            async (progress, token) => {
                for (let i = 0; i < items.length; i++) {
                    if (token.isCancellationRequested) {
                        NotificationManager.warning(`${operationName} cancelled. ${completed} of ${items.length} completed.`);
                        return;
                    }

                    const item = items[i];
                    progress.report({
                        message: `${i + 1}/${items.length}: ${item.label}`,
                        increment: (100 / items.length)
                    });

                    try {
                        const result = await operation(item);
                        if (result.success) {
                            completed++;
                        } else {
                            failed++;
                            errors.push(`${item.label}: ${result.error || 'Unknown error'}`);
                        }
                    } catch (error) {
                        failed++;
                        errors.push(`${item.label}: ${error}`);
                        ErrorHandler.logWarning(`${operationName} failed for ${item.label}: ${error}`);
                    }
                }
            }
        );

        // Show results
        if (failed === 0) {
            NotificationManager.success(
                `${operationName} completed successfully for all ${completed} item(s).`
            );
        } else {
            const action = await NotificationManager.error(
                `${operationName} completed with ${failed} failure(s). ${completed} item(s) succeeded.`,
                { actions: ['View Errors', 'OK'] }
            );

            if (action === 'View Errors') {
                const errorDoc = await vscode.workspace.openTextDocument({
                    content: errors.join('\n'),
                    language: 'text'
                });
                await vscode.window.showTextDocument(errorDoc);
            }
        }

        // Refresh views
        if (this.refreshCallback) {
            this.refreshCallback();
        }
    }

    /**
     * Select invoices for bulk operation
     */
    private async selectInvoices(title: string, statusFilter?: string): Promise<BulkItem[] | undefined> {
        const result = await this.cli.listInvoices();
        if (!result.success || !result.stdout) {
            NotificationManager.error('Failed to load invoices');
            return undefined;
        }

        const items: (vscode.QuickPickItem & { id: number })[] = [];
        const lines = this.parseListOutput(result.stdout);

        for (const parts of lines) {
            if (parts.length >= 6) {
                const id = parseInt(parts[0]);
                const invoiceNum = parts[1];
                const amount = parts[2];
                const status = parts[3];
                const client = parts[4];

                // Apply status filter if provided
                if (statusFilter && !status.toLowerCase().includes(statusFilter.toLowerCase())) {
                    continue;
                }

                items.push({
                    id,
                    label: `${invoiceNum} - ${client}`,
                    description: `${amount} • ${status}`,
                    picked: false
                });
            }
        }

        if (items.length === 0) {
            NotificationManager.info('No invoices available');
            return undefined;
        }

        const selected = await NotificationManager.selectMultiple(items, {
            placeHolder: title,
            canPickMany: true
        });

        if (!selected || selected.length === 0) {
            return undefined;
        }

        return selected.map(item => ({
            id: item.id,
            label: item.label,
            description: item.description
        }));
    }

    /**
     * Select clients for bulk operation
     */
    private async selectClients(title: string): Promise<BulkItem[] | undefined> {
        const result = await this.cli.listClients();
        if (!result.success || !result.stdout) {
            NotificationManager.error('Failed to load clients');
            return undefined;
        }

        const items: (vscode.QuickPickItem & { id: number })[] = [];
        const lines = this.parseListOutput(result.stdout);

        for (const parts of lines) {
            if (parts.length >= 3) {
                const id = parseInt(parts[0]);
                const name = parts[1];
                const email = parts[2];

                items.push({
                    id,
                    label: name,
                    description: email,
                    picked: false
                });
            }
        }

        if (items.length === 0) {
            NotificationManager.info('No clients available');
            return undefined;
        }

        const selected = await NotificationManager.selectMultiple(items, {
            placeHolder: title,
            canPickMany: true
        });

        if (!selected || selected.length === 0) {
            return undefined;
        }

        return selected.map(item => ({
            id: item.id,
            label: item.label,
            description: item.description
        }));
    }

    /**
     * Select expenses for bulk operation
     */
    private async selectExpenses(title: string): Promise<BulkItem[] | undefined> {
        const result = await this.cli.listExpenses();
        if (!result.success || !result.stdout) {
            NotificationManager.error('Failed to load expenses');
            return undefined;
        }

        const items: (vscode.QuickPickItem & { id: number })[] = [];
        const lines = this.parseListOutput(result.stdout);

        for (const parts of lines) {
            if (parts.length >= 4) {
                const id = parseInt(parts[0]);
                const description = parts[1];
                const amount = parts[2];
                const date = parts[3];

                items.push({
                    id,
                    label: description,
                    description: `${amount} • ${date}`,
                    picked: false
                });
            }
        }

        if (items.length === 0) {
            NotificationManager.info('No expenses available');
            return undefined;
        }

        const selected = await NotificationManager.selectMultiple(items, {
            placeHolder: title,
            canPickMany: true
        });

        if (!selected || selected.length === 0) {
            return undefined;
        }

        return selected.map(item => ({
            id: item.id,
            label: item.label,
            description: item.description
        }));
    }
}
