import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';
import { getSecureStorage } from '../utils/secureStorage';

/**
 * Security commands for managing database encryption and password storage
 */
export class SecurityCommands {
  private cli: UngCli;
  private outputChannel: vscode.OutputChannel;

  constructor(cli: UngCli, outputChannel: vscode.OutputChannel) {
    this.cli = cli;
    this.outputChannel = outputChannel;
  }

  /**
   * Show security status (encryption status and password storage)
   */
  async showSecurityStatus(): Promise<void> {
    const result = await this.cli.getSecurityStatus();

    this.outputChannel.clear();
    this.outputChannel.appendLine('=== Security Status ===\n');

    if (result.success && result.stdout) {
      this.outputChannel.appendLine(result.stdout);
    } else {
      this.outputChannel.appendLine('Failed to get security status');
      if (result.error) {
        this.outputChannel.appendLine(`Error: ${result.error}`);
      }
    }

    // Also show VS Code secure storage status
    const secureStorage = getSecureStorage();
    const hasPassword = await secureStorage.hasPassword();

    this.outputChannel.appendLine('\n=== VS Code Password Storage ===\n');
    if (hasPassword) {
      this.outputChannel.appendLine('Password saved in VS Code secure storage');
    } else {
      this.outputChannel.appendLine(
        'No password saved in VS Code secure storage'
      );
    }

    this.outputChannel.show();
  }

  /**
   * Save password to VS Code secure storage
   */
  async savePassword(): Promise<void> {
    const secureStorage = getSecureStorage();

    // Check if password is already saved
    const hasPassword = await secureStorage.hasPassword();
    if (hasPassword) {
      const replace = await vscode.window.showQuickPick(
        [
          {
            label: '$(check) Replace',
            value: true,
            description: 'Replace the existing password',
          },
          {
            label: '$(x) Cancel',
            value: false,
            description: 'Keep the existing password',
          },
        ],
        {
          placeHolder:
            'A password is already saved. What would you like to do?',
          title: 'Replace Password',
        }
      );

      if (!replace?.value) {
        return;
      }
    }

    // Get the password
    const password = await vscode.window.showInputBox({
      prompt: 'Enter your database password',
      password: true,
      title: 'Save Database Password',
      ignoreFocusOut: true,
    });

    if (!password) {
      return;
    }

    // Verify the password works by running a command
    const verifying = vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Verifying password...',
      },
      async () => {
        // Try to run a simple command with the password
        const result = await this.cli.execWithPassword(
          ['company', 'list'],
          password
        );
        return result.success;
      }
    );

    const isValid = await verifying;
    if (!isValid) {
      vscode.window.showErrorMessage(
        'Invalid password. Please check and try again.'
      );
      return;
    }

    // Save the password
    await secureStorage.savePassword(password);
    vscode.window.showInformationMessage('Password saved securely in VS Code');
  }

  /**
   * Clear the saved password from VS Code secure storage
   */
  async clearPassword(): Promise<void> {
    const secureStorage = getSecureStorage();

    // Check if password is saved
    const hasPassword = await secureStorage.hasPassword();
    if (!hasPassword) {
      vscode.window.showInformationMessage('No password is saved in VS Code');
      return;
    }

    // Confirm deletion
    const confirm = await vscode.window.showQuickPick(
      [
        {
          label: '$(trash) Clear Password',
          value: true,
          description: 'Remove the saved password',
        },
        {
          label: '$(x) Cancel',
          value: false,
          description: 'Keep the password',
        },
      ],
      {
        placeHolder: 'Are you sure you want to clear the saved password?',
        title: 'Clear Password',
      }
    );

    if (!confirm?.value) {
      return;
    }

    await secureStorage.deletePassword();
    vscode.window.showInformationMessage(
      'Password cleared from VS Code secure storage'
    );
  }

  /**
   * Open security management menu
   */
  async manageSecuritySettings(): Promise<void> {
    const secureStorage = getSecureStorage();
    const hasPassword = await secureStorage.hasPassword();
    const isEncrypted = await this.cli.isEncrypted();

    const options: Array<{
      label: string;
      description: string;
      action: () => Promise<void>;
    }> = [
      {
        label: '$(info) View Security Status',
        description: 'Show encryption and password storage status',
        action: () => this.showSecurityStatus(),
      },
    ];

    if (isEncrypted) {
      if (hasPassword) {
        options.push({
          label: '$(key) Update Saved Password',
          description: 'Replace the saved password with a new one',
          action: () => this.savePassword(),
        });
        options.push({
          label: '$(trash) Clear Saved Password',
          description: 'Remove the password from VS Code storage',
          action: () => this.clearPassword(),
        });
      } else {
        options.push({
          label: '$(lock) Save Password',
          description: 'Save your database password for automatic use',
          action: () => this.savePassword(),
        });
      }
    } else {
      options.push({
        label: '$(warning) Database Not Encrypted',
        description:
          'Run "ung security enable" in terminal to enable encryption',
        action: async () => {
          const terminal = vscode.window.createTerminal('UNG Security');
          terminal.show();
          terminal.sendText('ung security enable');
        },
      });
    }

    const selected = await vscode.window.showQuickPick(options, {
      placeHolder: 'Security Settings',
      title: 'Database Security',
    });

    if (selected) {
      await selected.action();
    }
  }
}
