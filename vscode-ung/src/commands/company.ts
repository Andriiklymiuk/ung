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
    async editCompany(companyId?: number): Promise<void> {
        const result = await this.cli.listCompanies();

        if (!result.success || !result.stdout) {
            vscode.window.showErrorMessage('Failed to fetch companies');
            return;
        }

        const companies = this.parseCompanyList(result.stdout);
        if (companies.length === 0) {
            vscode.window.showInformationMessage('No companies found. Create one first.');
            return;
        }

        let company;
        if (companyId) {
            company = companies.find(c => c.id === companyId);
        } else {
            // Select company to edit
            const selected = await vscode.window.showQuickPick(
                companies.map(c => ({
                    label: c.name,
                    description: c.email,
                    detail: c.address || undefined,
                    id: c.id
                })),
                { placeHolder: 'Select a company to edit' }
            );

            if (!selected) return;
            company = companies.find(c => c.id === selected.id);
        }

        if (!company) {
            vscode.window.showErrorMessage('Company not found');
            return;
        }

        // Show edit options
        const editField = await vscode.window.showQuickPick([
            { label: '$(organization) Edit Name', field: 'name', value: company.name },
            { label: '$(mail) Edit Email', field: 'email', value: company.email },
            { label: '$(location) Edit Address', field: 'address', value: company.address || '' },
            { label: '$(law) Edit Tax ID', field: 'taxId', value: company.taxId || '' },
            { label: '$(credit-card) Edit Bank Name', field: 'bankName', value: company.bankName || '' },
            { label: '$(key) Edit Bank Account', field: 'bankAccount', value: company.bankAccount || '' },
            { label: '$(globe) Edit Bank SWIFT', field: 'bankSwift', value: company.bankSwift || '' },
            { label: '$(edit) Edit All Fields', field: 'all', value: '' }
        ], { placeHolder: `Editing: ${company.name}` });

        if (!editField) return;

        const updates: {
            name?: string;
            email?: string;
            address?: string;
            taxId?: string;
            bankName?: string;
            bankAccount?: string;
            bankSwift?: string;
        } = {};

        if (editField.field === 'all') {
            // Edit all fields
            const fields = [
                { key: 'name', prompt: 'Company Name', current: company.name, required: true },
                { key: 'email', prompt: 'Company Email', current: company.email, required: true, isEmail: true },
                { key: 'address', prompt: 'Address', current: company.address || '' },
                { key: 'taxId', prompt: 'Tax ID', current: company.taxId || '' },
                { key: 'bankName', prompt: 'Bank Name', current: company.bankName || '' },
                { key: 'bankAccount', prompt: 'Bank Account', current: company.bankAccount || '' },
                { key: 'bankSwift', prompt: 'Bank SWIFT/BIC', current: company.bankSwift || '' }
            ];

            for (const field of fields) {
                const newValue = await vscode.window.showInputBox({
                    prompt: field.prompt,
                    value: field.current,
                    validateInput: field.required
                        ? (field.isEmail ? (v => v && v.includes('@') ? null : 'Valid email required') : (v => v ? null : `${field.prompt} is required`))
                        : undefined
                });

                if (newValue === undefined) return; // Cancelled
                if (newValue !== field.current) {
                    (updates as any)[field.key] = newValue;
                }
            }
        } else {
            // Edit single field
            const isRequired = editField.field === 'name' || editField.field === 'email';
            const isEmail = editField.field === 'email';

            const newValue = await vscode.window.showInputBox({
                prompt: `Edit ${editField.field}`,
                value: editField.value,
                validateInput: isRequired
                    ? (isEmail ? (v => v && v.includes('@') ? null : 'Valid email required') : (v => v ? null : 'This field is required'))
                    : undefined
            });

            if (newValue === undefined) return;
            if (newValue !== editField.value) {
                (updates as any)[editField.field] = newValue;
            }
        }

        if (Object.keys(updates).length === 0) {
            vscode.window.showInformationMessage('No changes made');
            return;
        }

        // Apply updates
        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Updating company...',
            cancellable: false
        }, async () => {
            const editResult = await this.cli.editCompany(company!.id, updates);

            if (editResult.success) {
                vscode.window.showInformationMessage(`Company "${company!.name}" updated successfully!`);
            } else {
                vscode.window.showErrorMessage(`Failed to update company: ${editResult.error}`);
            }
        });
    }

    /**
     * Parse company list from CLI output
     */
    private parseCompanyList(output: string): Array<{
        id: number;
        name: string;
        email: string;
        address?: string;
        taxId?: string;
        bankName?: string;
        bankAccount?: string;
        bankSwift?: string;
    }> {
        const lines = output.split('\n').filter(line => line.trim());
        const companies: Array<{
            id: number;
            name: string;
            email: string;
            address?: string;
            taxId?: string;
            bankName?: string;
            bankAccount?: string;
            bankSwift?: string;
        }> = [];

        for (let i = 1; i < lines.length; i++) { // Skip header
            const line = lines[i].trim();
            if (!line) continue;

            const parts = line.split(/\s{2,}/);
            if (parts.length >= 3) {
                const id = parseInt(parts[0]);
                if (!isNaN(id)) {
                    companies.push({
                        id,
                        name: parts[1],
                        email: parts[2],
                        address: parts[3] || undefined,
                        taxId: parts[4] || undefined,
                        bankAccount: parts[5] || undefined
                    });
                }
            }
        }

        return companies;
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
