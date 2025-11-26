import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';

/**
 * Client tree item types
 */
type ClientItemType = 'summary' | 'action' | 'client' | 'header';

/**
 * Client tree item
 */
export class ClientItem extends vscode.TreeItem {
  constructor(
    public readonly itemType: ClientItemType,
    public readonly itemId: number,
    public readonly name: string,
    public readonly email: string,
    public readonly collapsibleState: vscode.TreeItemCollapsibleState,
    public readonly contractCount?: number,
    public readonly totalRevenue?: number
  ) {
    super(name, collapsibleState);

    this.id = `${itemType}-${itemId}`;

    switch (itemType) {
      case 'summary':
        this.contextValue = 'summary';
        break;
      case 'action':
        this.contextValue = 'action';
        break;
      case 'header':
        this.contextValue = 'header';
        break;
      case 'client':
        this.tooltip = this.buildTooltip();
        this.description = email || '';
        this.contextValue = 'client';
        this.iconPath = new vscode.ThemeIcon(
          'person',
          new vscode.ThemeColor('charts.blue')
        );
        // Add click command to view client details
        this.command = {
          command: 'ung.viewClient',
          title: 'View Client',
          arguments: [itemId],
        };
        break;
    }
  }

  private buildTooltip(): string {
    let tooltip = `**${this.name}**\n\n`;
    if (this.email) tooltip += `Email: ${this.email}\n`;
    if (this.contractCount !== undefined)
      tooltip += `Contracts: ${this.contractCount}\n`;
    if (this.totalRevenue !== undefined)
      tooltip += `Total Revenue: $${this.totalRevenue.toFixed(2)}`;
    return tooltip;
  }
}

/**
 * Client tree data provider with summary metrics
 */
export class ClientProvider implements vscode.TreeDataProvider<ClientItem> {
  private _onDidChangeTreeData: vscode.EventEmitter<
    ClientItem | undefined | null | undefined
  > = new vscode.EventEmitter<ClientItem | undefined | null | undefined>();
  readonly onDidChangeTreeData: vscode.Event<
    ClientItem | undefined | null | undefined
  > = this._onDidChangeTreeData.event;

  constructor(private cli: UngCli) {}

  refresh(): void {
    this._onDidChangeTreeData.fire(undefined);
  }

  getTreeItem(element: ClientItem): vscode.TreeItem {
    return element;
  }

  async getChildren(element?: ClientItem): Promise<ClientItem[]> {
    if (element) {
      return [];
    }

    try {
      const result = await this.cli.listClients();

      if (!result.success || !result.stdout) {
        return this.getEmptyState();
      }

      // Parse the CLI output
      const clients = this.parseClientOutput(result.stdout);

      if (clients.length === 0) {
        return this.getEmptyState();
      }

      return this.buildTreeItems(clients);
    } catch (error) {
      vscode.window.showErrorMessage(`Failed to load clients: ${error}`);
      return this.getEmptyState();
    }
  }

  /**
   * Build tree items with summary at top
   */
  private buildTreeItems(clients: ClientItem[]): ClientItem[] {
    const items: ClientItem[] = [];

    // Summary section
    const summaryItem = new ClientItem(
      'summary',
      0,
      `${clients.length} Client${clients.length !== 1 ? 's' : ''}`,
      '',
      vscode.TreeItemCollapsibleState.None
    );
    summaryItem.iconPath = new vscode.ThemeIcon(
      'organization',
      new vscode.ThemeColor('charts.green')
    );
    summaryItem.description = 'Total';
    items.push(summaryItem);

    // Quick actions
    const addClientItem = new ClientItem(
      'action',
      -1,
      'Add New Client',
      '',
      vscode.TreeItemCollapsibleState.None
    );
    addClientItem.iconPath = new vscode.ThemeIcon(
      'person-add',
      new vscode.ThemeColor('charts.blue')
    );
    addClientItem.command = {
      command: 'ung.createClient',
      title: 'Add Client',
    };
    items.push(addClientItem);

    // Separator-like header with better formatting
    const clientsHeader = new ClientItem(
      'header',
      -2,
      '── Clients ──',
      '',
      vscode.TreeItemCollapsibleState.None
    );
    clientsHeader.iconPath = new vscode.ThemeIcon(
      'account',
      new vscode.ThemeColor('charts.gray')
    );
    items.push(clientsHeader);

    // Add all clients
    items.push(...clients);

    return items;
  }

  /**
   * Empty state with helpful message
   */
  private getEmptyState(): ClientItem[] {
    const emptyItem = new ClientItem(
      'summary',
      0,
      'No clients yet',
      '',
      vscode.TreeItemCollapsibleState.None
    );
    emptyItem.iconPath = new vscode.ThemeIcon('info');
    emptyItem.description = 'Add your first client';

    const addItem = new ClientItem(
      'action',
      -1,
      'Add Your First Client',
      '',
      vscode.TreeItemCollapsibleState.None
    );
    addItem.iconPath = new vscode.ThemeIcon(
      'person-add',
      new vscode.ThemeColor('charts.green')
    );
    addItem.command = {
      command: 'ung.createClient',
      title: 'Add Client',
    };

    return [emptyItem, addItem];
  }

  /**
   * Parse client list output from CLI
   */
  private parseClientOutput(output: string): ClientItem[] {
    const lines = output.split('\n').filter((line) => line.trim());
    const clients: ClientItem[] = [];

    for (let i = 1; i < lines.length; i++) {
      // Skip header
      const line = lines[i].trim();
      if (!line) continue;

      const parts = line.split(/\s{2,}/); // Split by multiple spaces
      if (parts.length >= 3) {
        const id = parseInt(parts[0], 10);
        const name = parts[1];
        const email = parts[2];

        if (!Number.isNaN(id)) {
          clients.push(
            new ClientItem(
              'client',
              id,
              name,
              email,
              vscode.TreeItemCollapsibleState.None
            )
          );
        }
      }
    }

    return clients;
  }
}
