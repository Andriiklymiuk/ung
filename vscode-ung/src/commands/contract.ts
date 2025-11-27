import * as os from 'node:os';
import * as path from 'node:path';
import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';
import { CURRENCIES, Formatter } from '../utils/formatting';

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
    // Get clients for selection
    const clientsResult = await this.cli.listClients();
    if (!clientsResult.success || !clientsResult.stdout) {
      vscode.window.showErrorMessage(
        'Failed to fetch clients. Please create a client first.'
      );
      return;
    }

    // Parse clients
    const clientLines = clientsResult.stdout.trim().split('\n');
    if (clientLines.length < 2) {
      vscode.window.showErrorMessage(
        'No clients found. Please create a client first.'
      );
      return;
    }

    const clients: Array<{ id: number; name: string; email: string }> = [];
    for (let i = 1; i < clientLines.length; i++) {
      const parts = clientLines[i]
        .split(/\s{2,}/)
        .map((p) => p.trim())
        .filter((p) => p);
      if (parts.length >= 3) {
        clients.push({
          id: parseInt(parts[0], 10),
          name: parts[1],
          email: parts[2],
        });
      }
    }

    if (clients.length === 0) {
      vscode.window.showErrorMessage(
        'No clients found. Please create a client first.'
      );
      return;
    }

    // Step 1: Select client
    const clientItems = clients.map((c) => ({
      label: c.name,
      description: c.email,
      id: c.id,
    }));

    const selectedClient = await vscode.window.showQuickPick(clientItems, {
      placeHolder: 'Select a client for this contract',
      title: 'Create Contract - Step 1: Select Client',
    });

    if (!selectedClient) return;

    // Step 2: Enter contract name
    const contractName = await vscode.window.showInputBox({
      prompt: 'Enter contract name',
      placeHolder: 'e.g., Website Development Q1 2025',
      title: 'Create Contract - Step 2: Contract Name',
      validateInput: (v) => (v ? null : 'Contract name is required'),
    });

    if (!contractName) return;

    // Step 3: Select contract type
    const contractTypes = [
      {
        label: 'Hourly Rate',
        description: 'Bill by the hour',
        value: 'hourly' as const,
      },
      {
        label: 'Fixed Price',
        description: 'One-time fixed payment',
        value: 'fixed_price' as const,
      },
      {
        label: 'Retainer',
        description: 'Monthly retainer fee',
        value: 'retainer' as const,
      },
    ];

    const selectedType = await vscode.window.showQuickPick(contractTypes, {
      placeHolder: 'Select contract type',
      title: 'Create Contract - Step 3: Contract Type',
    });

    if (!selectedType) return;

    // Step 4: Enter rate/price based on type
    let rate: number | undefined;
    let price: number | undefined;

    if (selectedType.value === 'hourly') {
      const rateStr = await vscode.window.showInputBox({
        prompt: 'Enter hourly rate',
        placeHolder: 'e.g., 75.00',
        title: 'Create Contract - Step 4: Hourly Rate',
        validateInput: (v) => {
          if (!v) return 'Hourly rate is required for hourly contracts';
          const num = parseFloat(v);
          return Number.isNaN(num) || num <= 0
            ? 'Enter a valid positive number'
            : null;
        },
      });

      if (!rateStr) return;
      rate = parseFloat(rateStr);
    } else if (selectedType.value === 'fixed_price') {
      const priceStr = await vscode.window.showInputBox({
        prompt: 'Enter fixed price',
        placeHolder: 'e.g., 5000.00',
        title: 'Create Contract - Step 4: Fixed Price',
        validateInput: (v) => {
          if (!v) return 'Fixed price is required for fixed-price contracts';
          const num = parseFloat(v);
          return Number.isNaN(num) || num <= 0
            ? 'Enter a valid positive number'
            : null;
        },
      });

      if (!priceStr) return;
      price = parseFloat(priceStr);
    } else {
      // Retainer - ask for monthly price
      const priceStr = await vscode.window.showInputBox({
        prompt: 'Enter monthly retainer amount',
        placeHolder: 'e.g., 2000.00',
        title: 'Create Contract - Step 4: Retainer Amount',
        validateInput: (v) => {
          if (!v) return null; // Optional for retainer
          const num = parseFloat(v);
          return Number.isNaN(num) || num < 0
            ? 'Enter a valid positive number'
            : null;
        },
      });

      if (priceStr) {
        price = parseFloat(priceStr);
      }
    }

    // Step 5: Select currency
    const currencies = [...CURRENCIES];
    const selectedCurrency = await vscode.window.showQuickPick(currencies, {
      placeHolder: 'Select currency',
      title: 'Create Contract - Step 5: Currency',
    });

    if (!selectedCurrency) return;

    // Create the contract
    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Creating contract...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.createContract({
          clientId: selectedClient.id,
          name: contractName,
          type: selectedType.value,
          rate,
          price,
          currency: selectedCurrency,
        });

        if (result.success) {
          vscode.window.showInformationMessage(
            `Contract "${contractName}" created successfully!`
          );
          if (this.refreshCallback) {
            this.refreshCallback();
          }
        } else {
          vscode.window.showErrorMessage(
            `Failed to create contract: ${result.error}`
          );
        }
      }
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
    const currencyPattern = CURRENCIES.join('|');
    const rateMatch = contract.ratePrice.match(/(\d+(?:\.\d+)?)\s*\/hr/i);
    const priceMatch = contract.ratePrice.match(
      new RegExp(`(\\d+(?:\\.\\d+)?)\\s*(?:${currencyPattern})`, 'i')
    );
    const currencyMatch = contract.ratePrice.match(
      new RegExp(`(${currencyPattern})`, 'i')
    );

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

      const newCurrency = await vscode.window.showQuickPick([...CURRENCIES], {
        placeHolder: 'Select currency',
        title: `Current: ${currentCurrency}`,
      });
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
      const newCurrency = await vscode.window.showQuickPick([...CURRENCIES], {
        placeHolder: 'Select currency',
      });
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
      // Let user select a contract to delete
      const result = await this.cli.listContracts();
      if (!result.success || !result.stdout) {
        vscode.window.showErrorMessage('Failed to fetch contracts');
        return;
      }

      const contracts = this.parseContractList(result.stdout);
      if (contracts.length === 0) {
        vscode.window.showInformationMessage('No contracts found.');
        return;
      }

      const selected = await vscode.window.showQuickPick(
        contracts.map((c) => ({
          label: c.name,
          description: `${c.type} • ${c.ratePrice}`,
          detail: `${c.client} • ${c.active ? 'Active' : 'Inactive'}`,
          id: c.id,
          contractNum: c.contractNum,
        })),
        { placeHolder: 'Select a contract to delete' }
      );

      if (!selected) return;
      contractId = selected.id;
    }

    // Get contract details for confirmation and potential revert
    const contractsResult = await this.cli.listContracts();
    let contractName = `Contract #${contractId}`;
    let contractData: {
      name: string;
      client: string;
      clientId?: number;
      type: string;
      ratePrice: string;
      active: boolean;
    } | null = null;

    if (contractsResult.success && contractsResult.stdout) {
      const contract = this.parseContractFromOutput(
        contractsResult.stdout,
        contractId
      );
      if (contract) {
        contractName = `${contract.contractNum} - ${contract.name}`;
        contractData = {
          name: contract.name,
          client: contract.client,
          type: contract.type,
          ratePrice: contract.ratePrice,
          active: contract.active,
        };
      }
    }

    // Confirm deletion
    const confirm = await vscode.window.showWarningMessage(
      `Are you sure you want to delete "${contractName}"? This action cannot be undone.`,
      { modal: true },
      'Delete'
    );

    if (confirm !== 'Delete') return;

    // Delete the contract
    let deleteSuccess = false;
    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Deleting contract...',
        cancellable: false,
      },
      async () => {
        const deleteResult = await this.cli.deleteContract(contractId!);

        if (deleteResult.success) {
          deleteSuccess = true;
          if (this.refreshCallback) {
            this.refreshCallback();
          }
        } else {
          vscode.window.showErrorMessage(
            `Failed to delete contract: ${deleteResult.error}`
          );
        }
      }
    );

    // Show success message with revert option
    if (deleteSuccess && contractData) {
      const action = await vscode.window.showInformationMessage(
        `Contract "${contractName}" deleted successfully!`,
        'Revert'
      );

      if (action === 'Revert') {
        await this.revertContractDeletion(contractData);
      }
    } else if (deleteSuccess) {
      vscode.window.showInformationMessage(
        `Contract "${contractName}" deleted successfully!`
      );
    }
  }

  /**
   * Revert a contract deletion by re-creating the contract
   */
  private async revertContractDeletion(contractData: {
    name: string;
    client: string;
    type: string;
    ratePrice: string;
    active: boolean;
  }): Promise<void> {
    // First, we need to find the client ID from the client name
    const clientsResult = await this.cli.listClients();
    if (!clientsResult.success || !clientsResult.stdout) {
      vscode.window.showErrorMessage(
        'Failed to revert: Could not fetch clients'
      );
      return;
    }

    // Parse clients to find the matching one
    const clientLines = clientsResult.stdout.trim().split('\n');
    let clientId: number | undefined;
    for (let i = 1; i < clientLines.length; i++) {
      const parts = clientLines[i]
        .split(/\s{2,}/)
        .map((p) => p.trim())
        .filter((p) => p);
      if (parts.length >= 2 && parts[1] === contractData.client) {
        clientId = parseInt(parts[0], 10);
        break;
      }
    }

    if (!clientId) {
      vscode.window.showErrorMessage(
        `Failed to revert: Client "${contractData.client}" not found`
      );
      return;
    }

    // Parse rate and price from ratePrice string
    const currencyPattern = CURRENCIES.join('|');
    const rateMatch = contractData.ratePrice.match(/(\d+(?:\.\d+)?)\s*\/hr/i);
    const priceMatch = contractData.ratePrice.match(
      new RegExp(`(\\d+(?:\\.\\d+)?)\\s*(?:${currencyPattern})`, 'i')
    );
    const currencyMatch = contractData.ratePrice.match(
      new RegExp(`(${currencyPattern})`, 'i')
    );

    const rate = rateMatch ? parseFloat(rateMatch[1]) : undefined;
    const price = priceMatch ? parseFloat(priceMatch[1]) : undefined;
    const currency = currencyMatch ? currencyMatch[1] : 'USD';

    // Map display type to CLI type
    let contractType: 'hourly' | 'fixed_price' | 'retainer' = 'hourly';
    const typeLower = contractData.type.toLowerCase();
    if (typeLower.includes('fixed') || typeLower.includes('price')) {
      contractType = 'fixed_price';
    } else if (typeLower.includes('retainer')) {
      contractType = 'retainer';
    }

    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Reverting contract deletion...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.createContract({
          clientId,
          name: contractData.name,
          type: contractType,
          rate,
          price,
          currency,
        });

        if (result.success) {
          vscode.window.showInformationMessage(
            `Contract "${contractData.name}" restored successfully!`
          );
          if (this.refreshCallback) {
            this.refreshCallback();
          }
        } else {
          vscode.window.showErrorMessage(
            `Failed to restore contract: ${result.error}`
          );
        }
      }
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
      const revealText = Formatter.getRevealInFileManagerText();
      const action = await vscode.window.showInformationMessage(
        `Contract PDF: ${pdfPath}`,
        'Open Again',
        revealText
      );

      if (action === 'Open Again') {
        await vscode.env.openExternal(vscode.Uri.file(pdfPath));
      } else if (action === revealText) {
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
