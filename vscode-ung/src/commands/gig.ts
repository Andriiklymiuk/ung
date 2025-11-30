import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';

/**
 * Gig status configuration
 * Simplified flow: todo → in_progress → sent → done
 */
const GIG_STATUSES = [
  {
    value: 'todo',
    label: 'Todo',
    icon: 'circle-outline',
    color: 'charts.gray',
  },
  {
    value: 'in_progress',
    label: 'In Progress',
    icon: 'play-circle',
    color: 'charts.blue',
  },
  {
    value: 'sent',
    label: 'Sent',
    icon: 'package',
    color: 'charts.orange',
  },
  {
    value: 'done',
    label: 'Done',
    icon: 'pass-filled',
    color: 'charts.green',
  },
  {
    value: 'on_hold',
    label: 'On Hold',
    icon: 'debug-pause',
    color: 'charts.yellow',
  },
  {
    value: 'cancelled',
    label: 'Cancelled',
    icon: 'circle-slash',
    color: 'charts.red',
  },
];

const GIG_TYPES = [
  { value: 'hourly', label: 'Hourly', description: 'Billed by the hour' },
  { value: 'fixed', label: 'Fixed Price', description: 'One-time project fee' },
  { value: 'retainer', label: 'Retainer', description: 'Monthly recurring' },
];

/**
 * Parsed gig from CLI output
 */
interface ParsedGig {
  id: number;
  name: string;
  client: string;
  status: string;
  hours: number;
  type: string;
  project: string;
}

/**
 * Gig command handlers
 */
export class GigCommands {
  constructor(
    private cli: UngCli,
    private refreshCallback?: () => void
  ) {}

  /**
   * Create a new gig
   */
  async createGig(): Promise<void> {
    // Get gig name
    const name = await vscode.window.showInputBox({
      prompt: 'Gig Name',
      placeHolder: 'e.g., Website Redesign for Acme',
      validateInput: (value) => (value ? null : 'Name is required'),
    });

    if (!name) return;

    // Select client (optional)
    const clientResult = await this.cli.listClients();
    let clientId: number | undefined;

    if (clientResult.success && clientResult.stdout) {
      const clients = this.parseClientsFromOutput(clientResult.stdout);
      if (clients.length > 0) {
        const clientItems = [
          {
            label: '$(dash) No Client',
            description: 'Skip client selection',
            id: undefined,
          },
          ...clients.map((c) => ({
            label: `$(person) ${c.name}`,
            description: c.email,
            id: c.id,
          })),
        ];

        const selectedClient = await vscode.window.showQuickPick(clientItems, {
          placeHolder: 'Select a client (optional)',
        });

        if (selectedClient?.id) {
          clientId = selectedClient.id;
        }
      }
    }

    // Select gig type
    const gigType = await vscode.window.showQuickPick(
      GIG_TYPES.map((t) => ({
        label: `$(symbol-enum) ${t.label}`,
        description: t.description,
        value: t.value,
      })),
      { placeHolder: 'Select gig type' }
    );

    if (!gigType) return;

    // Get hourly rate if hourly type
    let rate: number | undefined;
    if (gigType.value === 'hourly') {
      const rateStr = await vscode.window.showInputBox({
        prompt: 'Hourly Rate (optional)',
        placeHolder: 'e.g., 150',
        validateInput: (value) => {
          if (!value) return null;
          if (Number.isNaN(Number(value))) return 'Must be a valid number';
          if (Number(value) < 0) return 'Rate must be positive';
          return null;
        },
      });
      if (rateStr) {
        rate = Number(rateStr);
      }
    }

    // Create the gig
    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Creating gig...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.createGig({
          name,
          clientId,
          type: gigType.value,
          rate,
        });

        if (result.success) {
          vscode.window.showInformationMessage(`Gig "${name}" created!`);
          this.refreshCallback?.();
        } else {
          vscode.window.showErrorMessage(
            `Failed to create gig: ${result.error}`
          );
        }
      }
    );
  }

  /**
   * View gig details and actions
   */
  async viewGig(gigId?: number): Promise<void> {
    if (!gigId) {
      // Show gig list to select
      const gig = await this.selectGig('Select a gig to view');
      if (!gig) return;
      gigId = gig.id;
    }

    const result = await this.cli.showGig(gigId);
    if (!result.success) {
      vscode.window.showErrorMessage(`Failed to load gig: ${result.error}`);
      return;
    }

    // Show action menu
    const actions = [
      { label: '$(arrow-right) Move to...', action: 'move' },
      { label: '$(add) Add Task', action: 'addTask' },
      { label: '$(edit) Edit Gig', action: 'edit' },
      { label: '$(trash) Delete Gig', action: 'delete' },
      { label: '$(close) Close', action: 'close' },
    ];

    const selected = await vscode.window.showQuickPick(actions, {
      placeHolder: `Gig #${gigId}`,
      title: 'Gig Actions',
    });

    if (selected) {
      switch (selected.action) {
        case 'move':
          await this.moveGig(gigId);
          break;
        case 'addTask':
          await this.addTask(gigId);
          break;
        case 'edit':
          await this.editGig(gigId);
          break;
        case 'delete':
          await this.deleteGig(gigId);
          break;
      }
    }
  }

  /**
   * Move gig to a different status
   */
  async moveGig(gigId?: number): Promise<void> {
    if (!gigId) {
      const gig = await this.selectGig('Select a gig to move');
      if (!gig) return;
      gigId = gig.id;
    }

    const statusItems = GIG_STATUSES.map((s) => ({
      label: `$(${s.icon}) ${s.label}`,
      value: s.value,
    }));

    const selected = await vscode.window.showQuickPick(statusItems, {
      placeHolder: 'Move to status',
      title: 'Move Gig',
    });

    if (!selected) return;

    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: `Moving gig to ${selected.label}...`,
        cancellable: false,
      },
      async () => {
        const result = await this.cli.moveGig(gigId!, selected.value);
        if (result.success) {
          vscode.window.showInformationMessage(
            `Gig moved to ${selected.label}!`
          );
          this.refreshCallback?.();
        } else {
          vscode.window.showErrorMessage(`Failed to move gig: ${result.error}`);
        }
      }
    );
  }

  /**
   * Edit a gig
   */
  async editGig(gigId?: number): Promise<void> {
    if (!gigId) {
      const gig = await this.selectGig('Select a gig to edit');
      if (!gig) return;
      gigId = gig.id;
    }

    vscode.window.showInformationMessage(
      `Edit gig functionality coming soon. Use CLI: ung gig show ${gigId}`
    );
  }

  /**
   * Delete a gig
   */
  async deleteGig(gigId?: number): Promise<void> {
    if (!gigId) {
      const gig = await this.selectGig('Select a gig to delete');
      if (!gig) return;
      gigId = gig.id;
    }

    const confirm = await vscode.window.showWarningMessage(
      `Delete gig #${gigId}? This cannot be undone!`,
      { modal: true },
      'Yes, Delete',
      'Cancel'
    );

    if (confirm !== 'Yes, Delete') return;

    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Deleting gig...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.deleteGig(gigId!);
        if (result.success) {
          vscode.window.showInformationMessage('Gig deleted!');
          this.refreshCallback?.();
        } else {
          vscode.window.showErrorMessage(
            `Failed to delete gig: ${result.error}`
          );
        }
      }
    );
  }

  /**
   * Add a task to a gig
   */
  async addTask(gigId?: number): Promise<void> {
    if (!gigId) {
      const gig = await this.selectGig('Select a gig to add task to');
      if (!gig) return;
      gigId = gig.id;
    }

    const title = await vscode.window.showInputBox({
      prompt: 'Task Title',
      placeHolder: 'e.g., Design mockups',
      validateInput: (value) => (value ? null : 'Title is required'),
    });

    if (!title) return;

    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: 'Adding task...',
        cancellable: false,
      },
      async () => {
        const result = await this.cli.addGigTask(gigId!, title);
        if (result.success) {
          vscode.window.showInformationMessage(`Task "${title}" added!`);
          this.refreshCallback?.();
        } else {
          vscode.window.showErrorMessage(`Failed to add task: ${result.error}`);
        }
      }
    );
  }

  /**
   * Toggle task completion
   */
  async toggleTask(taskId: number): Promise<void> {
    const result = await this.cli.toggleGigTask(taskId);
    if (result.success) {
      this.refreshCallback?.();
    } else {
      vscode.window.showErrorMessage(`Failed to toggle task: ${result.error}`);
    }
  }

  /**
   * Filter gigs by status
   */
  async filterGigs(): Promise<void> {
    const statusItems = [
      { label: '$(list-flat) All Gigs', value: undefined },
      ...GIG_STATUSES.map((s) => ({
        label: `$(${s.icon}) ${s.label}`,
        value: s.value,
      })),
    ];

    const selected = await vscode.window.showQuickPick(statusItems, {
      placeHolder: 'Filter by status',
      title: 'Filter Gigs',
    });

    if (selected) {
      // This would be used by the panel to filter
      vscode.commands.executeCommand('ung.filterGigsBy', selected.value);
    }
  }

  /**
   * Select a gig from the list
   */
  private async selectGig(placeholder: string): Promise<ParsedGig | undefined> {
    const result = await this.cli.listGigs();
    if (!result.success || !result.stdout) {
      vscode.window.showErrorMessage('Failed to load gigs');
      return undefined;
    }

    const gigs = this.parseGigsFromOutput(result.stdout);
    if (gigs.length === 0) {
      vscode.window.showInformationMessage('No gigs found. Create one first!');
      return undefined;
    }

    const items = gigs.map((g) => ({
      label: `$(briefcase) ${g.name}`,
      description: `${g.client || 'No client'} - ${g.status}`,
      detail: `${g.hours}h tracked - ${g.type}`,
      gig: g,
    }));

    const selected = await vscode.window.showQuickPick(items, {
      placeHolder: placeholder,
      matchOnDescription: true,
      matchOnDetail: true,
    });

    return selected?.gig;
  }

  /**
   * Parse gigs from CLI output
   */
  parseGigsFromOutput(output: string): ParsedGig[] {
    const lines = output.trim().split('\n');
    if (lines.length < 2) return [];

    const gigs: ParsedGig[] = [];

    for (let i = 1; i < lines.length; i++) {
      const line = lines[i].trim();
      if (!line || line.startsWith('--')) continue;

      // CLI output columns: ID, NAME, CLIENT, PROJECT, STATUS, HOURS
      const parts = line.split(/\s{2,}/).filter((p) => p.trim());
      if (parts.length >= 5) {
        const id = parseInt(parts[0], 10);
        if (!Number.isNaN(id)) {
          gigs.push({
            id,
            name: parts[1] || '',
            client: parts[2] || '-',
            project: parts[3] || '-',
            status: parts[4] || 'todo',
            hours: parseFloat(parts[5]) || 0,
            type: 'hourly', // Default type, not in CLI output
          });
        }
      }
    }

    return gigs;
  }

  /**
   * Parse clients from CLI output
   */
  private parseClientsFromOutput(
    output: string
  ): Array<{ id: number; name: string; email: string }> {
    const lines = output.trim().split('\n');
    if (lines.length < 2) return [];

    const clients: Array<{ id: number; name: string; email: string }> = [];

    for (let i = 1; i < lines.length; i++) {
      const line = lines[i].trim();
      if (!line) continue;

      const parts = line.split(/\s{2,}/).filter((p) => p.trim());
      if (parts.length >= 2) {
        const id = parseInt(parts[0], 10);
        if (!Number.isNaN(id)) {
          clients.push({
            id,
            name: parts[1] || '',
            email: parts[2] || '',
          });
        }
      }
    }

    return clients;
  }
}

export { GIG_STATUSES, GIG_TYPES };
