import * as vscode from 'vscode';

/**
 * Custom error types for UNG extension
 */
export enum ErrorType {
  CLI_NOT_INSTALLED = 'CLI_NOT_INSTALLED',
  CLI_EXECUTION_FAILED = 'CLI_EXECUTION_FAILED',
  PARSE_ERROR = 'PARSE_ERROR',
  VALIDATION_ERROR = 'VALIDATION_ERROR',
  NOT_FOUND = 'NOT_FOUND',
  PERMISSION_DENIED = 'PERMISSION_DENIED',
  NETWORK_ERROR = 'NETWORK_ERROR',
  TIMEOUT = 'TIMEOUT',
  UNKNOWN = 'UNKNOWN',
}

/**
 * Custom error class for UNG operations
 */
export class UngError extends Error {
  constructor(
    public type: ErrorType,
    message: string,
    public details?: unknown
  ) {
    super(message);
    this.name = 'UngError';
  }
}

/**
 * Centralized error handler with user-friendly messages
 */
export class ErrorHandler {
  private static outputChannel: vscode.OutputChannel;

  /**
   * Initialize error handler with output channel
   */
  static initialize(outputChannel: vscode.OutputChannel): void {
    ErrorHandler.outputChannel = outputChannel;
  }

  /**
   * Handle error with appropriate user message
   */
  static handle(error: Error | UngError | string, context?: string): void {
    const errorMessage = ErrorHandler.getErrorMessage(error);
    const userMessage = ErrorHandler.getUserFriendlyMessage(error, context);

    // Log to output channel
    if (ErrorHandler.outputChannel) {
      const timestamp = new Date().toISOString();
      ErrorHandler.outputChannel.appendLine(
        `[${timestamp}] ERROR: ${errorMessage}`
      );
      if (error instanceof Error && error.stack) {
        ErrorHandler.outputChannel.appendLine(error.stack);
      }
    }

    // Show user-friendly message
    vscode.window.showErrorMessage(userMessage);
  }

  /**
   * Handle error with options
   */
  static async handleWithActions(
    error: Error | UngError | string,
    context?: string,
    actions?: string[]
  ): Promise<string | undefined> {
    const userMessage = ErrorHandler.getUserFriendlyMessage(error, context);
    return vscode.window.showErrorMessage(userMessage, ...(actions || []));
  }

  /**
   * Get raw error message
   */
  private static getErrorMessage(error: Error | UngError | string): string {
    if (typeof error === 'string') {
      return error;
    }
    return error.message;
  }

  /**
   * Get user-friendly error message
   */
  private static getUserFriendlyMessage(
    error: Error | UngError | string,
    context?: string
  ): string {
    const prefix = context ? `${context}: ` : '';

    if (typeof error === 'string') {
      return prefix + error;
    }

    if (error instanceof UngError) {
      switch (error.type) {
        case ErrorType.CLI_NOT_INSTALLED:
          return `${prefix}UNG CLI is not installed or not in PATH. Please install it first.`;
        case ErrorType.CLI_EXECUTION_FAILED:
          return `${prefix}Failed to execute UNG command. ${error.message}`;
        case ErrorType.PARSE_ERROR:
          return `${prefix}Failed to parse CLI output. The data format may have changed.`;
        case ErrorType.VALIDATION_ERROR:
          return `${prefix}Validation error: ${error.message}`;
        case ErrorType.NOT_FOUND:
          return `${prefix}Resource not found: ${error.message}`;
        case ErrorType.PERMISSION_DENIED:
          return `${prefix}Permission denied: ${error.message}`;
        case ErrorType.NETWORK_ERROR:
          return `${prefix}Network error: ${error.message}`;
        case ErrorType.TIMEOUT:
          return `${prefix}Operation timed out. Please try again.`;
        default:
          return `${prefix}${error.message}`;
      }
    }

    return `${prefix}${error.message}`;
  }

  /**
   * Create UngError from CLI result
   */
  static fromCliError(error: string, stderr?: string): UngError {
    if (error.includes('not found') || error.includes('command not found')) {
      return new UngError(ErrorType.CLI_NOT_INSTALLED, error);
    }
    if (error.includes('timeout') || error.includes('timed out')) {
      return new UngError(ErrorType.TIMEOUT, error);
    }
    if (error.includes('permission denied') || error.includes('EACCES')) {
      return new UngError(ErrorType.PERMISSION_DENIED, error);
    }
    return new UngError(ErrorType.CLI_EXECUTION_FAILED, error, { stderr });
  }

  /**
   * Validate input and throw error if invalid
   */
  static validateInput(value: string | undefined, fieldName: string): string {
    if (!value || value.trim() === '') {
      throw new UngError(
        ErrorType.VALIDATION_ERROR,
        `${fieldName} is required`
      );
    }
    return value.trim();
  }

  /**
   * Validate numeric input
   */
  static validateNumber(
    value: string | undefined,
    fieldName: string,
    min?: number,
    max?: number
  ): number {
    const validatedString = ErrorHandler.validateInput(value, fieldName);
    const num = Number(validatedString);

    if (Number.isNaN(num)) {
      throw new UngError(
        ErrorType.VALIDATION_ERROR,
        `${fieldName} must be a valid number`
      );
    }

    if (min !== undefined && num < min) {
      throw new UngError(
        ErrorType.VALIDATION_ERROR,
        `${fieldName} must be at least ${min}`
      );
    }

    if (max !== undefined && num > max) {
      throw new UngError(
        ErrorType.VALIDATION_ERROR,
        `${fieldName} must be at most ${max}`
      );
    }

    return num;
  }

  /**
   * Log warning without showing to user
   */
  static logWarning(message: string): void {
    if (ErrorHandler.outputChannel) {
      const timestamp = new Date().toISOString();
      ErrorHandler.outputChannel.appendLine(
        `[${timestamp}] WARNING: ${message}`
      );
    }
  }

  /**
   * Log info message
   */
  static logInfo(message: string): void {
    if (ErrorHandler.outputChannel) {
      const timestamp = new Date().toISOString();
      ErrorHandler.outputChannel.appendLine(`[${timestamp}] INFO: ${message}`);
    }
  }
}
