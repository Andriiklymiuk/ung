import * as vscode from 'vscode';
import { UngCli } from './cli/ungCli';
import { CompanyCommands } from './commands/company';
import { ClientCommands } from './commands/client';
import { ContractCommands } from './commands/contract';
import { InvoiceCommands } from './commands/invoice';
import { ExpenseCommands } from './commands/expense';
import { TrackingCommands } from './commands/tracking';
import { InvoiceProvider } from './views/invoiceProvider';
import { ContractProvider } from './views/contractProvider';
import { ClientProvider } from './views/clientProvider';
import { ExpenseProvider } from './views/expenseProvider';
import { TrackingProvider } from './views/trackingProvider';
import { DashboardProvider } from './views/dashboardProvider';
import { InvoicePanel } from './webview/invoicePanel';
import { ExportPanel } from './webview/exportPanel';
import { StatusBarManager } from './utils/statusBar';

/**
 * Extension activation
 */
export async function activate(context: vscode.ExtensionContext) {
    console.log('UNG extension is now active!');

    // Create output channel
    const outputChannel = vscode.window.createOutputChannel('UNG Operations');
    context.subscriptions.push(outputChannel);

    // Initialize CLI wrapper
    const cli = new UngCli(outputChannel);

    // Register install command (always available)
    context.subscriptions.push(
        vscode.commands.registerCommand('ung.installCli', async () => {
            const action = await vscode.window.showQuickPick(
                [
                    { label: '$(terminal) Install via Homebrew', description: 'Recommended for macOS/Linux', value: 'homebrew' },
                    { label: '$(cloud-download) Download from GitHub', description: 'Download binary directly', value: 'github' },
                    { label: '$(book) View Instructions', description: 'Open installation guide', value: 'docs' }
                ],
                { placeHolder: 'Choose installation method' }
            );

            if (action?.value === 'homebrew') {
                const terminal = vscode.window.createTerminal('UNG Installation');
                terminal.show();
                terminal.sendText('brew tap Andriiklymiuk/tools && brew install ung');
                vscode.window.showInformationMessage('Installing UNG CLI... Reload VSCode after installation completes.', 'Reload').then(choice => {
                    if (choice === 'Reload') {
                        vscode.commands.executeCommand('workbench.action.reloadWindow');
                    }
                });
            } else if (action?.value === 'github') {
                vscode.env.openExternal(vscode.Uri.parse('https://github.com/Andriiklymiuk/ung/releases/latest'));
            } else if (action?.value === 'docs') {
                vscode.env.openExternal(vscode.Uri.parse('https://github.com/Andriiklymiuk/ung#installation'));
            }
        })
    );

    // Register check for updates command
    context.subscriptions.push(
        vscode.commands.registerCommand('ung.checkForUpdates', async () => {
            await vscode.window.withProgress(
                { location: vscode.ProgressLocation.Notification, title: 'Checking for UNG updates...' },
                async () => {
                    try {
                        const currentVersion = await cli.getVersion();
                        const response = await fetch('https://api.github.com/repos/Andriiklymiuk/ung/releases/latest');
                        const data = await response.json() as { tag_name?: string };
                        const latestVersion = data.tag_name?.replace('v', '') || '';

                        if (!currentVersion || !latestVersion) {
                            vscode.window.showErrorMessage('Could not check for updates.');
                            return;
                        }

                        const current = currentVersion.replace('v', '').trim();
                        if (current === latestVersion) {
                            vscode.window.showInformationMessage(`UNG CLI is up to date (v${current})`);
                        } else {
                            const action = await vscode.window.showInformationMessage(
                                `UNG CLI update available: v${current} → v${latestVersion}`,
                                'Update Now',
                                'View Release Notes'
                            );

                            if (action === 'Update Now') {
                                const terminal = vscode.window.createTerminal('UNG Update');
                                terminal.show();
                                terminal.sendText('brew upgrade ung');
                                vscode.window.showInformationMessage('Updating UNG CLI... Reload VSCode after update completes.', 'Reload').then(choice => {
                                    if (choice === 'Reload') {
                                        vscode.commands.executeCommand('workbench.action.reloadWindow');
                                    }
                                });
                            } else if (action === 'View Release Notes') {
                                vscode.env.openExternal(vscode.Uri.parse('https://github.com/Andriiklymiuk/ung/releases/latest'));
                            }
                        }
                    } catch (error) {
                        vscode.window.showErrorMessage('Failed to check for updates. Please check your internet connection.');
                    }
                }
            );
        })
    );

    // Check if CLI is installed
    const isInstalled = await cli.isInstalled();
    vscode.commands.executeCommand('setContext', 'ung.cliInstalled', isInstalled);

    if (!isInstalled) {
        // Show error with install option
        const action = await vscode.window.showErrorMessage(
            'UNG CLI is not installed. Install it to use billing and time tracking features.',
            'Install Now'
        );

        if (action === 'Install Now') {
            vscode.commands.executeCommand('ung.installCli');
        }
        return;
    }

    // Show version info
    const version = await cli.getVersion();
    if (version) {
        outputChannel.appendLine(`UNG CLI version: ${version}`);
    }

    // Initialize status bar
    const statusBar = new StatusBarManager(cli);
    context.subscriptions.push(statusBar);
    await statusBar.start();

    // Initialize tree view providers
    const invoiceProvider = new InvoiceProvider(cli);
    const contractProvider = new ContractProvider(cli);
    const clientProvider = new ClientProvider(cli);
    const expenseProvider = new ExpenseProvider(cli);
    const trackingProvider = new TrackingProvider(cli);
    const dashboardProvider = new DashboardProvider(cli);

    // Register tree views
    const invoicesTree = vscode.window.createTreeView('ungInvoices', {
        treeDataProvider: invoiceProvider
    });
    const contractsTree = vscode.window.createTreeView('ungContracts', {
        treeDataProvider: contractProvider
    });
    const clientsTree = vscode.window.createTreeView('ungClients', {
        treeDataProvider: clientProvider
    });
    const expensesTree = vscode.window.createTreeView('ungExpenses', {
        treeDataProvider: expenseProvider
    });
    const trackingTree = vscode.window.createTreeView('ungTracking', {
        treeDataProvider: trackingProvider
    });
    const dashboardTree = vscode.window.createTreeView('ungDashboard', {
        treeDataProvider: dashboardProvider
    });

    context.subscriptions.push(invoicesTree, contractsTree, clientsTree, expensesTree, trackingTree, dashboardTree);

    // Initialize command handlers
    const companyCommands = new CompanyCommands(cli);
    const clientCommands = new ClientCommands(cli, () => clientProvider.refresh());
    const contractCommands = new ContractCommands(cli, () => contractProvider.refresh());
    const invoiceCommands = new InvoiceCommands(cli, () => invoiceProvider.refresh());
    const expenseCommands = new ExpenseCommands(cli, () => expenseProvider.refresh());
    const trackingCommands = new TrackingCommands(cli, statusBar, () => trackingProvider.refresh());

    // Register all commands

    // Company commands
    context.subscriptions.push(
        vscode.commands.registerCommand('ung.createCompany', () => companyCommands.createCompany()),
        vscode.commands.registerCommand('ung.editCompany', () => companyCommands.editCompany())
    );

    // Client commands
    context.subscriptions.push(
        vscode.commands.registerCommand('ung.createClient', () => clientCommands.createClient()),
        vscode.commands.registerCommand('ung.editClient', (item) => clientCommands.editClient(item?.itemId)),
        vscode.commands.registerCommand('ung.deleteClient', (item) => clientCommands.deleteClient(item?.itemId)),
        vscode.commands.registerCommand('ung.listClients', () => clientCommands.listClients()),
        vscode.commands.registerCommand('ung.refreshClients', () => clientProvider.refresh())
    );

    // Contract commands
    context.subscriptions.push(
        vscode.commands.registerCommand('ung.createContract', () => contractCommands.createContract()),
        vscode.commands.registerCommand('ung.viewContract', (item) => contractCommands.viewContract(item?.itemId)),
        vscode.commands.registerCommand('ung.editContract', (item) => contractCommands.editContract(item?.itemId)),
        vscode.commands.registerCommand('ung.deleteContract', (item) => contractCommands.deleteContract(item?.itemId)),
        vscode.commands.registerCommand('ung.generateContractPDF', (item) => contractCommands.generateContractPDF(item?.itemId)),
        vscode.commands.registerCommand('ung.refreshContracts', () => contractProvider.refresh())
    );

    // Invoice commands
    context.subscriptions.push(
        vscode.commands.registerCommand('ung.createInvoice', () => invoiceCommands.createInvoice()),
        vscode.commands.registerCommand('ung.generateFromTime', () => invoiceCommands.generateFromTime()),
        vscode.commands.registerCommand('ung.viewInvoice', (item) => {
            if (item?.itemId) {
                InvoicePanel.createOrShow(cli, item.itemId);
            } else {
                invoiceCommands.viewInvoice(item?.itemId);
            }
        }),
        vscode.commands.registerCommand('ung.editInvoice', (item) => invoiceCommands.editInvoice(item?.itemId)),
        vscode.commands.registerCommand('ung.deleteInvoice', (item) => invoiceCommands.deleteInvoice(item?.itemId)),
        vscode.commands.registerCommand('ung.exportInvoice', (item) => invoiceCommands.exportInvoice(item?.itemId)),
        vscode.commands.registerCommand('ung.emailInvoice', (item) => invoiceCommands.emailInvoice(item?.itemId)),
        vscode.commands.registerCommand('ung.refreshInvoices', () => invoiceProvider.refresh())
    );

    // Expense commands
    context.subscriptions.push(
        vscode.commands.registerCommand('ung.logExpense', () => expenseCommands.logExpense()),
        vscode.commands.registerCommand('ung.editExpense', (item) => expenseCommands.editExpense(item?.itemId)),
        vscode.commands.registerCommand('ung.deleteExpense', (item) => expenseCommands.deleteExpense(item?.itemId)),
        vscode.commands.registerCommand('ung.viewExpenseReport', () => expenseCommands.viewExpenseReport()),
        vscode.commands.registerCommand('ung.refreshExpenses', () => expenseProvider.refresh())
    );

    // Tracking commands
    context.subscriptions.push(
        vscode.commands.registerCommand('ung.startTracking', () => trackingCommands.startTracking()),
        vscode.commands.registerCommand('ung.stopTracking', () => trackingCommands.stopTracking()),
        vscode.commands.registerCommand('ung.logTimeManually', () => trackingCommands.logTimeManually()),
        vscode.commands.registerCommand('ung.viewActiveSession', () => trackingCommands.viewActiveSession()),
        vscode.commands.registerCommand('ung.refreshTracking', () => trackingProvider.refresh())
    );

    // Dashboard commands
    context.subscriptions.push(
        vscode.commands.registerCommand('ung.refreshDashboard', () => dashboardProvider.refresh())
    );

    // Export wizard command
    context.subscriptions.push(
        vscode.commands.registerCommand('ung.openExportWizard', () => ExportPanel.createOrShow(cli))
    );

    // Show welcome message
    vscode.window.showInformationMessage('UNG Billing extension is ready!');
}

/**
 * Extension deactivation
 */
export function deactivate() {
    console.log('UNG extension is now deactivated');
}
