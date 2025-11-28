import { exec } from 'node:child_process';
import { promisify } from 'node:util';
import * as vscode from 'vscode';
import { getSecureStorage } from '../utils/secureStorage';

const execAsync = promisify(exec);

// Environment variable name for database password
const UNG_DB_PASSWORD_ENV = 'UNG_DB_PASSWORD';

/**
 * Result from CLI execution
 */
export interface CliResult<T = unknown> {
  success: boolean;
  data?: T;
  error?: string;
  stdout?: string;
  stderr?: string;
}

/**
 * Get the workspace folder path for CLI commands
 * This enables the CLI to detect local .ung/ configuration
 */
function getWorkspaceCwd(): string | undefined {
  const workspaceFolders = vscode.workspace.workspaceFolders;
  if (workspaceFolders && workspaceFolders.length > 0) {
    return workspaceFolders[0].uri.fsPath;
  }
  return undefined;
}

/**
 * Check if useGlobalConfig setting is enabled
 */
function shouldUseGlobalConfig(): boolean {
  const config = vscode.workspace.getConfiguration('ung');
  return config.get<boolean>('useGlobalConfig', true);
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
   *
   * By default, commands use the ung.useGlobalConfig setting to determine
   * whether to use global (~/.ung/) or local (.ung/) configuration.
   * Use options.useGlobal to explicitly override.
   *
   * If options.password is provided, it will be passed to the CLI via the
   * UNG_DB_PASSWORD environment variable (not command-line args for security).
   */
  async exec<T = unknown>(
    args: string[],
    options?: {
      parseJson?: boolean;
      cwd?: string;
      useGlobal?: boolean;
      password?: string;
    }
  ): Promise<CliResult<T>> {
    const parseJson = options?.parseJson ?? false;
    // Default to workspace folder for local config detection
    const cwd = options?.cwd ?? getWorkspaceCwd();

    // Use setting if useGlobal not explicitly provided
    const useGlobal = options?.useGlobal ?? shouldUseGlobalConfig();

    // If useGlobal is set, add the --global flag
    const commandArgs = useGlobal ? ['--global', ...args] : args;

    const command = `${this.CLI_COMMAND} ${commandArgs.join(' ')}`;
    this.outputChannel.appendLine(`> ${command}${cwd ? ` (cwd: ${cwd})` : ''}`);

    // Build environment with optional password (more secure than command-line args)
    const env = { ...process.env };

    // Use password from options, or try VSCode secure storage
    let password = options?.password;
    if (!password) {
      try {
        const secureStorage = getSecureStorage();
        password = await secureStorage.getPassword();
      } catch {
        // Secure storage not initialized yet, ignore
      }
    }

    if (password) {
      env[UNG_DB_PASSWORD_ENV] = password;
    }

    try {
      const { stdout, stderr } = await execAsync(command, {
        cwd,
        maxBuffer: 1024 * 1024 * 10, // 10MB buffer
        env,
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
        } catch (_err) {
          // Not valid JSON, return as string
          data = stdout as T;
        }
      } else {
        data = stdout as T;
      }

      return {
        success: true,
        data,
        stdout,
        stderr,
      };
    } catch (error: unknown) {
      const errorMessage =
        error instanceof Error ? error.message : String(error);
      const execError = error as { stdout?: string; stderr?: string };
      this.outputChannel.appendLine(`Error: ${errorMessage}`);

      return {
        success: false,
        error: errorMessage,
        stdout: execError.stdout,
        stderr: execError.stderr,
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
  async listCompanies(): Promise<CliResult<unknown[]>> {
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
    if (data.registrationAddress)
      args.push('--registration-address', data.registrationAddress);
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
  async listClients(): Promise<CliResult<unknown[]>> {
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
  async editClient(
    id: number,
    data: {
      name?: string;
      email?: string;
      address?: string;
      taxId?: string;
    }
  ): Promise<CliResult> {
    const args = ['client', 'edit', id.toString()];
    if (data.name) {
      const escapedName = data.name.replace(/'/g, "'\\''");
      args.push('--name', `'${escapedName}'`);
    }
    if (data.email) args.push('--email', data.email);
    if (data.address) {
      const escapedAddress = data.address.replace(/'/g, "'\\''");
      args.push('--address', `'${escapedAddress}'`);
    }
    if (data.taxId) args.push('--tax-id', data.taxId);

    return this.exec(args);
  }

  // ========== Company Commands (Edit) ==========

  /**
   * Edit a company
   */
  async editCompany(
    id: number,
    data: {
      name?: string;
      email?: string;
      address?: string;
      taxId?: string;
      bankName?: string;
      bankAccount?: string;
      bankSwift?: string;
    }
  ): Promise<CliResult> {
    const args = ['company', 'edit', id.toString()];
    if (data.name) {
      const escapedName = data.name.replace(/'/g, "'\\''");
      args.push('--name', `'${escapedName}'`);
    }
    if (data.email) args.push('--email', data.email);
    if (data.address) {
      const escapedAddress = data.address.replace(/'/g, "'\\''");
      args.push('--address', `'${escapedAddress}'`);
    }
    if (data.taxId) args.push('--tax-id', data.taxId);
    if (data.bankName) {
      const escapedBankName = data.bankName.replace(/'/g, "'\\''");
      args.push('--bank-name', `'${escapedBankName}'`);
    }
    if (data.bankAccount) args.push('--bank-account', data.bankAccount);
    if (data.bankSwift) args.push('--bank-swift', data.bankSwift);

    return this.exec(args);
  }

  // ========== Contract Commands ==========

  /**
   * List all contracts
   */
  async listContracts(): Promise<CliResult<unknown[]>> {
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
  async editContract(
    id: number,
    data: {
      name?: string;
      rate?: number;
      price?: number;
      currency?: string;
      active?: boolean;
    }
  ): Promise<CliResult> {
    const args = ['contract', 'edit', id.toString()];
    if (data.name) {
      const escapedName = data.name.replace(/'/g, "'\\''");
      args.push('--name', `'${escapedName}'`);
    }
    if (data.rate !== undefined) args.push('--rate', data.rate.toString());
    if (data.price !== undefined) args.push('--price', data.price.toString());
    if (data.currency) args.push('--currency', data.currency);
    if (data.active !== undefined)
      args.push('--active', data.active.toString());

    return this.exec(args);
  }

  /**
   * Create a new contract
   * Note: Uses shell quoting for names with spaces
   */
  async createContract(data: {
    clientId: number;
    name: string;
    type: 'hourly' | 'fixed_price' | 'retainer';
    rate?: number;
    price?: number;
    currency?: string;
  }): Promise<CliResult> {
    const args = ['contract', 'add'];
    args.push('--client', data.clientId.toString());
    // Properly escape the name for shell - use single quotes to avoid shell interpretation
    const escapedName = data.name.replace(/'/g, "'\\''");
    args.push('--name', `'${escapedName}'`);
    args.push('--type', data.type);
    if (data.rate !== undefined) args.push('--rate', data.rate.toString());
    if (data.price !== undefined) args.push('--price', data.price.toString());
    if (data.currency) args.push('--currency', data.currency);
    // Add --yes to skip any interactive prompts
    args.push('--yes');

    return this.exec(args);
  }

  /**
   * Delete a contract
   */
  async deleteContract(id: number): Promise<CliResult> {
    return this.exec(['contract', 'delete', id.toString(), '--yes']);
  }

  // ========== Invoice Commands ==========

  /**
   * List all invoices
   */
  async listInvoices(): Promise<CliResult<unknown[]>> {
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
  async updateInvoiceStatus(
    id: number,
    status: 'pending' | 'sent' | 'paid' | 'overdue'
  ): Promise<CliResult> {
    return this.exec(['invoice', 'mark', id.toString(), '--status', status]);
  }

  /**
   * Generate invoice from time tracking
   */
  async generateInvoiceFromTime(
    clientName: string,
    options?: { pdf?: boolean; email?: boolean; emailApp?: string }
  ): Promise<CliResult> {
    return this.invoiceAction({
      client: clientName,
      pdf: options?.pdf,
      email: options?.email,
      emailApp: options?.emailApp,
    });
  }

  /**
   * Generate invoices for all clients with unbilled time
   * Note: This runs interactively in CLI, so for VS Code we use manual approach
   */
  async generateAllInvoices(options?: {
    pdf?: boolean;
    email?: boolean;
    emailApp?: string;
  }): Promise<CliResult> {
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
  async listExpenses(): Promise<CliResult<unknown[]>> {
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
  async listTrackingSessions(): Promise<CliResult<unknown[]>> {
    return this.exec(['track', 'ls']);
  }

  /**
   * List time entries (alias for tracking sessions)
   */
  async listTimeEntries(): Promise<CliResult<unknown[]>> {
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
    if (data.billable !== undefined)
      args.push('--billable', data.billable.toString());
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
  async generateRecurringInvoices(options?: {
    all?: boolean;
    dryRun?: boolean;
  }): Promise<CliResult> {
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
    if (options?.workMinutes)
      args.push('--work', options.workMinutes.toString());
    if (options?.breakMinutes)
      args.push('--break', options.breakMinutes.toString());
    if (options?.client) args.push('--client', options.client);
    if (options?.project) args.push('--project', options.project);
    if (options?.autoTrack) args.push('--track');
    return this.exec(args);
  }

  // ========== Export Commands ==========

  /**
   * Export data for accounting software
   */
  async exportData(
    format: string,
    dataTypes: string[],
    options?: { year?: number }
  ): Promise<CliResult> {
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
  async importData(
    filePath: string,
    dataType: string,
    dryRun?: boolean
  ): Promise<CliResult> {
    const args = ['import', '--file', filePath, '--type', dataType];
    if (dryRun) args.push('--dry-run');
    return this.exec(args);
  }

  /**
   * Import data from SQLite database (encrypted or not)
   * Password is passed via environment variable for security (not visible in process list)
   */
  async importDatabase(
    filePath: string,
    password?: string
  ): Promise<CliResult> {
    const args = ['import', 'db', filePath];
    // Pass password via env var (more secure than command-line args)
    return this.exec(args, { password });
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

  /**
   * Check if UNG is initialized (has a valid .ung directory with database or config)
   * Returns true if either local or global .ung exists with content
   */
  async isInitialized(): Promise<boolean> {
    try {
      // Run a simple command that requires initialization
      // If it succeeds, we're initialized
      const result = await this.exec(['company', 'list']);
      // Check if the output indicates the onboarding message
      if (result.stdout?.includes("hasn't been set up yet")) {
        return false;
      }
      return result.success;
    } catch {
      return false;
    }
  }

  /**
   * Initialize UNG with config init
   * @param global Whether to use global config
   */
  async initialize(global: boolean = true): Promise<CliResult> {
    const args = ['config', 'init'];
    if (global) {
      args.push('--global');
    }
    return this.exec(args, { useGlobal: false }); // Don't add --global flag twice
  }

  // ========== Invoice Edit/Delete Commands ==========

  /**
   * Edit an invoice
   */
  async editInvoice(
    id: number,
    data: {
      amount?: number;
      dueDate?: string;
      description?: string;
    }
  ): Promise<CliResult> {
    const args = ['invoice', 'edit', id.toString()];
    if (data.amount !== undefined)
      args.push('--amount', data.amount.toString());
    if (data.dueDate) args.push('--due', data.dueDate);
    if (data.description) {
      const escapedDesc = data.description.replace(/'/g, "'\\''");
      args.push('--description', `'${escapedDesc}'`);
    }
    return this.exec(args);
  }

  /**
   * Delete an invoice
   * Uses --yes flag to skip interactive confirmation (handled by VS Code dialog)
   */
  async deleteInvoice(id: number): Promise<CliResult> {
    // Use --yes flag to skip CLI confirmation since VS Code handles it
    return this.exec(['invoice', 'delete', id.toString(), '--yes']);
  }

  // ========== Tracking Edit/Delete Commands ==========

  /**
   * Edit a tracking session
   */
  async editTrackingSession(
    id: number,
    data: {
      hours?: number;
      project?: string;
      notes?: string;
    }
  ): Promise<CliResult> {
    const args = ['track', 'edit', id.toString()];
    if (data.hours !== undefined) args.push('--hours', data.hours.toString());
    if (data.project) {
      const escapedProject = data.project.replace(/'/g, "'\\''");
      args.push('--project', `'${escapedProject}'`);
    }
    if (data.notes) {
      const escapedNotes = data.notes.replace(/'/g, "'\\''");
      args.push('--notes', `'${escapedNotes}'`);
    }
    return this.exec(args);
  }

  /**
   * Delete a tracking session (soft delete)
   * Uses --yes flag to skip CLI confirmation since VS Code handles it
   */
  async deleteTrackingSession(id: number): Promise<CliResult> {
    return this.exec(['track', 'delete', id.toString(), '--yes']);
  }

  // ========== Security Commands ==========

  /**
   * Get security status (encryption and keychain info)
   */
  async getSecurityStatus(): Promise<CliResult> {
    return this.exec(['security', 'status']);
  }

  /**
   * Check if database is encrypted
   */
  async isEncrypted(): Promise<boolean> {
    const result = await this.exec(['security', 'status']);
    if (result.success && result.stdout) {
      const output = result.stdout.toLowerCase();
      // Make sure we're checking for actual encryption, not "not encrypted"
      // Look for patterns like "encrypted: yes" or "encryption: enabled"
      // and exclude patterns like "not encrypted" or "encryption: no"
      if (
        output.includes('not encrypted') ||
        output.includes('encryption: no')
      ) {
        return false;
      }
      // Check for positive encryption indicators
      return (
        output.includes('encrypted: yes') ||
        output.includes('encryption: yes') ||
        output.includes('encryption: enabled') ||
        (output.includes('encrypted') && !output.includes('not '))
      );
    }
    return false;
  }

  /**
   * Execute a command with the provided password
   * Useful for operations on encrypted databases
   */
  async execWithPassword<T = unknown>(
    args: string[],
    password: string,
    options?: { parseJson?: boolean; cwd?: string; useGlobal?: boolean }
  ): Promise<CliResult<T>> {
    return this.exec<T>(args, { ...options, password });
  }

  /**
   * Check if database requires a password and if one is available
   * Returns true if DB is not encrypted OR if password is available in keychain
   */
  async isPasswordAvailable(): Promise<boolean> {
    // Run security status to check encryption and keychain status
    const result = await this.exec(['security', 'status']);
    if (!result.success || !result.stdout) {
      return true; // Assume OK if we can't check
    }

    const output = result.stdout;
    const isEncrypted = output.includes('Encrypted');
    const hasKeychainPassword = output.includes('Password saved in');

    // If not encrypted, no password needed
    if (!isEncrypted) {
      return true;
    }

    // If encrypted and password is in keychain, we're good
    if (hasKeychainPassword) {
      return true;
    }

    // Encrypted but no password in keychain - need VSCode to provide it
    return false;
  }
}
