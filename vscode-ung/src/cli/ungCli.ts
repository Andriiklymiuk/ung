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
}
