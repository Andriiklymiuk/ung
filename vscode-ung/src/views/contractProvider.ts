import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';
import { Formatter } from '../utils/formatting';

type ContractTreeItem =
  | ContractSectionItem
  | ContractSummaryItem
  | ContractItem
  | ContractActionItem;

/**
 * Contract section for grouping by type
 */
class ContractSectionItem extends vscode.TreeItem {
  constructor(
    public readonly label: string,
    public readonly section: string,
    public readonly icon: string,
    public readonly colorId: string,
    public readonly count: number,
    public readonly totalValue?: string
  ) {
    super(
      label,
      count > 0
        ? vscode.TreeItemCollapsibleState.Expanded
        : vscode.TreeItemCollapsibleState.Collapsed
    );
    this.description = totalValue ? `${count} • ${totalValue}` : `${count}`;
    this.iconPath = new vscode.ThemeIcon(icon, new vscode.ThemeColor(colorId));
    this.contextValue = `contract_section_${section}`;
  }
}

/**
 * Contract summary item
 */
class ContractSummaryItem extends vscode.TreeItem {
  constructor(label: string, value: string, icon: string, colorId?: string) {
    super(label, vscode.TreeItemCollapsibleState.None);
    this.description = value;
    this.iconPath = colorId
      ? new vscode.ThemeIcon(icon, new vscode.ThemeColor(colorId))
      : new vscode.ThemeIcon(icon);
    this.contextValue = 'contract_summary';
  }
}

/**
 * Contract action item
 */
class ContractActionItem extends vscode.TreeItem {
  constructor(
    label: string,
    icon: string,
    commandId: string,
    colorId?: string
  ) {
    super(label, vscode.TreeItemCollapsibleState.None);
    this.iconPath = colorId
      ? new vscode.ThemeIcon(icon, new vscode.ThemeColor(colorId))
      : new vscode.ThemeIcon(icon);
    this.command = {
      command: commandId,
      title: label,
    };
    this.contextValue = 'contract_action';
  }
}

/**
 * Contract tree item
 */
export class ContractItem extends vscode.TreeItem {
  constructor(
    public readonly itemId: number,
    public readonly name: string,
    public readonly clientName: string,
    public readonly type: string,
    public readonly rate: string,
    public readonly rateValue: number,
    public readonly currency: string,
    public readonly active: boolean,
    public readonly collapsibleState: vscode.TreeItemCollapsibleState
  ) {
    super(name, collapsibleState);

    this.id = `contract_${itemId}`;

    // Rich tooltip with markdown
    this.tooltip = new vscode.MarkdownString();
    this.tooltip.appendMarkdown(`**${name}**\n\n`);
    this.tooltip.appendMarkdown(`- Client: ${clientName}\n`);
    this.tooltip.appendMarkdown(`- Type: ${this.formatType(type)}\n`);
    this.tooltip.appendMarkdown(`- Rate: ${rate}\n`);
    this.tooltip.appendMarkdown(
      `- Status: ${active ? '🟢 Active' : '⚪ Inactive'}\n`
    );
    this.tooltip.appendMarkdown(`\n*Click to view details*`);

    this.description = `${clientName} • ${rate}`;
    this.contextValue = 'contract';

    // Set icon based on type and status
    this.iconPath = this.getContractIcon(type, active);

    // Add view command on click
    this.command = {
      command: 'ung.viewContract',
      title: 'View Contract',
      arguments: [{ itemId }],
    };
  }

  private formatType(type: string): string {
    const typeMap: { [key: string]: string } = {
      hourly: 'Hourly',
      retainer: 'Retainer',
      fixed: 'Fixed Price',
      fixed_price: 'Fixed Price',
      milestone: 'Milestone',
    };
    return typeMap[type.toLowerCase()] || type;
  }

  private getContractIcon(type: string, active: boolean): vscode.ThemeIcon {
    if (!active) {
      return new vscode.ThemeIcon(
        'circle-slash',
        new vscode.ThemeColor('charts.gray')
      );
    }

    const typeLower = type.toLowerCase();
    switch (typeLower) {
      case 'hourly':
        return new vscode.ThemeIcon(
          'clock',
          new vscode.ThemeColor('charts.blue')
        );
      case 'retainer':
        return new vscode.ThemeIcon(
          'calendar',
          new vscode.ThemeColor('charts.purple')
        );
      case 'fixed':
      case 'fixed_price':
        return new vscode.ThemeIcon(
          'package',
          new vscode.ThemeColor('charts.green')
        );
      case 'milestone':
        return new vscode.ThemeIcon(
          'milestone',
          new vscode.ThemeColor('charts.orange')
        );
      default:
        return new vscode.ThemeIcon(
          'file-code',
          new vscode.ThemeColor('charts.blue')
        );
    }
  }
}

interface ParsedContract {
  id: number;
  name: string;
  clientName: string;
  type: string;
  rate: string;
  rateValue: number;
  currency: string;
  active: boolean;
}

/**
 * Contract tree data provider with type grouping
 */
export class ContractProvider
  implements vscode.TreeDataProvider<ContractTreeItem>
{
  private _onDidChangeTreeData: vscode.EventEmitter<
    ContractTreeItem | undefined | null | undefined
  > = new vscode.EventEmitter<
    ContractTreeItem | undefined | null | undefined
  >();
  readonly onDidChangeTreeData: vscode.Event<
    ContractTreeItem | undefined | null | undefined
  > = this._onDidChangeTreeData.event;

  private cachedContracts: ParsedContract[] = [];

  constructor(private cli: UngCli) {}

  refresh(): void {
    this.cachedContracts = [];
    this._onDidChangeTreeData.fire(undefined);
  }

  getTreeItem(element: ContractTreeItem): vscode.TreeItem {
    return element;
  }

  async getChildren(element?: ContractTreeItem): Promise<ContractTreeItem[]> {
    if (!element) {
      return this.getRootItems();
    }

    if (element instanceof ContractSectionItem) {
      return this.getContractsForSection(element.section);
    }

    return [];
  }

  private async getRootItems(): Promise<ContractTreeItem[]> {
    await this.loadContracts();
    const items: ContractTreeItem[] = [];

    const activeContracts = this.cachedContracts.filter((c) => c.active);
    const inactiveContracts = this.cachedContracts.filter((c) => !c.active);

    // Summary stats
    if (activeContracts.length > 0) {
      // Calculate monthly value (rough estimate)
      const monthlyValue = activeContracts.reduce((sum, c) => {
        if (c.type.toLowerCase() === 'hourly') {
          return sum + c.rateValue * 160; // Assume 160 hours/month
        } else if (c.type.toLowerCase() === 'retainer') {
          return sum + c.rateValue;
        }
        return sum + c.rateValue;
      }, 0);

      items.push(
        new ContractSummaryItem(
          'Active Contracts',
          String(activeContracts.length),
          'briefcase',
          'charts.green'
        )
      );

      items.push(
        new ContractSummaryItem(
          'Est. Monthly Value',
          Formatter.formatCurrency(monthlyValue, 'USD'),
          'graph-line',
          'charts.blue'
        )
      );
    }

    // Quick actions
    items.push(
      new ContractActionItem(
        'Create New Contract',
        'add',
        'ung.createContract',
        'charts.blue'
      )
    );

    items.push(
      new ContractActionItem(
        'Add New Client',
        'person-add',
        'ung.createClient',
        'charts.purple'
      )
    );

    // Type sections for active contracts
    const typeGroups = [
      {
        key: 'hourly',
        label: 'Hourly Contracts',
        icon: 'clock',
        color: 'charts.blue',
      },
      {
        key: 'retainer',
        label: 'Retainers',
        icon: 'calendar',
        color: 'charts.purple',
      },
      {
        key: 'fixed',
        label: 'Fixed Price',
        icon: 'package',
        color: 'charts.green',
      },
      {
        key: 'milestone',
        label: 'Milestone',
        icon: 'milestone',
        color: 'charts.orange',
      },
    ];

    for (const group of typeGroups) {
      const contracts = activeContracts.filter((c) => {
        const typeLower = c.type.toLowerCase();
        return (
          typeLower === group.key ||
          (group.key === 'fixed' && typeLower === 'fixed_price')
        );
      });

      if (contracts.length > 0) {
        const totalValue = contracts.reduce((sum, c) => sum + c.rateValue, 0);
        items.push(
          new ContractSectionItem(
            group.label,
            group.key,
            group.icon,
            group.color,
            contracts.length,
            group.key === 'hourly'
              ? `${Formatter.formatCurrency(totalValue, 'USD')}/hr avg`
              : Formatter.formatCurrency(totalValue, 'USD')
          )
        );
      }
    }

    // Inactive contracts section
    if (inactiveContracts.length > 0) {
      items.push(
        new ContractSectionItem(
          'Inactive',
          'inactive',
          'circle-slash',
          'charts.gray',
          inactiveContracts.length
        )
      );
    }

    return items;
  }

  private async getContractsForSection(
    section: string
  ): Promise<ContractItem[]> {
    let filtered: ParsedContract[];

    if (section === 'inactive') {
      filtered = this.cachedContracts.filter((c) => !c.active);
    } else {
      filtered = this.cachedContracts.filter((c) => {
        if (!c.active) return false;
        const typeLower = c.type.toLowerCase();
        if (section === 'fixed') {
          return typeLower === 'fixed' || typeLower === 'fixed_price';
        }
        return typeLower === section;
      });
    }

    return filtered
      .sort((a, b) => b.rateValue - a.rateValue)
      .map(
        (c) =>
          new ContractItem(
            c.id,
            c.name,
            c.clientName,
            c.type,
            c.rate,
            c.rateValue,
            c.currency,
            c.active,
            vscode.TreeItemCollapsibleState.None
          )
      );
  }

  private async loadContracts(): Promise<void> {
    if (this.cachedContracts.length > 0) {
      return;
    }

    try {
      const result = await this.cli.listContracts();

      if (!result.success || !result.stdout) {
        this.cachedContracts = [];
        return;
      }

      this.cachedContracts = this.parseContractOutput(result.stdout);
    } catch (error) {
      vscode.window.showErrorMessage(`Failed to load contracts: ${error}`);
      this.cachedContracts = [];
    }
  }

  /**
   * Parse contract list output from CLI
   */
  private parseContractOutput(output: string): ParsedContract[] {
    const lines = output.split('\n').filter((line) => line.trim());
    const contracts: ParsedContract[] = [];

    for (let i = 1; i < lines.length; i++) {
      // Skip header
      const line = lines[i].trim();
      if (!line) continue;

      const parts = line.split(/\s{2,}/); // Split by multiple spaces
      if (parts.length >= 5) {
        const id = parseInt(parts[0], 10);
        const name = parts[1];
        const clientName = parts[2];
        const type = parts[3];
        const rate = parts[4];
        const active =
          parts[5]?.toLowerCase() === 'true' ||
          parts[5]?.toLowerCase() === 'active' ||
          parts[5]?.toLowerCase() === 'yes';

        // Parse rate value
        const rateMatch = rate.match(/[\d,]+\.?\d*/);
        const rateValue = rateMatch
          ? parseFloat(rateMatch[0].replace(/,/g, ''))
          : 0;

        // Parse currency from rate
        const currencyMatch = rate.match(/([A-Z]{3})/);
        const currency = currencyMatch ? currencyMatch[1] : 'USD';

        if (!Number.isNaN(id)) {
          contracts.push({
            id,
            name,
            clientName,
            type,
            rate,
            rateValue,
            currency,
            active,
          });
        }
      }
    }

    return contracts;
  }
}
