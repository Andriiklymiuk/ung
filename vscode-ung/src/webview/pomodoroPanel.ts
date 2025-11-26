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

    private constructor(panel: vscode.WebviewPanel, _extensionUri: vscode.Uri, statusBarItem: vscode.StatusBarItem) {
        this._panel = panel;
        this._statusBarItem = statusBarItem;

        this._panel.webview.html = this._getHtmlForWebview();

        this._panel.onDidDispose(() => this.dispose(), null, this._disposables);

        this._panel.webview.onDidReceiveMessage(
            message => {
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

    public static createOrShow(extensionUri: vscode.Uri, statusBarItem: vscode.StatusBarItem) {
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
                retainContextWhenHidden: true
            }
        );

        PomodoroPanel.currentPanel = new PomodoroPanel(panel, extensionUri, statusBarItem);
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
            vscode.window.showInformationMessage(`🎉 Pomodoro #${this._sessionsCompleted} complete! Take a break.`);
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
            breakMinutes: this._breakMinutes
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
            padding: 20px;
        }
        .container {
            text-align: center;
            max-width: 400px;
            width: 100%;
        }
        .timer-display {
            font-size: 72px;
            font-weight: bold;
            margin: 30px 0;
            font-variant-numeric: tabular-nums;
        }
        .timer-display.break {
            color: var(--vscode-charts-green);
        }
        .timer-display.work {
            color: var(--vscode-charts-red);
        }
        .status {
            font-size: 24px;
            margin-bottom: 20px;
            opacity: 0.8;
        }
        .sessions {
            font-size: 16px;
            margin-bottom: 30px;
            opacity: 0.7;
        }
        .controls {
            display: flex;
            gap: 10px;
            justify-content: center;
            flex-wrap: wrap;
        }
        button {
            background: var(--vscode-button-background);
            color: var(--vscode-button-foreground);
            border: none;
            padding: 12px 24px;
            font-size: 16px;
            border-radius: 8px;
            cursor: pointer;
            transition: all 0.2s ease;
            font-weight: 500;
        }
        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
        }
        button:active {
            transform: translateY(0);
        }
        button.secondary {
            background: var(--vscode-button-secondaryBackground);
            color: var(--vscode-button-secondaryForeground);
        }
        button:disabled {
            opacity: 0.5;
            cursor: not-allowed;
            transform: none;
            box-shadow: none;
        }
        .settings {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid var(--vscode-widget-border);
        }
        .settings h3 {
            margin-bottom: 15px;
            font-size: 14px;
            opacity: 0.7;
        }
        .setting-row {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 10px;
            margin-bottom: 10px;
        }
        .setting-row label {
            min-width: 80px;
            text-align: right;
        }
        .setting-row input {
            width: 60px;
            padding: 6px;
            background: var(--vscode-input-background);
            color: var(--vscode-input-foreground);
            border: 1px solid var(--vscode-input-border);
            border-radius: 4px;
            text-align: center;
        }
        .progress-ring {
            margin: 20px auto;
        }
        .progress-ring circle {
            transition: stroke-dashoffset 1s linear;
        }
        .icon {
            font-size: 48px;
            margin-bottom: 10px;
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
