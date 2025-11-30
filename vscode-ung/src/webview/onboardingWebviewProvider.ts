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
        // Quick actions for ready state
        case 'startTracking':
          vscode.commands.executeCommand('ung.startTracking');
          break;
        case 'createInvoice':
          vscode.commands.executeCommand('ung.createInvoice');
          break;
        case 'addGig':
          vscode.commands.executeCommand('ung.addGig');
          break;
        case 'setGoal':
          vscode.commands.executeCommand('ung.setGoal');
          break;
        case 'openDashboard':
          vscode.commands.executeCommand('ung.openDashboard');
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
            position: relative;
            overflow-x: hidden;
        }

        .container {
            max-width: 100%;
            position: relative;
            z-index: 1;
        }

        /* ===== ANIMATED GRADIENT BACKGROUND ===== */
        .gradient-bg {
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            z-index: 0;
            overflow: hidden;
            pointer-events: none;
        }

        .gradient-orb {
            position: absolute;
            border-radius: 50%;
            filter: blur(60px);
            opacity: 0.15;
            animation: float-orb 20s ease-in-out infinite;
        }

        .gradient-orb.orb-1 {
            width: 300px;
            height: 300px;
            background: linear-gradient(135deg, #667eea, #764ba2);
            top: -100px;
            left: -100px;
            animation-delay: 0s;
        }

        .gradient-orb.orb-2 {
            width: 250px;
            height: 250px;
            background: linear-gradient(135deg, #f093fb, #f5576c);
            bottom: -80px;
            right: -80px;
            animation-delay: -7s;
        }

        .gradient-orb.orb-3 {
            width: 200px;
            height: 200px;
            background: linear-gradient(135deg, #4facfe, #00f2fe);
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            animation-delay: -14s;
        }

        @keyframes float-orb {
            0%, 100% { transform: translate(0, 0) scale(1); }
            25% { transform: translate(30px, -30px) scale(1.1); }
            50% { transform: translate(-20px, 20px) scale(0.95); }
            75% { transform: translate(20px, 30px) scale(1.05); }
        }

        .gradient-orb.orb-3 {
            animation-name: float-orb-center;
        }

        @keyframes float-orb-center {
            0%, 100% { transform: translate(-50%, -50%) scale(1); }
            25% { transform: translate(-40%, -60%) scale(1.1); }
            50% { transform: translate(-60%, -40%) scale(0.95); }
            75% { transform: translate(-45%, -55%) scale(1.05); }
        }

        /* ===== FLOATING PARTICLES ===== */
        .particles {
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            z-index: 0;
            pointer-events: none;
            overflow: hidden;
        }

        .particle {
            position: absolute;
            width: 4px;
            height: 4px;
            background: var(--vscode-textLink-foreground);
            border-radius: 50%;
            opacity: 0;
            animation: particle-float 15s ease-in-out infinite;
        }

        @keyframes particle-float {
            0% { opacity: 0; transform: translateY(100vh) scale(0); }
            10% { opacity: 0.6; }
            90% { opacity: 0.6; }
            100% { opacity: 0; transform: translateY(-20px) scale(1); }
        }

        /* ===== PROGRESS BAR ===== */
        .progress-container {
            margin-bottom: 24px;
        }

        .progress-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
        }

        .progress-step {
            font-size: 11px;
            font-weight: 600;
            color: var(--vscode-textLink-foreground);
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }

        .progress-percent {
            font-size: 11px;
            color: var(--vscode-descriptionForeground);
        }

        .progress-bar {
            display: flex;
            gap: 4px;
            height: 4px;
        }

        .progress-segment {
            flex: 1;
            background: var(--vscode-input-background);
            border-radius: 2px;
            overflow: hidden;
            position: relative;
        }

        .progress-segment.active::after,
        .progress-segment.completed::after {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            height: 100%;
            background: linear-gradient(90deg, var(--vscode-textLink-foreground), var(--vscode-textLink-activeForeground, var(--vscode-textLink-foreground)));
            border-radius: 2px;
        }

        .progress-segment.completed::after {
            width: 100%;
        }

        .progress-segment.active::after {
            width: 50%;
            animation: progress-pulse 2s ease-in-out infinite;
        }

        @keyframes progress-pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.6; }
        }

        /* ===== LOADING SPINNER - ENHANCED ===== */
        .loading-container {
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            padding: 60px 20px;
            animation: fade-in 0.5s ease-out;
        }

        @keyframes fade-in {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .spinner {
            width: 48px;
            height: 48px;
            border: 3px solid var(--vscode-input-background);
            border-top: 3px solid var(--vscode-textLink-foreground);
            border-radius: 50%;
            animation: spin 1s linear infinite;
            position: relative;
        }

        .spinner::before {
            content: '';
            position: absolute;
            top: -3px;
            left: -3px;
            right: -3px;
            bottom: -3px;
            border: 3px solid transparent;
            border-top: 3px solid var(--vscode-textLink-activeForeground, var(--vscode-textLink-foreground));
            border-radius: 50%;
            animation: spin 2s linear infinite reverse;
            opacity: 0.5;
        }

        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }

        .loading-text {
            margin-top: 20px;
            color: var(--vscode-descriptionForeground);
            font-size: 13px;
            animation: pulse-text 2s ease-in-out infinite;
        }

        @keyframes pulse-text {
            0%, 100% { opacity: 0.7; }
            50% { opacity: 1; }
        }

        /* ===== STAGGERED ANIMATIONS ===== */
        .animate-in {
            animation: slide-up-fade 0.6s ease-out forwards;
            opacity: 0;
        }

        .animate-in.delay-1 { animation-delay: 0.1s; }
        .animate-in.delay-2 { animation-delay: 0.2s; }
        .animate-in.delay-3 { animation-delay: 0.3s; }
        .animate-in.delay-4 { animation-delay: 0.4s; }
        .animate-in.delay-5 { animation-delay: 0.5s; }
        .animate-in.delay-6 { animation-delay: 0.6s; }

        @keyframes slide-up-fade {
            from {
                opacity: 0;
                transform: translateY(20px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }

        /* Feature items stagger */
        .feature-card {
            animation: feature-pop 0.4s ease-out forwards;
            opacity: 0;
        }

        .feature-card:nth-child(1) { animation-delay: 0.3s; }
        .feature-card:nth-child(2) { animation-delay: 0.4s; }
        .feature-card:nth-child(3) { animation-delay: 0.5s; }
        .feature-card:nth-child(4) { animation-delay: 0.6s; }
        .feature-card:nth-child(5) { animation-delay: 0.7s; }
        .feature-card:nth-child(6) { animation-delay: 0.8s; }

        @keyframes feature-pop {
            from {
                opacity: 0;
                transform: scale(0.8) translateY(10px);
            }
            to {
                opacity: 1;
                transform: scale(1) translateY(0);
            }
        }

        /* ===== TESTIMONIAL CAROUSEL ===== */
        .testimonials {
            margin: 20px 0;
            position: relative;
            overflow: hidden;
        }

        .testimonial-track {
            display: flex;
            transition: transform 0.5s cubic-bezier(0.4, 0, 0.2, 1);
        }

        .testimonial-card {
            flex: 0 0 100%;
            padding: 16px;
            background: linear-gradient(135deg,
                color-mix(in srgb, var(--vscode-textLink-foreground) 10%, var(--vscode-input-background)),
                var(--vscode-input-background));
            border-radius: 12px;
            border: 1px solid color-mix(in srgb, var(--vscode-textLink-foreground) 20%, transparent);
            position: relative;
            overflow: hidden;
        }

        .testimonial-card::before {
            content: '"';
            position: absolute;
            top: -10px;
            left: 10px;
            font-size: 60px;
            font-family: Georgia, serif;
            color: var(--vscode-textLink-foreground);
            opacity: 0.15;
            line-height: 1;
        }

        .testimonial-quote {
            font-size: 13px;
            line-height: 1.6;
            color: var(--vscode-foreground);
            margin-bottom: 12px;
            font-style: italic;
            position: relative;
            z-index: 1;
        }

        .testimonial-author {
            display: flex;
            align-items: center;
            gap: 10px;
        }

        .testimonial-avatar {
            width: 32px;
            height: 32px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 14px;
            font-weight: 600;
            color: white;
        }

        .testimonial-info {
            flex: 1;
        }

        .testimonial-name {
            font-size: 12px;
            font-weight: 600;
            color: var(--vscode-foreground);
        }

        .testimonial-role {
            font-size: 11px;
            color: var(--vscode-descriptionForeground);
        }

        .testimonial-feature {
            font-size: 10px;
            padding: 3px 8px;
            background: var(--vscode-badge-background);
            color: var(--vscode-badge-foreground);
            border-radius: 10px;
            margin-left: auto;
        }

        .testimonial-dots {
            display: flex;
            justify-content: center;
            gap: 6px;
            margin-top: 12px;
        }

        .testimonial-dot {
            width: 6px;
            height: 6px;
            border-radius: 50%;
            background: var(--vscode-input-background);
            border: 1px solid var(--vscode-input-border, var(--vscode-input-background));
            cursor: pointer;
            transition: all 0.3s ease;
        }

        .testimonial-dot.active {
            background: var(--vscode-textLink-foreground);
            transform: scale(1.3);
        }

        /* Respect reduced motion */
        @media (prefers-reduced-motion: reduce) {
            .gradient-orb,
            .particle,
            .animate-in,
            .feature-card,
            .spinner::before,
            .progress-segment.active::after {
                animation: none !important;
                opacity: 1 !important;
                transform: none !important;
            }
            .testimonial-track {
                transition: none;
            }
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

        /* Focus styles for keyboard accessibility */
        .btn:focus {
            outline: 2px solid var(--vscode-focusBorder);
            outline-offset: 2px;
        }

        .btn:focus:not(:focus-visible) {
            outline: none;
        }

        .btn:focus-visible {
            outline: 2px solid var(--vscode-focusBorder);
            outline-offset: 2px;
        }

        .link:focus {
            outline: 2px solid var(--vscode-focusBorder);
            outline-offset: 2px;
            border-radius: 6px;
        }

        .link:focus:not(:focus-visible) {
            outline: none;
        }

        .collapsible-header:focus {
            outline: 2px solid var(--vscode-focusBorder);
            outline-offset: -2px;
        }

        .collapsible-header:focus:not(:focus-visible) {
            outline: none;
        }

        /* Screen reader only text */
        .sr-only {
            position: absolute;
            width: 1px;
            height: 1px;
            padding: 0;
            margin: -1px;
            overflow: hidden;
            clip: rect(0, 0, 0, 0);
            white-space: nowrap;
            border: 0;
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
    <!-- Animated Gradient Background -->
    <div class="gradient-bg" aria-hidden="true">
        <div class="gradient-orb orb-1"></div>
        <div class="gradient-orb orb-2"></div>
        <div class="gradient-orb orb-3"></div>
    </div>

    <!-- Floating Particles -->
    <div class="particles" id="particles" aria-hidden="true"></div>

    <div class="container">
        ${this._state === 'loading' ? this._getLoadingHtml() : ''}
        ${this._state === 'not-installed' ? this._getNotInstalledHtml(platform) : ''}
        ${this._state === 'not-initialized' ? this._getNotInitializedHtml() : ''}
        ${this._state === 'ready' ? this._getReadyHtml() : ''}
    </div>

    <script>
        // Generate floating particles
        (function() {
            if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
            const particles = document.getElementById('particles');
            for (let i = 0; i < 15; i++) {
                const particle = document.createElement('div');
                particle.className = 'particle';
                particle.style.left = Math.random() * 100 + '%';
                particle.style.animationDelay = Math.random() * 15 + 's';
                particle.style.animationDuration = (10 + Math.random() * 10) + 's';
                particles.appendChild(particle);
            }
        })();
        const vscode = acquireVsCodeApi();

        // Handle button clicks and keyboard activation
        document.querySelectorAll('[data-command]').forEach(btn => {
            const handleActivation = () => {
                const command = btn.getAttribute('data-command');
                const platform = btn.getAttribute('data-platform');
                vscode.postMessage({ command, platform });
            };

            btn.addEventListener('click', handleActivation);

            // Keyboard support for Enter and Space
            btn.addEventListener('keydown', (e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    handleActivation();
                }
            });
        });

        // Handle collapsible sections with keyboard support
        document.querySelectorAll('.collapsible-header').forEach(header => {
            const toggleCollapsible = () => {
                header.parentElement.classList.toggle('open');
                const isOpen = header.parentElement.classList.contains('open');
                header.setAttribute('aria-expanded', isOpen ? 'true' : 'false');
            };

            header.addEventListener('click', toggleCollapsible);

            // Keyboard support for collapsibles
            header.addEventListener('keydown', (e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    toggleCollapsible();
                }
            });
        });

        // Set focus to first actionable element on load
        setTimeout(() => {
            const firstButton = document.querySelector('.btn');
            if (firstButton) firstButton.focus();
        }, 100);

        // Testimonial carousel auto-rotation
        (function() {
            const track = document.getElementById('testimonialTrack');
            const dots = document.querySelectorAll('.testimonial-dot');
            if (!track || dots.length === 0) return;

            let currentIndex = 0;
            const totalSlides = dots.length;
            let autoRotateInterval;

            function goToSlide(index) {
                currentIndex = index;
                track.style.transform = 'translateX(-' + (index * 100) + '%)';
                dots.forEach((dot, i) => {
                    dot.classList.toggle('active', i === index);
                    dot.setAttribute('aria-selected', i === index ? 'true' : 'false');
                });
            }

            function nextSlide() {
                goToSlide((currentIndex + 1) % totalSlides);
            }

            // Auto-rotate every 5 seconds
            function startAutoRotate() {
                if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
                autoRotateInterval = setInterval(nextSlide, 5000);
            }

            function stopAutoRotate() {
                clearInterval(autoRotateInterval);
            }

            // Handle dot clicks
            dots.forEach((dot, index) => {
                dot.addEventListener('click', () => {
                    stopAutoRotate();
                    goToSlide(index);
                    startAutoRotate();
                });
                dot.addEventListener('keydown', (e) => {
                    if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        stopAutoRotate();
                        goToSlide(index);
                        startAutoRotate();
                    }
                });
            });

            // Pause on hover
            track.parentElement.addEventListener('mouseenter', stopAutoRotate);
            track.parentElement.addEventListener('mouseleave', startAutoRotate);

            startAutoRotate();
        })();
    </script>
</body>
</html>`;
  }

  private _getLoadingHtml(): string {
    return `
        <div class="loading-container" role="status" aria-live="polite">
            <div class="spinner" aria-hidden="true"></div>
            <div class="loading-text">Checking UNG CLI status...</div>
        </div>
    `;
  }

  private _getNotInstalledHtml(platform: string): string {
    const isMac = platform === 'darwin';
    const isWindows = platform === 'win32';
    const isLinux = platform === 'linux';

    return `
        <!-- Progress Bar -->
        <div class="progress-container animate-in">
            <div class="progress-header">
                <span class="progress-step">Step 1 of 3</span>
                <span class="progress-percent">0%</span>
            </div>
            <div class="progress-bar" role="progressbar" aria-valuenow="0" aria-valuemin="0" aria-valuemax="100">
                <div class="progress-segment active"></div>
                <div class="progress-segment"></div>
                <div class="progress-segment"></div>
            </div>
        </div>

        <div class="header animate-in delay-1">
            <div class="logo-container">
                <span class="logo-icon">U</span>
            </div>
            <h1 class="title">Your Freelance Journey Starts Here</h1>
            <p class="subtitle">From tracking time to getting paid — all in one place</p>
        </div>

        <div class="section animate-in delay-2">
            <div class="section-header">
                <span class="section-icon">🚀</span>
                <h2 class="section-title">Quick Install</h2>
            </div>

            ${
              isMac || isLinux
                ? `
            <button class="btn btn-primary" data-command="installHomebrew" aria-label="Install UNG CLI via Homebrew - Recommended">
                <span class="btn-icon" aria-hidden="true">🍺</span>
                <span class="btn-content">
                    <span class="btn-label">Install via Homebrew</span>
                    <span class="btn-description">brew install ung</span>
                </span>
                <span class="btn-badge" aria-hidden="true">Recommended</span>
            </button>
            `
                : ''
            }

            ${
              isWindows
                ? `
            <button class="btn btn-primary" data-command="installScoop" aria-label="Install UNG CLI via Scoop - Recommended">
                <span class="btn-icon" aria-hidden="true">🪣</span>
                <span class="btn-content">
                    <span class="btn-label">Install via Scoop</span>
                    <span class="btn-description">scoop install ung</span>
                </span>
                <span class="btn-badge" aria-hidden="true">Recommended</span>
            </button>
            `
                : ''
            }

            <button class="btn btn-secondary" data-command="downloadBinary" data-platform="${platform}" aria-label="Download UNG binary for ${isMac ? 'macOS' : isWindows ? 'Windows' : 'Linux'}">
                <span class="btn-icon" aria-hidden="true">📦</span>
                <span class="btn-content">
                    <span class="btn-label">Download Binary</span>
                    <span class="btn-description">Direct download for ${isMac ? 'macOS' : isWindows ? 'Windows' : 'Linux'}</span>
                </span>
            </button>

            <button class="btn btn-secondary" data-command="installGo" aria-label="Install UNG CLI via Go toolchain">
                <span class="btn-icon" aria-hidden="true">🐹</span>
                <span class="btn-content">
                    <span class="btn-label">Install via Go</span>
                    <span class="btn-description">go install github.com/Andriiklymiuk/ung@latest</span>
                </span>
            </button>
        </div>

        <!-- Testimonial Carousel -->
        <div class="testimonials animate-in delay-3" aria-label="User testimonials">
            <div class="testimonial-track" id="testimonialTrack">
                <div class="testimonial-card">
                    <p class="testimonial-quote">Dig killed 6 of my ideas in one afternoon. The 7th scored 73. Shipping next week.</p>
                    <div class="testimonial-author">
                        <div class="testimonial-avatar" style="background: linear-gradient(135deg, #667eea, #764ba2);">M</div>
                        <div class="testimonial-info">
                            <div class="testimonial-name">Marcus Chen</div>
                            <div class="testimonial-role">Indie Hacker</div>
                        </div>
                        <span class="testimonial-feature">Dig</span>
                    </div>
                </div>
                <div class="testimonial-card">
                    <p class="testimonial-quote">Discovered I was billing 4 hours but working 9. Raised my rates 3x. Now I work less, earn more.</p>
                    <div class="testimonial-author">
                        <div class="testimonial-avatar" style="background: linear-gradient(135deg, #f093fb, #f5576c);">J</div>
                        <div class="testimonial-info">
                            <div class="testimonial-name">Jake Rivera</div>
                            <div class="testimonial-role">Full-Stack Dev</div>
                        </div>
                        <span class="testimonial-feature">Tracking</span>
                    </div>
                </div>
                <div class="testimonial-card">
                    <p class="testimonial-quote">Client ghosted for 47 days. Sent invoice with overdue badge. Paid in 6 hours.</p>
                    <div class="testimonial-author">
                        <div class="testimonial-avatar" style="background: linear-gradient(135deg, #4facfe, #00f2fe);">E</div>
                        <div class="testimonial-info">
                            <div class="testimonial-name">Emma Lindqvist</div>
                            <div class="testimonial-role">UX Designer</div>
                        </div>
                        <span class="testimonial-feature">Invoices</span>
                    </div>
                </div>
                <div class="testimonial-card">
                    <p class="testimonial-quote">$8k contract from a Hunt alert while I was sleeping. Best ROI of any tool I own.</p>
                    <div class="testimonial-author">
                        <div class="testimonial-avatar" style="background: linear-gradient(135deg, #11998e, #38ef7d);">N</div>
                        <div class="testimonial-info">
                            <div class="testimonial-name">Nina Volkov</div>
                            <div class="testimonial-role">DevOps Consultant</div>
                        </div>
                        <span class="testimonial-feature">Hunt</span>
                    </div>
                </div>
            </div>
            <div class="testimonial-dots" role="tablist" aria-label="Testimonial navigation">
                <button class="testimonial-dot active" data-index="0" role="tab" aria-selected="true" aria-label="Testimonial 1"></button>
                <button class="testimonial-dot" data-index="1" role="tab" aria-selected="false" aria-label="Testimonial 2"></button>
                <button class="testimonial-dot" data-index="2" role="tab" aria-selected="false" aria-label="Testimonial 3"></button>
                <button class="testimonial-dot" data-index="3" role="tab" aria-selected="false" aria-label="Testimonial 4"></button>
            </div>
        </div>

        <div class="divider"></div>

        <div class="section animate-in delay-4">
            <div class="collapsible">
                <div class="collapsible-header" role="button" tabindex="0" aria-expanded="false" aria-controls="features-content">
                    <span class="collapsible-icon" aria-hidden="true">▶</span>
                    <span class="collapsible-title">What you'll get</span>
                </div>
                <div class="collapsible-content" id="features-content" role="region">
                    <ul class="features-grid" role="list" aria-label="Features included">
                        <li class="feature-card" role="listitem">
                            <span class="feature-icon" aria-hidden="true">📄</span>
                            <span class="feature-text">Invoices</span>
                        </li>
                        <li class="feature-card" role="listitem">
                            <span class="feature-icon" aria-hidden="true">⏱️</span>
                            <span class="feature-text">Time Tracking</span>
                        </li>
                        <li class="feature-card" role="listitem">
                            <span class="feature-icon" aria-hidden="true">👥</span>
                            <span class="feature-text">Clients</span>
                        </li>
                        <li class="feature-card" role="listitem">
                            <span class="feature-icon" aria-hidden="true">📝</span>
                            <span class="feature-text">Contracts</span>
                        </li>
                        <li class="feature-card" role="listitem">
                            <span class="feature-icon" aria-hidden="true">💳</span>
                            <span class="feature-text">Expenses</span>
                        </li>
                        <li class="feature-card" role="listitem">
                            <span class="feature-icon" aria-hidden="true">📊</span>
                            <span class="feature-text">Reports</span>
                        </li>
                    </ul>
                    <div class="highlights" role="list" aria-label="Key benefits">
                        <span class="highlight-tag" role="listitem"><span class="highlight-icon" aria-hidden="true">🔒</span> Privacy First</span>
                        <span class="highlight-tag" role="listitem"><span class="highlight-icon" aria-hidden="true">📴</span> Offline</span>
                        <span class="highlight-tag" role="listitem"><span class="highlight-icon" aria-hidden="true">🌍</span> Multi-Currency</span>
                        <span class="highlight-tag" role="listitem"><span class="highlight-icon" aria-hidden="true">⚡</span> Fast</span>
                    </div>
                </div>
            </div>
        </div>

        <div class="divider" role="separator"></div>

        <nav class="footer-links animate-in delay-5" aria-label="Additional links">
            <button class="link" data-command="openDocs" aria-label="Open documentation">
                <span class="link-icon" aria-hidden="true">📚</span>
                Documentation
            </button>
            <button class="link" data-command="recheckCli" aria-label="Recheck CLI installation status">
                <span class="link-icon" aria-hidden="true">🔄</span>
                Recheck
            </button>
        </nav>
    `;
  }

  private _getNotInitializedHtml(): string {
    return `
        <!-- Progress Bar -->
        <div class="progress-container animate-in">
            <div class="progress-header">
                <span class="progress-step">Step 2 of 3</span>
                <span class="progress-percent">33%</span>
            </div>
            <div class="progress-bar" role="progressbar" aria-valuenow="33" aria-valuemin="0" aria-valuemax="100">
                <div class="progress-segment completed"></div>
                <div class="progress-segment active"></div>
                <div class="progress-segment"></div>
            </div>
        </div>

        <div class="header animate-in delay-1">
            <div class="logo-container">
                <span class="logo-icon">U</span>
            </div>
            <h1 class="title">Where Should We Store Your Data?</h1>
            <p class="subtitle">One click away from tracking time and getting paid</p>
        </div>

        <div style="text-align: center;" class="animate-in delay-2">
            <span class="status-badge success">
                <span>✓ CLI Installed</span>
            </span>
        </div>

        <div class="section animate-in delay-2">
            <div class="section-header">
                <span class="section-icon">📂</span>
                <h2 class="section-title">Choose Your Setup</h2>
            </div>

            <button class="btn btn-primary" data-command="initGlobal" aria-label="Global setup - Store data in home directory, access from any project - Recommended">
                <span class="btn-icon" aria-hidden="true">🏠</span>
                <span class="btn-content">
                    <span class="btn-label">Global Setup</span>
                    <span class="btn-description">Store data in ~/.ung/ - Access from any project</span>
                </span>
                <span class="btn-badge" aria-hidden="true">Recommended</span>
            </button>

            <button class="btn btn-secondary" data-command="initLocal" aria-label="Project setup - Store data in current workspace only">
                <span class="btn-icon" aria-hidden="true">📁</span>
                <span class="btn-content">
                    <span class="btn-label">Project Setup</span>
                    <span class="btn-description">Store data in .ung/ - Isolated to this workspace</span>
                </span>
            </button>
        </div>

        <div class="info-box animate-in delay-3">
            <span class="info-box-icon">💡</span>
            <div class="info-box-content">
                <div class="info-box-title">Which should I choose?</div>
                <div class="info-box-text">
                    Global setup is ideal for most freelancers. Choose project setup only if you need completely separate billing data per project.
                </div>
            </div>
        </div>

        <!-- Testimonial Carousel -->
        <div class="testimonials animate-in delay-4" aria-label="User testimonials">
            <div class="testimonial-track" id="testimonialTrack">
                <div class="testimonial-card">
                    <p class="testimonial-quote">Had 23 projects marked "in progress". Kanban showed 19 were actually dead. Shipped the other 4 in 2 weeks.</p>
                    <div class="testimonial-author">
                        <div class="testimonial-avatar" style="background: linear-gradient(135deg, #fa709a, #fee140);">D</div>
                        <div class="testimonial-info">
                            <div class="testimonial-name">David Morrison</div>
                            <div class="testimonial-role">Mobile Developer</div>
                        </div>
                        <span class="testimonial-feature">Kanban</span>
                    </div>
                </div>
                <div class="testimonial-card">
                    <p class="testimonial-quote">Goal bar turned red at $2k. Panicked. Hustled. Hit $11k that month. Never going back to flying blind.</p>
                    <div class="testimonial-author">
                        <div class="testimonial-avatar" style="background: linear-gradient(135deg, #a8edea, #fed6e3);">S</div>
                        <div class="testimonial-info">
                            <div class="testimonial-name">Sarah Okonkwo</div>
                            <div class="testimonial-role">Marketing Consultant</div>
                        </div>
                        <span class="testimonial-feature">Goals</span>
                    </div>
                </div>
                <div class="testimonial-card">
                    <p class="testimonial-quote">Focus timer saved me 4.5 hours per week. Built my MVP in 6 weeks instead of the 6 months I'd planned.</p>
                    <div class="testimonial-author">
                        <div class="testimonial-avatar" style="background: linear-gradient(135deg, #667eea, #764ba2);">R</div>
                        <div class="testimonial-info">
                            <div class="testimonial-name">Raj Patel</div>
                            <div class="testimonial-role">Startup Founder</div>
                        </div>
                        <span class="testimonial-feature">Focus</span>
                    </div>
                </div>
            </div>
            <div class="testimonial-dots" role="tablist" aria-label="Testimonial navigation">
                <button class="testimonial-dot active" data-index="0" role="tab" aria-selected="true" aria-label="Testimonial 1"></button>
                <button class="testimonial-dot" data-index="1" role="tab" aria-selected="false" aria-label="Testimonial 2"></button>
                <button class="testimonial-dot" data-index="2" role="tab" aria-selected="false" aria-label="Testimonial 3"></button>
            </div>
        </div>

        <div class="divider"></div>

        <nav class="footer-links animate-in delay-5" aria-label="Additional links">
            <button class="link" data-command="openDocs" aria-label="Open documentation">
                <span class="link-icon" aria-hidden="true">📚</span>
                Documentation
            </button>
            <button class="link" data-command="recheckCli" aria-label="Recheck CLI installation status">
                <span class="link-icon" aria-hidden="true">🔄</span>
                Recheck
            </button>
        </nav>
    `;
  }

  private _getReadyHtml(): string {
    return `
        <style>
            .confetti {
                position: fixed;
                top: 0;
                left: 0;
                right: 0;
                bottom: 0;
                pointer-events: none;
                z-index: 1000;
            }
            .confetti-piece {
                position: absolute;
                width: 10px;
                height: 10px;
                animation: confetti-fall 3s ease-out forwards;
                opacity: 0;
            }
            @keyframes confetti-fall {
                0% {
                    transform: translateY(-20px) rotate(0deg);
                    opacity: 1;
                }
                100% {
                    transform: translateY(100vh) rotate(720deg);
                    opacity: 0;
                }
            }
            .success-icon {
                animation: success-bounce 0.6s cubic-bezier(0.68, -0.55, 0.265, 1.55) forwards;
            }
            @keyframes success-bounce {
                0% { transform: scale(0); opacity: 0; }
                50% { transform: scale(1.2); }
                100% { transform: scale(1); opacity: 1; }
            }
            .success-checkmark {
                width: 80px;
                height: 80px;
                border-radius: 50%;
                display: flex;
                align-items: center;
                justify-content: center;
                background: linear-gradient(135deg, var(--vscode-charts-green, #4caf50), color-mix(in srgb, var(--vscode-charts-green, #4caf50) 70%, #000));
                box-shadow: 0 8px 32px rgba(76, 175, 80, 0.3);
                margin: 0 auto 20px;
                position: relative;
                overflow: hidden;
            }
            .success-checkmark::before {
                content: '';
                position: absolute;
                width: 200%;
                height: 200%;
                background: linear-gradient(45deg, transparent, rgba(255,255,255,0.2), transparent);
                animation: shine 2s ease-in-out infinite;
            }
            @keyframes shine {
                0% { transform: translateX(-100%) rotate(45deg); }
                100% { transform: translateX(100%) rotate(45deg); }
            }
            .pulse-ring {
                position: absolute;
                width: 100%;
                height: 100%;
                border-radius: 50%;
                border: 2px solid var(--vscode-charts-green, #4caf50);
                animation: pulse-ring 2s ease-out infinite;
            }
            @keyframes pulse-ring {
                0% { transform: scale(1); opacity: 0.8; }
                100% { transform: scale(1.5); opacity: 0; }
            }
            /* Respect user's motion preferences for accessibility */
            @media (prefers-reduced-motion: reduce) {
                .confetti-piece, .success-icon, .success-checkmark::before, .pulse-ring {
                    animation: none;
                    opacity: 1;
                    transform: none;
                }
                .confetti {
                    display: none;
                }
            }
        </style>
        <div class="confetti" id="confetti" aria-hidden="true"></div>
        <script>
            (function() {
                // Only show confetti if user doesn't prefer reduced motion
                if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
                    return;
                }
                const confetti = document.getElementById('confetti');
                const colors = ['#667eea', '#764ba2', '#f093fb', '#f5576c', '#4facfe', '#00f2fe', '#11998e', '#38ef7d'];
                for (let i = 0; i < 60; i++) {
                    const piece = document.createElement('div');
                    piece.className = 'confetti-piece';
                    piece.style.left = Math.random() * 100 + '%';
                    piece.style.backgroundColor = colors[Math.floor(Math.random() * colors.length)];
                    piece.style.animationDelay = Math.random() * 1.5 + 's';
                    piece.style.borderRadius = Math.random() > 0.5 ? '50%' : Math.random() > 0.5 ? '2px' : '0';
                    piece.style.width = (5 + Math.random() * 10) + 'px';
                    piece.style.height = (5 + Math.random() * 10) + 'px';
                    confetti.appendChild(piece);
                }
                setTimeout(() => confetti.remove(), 5000);
            })();
        </script>

        <!-- Progress Bar - Complete! -->
        <div class="progress-container animate-in">
            <div class="progress-header">
                <span class="progress-step">Complete!</span>
                <span class="progress-percent">100%</span>
            </div>
            <div class="progress-bar" role="progressbar" aria-valuenow="100" aria-valuemin="0" aria-valuemax="100">
                <div class="progress-segment completed"></div>
                <div class="progress-segment completed"></div>
                <div class="progress-segment completed"></div>
            </div>
        </div>

        <div class="success-container animate-in delay-1">
            <div class="success-checkmark">
                <div class="pulse-ring"></div>
                <span style="font-size: 36px; position: relative; z-index: 1;">✓</span>
            </div>
            <h1 class="success-title">Your Journey Begins Now!</h1>
            <p class="success-message">Track time. Manage gigs. Get paid. All from VS Code.</p>
        </div>

        <div class="section animate-in delay-2">
            <div class="section-header">
                <span class="section-icon">⚡</span>
                <h2 class="section-title">What's First?</h2>
            </div>

            <button class="btn btn-primary" data-command="startTracking" aria-label="Start tracking time - Begin a work session now">
                <span class="btn-icon" aria-hidden="true">▶️</span>
                <span class="btn-content">
                    <span class="btn-label">Start Tracking Time</span>
                    <span class="btn-description">Begin a work session now</span>
                </span>
            </button>

            <button class="btn btn-secondary" data-command="createInvoice" aria-label="Create invoice - Bill your clients">
                <span class="btn-icon" aria-hidden="true">📄</span>
                <span class="btn-content">
                    <span class="btn-label">Create Invoice</span>
                    <span class="btn-description">Bill your clients</span>
                </span>
            </button>

            <button class="btn btn-secondary" data-command="addGig" aria-label="Add a gig - Track projects on your kanban board">
                <span class="btn-icon" aria-hidden="true">🎯</span>
                <span class="btn-content">
                    <span class="btn-label">Add a Gig</span>
                    <span class="btn-description">Track projects on your kanban board</span>
                </span>
            </button>

            <button class="btn btn-secondary" data-command="setGoal" aria-label="Set income goal - Track progress toward your target">
                <span class="btn-icon" aria-hidden="true">📊</span>
                <span class="btn-content">
                    <span class="btn-label">Set Income Goal</span>
                    <span class="btn-description">Track progress toward your target</span>
                </span>
            </button>
        </div>

        <div class="info-box animate-in delay-3">
            <span class="info-box-icon">⌨️</span>
            <div class="info-box-content">
                <div class="info-box-title">Pro Tip</div>
                <div class="info-box-text">
                    Use <strong>Ctrl+Alt+T</strong> (Cmd+Alt+T on Mac) to quickly start/stop time tracking from anywhere in VS Code.
                </div>
            </div>
        </div>

        <!-- Final Testimonial -->
        <div class="testimonials animate-in delay-4" aria-label="User testimonials">
            <div class="testimonial-track" id="testimonialTrack">
                <div class="testimonial-card">
                    <p class="testimonial-quote">Set up recurring invoices once. $34k collected on autopilot last year while I focused on actual work.</p>
                    <div class="testimonial-author">
                        <div class="testimonial-avatar" style="background: linear-gradient(135deg, #667eea, #764ba2);">P</div>
                        <div class="testimonial-info">
                            <div class="testimonial-name">Priya Krishnamurthy</div>
                            <div class="testimonial-role">Backend Engineer</div>
                        </div>
                        <span class="testimonial-feature">Invoices</span>
                    </div>
                </div>
            </div>
        </div>

        <div class="divider"></div>

        <nav class="footer-links animate-in delay-5" aria-label="Additional links">
            <button class="link" data-command="openDashboard" aria-label="Open dashboard">
                <span class="link-icon" aria-hidden="true">📊</span>
                Dashboard
            </button>
            <button class="link" data-command="openDocs" aria-label="Open documentation">
                <span class="link-icon" aria-hidden="true">📚</span>
                Documentation
            </button>
        </nav>
    `;
  }
}
