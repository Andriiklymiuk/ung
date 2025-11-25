import * as vscode from 'vscode';

/**
 * Configuration manager for UNG extension
 * Minimal configuration - most settings handled by CLI
 */
export class Config {
    /**
     * Check if auto-refresh is enabled
     */
    static isAutoRefreshEnabled(): boolean {
        const config = vscode.workspace.getConfiguration('ung');
        return config.get<boolean>('autoRefresh', true);
    }

    /**
     * Check if status bar should be shown
     */
    static shouldShowStatusBar(): boolean {
        const config = vscode.workspace.getConfiguration('ung');
        return config.get<boolean>('showStatusBar', true);
    }

    /**
     * Get default currency
     */
    static getDefaultCurrency(): string {
        const config = vscode.workspace.getConfiguration('ung');
        return config.get<string>('defaultCurrency', 'USD');
    }
}
