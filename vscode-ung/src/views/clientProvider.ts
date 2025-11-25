import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';

/**
 * Client tree item
 */
export class ClientItem extends vscode.TreeItem {
    constructor(
        public readonly id: number,
        public readonly name: string,
        public readonly email: string,
        public readonly collapsibleState: vscode.TreeItemCollapsibleState
    ) {
        super(name, collapsibleState);

        this.tooltip = `${name}\n${email}`;
        this.description = email;
        this.contextValue = 'client';
        this.iconPath = new vscode.ThemeIcon('person');
    }
}

/**
 * Client tree data provider
 */
export class ClientProvider implements vscode.TreeDataProvider<ClientItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<ClientItem | undefined | null | void> = new vscode.EventEmitter<ClientItem | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<ClientItem | undefined | null | void> = this._onDidChangeTreeData.event;

    constructor(private cli: UngCli) {}

    refresh(): void {
        this._onDidChangeTreeData.fire();
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
                return [];
            }

            // Parse the CLI output
            const clients = this.parseClientOutput(result.stdout);
            return clients;
        } catch (error) {
            vscode.window.showErrorMessage(`Failed to load clients: ${error}`);
            return [];
        }
    }

    /**
     * Parse client list output from CLI
     */
    private parseClientOutput(output: string): ClientItem[] {
        const lines = output.split('\n').filter(line => line.trim());
        const clients: ClientItem[] = [];

        for (let i = 1; i < lines.length; i++) { // Skip header
            const line = lines[i].trim();
            if (!line) continue;

            const parts = line.split(/\s{2,}/); // Split by multiple spaces
            if (parts.length >= 3) {
                const id = parseInt(parts[0]);
                const name = parts[1];
                const email = parts[2];

                if (!isNaN(id)) {
                    clients.push(new ClientItem(
                        id,
                        name,
                        email,
                        vscode.TreeItemCollapsibleState.None
                    ));
                }
            }
        }

        return clients;
    }
}
