import * as vscode from 'vscode';

export class PomodoroPanel {
  public static currentPanel: PomodoroPanel | undefined;
  private readonly _panel: vscode.WebviewPanel;
  private _disposables: vscode.Disposable[] = [];
  private _timer: NodeJS.Timeout | undefined;
  private _statusBarItem: vscode.StatusBarItem;

  // Timer state
  private _isRunning = false;
  private _isBreak = false;
  private _secondsRemaining = 0;
  private _workMinutes = 25;
  private _breakMinutes = 5;
  private _sessionsCompleted = 0;

  private constructor(
    panel: vscode.WebviewPanel,
    _extensionUri: vscode.Uri,
    statusBarItem: vscode.StatusBarItem
  ) {
    this._panel = panel;
    this._statusBarItem = statusBarItem;

    this._panel.webview.html = this._getHtmlForWebview();

    this._panel.onDidDispose(() => this.dispose(), null, this._disposables);

    this._panel.webview.onDidReceiveMessage(
      (message) => {
        switch (message.command) {
          case 'start':
            this._startTimer(message.workMinutes, message.breakMinutes);
            break;
          case 'pause':
            this._pauseTimer();
            break;
          case 'resume':
            this._resumeTimer();
            break;
          case 'stop':
            this._stopTimer();
            break;
          case 'skip':
            this._skipToNext();
            break;
        }
      },
      null,
      this._disposables
    );
  }

  public static createOrShow(
    extensionUri: vscode.Uri,
    statusBarItem: vscode.StatusBarItem
  ) {
    const column = vscode.ViewColumn.Beside;

    if (PomodoroPanel.currentPanel) {
      PomodoroPanel.currentPanel._panel.reveal(column);
      return;
    }

    const panel = vscode.window.createWebviewPanel(
      'ungPomodoro',
      '🍅 Pomodoro Timer',
      column,
      {
        enableScripts: true,
        retainContextWhenHidden: true,
      }
    );

    PomodoroPanel.currentPanel = new PomodoroPanel(
      panel,
      extensionUri,
      statusBarItem
    );
  }

  private _startTimer(workMinutes: number, breakMinutes: number) {
    this._workMinutes = workMinutes;
    this._breakMinutes = breakMinutes;
    this._secondsRemaining = workMinutes * 60;
    this._isRunning = true;
    this._isBreak = false;

    this._tick();
    this._timer = setInterval(() => this._tick(), 1000);
    this._updateStatusBar();
  }

  private _tick() {
    if (!this._isRunning) return;

    this._secondsRemaining--;
    this._updateWebview();
    this._updateStatusBar();

    if (this._secondsRemaining <= 0) {
      this._onTimerComplete();
    }
  }

  private _onTimerComplete() {
    if (this._isBreak) {
      // Break finished, start new work session
      this._isBreak = false;
      this._secondsRemaining = this._workMinutes * 60;
      vscode.window.showInformationMessage('🍅 Break over! Time to focus.');
      this._playSound();
    } else {
      // Work session finished
      this._sessionsCompleted++;
      this._isBreak = true;
      this._secondsRemaining = this._breakMinutes * 60;
      vscode.window.showInformationMessage(
        `🎉 Pomodoro #${this._sessionsCompleted} complete! Take a break.`
      );
      this._playSound();
    }
    this._updateWebview();
  }

  private _pauseTimer() {
    this._isRunning = false;
    if (this._timer) {
      clearInterval(this._timer);
      this._timer = undefined;
    }
    this._updateStatusBar();
  }

  private _resumeTimer() {
    this._isRunning = true;
    this._timer = setInterval(() => this._tick(), 1000);
    this._updateStatusBar();
  }

  private _stopTimer() {
    this._isRunning = false;
    this._isBreak = false;
    this._secondsRemaining = 0;
    if (this._timer) {
      clearInterval(this._timer);
      this._timer = undefined;
    }
    this._updateWebview();
    this._updateStatusBar();
  }

  private _skipToNext() {
    this._secondsRemaining = 0;
    this._onTimerComplete();
  }

  private _playSound() {
    // Use VS Code's built-in notification sound
    vscode.commands.executeCommand('editor.action.playSound', 'terminalBell');
  }

  private _updateStatusBar() {
    if (this._isRunning || this._secondsRemaining > 0) {
      const minutes = Math.floor(this._secondsRemaining / 60);
      const seconds = this._secondsRemaining % 60;
      const timeStr = `${minutes}:${seconds.toString().padStart(2, '0')}`;
      const icon = this._isBreak ? '☕' : '🍅';
      const status = this._isRunning ? '' : ' (paused)';
      this._statusBarItem.text = `${icon} ${timeStr}${status}`;
      this._statusBarItem.tooltip = this._isBreak ? 'Break time' : 'Focus time';
      this._statusBarItem.show();
    } else {
      this._statusBarItem.hide();
    }
  }

  private _updateWebview() {
    this._panel.webview.postMessage({
      type: 'update',
      secondsRemaining: this._secondsRemaining,
      isRunning: this._isRunning,
      isBreak: this._isBreak,
      sessionsCompleted: this._sessionsCompleted,
      workMinutes: this._workMinutes,
      breakMinutes: this._breakMinutes,
    });
  }

  private _getHtmlForWebview() {
    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Pomodoro Timer</title>
    <style>
        /* ==============================================
           UNG Design System - Pomodoro Timer
           Aligned with macOS DesignTokens.swift
           ============================================== */
        :root {
            /* Brand Colors */
            --ung-brand: #3373E8;
            --ung-brand-light: rgba(51, 115, 232, 0.15);
            --ung-brand-dark: #2660CC;

            /* Semantic Colors */
            --ung-success: #33A756;
            --ung-success-light: rgba(51, 167, 86, 0.12);
            --ung-warning: #F29932;
            --ung-warning-light: rgba(242, 153, 50, 0.12);
            --ung-error: #E65A5A;
            --ung-error-light: rgba(230, 90, 90, 0.12);

            /* Spacing */
            --space-xs: 8px;
            --space-sm: 12px;
            --space-md: 16px;
            --space-lg: 24px;
            --space-xl: 32px;

            /* Border Radius */
            --radius-sm: 8px;
            --radius-md: 12px;
            --radius-full: 9999px;

            /* Transitions */
            --transition-micro: 0.1s cubic-bezier(0.4, 0, 0.2, 1);
            --transition-quick: 0.15s cubic-bezier(0.4, 0, 0.2, 1);
            --transition-bounce: 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: var(--vscode-font-family);
            background: var(--vscode-editor-background);
            color: var(--vscode-editor-foreground);
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            padding: var(--space-lg);
        }

        .container {
            text-align: center;
            max-width: 400px;
            width: 100%;
        }

        .timer-display {
            font-size: 64px;
            font-weight: 700;
            margin: var(--space-lg) 0;
            font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, monospace;
            font-variant-numeric: tabular-nums;
            letter-spacing: -2px;
            transition: color var(--transition-quick), transform var(--transition-bounce);
        }

        .timer-display.break {
            color: var(--ung-success);
        }

        .timer-display.work {
            color: var(--ung-warning);
        }

        .timer-display:hover {
            transform: scale(1.02);
        }

        .status {
            font-size: 20px;
            font-weight: 500;
            margin-bottom: var(--space-md);
            opacity: 0.85;
            transition: opacity var(--transition-quick);
        }

        .sessions {
            font-size: 14px;
            margin-bottom: var(--space-xl);
            color: var(--vscode-descriptionForeground);
            padding: var(--space-xs) var(--space-md);
            background: var(--vscode-input-background);
            border-radius: var(--radius-full);
            display: inline-block;
        }

        .controls {
            display: flex;
            gap: var(--space-sm);
            justify-content: center;
            flex-wrap: wrap;
        }

        button {
            background: var(--ung-brand);
            color: white;
            border: none;
            padding: var(--space-sm) var(--space-lg);
            font-size: 14px;
            border-radius: var(--radius-sm);
            cursor: pointer;
            transition: all var(--transition-quick);
            font-weight: 600;
            display: inline-flex;
            align-items: center;
            gap: var(--space-xs);
        }

        button:hover {
            background: var(--ung-brand-dark);
            transform: translateY(-2px);
            box-shadow: 0 6px 16px rgba(51, 115, 232, 0.3);
        }

        button:active {
            transform: translateY(0) scale(0.98);
        }

        button.secondary {
            background: var(--vscode-input-background);
            color: var(--vscode-editor-foreground);
            border: 1px solid var(--vscode-widget-border);
        }

        button.secondary:hover {
            background: var(--vscode-list-hoverBackground);
            border-color: var(--ung-brand);
            box-shadow: none;
        }

        button:disabled {
            opacity: 0.5;
            cursor: not-allowed;
            transform: none;
            box-shadow: none;
        }

        .settings {
            margin-top: var(--space-xl);
            padding-top: var(--space-lg);
            border-top: 1px solid var(--vscode-widget-border);
        }

        .settings h3 {
            margin-bottom: var(--space-md);
            font-size: 12px;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            color: var(--vscode-descriptionForeground);
        }

        .setting-row {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: var(--space-sm);
            margin-bottom: var(--space-sm);
        }

        .setting-row label {
            min-width: 60px;
            text-align: right;
            font-size: 13px;
            font-weight: 500;
        }

        .setting-row input {
            width: 64px;
            padding: var(--space-xs);
            background: var(--vscode-input-background);
            color: var(--vscode-input-foreground);
            border: 1px solid var(--vscode-input-border);
            border-radius: var(--radius-sm);
            text-align: center;
            font-size: 14px;
            font-weight: 600;
            transition: border-color var(--transition-quick), box-shadow var(--transition-quick);
        }

        .setting-row input:focus {
            outline: none;
            border-color: var(--ung-brand);
            box-shadow: 0 0 0 3px var(--ung-brand-light);
        }

        .progress-ring {
            margin: var(--space-lg) auto;
            filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.1));
        }

        .progress-ring circle {
            transition: stroke-dashoffset 1s linear, stroke var(--transition-quick);
        }

        .icon {
            font-size: 48px;
            margin-bottom: var(--space-xs);
            transition: transform var(--transition-bounce);
            display: inline-block;
        }

        .icon:hover {
            transform: scale(1.1) rotate(5deg);
        }

        /* Focus States */
        button:focus-visible,
        input:focus-visible {
            outline: 2px solid var(--ung-brand);
            outline-offset: 2px;
        }

        /* Reduced Motion */
        @media (prefers-reduced-motion: reduce) {
            *, *::before, *::after {
                animation-duration: 0.01ms !important;
                transition-duration: 0.01ms !important;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon" id="icon">🍅</div>
        <div class="status" id="status">Ready to focus</div>

        <svg class="progress-ring" width="200" height="200">
            <circle
                id="progress-bg"
                stroke="var(--vscode-widget-border)"
                stroke-width="8"
                fill="transparent"
                r="90"
                cx="100"
                cy="100"
            />
            <circle
                id="progress"
                stroke="var(--vscode-charts-red)"
                stroke-width="8"
                fill="transparent"
                r="90"
                cx="100"
                cy="100"
                stroke-linecap="round"
                transform="rotate(-90 100 100)"
            />
        </svg>

        <div class="timer-display work" id="timer">25:00</div>
        <div class="sessions" id="sessions">Sessions completed: 0</div>

        <div class="controls" id="controls">
            <button id="startBtn" onclick="start()">▶ Start</button>
        </div>

        <div class="settings" id="settings">
            <h3>Settings</h3>
            <div class="setting-row">
                <label>Work:</label>
                <input type="number" id="workMinutes" value="25" min="1" max="60"> min
            </div>
            <div class="setting-row">
                <label>Break:</label>
                <input type="number" id="breakMinutes" value="5" min="1" max="30"> min
            </div>
        </div>
    </div>

    <script>
        const vscode = acquireVsCodeApi();

        let isRunning = false;
        let isPaused = false;
        let totalSeconds = 25 * 60;

        const circumference = 2 * Math.PI * 90;
        document.getElementById('progress').style.strokeDasharray = circumference;
        document.getElementById('progress').style.strokeDashoffset = 0;

        function formatTime(seconds) {
            const m = Math.floor(seconds / 60);
            const s = seconds % 60;
            return m + ':' + s.toString().padStart(2, '0');
        }

        function updateProgress(remaining, total) {
            const progress = document.getElementById('progress');
            const offset = circumference * (1 - remaining / total);
            progress.style.strokeDashoffset = offset;
        }

        function start() {
            const workMinutes = parseInt(document.getElementById('workMinutes').value) || 25;
            const breakMinutes = parseInt(document.getElementById('breakMinutes').value) || 5;
            vscode.postMessage({ command: 'start', workMinutes, breakMinutes });
        }

        function pause() {
            vscode.postMessage({ command: 'pause' });
        }

        function resume() {
            vscode.postMessage({ command: 'resume' });
        }

        function stop() {
            vscode.postMessage({ command: 'stop' });
        }

        function skip() {
            vscode.postMessage({ command: 'skip' });
        }

        window.addEventListener('message', event => {
            const message = event.data;
            if (message.type === 'update') {
                const timer = document.getElementById('timer');
                const status = document.getElementById('status');
                const icon = document.getElementById('icon');
                const sessions = document.getElementById('sessions');
                const controls = document.getElementById('controls');
                const settings = document.getElementById('settings');
                const progress = document.getElementById('progress');

                timer.textContent = formatTime(message.secondsRemaining);
                sessions.textContent = 'Sessions completed: ' + message.sessionsCompleted;

                // Calculate total for progress
                totalSeconds = message.isBreak ? message.breakMinutes * 60 : message.workMinutes * 60;
                updateProgress(message.secondsRemaining, totalSeconds);

                if (message.isBreak) {
                    timer.className = 'timer-display break';
                    status.textContent = '☕ Break time';
                    icon.textContent = '☕';
                    progress.style.stroke = 'var(--vscode-charts-green)';
                } else {
                    timer.className = 'timer-display work';
                    status.textContent = message.isRunning ? '🎯 Focus mode' : (message.secondsRemaining > 0 ? '⏸ Paused' : 'Ready to focus');
                    icon.textContent = '🍅';
                    progress.style.stroke = 'var(--vscode-charts-red)';
                }

                // Update controls
                if (message.secondsRemaining > 0) {
                    settings.style.display = 'none';
                    if (message.isRunning) {
                        controls.innerHTML = \`
                            <button onclick="pause()">⏸ Pause</button>
                            <button class="secondary" onclick="skip()">⏭ Skip</button>
                            <button class="secondary" onclick="stop()">⏹ Stop</button>
                        \`;
                    } else {
                        controls.innerHTML = \`
                            <button onclick="resume()">▶ Resume</button>
                            <button class="secondary" onclick="stop()">⏹ Stop</button>
                        \`;
                    }
                } else {
                    settings.style.display = 'block';
                    controls.innerHTML = '<button id="startBtn" onclick="start()">▶ Start</button>';
                    updateProgress(1, 1); // Reset progress ring
                }
            }
        });
    </script>
</body>
</html>`;
  }

  public dispose() {
    PomodoroPanel.currentPanel = undefined;

    if (this._timer) {
      clearInterval(this._timer);
    }

    this._statusBarItem.hide();
    this._panel.dispose();

    while (this._disposables.length) {
      const x = this._disposables.pop();
      if (x) {
        x.dispose();
      }
    }
  }
}
