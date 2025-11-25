import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';

/**
 * Company command handlers
 */
export class CompanyCommands {
    constructor(private cli: UngCli) {}

    /**
     * Create a new company
     */
    async createCompany(): Promise<void> {
        const name = await vscode.window.showInputBox({
            prompt: 'Company Name',
            placeHolder: 'e.g., Acme Corporation',
            validateInput: (value) => value ? null : 'Company name is required'
        });

        if (!name) return;

        const email = await vscode.window.showInputBox({
            prompt: 'Company Email',
            placeHolder: 'e.g., contact@acme.com',
            validateInput: (value) => {
                if (!value) return 'Email is required';
                if (!value.includes('@')) return 'Invalid email format';
                return null;
            }
        });

        if (!email) return;

        const phone = await vscode.window.showInputBox({
            prompt: 'Phone (optional)',
            placeHolder: 'e.g., +1-555-0100'
        });

        const address = await vscode.window.showInputBox({
            prompt: 'Address (optional)',
            placeHolder: 'e.g., 123 Business St, City, State'
        });

        const registrationAddress = await vscode.window.showInputBox({
            prompt: 'Registration Address (optional)',
            placeHolder: 'Official registration address'
        });

        const taxId = await vscode.window.showInputBox({
            prompt: 'Tax ID (optional)',
            placeHolder: 'e.g., 12-3456789'
        });

        const bankName = await vscode.window.showInputBox({
            prompt: 'Bank Name (optional)',
            placeHolder: 'e.g., First National Bank'
        });

        const bankAccount = await vscode.window.showInputBox({
            prompt: 'Bank Account (optional)',
            placeHolder: 'Account number'
        });

        const bankSwift = await vscode.window.showInputBox({
            prompt: 'Bank SWIFT/BIC (optional)',
            placeHolder: 'e.g., ABCDUS33'
        });

        // Show progress
        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Creating company...',
            cancellable: false
        }, async () => {
            const result = await this.cli.createCompany({
                name,
                email,
                phone,
                address,
                registrationAddress,
                taxId,
                bankName,
                bankAccount,
                bankSwift
            });

            if (result.success) {
                vscode.window.showInformationMessage(`Company "${name}" created successfully!`);
            } else {
                vscode.window.showErrorMessage(`Failed to create company: ${result.error}`);
            }
        });
    }

    /**
     * Edit existing company
     */
    async editCompany(): Promise<void> {
        const result = await this.cli.listCompanies();

        if (!result.success || !result.stdout) {
            vscode.window.showErrorMessage('Failed to fetch companies');
            return;
        }

        // For now, just show a message that editing is not yet implemented
        // In a full implementation, you would parse the list and allow editing
        vscode.window.showInformationMessage('Company editing will be available in a future version. Use the CLI: ung company edit [id]');
    }

    /**
     * List all companies
     */
    async listCompanies(): Promise<void> {
        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Fetching companies...',
            cancellable: false
        }, async () => {
            const result = await this.cli.listCompanies();

            if (result.success && result.stdout) {
                // Show output in a new document
                const doc = await vscode.workspace.openTextDocument({
                    content: result.stdout,
                    language: 'plaintext'
                });
                await vscode.window.showTextDocument(doc);
            } else {
                vscode.window.showErrorMessage(`Failed to list companies: ${result.error}`);
            }
        });
    }
}
