import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';
import { StatusBarManager } from '../utils/statusBar';

/**
 * Time tracking command handlers
 */
export class TrackingCommands {
    constructor(
        private cli: UngCli,
        private statusBar: StatusBarManager,
        private refreshCallback?: () => void
    ) {}

    /**
     * Start time tracking
     */
    async startTracking(): Promise<void> {
        // Check if there's already an active session
        const currentResult = await this.cli.getCurrentSession();
        if (currentResult.success && currentResult.stdout && !currentResult.stdout.includes('No active')) {
            vscode.window.showWarningMessage('There is already an active tracking session. Stop it first.');
            return;
        }

        const project = await vscode.window.showInputBox({
            prompt: 'Project Name',
            placeHolder: 'e.g., Website Development'
        });

        const clientIdStr = await vscode.window.showInputBox({
            prompt: 'Client ID (optional)',
            placeHolder: 'Leave empty for no client'
        });

        const billable = await vscode.window.showQuickPick(
            ['Yes', 'No'],
            { placeHolder: 'Is this billable?' }
        );

        const notes = await vscode.window.showInputBox({
            prompt: 'Notes (optional)',
            placeHolder: 'What are you working on?'
        });

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Starting time tracking...',
            cancellable: false
        }, async () => {
            const result = await this.cli.startTracking({
                clientId: clientIdStr ? Number(clientIdStr) : undefined,
                project,
                billable: billable === 'Yes',
                notes
            });

            if (result.success) {
                vscode.window.showInformationMessage('Time tracking started!');
                await this.statusBar.forceUpdate();
                if (this.refreshCallback) {
                    this.refreshCallback();
                }
            } else {
                vscode.window.showErrorMessage(`Failed to start tracking: ${result.error}`);
            }
        });
    }

    /**
     * Stop time tracking
     */
    async stopTracking(): Promise<void> {
        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Stopping time tracking...',
            cancellable: false
        }, async () => {
            const result = await this.cli.stopTracking();

            if (result.success) {
                vscode.window.showInformationMessage('Time tracking stopped!');
                await this.statusBar.forceUpdate();
                if (this.refreshCallback) {
                    this.refreshCallback();
                }
            } else {
                vscode.window.showErrorMessage(`Failed to stop tracking: ${result.error}`);
            }
        });
    }

    /**
     * Log time manually
     */
    async logTimeManually(): Promise<void> {
        vscode.window.showInformationMessage(
            'Manual time logging is interactive. Please use the CLI: ung track log'
        );
    }

    /**
     * View active tracking session
     */
    async viewActiveSession(): Promise<void> {
        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Fetching active session...',
            cancellable: false
        }, async () => {
            const result = await this.cli.getCurrentSession();

            if (result.success && result.stdout) {
                // Show output in a new document
                const doc = await vscode.workspace.openTextDocument({
                    content: result.stdout,
                    language: 'plaintext'
                });
                await vscode.window.showTextDocument(doc);
            } else {
                vscode.window.showErrorMessage(`Failed to fetch session: ${result.error}`);
            }
        });
    }
}
