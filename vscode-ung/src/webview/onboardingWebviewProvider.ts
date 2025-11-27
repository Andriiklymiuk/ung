import * as vscode from 'vscode';

/**
 * Onboarding webview provider for the sidebar
 * Shows a polished welcome experience when CLI is not installed or not initialized
 */
export class OnboardingWebviewProvider implements vscode.WebviewViewProvider {
  public static readonly viewType = 'ungOnboarding';

  private _view?: vscode.WebviewView;
  private _extensionUri: vscode.Uri;
  private _state: 'not-installed' | 'not-initialized' | 'ready';

  constructor(
    extensionUri: vscode.Uri,
    private readonly checkCliInstalled: () => Promise<boolean>,
    private readonly checkIsInitialized: () => Promise<boolean>
  ) {
    this._extensionUri = extensionUri;
    this._state = 'not-installed';
  }

  public resolveWebviewView(
    webviewView: vscode.WebviewView,
    _context: vscode.WebviewViewResolveContext,
    _token: vscode.CancellationToken
  ) {
    this._view = webviewView;

    webviewView.webview.options = {
      enableScripts: true,
      localResourceRoots: [this._extensionUri],
    };

    // Handle messages from the webview
    webviewView.webview.onDidReceiveMessage(async (message) => {
      switch (message.command) {
        case 'installHomebrew':
          vscode.commands.executeCommand('ung.installViaHomebrew');
          break;
        case 'installScoop':
          vscode.commands.executeCommand('ung.installViaScoop');
          break;
        case 'installGo':
          vscode.commands.executeCommand('ung.installViaGo');
          break;
        case 'downloadBinary':
          vscode.commands.executeCommand(
            'ung.downloadBinary',
            message.platform
          );
          break;
        case 'initGlobal':
          vscode.commands.executeCommand('ung.initializeGlobal');
          break;
        case 'initLocal':
          vscode.commands.executeCommand('ung.initializeLocal');
          break;
        case 'openDocs':
          vscode.commands.executeCommand('ung.openDocs');
          break;
        case 'recheckCli':
          vscode.commands.executeCommand('ung.recheckCli');
          break;
        case 'refresh':
          await this.refresh();
          break;
      }
    });

    this.refresh();
  }

  public async refresh(): Promise<void> {
    if (!this._view) {
      return;
    }

    // Check current state
    const cliInstalled = await this.checkCliInstalled();
    if (!cliInstalled) {
      this._state = 'not-installed';
    } else {
      const isInitialized = await this.checkIsInitialized();
      this._state = isInitialized ? 'ready' : 'not-initialized';
    }

    this._view.webview.html = this._getHtmlForWebview();
  }

  private _getHtmlForWebview(): string {
    const platform = process.platform;

    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to UNG</title>
    <style>
        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            font-family: var(--vscode-font-family);
            color: var(--vscode-foreground);
            background-color: var(--vscode-sideBar-background);
            padding: 16px;
            line-height: 1.5;
        }

        .container {
            max-width: 100%;
        }

        /* Header Section */
        .header {
            text-align: center;
            padding: 24px 0;
            border-bottom: 1px solid var(--vscode-panel-border);
            margin-bottom: 20px;
        }

        .logo {
            font-size: 32px;
            margin-bottom: 8px;
        }

        .title {
            font-size: 20px;
            font-weight: 600;
            color: var(--vscode-foreground);
            margin-bottom: 4px;
        }

        .subtitle {
            font-size: 13px;
            color: var(--vscode-descriptionForeground);
        }

        /* Section Styles */
        .section {
            margin-bottom: 20px;
        }

        .section-title {
            font-size: 13px;
            font-weight: 600;
            color: var(--vscode-foreground);
            margin-bottom: 12px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }

        .section-description {
            font-size: 12px;
            color: var(--vscode-descriptionForeground);
            margin-bottom: 12px;
        }

        /* Button Styles */
        .btn {
            display: flex;
            align-items: center;
            width: 100%;
            padding: 10px 12px;
            margin-bottom: 8px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 13px;
            font-family: var(--vscode-font-family);
            transition: background-color 0.15s ease;
            text-align: left;
        }

        .btn-primary {
            background-color: var(--vscode-button-background);
            color: var(--vscode-button-foreground);
        }

        .btn-primary:hover {
            background-color: var(--vscode-button-hoverBackground);
        }

        .btn-secondary {
            background-color: var(--vscode-button-secondaryBackground);
            color: var(--vscode-button-secondaryForeground);
        }

        .btn-secondary:hover {
            background-color: var(--vscode-button-secondaryHoverBackground);
        }

        .btn-icon {
            margin-right: 10px;
            font-size: 16px;
            width: 20px;
            text-align: center;
        }

        .btn-content {
            flex: 1;
        }

        .btn-label {
            display: block;
            font-weight: 500;
        }

        .btn-description {
            display: block;
            font-size: 11px;
            color: var(--vscode-descriptionForeground);
            margin-top: 2px;
        }

        .btn-badge {
            font-size: 10px;
            padding: 2px 6px;
            background-color: var(--vscode-badge-background);
            color: var(--vscode-badge-foreground);
            border-radius: 10px;
            margin-left: 8px;
        }

        /* Feature List */
        .features {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 8px;
            margin-top: 12px;
        }

        .feature {
            display: flex;
            align-items: center;
            padding: 8px;
            background-color: var(--vscode-input-background);
            border-radius: 4px;
            font-size: 11px;
        }

        .feature-icon {
            margin-right: 8px;
            opacity: 0.8;
        }

        /* Info Box */
        .info-box {
            padding: 12px;
            background-color: var(--vscode-textBlockQuote-background);
            border-left: 3px solid var(--vscode-textLink-foreground);
            border-radius: 0 4px 4px 0;
            margin: 16px 0;
        }

        .info-box p {
            font-size: 12px;
            color: var(--vscode-descriptionForeground);
        }

        /* Divider */
        .divider {
            height: 1px;
            background-color: var(--vscode-panel-border);
            margin: 20px 0;
        }

        /* Link Styles */
        .link {
            color: var(--vscode-textLink-foreground);
            text-decoration: none;
            cursor: pointer;
            font-size: 12px;
        }

        .link:hover {
            text-decoration: underline;
        }

        /* Warning Box */
        .warning-box {
            display: flex;
            align-items: center;
            padding: 12px;
            background-color: var(--vscode-inputValidation-warningBackground);
            border: 1px solid var(--vscode-inputValidation-warningBorder);
            border-radius: 4px;
            margin-bottom: 16px;
        }

        .warning-icon {
            margin-right: 10px;
            font-size: 18px;
        }

        .warning-text {
            font-size: 12px;
        }

        /* Collapsible sections */
        .collapsible {
            margin-bottom: 8px;
        }

        .collapsible-header {
            display: flex;
            align-items: center;
            padding: 8px 12px;
            background-color: var(--vscode-input-background);
            border-radius: 4px;
            cursor: pointer;
            font-size: 13px;
            font-weight: 500;
        }

        .collapsible-header:hover {
            background-color: var(--vscode-list-hoverBackground);
        }

        .collapsible-icon {
            margin-right: 8px;
            transition: transform 0.2s ease;
        }

        .collapsible-content {
            padding: 12px;
            display: none;
        }

        .collapsible.open .collapsible-content {
            display: block;
        }

        .collapsible.open .collapsible-icon {
            transform: rotate(90deg);
        }
    </style>
</head>
<body>
    <div class="container">
        ${this._state === 'not-installed' ? this._getNotInstalledHtml(platform) : ''}
        ${this._state === 'not-initialized' ? this._getNotInitializedHtml() : ''}
        ${this._state === 'ready' ? this._getReadyHtml() : ''}
    </div>

    <script>
        const vscode = acquireVsCodeApi();

        // Handle button clicks
        document.querySelectorAll('[data-command]').forEach(btn => {
            btn.addEventListener('click', () => {
                const command = btn.getAttribute('data-command');
                const platform = btn.getAttribute('data-platform');
                vscode.postMessage({ command, platform });
            });
        });

        // Handle collapsible sections
        document.querySelectorAll('.collapsible-header').forEach(header => {
            header.addEventListener('click', () => {
                header.parentElement.classList.toggle('open');
            });
        });
    </script>
</body>
</html>`;
  }

  private _getNotInstalledHtml(platform: string): string {
    const isMac = platform === 'darwin';
    const isWindows = platform === 'win32';
    const isLinux = platform === 'linux';

    return `
        <div class="header">
            <div class="logo">🚀</div>
            <h1 class="title">Welcome to UNG!</h1>
            <p class="subtitle">Your all-in-one freelance business toolkit</p>
        </div>

        <div class="warning-box">
            <span class="warning-icon">⚠️</span>
            <span class="warning-text">UNG CLI is required to use this extension</span>
        </div>

        <div class="section">
            <h2 class="section-title">Install UNG CLI</h2>
            <p class="section-description">Choose your preferred installation method:</p>

            ${
              isMac || isLinux
                ? `
            <button class="btn btn-primary" data-command="installHomebrew">
                <span class="btn-icon">🍺</span>
                <span class="btn-content">
                    <span class="btn-label">Install via Homebrew</span>
                    <span class="btn-description">brew tap Andriiklymiuk/tools && brew install ung</span>
                </span>
                <span class="btn-badge">Recommended</span>
            </button>
            `
                : ''
            }

            ${
              isWindows
                ? `
            <button class="btn btn-primary" data-command="installScoop">
                <span class="btn-icon">🥄</span>
                <span class="btn-content">
                    <span class="btn-label">Install via Scoop</span>
                    <span class="btn-description">scoop bucket add ung && scoop install ung</span>
                </span>
                <span class="btn-badge">Recommended</span>
            </button>
            `
                : ''
            }

            <button class="btn btn-secondary" data-command="downloadBinary" data-platform="${platform}">
                <span class="btn-icon">📦</span>
                <span class="btn-content">
                    <span class="btn-label">Download Binary</span>
                    <span class="btn-description">Direct download for ${isMac ? 'macOS' : isWindows ? 'Windows' : 'Linux'}</span>
                </span>
            </button>

            <button class="btn btn-secondary" data-command="installGo">
                <span class="btn-icon">🐹</span>
                <span class="btn-content">
                    <span class="btn-label">Install via Go</span>
                    <span class="btn-description">go install github.com/Andriiklymiuk/ung@latest</span>
                </span>
            </button>
        </div>

        <div class="divider"></div>

        <div class="section">
            <div class="collapsible">
                <div class="collapsible-header">
                    <span class="collapsible-icon">▶</span>
                    <span>✨ Features</span>
                </div>
                <div class="collapsible-content">
                    <div class="features">
                        <div class="feature"><span class="feature-icon">📄</span> Invoice Generation</div>
                        <div class="feature"><span class="feature-icon">⏱️</span> Time Tracking</div>
                        <div class="feature"><span class="feature-icon">👥</span> Client Management</div>
                        <div class="feature"><span class="feature-icon">📝</span> Contracts</div>
                        <div class="feature"><span class="feature-icon">💳</span> Expense Tracking</div>
                        <div class="feature"><span class="feature-icon">📊</span> Revenue Reports</div>
                        <div class="feature"><span class="feature-icon">🌍</span> Multi-Currency</div>
                        <div class="feature"><span class="feature-icon">📧</span> Email Integration</div>
                    </div>
                </div>
            </div>
        </div>

        <div class="divider"></div>

        <div style="text-align: center;">
            <a class="link" data-command="openDocs">📚 View Documentation</a>
            <span style="margin: 0 8px; color: var(--vscode-descriptionForeground);">•</span>
            <a class="link" data-command="recheckCli">🔄 Recheck Installation</a>
        </div>
    `;
  }

  private _getNotInitializedHtml(): string {
    return `
        <div class="header">
            <div class="logo">🎉</div>
            <h1 class="title">UNG CLI Installed!</h1>
            <p class="subtitle">Let's set up your workspace</p>
        </div>

        <div class="section">
            <h2 class="section-title">Choose Your Setup</h2>
            <p class="section-description">Select where to store your billing data:</p>

            <button class="btn btn-primary" data-command="initGlobal">
                <span class="btn-icon">🏠</span>
                <span class="btn-content">
                    <span class="btn-label">Global Setup</span>
                    <span class="btn-description">Store data in ~/.ung/ - Perfect for personal use across all projects</span>
                </span>
                <span class="btn-badge">Recommended</span>
            </button>

            <button class="btn btn-secondary" data-command="initLocal">
                <span class="btn-icon">📁</span>
                <span class="btn-content">
                    <span class="btn-label">Project-Specific Setup</span>
                    <span class="btn-description">Create .ung/ in this workspace - Great for project-specific billing</span>
                </span>
            </button>
        </div>

        <div class="info-box">
            <p><strong>💡 Not sure which to choose?</strong><br>
            Global setup is best for most freelancers. Use project-specific if you need separate billing per project.</p>
        </div>

        <div class="divider"></div>

        <div class="section">
            <div class="collapsible open">
                <div class="collapsible-header">
                    <span class="collapsible-icon">▶</span>
                    <span>🎯 What You Can Do</span>
                </div>
                <div class="collapsible-content">
                    <div class="features">
                        <div class="feature"><span class="feature-icon">⏱️</span> Track Time</div>
                        <div class="feature"><span class="feature-icon">📄</span> Create Invoices</div>
                        <div class="feature"><span class="feature-icon">👥</span> Manage Clients</div>
                        <div class="feature"><span class="feature-icon">📝</span> Handle Contracts</div>
                        <div class="feature"><span class="feature-icon">💳</span> Track Expenses</div>
                        <div class="feature"><span class="feature-icon">📊</span> View Reports</div>
                    </div>
                </div>
            </div>

            <div class="collapsible">
                <div class="collapsible-header">
                    <span class="collapsible-icon">▶</span>
                    <span>💎 Why UNG?</span>
                </div>
                <div class="collapsible-content">
                    <div class="features">
                        <div class="feature"><span class="feature-icon">🔒</span> Privacy First</div>
                        <div class="feature"><span class="feature-icon">📴</span> Works Offline</div>
                        <div class="feature"><span class="feature-icon">🌍</span> Multi-Currency</div>
                        <div class="feature"><span class="feature-icon">💻</span> VS Code Native</div>
                        <div class="feature"><span class="feature-icon">🔓</span> Open Source</div>
                        <div class="feature"><span class="feature-icon">⚡</span> Fast CLI</div>
                    </div>
                </div>
            </div>
        </div>

        <div class="divider"></div>

        <div style="text-align: center;">
            <a class="link" data-command="openDocs">📚 View Documentation</a>
        </div>
    `;
  }

  private _getReadyHtml(): string {
    return `
        <div class="header">
            <div class="logo">✅</div>
            <h1 class="title">All Set!</h1>
            <p class="subtitle">UNG is ready to use</p>
        </div>

        <div class="info-box">
            <p>You're all set! Check out the dashboard and other views in the sidebar to get started.</p>
        </div>
    `;
  }
}
