import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';

/**
 * Tracking session tree item
 */
export class TrackingItem extends vscode.TreeItem {
    constructor(
        public readonly itemId: number,
        public readonly project: string,
        public readonly client: string,
        public readonly duration: string,
        public readonly date: string,
        public readonly billable: boolean,
        public readonly collapsibleState: vscode.TreeItemCollapsibleState
    ) {
        super(project || 'Untitled Session', collapsibleState);

        this.id = String(itemId);
        this.tooltip = `${project || 'Untitled Session'}\nClient: ${client}\nDuration: ${duration}\nDate: ${date}\nBillable: ${billable}`;
        this.description = `${client} • ${duration}`;
        this.contextValue = 'tracking';

        // Set icon based on billable status
        this.iconPath = billable
            ? new vscode.ThemeIcon('clock', new vscode.ThemeColor('charts.green'))
            : new vscode.ThemeIcon('clock', new vscode.ThemeColor('charts.gray'));
    }
}

/**
 * Tracking sessions tree data provider
 */
export class TrackingProvider implements vscode.TreeDataProvider<TrackingItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<TrackingItem | undefined | null | void> = new vscode.EventEmitter<TrackingItem | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<TrackingItem | undefined | null | void> = this._onDidChangeTreeData.event;

    constructor(private cli: UngCli) {}

    refresh(): void {
        this._onDidChangeTreeData.fire();
    }

    getTreeItem(element: TrackingItem): vscode.TreeItem {
        return element;
    }

    async getChildren(element?: TrackingItem): Promise<TrackingItem[]> {
        if (element) {
            return [];
        }

        try {
            const result = await this.cli.listTrackingSessions();

            if (!result.success || !result.stdout) {
                return [];
            }

            // Parse the CLI output
            const sessions = this.parseTrackingOutput(result.stdout);
            return sessions;
        } catch (error) {
            vscode.window.showErrorMessage(`Failed to load tracking sessions: ${error}`);
            return [];
        }
    }

    /**
     * Parse tracking sessions output from CLI
     */
    private parseTrackingOutput(output: string): TrackingItem[] {
        const lines = output.split('\n').filter(line => line.trim());
        const sessions: TrackingItem[] = [];

        for (let i = 1; i < lines.length; i++) { // Skip header
            const line = lines[i].trim();
            if (!line) continue;

            const parts = line.split(/\s{2,}/); // Split by multiple spaces
            if (parts.length >= 6) {
                const id = parseInt(parts[0]);
                const project = parts[1] || 'Untitled';
                const client = parts[2] || '-';
                const date = parts[3];
                const duration = parts[4];
                const billable = parts[5]?.toLowerCase() === 'yes';

                if (!isNaN(id)) {
                    sessions.push(new TrackingItem(
                        id,
                        project,
                        client,
                        duration,
                        date,
                        billable,
                        vscode.TreeItemCollapsibleState.None
                    ));
                }
            }
        }

        return sessions;
    }
}
