import type { CliResult } from '../cli/ungCli';
import { ErrorHandler, ErrorType, UngError } from './errors';

/**
 * Command queue item
 */
interface QueuedCommand<T = unknown> {
  id: string;
  execute: () => Promise<CliResult<T>>;
  resolve: (result: CliResult<T>) => void;
  reject: (error: Error) => void;
  timeout?: number;
  retries: number;
  maxRetries: number;
}

/**
 * Command queue for managing CLI execution
 * Prevents concurrent CLI calls and provides retry logic
 */
export class CommandQueue {
  private queue: QueuedCommand[] = [];
  private processing: boolean = false;
  private currentCommand: QueuedCommand | null = null;
  private readonly DEFAULT_TIMEOUT = 30000; // 30 seconds
  private readonly DEFAULT_MAX_RETRIES = 2;

  /**
   * Add command to queue
   */
  async enqueue<T = unknown>(
    execute: () => Promise<CliResult<T>>,
    options?: {
      timeout?: number;
      maxRetries?: number;
      priority?: boolean;
    }
  ): Promise<CliResult<T>> {
    return new Promise((resolve, reject) => {
      const command: QueuedCommand<T> = {
        id: this.generateId(),
        execute,
        resolve,
        reject,
        timeout: options?.timeout || this.DEFAULT_TIMEOUT,
        retries: 0,
        maxRetries: options?.maxRetries ?? this.DEFAULT_MAX_RETRIES,
      };

      if (options?.priority) {
        this.queue.unshift(command as QueuedCommand);
      } else {
        this.queue.push(command as QueuedCommand);
      }

      this.processQueue();
    });
  }

  /**
   * Process the command queue
   */
  private async processQueue(): Promise<void> {
    if (this.processing || this.queue.length === 0) {
      return;
    }

    this.processing = true;

    while (this.queue.length > 0) {
      const command = this.queue.shift()!;
      this.currentCommand = command;

      try {
        const result = await this.executeWithTimeout(command);
        command.resolve(result);
      } catch (error) {
        if (command.retries < command.maxRetries && this.shouldRetry(error)) {
          // Retry the command
          command.retries++;
          ErrorHandler.logWarning(
            `Retrying command (attempt ${command.retries + 1}/${command.maxRetries + 1})`
          );
          this.queue.unshift(command);
        } else {
          command.reject(error as Error);
        }
      }

      this.currentCommand = null;
    }

    this.processing = false;
  }

  /**
   * Execute command with timeout
   */
  private async executeWithTimeout<T>(
    command: QueuedCommand<T>
  ): Promise<CliResult<T>> {
    if (!command.timeout) {
      return command.execute();
    }

    return new Promise((resolve, reject) => {
      const timeoutId = setTimeout(() => {
        reject(
          new UngError(
            ErrorType.TIMEOUT,
            `Command timed out after ${command.timeout}ms`
          )
        );
      }, command.timeout);

      command
        .execute()
        .then((result) => {
          clearTimeout(timeoutId);
          resolve(result);
        })
        .catch((error) => {
          clearTimeout(timeoutId);
          reject(error);
        });
    });
  }

  /**
   * Check if error is retryable
   */
  private shouldRetry(error: unknown): boolean {
    if (error instanceof UngError) {
      // Don't retry validation errors or not found errors
      if (
        error.type === ErrorType.VALIDATION_ERROR ||
        error.type === ErrorType.NOT_FOUND
      ) {
        return false;
      }
      // Retry timeouts and network errors
      if (
        error.type === ErrorType.TIMEOUT ||
        error.type === ErrorType.NETWORK_ERROR
      ) {
        return true;
      }
    }

    // Don't retry by default
    return false;
  }

  /**
   * Generate unique command ID
   */
  private generateId(): string {
    return `cmd_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  /**
   * Get queue status
   */
  getStatus(): {
    queueLength: number;
    processing: boolean;
    currentCommand: string | null;
  } {
    return {
      queueLength: this.queue.length,
      processing: this.processing,
      currentCommand: this.currentCommand?.id || null,
    };
  }

  /**
   * Clear the queue
   */
  clear(): void {
    this.queue.forEach((command) => {
      command.reject(new UngError(ErrorType.UNKNOWN, 'Queue was cleared'));
    });
    this.queue = [];
  }
}

/**
 * Global command queue instance
 */
export const globalCommandQueue = new CommandQueue();
