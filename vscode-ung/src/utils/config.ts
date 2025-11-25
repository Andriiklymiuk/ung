import * as vscode from 'vscode';

/**
 * Configuration manager for UNG extension
 */
export class Config {
    /**
     * Get the CLI path
     */
    static getCliPath(): string {
        const config = vscode.workspace.getConfiguration('ung');
        return config.get<string>('cliPath', 'ung');
    }

    /**
     * Get the default currency
     */
    static getDefaultCurrency(): string {
        const config = vscode.workspace.getConfiguration('ung');
        return config.get<string>('defaultCurrency', 'USD');
    }

    /**
     * Check if auto-refresh is enabled
     */
    static isAutoRefreshEnabled(): boolean {
        const config = vscode.workspace.getConfiguration('ung');
        return config.get<boolean>('autoRefresh', true);
    }

    /**
     * Get the date format
     */
    static getDateFormat(): string {
        const config = vscode.workspace.getConfiguration('ung');
        return config.get<string>('dateFormat', 'YYYY-MM-DD');
    }

    /**
     * Check if status bar should be shown
     */
    static shouldShowStatusBar(): boolean {
        const config = vscode.workspace.getConfiguration('ung');
        return config.get<boolean>('showStatusBar', true);
    }

    /**
     * Update CLI path
     */
    static async setCliPath(path: string): Promise<void> {
        const config = vscode.workspace.getConfiguration('ung');
        await config.update('cliPath', path, vscode.ConfigurationTarget.Global);
    }

    /**
     * Update default currency
     */
    static async setDefaultCurrency(currency: string): Promise<void> {
        const config = vscode.workspace.getConfiguration('ung');
        await config.update('defaultCurrency', currency, vscode.ConfigurationTarget.Global);
    }
}
