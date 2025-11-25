import * as vscode from 'vscode';
import { UngCli, CliResult } from '../cli/ungCli';
import { ErrorHandler, UngError, ErrorType } from './errors';
import { NotificationManager } from './notifications';
import { globalCommandQueue } from './commandQueue';

/**
 * Base command class with common patterns
 */
export abstract class BaseCommand {
    constructor(
        protected cli: UngCli,
        protected refreshCallback?: () => void
    ) {}

    /**
     * Execute CLI command with error handling and notifications
     */
    protected async executeCommand<T = any>(
        operation: () => Promise<CliResult<T>>,
        options?: {
            successMessage?: string;
            errorMessage?: string;
            showProgress?: boolean;
            progressTitle?: string;
            refresh?: boolean;
        }
    ): Promise<CliResult<T>> {
        try {
            let result: CliResult<T>;

            if (options?.showProgress) {
                result = await NotificationManager.withProgress(
                    options.progressTitle || 'Processing...',
                    operation
                );
            } else {
                result = await operation();
            }

            if (result.success) {
                if (options?.successMessage) {
                    NotificationManager.success(options.successMessage);
                }

                if (options?.refresh !== false && this.refreshCallback) {
                    this.refreshCallback();
                }
            } else {
                const errorMsg = options?.errorMessage || 'Operation failed';
                ErrorHandler.handle(
                    result.error || 'Unknown error',
                    errorMsg
                );
            }

            return result;
        } catch (error) {
            const errorMsg = options?.errorMessage || 'Operation failed';
            ErrorHandler.handle(error as Error, errorMsg);
            throw error;
        }
    }

    /**
     * Execute command through queue
     */
    protected async executeQueued<T = any>(
        operation: () => Promise<CliResult<T>>,
        options?: {
            successMessage?: string;
            errorMessage?: string;
            showProgress?: boolean;
            progressTitle?: string;
            refresh?: boolean;
            priority?: boolean;
            timeout?: number;
            maxRetries?: number;
        }
    ): Promise<CliResult<T>> {
        try {
            const result = await globalCommandQueue.enqueue(operation, {
                priority: options?.priority,
                timeout: options?.timeout,
                maxRetries: options?.maxRetries
            });

            if (result.success) {
                if (options?.successMessage) {
                    NotificationManager.success(options.successMessage);
                }

                if (options?.refresh !== false && this.refreshCallback) {
                    this.refreshCallback();
                }
            } else {
                const errorMsg = options?.errorMessage || 'Operation failed';
                ErrorHandler.handle(
                    result.error || 'Unknown error',
                    errorMsg
                );
            }

            return result;
        } catch (error) {
            const errorMsg = options?.errorMessage || 'Operation failed';
            ErrorHandler.handle(error as Error, errorMsg);
            throw error;
        }
    }

    /**
     * Validate ID parameter
     */
    protected validateId(id: number | undefined, entityName: string): number {
        if (id === undefined || id === null) {
            throw new UngError(
                ErrorType.VALIDATION_ERROR,
                `No ${entityName} selected`
            );
        }
        return id;
    }

    /**
     * Get input from user with validation
     */
    protected async getInput(
        prompt: string,
        options?: {
            placeholder?: string;
            defaultValue?: string;
            required?: boolean;
            validator?: (value: string) => string | null;
        }
    ): Promise<string | undefined> {
        return NotificationManager.input({
            prompt,
            placeHolder: options?.placeholder,
            value: options?.defaultValue,
            validateInput: (value) => {
                if (options?.required && !value.trim()) {
                    return 'This field is required';
                }
                if (options?.validator) {
                    return options.validator(value);
                }
                return null;
            }
        });
    }

    /**
     * Get number input from user
     */
    protected async getNumberInput(
        prompt: string,
        options?: {
            placeholder?: string;
            defaultValue?: string;
            min?: number;
            max?: number;
            required?: boolean;
        }
    ): Promise<number | undefined> {
        const value = await this.getInput(prompt, {
            placeholder: options?.placeholder,
            defaultValue: options?.defaultValue,
            required: options?.required,
            validator: (val) => {
                if (!val && !options?.required) {
                    return null;
                }

                const num = Number(val);
                if (isNaN(num)) {
                    return 'Must be a valid number';
                }

                if (options?.min !== undefined && num < options.min) {
                    return `Must be at least ${options.min}`;
                }

                if (options?.max !== undefined && num > options.max) {
                    return `Must be at most ${options.max}`;
                }

                return null;
            }
        });

        return value ? Number(value) : undefined;
    }

    /**
     * Get date input from user
     */
    protected async getDateInput(
        prompt: string,
        options?: {
            placeholder?: string;
            defaultValue?: string;
            required?: boolean;
        }
    ): Promise<string | undefined> {
        return this.getInput(prompt, {
            placeholder: options?.placeholder || 'YYYY-MM-DD',
            defaultValue: options?.defaultValue,
            required: options?.required,
            validator: (val) => {
                if (!val && !options?.required) {
                    return null;
                }

                const dateRegex = /^\d{4}-\d{2}-\d{2}$/;
                if (!dateRegex.test(val)) {
                    return 'Must be in YYYY-MM-DD format';
                }

                const date = new Date(val);
                if (isNaN(date.getTime())) {
                    return 'Invalid date';
                }

                return null;
            }
        });
    }

    /**
     * Show quick pick selection
     */
    protected async showQuickPick<T extends vscode.QuickPickItem>(
        items: T[],
        options?: vscode.QuickPickOptions
    ): Promise<T | undefined> {
        return NotificationManager.select(items, options);
    }

    /**
     * Show confirmation dialog
     */
    protected async confirm(
        message: string,
        options?: {
            detail?: string;
            confirmLabel?: string;
            cancelLabel?: string;
        }
    ): Promise<boolean> {
        return NotificationManager.confirm(message, options);
    }

    /**
     * Parse list output from CLI
     */
    protected parseListOutput(output: string, skipLines: number = 1): string[][] {
        const lines = output.split('\n').filter(line => line.trim());
        const data: string[][] = [];

        for (let i = skipLines; i < lines.length; i++) {
            const line = lines[i].trim();
            if (!line) continue;

            // Split by multiple spaces (2 or more)
            const parts = line.split(/\s{2,}/);
            data.push(parts);
        }

        return data;
    }

    /**
     * Parse CLI output for single entity
     */
    protected parseEntityOutput(output: string): Map<string, string> {
        const result = new Map<string, string>();
        const lines = output.split('\n');

        for (const line of lines) {
            const match = line.match(/^([^:]+):\s*(.+)$/);
            if (match) {
                result.set(match[1].trim(), match[2].trim());
            }
        }

        return result;
    }

    /**
     * Extract ID from CLI output
     */
    protected extractIdFromOutput(output: string): number | null {
        const idMatch = output.match(/ID[:\s]+(\d+)/i);
        if (idMatch) {
            return parseInt(idMatch[1], 10);
        }

        const createdMatch = output.match(/created.*#(\d+)/i);
        if (createdMatch) {
            return parseInt(createdMatch[1], 10);
        }

        return null;
    }

    /**
     * Handle delete operation with confirmation
     */
    protected async handleDelete(
        entityName: string,
        id: number,
        deleteOperation: () => Promise<CliResult>
    ): Promise<void> {
        const confirmed = await this.confirm(
            `Are you sure you want to delete this ${entityName}?`,
            {
                detail: `${entityName} ID: ${id}`,
                confirmLabel: 'Delete',
                cancelLabel: 'Cancel'
            }
        );

        if (!confirmed) {
            return;
        }

        await this.executeCommand(
            deleteOperation,
            {
                successMessage: `${entityName} deleted successfully`,
                errorMessage: `Failed to delete ${entityName}`,
                showProgress: true,
                progressTitle: `Deleting ${entityName}...`
            }
        );
    }

    /**
     * Create options for currency selection
     */
    protected getCurrencyOptions(): vscode.QuickPickItem[] {
        return [
            { label: 'USD', description: 'US Dollar' },
            { label: 'EUR', description: 'Euro' },
            { label: 'GBP', description: 'British Pound' },
            { label: 'UAH', description: 'Ukrainian Hryvnia' },
            { label: 'CAD', description: 'Canadian Dollar' },
            { label: 'AUD', description: 'Australian Dollar' }
        ];
    }
}
