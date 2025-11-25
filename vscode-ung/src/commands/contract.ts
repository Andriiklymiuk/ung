import * as vscode from 'vscode';
import * as os from 'os';
import * as path from 'path';
import { UngCli } from '../cli/ungCli';

/**
 * Contract command handlers
 */
export class ContractCommands {
    private refreshCallback?: () => void;

    constructor(private cli: UngCli, refreshCallback?: () => void) {
        this.refreshCallback = refreshCallback;
    }

    /**
     * Refresh callback getter for potential future use
     */
    protected getRefreshCallback(): (() => void) | undefined {
        return this.refreshCallback;
    }

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

        // Get contract list and find the specific contract
        const result = await this.cli.listContracts();
        if (!result.success) {
            vscode.window.showErrorMessage('Failed to fetch contracts');
            return;
        }

        const contract = this.parseContractFromOutput(result.stdout || '', contractId);
        if (!contract) {
            vscode.window.showErrorMessage(`Contract ${contractId} not found`);
            return;
        }

        // Show contract details in a quick pick with actions
        const actions = [
            { label: '$(file-pdf) Generate PDF', action: 'pdf' },
            { label: '$(mail) Email Contract', action: 'email' },
            { label: '$(close) Close', action: 'close' }
        ];

        const detail = `${contract.name} | ${contract.client} | ${contract.type} | ${contract.ratePrice} | ${contract.active ? 'Active' : 'Inactive'}`;

        const selected = await vscode.window.showQuickPick(actions, {
            placeHolder: `Contract: ${contract.contractNum}`,
            title: detail
        });

        if (selected) {
            switch (selected.action) {
                case 'pdf':
                    await this.generateContractPDF(contractId);
                    break;
                case 'email':
                    await this.emailContract(contractId);
                    break;
            }
        }
    }

    /**
     * Parse a specific contract from CLI output
     */
    private parseContractFromOutput(output: string, contractId: number): {
        id: number;
        contractNum: string;
        name: string;
        client: string;
        type: string;
        ratePrice: string;
        active: boolean;
    } | null {
        const lines = output.trim().split('\n');
        if (lines.length < 2) return null;

        // Skip header line
        const dataLines = lines.slice(1);

        for (const line of dataLines) {
            // Parse: ID  CONTRACT#  NAME  CLIENT  TYPE  RATE/PRICE  ACTIVE
            const parts = line.split(/\s{2,}/).map(p => p.trim()).filter(p => p);
            if (parts.length >= 6) {
                const id = parseInt(parts[0], 10);
                if (id === contractId) {
                    return {
                        id,
                        contractNum: parts[1],
                        name: parts[2],
                        client: parts[3],
                        type: parts[4],
                        ratePrice: parts[5],
                        active: parts[6] === '✓'
                    };
                }
            }
        }

        return null;
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

        let pdfPath: string | undefined;

        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Generating contract PDF...',
            cancellable: false
        }, async () => {
            const result = await this.cli.generateContractPDF(contractId);

            if (result.success) {
                pdfPath = this.parsePDFPath(result.stdout || '');
            } else {
                vscode.window.showErrorMessage(`Failed to generate PDF: ${result.error}`);
            }
        });

        if (pdfPath) {
            // Auto-open the PDF
            await vscode.env.openExternal(vscode.Uri.file(pdfPath));

            // Also show notification with buttons
            const action = await vscode.window.showInformationMessage(
                `Contract PDF: ${pdfPath}`,
                'Open Again',
                'Show in Finder'
            );

            if (action === 'Open Again') {
                await vscode.env.openExternal(vscode.Uri.file(pdfPath));
            } else if (action === 'Show in Finder') {
                await vscode.commands.executeCommand('revealFileInOS', vscode.Uri.file(pdfPath));
            }
        } else {
            // Fallback if path parsing fails
            const contractsDir = path.join(os.homedir(), '.ung', 'contracts');
            const action = await vscode.window.showInformationMessage(
                'Contract PDF generated!',
                'Open Contracts Folder'
            );
            if (action === 'Open Contracts Folder') {
                await vscode.env.openExternal(vscode.Uri.file(contractsDir));
            }
        }
    }

    /**
     * Parse PDF path from CLI output
     */
    private parsePDFPath(output: string): string | undefined {
        // Match: "✓ PDF generated successfully: /path/to/file.pdf"
        const match = output.match(/PDF (?:generated successfully|saved to):\s*(.+\.pdf)/i);
        return match ? match[1].trim() : undefined;
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
