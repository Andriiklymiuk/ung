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
import { Config } from './utils/config';

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

    // Check if CLI is installed
    const isInstalled = await cli.isInstalled();
    if (!isInstalled) {
        const action = await vscode.window.showWarningMessage(
            'UNG CLI is not installed or not in PATH. Please install it first.',
            'Open Installation Guide'
        );

        if (action === 'Open Installation Guide') {
            vscode.env.openExternal(vscode.Uri.parse('https://github.com/Andriiklymiuk/ung#installation'));
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
        vscode.commands.registerCommand('ung.editClient', (item) => clientCommands.editClient(item?.id)),
        vscode.commands.registerCommand('ung.deleteClient', (item) => clientCommands.deleteClient(item?.id)),
        vscode.commands.registerCommand('ung.listClients', () => clientCommands.listClients()),
        vscode.commands.registerCommand('ung.refreshClients', () => clientProvider.refresh())
    );

    // Contract commands
    context.subscriptions.push(
        vscode.commands.registerCommand('ung.createContract', () => contractCommands.createContract()),
        vscode.commands.registerCommand('ung.viewContract', (item) => contractCommands.viewContract(item?.id)),
        vscode.commands.registerCommand('ung.editContract', (item) => contractCommands.editContract(item?.id)),
        vscode.commands.registerCommand('ung.deleteContract', (item) => contractCommands.deleteContract(item?.id)),
        vscode.commands.registerCommand('ung.generateContractPDF', (item) => contractCommands.generateContractPDF(item?.id)),
        vscode.commands.registerCommand('ung.refreshContracts', () => contractProvider.refresh())
    );

    // Invoice commands
    context.subscriptions.push(
        vscode.commands.registerCommand('ung.createInvoice', () => invoiceCommands.createInvoice()),
        vscode.commands.registerCommand('ung.generateFromTime', () => invoiceCommands.generateFromTime()),
        vscode.commands.registerCommand('ung.viewInvoice', (item) => {
            if (item?.id) {
                InvoicePanel.createOrShow(cli, item.id);
            } else {
                invoiceCommands.viewInvoice(item?.id);
            }
        }),
        vscode.commands.registerCommand('ung.editInvoice', (item) => invoiceCommands.editInvoice(item?.id)),
        vscode.commands.registerCommand('ung.deleteInvoice', (item) => invoiceCommands.deleteInvoice(item?.id)),
        vscode.commands.registerCommand('ung.exportInvoice', (item) => invoiceCommands.exportInvoice(item?.id)),
        vscode.commands.registerCommand('ung.emailInvoice', (item) => invoiceCommands.emailInvoice(item?.id)),
        vscode.commands.registerCommand('ung.refreshInvoices', () => invoiceProvider.refresh())
    );

    // Expense commands
    context.subscriptions.push(
        vscode.commands.registerCommand('ung.logExpense', () => expenseCommands.logExpense()),
        vscode.commands.registerCommand('ung.editExpense', (item) => expenseCommands.editExpense(item?.id)),
        vscode.commands.registerCommand('ung.deleteExpense', (item) => expenseCommands.deleteExpense(item?.id)),
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
