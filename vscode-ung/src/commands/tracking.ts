import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';
import type { StatusBarManager } from '../utils/statusBar';

/**
 * Time tracking command handlers
 */
export class TrackingCommands {
  constructor(
    private cli: UngCli,
    private statusBar: StatusBarManager,
    private refreshCallback?: () => void
  ) {}

  /**
   * Start time tracking
   */
  async startTracking(): Promise<void> {
    // Check if there's already an active session
    const currentResult = await this.cli.getCurrentSession();
    if (
      currentResult.success &&
      currentResult.stdout &&
      !currentResult.stdout.includes('No active')
    ) {
      vscode.window.showWarningMessage(
        'There is already an active tracking session. Stop it first.'
      );
      return;
    }

    const project = await vscode.window.showInputBox({
      prompt: 'Project Name',
      placeHolder: 'e.g., Website Development',
    });

    const clientIdStr = await vscode.window.showInputBox({
      prompt: 'Client ID (optional)',
      placeHolder: 'Leave empty for no client',
    });

    const billable = await vscode.window.showQuickPick(['Yes', 'No'], {
      placeHolder: 'Is this billable?',
    });

    const notes = await vscode.window.showInputBox({
      prompt: 'Notes (optional)',
      placeHolder: 'What are you working on?',
    });

    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Starting time tracking...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.startTracking({
          clientId: clientIdStr ? Number(clientIdStr) : undefined,
          project,
          billable: billable === 'Yes',
          notes,
        });

        if (result.success) {
          vscode.window.showInformationMessage('Time tracking started!');
          await this.statusBar.forceUpdate();
          if (this.refreshCallback) {
            this.refreshCallback();
          }
        } else {
          vscode.window.showErrorMessage(
            `Failed to start tracking: ${result.error}`
          );
        }
      }
    );
  }

  /**
   * Stop time tracking
   */
  async stopTracking(): Promise<void> {
    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Stopping time tracking...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.stopTracking();

        if (result.success) {
          vscode.window.showInformationMessage('Time tracking stopped!');
          await this.statusBar.forceUpdate();
          if (this.refreshCallback) {
            this.refreshCallback();
          }
        } else {
          vscode.window.showErrorMessage(
            `Failed to stop tracking: ${result.error}`
          );
        }
      }
    );
  }

  /**
   * Log time manually
   */
  async logTimeManually(): Promise<void> {
    // Get contracts to select from
    const contractResult = await this.cli.listContracts();
    if (!contractResult.success) {
      vscode.window.showErrorMessage('Failed to fetch contracts');
      return;
    }

    // Parse contracts from CLI output
    const contracts = this.parseContractsFromOutput(
      contractResult.stdout || ''
    );
    if (contracts.length === 0) {
      vscode.window.showErrorMessage(
        'No contracts found. Create one first with "ung contract add"'
      );
      return;
    }

    // Show contract dropdown with client name and rate info
    const contractItems = contracts.map((c) => ({
      label: c.client,
      description: `${c.type} - ${c.ratePrice}`,
      detail: c.name,
      contract: c,
    }));

    const selectedContract = await vscode.window.showQuickPick(contractItems, {
      placeHolder: 'Select a contract to log time for',
      matchOnDescription: true,
      matchOnDetail: true,
    });

    if (!selectedContract) return;

    // Ask for hours
    const hoursStr = await vscode.window.showInputBox({
      prompt: `Hours worked for ${selectedContract.contract.client}`,
      placeHolder: 'e.g., 2.5',
      validateInput: (value) => {
        if (!value) return 'Hours is required';
        if (Number.isNaN(Number(value))) return 'Must be a valid number';
        if (Number(value) <= 0) return 'Hours must be greater than 0';
        return null;
      },
    });

    if (!hoursStr) return;

    // Ask for project name (optional)
    const project = await vscode.window.showInputBox({
      prompt: 'Project name (optional)',
      placeHolder: 'e.g., Website Development',
    });

    // Ask for notes (optional)
    const notes = await vscode.window.showInputBox({
      prompt: 'Notes (optional)',
      placeHolder: 'What did you work on?',
    });

    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Logging time...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.logTime({
          contractId: selectedContract.contract.id,
          hours: Number(hoursStr),
          project: project || undefined,
          notes: notes || undefined,
        });

        if (result.success) {
          vscode.window.showInformationMessage(
            `Logged ${hoursStr} hours for ${selectedContract.contract.client}!`
          );
          if (this.refreshCallback) {
            this.refreshCallback();
          }
        } else {
          vscode.window.showErrorMessage(`Failed to log time: ${result.error}`);
        }
      }
    );
  }

  /**
   * Parse contracts from CLI output (tabular format)
   */
  private parseContractsFromOutput(output: string): Array<{
    id: number;
    contractNum: string;
    name: string;
    client: string;
    type: string;
    ratePrice: string;
    active: boolean;
  }> {
    const lines = output.trim().split('\n');
    if (lines.length < 2) return [];

    // Skip header line
    const dataLines = lines.slice(1);
    const contracts: Array<{
      id: number;
      contractNum: string;
      name: string;
      client: string;
      type: string;
      ratePrice: string;
      active: boolean;
    }> = [];

    for (const line of dataLines) {
      // Parse: ID  CONTRACT#  NAME  CLIENT  TYPE  RATE/PRICE  ACTIVE
      const parts = line
        .split(/\s{2,}/)
        .map((p) => p.trim())
        .filter((p) => p);
      if (parts.length >= 6) {
        const id = parseInt(parts[0], 10);
        if (!Number.isNaN(id)) {
          contracts.push({
            id,
            contractNum: parts[1],
            name: parts[2],
            client: parts[3],
            type: parts[4],
            ratePrice: parts[5],
            active: parts[6] === '✓',
          });
        }
      }
    }

    return contracts;
  }

  /**
   * View active tracking session
   */
  async viewActiveSession(): Promise<void> {
    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Fetching active session...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.getCurrentSession();

        if (result.success && result.stdout) {
          // Show output in a new document
          const doc = await vscode.workspace.openTextDocument({
            content: result.stdout,
            language: 'plaintext',
          });
          await vscode.window.showTextDocument(doc);
        } else {
          vscode.window.showErrorMessage(
            `Failed to fetch session: ${result.error}`
          );
        }
      }
    );
  }

  /**
   * View tracking session with action menu
   */
  async viewTrackingSession(sessionId?: number): Promise<void> {
    if (!sessionId) {
      vscode.window.showErrorMessage('No session selected');
      return;
    }

    const actions = [
      { label: '$(edit) Edit Session', action: 'edit' },
      { label: '$(trash) Delete Session', action: 'delete' },
      { label: '$(close) Close', action: 'close' },
    ];

    const selected = await vscode.window.showQuickPick(actions, {
      placeHolder: `Session #${sessionId}`,
      title: 'Session Actions',
    });

    if (selected) {
      switch (selected.action) {
        case 'edit':
          await this.editTrackingSession(sessionId);
          break;
        case 'delete':
          await this.deleteTrackingSession(sessionId);
          break;
      }
    }
  }

  /**
   * Edit a tracking session
   */
  async editTrackingSession(sessionId?: number): Promise<void> {
    if (!sessionId) {
      vscode.window.showErrorMessage('No session selected');
      return;
    }

    // Ask what to edit
    const editOptions = [
      { label: '$(symbol-number) Hours', value: 'hours' },
      { label: '$(folder) Project Name', value: 'project' },
      { label: '$(note) Notes', value: 'notes' },
    ];

    const selected = await vscode.window.showQuickPick(editOptions, {
      placeHolder: 'What would you like to edit?',
      title: `Edit Session #${sessionId}`,
    });

    if (!selected) return;

    const editData: { hours?: number; project?: string; notes?: string } = {};

    switch (selected.value) {
      case 'hours': {
        const newHours = await vscode.window.showInputBox({
          prompt: 'New hours value',
          placeHolder: 'e.g., 2.5',
          validateInput: (value) => {
            if (!value) return 'Hours is required';
            if (Number.isNaN(Number(value))) return 'Must be a valid number';
            if (Number(value) <= 0) return 'Hours must be greater than 0';
            return null;
          },
        });
        if (!newHours) return;
        editData.hours = Number(newHours);
        break;
      }
      case 'project': {
        const newProject = await vscode.window.showInputBox({
          prompt: 'New project name',
          placeHolder: 'e.g., Website Development',
        });
        if (!newProject) return;
        editData.project = newProject;
        break;
      }
      case 'notes': {
        const newNotes = await vscode.window.showInputBox({
          prompt: 'New notes',
          placeHolder: 'What did you work on?',
        });
        if (!newNotes) return;
        editData.notes = newNotes;
        break;
      }
    }

    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Updating tracking session...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.editTrackingSession(sessionId, editData);

        if (result.success) {
          vscode.window.showInformationMessage('Session updated successfully!');
          if (this.refreshCallback) {
            this.refreshCallback();
          }
        } else {
          vscode.window.showErrorMessage(
            `Failed to update session: ${result.error}`
          );
        }
      }
    );
  }

  /**
   * Delete a tracking session
   */
  async deleteTrackingSession(sessionId?: number): Promise<void> {
    if (!sessionId) {
      vscode.window.showErrorMessage('No session selected');
      return;
    }

    // Get session details for potential revert
    const sessionsResult = await this.cli.listTrackingSessions();
    let sessionData: {
      project: string;
      client: string;
      hours: number;
      billable: boolean;
      notes?: string;
    } | null = null;

    if (sessionsResult.success && sessionsResult.stdout) {
      const session = this.parseSessionFromOutput(
        sessionsResult.stdout,
        sessionId
      );
      if (session) {
        sessionData = session;
      }
    }

    const confirm = await vscode.window.showWarningMessage(
      sessionData
        ? `Delete tracking session "${sessionData.project}"? This is a soft delete.`
        : `Delete tracking session #${sessionId}? This is a soft delete.`,
      { modal: true },
      'Yes, Delete',
      'Cancel'
    );

    if (confirm !== 'Yes, Delete') return;

    let deleteSuccess = false;
    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Deleting session...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.deleteTrackingSession(sessionId);

        if (result.success) {
          deleteSuccess = true;
          if (this.refreshCallback) {
            this.refreshCallback();
          }
        } else {
          vscode.window.showErrorMessage(
            `Failed to delete session: ${result.error}`
          );
        }
      }
    );

    // Show success message with revert option
    if (deleteSuccess && sessionData) {
      const action = await vscode.window.showInformationMessage(
        `Session "${sessionData.project}" deleted successfully!`,
        'Revert'
      );

      if (action === 'Revert') {
        await this.revertSessionDeletion(sessionData);
      }
    } else if (deleteSuccess) {
      vscode.window.showInformationMessage('Session deleted successfully!');
    }
  }

  /**
   * Parse a specific session from CLI output
   */
  private parseSessionFromOutput(
    output: string,
    sessionId: number
  ): {
    project: string;
    client: string;
    hours: number;
    billable: boolean;
    notes?: string;
  } | null {
    const lines = output.split('\n').filter((line) => line.trim());

    for (let i = 1; i < lines.length; i++) {
      const line = lines[i].trim();
      if (!line) continue;

      const parts = line.split(/\s{2,}/);
      if (parts.length >= 6) {
        const id = parseInt(parts[0], 10);
        if (id === sessionId) {
          const project = parts[1] || 'Untitled';
          const client = parts[2] || '';
          const duration = parts[4];
          const billable =
            parts[5]?.toLowerCase() === 'yes' ||
            parts[5]?.toLowerCase() === 'true';

          // Parse duration to hours
          let hours = 0;
          const hourMatch = duration.match(/(\d+(?:\.\d+)?)\s*h/i);
          const minMatch = duration.match(/(\d+)\s*m/i);
          const colonMatch = duration.match(/(\d+):(\d+)/);

          if (colonMatch) {
            hours =
              parseInt(colonMatch[1], 10) + parseInt(colonMatch[2], 10) / 60;
          } else {
            if (hourMatch) {
              hours += parseFloat(hourMatch[1]);
            }
            if (minMatch) {
              hours += parseInt(minMatch[1], 10) / 60;
            }
          }

          return {
            project,
            client,
            hours: Math.round(hours * 100) / 100,
            billable,
          };
        }
      }
    }

    return null;
  }

  /**
   * Revert a session deletion by re-logging the time
   */
  private async revertSessionDeletion(sessionData: {
    project: string;
    client: string;
    hours: number;
    billable: boolean;
    notes?: string;
  }): Promise<void> {
    // First, we need to find the contract ID from the client name
    const contractsResult = await this.cli.listContracts();
    if (!contractsResult.success || !contractsResult.stdout) {
      vscode.window.showErrorMessage(
        'Failed to revert: Could not fetch contracts'
      );
      return;
    }

    // Parse contracts to find one matching the client
    const contracts = this.parseContractsFromOutput(contractsResult.stdout);
    const matchingContract = contracts.find(
      (c) => c.client === sessionData.client
    );

    if (!matchingContract) {
      vscode.window.showErrorMessage(
        `Failed to revert: No contract found for client "${sessionData.client}"`
      );
      return;
    }

    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Reverting session deletion...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.logTime({
          contractId: matchingContract.id,
          hours: sessionData.hours,
          project: sessionData.project,
          notes: sessionData.notes,
        });

        if (result.success) {
          vscode.window.showInformationMessage(
            `Session "${sessionData.project}" restored successfully!`
          );
          if (this.refreshCallback) {
            this.refreshCallback();
          }
        } else {
          vscode.window.showErrorMessage(
            `Failed to restore session: ${result.error}`
          );
        }
      }
    );
  }
}
