import * as vscode from 'vscode';
import { UngCli } from './cli/ungCli';
import { ClientCommands } from './commands/client';
import { CommandCenter } from './commands/commandCenter';
import { CompanyCommands } from './commands/company';
import { ContractCommands } from './commands/contract';
import { ExpenseCommands } from './commands/expense';
import { InvoiceCommands } from './commands/invoice';
import { SearchCommands } from './commands/search';
import { SecurityCommands } from './commands/security';
import { TrackingCommands } from './commands/tracking';
import { getSecureStorage, initSecureStorage } from './utils/secureStorage';
import { StatusBarManager } from './utils/statusBar';
import { ClientProvider } from './views/clientProvider';
import { ContractProvider } from './views/contractProvider';
import { ExpenseProvider } from './views/expenseProvider';
import { InvoiceProvider } from './views/invoiceProvider';
import { TrackingProvider } from './views/trackingProvider';
import { DashboardWebviewProvider } from './webview/dashboardWebviewProvider';
import { ExpensePanel } from './webview/expensePanel';
import { ExportPanel } from './webview/exportPanel';
import { MainDashboardPanel } from './webview/mainDashboardPanel';
import { OnboardingWebviewProvider } from './webview/onboardingWebviewProvider';
import { PomodoroPanel } from './webview/pomodoroPanel';
import { StatisticsPanel } from './webview/statisticsPanel';
import { TemplateEditorPanel } from './webview/templateEditorPanel';

/**
 * Extension activation
 */
export async function activate(context: vscode.ExtensionContext) {
  console.log('UNG extension is now active!');

  // Create output channel
  const outputChannel = vscode.window.createOutputChannel('UNG Operations');
  context.subscriptions.push(outputChannel);

  // Initialize secure storage for password management
  initSecureStorage(context);

  // Initialize CLI wrapper
  const cli = new UngCli(outputChannel);

  // Register Onboarding webview provider for sidebar
  const onboardingProvider = new OnboardingWebviewProvider(
    context.extensionUri,
    async () => cli.isInstalled(),
    async () => cli.isInitialized()
  );
  context.subscriptions.push(
    vscode.window.registerWebviewViewProvider(
      OnboardingWebviewProvider.viewType,
      onboardingProvider
    )
  );

  // Register Dashboard webview provider for sidebar
  const dashboardWebviewProvider = new DashboardWebviewProvider(
    context.extensionUri,
    cli
  );
  context.subscriptions.push(
    vscode.window.registerWebviewViewProvider(
      DashboardWebviewProvider.viewType,
      dashboardWebviewProvider
    )
  );

  // Register install commands (always available)
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.installCli', async () => {
      const action = await vscode.window.showQuickPick(
        [
          {
            label: '$(terminal) Install via Homebrew',
            description: 'Recommended for macOS/Linux',
            value: 'homebrew',
          },
          {
            label: '$(book) View Instructions',
            description: 'Open installation guide',
            value: 'docs',
          },
        ],
        { placeHolder: 'Choose installation method' }
      );

      if (action?.value === 'homebrew') {
        vscode.commands.executeCommand('ung.installViaHomebrew');
      } else if (action?.value === 'docs') {
        vscode.env.openExternal(
          vscode.Uri.parse(
            'https://andriiklymiuk.github.io/ung/docs/installation'
          )
        );
      }
    }),

    vscode.commands.registerCommand('ung.installViaHomebrew', () => {
      const terminal = vscode.window.createTerminal('UNG Installation');
      terminal.show();
      terminal.sendText('brew tap Andriiklymiuk/tools && brew install ung');
      vscode.window
        .showInformationMessage(
          'Installing UNG CLI via Homebrew... Reload VS Code after installation completes.',
          'Reload'
        )
        .then((choice) => {
          if (choice === 'Reload') {
            vscode.commands.executeCommand('workbench.action.reloadWindow');
          }
        });
    }),

    vscode.commands.registerCommand('ung.installViaScoop', () => {
      const terminal = vscode.window.createTerminal('UNG Installation');
      terminal.show();
      terminal.sendText(
        'scoop bucket add ung https://github.com/Andriiklymiuk/scoop-bucket && scoop install ung'
      );
      vscode.window
        .showInformationMessage(
          'Installing UNG CLI via Scoop... Reload VS Code after installation completes.',
          'Reload'
        )
        .then((choice) => {
          if (choice === 'Reload') {
            vscode.commands.executeCommand('workbench.action.reloadWindow');
          }
        });
    }),

    vscode.commands.registerCommand('ung.installViaGo', () => {
      const terminal = vscode.window.createTerminal('UNG Installation');
      terminal.show();
      terminal.sendText('go install github.com/Andriiklymiuk/ung@latest');
      vscode.window
        .showInformationMessage(
          'Installing UNG CLI via Go... Reload VS Code after installation completes.',
          'Reload'
        )
        .then((choice) => {
          if (choice === 'Reload') {
            vscode.commands.executeCommand('workbench.action.reloadWindow');
          }
        });
    }),

    vscode.commands.registerCommand(
      'ung.downloadBinary',
      (platform?: string) => {
        const url =
          platform === 'darwin'
            ? 'https://github.com/Andriiklymiuk/ung/releases/latest/download/ung_darwin_amd64.tar.gz'
            : platform === 'windows'
              ? 'https://github.com/Andriiklymiuk/ung/releases/latest/download/ung_windows_amd64.zip'
              : 'https://github.com/Andriiklymiuk/ung/releases/latest/download/ung_linux_amd64.tar.gz';
        vscode.env.openExternal(vscode.Uri.parse(url));
      }
    ),

    vscode.commands.registerCommand('ung.openDocs', () => {
      vscode.env.openExternal(
        vscode.Uri.parse('https://andriiklymiuk.github.io/ung/docs/intro')
      );
    }),

    vscode.commands.registerCommand('ung.recheckCli', async () => {
      const isNowInstalled = await cli.isInstalled();
      vscode.commands.executeCommand(
        'setContext',
        'ung.cliInstalled',
        isNowInstalled
      );
      // Refresh the onboarding webview
      await onboardingProvider.refresh();
      if (isNowInstalled) {
        // Silently reload to enable features
        vscode.commands.executeCommand('workbench.action.reloadWindow');
      } else {
        vscode.window.showWarningMessage(
          'UNG CLI not found. Please install it first.'
        );
      }
    })
  );

  // Register check for updates command
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.checkForUpdates', async () => {
      await vscode.window.withProgress(
        {
          location: vscode.ProgressLocation.Notification,
          title: 'Checking for UNG updates...',
        },
        async () => {
          try {
            const currentVersion = await cli.getVersion();
            const response = await fetch(
              'https://api.github.com/repos/Andriiklymiuk/ung/releases/latest'
            );
            const data = (await response.json()) as { tag_name?: string };
            const latestVersion = data.tag_name?.replace('v', '') || '';

            if (!currentVersion || !latestVersion) {
              vscode.window.showErrorMessage('Could not check for updates.');
              return;
            }

            // Parse version number from output like "ung version 1.0.17" or "v1.0.17"
            const versionMatch = currentVersion.match(/(\d+\.\d+\.\d+)/);
            const current = versionMatch
              ? versionMatch[1]
              : currentVersion.replace(/[^0-9.]/g, '').trim();
            const latest = latestVersion.replace(/[^0-9.]/g, '').trim();

            if (current === latest) {
              vscode.window.showInformationMessage(
                `UNG CLI is up to date (v${current})`
              );
            } else {
              const action = await vscode.window.showInformationMessage(
                `UNG CLI update available: v${current} → v${latest}`,
                'Update Now'
              );

              if (action === 'Update Now') {
                const terminal = vscode.window.createTerminal('UNG Update');
                terminal.show();
                terminal.sendText('brew upgrade ung');
                vscode.window
                  .showInformationMessage(
                    'Updating UNG CLI... Reload VSCode after update completes.',
                    'Reload'
                  )
                  .then((choice) => {
                    if (choice === 'Reload') {
                      vscode.commands.executeCommand(
                        'workbench.action.reloadWindow'
                      );
                    }
                  });
              }
            }
          } catch (_error) {
            vscode.window.showErrorMessage(
              'Failed to check for updates. Please check your internet connection.'
            );
          }
        }
      );
    })
  );

  // Check if CLI is installed
  const isInstalled = await cli.isInstalled();
  vscode.commands.executeCommand('setContext', 'ung.cliInstalled', isInstalled);

  if (!isInstalled) {
    // CLI not installed - the onboarding webview will show installation options
    // No blocking alert needed, the webview handles the UX
    // Refresh the onboarding provider to ensure it has the correct state
    await onboardingProvider.refresh();
    return;
  }

  // Check if UNG is initialized
  const isInitialized = await cli.isInitialized();
  vscode.commands.executeCommand(
    'setContext',
    'ung.isInitialized',
    isInitialized
  );

  // Register initialization commands
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.initializeGlobal', async () => {
      await vscode.window.withProgress(
        {
          location: vscode.ProgressLocation.Notification,
          title: 'Setting up UNG globally...',
        },
        async () => {
          const result = await cli.initialize(true);
          if (result.success) {
            vscode.commands.executeCommand(
              'setContext',
              'ung.isInitialized',
              true
            );
            vscode.window
              .showInformationMessage(
                'UNG is now set up! Your data will be stored in ~/.ung/',
                'Get Started'
              )
              .then((choice) => {
                if (choice === 'Get Started') {
                  vscode.commands.executeCommand(
                    'workbench.action.reloadWindow'
                  );
                }
              });
          } else {
            vscode.window.showErrorMessage(`Setup failed: ${result.error}`);
          }
        }
      );
    }),

    vscode.commands.registerCommand('ung.initializeLocal', async () => {
      await vscode.window.withProgress(
        {
          location: vscode.ProgressLocation.Notification,
          title: 'Setting up UNG for this workspace...',
        },
        async () => {
          // Update the setting to use local config BEFORE initialization
          const config = vscode.workspace.getConfiguration('ung');
          await config.update(
            'useGlobalConfig',
            false,
            vscode.ConfigurationTarget.Workspace
          );

          const result = await cli.initialize(false);
          if (result.success) {
            vscode.commands.executeCommand(
              'setContext',
              'ung.isInitialized',
              true
            );
            vscode.window
              .showInformationMessage(
                'UNG is now set up! Your data will be stored in .ung/ folder',
                'Get Started'
              )
              .then((choice) => {
                if (choice === 'Get Started') {
                  vscode.commands.executeCommand(
                    'workbench.action.reloadWindow'
                  );
                }
              });
          } else {
            vscode.window.showErrorMessage(`Setup failed: ${result.error}`);
          }
        }
      );
    }),

    vscode.commands.registerCommand('ung.refreshSetupRequired', () =>
      onboardingProvider.refresh()
    )
  );

  // If not initialized, the onboarding webview will show setup options
  // No blocking alert needed, the webview handles the UX
  if (!isInitialized) {
    // Refresh the onboarding provider to ensure it has the correct state
    await onboardingProvider.refresh();
    return;
  }

  // Show version info
  const version = await cli.getVersion();
  if (version) {
    outputChannel.appendLine(`UNG CLI version: ${version}`);
  }

  // Check if database is encrypted and password is available
  // If encrypted but no password in keychain, prompt user to save it
  const checkEncryptedDatabase = async () => {
    const isEncrypted = await cli.isEncrypted();
    if (!isEncrypted) {
      return; // Not encrypted, no password needed
    }

    const secureStorage = getSecureStorage();
    const hasVSCodePassword = await secureStorage.hasPassword();
    const passwordAvailable = await cli.isPasswordAvailable();

    if (!passwordAvailable && !hasVSCodePassword) {
      // Database is encrypted but no password is saved anywhere
      const action = await vscode.window.showWarningMessage(
        'Your database is encrypted but no password is saved. Save your password for seamless access.',
        'Save Password',
        'Later'
      );

      if (action === 'Save Password') {
        vscode.commands.executeCommand('ung.savePassword');
      }
    }
  };

  // Run the check (non-blocking)
  checkEncryptedDatabase();

  // Initialize status bar
  const statusBar = new StatusBarManager(cli);
  context.subscriptions.push(statusBar);
  await statusBar.start();

  // Initialize pomodoro status bar item
  const pomodoroStatusBar = vscode.window.createStatusBarItem(
    vscode.StatusBarAlignment.Right,
    200
  );
  pomodoroStatusBar.command = 'ung.startPomodoro';
  context.subscriptions.push(pomodoroStatusBar);

  // Initialize providers for command refresh callbacks
  // Note: Tree views have been consolidated into the single dashboard view
  const invoiceProvider = new InvoiceProvider(cli);
  const contractProvider = new ContractProvider(cli);
  const clientProvider = new ClientProvider(cli);
  const expenseProvider = new ExpenseProvider(cli);
  const trackingProvider = new TrackingProvider(cli);

  // Initialize command handlers
  const companyCommands = new CompanyCommands(cli);
  const clientCommands = new ClientCommands(cli, () => {
    clientProvider.refresh();
    dashboardWebviewProvider.refresh();
  });
  const contractCommands = new ContractCommands(cli, () => {
    contractProvider.refresh();
    dashboardWebviewProvider.refresh();
  });
  const invoiceCommands = new InvoiceCommands(cli, () => {
    invoiceProvider.refresh();
    dashboardWebviewProvider.refresh();
  });
  const expenseCommands = new ExpenseCommands(cli, () =>
    expenseProvider.refresh()
  );
  const trackingCommands = new TrackingCommands(cli, statusBar, () => {
    trackingProvider.refresh();
    dashboardWebviewProvider.refresh();
  });
  const securityCommands = new SecurityCommands(cli, outputChannel);

  // Register all commands

  // Company commands
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.createCompany', async () => {
      await companyCommands.createCompany();
      dashboardWebviewProvider.refresh();
    }),
    vscode.commands.registerCommand('ung.editCompany', () =>
      companyCommands.editCompany()
    )
  );

  // Client commands
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.createClient', async () => {
      await clientCommands.createClient();
      dashboardWebviewProvider.refresh();
    }),
    vscode.commands.registerCommand('ung.viewClient', (clientId) =>
      clientCommands.viewClient(clientId)
    ),
    vscode.commands.registerCommand('ung.editClient', (item) =>
      clientCommands.editClient(item?.itemId)
    ),
    vscode.commands.registerCommand('ung.deleteClient', (item) =>
      clientCommands.deleteClient(item?.itemId)
    ),
    vscode.commands.registerCommand('ung.listClients', () =>
      clientCommands.listClients()
    ),
    vscode.commands.registerCommand('ung.refreshClients', () =>
      clientProvider.refresh()
    )
  );

  // Contract commands
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.createContract', async () => {
      await contractCommands.createContract();
      dashboardWebviewProvider.refresh();
    }),
    vscode.commands.registerCommand('ung.viewContract', (item) =>
      contractCommands.viewContract(item?.itemId)
    ),
    vscode.commands.registerCommand('ung.editContract', (item) =>
      contractCommands.editContract(item?.itemId)
    ),
    vscode.commands.registerCommand('ung.deleteContract', (item) =>
      contractCommands.deleteContract(item?.itemId)
    ),
    vscode.commands.registerCommand('ung.generateContractPDF', (item) =>
      contractCommands.generateContractPDF(item?.itemId)
    ),
    vscode.commands.registerCommand('ung.refreshContracts', () =>
      contractProvider.refresh()
    )
  );

  // Invoice commands
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.createInvoice', () =>
      invoiceCommands.createInvoice()
    ),
    vscode.commands.registerCommand('ung.generateFromTime', () =>
      invoiceCommands.generateFromTime()
    ),
    vscode.commands.registerCommand('ung.viewInvoice', (item) => {
      invoiceCommands.viewInvoice(item?.itemId);
    }),
    vscode.commands.registerCommand('ung.editInvoice', (item) =>
      invoiceCommands.editInvoice(item?.itemId)
    ),
    vscode.commands.registerCommand('ung.deleteInvoice', (item) =>
      invoiceCommands.deleteInvoice(item?.itemId)
    ),
    vscode.commands.registerCommand('ung.exportInvoice', (item) =>
      invoiceCommands.exportInvoice(item?.itemId)
    ),
    vscode.commands.registerCommand('ung.emailInvoice', (item) =>
      invoiceCommands.emailInvoice(item?.itemId)
    ),
    vscode.commands.registerCommand('ung.markInvoicePaid', (item) =>
      invoiceCommands.markAsPaid(item?.itemId)
    ),
    vscode.commands.registerCommand('ung.markInvoiceSent', (item) =>
      invoiceCommands.markAsSent(item?.itemId)
    ),
    vscode.commands.registerCommand('ung.changeInvoiceStatus', (item) =>
      invoiceCommands.changeInvoiceStatus(item?.itemId)
    ),
    vscode.commands.registerCommand('ung.refreshInvoices', () =>
      invoiceProvider.refresh()
    ),
    vscode.commands.registerCommand('ung.generateAllInvoices', () =>
      invoiceCommands.generateAllInvoices()
    ),
    vscode.commands.registerCommand('ung.sendAllInvoices', () =>
      invoiceCommands.sendAllInvoices()
    )
  );

  // Expense commands
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.logExpense', () =>
      expenseCommands.logExpense()
    ),
    vscode.commands.registerCommand('ung.editExpense', (item) =>
      expenseCommands.editExpense(item?.itemId)
    ),
    vscode.commands.registerCommand('ung.deleteExpense', (item) =>
      expenseCommands.deleteExpense(item?.itemId)
    ),
    vscode.commands.registerCommand('ung.viewExpenseReport', () =>
      ExpensePanel.createOrShow(cli)
    ),
    vscode.commands.registerCommand('ung.refreshExpenses', () =>
      expenseProvider.refresh()
    )
  );

  // Tracking commands
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.startTracking', () =>
      trackingCommands.startTracking()
    ),
    vscode.commands.registerCommand('ung.stopTracking', () =>
      trackingCommands.stopTracking()
    ),
    vscode.commands.registerCommand('ung.logTimeManually', () =>
      trackingCommands.logTimeManually()
    ),
    vscode.commands.registerCommand('ung.viewActiveSession', () =>
      trackingCommands.viewActiveSession()
    ),
    vscode.commands.registerCommand('ung.viewTrackingSession', (item) =>
      trackingCommands.viewTrackingSession(item?.itemId)
    ),
    vscode.commands.registerCommand('ung.editTrackingSession', (item) =>
      trackingCommands.editTrackingSession(item?.itemId)
    ),
    vscode.commands.registerCommand('ung.deleteTrackingSession', (item) =>
      trackingCommands.deleteTrackingSession(item?.itemId)
    ),
    vscode.commands.registerCommand('ung.refreshTracking', () =>
      trackingProvider.refresh()
    )
  );

  // Dashboard commands
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.refreshDashboard', () =>
      dashboardWebviewProvider.refresh()
    )
  );

  // Export wizard command
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.openExportWizard', () =>
      ExportPanel.createOrShow(cli)
    )
  );

  // Statistics panel command
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.openStatistics', () =>
      StatisticsPanel.createOrShow(cli)
    )
  );

  // Template Editor panel command
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.openTemplateEditor', () =>
      TemplateEditorPanel.createOrShow(cli)
    )
  );

  // Main Dashboard panel command
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.openDashboard', () =>
      MainDashboardPanel.createOrShow(cli)
    )
  );

  // Search commands
  const searchCommands = new SearchCommands(cli);
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.search', () =>
      searchCommands.universalSearch()
    ),
    vscode.commands.registerCommand('ung.searchInvoices', () =>
      searchCommands.searchInvoices()
    ),
    vscode.commands.registerCommand('ung.searchContracts', () =>
      searchCommands.searchContracts()
    ),
    vscode.commands.registerCommand('ung.searchClients', () =>
      searchCommands.searchClients()
    )
  );

  // Duplicate invoice command
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.duplicateInvoice', (item) =>
      invoiceCommands.duplicateInvoice(item?.itemId)
    )
  );

  // Open invoice PDF command
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.openInvoicePDF', (item) =>
      invoiceCommands.openInvoicePDF(item?.invoiceNum)
    )
  );

  // Quick start tracking - simplified start tracking
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.quickStart', async () => {
      await trackingCommands.startTracking();
      await statusBar.forceUpdate();
    })
  );

  // Toggle tracking command (for status bar)
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.toggleTracking', async () => {
      if (statusBar.getIsTracking()) {
        await trackingCommands.stopTracking();
      } else {
        await trackingCommands.startTracking();
      }
      await statusBar.forceUpdate();
    })
  );

  // Command Center - unified access to all commands
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.commandCenter', () =>
      CommandCenter.show()
    )
  );

  // Quick Actions - most common operations
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.quickActions', () =>
      CommandCenter.showQuickActions()
    )
  );

  // Refresh all command
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.refreshAll', () => {
      invoiceProvider.refresh();
      contractProvider.refresh();
      clientProvider.refresh();
      expenseProvider.refresh();
      trackingProvider.refresh();
      dashboardWebviewProvider.refresh();
      statusBar.forceUpdate();
      vscode.window.showInformationMessage('All views refreshed!');
    })
  );

  // Pomodoro timer command - opens native VS Code panel
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.startPomodoro', async () => {
      PomodoroPanel.createOrShow(context.extensionUri, pomodoroStatusBar);
    })
  );

  // Recurring invoice commands
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.manageRecurring', async () => {
      const actions = [
        { label: '$(list-flat) View Recurring Invoices', action: 'list' },
        { label: '$(play) Generate Due Invoices', action: 'generate' },
        { label: '$(add) Create New (opens terminal)', action: 'add' },
      ];

      const selected = await vscode.window.showQuickPick(actions, {
        placeHolder: 'Manage Recurring Invoices',
        title: 'Recurring Invoices',
      });

      if (!selected) return;

      switch (selected.action) {
        case 'list': {
          const result = await cli.listRecurringInvoices();
          if (result.success && result.stdout) {
            outputChannel.clear();
            outputChannel.appendLine('=== Recurring Invoices ===\n');
            outputChannel.appendLine(result.stdout);
            outputChannel.show();
          } else {
            vscode.window.showInformationMessage(
              'No recurring invoices found. Use terminal to create one: ung recurring add'
            );
          }
          break;
        }
        case 'generate': {
          const confirm = await vscode.window.showWarningMessage(
            'Generate all due recurring invoices?',
            { modal: true },
            'Generate',
            'Preview First'
          );

          if (confirm === 'Preview First') {
            const result = await cli.generateRecurringInvoices({
              dryRun: true,
            });
            outputChannel.clear();
            outputChannel.appendLine('=== Recurring Invoice Preview ===\n');
            outputChannel.appendLine(result.stdout || 'No invoices due');
            outputChannel.show();
          } else if (confirm === 'Generate') {
            await vscode.window.withProgress(
              {
                location: vscode.ProgressLocation.Notification,
                title: 'Generating recurring invoices...',
              },
              async () => {
                const result = await cli.generateRecurringInvoices();
                if (result.success) {
                  vscode.window.showInformationMessage(
                    'Recurring invoices generated!'
                  );
                  invoiceProvider.refresh();
                } else {
                  vscode.window.showErrorMessage(`Failed: ${result.error}`);
                }
              }
            );
          }
          break;
        }
        case 'add': {
          const terminal = vscode.window.createTerminal('UNG Recurring');
          terminal.show();
          terminal.sendText('ung recurring add');
          break;
        }
      }
    })
  );

  // Export data command
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.exportData', async () => {
      const formats = [
        {
          label: '$(file) CSV (Universal)',
          value: 'csv',
          description: 'Works with any spreadsheet',
        },
        {
          label: '$(book) QuickBooks (IIF)',
          value: 'quickbooks',
          description: 'Import into QuickBooks',
        },
        {
          label: '$(json) JSON (Custom)',
          value: 'json',
          description: 'For custom integrations',
        },
      ];

      const format = await vscode.window.showQuickPick(formats, {
        placeHolder: 'Select export format',
        title: 'Export Data',
      });

      if (!format) return;

      const dataTypes = [
        { label: '$(file-text) Invoices', value: 'invoices', picked: true },
        { label: '$(credit-card) Expenses', value: 'expenses', picked: true },
        { label: '$(clock) Time Tracking', value: 'time', picked: true },
      ];

      const selected = await vscode.window.showQuickPick(dataTypes, {
        placeHolder: 'Select data to export',
        canPickMany: true,
        title: 'What to Export',
      });

      if (!selected || selected.length === 0) return;

      await vscode.window.withProgress(
        {
          location: vscode.ProgressLocation.Notification,
          title: 'Exporting data...',
        },
        async () => {
          const result = await cli.exportData(
            format.value,
            selected.map((s) => s.value)
          );
          if (result.success) {
            const openFolder = await vscode.window.showInformationMessage(
              'Export complete! Files saved to ~/.ung/exports',
              'Open Folder',
              'Show Output'
            );
            if (openFolder === 'Open Folder') {
              const homeDir = process.env.HOME || '';
              vscode.commands.executeCommand(
                'revealFileInOS',
                vscode.Uri.file(`${homeDir}/.ung/exports`)
              );
            } else if (openFolder === 'Show Output') {
              outputChannel.clear();
              outputChannel.appendLine(result.stdout || 'Export completed');
              outputChannel.show();
            }
          } else {
            vscode.window.showErrorMessage(`Export failed: ${result.error}`);
          }
        }
      );
    })
  );

  // Backup & Sync command
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.syncData', async () => {
      const actions = [
        {
          label: '$(cloud-upload) Create Backup',
          value: 'backup',
          description: 'Save all data to backup file',
        },
        {
          label: '$(cloud-download) Restore from Backup',
          value: 'restore',
          description: 'Restore data from backup',
        },
        {
          label: '$(list-flat) List Backups',
          value: 'list',
          description: 'Show available backups',
        },
      ];

      const action = await vscode.window.showQuickPick(actions, {
        placeHolder: 'Select sync action',
        title: 'Backup & Sync',
      });

      if (!action) return;

      switch (action.value) {
        case 'backup': {
          await vscode.window.withProgress(
            {
              location: vscode.ProgressLocation.Notification,
              title: 'Creating backup...',
            },
            async () => {
              const result = await cli.createBackup();
              if (result.success) {
                vscode.window.showInformationMessage(
                  'Backup created successfully! Check ~/.ung/backups'
                );
                outputChannel.clear();
                outputChannel.appendLine(result.stdout || 'Backup complete');
                outputChannel.show();
              } else {
                vscode.window.showErrorMessage(
                  `Backup failed: ${result.error}`
                );
              }
            }
          );
          break;
        }
        case 'restore': {
          const terminal = vscode.window.createTerminal('UNG Restore');
          terminal.show();
          terminal.sendText('ung sync restore');
          break;
        }
        case 'list': {
          const result = await cli.listBackups();
          outputChannel.clear();
          outputChannel.appendLine('=== Available Backups ===\n');
          outputChannel.appendLine(result.stdout || 'No backups found');
          outputChannel.show();
          break;
        }
      }
    })
  );

  // Import data command
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.importData', async () => {
      const dataTypes = [
        {
          label: '$(database) SQLite Database',
          value: 'sqlite',
          description: 'Full import from another UNG database',
        },
        {
          label: '$(person) Clients CSV',
          value: 'clients',
          description: 'name, email, address, tax_id',
        },
        {
          label: '$(credit-card) Expenses CSV',
          value: 'expenses',
          description: 'date, description, amount, category',
        },
        {
          label: '$(clock) Time Entries CSV',
          value: 'time',
          description: 'date, client, project, hours',
        },
      ];

      const dataType = await vscode.window.showQuickPick(dataTypes, {
        placeHolder: 'What data do you want to import?',
        title: 'Import Data',
      });

      if (!dataType) return;

      // Handle SQLite import
      if (dataType.value === 'sqlite') {
        const files = await vscode.window.showOpenDialog({
          canSelectMany: false,
          filters: {
            'SQLite Database': ['db', 'sqlite', 'sqlite3'],
            'All Files': ['*'],
          },
          title: 'Select UNG database file to import',
        });

        if (!files || files.length === 0) return;

        const filePath = files[0].fsPath;

        // Ask about encryption
        const isEncrypted = await vscode.window.showQuickPick(
          [
            { label: '$(unlock) Not encrypted', value: false },
            { label: '$(lock) Password protected', value: true },
          ],
          { placeHolder: 'Is the database encrypted?' }
        );

        if (!isEncrypted) return;

        let password: string | undefined;
        if (isEncrypted.value) {
          // Try to get password from secure storage first
          const secureStorage = getSecureStorage();
          password = await secureStorage.getOrPromptPassword({
            title: 'Import Database Password',
            prompt: 'Enter the password for the encrypted database',
            offerToSave: false, // Don't offer to save import passwords
          });
          if (!password) return;
        }

        await vscode.window.withProgress(
          {
            location: vscode.ProgressLocation.Notification,
            title: 'Importing database...',
          },
          async () => {
            const result = await cli.importDatabase(filePath, password);
            outputChannel.clear();
            outputChannel.appendLine('=== Database Import Results ===\n');
            outputChannel.appendLine(result.stdout || 'No output');
            if (result.stderr) {
              outputChannel.appendLine(`\nErrors:\n${result.stderr}`);
            }
            outputChannel.show();

            if (result.success) {
              vscode.window.showInformationMessage(
                'Database import completed!'
              );
              // Refresh all views
              clientProvider.refresh();
              expenseProvider.refresh();
              trackingProvider.refresh();
              invoiceProvider.refresh();
              contractProvider.refresh();
            }
          }
        );
        return;
      }

      // Show file picker for CSV
      const files = await vscode.window.showOpenDialog({
        canSelectMany: false,
        filters: { 'CSV Files': ['csv'], 'All Files': ['*'] },
        title: 'Select CSV file to import',
      });

      if (!files || files.length === 0) return;

      const filePath = files[0].fsPath;

      // Preview first
      const action = await vscode.window.showWarningMessage(
        `Import ${dataType.label} from ${filePath.split('/').pop()}?`,
        { modal: true },
        'Preview First',
        'Import Now'
      );

      if (!action) return;

      const dryRun = action === 'Preview First';

      await vscode.window.withProgress(
        {
          location: vscode.ProgressLocation.Notification,
          title: dryRun ? 'Previewing import...' : 'Importing data...',
        },
        async () => {
          const result = await cli.importData(filePath, dataType.value, dryRun);
          outputChannel.clear();
          outputChannel.appendLine(
            `=== Import ${dryRun ? 'Preview' : 'Results'} ===\n`
          );
          outputChannel.appendLine(result.stdout || 'No output');
          if (result.stderr) {
            outputChannel.appendLine(`\nErrors:\n${result.stderr}`);
          }
          outputChannel.show();

          if (result.success && !dryRun) {
            vscode.window.showInformationMessage('Import completed!');
            // Refresh relevant views
            if (dataType.value === 'clients') {
              clientProvider.refresh();
            } else if (dataType.value === 'expenses') {
              expenseProvider.refresh();
            } else if (dataType.value === 'time') {
              trackingProvider.refresh();
            }
          }
        }
      );
    })
  );

  // Income goal commands
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.setIncomeGoal', async () => {
      const period = await vscode.window.showQuickPick(
        [
          { label: '$(calendar) Monthly', value: 'monthly' },
          { label: '$(milestone) Quarterly', value: 'quarterly' },
          { label: '$(rocket) Yearly', value: 'yearly' },
        ],
        { placeHolder: 'Select goal period', title: 'Set Income Goal' }
      );

      if (!period) return;

      const amount = await vscode.window.showInputBox({
        prompt: `Enter ${period.label.replace('$(calendar) ', '').replace('$(milestone) ', '').replace('$(rocket) ', '')} income target`,
        placeHolder: 'e.g., 10000',
        validateInput: (val) => {
          const num = parseFloat(val);
          if (Number.isNaN(num) || num <= 0) {
            return 'Please enter a positive number';
          }
          return null;
        },
      });

      if (!amount) return;

      const description = await vscode.window.showInputBox({
        prompt: 'Goal description (optional)',
        placeHolder: 'e.g., Q4 revenue target',
      });

      const terminal = vscode.window.createTerminal('UNG Goal');
      terminal.show();
      let cmd = `ung goal set ${amount} -p ${period.value}`;
      if (description) {
        cmd += ` -d "${description}"`;
      }
      terminal.sendText(cmd);
    }),

    vscode.commands.registerCommand('ung.viewGoalStatus', async () => {
      const terminal = vscode.window.createTerminal('UNG Goal');
      terminal.show();
      terminal.sendText('ung goal status');
    })
  );

  // Rate calculator commands
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.calculateRate', async () => {
      const targetType = await vscode.window.showQuickPick(
        [
          { label: '$(symbol-number) Annual income', value: 'annual' },
          { label: '$(calendar) Monthly income', value: 'monthly' },
        ],
        { placeHolder: 'Calculate rate from:', title: 'Rate Calculator' }
      );

      if (!targetType) return;

      const amount = await vscode.window.showInputBox({
        prompt: `Enter target ${targetType.value} income`,
        placeHolder: 'e.g., 100000',
        validateInput: (val) => {
          const num = parseFloat(val);
          if (Number.isNaN(num) || num <= 0) {
            return 'Please enter a positive number';
          }
          return null;
        },
      });

      if (!amount) return;

      const hours = await vscode.window.showInputBox({
        prompt: 'Billable hours per week',
        placeHolder: '40',
        value: '40',
      });

      const terminal = vscode.window.createTerminal('UNG Rate');
      terminal.show();
      const flag = targetType.value === 'annual' ? '--annual' : '--monthly';
      terminal.sendText(
        `ung rate calc ${flag} ${amount} --hours ${hours || '40'}`
      );
    }),

    vscode.commands.registerCommand('ung.analyzeRates', async () => {
      const terminal = vscode.window.createTerminal('UNG Rate');
      terminal.show();
      terminal.sendText('ung rate analyze');
    })
  );

  // Profit dashboard command
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.openProfitDashboard', async () => {
      const terminal = vscode.window.createTerminal('UNG Profit');
      terminal.show();
      terminal.sendText('ung profit');
    })
  );

  // Weekly and monthly report commands
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.weeklyReport', async () => {
      const options = await vscode.window.showQuickPick(
        [
          { label: '$(calendar) This Week', value: '' },
          { label: '$(history) Last Week', value: '--last' },
        ],
        { placeHolder: 'Select report period', title: 'Weekly Report' }
      );

      if (!options) return;

      const terminal = vscode.window.createTerminal('UNG Report');
      terminal.show();
      terminal.sendText(`ung report weekly ${options.value}`.trim());
    }),

    vscode.commands.registerCommand('ung.monthlyReport', async () => {
      const terminal = vscode.window.createTerminal('UNG Report');
      terminal.show();
      terminal.sendText('ung report monthly');
    })
  );

  // Add new commands to quick actions
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.businessInsights', async () => {
      const actions = [
        {
          label: '$(dashboard) Profit Dashboard',
          command: 'ung.openProfitDashboard',
        },
        { label: '$(calendar) Weekly Report', command: 'ung.weeklyReport' },
        { label: '$(history) Monthly Report', command: 'ung.monthlyReport' },
        { label: '$(target) Set Income Goal', command: 'ung.setIncomeGoal' },
        { label: '$(pulse) View Goal Progress', command: 'ung.viewGoalStatus' },
        {
          label: '$(symbol-number) Calculate Rate',
          command: 'ung.calculateRate',
        },
        {
          label: '$(graph-line) Analyze Actual Rates',
          command: 'ung.analyzeRates',
        },
      ];

      const selected = await vscode.window.showQuickPick(actions, {
        placeHolder: 'Business Insights',
        title: 'Financial Tools',
      });

      if (selected) {
        vscode.commands.executeCommand(selected.command);
      }
    })
  );

  // Security commands
  context.subscriptions.push(
    vscode.commands.registerCommand('ung.securityStatus', () =>
      securityCommands.showSecurityStatus()
    ),
    vscode.commands.registerCommand('ung.savePassword', () =>
      securityCommands.savePassword()
    ),
    vscode.commands.registerCommand('ung.clearPassword', () =>
      securityCommands.clearPassword()
    ),
    vscode.commands.registerCommand('ung.securitySettings', () =>
      securityCommands.manageSecuritySettings()
    )
  );
}

/**
 * Extension deactivation
 */
export function deactivate() {
  console.log('UNG extension is now deactivated');
}
