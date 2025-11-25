import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';
import { Formatter } from '../utils/formatting';

/**
 * Contract tree item
 */
export class ContractItem extends vscode.TreeItem {
    constructor(
        public readonly id: number,
        public readonly name: string,
        public readonly clientName: string,
        public readonly type: string,
        public readonly rate: string,
        public readonly active: boolean,
        public readonly collapsibleState: vscode.TreeItemCollapsibleState
    ) {
        super(name, collapsibleState);

        this.tooltip = `${name}\nClient: ${clientName}\nType: ${type}\nRate: ${rate}\nActive: ${active}`;
        this.description = `${clientName} • ${rate}`;
        this.contextValue = 'contract';

        // Set icon
        this.iconPath = active
            ? new vscode.ThemeIcon('file-code', new vscode.ThemeColor('charts.green'))
            : new vscode.ThemeIcon('file-code', new vscode.ThemeColor('charts.gray'));
    }
}

/**
 * Contract tree data provider
 */
export class ContractProvider implements vscode.TreeDataProvider<ContractItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<ContractItem | undefined | null | void> = new vscode.EventEmitter<ContractItem | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<ContractItem | undefined | null | void> = this._onDidChangeTreeData.event;

    constructor(private cli: UngCli) {}

    refresh(): void {
        this._onDidChangeTreeData.fire();
    }

    getTreeItem(element: ContractItem): vscode.TreeItem {
        return element;
    }

    async getChildren(element?: ContractItem): Promise<ContractItem[]> {
        if (element) {
            return [];
        }

        try {
            const result = await this.cli.listContracts();

            if (!result.success || !result.stdout) {
                return [];
            }

            // Parse the CLI output
            const contracts = this.parseContractOutput(result.stdout);
            return contracts;
        } catch (error) {
            vscode.window.showErrorMessage(`Failed to load contracts: ${error}`);
            return [];
        }
    }

    /**
     * Parse contract list output from CLI
     */
    private parseContractOutput(output: string): ContractItem[] {
        const lines = output.split('\n').filter(line => line.trim());
        const contracts: ContractItem[] = [];

        for (let i = 1; i < lines.length; i++) { // Skip header
            const line = lines[i].trim();
            if (!line) continue;

            const parts = line.split(/\s{2,}/); // Split by multiple spaces
            if (parts.length >= 5) {
                const id = parseInt(parts[0]);
                const name = parts[1];
                const clientName = parts[2];
                const type = parts[3];
                const rate = parts[4];
                const active = parts[5]?.toLowerCase() === 'true' || parts[5]?.toLowerCase() === 'active';

                if (!isNaN(id)) {
                    contracts.push(new ContractItem(
                        id,
                        name,
                        clientName,
                        type,
                        rate,
                        active,
                        vscode.TreeItemCollapsibleState.None
                    ));
                }
            }
        }

        return contracts;
    }
}
