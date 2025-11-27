import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';
import { Formatter } from '../utils/formatting';

/**
 * Dashboard metric item
 */
class DashboardMetricItem extends vscode.TreeItem {
  constructor(
    public readonly label: string,
    public readonly value: string,
    public readonly icon: string,
    public readonly colorId?: string
  ) {
    super(label, vscode.TreeItemCollapsibleState.None);
    this.description = value;
    this.tooltip = `${label}: ${value}`;
    this.iconPath = colorId
      ? new vscode.ThemeIcon(icon, new vscode.ThemeColor(colorId))
      : new vscode.ThemeIcon(icon);
  }
}

/**
 * Dashboard action item for quick actions
 */
class DashboardActionItem extends vscode.TreeItem {
  constructor(
    public readonly label: string,
    public readonly icon: string,
    public readonly commandId: string,
    public readonly colorId?: string
  ) {
    super(label, vscode.TreeItemCollapsibleState.None);
    this.iconPath = colorId
      ? new vscode.ThemeIcon(icon, new vscode.ThemeColor(colorId))
      : new vscode.ThemeIcon(icon);
    this.command = {
      command: commandId,
      title: label,
    };
    this.contextValue = 'action';
  }
}

/**
 * Dashboard section item
 */
class DashboardSectionItem extends vscode.TreeItem {
  constructor(
    public readonly label: string,
    public readonly section: string,
    public readonly icon: string = 'folder',
    collapsibleState: vscode.TreeItemCollapsibleState = vscode
      .TreeItemCollapsibleState.Expanded
  ) {
    super(label, collapsibleState);
    this.contextValue = `dashboard_${section}`;
    this.iconPath = new vscode.ThemeIcon(icon);
  }
}

type DashboardItem =
  | DashboardSectionItem
  | DashboardMetricItem
  | DashboardActionItem;

/**
 * Dashboard metrics interface
 */
interface DashboardMetrics {
  totalMonthlyRevenue: number;
  hourlyRevenue: number;
  retainerRevenue: number;
  projectedHours: number;
  averageHourlyRate: number;
  activeContracts: number;
  totalClients: number;
  pendingInvoices: number;
  unpaidAmount: number;
  currency: string;
}

/**
 * Dashboard tree data provider
 */
export class DashboardProvider
  implements vscode.TreeDataProvider<DashboardItem>
{
  private _onDidChangeTreeData: vscode.EventEmitter<
    DashboardItem | undefined | null | undefined
  > = new vscode.EventEmitter<DashboardItem | undefined | null | undefined>();
  readonly onDidChangeTreeData: vscode.Event<
    DashboardItem | undefined | null | undefined
  > = this._onDidChangeTreeData.event;

  private cachedMetrics: DashboardMetrics | null = null;
  private cacheTimestamp: number = 0;
  private readonly CACHE_DURATION = 60000; // 1 minute
  private activeTracking: {
    project: string;
    client: string;
    duration: string;
  } | null = null;

  constructor(private cli: UngCli) {}

  refresh(): void {
    this.cachedMetrics = null;
    this._onDidChangeTreeData.fire(undefined);
  }

  getTreeItem(element: DashboardItem): vscode.TreeItem {
    return element;
  }

  async getChildren(element?: DashboardItem): Promise<DashboardItem[]> {
    if (!element) {
      // Root level - show dashboard sections
      const sections: DashboardItem[] = [];

      // Check for active tracking session
      await this.checkActiveTracking();
      if (this.activeTracking) {
        sections.push(
          new DashboardSectionItem(
            'Active Session',
            'active',
            'pulse',
            vscode.TreeItemCollapsibleState.Expanded
          )
        );
      }

      sections.push(
        new DashboardSectionItem(
          'Quick Actions',
          'actions',
          'zap',
          vscode.TreeItemCollapsibleState.Expanded
        ),
        new DashboardSectionItem('Revenue Overview', 'revenue', 'graph-line'),
        new DashboardSectionItem('Business Summary', 'stats', 'dashboard')
      );

      return sections;
    }

    if (element instanceof DashboardSectionItem) {
      switch (element.section) {
        case 'active':
          return this.getActiveTrackingItems();
        case 'actions':
          return this.getQuickActions();
        case 'revenue':
          return this.getRevenueMetrics();
        case 'stats':
          return this.getStatsMetrics();
        default:
          return [];
      }
    }

    return [];
  }

  private async checkActiveTracking(): Promise<void> {
    try {
      const result = await this.cli.exec(['tracking', 'status']);
      if (
        result.success &&
        result.stdout &&
        !result.stdout.includes('No active')
      ) {
        // Parse active session
        const lines = result.stdout.split('\n');
        let project = 'Unknown';
        let client = '';
        let duration = '0:00';

        for (const line of lines) {
          if (line.includes('Project:')) {
            project = line.split(':')[1]?.trim() || 'Unknown';
          } else if (line.includes('Client:')) {
            client = line.split(':')[1]?.trim() || '';
          } else if (line.includes('Duration:') || line.includes('Elapsed:')) {
            duration = line.split(':').slice(1).join(':').trim() || '0:00';
          }
        }

        this.activeTracking = { project, client, duration };
      } else {
        this.activeTracking = null;
      }
    } catch {
      this.activeTracking = null;
    }
  }

  private getActiveTrackingItems(): DashboardItem[] {
    if (!this.activeTracking) {
      return [];
    }

    const items: DashboardItem[] = [];

    items.push(
      new DashboardMetricItem(
        this.activeTracking.project,
        this.activeTracking.duration,
        'record',
        'charts.red'
      )
    );

    if (this.activeTracking.client) {
      items.push(
        new DashboardMetricItem(
          'Client',
          this.activeTracking.client,
          'person',
          'charts.blue'
        )
      );
    }

    items.push(
      new DashboardActionItem(
        'Stop Tracking',
        'debug-stop',
        'ung.stopTracking',
        'charts.red'
      )
    );

    return items;
  }

  private getQuickActions(): DashboardItem[] {
    const actions: DashboardItem[] = [];

    // Most common actions with intuitive icons
    if (!this.activeTracking) {
      actions.push(
        new DashboardActionItem(
          'Start Time Tracking',
          'play',
          'ung.startTracking',
          'charts.green'
        )
      );
    }

    actions.push(
      new DashboardActionItem(
        'Create Invoice',
        'new-file',
        'ung.createInvoice',
        'charts.blue'
      )
    );

    actions.push(
      new DashboardActionItem(
        'Log Expense',
        'credit-card',
        'ung.logExpense',
        'charts.orange'
      )
    );

    actions.push(
      new DashboardActionItem('Add Client', 'person-add', 'ung.createClient')
    );

    actions.push(
      new DashboardActionItem('New Contract', 'file-add', 'ung.createContract')
    );

    actions.push(
      new DashboardActionItem(
        'Edit PDF Template',
        'file-code',
        'ung.openTemplateEditor',
        'charts.yellow'
      )
    );

    return actions;
  }

  private async getRevenueMetrics(): Promise<DashboardMetricItem[]> {
    try {
      const metrics = await this.getDashboardMetrics();
      const items: DashboardMetricItem[] = [];

      items.push(
        new DashboardMetricItem(
          'Total Revenue',
          Formatter.formatCurrency(
            metrics.totalMonthlyRevenue || 0,
            metrics.currency || 'USD'
          ),
          'graph-line',
          'charts.green'
        )
      );

      items.push(
        new DashboardMetricItem(
          'From Hourly',
          Formatter.formatCurrency(
            metrics.hourlyRevenue || 0,
            metrics.currency || 'USD'
          ),
          'clock',
          'charts.blue'
        )
      );

      items.push(
        new DashboardMetricItem(
          'From Retainers',
          Formatter.formatCurrency(
            metrics.retainerRevenue || 0,
            metrics.currency || 'USD'
          ),
          'calendar',
          'charts.purple'
        )
      );

      items.push(
        new DashboardMetricItem(
          'Hours This Month',
          `${(metrics.projectedHours || 0).toFixed(1)}h`,
          'watch',
          'charts.yellow'
        )
      );

      if (metrics.averageHourlyRate > 0) {
        items.push(
          new DashboardMetricItem(
            'Avg. Hourly Rate',
            `$${metrics.averageHourlyRate.toFixed(0)}/hr`,
            'tag',
            'charts.orange'
          )
        );
      }

      return items;
    } catch (_error) {
      return [
        new DashboardMetricItem(
          'Error',
          'Failed to load revenue',
          'error',
          'errorForeground'
        ),
      ];
    }
  }

  private async getStatsMetrics(): Promise<DashboardMetricItem[]> {
    try {
      const metrics = await this.getDashboardMetrics();
      const items: DashboardMetricItem[] = [];

      items.push(
        new DashboardMetricItem(
          'Active Clients',
          String(metrics.totalClients || 0),
          'people',
          'charts.blue'
        )
      );

      items.push(
        new DashboardMetricItem(
          'Active Contracts',
          String(metrics.activeContracts || 0),
          'briefcase',
          'charts.purple'
        )
      );

      const pendingCount = metrics.pendingInvoices || 0;
      items.push(
        new DashboardMetricItem(
          'Pending Invoices',
          String(pendingCount),
          pendingCount > 0 ? 'bell' : 'bell-slash',
          pendingCount > 0 ? 'charts.yellow' : 'charts.gray'
        )
      );

      const unpaidAmount = metrics.unpaidAmount || 0;
      items.push(
        new DashboardMetricItem(
          'Unpaid Amount',
          Formatter.formatCurrency(unpaidAmount, metrics.currency || 'USD'),
          unpaidAmount > 0 ? 'alert' : 'check',
          unpaidAmount > 0 ? 'charts.red' : 'charts.green'
        )
      );

      return items;
    } catch (_error) {
      return [
        new DashboardMetricItem(
          'Error',
          'Failed to load stats',
          'error',
          'errorForeground'
        ),
      ];
    }
  }

  private async getDashboardMetrics(): Promise<DashboardMetrics> {
    // Return cached metrics if still fresh
    const now = Date.now();
    if (this.cachedMetrics && now - this.cacheTimestamp < this.CACHE_DURATION) {
      return this.cachedMetrics;
    }

    try {
      const result = await this.cli.exec(['dashboard']);

      if (!result.success || !result.stdout) {
        return this.getDefaultMetrics();
      }

      // Parse dashboard output
      const metrics = this.parseDashboardOutput(result.stdout);
      this.cachedMetrics = metrics;
      this.cacheTimestamp = now;

      return metrics;
    } catch (error) {
      console.error('Failed to parse dashboard:', error);
      return this.getDefaultMetrics();
    }
  }

  private parseDashboardOutput(output: string): DashboardMetrics {
    const metrics: DashboardMetrics = {
      totalMonthlyRevenue: 0,
      hourlyRevenue: 0,
      retainerRevenue: 0,
      projectedHours: 0,
      averageHourlyRate: 0,
      activeContracts: 0,
      totalClients: 0,
      pendingInvoices: 0,
      unpaidAmount: 0,
      currency: 'USD',
    };

    const lines = output.split('\n');

    for (const line of lines) {
      if (line.includes('Total Monthly Revenue') || line.includes('TOTAL')) {
        const match = line.match(/\$?([0-9,]+\.?[0-9]*)/);
        if (match) {
          metrics.totalMonthlyRevenue = parseFloat(match[1].replace(/,/g, ''));
        }
      }
      if (line.includes('Hourly Contracts')) {
        const match = line.match(/\$?([0-9,]+\.?[0-9]*)/);
        if (match) {
          metrics.hourlyRevenue = parseFloat(match[1].replace(/,/g, ''));
        }
      }
      if (line.includes('Retainer Contracts')) {
        const match = line.match(/\$?([0-9,]+\.?[0-9]*)/);
        if (match) {
          metrics.retainerRevenue = parseFloat(match[1].replace(/,/g, ''));
        }
      }
      if (line.includes('Projected Hours')) {
        const match = line.match(/([0-9,]+\.?[0-9]*)/);
        if (match) {
          metrics.projectedHours = parseFloat(match[1].replace(/,/g, ''));
        }
      }
      if (line.includes('Average Rate')) {
        const match = line.match(/\$?([0-9,]+\.?[0-9]*)/);
        if (match) {
          metrics.averageHourlyRate = parseFloat(match[1].replace(/,/g, ''));
        }
      }
      if (line.includes('Total Clients')) {
        const match = line.match(/:\s*(\d+)/);
        if (match) {
          metrics.totalClients = parseInt(match[1], 10);
        }
      }
      if (line.includes('Pending Invoices')) {
        const match = line.match(/:\s*(\d+)/);
        if (match) {
          metrics.pendingInvoices = parseInt(match[1], 10);
        }
      }
      if (line.includes('Unpaid Amount')) {
        const match = line.match(/\$?([0-9,]+\.?[0-9]*)/);
        if (match) {
          metrics.unpaidAmount = parseFloat(match[1].replace(/,/g, ''));
        }
      }
    }

    return metrics;
  }

  private getDefaultMetrics(): DashboardMetrics {
    return {
      totalMonthlyRevenue: 0,
      hourlyRevenue: 0,
      retainerRevenue: 0,
      projectedHours: 0,
      averageHourlyRate: 0,
      activeContracts: 0,
      totalClients: 0,
      pendingInvoices: 0,
      unpaidAmount: 0,
      currency: 'USD',
    };
  }
}
