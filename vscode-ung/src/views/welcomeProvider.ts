import * as vscode from 'vscode';

/**
 * Welcome item types
 */
type WelcomeItemType = 'header' | 'feature' | 'action' | 'info' | 'separator';

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
      .TreeItemCollapsibleState.None
  ) {
    super(label, collapsibleState);

    if (icon) {
      this.iconPath = new vscode.ThemeIcon(icon);
    }

    if (command) {
      this.command = command;
    }

    this.contextValue = itemType;
  }
}

/**
 * Welcome provider when UNG CLI is not installed
 * Shows installation options and feature overview
 */
export class WelcomeProvider implements vscode.TreeDataProvider<WelcomeItem> {
  private _onDidChangeTreeData: vscode.EventEmitter<
    WelcomeItem | undefined | null | undefined
  > = new vscode.EventEmitter<WelcomeItem | undefined | null | undefined>();
  readonly onDidChangeTreeData: vscode.Event<
    WelcomeItem | undefined | null | undefined
  > = this._onDidChangeTreeData.event;

  refresh(): void {
    this._onDidChangeTreeData.fire(undefined);
  }

  getTreeItem(element: WelcomeItem): vscode.TreeItem {
    return element;
  }

  async getChildren(element?: WelcomeItem): Promise<WelcomeItem[]> {
    if (element) {
      // Children based on parent type
      if (element.label === 'Installation Options') {
        return this.getInstallationOptions();
      }
      if (element.label === 'Features') {
        return this.getFeaturesList();
      }
      if (element.label === 'Quick Links') {
        return this.getQuickLinks();
      }
      return [];
    }

    // Root level items
    return [
      new WelcomeItem(
        'UNG CLI Required',
        'header',
        'warning',
        undefined,
        vscode.TreeItemCollapsibleState.None
      ),
      new WelcomeItem(
        'Installation Options',
        'header',
        'cloud-download',
        undefined,
        vscode.TreeItemCollapsibleState.Expanded
      ),
      new WelcomeItem(
        'Features',
        'header',
        'star',
        undefined,
        vscode.TreeItemCollapsibleState.Expanded
      ),
      new WelcomeItem(
        'Quick Links',
        'header',
        'link-external',
        undefined,
        vscode.TreeItemCollapsibleState.Expanded
      ),
    ];
  }

  private getInstallationOptions(): WelcomeItem[] {
    const platform = process.platform;
    const items: WelcomeItem[] = [];

    // Homebrew (macOS/Linux)
    if (platform === 'darwin' || platform === 'linux') {
      items.push(
        new WelcomeItem(
          'Install via Homebrew (Recommended)',
          'action',
          'terminal',
          {
            command: 'ung.installViaHomebrew',
            title: 'Install via Homebrew',
          }
        )
      );
    }

    // Platform-specific installers
    if (platform === 'darwin') {
      items.push(
        new WelcomeItem('Download macOS Binary', 'action', 'desktop-download', {
          command: 'ung.downloadBinary',
          title: 'Download Binary',
          arguments: ['darwin'],
        })
      );
    } else if (platform === 'win32') {
      items.push(
        new WelcomeItem(
          'Download Windows Installer',
          'action',
          'desktop-download',
          {
            command: 'ung.downloadBinary',
            title: 'Download Binary',
            arguments: ['windows'],
          }
        )
      );
      items.push(
        new WelcomeItem('Install via Scoop', 'action', 'terminal', {
          command: 'ung.installViaScoop',
          title: 'Install via Scoop',
        })
      );
    } else {
      items.push(
        new WelcomeItem('Download Linux Binary', 'action', 'desktop-download', {
          command: 'ung.downloadBinary',
          title: 'Download Binary',
          arguments: ['linux'],
        })
      );
    }

    // Universal options
    items.push(
      new WelcomeItem('Install via Go', 'action', 'package', {
        command: 'ung.installViaGo',
        title: 'Install via Go',
      })
    );

    items.push(
      new WelcomeItem('View All Releases', 'action', 'github', {
        command: 'ung.openReleases',
        title: 'Open GitHub Releases',
      })
    );

    return items;
  }

  private getFeaturesList(): WelcomeItem[] {
    return [
      new WelcomeItem('Invoice Generation & PDF Export', 'feature', 'file-pdf'),
      new WelcomeItem('Time Tracking with Timer', 'feature', 'clock'),
      new WelcomeItem(
        'Client & Contract Management',
        'feature',
        'organization'
      ),
      new WelcomeItem('Expense Tracking', 'feature', 'credit-card'),
      new WelcomeItem('Revenue Dashboard', 'feature', 'graph'),
      new WelcomeItem('Multi-Currency Support', 'feature', 'globe'),
      new WelcomeItem('Email Integration', 'feature', 'mail'),
      new WelcomeItem('Tax & VAT Calculations', 'feature', 'law'),
    ];
  }

  private getQuickLinks(): WelcomeItem[] {
    return [
      new WelcomeItem('Documentation', 'action', 'book', {
        command: 'ung.openDocs',
        title: 'Open Documentation',
      }),
      new WelcomeItem('GitHub Repository', 'action', 'github', {
        command: 'ung.openGitHub',
        title: 'Open GitHub',
      }),
      new WelcomeItem('Report an Issue', 'action', 'bug', {
        command: 'ung.reportIssue',
        title: 'Report Issue',
      }),
      new WelcomeItem('Check CLI Installation', 'action', 'refresh', {
        command: 'ung.recheckCli',
        title: 'Recheck CLI',
      }),
    ];
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
