import * as vscode from 'vscode';

/**
 * Onboarding webview provider for the sidebar
 * Shows a polished welcome experience when CLI is not installed or not initialized
 */
export class OnboardingWebviewProvider implements vscode.WebviewViewProvider {
  public static readonly viewType = 'ungOnboarding';

  private _view?: vscode.WebviewView;
  private _extensionUri: vscode.Uri;
  private _state: 'loading' | 'not-installed' | 'not-initialized' | 'ready';

  constructor(
    extensionUri: vscode.Uri,
    private readonly checkCliInstalled: () => Promise<boolean>,
    private readonly checkIsInitialized: () => Promise<boolean>
  ) {
    this._extensionUri = extensionUri;
    this._state = 'loading';
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

    // Set initial HTML immediately to avoid blank screen
    // Shows loading state while async checks run
    webviewView.webview.html = this._getHtmlForWebview();

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

    // Run async refresh to update state based on actual CLI checks
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
            padding: 20px;
            line-height: 1.6;
        }

        .container {
            max-width: 100%;
        }

        /* Loading Spinner */
        .loading-container {
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            padding: 60px 20px;
        }

        .spinner {
            width: 40px;
            height: 40px;
            border: 3px solid var(--vscode-input-background);
            border-top: 3px solid var(--vscode-textLink-foreground);
            border-radius: 50%;
            animation: spin 1s linear infinite;
        }

        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }

        .loading-text {
            margin-top: 16px;
            color: var(--vscode-descriptionForeground);
            font-size: 13px;
        }

        /* Header Section */
        .header {
            text-align: center;
            padding: 20px 0 24px;
            margin-bottom: 24px;
        }

        .logo-container {
            width: 64px;
            height: 64px;
            margin: 0 auto 16px;
            background: linear-gradient(135deg, var(--vscode-textLink-foreground), var(--vscode-textLink-activeForeground, var(--vscode-textLink-foreground)));
            border-radius: 16px;
            display: flex;
            align-items: center;
            justify-content: center;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        }

        .logo-icon {
            font-size: 28px;
            filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.2));
        }

        .title {
            font-size: 22px;
            font-weight: 600;
            color: var(--vscode-foreground);
            margin-bottom: 6px;
            letter-spacing: -0.3px;
        }

        .subtitle {
            font-size: 13px;
            color: var(--vscode-descriptionForeground);
            max-width: 280px;
            margin: 0 auto;
        }

        /* Status Badge */
        .status-badge {
            display: inline-flex;
            align-items: center;
            gap: 6px;
            padding: 6px 12px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: 500;
            margin-bottom: 20px;
        }

        .status-badge.warning {
            background-color: color-mix(in srgb, var(--vscode-editorWarning-foreground) 15%, transparent);
            color: var(--vscode-editorWarning-foreground);
            border: 1px solid color-mix(in srgb, var(--vscode-editorWarning-foreground) 30%, transparent);
        }

        .status-badge.success {
            background-color: color-mix(in srgb, var(--vscode-charts-green, #4caf50) 15%, transparent);
            color: var(--vscode-charts-green, #4caf50);
            border: 1px solid color-mix(in srgb, var(--vscode-charts-green, #4caf50) 30%, transparent);
        }

        .status-badge.info {
            background-color: color-mix(in srgb, var(--vscode-textLink-foreground) 15%, transparent);
            color: var(--vscode-textLink-foreground);
            border: 1px solid color-mix(in srgb, var(--vscode-textLink-foreground) 30%, transparent);
        }

        /* Section Styles */
        .section {
            margin-bottom: 24px;
        }

        .section-header {
            display: flex;
            align-items: center;
            gap: 8px;
            margin-bottom: 12px;
        }

        .section-icon {
            font-size: 14px;
            opacity: 0.8;
        }

        .section-title {
            font-size: 11px;
            font-weight: 600;
            color: var(--vscode-descriptionForeground);
            text-transform: uppercase;
            letter-spacing: 0.8px;
        }

        .section-description {
            font-size: 13px;
            color: var(--vscode-descriptionForeground);
            margin-bottom: 14px;
        }

        /* Button Styles */
        .btn {
            display: flex;
            align-items: center;
            width: 100%;
            padding: 12px 14px;
            margin-bottom: 10px;
            border: 1px solid transparent;
            border-radius: 8px;
            cursor: pointer;
            font-size: 13px;
            font-family: var(--vscode-font-family);
            transition: all 0.15s ease;
            text-align: left;
            position: relative;
            overflow: hidden;
        }

        .btn::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: linear-gradient(135deg, rgba(255,255,255,0.1), transparent);
            opacity: 0;
            transition: opacity 0.15s ease;
        }

        .btn:hover::before {
            opacity: 1;
        }

        .btn-primary {
            background-color: var(--vscode-button-background);
            color: var(--vscode-button-foreground);
            border-color: var(--vscode-button-background);
        }

        .btn-primary:hover {
            background-color: var(--vscode-button-hoverBackground);
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        }

        .btn-secondary {
            background-color: var(--vscode-input-background);
            color: var(--vscode-foreground);
            border-color: var(--vscode-input-border, var(--vscode-input-background));
        }

        .btn-secondary:hover {
            background-color: var(--vscode-list-hoverBackground);
            border-color: var(--vscode-focusBorder);
        }

        .btn-icon {
            width: 32px;
            height: 32px;
            display: flex;
            align-items: center;
            justify-content: center;
            margin-right: 12px;
            border-radius: 6px;
            background-color: rgba(255, 255, 255, 0.1);
            font-size: 16px;
            flex-shrink: 0;
        }

        .btn-secondary .btn-icon {
            background-color: var(--vscode-badge-background);
        }

        .btn-content {
            flex: 1;
            min-width: 0;
        }

        .btn-label {
            display: block;
            font-weight: 500;
            margin-bottom: 2px;
        }

        .btn-description {
            display: block;
            font-size: 11px;
            color: var(--vscode-descriptionForeground);
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .btn-primary .btn-description {
            color: rgba(255, 255, 255, 0.75);
        }

        .btn-badge {
            font-size: 10px;
            padding: 3px 8px;
            background-color: rgba(255, 255, 255, 0.2);
            color: inherit;
            border-radius: 12px;
            font-weight: 500;
            flex-shrink: 0;
            margin-left: 8px;
        }

        .btn-secondary .btn-badge {
            background-color: var(--vscode-badge-background);
            color: var(--vscode-badge-foreground);
        }

        /* Feature Cards */
        .features-grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 8px;
            margin-top: 16px;
        }

        .feature-card {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 10px 12px;
            background-color: var(--vscode-input-background);
            border-radius: 6px;
            font-size: 12px;
            transition: background-color 0.15s ease;
        }

        .feature-card:hover {
            background-color: var(--vscode-list-hoverBackground);
        }

        .feature-icon {
            font-size: 14px;
            opacity: 0.85;
        }

        .feature-text {
            color: var(--vscode-foreground);
            font-weight: 450;
        }

        /* Info Box */
        .info-box {
            display: flex;
            gap: 12px;
            padding: 14px;
            background-color: color-mix(in srgb, var(--vscode-textLink-foreground) 8%, var(--vscode-input-background));
            border-left: 3px solid var(--vscode-textLink-foreground);
            border-radius: 0 8px 8px 0;
            margin: 20px 0;
        }

        .info-box-icon {
            font-size: 18px;
            flex-shrink: 0;
        }

        .info-box-content {
            flex: 1;
        }

        .info-box-title {
            font-size: 12px;
            font-weight: 600;
            color: var(--vscode-foreground);
            margin-bottom: 4px;
        }

        .info-box-text {
            font-size: 12px;
            color: var(--vscode-descriptionForeground);
            line-height: 1.5;
        }

        /* Divider */
        .divider {
            height: 1px;
            background: linear-gradient(to right, transparent, var(--vscode-panel-border), transparent);
            margin: 24px 0;
        }

        /* Link Styles */
        .footer-links {
            display: flex;
            justify-content: center;
            gap: 16px;
            flex-wrap: wrap;
        }

        .link {
            display: inline-flex;
            align-items: center;
            gap: 6px;
            color: var(--vscode-textLink-foreground);
            text-decoration: none;
            cursor: pointer;
            font-size: 12px;
            padding: 6px 10px;
            border-radius: 6px;
            transition: all 0.15s ease;
        }

        .link:hover {
            background-color: color-mix(in srgb, var(--vscode-textLink-foreground) 10%, transparent);
        }

        .link-icon {
            font-size: 14px;
        }

        /* Collapsible sections */
        .collapsible {
            margin-bottom: 8px;
            border-radius: 8px;
            overflow: hidden;
            background-color: var(--vscode-input-background);
        }

        .collapsible-header {
            display: flex;
            align-items: center;
            padding: 12px 14px;
            cursor: pointer;
            font-size: 13px;
            font-weight: 500;
            transition: background-color 0.15s ease;
        }

        .collapsible-header:hover {
            background-color: var(--vscode-list-hoverBackground);
        }

        .collapsible-icon {
            margin-right: 10px;
            font-size: 10px;
            transition: transform 0.2s ease;
            color: var(--vscode-descriptionForeground);
        }

        .collapsible-title {
            flex: 1;
        }

        .collapsible-content {
            padding: 0 14px 14px;
            display: none;
        }

        .collapsible.open .collapsible-content {
            display: block;
        }

        .collapsible.open .collapsible-icon {
            transform: rotate(90deg);
        }

        /* Highlight Points */
        .highlights {
            display: flex;
            flex-wrap: wrap;
            gap: 8px;
            margin-top: 12px;
        }

        .highlight-tag {
            display: inline-flex;
            align-items: center;
            gap: 4px;
            padding: 4px 10px;
            background-color: var(--vscode-input-background);
            border-radius: 12px;
            font-size: 11px;
            color: var(--vscode-descriptionForeground);
        }

        .highlight-icon {
            font-size: 12px;
        }

        /* Success State */
        .success-container {
            text-align: center;
            padding: 20px 0;
        }

        .success-icon {
            width: 80px;
            height: 80px;
            margin: 0 auto 20px;
            background: linear-gradient(135deg, var(--vscode-charts-green, #4caf50), color-mix(in srgb, var(--vscode-charts-green, #4caf50) 70%, black));
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 36px;
            box-shadow: 0 8px 24px rgba(76, 175, 80, 0.25);
        }

        .success-title {
            font-size: 20px;
            font-weight: 600;
            margin-bottom: 8px;
        }

        .success-message {
            color: var(--vscode-descriptionForeground);
            font-size: 13px;
            margin-bottom: 20px;
        }
    </style>
</head>
<body>
    <div class="container">
        ${this._state === 'loading' ? this._getLoadingHtml() : ''}
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

  private _getLoadingHtml(): string {
    return `
        <div class="loading-container">
            <div class="spinner"></div>
            <div class="loading-text">Checking UNG CLI status...</div>
        </div>
    `;
  }

  private _getNotInstalledHtml(platform: string): string {
    const isMac = platform === 'darwin';
    const isWindows = platform === 'win32';
    const isLinux = platform === 'linux';

    return `
        <div class="header">
            <div class="logo-container">
                <span class="logo-icon">U</span>
            </div>
            <h1 class="title">Welcome to UNG</h1>
            <p class="subtitle">Professional billing and time tracking for freelancers</p>
        </div>

        <div style="text-align: center;">
            <span class="status-badge warning">
                <span>CLI Required</span>
            </span>
        </div>

        <div class="section">
            <div class="section-header">
                <span class="section-icon">1</span>
                <h2 class="section-title">Install UNG CLI</h2>
            </div>

            ${
              isMac || isLinux
                ? `
            <button class="btn btn-primary" data-command="installHomebrew">
                <span class="btn-icon">🍺</span>
                <span class="btn-content">
                    <span class="btn-label">Install via Homebrew</span>
                    <span class="btn-description">brew install ung</span>
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
                <span class="btn-icon">🪣</span>
                <span class="btn-content">
                    <span class="btn-label">Install via Scoop</span>
                    <span class="btn-description">scoop install ung</span>
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
                    <span class="collapsible-title">What you'll get</span>
                </div>
                <div class="collapsible-content">
                    <div class="features-grid">
                        <div class="feature-card">
                            <span class="feature-icon">📄</span>
                            <span class="feature-text">Invoices</span>
                        </div>
                        <div class="feature-card">
                            <span class="feature-icon">⏱️</span>
                            <span class="feature-text">Time Tracking</span>
                        </div>
                        <div class="feature-card">
                            <span class="feature-icon">👥</span>
                            <span class="feature-text">Clients</span>
                        </div>
                        <div class="feature-card">
                            <span class="feature-icon">📝</span>
                            <span class="feature-text">Contracts</span>
                        </div>
                        <div class="feature-card">
                            <span class="feature-icon">💳</span>
                            <span class="feature-text">Expenses</span>
                        </div>
                        <div class="feature-card">
                            <span class="feature-icon">📊</span>
                            <span class="feature-text">Reports</span>
                        </div>
                    </div>
                    <div class="highlights">
                        <span class="highlight-tag"><span class="highlight-icon">🔒</span> Privacy First</span>
                        <span class="highlight-tag"><span class="highlight-icon">📴</span> Offline</span>
                        <span class="highlight-tag"><span class="highlight-icon">🌍</span> Multi-Currency</span>
                        <span class="highlight-tag"><span class="highlight-icon">⚡</span> Fast</span>
                    </div>
                </div>
            </div>
        </div>

        <div class="divider"></div>

        <div class="footer-links">
            <a class="link" data-command="openDocs">
                <span class="link-icon">📚</span>
                Documentation
            </a>
            <a class="link" data-command="recheckCli">
                <span class="link-icon">🔄</span>
                Recheck
            </a>
        </div>
    `;
  }

  private _getNotInitializedHtml(): string {
    return `
        <div class="header">
            <div class="logo-container">
                <span class="logo-icon">U</span>
            </div>
            <h1 class="title">Almost There!</h1>
            <p class="subtitle">UNG CLI detected. Let's set up your workspace.</p>
        </div>

        <div style="text-align: center;">
            <span class="status-badge success">
                <span>CLI Installed</span>
            </span>
        </div>

        <div class="section">
            <div class="section-header">
                <span class="section-icon">2</span>
                <h2 class="section-title">Choose Setup Type</h2>
            </div>

            <button class="btn btn-primary" data-command="initGlobal">
                <span class="btn-icon">🏠</span>
                <span class="btn-content">
                    <span class="btn-label">Global Setup</span>
                    <span class="btn-description">Store data in ~/.ung/ - Access from any project</span>
                </span>
                <span class="btn-badge">Recommended</span>
            </button>

            <button class="btn btn-secondary" data-command="initLocal">
                <span class="btn-icon">📁</span>
                <span class="btn-content">
                    <span class="btn-label">Project Setup</span>
                    <span class="btn-description">Store data in .ung/ - Isolated to this workspace</span>
                </span>
            </button>
        </div>

        <div class="info-box">
            <span class="info-box-icon">💡</span>
            <div class="info-box-content">
                <div class="info-box-title">Which should I choose?</div>
                <div class="info-box-text">
                    Global setup is ideal for most freelancers. Choose project setup only if you need completely separate billing data per project.
                </div>
            </div>
        </div>

        <div class="divider"></div>

        <div class="section">
            <div class="collapsible open">
                <div class="collapsible-header">
                    <span class="collapsible-icon">▶</span>
                    <span class="collapsible-title">What you can do</span>
                </div>
                <div class="collapsible-content">
                    <div class="features-grid">
                        <div class="feature-card">
                            <span class="feature-icon">⏱️</span>
                            <span class="feature-text">Track Time</span>
                        </div>
                        <div class="feature-card">
                            <span class="feature-icon">📄</span>
                            <span class="feature-text">Create Invoices</span>
                        </div>
                        <div class="feature-card">
                            <span class="feature-icon">👥</span>
                            <span class="feature-text">Manage Clients</span>
                        </div>
                        <div class="feature-card">
                            <span class="feature-icon">📝</span>
                            <span class="feature-text">Handle Contracts</span>
                        </div>
                        <div class="feature-card">
                            <span class="feature-icon">💳</span>
                            <span class="feature-text">Track Expenses</span>
                        </div>
                        <div class="feature-card">
                            <span class="feature-icon">📊</span>
                            <span class="feature-text">View Reports</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="divider"></div>

        <div class="footer-links">
            <a class="link" data-command="openDocs">
                <span class="link-icon">📚</span>
                Documentation
            </a>
            <a class="link" data-command="recheckCli">
                <span class="link-icon">🔄</span>
                Recheck
            </a>
        </div>
    `;
  }

  private _getReadyHtml(): string {
    return `
        <div class="success-container">
            <div class="success-icon">✓</div>
            <h1 class="success-title">All Set!</h1>
            <p class="success-message">UNG is ready. Check out the sidebar views to get started.</p>
        </div>

        <div class="info-box">
            <span class="info-box-icon">🚀</span>
            <div class="info-box-content">
                <div class="info-box-title">Quick Start</div>
                <div class="info-box-text">
                    Use the Dashboard, Invoices, Clients, and Time Tracking views in the sidebar to manage your freelance business.
                </div>
            </div>
        </div>

        <div class="divider"></div>

        <div class="footer-links">
            <a class="link" data-command="openDocs">
                <span class="link-icon">📚</span>
                Documentation
            </a>
        </div>
    `;
  }
}
