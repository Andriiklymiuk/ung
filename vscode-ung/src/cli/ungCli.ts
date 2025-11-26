import * as vscode from 'vscode';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

/**
 * Result from CLI execution
 */
export interface CliResult<T = any> {
    success: boolean;
    data?: T;
    error?: string;
    stdout?: string;
    stderr?: string;
}

/**
 * CLI wrapper for executing UNG commands
 */
export class UngCli {
    private outputChannel: vscode.OutputChannel;
    private readonly CLI_COMMAND = 'ung';

    constructor(outputChannel: vscode.OutputChannel) {
        this.outputChannel = outputChannel;
    }

    /**
     * Execute a UNG CLI command
     * @param args Command arguments
     * @param options Execution options
     * @returns Promise with result
     */
    async exec<T = any>(args: string[], options?: { parseJson?: boolean; cwd?: string }): Promise<CliResult<T>> {
        const parseJson = options?.parseJson ?? false;
        const cwd = options?.cwd ?? undefined;

        const command = `${this.CLI_COMMAND} ${args.join(' ')}`;
        this.outputChannel.appendLine(`> ${command}`);

        try {
            const { stdout, stderr } = await execAsync(command, {
                cwd,
                maxBuffer: 1024 * 1024 * 10 // 10MB buffer
            });

            if (stderr) {
                this.outputChannel.appendLine(`stderr: ${stderr}`);
            }

            if (stdout) {
                this.outputChannel.appendLine(stdout);
            }

            // Parse JSON output if requested
            let data: T | undefined;
            if (parseJson && stdout.trim()) {
                try {
                    data = JSON.parse(stdout);
                } catch (err) {
                    // Not valid JSON, return as string
                    data = stdout as any;
                }
            } else {
                data = stdout as any;
            }

            return {
                success: true,
                data,
                stdout,
                stderr
            };
        } catch (error: any) {
            const errorMessage = error.message || String(error);
            this.outputChannel.appendLine(`Error: ${errorMessage}`);

            return {
                success: false,
                error: errorMessage,
                stdout: error.stdout,
                stderr: error.stderr
            };
        }
    }

    /**
     * Check if UNG CLI is installed
     */
    async isInstalled(): Promise<boolean> {
        try {
            const result = await this.exec(['--version']);
            return result.success;
        } catch {
            return false;
        }
    }

    /**
     * Get CLI version
     */
    async getVersion(): Promise<string | null> {
        const result = await this.exec(['--version']);
        if (result.success && result.stdout) {
            return result.stdout.trim();
        }
        return null;
    }

    // ========== Company Commands ==========

    /**
     * List all companies
     */
    async listCompanies(): Promise<CliResult<any[]>> {
        return this.exec(['company', 'ls']);
    }

    /**
     * Create a new company
     */
    async createCompany(data: {
        name: string;
        email: string;
        phone?: string;
        address?: string;
        registrationAddress?: string;
        taxId?: string;
        bankName?: string;
        bankAccount?: string;
        bankSwift?: string;
    }): Promise<CliResult> {
        const args = ['company', 'add'];
        args.push('--name', data.name);
        args.push('--email', data.email);
        if (data.phone) args.push('--phone', data.phone);
        if (data.address) args.push('--address', data.address);
        if (data.registrationAddress) args.push('--registration-address', data.registrationAddress);
        if (data.taxId) args.push('--tax-id', data.taxId);
        if (data.bankName) args.push('--bank-name', data.bankName);
        if (data.bankAccount) args.push('--bank-account', data.bankAccount);
        if (data.bankSwift) args.push('--bank-swift', data.bankSwift);

        return this.exec(args);
    }

    // ========== Client Commands ==========

    /**
     * List all clients
     */
    async listClients(): Promise<CliResult<any[]>> {
        return this.exec(['client', 'ls']);
    }

    /**
     * Create a new client
     */
    async createClient(data: {
        name: string;
        email: string;
        address?: string;
        taxId?: string;
    }): Promise<CliResult> {
        const args = ['client', 'add'];
        args.push('--name', data.name);
        args.push('--email', data.email);
        if (data.address) args.push('--address', data.address);
        if (data.taxId) args.push('--tax-id', data.taxId);

        return this.exec(args);
    }

    /**
     * Delete a client
     */
    async deleteClient(id: number): Promise<CliResult> {
        return this.exec(['client', 'delete', id.toString()]);
    }

    /**
     * Edit a client
     */
    async editClient(id: number, data: {
        name?: string;
        email?: string;
        address?: string;
        taxId?: string;
    }): Promise<CliResult> {
        const args = ['client', 'edit', id.toString()];
        if (data.name) args.push('--name', `"${data.name}"`);
        if (data.email) args.push('--email', data.email);
        if (data.address) args.push('--address', `"${data.address}"`);
        if (data.taxId) args.push('--tax-id', data.taxId);

        return this.exec(args);
    }

    // ========== Company Commands (Edit) ==========

    /**
     * Edit a company
     */
    async editCompany(id: number, data: {
        name?: string;
        email?: string;
        address?: string;
        taxId?: string;
        bankName?: string;
        bankAccount?: string;
        bankSwift?: string;
    }): Promise<CliResult> {
        const args = ['company', 'edit', id.toString()];
        if (data.name) args.push('--name', `"${data.name}"`);
        if (data.email) args.push('--email', data.email);
        if (data.address) args.push('--address', `"${data.address}"`);
        if (data.taxId) args.push('--tax-id', data.taxId);
        if (data.bankName) args.push('--bank-name', `"${data.bankName}"`);
        if (data.bankAccount) args.push('--bank-account', data.bankAccount);
        if (data.bankSwift) args.push('--bank-swift', data.bankSwift);

        return this.exec(args);
    }

    // ========== Contract Commands ==========

    /**
     * List all contracts
     */
    async listContracts(): Promise<CliResult<any[]>> {
        return this.exec(['contract', 'ls']);
    }

    /**
     * Generate contract PDF
     */
    async generateContractPDF(id: number): Promise<CliResult> {
        return this.exec(['contract', 'pdf', id.toString()]);
    }

    /**
     * Email contract
     */
    async emailContract(id: number): Promise<CliResult> {
        return this.exec(['contract', 'email', id.toString()]);
    }

    /**
     * Edit a contract
     */
    async editContract(id: number, data: {
        name?: string;
        rate?: number;
        price?: number;
        currency?: string;
        active?: boolean;
    }): Promise<CliResult> {
        const args = ['contract', 'edit', id.toString()];
        if (data.name) args.push('--name', `"${data.name}"`);
        if (data.rate !== undefined) args.push('--rate', data.rate.toString());
        if (data.price !== undefined) args.push('--price', data.price.toString());
        if (data.currency) args.push('--currency', data.currency);
        if (data.active !== undefined) args.push('--active', data.active.toString());

        return this.exec(args);
    }

    // ========== Invoice Commands ==========

    /**
     * List all invoices
     */
    async listInvoices(): Promise<CliResult<any[]>> {
        return this.exec(['invoice', 'ls']);
    }

    /**
     * Create a new invoice (manual)
     */
    async createInvoice(data: {
        clientName: string;
        amount: number;
        currency?: string;
    }): Promise<CliResult> {
        const args = ['invoice', 'new'];
        args.push('--client-name', data.clientName);
        args.push('--price', data.amount.toString());
        if (data.currency) args.push('--currency', data.currency);

        return this.exec(args);
    }

    /**
     * Unified invoice action - generate from time, PDF, email
     */
    async invoiceAction(options: {
        client?: string;
        id?: number;
        pdf?: boolean;
        email?: boolean;
        emailApp?: string;
        batch?: boolean;
    }): Promise<CliResult> {
        const args = ['invoice'];
        if (options.client) args.push('--client', options.client);
        if (options.id) args.push('--id', options.id.toString());
        if (options.pdf) args.push('--pdf');
        if (options.email) args.push('--email');
        if (options.emailApp) args.push('--email-app', options.emailApp);
        if (options.batch) args.push('--batch');
        return this.exec(args);
    }

    /**
     * Generate invoice PDF for existing invoice
     */
    async generateInvoicePDF(id: number): Promise<CliResult> {
        return this.invoiceAction({ id, pdf: true });
    }

    /**
     * Email invoice
     */
    async emailInvoice(id: number, emailApp: string): Promise<CliResult> {
        return this.invoiceAction({ id, email: true, emailApp });
    }

    /**
     * Mark invoice as paid
     */
    async markInvoicePaid(id: number): Promise<CliResult> {
        return this.exec(['invoice', 'mark', id.toString(), '--status', 'paid']);
    }

    /**
     * Mark invoice as sent
     */
    async markInvoiceSent(id: number): Promise<CliResult> {
        return this.exec(['invoice', 'mark', id.toString(), '--status', 'sent']);
    }

    /**
     * Update invoice status
     */
    async updateInvoiceStatus(id: number, status: 'pending' | 'sent' | 'paid' | 'overdue'): Promise<CliResult> {
        return this.exec(['invoice', 'mark', id.toString(), '--status', status]);
    }

    /**
     * Generate invoice from time tracking
     */
    async generateInvoiceFromTime(clientName: string, options?: { pdf?: boolean; email?: boolean; emailApp?: string }): Promise<CliResult> {
        return this.invoiceAction({
            client: clientName,
            pdf: options?.pdf,
            email: options?.email,
            emailApp: options?.emailApp
        });
    }

    /**
     * Generate invoices for all clients with unbilled time
     * Note: This runs interactively in CLI, so for VS Code we use manual approach
     */
    async generateAllInvoices(options?: { pdf?: boolean; email?: boolean; emailApp?: string }): Promise<CliResult> {
        const args = ['invoice', 'generate-all'];
        if (options?.pdf) args.push('--pdf');
        if (options?.email) args.push('--email');
        if (options?.emailApp) args.push('--email-app', options.emailApp);
        return this.exec(args);
    }

    /**
     * Send emails for all pending invoices
     * Note: This runs interactively in CLI
     */
    async sendAllInvoices(emailApp?: string): Promise<CliResult> {
        const args = ['invoice', 'send-all'];
        if (emailApp) args.push('--email-app', emailApp);
        return this.exec(args);
    }

    /**
     * Get unbilled time summary (for VS Code bulk generation preview)
     */
    async getUnbilledTimeSummary(): Promise<CliResult> {
        return this.exec(['track', 'ls', '--unbilled']);
    }

    // ========== Expense Commands ==========

    /**
     * List all expenses
     */
    async listExpenses(): Promise<CliResult<any[]>> {
        return this.exec(['expense', 'ls']);
    }

    /**
     * Get expense report
     */
    async getExpenseReport(): Promise<CliResult> {
        return this.exec(['expense', 'report']);
    }

    // ========== Time Tracking Commands ==========

    /**
     * List tracking sessions
     */
    async listTrackingSessions(): Promise<CliResult<any[]>> {
        return this.exec(['track', 'ls']);
    }

    /**
     * List time entries (alias for tracking sessions)
     */
    async listTimeEntries(): Promise<CliResult<any[]>> {
        return this.exec(['track', 'ls']);
    }

    /**
     * Start time tracking
     */
    async startTracking(data: {
        clientId?: number;
        project?: string;
        billable?: boolean;
        notes?: string;
    }): Promise<CliResult> {
        const args = ['track', 'start'];
        if (data.clientId) args.push('--client', data.clientId.toString());
        if (data.project) args.push('--project', data.project);
        if (data.billable !== undefined) args.push('--billable', data.billable.toString());
        if (data.notes) args.push('--notes', data.notes);

        return this.exec(args);
    }

    /**
     * Stop time tracking
     */
    async stopTracking(): Promise<CliResult> {
        return this.exec(['track', 'stop']);
    }

    /**
     * Get current tracking session
     */
    async getCurrentSession(): Promise<CliResult> {
        return this.exec(['track', 'now']);
    }

    /**
     * Log time manually
     */
    async logTime(data: {
        contractId: number;
        hours: number;
        project?: string;
        notes?: string;
    }): Promise<CliResult> {
        const args = ['track', 'log'];
        args.push('--contract', data.contractId.toString());
        args.push('--hours', data.hours.toString());
        if (data.project) args.push('--project', data.project);
        if (data.notes) args.push('--notes', data.notes);

        return this.exec(args);
    }

    // ========== Recurring Invoice Commands ==========

    /**
     * List recurring invoices
     */
    async listRecurringInvoices(): Promise<CliResult> {
        return this.exec(['recurring', 'ls']);
    }

    /**
     * Generate due recurring invoices
     */
    async generateRecurringInvoices(options?: { all?: boolean; dryRun?: boolean }): Promise<CliResult> {
        const args = ['recurring', 'generate'];
        if (options?.all) args.push('--all');
        if (options?.dryRun) args.push('--dry-run');
        return this.exec(args);
    }

    /**
     * Pause a recurring invoice
     */
    async pauseRecurringInvoice(id: number): Promise<CliResult> {
        return this.exec(['recurring', 'pause', id.toString()]);
    }

    /**
     * Resume a recurring invoice
     */
    async resumeRecurringInvoice(id: number): Promise<CliResult> {
        return this.exec(['recurring', 'resume', id.toString()]);
    }

    /**
     * Delete a recurring invoice
     */
    async deleteRecurringInvoice(id: number): Promise<CliResult> {
        return this.exec(['recurring', 'delete', id.toString()]);
    }

    // ========== Pomodoro Commands ==========

    /**
     * Start pomodoro timer (note: runs in terminal for interactive experience)
     */
    async startPomodoro(options?: {
        workMinutes?: number;
        breakMinutes?: number;
        client?: string;
        project?: string;
        autoTrack?: boolean;
    }): Promise<CliResult> {
        const args = ['pomodoro'];
        if (options?.workMinutes) args.push('--work', options.workMinutes.toString());
        if (options?.breakMinutes) args.push('--break', options.breakMinutes.toString());
        if (options?.client) args.push('--client', options.client);
        if (options?.project) args.push('--project', options.project);
        if (options?.autoTrack) args.push('--track');
        return this.exec(args);
    }

    // ========== Export Commands ==========

    /**
     * Export data for accounting software
     */
    async exportData(format: string, dataTypes: string[], options?: { year?: number }): Promise<CliResult> {
        const args = ['export', '--format', format];
        for (const dt of dataTypes) {
            args.push(`--${dt}`);
        }
        if (options?.year) args.push('--year', options.year.toString());
        return this.exec(args);
    }

    // ========== Sync Commands ==========

    /**
     * Create a backup of all data
     */
    async createBackup(outputPath?: string): Promise<CliResult> {
        const args = ['sync', 'backup'];
        if (outputPath) args.push('--output', outputPath);
        return this.exec(args);
    }

    /**
     * List available backups
     */
    async listBackups(): Promise<CliResult> {
        return this.exec(['sync', 'ls']);
    }

    /**
     * Restore from backup (interactive, best run in terminal)
     */
    async restoreBackup(file?: string, force?: boolean): Promise<CliResult> {
        const args = ['sync', 'restore'];
        if (file) args.push(file);
        if (force) args.push('--force');
        return this.exec(args);
    }

    // ========== Import Commands ==========

    /**
     * Import data from CSV file
     */
    async importData(filePath: string, dataType: string, dryRun?: boolean): Promise<CliResult> {
        const args = ['import', '--file', filePath, '--type', dataType];
        if (dryRun) args.push('--dry-run');
        return this.exec(args);
    }

    /**
     * Import data from SQLite database (encrypted or not)
     */
    async importDatabase(filePath: string, password?: string): Promise<CliResult> {
        const args = ['import', 'db', filePath];
        if (password) args.push('--password', password);
        return this.exec(args);
    }

    // ========== Database Commands ==========

    /**
     * Get database info/statistics
     */
    async getDatabaseInfo(): Promise<CliResult> {
        return this.exec(['database', 'info']);
    }

    /**
     * Get current database path
     */
    async getDatabaseCurrent(): Promise<CliResult> {
        return this.exec(['database', 'current']);
    }

    /**
     * List available databases
     */
    async listDatabases(): Promise<CliResult> {
        return this.exec(['database', 'list']);
    }

    /**
     * Switch to a different database
     */
    async switchDatabase(dbPath: string): Promise<CliResult> {
        return this.exec(['database', 'switch', dbPath]);
    }

    // ========== Additional Tracking Commands ==========

    /**
     * Get unbilled time sessions (for invoicing)
     */
    async getUnbilledSessions(): Promise<CliResult> {
        return this.exec(['track', 'ls', '--unbilled']);
    }

    // ========== Report Commands ==========

    /**
     * Get monthly report
     */
    async getMonthlyReport(): Promise<CliResult> {
        return this.exec(['report', 'monthly']);
    }

    /**
     * Get revenue report
     */
    async getRevenueReport(): Promise<CliResult> {
        return this.exec(['report', 'revenue']);
    }

    /**
     * Get unpaid invoices report
     */
    async getUnpaidReport(): Promise<CliResult> {
        return this.exec(['report', 'unpaid']);
    }

    /**
     * Get overdue invoices report
     */
    async getOverdueReport(): Promise<CliResult> {
        return this.exec(['report', 'overdue']);
    }

    /**
     * Get dashboard data
     */
    async getDashboard(): Promise<CliResult> {
        return this.exec(['dashboard']);
    }

    // ========== Invoice Edit/Delete Commands ==========

    /**
     * Edit an invoice
     */
    async editInvoice(id: number, data: {
        amount?: number;
        dueDate?: string;
        description?: string;
    }): Promise<CliResult> {
        const args = ['invoice', 'edit', id.toString()];
        if (data.amount !== undefined) args.push('--amount', data.amount.toString());
        if (data.dueDate) args.push('--due', data.dueDate);
        if (data.description) args.push('--description', `"${data.description}"`);
        return this.exec(args);
    }

    /**
     * Delete an invoice (requires confirmation in terminal)
     * For VS Code, we handle confirmation separately
     */
    async deleteInvoice(id: number): Promise<CliResult> {
        // Note: CLI requires interactive confirmation
        // For VS Code, we'll handle confirmation via VS Code dialog
        return this.exec(['invoice', 'delete', id.toString()]);
    }

    // ========== Tracking Edit/Delete Commands ==========

    /**
     * Edit a tracking session
     */
    async editTrackingSession(id: number, data: {
        hours?: number;
        project?: string;
        notes?: string;
    }): Promise<CliResult> {
        const args = ['track', 'edit', id.toString()];
        if (data.hours !== undefined) args.push('--hours', data.hours.toString());
        if (data.project) args.push('--project', `"${data.project}"`);
        if (data.notes) args.push('--notes', `"${data.notes}"`);
        return this.exec(args);
    }

    /**
     * Delete a tracking session (soft delete)
     */
    async deleteTrackingSession(id: number): Promise<CliResult> {
        // Note: CLI requires interactive confirmation
        // For VS Code, we'll handle confirmation via VS Code dialog
        return this.exec(['track', 'delete', id.toString()]);
    }
}
