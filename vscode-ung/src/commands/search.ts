import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';

interface SearchResult {
    type: 'invoice' | 'contract' | 'client' | 'expense' | 'time';
    id: number;
    label: string;
    description: string;
    detail: string;
}

/**
 * Universal search across all UNG data
 */
export class SearchCommands {
    constructor(private cli: UngCli) {}

    /**
     * Show universal search dialog
     */
    async universalSearch(): Promise<void> {
        const quickPick = vscode.window.createQuickPick<SearchResult & vscode.QuickPickItem>();
        quickPick.placeholder = 'Search invoices, contracts, clients, expenses...';
        quickPick.matchOnDescription = true;
        quickPick.matchOnDetail = true;

        quickPick.busy = true;
        const allResults = await this.loadAllData();
        quickPick.items = allResults;
        quickPick.busy = false;

        quickPick.onDidChangeValue(value => {
            if (!value) {
                quickPick.items = allResults;
            }
        });

        quickPick.onDidAccept(async () => {
            const selected = quickPick.selectedItems[0];
            if (selected) {
                quickPick.hide();
                await this.handleSelection(selected);
            }
        });

        quickPick.onDidHide(() => quickPick.dispose());
        quickPick.show();
    }

    async searchInvoices(): Promise<void> {
        const results = await this.loadInvoices();
        const selected = await vscode.window.showQuickPick(results, {
            placeHolder: 'Search invoices...',
            matchOnDescription: true,
            matchOnDetail: true
        });
        if (selected) await this.handleSelection(selected);
    }

    async searchContracts(): Promise<void> {
        const results = await this.loadContracts();
        const selected = await vscode.window.showQuickPick(results, {
            placeHolder: 'Search contracts...',
            matchOnDescription: true,
            matchOnDetail: true
        });
        if (selected) await this.handleSelection(selected);
    }

    async searchClients(): Promise<void> {
        const results = await this.loadClients();
        const selected = await vscode.window.showQuickPick(results, {
            placeHolder: 'Search clients...',
            matchOnDescription: true,
            matchOnDetail: true
        });
        if (selected) await this.handleSelection(selected);
    }

    private async loadAllData(): Promise<Array<SearchResult & vscode.QuickPickItem>> {
        const [invoices, contracts, clients, expenses] = await Promise.all([
            this.loadInvoices(),
            this.loadContracts(),
            this.loadClients(),
            this.loadExpenses()
        ]);
        return [...invoices, ...contracts, ...clients, ...expenses];
    }

    private async loadInvoices(): Promise<Array<SearchResult & vscode.QuickPickItem>> {
        const result = await this.cli.listInvoices();
        if (!result.success || !result.stdout) return [];
        const lines = result.stdout.trim().split('\n');
        if (lines.length < 2) return [];

        const items: Array<SearchResult & vscode.QuickPickItem> = [];
        for (const line of lines.slice(1)) {
            const parts = line.split(/\s{2,}/).filter((p: string) => p.trim());
            const id = parseInt(parts[0], 10);
            if (!isNaN(id)) {
                items.push({
                    type: 'invoice',
                    id,
                    label: `$(file-text) ${parts[1] || ''}`,
                    description: `${parts[2] || 'Unknown'} - ${parts[3] || ''}`,
                    detail: `Invoice - ${parts[4] || ''}`
                });
            }
        }
        return items;
    }

    private async loadContracts(): Promise<Array<SearchResult & vscode.QuickPickItem>> {
        const result = await this.cli.listContracts();
        if (!result.success || !result.stdout) return [];
        const lines = result.stdout.trim().split('\n');
        if (lines.length < 2) return [];

        const items: Array<SearchResult & vscode.QuickPickItem> = [];
        for (const line of lines.slice(1)) {
            const parts = line.split(/\s{2,}/).filter((p: string) => p.trim());
            const id = parseInt(parts[0], 10);
            if (!isNaN(id)) {
                const isActive = line.includes('✓');
                items.push({
                    type: 'contract',
                    id,
                    label: `$(file-code) ${parts[2] || 'Unknown'}`,
                    description: `${parts[3] || ''} - ${parts[4] || ''} - ${parts[5] || ''}`,
                    detail: `Contract - ${isActive ? 'Active' : 'Inactive'}`
                });
            }
        }
        return items;
    }

    private async loadClients(): Promise<Array<SearchResult & vscode.QuickPickItem>> {
        const result = await this.cli.listClients();
        if (!result.success || !result.stdout) return [];
        const lines = result.stdout.trim().split('\n');
        if (lines.length < 2) return [];

        const items: Array<SearchResult & vscode.QuickPickItem> = [];
        for (const line of lines.slice(1)) {
            const parts = line.split(/\s{2,}/).filter((p: string) => p.trim());
            const id = parseInt(parts[0], 10);
            if (!isNaN(id)) {
                items.push({
                    type: 'client',
                    id,
                    label: `$(person) ${parts[1] || 'Unknown'}`,
                    description: parts[2] || '',
                    detail: 'Client'
                });
            }
        }
        return items;
    }

    private async loadExpenses(): Promise<Array<SearchResult & vscode.QuickPickItem>> {
        const result = await this.cli.listExpenses();
        if (!result.success || !result.stdout) return [];
        const lines = result.stdout.trim().split('\n');
        if (lines.length < 2) return [];

        const items: Array<SearchResult & vscode.QuickPickItem> = [];
        for (const line of lines.slice(1)) {
            const parts = line.split(/\s{2,}/).filter((p: string) => p.trim());
            const id = parseInt(parts[0], 10);
            if (!isNaN(id)) {
                items.push({
                    type: 'expense',
                    id,
                    label: `$(credit-card) ${parts[4] || parts[3] || 'Expense'}`,
                    description: `${parts[1] || ''} - ${parts[2] || ''}`,
                    detail: `Expense - ${parts[3] || ''}`
                });
            }
        }
        return items;
    }

    private async handleSelection(selected: SearchResult): Promise<void> {
        const actions = this.getActionsForType(selected.type);
        const action = await vscode.window.showQuickPick(actions, {
            placeHolder: `Actions for ${selected.label.replace(/\$\([^)]+\)\s*/, '')}`
        });
        if (action) {
            await vscode.commands.executeCommand(action.command, { itemId: selected.id });
        }
    }

    private getActionsForType(type: string): Array<{ label: string; command: string }> {
        switch (type) {
            case 'invoice':
                return [
                    { label: '$(eye) View Invoice', command: 'ung.viewInvoice' },
                    { label: '$(file-pdf) Export PDF', command: 'ung.exportInvoice' },
                    { label: '$(mail) Email Invoice', command: 'ung.emailInvoice' },
                    { label: '$(check) Mark as Paid', command: 'ung.markInvoicePaid' },
                    { label: '$(edit) Edit Invoice', command: 'ung.editInvoice' }
                ];
            case 'contract':
                return [
                    { label: '$(eye) View Contract', command: 'ung.viewContract' },
                    { label: '$(file-pdf) Generate PDF', command: 'ung.generateContractPDF' },
                    { label: '$(edit) Edit Contract', command: 'ung.editContract' }
                ];
            case 'client':
                return [
                    { label: '$(edit) Edit Client', command: 'ung.editClient' },
                    { label: '$(trash) Delete Client', command: 'ung.deleteClient' }
                ];
            case 'expense':
                return [
                    { label: '$(edit) Edit Expense', command: 'ung.editExpense' },
                    { label: '$(trash) Delete Expense', command: 'ung.deleteExpense' }
                ];
            default:
                return [];
        }
    }
}
