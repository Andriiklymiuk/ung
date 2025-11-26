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

    /**
     * View tracking session with action menu
     */
    async viewTrackingSession(sessionId?: number): Promise<void> {
        if (!sessionId) {
            vscode.window.showErrorMessage('No session selected');
            return;
        }

        const actions = [
            { label: '$(edit) Edit Session', action: 'edit' },
            { label: '$(trash) Delete Session', action: 'delete' },
            { label: '$(close) Close', action: 'close' }
        ];

        const selected = await vscode.window.showQuickPick(actions, {
            placeHolder: `Session #${sessionId}`,
            title: 'Session Actions'
        });

        if (selected) {
            switch (selected.action) {
                case 'edit':
                    await this.editTrackingSession(sessionId);
                    break;
                case 'delete':
                    await this.deleteTrackingSession(sessionId);
                    break;
            }
        }
    }

    /**
     * Edit a tracking session
     */
    async editTrackingSession(sessionId?: number): Promise<void> {
        if (!sessionId) {
            vscode.window.showErrorMessage('No session selected');
            return;
        }

        // Ask what to edit
        const editOptions = [
            { label: '$(symbol-number) Hours', value: 'hours' },
            { label: '$(folder) Project Name', value: 'project' },
            { label: '$(note) Notes', value: 'notes' }
        ];

        const selected = await vscode.window.showQuickPick(editOptions, {
            placeHolder: 'What would you like to edit?',
            title: `Edit Session #${sessionId}`
        });

        if (!selected) return;

        const editData: { hours?: number; project?: string; notes?: string } = {};

        switch (selected.value) {
            case 'hours': {
                const newHours = await vscode.window.showInputBox({
                    prompt: 'New hours value',
                    placeHolder: 'e.g., 2.5',
                    validateInput: value => {
                        if (!value) return 'Hours is required';
                        if (isNaN(Number(value))) return 'Must be a valid number';
                        if (Number(value) <= 0) return 'Hours must be greater than 0';
                        return null;
                    }
                });
                if (!newHours) return;
                editData.hours = Number(newHours);
                break;
            }
            case 'project': {
                const newProject = await vscode.window.showInputBox({
                    prompt: 'New project name',
                    placeHolder: 'e.g., Website Development'
                });
                if (!newProject) return;
                editData.project = newProject;
                break;
            }
            case 'notes': {
                const newNotes = await vscode.window.showInputBox({
                    prompt: 'New notes',
                    placeHolder: 'What did you work on?'
                });
                if (!newNotes) return;
                editData.notes = newNotes;
                break;
            }
        }

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Updating tracking session...',
            cancellable: false
        }, async () => {
            const result = await this.cli.editTrackingSession(sessionId, editData);

            if (result.success) {
                vscode.window.showInformationMessage('Session updated successfully!');
                if (this.refreshCallback) {
                    this.refreshCallback();
                }
            } else {
                vscode.window.showErrorMessage(`Failed to update session: ${result.error}`);
            }
        });
    }

    /**
     * Delete a tracking session
     */
    async deleteTrackingSession(sessionId?: number): Promise<void> {
        if (!sessionId) {
            vscode.window.showErrorMessage('No session selected');
            return;
        }

        const confirm = await vscode.window.showWarningMessage(
            `Delete tracking session #${sessionId}? This is a soft delete.`,
            { modal: true },
            'Yes, Delete',
            'Cancel'
        );

        if (confirm !== 'Yes, Delete') return;

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Deleting session...',
            cancellable: false
        }, async () => {
            const result = await this.cli.deleteTrackingSession(sessionId);

            if (result.success) {
                vscode.window.showInformationMessage('Session deleted successfully!');
                if (this.refreshCallback) {
                    this.refreshCallback();
                }
            } else {
                vscode.window.showErrorMessage(`Failed to delete session: ${result.error}`);
            }
        });
    }
}
