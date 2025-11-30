import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';
import { GIG_STATUSES } from '../commands/gig';

type GigTreeItem = GigSectionItem | GigItem | GigActionItem | GigSummaryItem;

/**
 * Gig status section
 */
class GigSectionItem extends vscode.TreeItem {
  constructor(
    public readonly label: string,
    public readonly section: string,
    public readonly icon: string,
    public readonly colorId: string,
    public readonly count: number
  ) {
    super(
      label,
      count > 0
        ? vscode.TreeItemCollapsibleState.Expanded
        : vscode.TreeItemCollapsibleState.Collapsed
    );
    this.description = `${count}`;
    this.iconPath = new vscode.ThemeIcon(icon, new vscode.ThemeColor(colorId));
    this.contextValue = `gig_section_${section}`;
  }
}

/**
 * Summary item for gig stats
 */
class GigSummaryItem extends vscode.TreeItem {
  constructor(label: string, value: string, icon: string, colorId?: string) {
    super(label, vscode.TreeItemCollapsibleState.None);
    this.description = value;
    this.iconPath = colorId
      ? new vscode.ThemeIcon(icon, new vscode.ThemeColor(colorId))
      : new vscode.ThemeIcon(icon);
    this.contextValue = 'gig_summary';
  }
}

/**
 * Quick action item
 */
class GigActionItem extends vscode.TreeItem {
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
    this.contextValue = 'gig_action';
  }
}

/**
 * Individual gig item
 */
export class GigItem extends vscode.TreeItem {
  constructor(
    public readonly itemId: number,
    public readonly gigName: string,
    public readonly clientName: string,
    public readonly status: string,
    public readonly hours: number,
    public readonly gigType: string
  ) {
    super(gigName, vscode.TreeItemCollapsibleState.None);

    this.id = `gig_${itemId}`;
    this.tooltip = new vscode.MarkdownString();
    this.tooltip.appendMarkdown(`**${gigName}**\n\n`);
    this.tooltip.appendMarkdown(`- Client: ${clientName || 'None'}\n`);
    this.tooltip.appendMarkdown(`- Status: ${status}\n`);
    this.tooltip.appendMarkdown(`- Hours: ${hours.toFixed(1)}\n`);
    this.tooltip.appendMarkdown(`- Type: ${gigType}\n`);
    this.tooltip.appendMarkdown(`\n*Click to view actions*`);

    this.description = `${clientName || '-'} • ${hours.toFixed(1)}h`;
    this.contextValue = 'gig';

    this.iconPath = this.getStatusIcon(status);

    this.command = {
      command: 'ung.viewGig',
      title: 'View Gig',
      arguments: [{ itemId }],
    };
  }

  private getStatusIcon(status: string): vscode.ThemeIcon {
    const statusConfig = GIG_STATUSES.find(
      (s) => s.value === status.toLowerCase().replace(' ', '_')
    );
    if (statusConfig) {
      return new vscode.ThemeIcon(
        statusConfig.icon,
        new vscode.ThemeColor(statusConfig.color)
      );
    }
    return new vscode.ThemeIcon('circle-outline');
  }
}

interface ParsedGig {
  id: number;
  name: string;
  client: string;
  status: string;
  hours: number;
  type: string;
}

/**
 * Tree data provider for gigs grouped by status
 */
export class GigProvider implements vscode.TreeDataProvider<GigTreeItem> {
  private _onDidChangeTreeData: vscode.EventEmitter<
    GigTreeItem | undefined | null
  > = new vscode.EventEmitter<GigTreeItem | undefined | null>();
  readonly onDidChangeTreeData: vscode.Event<GigTreeItem | undefined | null> =
    this._onDidChangeTreeData.event;

  private cachedGigs: ParsedGig[] = [];

  constructor(private cli: UngCli) {}

  refresh(): void {
    this.cachedGigs = [];
    this._onDidChangeTreeData.fire(undefined);
  }

  getTreeItem(element: GigTreeItem): vscode.TreeItem {
    return element;
  }

  async getChildren(element?: GigTreeItem): Promise<GigTreeItem[]> {
    if (!element) {
      return this.getRootItems();
    }

    if (element instanceof GigSectionItem) {
      return this.getGigsForSection(element.section);
    }

    return [];
  }

  private async getRootItems(): Promise<GigTreeItem[]> {
    await this.loadGigs();
    const items: GigTreeItem[] = [];

    // Summary stats
    const totalGigs = this.cachedGigs.length;
    const activeGigs = this.cachedGigs.filter(
      (g) => g.status === 'active'
    ).length;
    const totalHours = this.cachedGigs.reduce((sum, g) => sum + g.hours, 0);

    if (totalGigs > 0) {
      items.push(
        new GigSummaryItem(
          'Total Gigs',
          totalGigs.toString(),
          'briefcase',
          'charts.blue'
        )
      );

      if (activeGigs > 0) {
        items.push(
          new GigSummaryItem(
            'Active',
            activeGigs.toString(),
            'play-circle',
            'charts.green'
          )
        );
      }

      if (totalHours > 0) {
        items.push(
          new GigSummaryItem(
            'Hours Tracked',
            `${totalHours.toFixed(1)}h`,
            'clock',
            'charts.purple'
          )
        );
      }
    }

    // Quick actions
    items.push(
      new GigActionItem('Create New Gig', 'add', 'ung.createGig', 'charts.blue')
    );

    items.push(
      new GigActionItem(
        'Open Kanban Board',
        'layout',
        'ung.openGigBoard',
        'charts.purple'
      )
    );

    // Status sections - only show active workflow statuses
    const workflowStatuses = [
      'pipeline',
      'negotiating',
      'active',
      'delivered',
      'invoiced',
      'complete',
    ];

    for (const statusValue of workflowStatuses) {
      const statusConfig = GIG_STATUSES.find((s) => s.value === statusValue);
      if (!statusConfig) continue;

      const count = this.cachedGigs.filter(
        (g) => g.status.toLowerCase().replace(' ', '_') === statusValue
      ).length;

      // Show all workflow statuses for kanban-like view
      items.push(
        new GigSectionItem(
          statusConfig.label,
          statusValue,
          statusConfig.icon,
          statusConfig.color,
          count
        )
      );
    }

    return items;
  }

  private async getGigsForSection(section: string): Promise<GigItem[]> {
    return this.cachedGigs
      .filter((g) => g.status.toLowerCase().replace(' ', '_') === section)
      .map(
        (g) => new GigItem(g.id, g.name, g.client, g.status, g.hours, g.type)
      );
  }

  private async loadGigs(): Promise<void> {
    if (this.cachedGigs.length > 0) {
      return;
    }

    try {
      const result = await this.cli.listGigs();

      if (!result.success || !result.stdout) {
        this.cachedGigs = [];
        return;
      }

      this.cachedGigs = this.parseGigOutput(result.stdout);
    } catch (error) {
      vscode.window.showErrorMessage(`Failed to load gigs: ${error}`);
      this.cachedGigs = [];
    }
  }

  private parseGigOutput(output: string): ParsedGig[] {
    const lines = output.trim().split('\n');
    if (lines.length < 2) return [];

    const gigs: ParsedGig[] = [];

    for (let i = 1; i < lines.length; i++) {
      const line = lines[i].trim();
      if (!line || line.startsWith('--')) continue;

      const parts = line.split(/\s{2,}/).filter((p) => p.trim());
      if (parts.length >= 5) {
        const id = parseInt(parts[0], 10);
        if (!Number.isNaN(id)) {
          gigs.push({
            id,
            name: parts[1] || '',
            client: parts[2] || '-',
            status: parts[3] || 'pipeline',
            hours: parseFloat(parts[4]) || 0,
            type: parts[5] || 'hourly',
          });
        }
      }
    }

    return gigs;
  }
}
