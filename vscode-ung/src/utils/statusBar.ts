import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';
import { Config } from './config';
import { Formatter } from './formatting';

/**
 * Status bar manager for active time tracking
 */
export class StatusBarManager {
  private statusBarItem: vscode.StatusBarItem;
  private earningsItem: vscode.StatusBarItem;
  private quickActionItem: vscode.StatusBarItem;
  private cli: UngCli;
  private updateInterval: NodeJS.Timeout | null = null;
  private activeSessionData: {
    project?: string;
    started?: string;
    client?: string;
  } | null = null;
  private isTracking: boolean = false;

  constructor(cli: UngCli) {
    this.cli = cli;

    // Main tracking status bar item
    this.statusBarItem = vscode.window.createStatusBarItem(
      vscode.StatusBarAlignment.Left,
      100
    );
    this.statusBarItem.command = 'ung.toggleTracking';

    // Earnings today item
    this.earningsItem = vscode.window.createStatusBarItem(
      vscode.StatusBarAlignment.Left,
      99
    );
    this.earningsItem.command = 'ung.openStatistics';

    // Quick action button
    this.quickActionItem = vscode.window.createStatusBarItem(
      vscode.StatusBarAlignment.Left,
      98
    );
    this.quickActionItem.text = '$(add)';
    this.quickActionItem.tooltip = 'UNG Quick Actions';
    this.quickActionItem.command = 'ung.quickActions';
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
    await this.updateEarnings();

    // Show quick action button
    this.quickActionItem.show();

    // Update every 5 seconds
    this.updateInterval = setInterval(() => {
      this.update();
    }, 5000);

    // Update earnings every minute
    setInterval(() => {
      this.updateEarnings();
    }, 60000);
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
    this.earningsItem.hide();
    this.quickActionItem.hide();
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
        this.isTracking = false;
        this.statusBarItem.text = '$(play) Start Tracking';
        this.statusBarItem.tooltip = 'Click to start time tracking';
        this.statusBarItem.backgroundColor = undefined;
        this.statusBarItem.show();
        this.activeSessionData = null;
      } else {
        // Parse session data from output
        this.activeSessionData = this.parseSessionData(output);
        this.isTracking = true;

        if (this.activeSessionData?.started) {
          const elapsed = this.calculateElapsed(this.activeSessionData.started);
          const project = this.activeSessionData.project || 'Tracking';

          this.statusBarItem.text = `$(debug-stop) ${Formatter.formatDuration(elapsed)} - ${project}`;
          this.statusBarItem.tooltip = `Active Time Tracking\nProject: ${project}\nStarted: ${this.activeSessionData.started}\nClick to stop`;
          this.statusBarItem.backgroundColor = new vscode.ThemeColor(
            'statusBarItem.warningBackground'
          );
          this.statusBarItem.show();
        } else {
          this.isTracking = false;
          this.statusBarItem.text = '$(play) Start Tracking';
          this.statusBarItem.tooltip = 'Click to start time tracking';
          this.statusBarItem.backgroundColor = undefined;
          this.statusBarItem.show();
        }
      }
    } else {
      this.isTracking = false;
      this.statusBarItem.text = '$(play) Start Tracking';
      this.statusBarItem.tooltip = 'Click to start time tracking';
      this.statusBarItem.backgroundColor = undefined;
      this.statusBarItem.show();
      this.activeSessionData = null;
    }
  }

  /**
   * Update earnings display
   */
  async updateEarnings(): Promise<void> {
    if (!Config.shouldShowStatusBar()) {
      this.earningsItem.hide();
      return;
    }

    try {
      const result = await this.cli.listTimeEntries();
      if (result.success && result.stdout) {
        const today = new Date().toISOString().split('T')[0];
        const lines = result.stdout.trim().split('\n');

        let todayHours = 0;

        for (const line of lines.slice(1)) {
          const parts = line.split(/\s{2,}/).filter((p: string) => p.trim());
          if (parts.length >= 3) {
            const dateStr = parts[1] || '';
            if (dateStr.includes(today.slice(5))) {
              // Match MM-DD
              const hoursMatch = parts[2]?.match(/([\d.]+)\s*h/);
              if (hoursMatch) {
                todayHours += parseFloat(hoursMatch[1]);
              }
            }
          }
        }

        if (todayHours > 0) {
          this.earningsItem.text = `$(graph) ${todayHours.toFixed(1)}h today`;
          this.earningsItem.tooltip = `Hours tracked today: ${todayHours.toFixed(1)}h\nClick to view statistics`;
          this.earningsItem.show();
        } else {
          this.earningsItem.text = '$(graph) 0h today';
          this.earningsItem.tooltip =
            'No time tracked today\nClick to view statistics';
          this.earningsItem.show();
        }
      }
    } catch {
      this.earningsItem.hide();
    }
  }

  /**
   * Check if currently tracking
   */
  getIsTracking(): boolean {
    return this.isTracking;
  }

  /**
   * Parse session data from CLI output
   */
  private parseSessionData(
    output: string
  ): { project?: string; started?: string; client?: string } | null {
    const lines = output.split('\n');
    const data: { project?: string; started?: string; client?: string } = {};

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
    this.earningsItem.dispose();
    this.quickActionItem.dispose();
  }
}
