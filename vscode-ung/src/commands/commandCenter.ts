import * as vscode from 'vscode';

interface CommandItem extends vscode.QuickPickItem {
    command?: string;
    category?: string;
    args?: any[];
}

/**
 * Command Center - Unified access to all UNG commands
 * Groups commands by category for easy discovery
 */
export class CommandCenter {
    private static categories: Map<string, CommandItem[]> = new Map([
        ['invoices', [
            { label: '$(add) Create Invoice', command: 'ung.createInvoice', description: 'Create a new invoice' },
            { label: '$(clock) Generate from Time', command: 'ung.generateFromTime', description: 'Create invoice from tracked time' },
            { label: '$(files) Generate All', command: 'ung.generateAllInvoices', description: 'Generate invoices for all clients' },
            { label: '$(send) Send All Pending', command: 'ung.sendAllInvoices', description: 'Email all pending invoices' },
            { label: '$(search) Search Invoices', command: 'ung.searchInvoices', description: 'Find an invoice' },
        ]],
        ['time', [
            { label: '$(play) Start Tracking', command: 'ung.startTracking', description: 'Start the timer' },
            { label: '$(debug-stop) Stop Tracking', command: 'ung.stopTracking', description: 'Stop the timer' },
            { label: '$(add) Log Time Manually', command: 'ung.logTimeManually', description: 'Add time without timer' },
            { label: '$(eye) View Active Session', command: 'ung.viewActiveSession', description: 'See current session' },
            { label: '$(watch) Pomodoro Timer', command: 'ung.startPomodoro', description: 'Focus timer' },
        ]],
        ['clients', [
            { label: '$(person-add) Add Client', command: 'ung.createClient', description: 'Create a new client' },
            { label: '$(list-flat) View Clients', command: 'ung.listClients', description: 'List all clients' },
            { label: '$(search) Search Clients', command: 'ung.searchClients', description: 'Find a client' },
        ]],
        ['contracts', [
            { label: '$(file-add) Create Contract', command: 'ung.createContract', description: 'Create a new contract' },
            { label: '$(search) Search Contracts', command: 'ung.searchContracts', description: 'Find a contract' },
        ]],
        ['expenses', [
            { label: '$(add) Log Expense', command: 'ung.logExpense', description: 'Add a new expense' },
            { label: '$(graph) Expense Report', command: 'ung.viewExpenseReport', description: 'View expense summary' },
        ]],
        ['reports', [
            { label: '$(dashboard) Dashboard', command: 'ung.openDashboard', description: 'Main dashboard overview' },
            { label: '$(graph) Statistics', command: 'ung.openStatistics', description: 'Charts and analytics' },
            { label: '$(calendar) Weekly Report', command: 'ung.weeklyReport', description: 'This week summary' },
            { label: '$(history) Monthly Report', command: 'ung.monthlyReport', description: 'This month summary' },
            { label: '$(pulse) Goal Progress', command: 'ung.viewGoalStatus', description: 'Track income goals' },
        ]],
        ['data', [
            { label: '$(export) Export Data', command: 'ung.exportData', description: 'Export for accounting' },
            { label: '$(folder-opened) Import CSV', command: 'ung.importData', description: 'Import from spreadsheet' },
            { label: '$(cloud) Backup & Sync', command: 'ung.syncData', description: 'Backup your data' },
        ]],
        ['settings', [
            { label: '$(target) Set Income Goal', command: 'ung.setIncomeGoal', description: 'Set monthly target' },
            { label: '$(symbol-number) Calculate Rate', command: 'ung.calculateRate', description: 'Find your hourly rate' },
            { label: '$(organization) Edit Company', command: 'ung.editCompany', description: 'Update company info' },
            { label: '$(sync) Check for Updates', command: 'ung.checkForUpdates', description: 'Update UNG CLI' },
            { label: '$(book) Documentation', command: 'ung.openDocs', description: 'View help docs' },
        ]],
    ]);

    private static categoryLabels: Map<string, string> = new Map([
        ['invoices', 'Invoices'],
        ['time', 'Time Tracking'],
        ['clients', 'Clients'],
        ['contracts', 'Contracts'],
        ['expenses', 'Expenses'],
        ['reports', 'Reports & Analytics'],
        ['data', 'Data Management'],
        ['settings', 'Settings & Help'],
    ]);

    /**
     * Show the main command center with categories
     */
    static async show(): Promise<void> {
        const items: CommandItem[] = [];

        // Add category headers with their commands
        for (const [categoryKey, commands] of this.categories) {
            const label = this.categoryLabels.get(categoryKey) || categoryKey;

            // Add category header
            items.push({
                label: `$(folder) ${label}`,
                kind: vscode.QuickPickItemKind.Separator,
                category: categoryKey
            });

            // Add top 3 commands from each category for quick access
            items.push(...commands.slice(0, 3));
        }

        const selected = await vscode.window.showQuickPick(items, {
            placeHolder: 'What would you like to do?',
            title: 'UNG Command Center',
            matchOnDescription: true,
        });

        if (selected?.command) {
            vscode.commands.executeCommand(selected.command, ...(selected.args || []));
        } else if (selected?.category) {
            // If they clicked a category header, show that category
            await this.showCategory(selected.category);
        }
    }

    /**
     * Show commands for a specific category
     */
    static async showCategory(categoryKey: string): Promise<void> {
        const commands = this.categories.get(categoryKey);
        if (!commands) return;

        const label = this.categoryLabels.get(categoryKey) || categoryKey;

        const items: CommandItem[] = [
            { label: '$(arrow-left) Back to Command Center', command: 'ung.commandCenter' },
            { label: '', kind: vscode.QuickPickItemKind.Separator },
            ...commands
        ];

        const selected = await vscode.window.showQuickPick(items, {
            placeHolder: `${label} Commands`,
            title: `UNG - ${label}`,
            matchOnDescription: true,
        });

        if (selected?.command) {
            vscode.commands.executeCommand(selected.command, ...(selected.args || []));
        }
    }

    /**
     * Quick Actions - Most common operations
     */
    static async showQuickActions(): Promise<void> {
        const items: CommandItem[] = [
            { label: '$(play) Start Time Tracking', command: 'ung.startTracking', description: 'Begin tracking time' },
            { label: '$(debug-stop) Stop Time Tracking', command: 'ung.stopTracking', description: 'Stop the timer' },
            { label: '', kind: vscode.QuickPickItemKind.Separator },
            { label: '$(add) Create Invoice', command: 'ung.createInvoice', description: 'New invoice' },
            { label: '$(clock) Generate from Time', command: 'ung.generateFromTime', description: 'Invoice from tracked time' },
            { label: '', kind: vscode.QuickPickItemKind.Separator },
            { label: '$(person-add) Add Client', command: 'ung.createClient', description: 'New client' },
            { label: '$(file-add) Create Contract', command: 'ung.createContract', description: 'New contract' },
            { label: '$(add) Log Expense', command: 'ung.logExpense', description: 'Record expense' },
            { label: '', kind: vscode.QuickPickItemKind.Separator },
            { label: '$(dashboard) Open Dashboard', command: 'ung.openDashboard', description: 'View overview' },
            { label: '$(graph) View Statistics', command: 'ung.openStatistics', description: 'Charts & reports' },
            { label: '', kind: vscode.QuickPickItemKind.Separator },
            { label: '$(list-unordered) All Commands...', command: 'ung.commandCenter', description: 'Browse all features' },
        ];

        const selected = await vscode.window.showQuickPick(items, {
            placeHolder: 'Quick Actions',
            title: 'UNG Quick Actions',
            matchOnDescription: true,
        });

        if (selected?.command) {
            vscode.commands.executeCommand(selected.command, ...(selected.args || []));
        }
    }

    /**
     * Invoice-specific actions
     */
    static async showInvoiceActions(): Promise<void> {
        await this.showCategory('invoices');
    }

    /**
     * Time tracking actions
     */
    static async showTimeActions(): Promise<void> {
        await this.showCategory('time');
    }

    /**
     * Reports actions
     */
    static async showReportActions(): Promise<void> {
        await this.showCategory('reports');
    }
}
