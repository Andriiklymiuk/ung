import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';

/**
 * Export wizard webview panel
 */
export class ExportPanel {
  public static currentPanel: ExportPanel | undefined;
  private readonly panel: vscode.WebviewPanel;
  private disposables: vscode.Disposable[] = [];

  private cli: UngCli;

  private constructor(panel: vscode.WebviewPanel, cli: UngCli) {
    this.cli = cli;
    this.panel = panel;

    // Set the webview's HTML content
    this.update();

    // Handle messages from the webview
    this.panel.webview.onDidReceiveMessage(
      (message) => {
        switch (message.command) {
          case 'export':
            this.performExport(message.options);
            break;
        }
      },
      null,
      this.disposables
    );

    // Handle panel disposed
    this.panel.onDidDispose(() => this.dispose(), null, this.disposables);
  }

  /**
   * Create or show export panel
   */
  public static createOrShow(cli: UngCli) {
    const column = vscode.window.activeTextEditor
      ? vscode.window.activeTextEditor.viewColumn
      : undefined;

    // If we already have a panel, show it
    if (ExportPanel.currentPanel) {
      ExportPanel.currentPanel.panel.reveal(column);
      return;
    }

    // Otherwise, create a new panel
    const panel = vscode.window.createWebviewPanel(
      'ungExport',
      'Export Wizard',
      column || vscode.ViewColumn.One,
      {
        enableScripts: true,
        retainContextWhenHidden: true,
      }
    );

    ExportPanel.currentPanel = new ExportPanel(panel, cli);
  }

  /**
   * Update webview content
   */
  private update() {
    this.panel.webview.html = this.getHtmlForWebview();
  }

  /**
   * Get CLI instance for export operations
   */
  protected getCli(): UngCli {
    return this.cli;
  }

  /**
   * Perform export based on selected options
   */
  private async performExport(options: {
    format: string;
    type: string;
    dateFrom: string;
    dateTo: string;
  }) {
    const cli = this.getCli();

    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: `Exporting ${options.type} to ${options.format.toUpperCase()}...`,
        cancellable: false,
      },
      async () => {
        // Build data types array based on selection
        const dataTypes: string[] = [];
        switch (options.type) {
          case 'invoices':
            dataTypes.push('invoices');
            break;
          case 'contracts':
            dataTypes.push('contracts');
            break;
          case 'expenses':
            dataTypes.push('expenses');
            break;
          case 'tracking':
            dataTypes.push('time');
            break;
          case 'all':
            dataTypes.push('invoices', 'expenses', 'time');
            break;
        }

        // Determine year from dates if provided
        let year: number | undefined;
        if (options.dateFrom) {
          const dateFromYear = new Date(options.dateFrom).getFullYear();
          year = dateFromYear;
        }

        const result = await cli.exportData(options.format, dataTypes, {
          year,
        });

        if (result.success) {
          // Parse output to find file path
          const output = result.stdout || '';
          const pathMatch =
            output.match(/exported to:\s*(.+\.(?:csv|pdf))/i) ||
            output.match(/saved to:\s*(.+\.(?:csv|pdf))/i) ||
            output.match(/(\/.+\.(?:csv|pdf))/);

          if (pathMatch) {
            const action = await vscode.window.showInformationMessage(
              `Export completed: ${pathMatch[1]}`,
              'Open File',
              'Show in Folder'
            );

            if (action === 'Open File') {
              await vscode.env.openExternal(vscode.Uri.file(pathMatch[1]));
            } else if (action === 'Show in Folder') {
              await vscode.commands.executeCommand(
                'revealFileInOS',
                vscode.Uri.file(pathMatch[1])
              );
            }
          } else {
            vscode.window.showInformationMessage(
              `Export completed successfully!\n${output || 'Data exported.'}`
            );
          }
        } else {
          vscode.window.showErrorMessage(
            `Export failed: ${result.error || 'Unknown error'}`
          );
        }
      }
    );
  }

  /**
   * Get HTML content for webview
   */
  private getHtmlForWebview(): string {
    return `<!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <title>Export Wizard</title>
            <style>
                body {
                    font-family: var(--vscode-font-family);
                    color: var(--vscode-foreground);
                    padding: 20px;
                }
                .container {
                    max-width: 600px;
                    margin: 0 auto;
                }
                h1 {
                    color: var(--vscode-titleBar-activeForeground);
                    border-bottom: 2px solid var(--vscode-panel-border);
                    padding-bottom: 10px;
                }
                .form-group {
                    margin: 20px 0;
                }
                label {
                    display: block;
                    margin-bottom: 5px;
                    font-weight: bold;
                }
                select, input {
                    width: 100%;
                    padding: 8px;
                    background-color: var(--vscode-input-background);
                    color: var(--vscode-input-foreground);
                    border: 1px solid var(--vscode-input-border);
                    border-radius: 2px;
                }
                button {
                    background-color: var(--vscode-button-background);
                    color: var(--vscode-button-foreground);
                    border: none;
                    padding: 10px 20px;
                    cursor: pointer;
                    border-radius: 2px;
                    width: 100%;
                    margin-top: 20px;
                }
                button:hover {
                    background-color: var(--vscode-button-hoverBackground);
                }
            </style>
        </head>
        <body>
            <div class="container">
                <h1>Export Wizard</h1>

                <form id="exportForm">
                    <div class="form-group">
                        <label for="format">Export Format:</label>
                        <select id="format">
                            <option value="pdf">PDF</option>
                            <option value="csv">CSV</option>
                        </select>
                    </div>

                    <div class="form-group">
                        <label for="type">Export Type:</label>
                        <select id="type">
                            <option value="invoices">Invoices</option>
                            <option value="expenses">Expenses</option>
                            <option value="tracking">Time Tracking</option>
                            <option value="all">All Data</option>
                        </select>
                    </div>

                    <div class="form-group">
                        <label for="dateFrom">Date From (optional):</label>
                        <input type="date" id="dateFrom">
                    </div>

                    <div class="form-group">
                        <label for="dateTo">Date To (optional):</label>
                        <input type="date" id="dateTo">
                    </div>

                    <button type="submit">Export Data</button>
                </form>

                <div style="margin-top: 30px; padding: 16px; background: var(--vscode-textBlockQuote-background); border-radius: 4px; border-left: 4px solid var(--vscode-button-background);">
                    <p style="margin: 0; color: var(--vscode-descriptionForeground); font-size: 13px;">
                        <strong>Tip:</strong> CSV exports are great for spreadsheets and accounting software.
                        Date filters help narrow down the export to a specific period.
                    </p>
                </div>
            </div>

            <script>
                const vscode = acquireVsCodeApi();

                document.getElementById('exportForm').addEventListener('submit', (e) => {
                    e.preventDefault();

                    const options = {
                        format: document.getElementById('format').value,
                        type: document.getElementById('type').value,
                        dateFrom: document.getElementById('dateFrom').value,
                        dateTo: document.getElementById('dateTo').value
                    };

                    vscode.postMessage({
                        command: 'export',
                        options
                    });
                });
            </script>
        </body>
        </html>`;
  }

  /**
   * Dispose resources
   */
  public dispose() {
    ExportPanel.currentPanel = undefined;

    this.panel.dispose();

    while (this.disposables.length) {
      const disposable = this.disposables.pop();
      if (disposable) {
        disposable.dispose();
      }
    }
  }
}
