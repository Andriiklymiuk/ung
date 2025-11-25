import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';
import { Formatter } from '../utils/formatting';

/**
 * Dashboard metric item
 */
class DashboardMetricItem extends vscode.TreeItem {
    constructor(
        public readonly label: string,
        public readonly value: string,
        public readonly icon: string
    ) {
        super(label, vscode.TreeItemCollapsibleState.None);
        this.description = value;
        this.tooltip = `${label}: ${value}`;
        this.iconPath = new vscode.ThemeIcon(icon);
    }
}

/**
 * Dashboard section item
 */
class DashboardSectionItem extends vscode.TreeItem {
    constructor(
        public readonly label: string,
        public readonly section: string,
        public readonly icon: string = 'folder'
    ) {
        super(label, vscode.TreeItemCollapsibleState.Expanded);
        this.contextValue = `dashboard_${section}`;
        this.iconPath = new vscode.ThemeIcon(icon);
    }
}

/**
 * Dashboard tree data provider
 */
export class DashboardProvider implements vscode.TreeDataProvider<DashboardSectionItem | DashboardMetricItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<DashboardSectionItem | DashboardMetricItem | undefined | null | void> =
        new vscode.EventEmitter<DashboardSectionItem | DashboardMetricItem | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<DashboardSectionItem | DashboardMetricItem | undefined | null | void> =
        this._onDidChangeTreeData.event;

    private cachedMetrics: any = null;
    private cacheTimestamp: number = 0;
    private readonly CACHE_DURATION = 60000; // 1 minute

    constructor(private cli: UngCli) {}

    refresh(): void {
        this.cachedMetrics = null;
        this._onDidChangeTreeData.fire();
    }

    getTreeItem(element: DashboardSectionItem | DashboardMetricItem): vscode.TreeItem {
        return element;
    }

    async getChildren(element?: DashboardSectionItem | DashboardMetricItem): Promise<(DashboardSectionItem | DashboardMetricItem)[]> {
        if (!element) {
            // Root level - show dashboard sections
            return [
                new DashboardSectionItem('Revenue', 'revenue', 'dollar'),
                new DashboardSectionItem('Contracts', 'contracts', 'briefcase'),
                new DashboardSectionItem('Quick Stats', 'stats', 'dashboard'),
            ];
        }

        if (element instanceof DashboardSectionItem) {
            if (element.section === 'revenue') {
                return this.getRevenueMetrics();
            } else if (element.section === 'contracts') {
                return this.getContractMetrics();
            } else if (element.section === 'stats') {
                return this.getStatsMetrics();
            }
        }

        return [];
    }

    private async getRevenueMetrics(): Promise<DashboardMetricItem[]> {
        try {
            const metrics = await this.getDashboardMetrics();
            const items: DashboardMetricItem[] = [];

            items.push(new DashboardMetricItem(
                'Monthly Revenue',
                Formatter.formatCurrency(metrics.totalMonthlyRevenue || 0, metrics.currency || 'USD'),
                'symbol-number'
            ));

            items.push(new DashboardMetricItem(
                'Hourly Revenue',
                Formatter.formatCurrency(metrics.hourlyRevenue || 0, metrics.currency || 'USD'),
                'symbol-numeric'
            ));

            items.push(new DashboardMetricItem(
                'Retainer Revenue',
                Formatter.formatCurrency(metrics.retainerRevenue || 0, metrics.currency || 'USD'),
                'symbol-numeric'
            ));

            items.push(new DashboardMetricItem(
                'Projected Hours',
                `${(metrics.projectedHours || 0).toFixed(1)}h`,
                'clock'
            ));

            if (metrics.averageHourlyRate > 0) {
                items.push(new DashboardMetricItem(
                    'Average Rate',
                    `$${metrics.averageHourlyRate.toFixed(0)}/hr`,
                    'symbol-number'
                ));
            }

            return items;
        } catch (error) {
            return [new DashboardMetricItem('Error', 'Failed to load revenue metrics', 'error')];
        }
    }

    private async getContractMetrics(): Promise<DashboardMetricItem[]> {
        try {
            const metrics = await this.getDashboardMetrics();
            const items: DashboardMetricItem[] = [];

            items.push(new DashboardMetricItem(
                'Active Contracts',
                String(metrics.activeContracts || 0),
                'briefcase'
            ));

            // Show top contracts if available
            if (metrics.topContracts && metrics.topContracts.length > 0) {
                for (let i = 0; i < Math.min(3, metrics.topContracts.length); i++) {
                    const contract = metrics.topContracts[i];
                    items.push(new DashboardMetricItem(
                        contract.clientName,
                        Formatter.formatCurrency(contract.monthlyRevenue || 0, contract.currency || 'USD'),
                        'briefcase'
                    ));
                }
            }

            return items;
        } catch (error) {
            return [new DashboardMetricItem('Error', 'Failed to load contract metrics', 'error')];
        }
    }

    private async getStatsMetrics(): Promise<DashboardMetricItem[]> {
        try {
            const metrics = await this.getDashboardMetrics();
            const items: DashboardMetricItem[] = [];

            items.push(new DashboardMetricItem(
                'Total Clients',
                String(metrics.totalClients || 0),
                'person'
            ));

            items.push(new DashboardMetricItem(
                'Pending Invoices',
                String(metrics.pendingInvoices || 0),
                'clock'
            ));

            items.push(new DashboardMetricItem(
                'Unpaid Amount',
                Formatter.formatCurrency(metrics.unpaidAmount || 0, metrics.currency || 'USD'),
                'warning'
            ));

            return items;
        } catch (error) {
            return [new DashboardMetricItem('Error', 'Failed to load statistics', 'error')];
        }
    }

    private async getDashboardMetrics(): Promise<any> {
        // Return cached metrics if still fresh
        const now = Date.now();
        if (this.cachedMetrics && (now - this.cacheTimestamp) < this.CACHE_DURATION) {
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

    private parseDashboardOutput(output: string): any {
        const metrics: any = {
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
            topContracts: []
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

    private getDefaultMetrics(): any {
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
            topContracts: []
        };
    }
}
