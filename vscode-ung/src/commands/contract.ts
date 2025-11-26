import * as os from 'node:os';
import * as path from 'node:path';
import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';

/**
 * Contract command handlers
 */
export class ContractCommands {
  private refreshCallback?: () => void;

  constructor(
    private cli: UngCli,
    refreshCallback?: () => void
  ) {
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

    const contract = this.parseContractFromOutput(
      result.stdout || '',
      contractId
    );
    if (!contract) {
      vscode.window.showErrorMessage(`Contract ${contractId} not found`);
      return;
    }

    // Show contract details in a quick pick with actions
    const actions = [
      { label: '$(file-pdf) Generate PDF', action: 'pdf' },
      { label: '$(mail) Email Contract', action: 'email' },
      { label: '$(close) Close', action: 'close' },
    ];

    const detail = `${contract.name} | ${contract.client} | ${contract.type} | ${contract.ratePrice} | ${contract.active ? 'Active' : 'Inactive'}`;

    const selected = await vscode.window.showQuickPick(actions, {
      placeHolder: `Contract: ${contract.contractNum}`,
      title: detail,
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
  private parseContractFromOutput(
    output: string,
    contractId: number
  ): {
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
      const parts = line
        .split(/\s{2,}/)
        .map((p) => p.trim())
        .filter((p) => p);
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
            active: parts[6] === '✓',
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
    const result = await this.cli.listContracts();

    if (!result.success || !result.stdout) {
      vscode.window.showErrorMessage('Failed to fetch contracts');
      return;
    }

    const contracts = this.parseContractList(result.stdout);
    if (contracts.length === 0) {
      vscode.window.showInformationMessage(
        'No contracts found. Create one first.'
      );
      return;
    }

    let contract:
      | {
          id: number;
          contractNum: string;
          name: string;
          client: string;
          type: string;
          ratePrice: string;
          active: boolean;
        }
      | undefined;
    if (contractId) {
      contract = contracts.find((c) => c.id === contractId);
    } else {
      // Select contract to edit
      const selected = await vscode.window.showQuickPick(
        contracts.map((c) => ({
          label: c.name,
          description: `${c.type} • ${c.ratePrice}`,
          detail: `${c.client} • ${c.active ? 'Active' : 'Inactive'}`,
          id: c.id,
        })),
        { placeHolder: 'Select a contract to edit' }
      );

      if (!selected) return;
      contract = contracts.find((c) => c.id === selected.id);
    }

    if (!contract) {
      vscode.window.showErrorMessage('Contract not found');
      return;
    }

    // Parse rate and price from ratePrice string
    const rateMatch = contract.ratePrice.match(/(\d+(?:\.\d+)?)\s*\/hr/i);
    const priceMatch = contract.ratePrice.match(
      /(\d+(?:\.\d+)?)\s*(?:USD|EUR|GBP|CHF|PLN)/i
    );
    const currencyMatch = contract.ratePrice.match(/(USD|EUR|GBP|CHF|PLN)/i);

    const currentRate = rateMatch ? parseFloat(rateMatch[1]) : undefined;
    const currentPrice = priceMatch ? parseFloat(priceMatch[1]) : undefined;
    const currentCurrency = currencyMatch ? currencyMatch[1] : 'USD';

    // Show edit options
    const editField = await vscode.window.showQuickPick(
      [
        { label: '$(edit) Edit Name', field: 'name', value: contract.name },
        {
          label: '$(symbol-number) Edit Hourly Rate',
          field: 'rate',
          value: currentRate?.toString() || '',
        },
        {
          label: '$(credit-card) Edit Fixed Price',
          field: 'price',
          value: currentPrice?.toString() || '',
        },
        {
          label: '$(globe) Edit Currency',
          field: 'currency',
          value: currentCurrency,
        },
        {
          label: '$(circle-filled) Toggle Active Status',
          field: 'active',
          value: contract.active ? 'true' : 'false',
        },
        { label: '$(edit) Edit All Fields', field: 'all', value: '' },
      ],
      { placeHolder: `Editing: ${contract.name}` }
    );

    if (!editField) return;

    const updates: {
      name?: string;
      rate?: number;
      price?: number;
      currency?: string;
      active?: boolean;
    } = {};

    if (editField.field === 'all') {
      // Edit all fields
      const newName = await vscode.window.showInputBox({
        prompt: 'Contract Name',
        value: contract.name,
        validateInput: (v) => (v ? null : 'Name is required'),
      });
      if (newName === undefined) return;
      if (newName !== contract.name) updates.name = newName;

      const newRate = await vscode.window.showInputBox({
        prompt: 'Hourly Rate (leave empty for fixed price contracts)',
        value: currentRate?.toString() || '',
        validateInput: (v) => {
          if (!v) return null;
          const num = parseFloat(v);
          return Number.isNaN(num) || num < 0
            ? 'Enter a valid positive number'
            : null;
        },
      });
      if (newRate === undefined) return;
      if (newRate && parseFloat(newRate) !== currentRate)
        updates.rate = parseFloat(newRate);

      const newPrice = await vscode.window.showInputBox({
        prompt: 'Fixed Price (leave empty for hourly contracts)',
        value: currentPrice?.toString() || '',
        validateInput: (v) => {
          if (!v) return null;
          const num = parseFloat(v);
          return Number.isNaN(num) || num < 0
            ? 'Enter a valid positive number'
            : null;
        },
      });
      if (newPrice === undefined) return;
      if (newPrice && parseFloat(newPrice) !== currentPrice)
        updates.price = parseFloat(newPrice);

      const newCurrency = await vscode.window.showQuickPick(
        ['USD', 'EUR', 'GBP', 'CHF', 'PLN'],
        { placeHolder: 'Select currency', title: `Current: ${currentCurrency}` }
      );
      if (newCurrency === undefined) return;
      if (newCurrency !== currentCurrency) updates.currency = newCurrency;

      const newActive = await vscode.window.showQuickPick(
        [
          { label: '$(pass-filled) Active', value: true },
          { label: '$(circle-slash) Inactive', value: false },
        ],
        { placeHolder: `Currently: ${contract.active ? 'Active' : 'Inactive'}` }
      );
      if (newActive === undefined) return;
      if (newActive.value !== contract.active) updates.active = newActive.value;
    } else if (editField.field === 'active') {
      // Toggle active status
      updates.active = !contract.active;
    } else if (editField.field === 'rate') {
      const newRate = await vscode.window.showInputBox({
        prompt: 'Hourly Rate',
        value: editField.value,
        validateInput: (v) => {
          if (!v) return 'Rate is required';
          const num = parseFloat(v);
          return Number.isNaN(num) || num < 0
            ? 'Enter a valid positive number'
            : null;
        },
      });
      if (newRate === undefined) return;
      if (parseFloat(newRate) !== currentRate)
        updates.rate = parseFloat(newRate);
    } else if (editField.field === 'price') {
      const newPrice = await vscode.window.showInputBox({
        prompt: 'Fixed Price',
        value: editField.value,
        validateInput: (v) => {
          if (!v) return 'Price is required';
          const num = parseFloat(v);
          return Number.isNaN(num) || num < 0
            ? 'Enter a valid positive number'
            : null;
        },
      });
      if (newPrice === undefined) return;
      if (parseFloat(newPrice) !== currentPrice)
        updates.price = parseFloat(newPrice);
    } else if (editField.field === 'currency') {
      const newCurrency = await vscode.window.showQuickPick(
        ['USD', 'EUR', 'GBP', 'CHF', 'PLN'],
        { placeHolder: 'Select currency' }
      );
      if (newCurrency === undefined) return;
      if (newCurrency !== currentCurrency) updates.currency = newCurrency;
    } else if (editField.field === 'name') {
      const newName = await vscode.window.showInputBox({
        prompt: 'Contract Name',
        value: editField.value,
        validateInput: (v) => (v ? null : 'Name is required'),
      });
      if (newName === undefined) return;
      if (newName !== contract.name) updates.name = newName;
    }

    if (Object.keys(updates).length === 0) {
      vscode.window.showInformationMessage('No changes made');
      return;
    }

    // Apply updates
    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Updating contract...',
        cancellable: false,
      },
      async () => {
        const editResult = await this.cli.editContract(contract!.id, updates);

        if (editResult.success) {
          vscode.window.showInformationMessage(
            `Contract "${contract!.name}" updated successfully!`
          );
          if (this.refreshCallback) {
            this.refreshCallback();
          }
        } else {
          vscode.window.showErrorMessage(
            `Failed to update contract: ${editResult.error}`
          );
        }
      }
    );
  }

  /**
   * Parse contract list from CLI output
   */
  private parseContractList(output: string): Array<{
    id: number;
    contractNum: string;
    name: string;
    client: string;
    type: string;
    ratePrice: string;
    active: boolean;
  }> {
    const lines = output.trim().split('\n');
    const contracts: Array<{
      id: number;
      contractNum: string;
      name: string;
      client: string;
      type: string;
      ratePrice: string;
      active: boolean;
    }> = [];

    if (lines.length < 2) return contracts;

    // Skip header line
    for (let i = 1; i < lines.length; i++) {
      const line = lines[i].trim();
      if (!line) continue;

      // Parse: ID  CONTRACT#  NAME  CLIENT  TYPE  RATE/PRICE  ACTIVE
      const parts = line
        .split(/\s{2,}/)
        .map((p) => p.trim())
        .filter((p) => p);
      if (parts.length >= 6) {
        const id = parseInt(parts[0], 10);
        if (!Number.isNaN(id)) {
          contracts.push({
            id,
            contractNum: parts[1],
            name: parts[2],
            client: parts[3],
            type: parts[4],
            ratePrice: parts[5],
            active: parts[6] === '✓',
          });
        }
      }
    }

    return contracts;
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

    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Generating contract PDF...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.generateContractPDF(contractId);

        if (result.success) {
          pdfPath = this.parsePDFPath(result.stdout || '');
        } else {
          vscode.window.showErrorMessage(
            `Failed to generate PDF: ${result.error}`
          );
        }
      }
    );

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
        await vscode.commands.executeCommand(
          'revealFileInOS',
          vscode.Uri.file(pdfPath)
        );
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
    const match = output.match(
      /PDF (?:generated successfully|saved to):\s*(.+\.pdf)/i
    );
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

    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Preparing contract email...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.emailContract(contractId);

        if (result.success) {
          vscode.window.showInformationMessage('Contract email prepared!');
        } else {
          vscode.window.showErrorMessage(
            `Failed to email contract: ${result.error}`
          );
        }
      }
    );
  }
}
