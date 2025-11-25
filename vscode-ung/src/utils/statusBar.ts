import * as vscode from 'vscode';
import { UngCli } from '../cli/ungCli';
import { Formatter } from './formatting';
import { Config } from './config';

/**
 * Status bar manager for active time tracking
 */
export class StatusBarManager {
    private statusBarItem: vscode.StatusBarItem;
    private cli: UngCli;
    private updateInterval: NodeJS.Timeout | null = null;
    private activeSessionData: any = null;

    constructor(cli: UngCli) {
        this.cli = cli;
        this.statusBarItem = vscode.window.createStatusBarItem(
            vscode.StatusBarAlignment.Left,
            100
        );
        this.statusBarItem.command = 'ung.viewActiveSession';
    }

    /**
     * Start monitoring active sessions
     */
    async start(): Promise<void> {
        if (!Config.shouldShowStatusBar()) {
            return;
        }

        // Update immediately
        await this.update();

        // Update every 5 seconds
        this.updateInterval = setInterval(() => {
            this.update();
        }, 5000);
    }

    /**
     * Stop monitoring
     */
    stop(): void {
        if (this.updateInterval) {
            clearInterval(this.updateInterval);
            this.updateInterval = null;
        }
        this.statusBarItem.hide();
    }

    /**
     * Update the status bar with current session
     */
    async update(): Promise<void> {
        if (!Config.shouldShowStatusBar()) {
            this.statusBarItem.hide();
            return;
        }

        const result = await this.cli.getCurrentSession();

        if (result.success && result.stdout) {
            // Parse the output to check if there's an active session
            const output = result.stdout;

            if (output.includes('No active tracking session')) {
                this.statusBarItem.hide();
                this.activeSessionData = null;
            } else {
                // Parse session data from output
                this.activeSessionData = this.parseSessionData(output);

                if (this.activeSessionData) {
                    const elapsed = this.calculateElapsed(this.activeSessionData.started);
                    const project = this.activeSessionData.project || 'Tracking';

                    this.statusBarItem.text = `$(clock) ${Formatter.formatDuration(elapsed)} - ${project}`;
                    this.statusBarItem.tooltip = `Active Time Tracking\nProject: ${project}\nStarted: ${this.activeSessionData.started}\nClick to view details`;
                    this.statusBarItem.show();
                } else {
                    this.statusBarItem.hide();
                }
            }
        } else {
            this.statusBarItem.hide();
            this.activeSessionData = null;
        }
    }

    /**
     * Parse session data from CLI output
     */
    private parseSessionData(output: string): any {
        const lines = output.split('\n');
        const data: any = {};

        for (const line of lines) {
            if (line.includes('Project:')) {
                data.project = line.split('Project:')[1].trim();
            } else if (line.includes('Started:')) {
                data.started = line.split('Started:')[1].trim();
            } else if (line.includes('Client:')) {
                data.client = line.split('Client:')[1].split('(')[0].trim();
            }
        }

        return data.started ? data : null;
    }

    /**
     * Calculate elapsed time in seconds
     */
    private calculateElapsed(startedStr: string): number {
        try {
            const started = new Date(startedStr);
            const now = new Date();
            return Math.floor((now.getTime() - started.getTime()) / 1000);
        } catch {
            return 0;
        }
    }

    /**
     * Force immediate update
     */
    async forceUpdate(): Promise<void> {
        await this.update();
    }

    /**
     * Dispose resources
     */
    dispose(): void {
        this.stop();
        this.statusBarItem.dispose();
    }
}
