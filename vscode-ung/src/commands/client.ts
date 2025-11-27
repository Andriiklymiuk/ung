import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';

/**
 * Client command handlers
 */
export class ClientCommands {
  constructor(
    private cli: UngCli,
    private refreshCallback?: () => void
  ) {}

  /**
   * Create a new client
   */
  async createClient(): Promise<void> {
    const name = await vscode.window.showInputBox({
      prompt: 'Client Name',
      placeHolder: 'e.g., Acme Corp',
      validateInput: (value) => (value ? null : 'Client name is required'),
    });

    if (!name) return;

    const email = await vscode.window.showInputBox({
      prompt: 'Client Email',
      placeHolder: 'e.g., billing@acme.com',
      validateInput: (value) => {
        if (!value) return 'Email is required';
        if (!value.includes('@')) return 'Invalid email format';
        return null;
      },
    });

    if (!email) return;

    const address = await vscode.window.showInputBox({
      prompt: 'Address (optional)',
      placeHolder: 'e.g., 456 Client Ave, City, State',
    });

    const taxId = await vscode.window.showInputBox({
      prompt: 'Tax ID (optional)',
      placeHolder: 'e.g., 98-7654321',
    });

    // Show progress
    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Creating client...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.createClient({
          name,
          email,
          address,
          taxId,
        });

        if (result.success) {
          vscode.window.showInformationMessage(
            `Client "${name}" created successfully!`
          );
          if (this.refreshCallback) {
            this.refreshCallback();
          }
        } else {
          vscode.window.showErrorMessage(
            `Failed to create client: ${result.error}`
          );
        }
      }
    );
  }

  /**
   * Delete a client
   */
  async deleteClient(clientId?: number): Promise<void> {
    if (!clientId) {
      vscode.window.showErrorMessage('No client selected');
      return;
    }

    // Get client details for confirmation and potential revert
    const clientsResult = await this.cli.listClients();
    let clientData: {
      name: string;
      email: string;
      address?: string;
      taxId?: string;
    } | null = null;

    if (clientsResult.success && clientsResult.stdout) {
      const clients = this.parseClientList(clientsResult.stdout);
      const client = clients.find((c) => c.id === clientId);
      if (client) {
        clientData = {
          name: client.name,
          email: client.email,
          address: client.address,
          taxId: client.taxId,
        };
      }
    }

    const confirm = await vscode.window.showWarningMessage(
      clientData
        ? `Are you sure you want to delete "${clientData.name}"?`
        : `Are you sure you want to delete this client?`,
      { modal: true },
      'Yes',
      'No'
    );

    if (confirm !== 'Yes') return;

    let deleteSuccess = false;
    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Deleting client...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.deleteClient(clientId);

        if (result.success) {
          deleteSuccess = true;
          if (this.refreshCallback) {
            this.refreshCallback();
          }
        } else {
          vscode.window.showErrorMessage(
            `Failed to delete client: ${result.error}`
          );
        }
      }
    );

    // Show success message with revert option
    if (deleteSuccess && clientData) {
      const action = await vscode.window.showInformationMessage(
        `Client "${clientData.name}" deleted successfully!`,
        'Revert'
      );

      if (action === 'Revert') {
        await this.revertClientDeletion(clientData);
      }
    } else if (deleteSuccess) {
      vscode.window.showInformationMessage('Client deleted successfully!');
    }
  }

  /**
   * Revert a client deletion by re-creating the client
   */
  private async revertClientDeletion(clientData: {
    name: string;
    email: string;
    address?: string;
    taxId?: string;
  }): Promise<void> {
    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Reverting client deletion...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.createClient({
          name: clientData.name,
          email: clientData.email,
          address: clientData.address,
          taxId: clientData.taxId,
        });

        if (result.success) {
          vscode.window.showInformationMessage(
            `Client "${clientData.name}" restored successfully!`
          );
          if (this.refreshCallback) {
            this.refreshCallback();
          }
        } else {
          vscode.window.showErrorMessage(
            `Failed to restore client: ${result.error}`
          );
        }
      }
    );
  }

  /**
   * Edit a client
   */
  async editClient(clientId?: number): Promise<void> {
    if (!clientId) {
      // Let user select a client first
      const result = await this.cli.listClients();
      if (!result.success || !result.stdout) {
        vscode.window.showErrorMessage('Failed to fetch clients');
        return;
      }

      const clients = this.parseClientList(result.stdout);
      if (clients.length === 0) {
        vscode.window.showInformationMessage(
          'No clients found. Create one first.'
        );
        return;
      }

      const selected = await vscode.window.showQuickPick(
        clients.map((c) => ({
          label: c.name,
          description: c.email,
          detail: c.address || undefined,
          id: c.id,
        })),
        { placeHolder: 'Select a client to edit' }
      );

      if (!selected) return;
      clientId = selected.id;
    }

    // Fetch current client data
    const result = await this.cli.listClients();
    if (!result.success || !result.stdout) {
      vscode.window.showErrorMessage('Failed to fetch client data');
      return;
    }

    const clients = this.parseClientList(result.stdout);
    const client = clients.find((c) => c.id === clientId);
    if (!client) {
      vscode.window.showErrorMessage('Client not found');
      return;
    }

    // Show edit options
    const editField = await vscode.window.showQuickPick(
      [
        { label: '$(edit) Edit Name', field: 'name', value: client.name },
        { label: '$(mail) Edit Email', field: 'email', value: client.email },
        {
          label: '$(location) Edit Address',
          field: 'address',
          value: client.address || '',
        },
        {
          label: '$(law) Edit Tax ID',
          field: 'taxId',
          value: client.taxId || '',
        },
        { label: '$(edit) Edit All Fields', field: 'all', value: '' },
      ],
      { placeHolder: `Editing: ${client.name}` }
    );

    if (!editField) return;

    const updates: {
      name?: string;
      email?: string;
      address?: string;
      taxId?: string;
    } = {};

    if (editField.field === 'all') {
      // Edit all fields
      const newName = await vscode.window.showInputBox({
        prompt: 'Client Name',
        value: client.name,
        validateInput: (v) => (v ? null : 'Name is required'),
      });
      if (newName === undefined) return;
      if (newName !== client.name) updates.name = newName;

      const newEmail = await vscode.window.showInputBox({
        prompt: 'Client Email',
        value: client.email,
        validateInput: (v) =>
          v?.includes('@') ? null : 'Valid email is required',
      });
      if (newEmail === undefined) return;
      if (newEmail !== client.email) updates.email = newEmail;

      const newAddress = await vscode.window.showInputBox({
        prompt: 'Address (optional)',
        value: client.address || '',
      });
      if (newAddress === undefined) return;
      if (newAddress !== (client.address || '')) updates.address = newAddress;

      const newTaxId = await vscode.window.showInputBox({
        prompt: 'Tax ID (optional)',
        value: client.taxId || '',
      });
      if (newTaxId === undefined) return;
      if (newTaxId !== (client.taxId || '')) updates.taxId = newTaxId;
    } else {
      // Edit single field
      const newValue = await vscode.window.showInputBox({
        prompt: `Edit ${editField.field}`,
        value: editField.value,
        validateInput:
          editField.field === 'email'
            ? (v) => (v?.includes('@') ? null : 'Valid email is required')
            : editField.field === 'name'
              ? (v) => (v ? null : 'Name is required')
              : undefined,
      });

      if (newValue === undefined) return;
      if (newValue !== editField.value) {
        updates[editField.field as keyof typeof updates] = newValue;
      }
    }

    if (Object.keys(updates).length === 0) {
      vscode.window.showInformationMessage('No changes made');
      return;
    }

    // Apply updates
    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Updating client...',
        cancellable: false,
      },
      async () => {
        const editResult = await this.cli.editClient(clientId!, updates);

        if (editResult.success) {
          vscode.window.showInformationMessage(
            `Client "${client.name}" updated successfully!`
          );
          if (this.refreshCallback) {
            this.refreshCallback();
          }
        } else {
          vscode.window.showErrorMessage(
            `Failed to update client: ${editResult.error}`
          );
        }
      }
    );
  }

  /**
   * Parse client list from CLI output
   */
  private parseClientList(output: string): Array<{
    id: number;
    name: string;
    email: string;
    address?: string;
    taxId?: string;
  }> {
    const lines = output.split('\n').filter((line) => line.trim());
    const clients: Array<{
      id: number;
      name: string;
      email: string;
      address?: string;
      taxId?: string;
    }> = [];

    for (let i = 1; i < lines.length; i++) {
      // Skip header
      const line = lines[i].trim();
      if (!line) continue;

      const parts = line.split(/\s{2,}/);
      if (parts.length >= 3) {
        const id = parseInt(parts[0], 10);
        if (!Number.isNaN(id)) {
          clients.push({
            id,
            name: parts[1],
            email: parts[2],
            address: parts[3] || undefined,
            taxId: parts[4] || undefined,
          });
        }
      }
    }

    return clients;
  }

  /**
   * View a client with action menu
   */
  async viewClient(clientId?: number): Promise<void> {
    if (!clientId) {
      // Let user select a client first
      const result = await this.cli.listClients();
      if (!result.success || !result.stdout) {
        vscode.window.showErrorMessage('Failed to fetch clients');
        return;
      }

      const clients = this.parseClientList(result.stdout);
      if (clients.length === 0) {
        vscode.window.showInformationMessage(
          'No clients found. Create one first.'
        );
        return;
      }

      const selected = await vscode.window.showQuickPick(
        clients.map((c) => ({
          label: c.name,
          description: c.email,
          detail: c.address || undefined,
          id: c.id,
        })),
        { placeHolder: 'Select a client' }
      );

      if (!selected) return;
      clientId = selected.id;
    }

    // Get client details
    const result = await this.cli.listClients();
    if (!result.success || !result.stdout) {
      vscode.window.showErrorMessage('Failed to fetch client data');
      return;
    }

    const clients = this.parseClientList(result.stdout);
    const client = clients.find((c) => c.id === clientId);
    if (!client) {
      vscode.window.showErrorMessage('Client not found');
      return;
    }

    // Show action menu
    const actions = [
      { label: '$(edit) Edit Client', action: 'edit' },
      { label: '$(file-pdf) View Contracts', action: 'contracts' },
      { label: '$(file-text) View Invoices', action: 'invoices' },
      { label: '$(trash) Delete Client', action: 'delete' },
      { label: '$(close) Close', action: 'close' },
    ];

    const selected = await vscode.window.showQuickPick(actions, {
      placeHolder: `${client.name} • ${client.email}`,
      title: 'Client Actions',
    });

    if (selected) {
      switch (selected.action) {
        case 'edit':
          await this.editClient(clientId);
          break;
        case 'contracts':
          // Open contracts filtered by client (show info message for now)
          vscode.window.showInformationMessage(
            `Contracts for ${client.name} - Use the Contracts panel to view.`
          );
          break;
        case 'invoices':
          // Open invoices filtered by client (show info message for now)
          vscode.window.showInformationMessage(
            `Invoices for ${client.name} - Use the Invoices panel to view.`
          );
          break;
        case 'delete':
          await this.deleteClient(clientId);
          break;
      }
    }
  }

  /**
   * List clients with quick pick
   */
  async listClients(): Promise<void> {
    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Fetching clients...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.listClients();

        if (result.success && result.stdout) {
          // Show output in a new document
          const doc = await vscode.workspace.openTextDocument({
            content: result.stdout,
            language: 'plaintext',
          });
          await vscode.window.showTextDocument(doc);
        } else {
          vscode.window.showErrorMessage(
            `Failed to list clients: ${result.error}`
          );
        }
      }
    );
  }
}
