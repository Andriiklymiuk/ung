import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';

/**
 * Contract command handlers
 */
export class ContractCommands {
    constructor(private cli: UngCli, private _refreshCallback?: () => void) {}

    /**
     * Create a new contract
     */
    async createContract(): Promise<void> {
        vscode.window.showInformationMessage(
            'Contract creation is interactive. Please use the CLI: ung contract add'
        );
    }

    /**
     * View contract details
     */
    async viewContract(contractId?: number): Promise<void> {
        if (!contractId) {
            vscode.window.showErrorMessage('No contract selected');
            return;
        }

        vscode.window.showInformationMessage(`Contract ID: ${contractId}. Full details coming soon!`);
    }

    /**
     * Edit a contract
     */
    async editContract(contractId?: number): Promise<void> {
        if (!contractId) {
            vscode.window.showErrorMessage('No contract selected');
            return;
        }

        vscode.window.showInformationMessage(
            'Contract editing will be available in a future version. Use the CLI: ung contract edit ' + contractId
        );
    }

    /**
     * Delete a contract
     */
    async deleteContract(contractId?: number): Promise<void> {
        if (!contractId) {
            vscode.window.showErrorMessage('No contract selected');
            return;
        }

        vscode.window.showInformationMessage(
            'Contract deletion will be available in a future version. Use the CLI to manage contracts.'
        );
    }

    /**
     * Generate contract PDF
     */
    async generateContractPDF(contractId?: number): Promise<void> {
        if (!contractId) {
            vscode.window.showErrorMessage('No contract selected');
            return;
        }

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Generating contract PDF...',
            cancellable: false
        }, async () => {
            const result = await this.cli.generateContractPDF(contractId);

            if (result.success) {
                vscode.window.showInformationMessage('Contract PDF generated successfully!');
            } else {
                vscode.window.showErrorMessage(`Failed to generate PDF: ${result.error}`);
            }
        });
    }

    /**
     * Email contract
     */
    async emailContract(contractId?: number): Promise<void> {
        if (!contractId) {
            vscode.window.showErrorMessage('No contract selected');
            return;
        }

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Preparing contract email...',
            cancellable: false
        }, async () => {
            const result = await this.cli.emailContract(contractId);

            if (result.success) {
                vscode.window.showInformationMessage('Contract email prepared!');
            } else {
                vscode.window.showErrorMessage(`Failed to email contract: ${result.error}`);
            }
        });
    }
}
