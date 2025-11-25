import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';
import { Formatter } from '../utils/formatting';

type TrackingTreeItem = TrackingSectionItem | TrackingSummaryItem | TrackingItem | TrackingActionItem;

/**
 * Tracking section for grouping
 */
class TrackingSectionItem extends vscode.TreeItem {
    constructor(
        public readonly label: string,
        public readonly section: string,
        public readonly icon: string,
        public readonly colorId: string,
        count?: number
    ) {
        super(label, vscode.TreeItemCollapsibleState.Expanded);
        if (count !== undefined) {
            this.description = `${count}`;
        }
        this.iconPath = new vscode.ThemeIcon(icon, new vscode.ThemeColor(colorId));
        this.contextValue = `tracking_section_${section}`;
    }
}

/**
 * Tracking summary item
 */
class TrackingSummaryItem extends vscode.TreeItem {
    constructor(
        label: string,
        value: string,
        icon: string,
        colorId?: string
    ) {
        super(label, vscode.TreeItemCollapsibleState.None);
        this.description = value;
        this.iconPath = colorId
            ? new vscode.ThemeIcon(icon, new vscode.ThemeColor(colorId))
            : new vscode.ThemeIcon(icon);
        this.contextValue = 'tracking_summary';
    }
}

/**
 * Tracking action item
 */
class TrackingActionItem extends vscode.TreeItem {
    constructor(
        label: string,
        icon: string,
        commandId: string,
        colorId?: string,
        tooltip?: string
    ) {
        super(label, vscode.TreeItemCollapsibleState.None);
        this.iconPath = colorId
            ? new vscode.ThemeIcon(icon, new vscode.ThemeColor(colorId))
            : new vscode.ThemeIcon(icon);
        this.command = {
            command: commandId,
            title: label
        };
        this.contextValue = 'tracking_action';
        if (tooltip) {
            this.tooltip = tooltip;
        }
    }
}

/**
 * Tracking session tree item
 */
export class TrackingItem extends vscode.TreeItem {
    constructor(
        public readonly itemId: number,
        public readonly project: string,
        public readonly client: string,
        public readonly duration: string,
        public readonly durationMinutes: number,
        public readonly date: string,
        public readonly billable: boolean,
        public readonly collapsibleState: vscode.TreeItemCollapsibleState
    ) {
        super(project || 'Untitled Session', collapsibleState);

        this.id = `tracking_${itemId}`;

        // Rich tooltip
        this.tooltip = new vscode.MarkdownString();
        this.tooltip.appendMarkdown(`**${project || 'Untitled Session'}**\n\n`);
        this.tooltip.appendMarkdown(`- Client: ${client || 'N/A'}\n`);
        this.tooltip.appendMarkdown(`- Duration: ${duration}\n`);
        this.tooltip.appendMarkdown(`- Date: ${Formatter.formatDate(date)}\n`);
        this.tooltip.appendMarkdown(`- Billable: ${billable ? '✅ Yes' : '❌ No'}\n`);

        this.description = `${client || 'No client'} • ${duration}`;
        this.contextValue = 'tracking';

        // Set icon based on billable status
        this.iconPath = billable
            ? new vscode.ThemeIcon('clock', new vscode.ThemeColor('charts.green'))
            : new vscode.ThemeIcon('clock', new vscode.ThemeColor('charts.gray'));
    }
}

interface ParsedSession {
    id: number;
    project: string;
    client: string;
    date: string;
    duration: string;
    durationMinutes: number;
    billable: boolean;
}

interface ActiveSession {
    project: string;
    client: string;
    duration: string;
    startTime: string;
}

/**
 * Tracking sessions tree data provider with enhanced features
 */
export class TrackingProvider implements vscode.TreeDataProvider<TrackingTreeItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<TrackingTreeItem | undefined | null | void> =
        new vscode.EventEmitter<TrackingTreeItem | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<TrackingTreeItem | undefined | null | void> =
        this._onDidChangeTreeData.event;

    private cachedSessions: ParsedSession[] = [];
    private activeSession: ActiveSession | null = null;

    constructor(private cli: UngCli) {}

    refresh(): void {
        this.cachedSessions = [];
        this.activeSession = null;
        this._onDidChangeTreeData.fire();
    }

    getTreeItem(element: TrackingTreeItem): vscode.TreeItem {
        return element;
    }

    async getChildren(element?: TrackingTreeItem): Promise<TrackingTreeItem[]> {
        if (!element) {
            return this.getRootItems();
        }

        if (element instanceof TrackingSectionItem) {
            return this.getItemsForSection(element.section);
        }

        return [];
    }

    private async getRootItems(): Promise<TrackingTreeItem[]> {
        await this.loadData();
        const items: TrackingTreeItem[] = [];

        // Active session section
        if (this.activeSession) {
            items.push(new TrackingSummaryItem(
                '🔴 Active Session',
                this.activeSession.project || 'Untitled',
                'record',
                'charts.red'
            ));

            items.push(new TrackingSummaryItem(
                'Duration',
                this.activeSession.duration,
                'clock',
                'charts.yellow'
            ));

            if (this.activeSession.client) {
                items.push(new TrackingSummaryItem(
                    'Client',
                    this.activeSession.client,
                    'person',
                    'charts.blue'
                ));
            }

            items.push(new TrackingActionItem(
                'Stop Tracking',
                'debug-stop',
                'ung.stopTracking',
                'charts.red',
                'Stop the current tracking session'
            ));
        } else {
            // Quick start tracking
            items.push(new TrackingActionItem(
                'Start Tracking',
                'play',
                'ung.startTracking',
                'charts.green',
                'Start a new time tracking session'
            ));
        }

        // Log time manually (always available)
        items.push(new TrackingActionItem(
            'Log Time Manually',
            'add',
            'ung.logTimeManually',
            'charts.blue',
            'Add a time entry manually'
        ));

        // Today's summary
        const today = new Date().toISOString().split('T')[0];
        const todaySessions = this.cachedSessions.filter(s => s.date.startsWith(today));
        const todayMinutes = todaySessions.reduce((sum, s) => sum + s.durationMinutes, 0);
        const todayBillable = todaySessions.filter(s => s.billable).reduce((sum, s) => sum + s.durationMinutes, 0);

        if (todaySessions.length > 0 || this.activeSession) {
            items.push(new TrackingSummaryItem(
                'Today',
                this.formatMinutes(todayMinutes),
                'calendar',
                'charts.purple'
            ));

            if (todayBillable > 0) {
                items.push(new TrackingSummaryItem(
                    'Billable Today',
                    this.formatMinutes(todayBillable),
                    'credit-card',
                    'charts.green'
                ));
            }
        }

        // This week's summary
        const weekStart = this.getWeekStart();
        const weekSessions = this.cachedSessions.filter(s => new Date(s.date) >= weekStart);
        const weekMinutes = weekSessions.reduce((sum, s) => sum + s.durationMinutes, 0);

        if (weekMinutes > 0) {
            items.push(new TrackingSummaryItem(
                'This Week',
                this.formatMinutes(weekMinutes),
                'calendar',
                'charts.blue'
            ));
        }

        // Recent sessions section
        if (this.cachedSessions.length > 0) {
            items.push(new TrackingSectionItem(
                'Recent Sessions',
                'recent',
                'history',
                'charts.gray',
                Math.min(10, this.cachedSessions.length)
            ));
        }

        // By client section (if there are multiple clients)
        const clientNames = [...new Set(this.cachedSessions.map(s => s.client).filter(c => c))];
        if (clientNames.length > 1) {
            items.push(new TrackingSectionItem(
                'By Client',
                'clients',
                'organization',
                'charts.purple',
                clientNames.length
            ));
        }

        return items;
    }

    private async getItemsForSection(section: string): Promise<TrackingTreeItem[]> {
        if (section === 'recent') {
            return this.cachedSessions
                .slice(0, 10)
                .map(s => new TrackingItem(
                    s.id,
                    s.project,
                    s.client,
                    s.duration,
                    s.durationMinutes,
                    s.date,
                    s.billable,
                    vscode.TreeItemCollapsibleState.None
                ));
        }

        if (section === 'clients') {
            const clientNames = [...new Set(this.cachedSessions.map(s => s.client).filter(c => c))];
            return clientNames.map(client => {
                const clientSessions = this.cachedSessions.filter(s => s.client === client);
                const totalMinutes = clientSessions.reduce((sum, s) => sum + s.durationMinutes, 0);
                return new TrackingSummaryItem(
                    client,
                    `${clientSessions.length} sessions • ${this.formatMinutes(totalMinutes)}`,
                    'person',
                    'charts.blue'
                );
            });
        }

        return [];
    }

    private async loadData(): Promise<void> {
        await Promise.all([
            this.loadActiveSession(),
            this.loadSessions()
        ]);
    }

    private async loadActiveSession(): Promise<void> {
        try {
            const result = await this.cli.exec(['tracking', 'status']);
            if (result.success && result.stdout && !result.stdout.includes('No active')) {
                const lines = result.stdout.split('\n');
                let project = '';
                let client = '';
                let duration = '0:00';
                let startTime = '';

                for (const line of lines) {
                    if (line.includes('Project:')) {
                        project = line.split(':').slice(1).join(':').trim();
                    } else if (line.includes('Client:')) {
                        client = line.split(':').slice(1).join(':').trim();
                    } else if (line.includes('Duration:') || line.includes('Elapsed:')) {
                        duration = line.split(':').slice(1).join(':').trim();
                    } else if (line.includes('Started:') || line.includes('Start:')) {
                        startTime = line.split(':').slice(1).join(':').trim();
                    }
                }

                this.activeSession = { project, client, duration, startTime };
            } else {
                this.activeSession = null;
            }
        } catch {
            this.activeSession = null;
        }
    }

    private async loadSessions(): Promise<void> {
        if (this.cachedSessions.length > 0) {
            return;
        }

        try {
            const result = await this.cli.listTrackingSessions();

            if (!result.success || !result.stdout) {
                this.cachedSessions = [];
                return;
            }

            this.cachedSessions = this.parseTrackingOutput(result.stdout);
        } catch (error) {
            vscode.window.showErrorMessage(`Failed to load tracking sessions: ${error}`);
            this.cachedSessions = [];
        }
    }

    /**
     * Parse tracking sessions output from CLI
     */
    private parseTrackingOutput(output: string): ParsedSession[] {
        const lines = output.split('\n').filter(line => line.trim());
        const sessions: ParsedSession[] = [];

        for (let i = 1; i < lines.length; i++) { // Skip header
            const line = lines[i].trim();
            if (!line) continue;

            const parts = line.split(/\s{2,}/); // Split by multiple spaces
            if (parts.length >= 6) {
                const id = parseInt(parts[0]);
                const project = parts[1] || 'Untitled';
                const client = parts[2] || '';
                const date = parts[3];
                const duration = parts[4];
                const billable = parts[5]?.toLowerCase() === 'yes' ||
                                parts[5]?.toLowerCase() === 'true';

                // Parse duration to minutes
                const durationMinutes = this.parseDurationToMinutes(duration);

                if (!isNaN(id)) {
                    sessions.push({
                        id,
                        project,
                        client,
                        date,
                        duration,
                        durationMinutes,
                        billable
                    });
                }
            }
        }

        // Sort by date (newest first)
        return sessions.sort((a, b) =>
            new Date(b.date).getTime() - new Date(a.date).getTime()
        );
    }

    private parseDurationToMinutes(duration: string): number {
        // Handle formats like "2h 30m", "2:30", "150m", "2.5h"
        let minutes = 0;

        const hourMatch = duration.match(/(\d+(?:\.\d+)?)\s*h/i);
        const minMatch = duration.match(/(\d+)\s*m/i);
        const colonMatch = duration.match(/(\d+):(\d+)/);

        if (colonMatch) {
            minutes = parseInt(colonMatch[1]) * 60 + parseInt(colonMatch[2]);
        } else {
            if (hourMatch) {
                minutes += parseFloat(hourMatch[1]) * 60;
            }
            if (minMatch) {
                minutes += parseInt(minMatch[1]);
            }
        }

        return Math.round(minutes);
    }

    private formatMinutes(minutes: number): string {
        const hours = Math.floor(minutes / 60);
        const mins = minutes % 60;
        if (hours > 0) {
            return mins > 0 ? `${hours}h ${mins}m` : `${hours}h`;
        }
        return `${mins}m`;
    }

    private getWeekStart(): Date {
        const now = new Date();
        const day = now.getDay();
        const diff = now.getDate() - day + (day === 0 ? -6 : 1);
        return new Date(now.setDate(diff));
    }
}
