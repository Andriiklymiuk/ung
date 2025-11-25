import * as vscode from 'vscode';

/**
 * Notification types
 */
export enum NotificationType {
    SUCCESS = 'success',
    ERROR = 'error',
    WARNING = 'warning',
    INFO = 'info'
}

/**
 * Notification options
 */
export interface NotificationOptions {
    modal?: boolean;
    actions?: string[];
    timeout?: number;
}

/**
 * Centralized notification manager for consistent user feedback
 */
export class NotificationManager {
    private static pendingNotifications: Map<string, NodeJS.Timeout> = new Map();

    /**
     * Show success notification
     */
    static success(message: string, options?: NotificationOptions): Promise<string | undefined> {
        return this.show(NotificationType.SUCCESS, message, options);
    }

    /**
     * Show error notification
     */
    static error(message: string, options?: NotificationOptions): Promise<string | undefined> {
        return this.show(NotificationType.ERROR, message, options);
    }

    /**
     * Show warning notification
     */
    static warning(message: string, options?: NotificationOptions): Promise<string | undefined> {
        return this.show(NotificationType.WARNING, message, options);
    }

    /**
     * Show info notification
     */
    static info(message: string, options?: NotificationOptions): Promise<string | undefined> {
        return this.show(NotificationType.INFO, message, options);
    }

    /**
     * Show notification with progress indicator
     */
    static async withProgress<T>(
        title: string,
        task: () => Promise<T>,
        successMessage?: string,
        errorMessage?: string
    ): Promise<T> {
        try {
            const result = await vscode.window.withProgress(
                {
                    location: vscode.ProgressLocation.Notification,
                    title,
                    cancellable: false
                },
                async () => task()
            );

            if (successMessage) {
                this.success(successMessage);
            }

            return result;
        } catch (error) {
            const msg = errorMessage || `${title} failed`;
            this.error(`${msg}: ${error}`);
            throw error;
        }
    }

    /**
     * Show notification with cancellable progress
     */
    static async withCancellableProgress<T>(
        title: string,
        task: (token: vscode.CancellationToken) => Promise<T>,
        successMessage?: string,
        errorMessage?: string
    ): Promise<T | undefined> {
        try {
            const result = await vscode.window.withProgress(
                {
                    location: vscode.ProgressLocation.Notification,
                    title,
                    cancellable: true
                },
                async (progress, token) => {
                    if (token.isCancellationRequested) {
                        return undefined;
                    }
                    return task(token);
                }
            );

            if (result !== undefined && successMessage) {
                this.success(successMessage);
            }

            return result;
        } catch (error) {
            const msg = errorMessage || `${title} failed`;
            this.error(`${msg}: ${error}`);
            throw error;
        }
    }

    /**
     * Show confirmation dialog
     */
    static async confirm(
        message: string,
        options?: { detail?: string; confirmLabel?: string; cancelLabel?: string }
    ): Promise<boolean> {
        const result = await vscode.window.showWarningMessage(
            message,
            {
                modal: true,
                detail: options?.detail
            },
            options?.confirmLabel || 'Yes',
            options?.cancelLabel || 'No'
        );

        return result === (options?.confirmLabel || 'Yes');
    }

    /**
     * Show quick pick selection
     */
    static async select<T extends vscode.QuickPickItem>(
        items: T[],
        options?: vscode.QuickPickOptions
    ): Promise<T | undefined> {
        return vscode.window.showQuickPick(items, options);
    }

    /**
     * Show multi-select quick pick
     */
    static async selectMultiple<T extends vscode.QuickPickItem>(
        items: T[],
        options?: vscode.QuickPickOptions
    ): Promise<T[] | undefined> {
        return vscode.window.showQuickPick(items, {
            ...options,
            canPickMany: true
        });
    }

    /**
     * Show input box
     */
    static async input(options: vscode.InputBoxOptions): Promise<string | undefined> {
        return vscode.window.showInputBox(options);
    }

    /**
     * Show status bar message temporarily
     */
    static statusBar(message: string, timeoutMs: number = 3000): vscode.Disposable {
        return vscode.window.setStatusBarMessage(message, timeoutMs);
    }

    /**
     * Show throttled notification (avoid spam)
     */
    static throttled(
        key: string,
        type: NotificationType,
        message: string,
        throttleMs: number = 5000
    ): void {
        // Check if there's a pending notification with this key
        if (this.pendingNotifications.has(key)) {
            return;
        }

        // Show notification
        this.show(type, message);

        // Set throttle
        const timeout = setTimeout(() => {
            this.pendingNotifications.delete(key);
        }, throttleMs);

        this.pendingNotifications.set(key, timeout);
    }

    /**
     * Clear throttled notification
     */
    static clearThrottled(key: string): void {
        const timeout = this.pendingNotifications.get(key);
        if (timeout) {
            clearTimeout(timeout);
            this.pendingNotifications.delete(key);
        }
    }

    /**
     * Core notification display method
     */
    private static show(
        type: NotificationType,
        message: string,
        options?: NotificationOptions
    ): Promise<string | undefined> {
        const modalOptions = options?.modal ? { modal: true } : {};
        const actions = options?.actions || [];

        switch (type) {
            case NotificationType.SUCCESS:
                return vscode.window.showInformationMessage(message, modalOptions, ...actions);
            case NotificationType.ERROR:
                return vscode.window.showErrorMessage(message, modalOptions, ...actions);
            case NotificationType.WARNING:
                return vscode.window.showWarningMessage(message, modalOptions, ...actions);
            case NotificationType.INFO:
                return vscode.window.showInformationMessage(message, modalOptions, ...actions);
        }
    }

    /**
     * Show operation result with appropriate notification
     */
    static operationResult(
        success: boolean,
        successMessage: string,
        errorMessage: string
    ): void {
        if (success) {
            this.success(successMessage);
        } else {
            this.error(errorMessage);
        }
    }
}
