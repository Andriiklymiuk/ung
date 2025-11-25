import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';
import { BaseCommand } from '../utils/baseCommand';
import { NotificationManager } from '../utils/notifications';
import { Formatter } from '../utils/formatting';

/**
 * Search result item
 */
interface SearchResult {
    type: 'invoice' | 'client' | 'contract' | 'expense';
    id: number;
    label: string;
    description: string;
    detail?: string;
}

/**
 * Search command handlers
 */
export class SearchCommands extends BaseCommand {
    constructor(cli: UngCli) {
        super(cli);
    }

    /**
     * Global search across all entities
     */
    async globalSearch(): Promise<void> {
        const query = await this.getInput('Search', {
            placeholder: 'Search invoices, clients, contracts, expenses...',
            required: true
        });

        if (!query) {
            return;
        }

        await NotificationManager.withProgress(
            'Searching...',
            async () => {
                const results = await this.performSearch(query);

                if (results.length === 0) {
                    NotificationManager.info(`No results found for "${query}"`);
                    return;
                }

                await this.showSearchResults(results, query);
            }
        );
    }

    /**
     * Search invoices
     */
    async searchInvoices(): Promise<void> {
        const query = await this.getInput('Search Invoices', {
            placeholder: 'Search by invoice number, client, amount...',
            required: true
        });

        if (!query) {
            return;
        }

        await NotificationManager.withProgress(
            'Searching invoices...',
            async () => {
                const results = await this.searchInvoiceData(query);
                await this.showSearchResults(results, query);
            }
        );
    }

    /**
     * Search clients
     */
    async searchClients(): Promise<void> {
        const query = await this.getInput('Search Clients', {
            placeholder: 'Search by name, email...',
            required: true
        });

        if (!query) {
            return;
        }

        await NotificationManager.withProgress(
            'Searching clients...',
            async () => {
                const results = await this.searchClientData(query);
                await this.showSearchResults(results, query);
            }
        );
    }

    /**
     * Filter invoices by status
     */
    async filterInvoicesByStatus(): Promise<void> {
        const status = await this.showQuickPick([
            { label: 'Paid', description: 'Show paid invoices' },
            { label: 'Pending', description: 'Show pending invoices' },
            { label: 'Overdue', description: 'Show overdue invoices' },
            { label: 'All', description: 'Show all invoices' }
        ], {
            placeHolder: 'Select invoice status to filter'
        });

        if (!status) {
            return;
        }

        await NotificationManager.withProgress(
            'Filtering invoices...',
            async () => {
                const results = await this.getInvoicesByStatus(status.label);
                await this.showSearchResults(results, `Status: ${status.label}`);
            }
        );
    }

    /**
     * Filter invoices by date range
     */
    async filterInvoicesByDateRange(): Promise<void> {
        const rangeType = await this.showQuickPick([
            { label: 'This Month', description: 'Invoices from current month' },
            { label: 'Last Month', description: 'Invoices from last month' },
            { label: 'This Quarter', description: 'Invoices from current quarter' },
            { label: 'This Year', description: 'Invoices from current year' },
            { label: 'Custom Range', description: 'Specify custom date range' }
        ], {
            placeHolder: 'Select date range'
        });

        if (!rangeType) {
            return;
        }

        let startDate: string;
        let endDate: string;

        if (rangeType.label === 'Custom Range') {
            const start = await this.getDateInput('Start Date', { required: true });
            const end = await this.getDateInput('End Date', { required: true });

            if (!start || !end) {
                return;
            }

            startDate = start;
            endDate = end;
        } else {
            const dates = this.getDateRangeForPeriod(rangeType.label);
            startDate = dates.start;
            endDate = dates.end;
        }

        await NotificationManager.withProgress(
            'Filtering invoices...',
            async () => {
                const results = await this.getInvoicesByDateRange(startDate, endDate);
                await this.showSearchResults(
                    results,
                    `${rangeType.label}: ${startDate} to ${endDate}`
                );
            }
        );
    }

    /**
     * Perform global search
     */
    private async performSearch(query: string): Promise<SearchResult[]> {
        const results: SearchResult[] = [];
        const lowerQuery = query.toLowerCase();

        // Search invoices
        const invoiceResults = await this.searchInvoiceData(query);
        results.push(...invoiceResults);

        // Search clients
        const clientResults = await this.searchClientData(query);
        results.push(...clientResults);

        // Search contracts
        const contractResults = await this.searchContractData(query);
        results.push(...contractResults);

        // Search expenses
        const expenseResults = await this.searchExpenseData(query);
        results.push(...expenseResults);

        return results;
    }

    /**
     * Search invoice data
     */
    private async searchInvoiceData(query: string): Promise<SearchResult[]> {
        const result = await this.cli.listInvoices();
        if (!result.success || !result.stdout) {
            return [];
        }

        const results: SearchResult[] = [];
        const lowerQuery = query.toLowerCase();
        const lines = this.parseListOutput(result.stdout);

        for (const parts of lines) {
            if (parts.length >= 6) {
                const id = parseInt(parts[0]);
                const invoiceNum = parts[1];
                const amount = parts[2];
                const status = parts[3];
                const client = parts[4];
                const dueDate = parts[5];

                const searchText = `${invoiceNum} ${client} ${amount} ${status}`.toLowerCase();
                if (searchText.includes(lowerQuery)) {
                    results.push({
                        type: 'invoice',
                        id,
                        label: `${invoiceNum} - ${client}`,
                        description: `${amount} • ${status}`,
                        detail: `Due: ${dueDate}`
                    });
                }
            }
        }

        return results;
    }

    /**
     * Search client data
     */
    private async searchClientData(query: string): Promise<SearchResult[]> {
        const result = await this.cli.listClients();
        if (!result.success || !result.stdout) {
            return [];
        }

        const results: SearchResult[] = [];
        const lowerQuery = query.toLowerCase();
        const lines = this.parseListOutput(result.stdout);

        for (const parts of lines) {
            if (parts.length >= 3) {
                const id = parseInt(parts[0]);
                const name = parts[1];
                const email = parts[2];

                const searchText = `${name} ${email}`.toLowerCase();
                if (searchText.includes(lowerQuery)) {
                    results.push({
                        type: 'client',
                        id,
                        label: name,
                        description: email,
                        detail: 'Client'
                    });
                }
            }
        }

        return results;
    }

    /**
     * Search contract data
     */
    private async searchContractData(query: string): Promise<SearchResult[]> {
        const result = await this.cli.listContracts();
        if (!result.success || !result.stdout) {
            return [];
        }

        const results: SearchResult[] = [];
        const lowerQuery = query.toLowerCase();
        const lines = this.parseListOutput(result.stdout);

        for (const parts of lines) {
            if (parts.length >= 4) {
                const id = parseInt(parts[0]);
                const client = parts[1];
                const type = parts[2];
                const rate = parts[3];

                const searchText = `${client} ${type} ${rate}`.toLowerCase();
                if (searchText.includes(lowerQuery)) {
                    results.push({
                        type: 'contract',
                        id,
                        label: `${client} - ${type}`,
                        description: rate,
                        detail: 'Contract'
                    });
                }
            }
        }

        return results;
    }

    /**
     * Search expense data
     */
    private async searchExpenseData(query: string): Promise<SearchResult[]> {
        const result = await this.cli.listExpenses();
        if (!result.success || !result.stdout) {
            return [];
        }

        const results: SearchResult[] = [];
        const lowerQuery = query.toLowerCase();
        const lines = this.parseListOutput(result.stdout);

        for (const parts of lines) {
            if (parts.length >= 4) {
                const id = parseInt(parts[0]);
                const description = parts[1];
                const amount = parts[2];
                const date = parts[3];

                const searchText = `${description} ${amount}`.toLowerCase();
                if (searchText.includes(lowerQuery)) {
                    results.push({
                        type: 'expense',
                        id,
                        label: description,
                        description: amount,
                        detail: `Date: ${date}`
                    });
                }
            }
        }

        return results;
    }

    /**
     * Get invoices by status
     */
    private async getInvoicesByStatus(status: string): Promise<SearchResult[]> {
        if (status === 'All') {
            return this.searchInvoiceData('');
        }
        return this.searchInvoiceData(status);
    }

    /**
     * Get invoices by date range
     */
    private async getInvoicesByDateRange(startDate: string, endDate: string): Promise<SearchResult[]> {
        const allInvoices = await this.searchInvoiceData('');
        const start = new Date(startDate);
        const end = new Date(endDate);

        return allInvoices.filter(invoice => {
            if (!invoice.detail) return false;
            const dueDateMatch = invoice.detail.match(/Due: (\d{4}-\d{2}-\d{2})/);
            if (!dueDateMatch) return false;

            const dueDate = new Date(dueDateMatch[1]);
            return dueDate >= start && dueDate <= end;
        });
    }

    /**
     * Get date range for period
     */
    private getDateRangeForPeriod(period: string): { start: string; end: string } {
        const now = new Date();
        let start: Date;
        let end: Date = now;

        switch (period) {
            case 'This Month':
                start = new Date(now.getFullYear(), now.getMonth(), 1);
                break;
            case 'Last Month':
                start = new Date(now.getFullYear(), now.getMonth() - 1, 1);
                end = new Date(now.getFullYear(), now.getMonth(), 0);
                break;
            case 'This Quarter':
                const quarter = Math.floor(now.getMonth() / 3);
                start = new Date(now.getFullYear(), quarter * 3, 1);
                break;
            case 'This Year':
                start = new Date(now.getFullYear(), 0, 1);
                break;
            default:
                start = new Date(now.getFullYear(), now.getMonth(), 1);
        }

        return {
            start: Formatter.formatDate(start),
            end: Formatter.formatDate(end)
        };
    }

    /**
     * Show search results
     */
    private async showSearchResults(results: SearchResult[], query: string): Promise<void> {
        if (results.length === 0) {
            NotificationManager.info(`No results found for "${query}"`);
            return;
        }

        const quickPickItems = results.map(result => ({
            label: result.label,
            description: result.description,
            detail: result.detail,
            result
        }));

        const selected = await vscode.window.showQuickPick(quickPickItems, {
            placeHolder: `${results.length} result(s) for "${query}"`,
            matchOnDescription: true,
            matchOnDetail: true
        });

        if (selected) {
            await this.openSearchResult(selected.result);
        }
    }

    /**
     * Open selected search result
     */
    private async openSearchResult(result: SearchResult): Promise<void> {
        switch (result.type) {
            case 'invoice':
                await vscode.commands.executeCommand('ung.viewInvoice', { id: result.id });
                break;
            case 'client':
                await vscode.commands.executeCommand('ung.editClient', { id: result.id });
                break;
            case 'contract':
                await vscode.commands.executeCommand('ung.viewContract', { id: result.id });
                break;
            case 'expense':
                await vscode.commands.executeCommand('ung.editExpense', { id: result.id });
                break;
        }
    }
}
