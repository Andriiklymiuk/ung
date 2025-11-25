import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';

/**
 * Client command handlers
 */
export class ClientCommands {
    constructor(private cli: UngCli, private refreshCallback?: () => void) {}

    /**
     * Create a new client
     */
    async createClient(): Promise<void> {
        const name = await vscode.window.showInputBox({
            prompt: 'Client Name',
            placeHolder: 'e.g., Acme Corp',
            validateInput: (value) => value ? null : 'Client name is required'
        });

        if (!name) return;

        const email = await vscode.window.showInputBox({
            prompt: 'Client Email',
            placeHolder: 'e.g., billing@acme.com',
            validateInput: (value) => {
                if (!value) return 'Email is required';
                if (!value.includes('@')) return 'Invalid email format';
                return null;
            }
        });

        if (!email) return;

        const address = await vscode.window.showInputBox({
            prompt: 'Address (optional)',
            placeHolder: 'e.g., 456 Client Ave, City, State'
        });

        const taxId = await vscode.window.showInputBox({
            prompt: 'Tax ID (optional)',
            placeHolder: 'e.g., 98-7654321'
        });

        // Show progress
        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Creating client...',
            cancellable: false
        }, async () => {
            const result = await this.cli.createClient({
                name,
                email,
                address,
                taxId
            });

            if (result.success) {
                vscode.window.showInformationMessage(`Client "${name}" created successfully!`);
                if (this.refreshCallback) {
                    this.refreshCallback();
                }
            } else {
                vscode.window.showErrorMessage(`Failed to create client: ${result.error}`);
            }
        });
    }

    /**
     * Delete a client
     */
    async deleteClient(clientId?: number): Promise<void> {
        if (!clientId) {
            vscode.window.showErrorMessage('No client selected');
            return;
        }

        const confirm = await vscode.window.showWarningMessage(
            `Are you sure you want to delete this client?`,
            { modal: true },
            'Yes', 'No'
        );

        if (confirm !== 'Yes') return;

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Deleting client...',
            cancellable: false
        }, async () => {
            const result = await this.cli.deleteClient(clientId);

            if (result.success) {
                vscode.window.showInformationMessage('Client deleted successfully!');
                if (this.refreshCallback) {
                    this.refreshCallback();
                }
            } else {
                vscode.window.showErrorMessage(`Failed to delete client: ${result.error}`);
            }
        });
    }

    /**
     * Edit a client
     */
    async editClient(clientId?: number): Promise<void> {
        vscode.window.showInformationMessage('Client editing will be available in a future version. Use the CLI: ung client edit [id]');
    }

    /**
     * List clients with quick pick
     */
    async listClients(): Promise<void> {
        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Fetching clients...',
            cancellable: false
        }, async () => {
            const result = await this.cli.listClients();

            if (result.success && result.stdout) {
                // Show output in a new document
                const doc = await vscode.workspace.openTextDocument({
                    content: result.stdout,
                    language: 'plaintext'
                });
                await vscode.window.showTextDocument(doc);
            } else {
                vscode.window.showErrorMessage(`Failed to list clients: ${result.error}`);
            }
        });
    }
}
