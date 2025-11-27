import * as vscode from 'vscode';

/**
 * Welcome item types
 */
type WelcomeItemType =
  | 'header'
  | 'feature'
  | 'action'
  | 'info'
  | 'separator'
  | 'step'
  | 'benefit';

/**
 * Welcome tree item with rich formatting
 */
export class WelcomeItem extends vscode.TreeItem {
  constructor(
    label: string,
    public readonly itemType: WelcomeItemType,
    public readonly icon?: string,
    public readonly command?: vscode.Command,
    collapsibleState: vscode.TreeItemCollapsibleState = vscode
      .TreeItemCollapsibleState.None,
    description?: string,
    tooltip?: string
  ) {
    super(label, collapsibleState);

    if (icon) {
      this.iconPath = new vscode.ThemeIcon(icon);
    }

    if (command) {
      this.command = command;
    }

    if (description) {
      this.description = description;
    }

    if (tooltip) {
      this.tooltip = tooltip;
    }

    this.contextValue = itemType;
  }
}

/**
 * Getting Started provider for new users who have CLI installed
 * Shows onboarding steps and quick setup
 */
export class GettingStartedProvider
  implements vscode.TreeDataProvider<WelcomeItem>
{
  private _onDidChangeTreeData: vscode.EventEmitter<
    WelcomeItem | undefined | null | undefined
  > = new vscode.EventEmitter<WelcomeItem | undefined | null | undefined>();
  readonly onDidChangeTreeData: vscode.Event<
    WelcomeItem | undefined | null | undefined
  > = this._onDidChangeTreeData.event;

  private hasCompany: boolean = false;
  private hasClient: boolean = false;
  private hasContract: boolean = false;

  constructor(
    private checkCompany: () => Promise<boolean>,
    private checkClient: () => Promise<boolean>,
    private checkContract: () => Promise<boolean>
  ) {}

  async refresh(): Promise<void> {
    this.hasCompany = await this.checkCompany();
    this.hasClient = await this.checkClient();
    this.hasContract = await this.checkContract();
    this._onDidChangeTreeData.fire(undefined);
  }

  getTreeItem(element: WelcomeItem): vscode.TreeItem {
    return element;
  }

  async getChildren(element?: WelcomeItem): Promise<WelcomeItem[]> {
    if (element) {
      return [];
    }

    const items: WelcomeItem[] = [];

    // Step 1: Create Company
    if (!this.hasCompany) {
      items.push(
        new WelcomeItem(
          '1. Set Up Your Company',
          'action',
          'circle-large-outline',
          {
            command: 'ung.createCompany',
            title: 'Create Company',
          }
        )
      );
    } else {
      const item = new WelcomeItem('1. Company Created', 'info', 'pass');
      item.description = 'Done';
      items.push(item);
    }

    // Step 2: Add Client
    if (!this.hasClient) {
      items.push(
        new WelcomeItem(
          '2. Add Your First Client',
          'action',
          this.hasCompany ? 'circle-large-outline' : 'circle-slash',
          this.hasCompany
            ? {
                command: 'ung.createClient',
                title: 'Create Client',
              }
            : undefined
        )
      );
    } else {
      const item = new WelcomeItem('2. Client Added', 'info', 'pass');
      item.description = 'Done';
      items.push(item);
    }

    // Step 3: Create Contract
    if (!this.hasContract) {
      items.push(
        new WelcomeItem(
          '3. Create a Contract',
          'action',
          this.hasClient ? 'circle-large-outline' : 'circle-slash',
          this.hasClient
            ? {
                command: 'ung.createContract',
                title: 'Create Contract',
              }
            : undefined
        )
      );
    } else {
      const item = new WelcomeItem('3. Contract Created', 'info', 'pass');
      item.description = 'Done';
      items.push(item);
    }

    // Quick Actions (always show)
    items.push(new WelcomeItem('', 'separator', undefined));

    items.push(
      new WelcomeItem('Start Time Tracking', 'action', 'play', {
        command: 'ung.startTracking',
        title: 'Start Tracking',
      })
    );

    items.push(
      new WelcomeItem('Create Invoice', 'action', 'new-file', {
        command: 'ung.createInvoice',
        title: 'Create Invoice',
      })
    );

    items.push(
      new WelcomeItem('View Documentation', 'action', 'book', {
        command: 'ung.openDocs',
        title: 'Open Docs',
      })
    );

    return items;
  }
}
