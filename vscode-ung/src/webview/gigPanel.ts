import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';

/**
 * Gig Kanban Board Panel
 */
export class GigPanel {
  public static currentPanel: GigPanel | undefined;
  private static readonly viewType = 'ungGigBoard';

  private readonly _panel: vscode.WebviewPanel;
  private readonly _cli: UngCli;
  private _disposables: vscode.Disposable[] = [];
  private static _onChangeCallback: (() => void) | undefined;

  public static setOnChangeCallback(callback: () => void): void {
    GigPanel._onChangeCallback = callback;
  }

  public static createOrShow(cli: UngCli): void {
    const column = vscode.window.activeTextEditor
      ? vscode.window.activeTextEditor.viewColumn
      : undefined;

    if (GigPanel.currentPanel) {
      GigPanel.currentPanel._panel.reveal(column);
      GigPanel.currentPanel.refresh();
      return;
    }

    const panel = vscode.window.createWebviewPanel(
      GigPanel.viewType,
      'Gig Board',
      column || vscode.ViewColumn.One,
      {
        enableScripts: true,
        retainContextWhenHidden: true,
      }
    );

    GigPanel.currentPanel = new GigPanel(panel, cli);
  }

  private constructor(panel: vscode.WebviewPanel, cli: UngCli) {
    this._panel = panel;
    this._cli = cli;

    this._update();

    this._panel.onDidDispose(() => this.dispose(), null, this._disposables);

    this._panel.webview.onDidReceiveMessage(
      async (message) => {
        switch (message.command) {
          case 'moveGig':
            await this.moveGig(message.gigId, message.newStatus);
            break;
          case 'createGig':
            await this.createGig(message.name, message.status);
            break;
          case 'deleteGig':
            await this.deleteGig(message.gigId);
            break;
          case 'addTask':
            await this.addTask(message.gigId, message.title);
            break;
          case 'toggleTask':
            await this.toggleTask(message.taskId);
            break;
          case 'refresh':
            await this.refresh();
            break;
        }
      },
      null,
      this._disposables
    );
  }

  private async moveGig(gigId: number, newStatus: string): Promise<void> {
    const result = await this._cli.moveGig(gigId, newStatus);
    if (result.success) {
      GigPanel._onChangeCallback?.();
      await this.refresh();
    } else {
      this._panel.webview.postMessage({
        command: 'error',
        message: `Failed to move gig: ${result.error}`,
      });
    }
  }

  private async createGig(name: string, status: string): Promise<void> {
    const result = await this._cli.createGig({ name, status });
    if (result.success) {
      GigPanel._onChangeCallback?.();
      await this.refresh();
    } else {
      this._panel.webview.postMessage({
        command: 'error',
        message: `Failed to create gig: ${result.error}`,
      });
    }
  }

  private async deleteGig(gigId: number): Promise<void> {
    const result = await this._cli.deleteGig(gigId);
    if (result.success) {
      GigPanel._onChangeCallback?.();
      await this.refresh();
    } else {
      this._panel.webview.postMessage({
        command: 'error',
        message: `Failed to delete gig: ${result.error}`,
      });
    }
  }

  private async addTask(gigId: number, title: string): Promise<void> {
    const result = await this._cli.addGigTask(gigId, title);
    if (result.success) {
      await this.refresh();
    } else {
      this._panel.webview.postMessage({
        command: 'error',
        message: `Failed to add task: ${result.error}`,
      });
    }
  }

  private async toggleTask(taskId: number): Promise<void> {
    const result = await this._cli.toggleGigTask(taskId);
    if (result.success) {
      await this.refresh();
    }
  }

  public async refresh(): Promise<void> {
    this._update();
  }

  private async _update(): Promise<void> {
    const webview = this._panel.webview;
    const gigs = await this.loadGigs();
    webview.html = this._getHtmlForWebview(webview, gigs);
  }

  private async loadGigs(): Promise<
    Array<{
      id: number;
      name: string;
      client: string;
      status: string;
      hours: number;
      type: string;
    }>
  > {
    const result = await this._cli.listGigs();
    if (!result.success || !result.stdout) {
      return [];
    }

    const lines = result.stdout.trim().split('\n');
    if (lines.length < 2) return [];

    const gigs: Array<{
      id: number;
      name: string;
      client: string;
      status: string;
      hours: number;
      type: string;
    }> = [];

    for (let i = 1; i < lines.length; i++) {
      const line = lines[i].trim();
      if (!line || line.startsWith('--')) continue;

      const parts = line.split(/\s{2,}/).filter((p) => p.trim());
      if (parts.length >= 5) {
        const id = parseInt(parts[0], 10);
        if (!Number.isNaN(id)) {
          gigs.push({
            id,
            name: parts[1] || '',
            client: parts[2] || '-',
            status: parts[3] || 'pipeline',
            hours: parseFloat(parts[4]) || 0,
            type: parts[5] || 'hourly',
          });
        }
      }
    }

    return gigs;
  }

  private _getHtmlForWebview(
    _webview: vscode.Webview,
    gigs: Array<{
      id: number;
      name: string;
      client: string;
      status: string;
      hours: number;
      type: string;
    }>
  ): string {
    const columns = [
      { key: 'pipeline', label: 'Pipeline', icon: '📋', color: '#6b7280' },
      {
        key: 'negotiating',
        label: 'Negotiating',
        icon: '💬',
        color: '#a855f7',
      },
      { key: 'active', label: 'Active', icon: '🚀', color: '#3b82f6' },
      { key: 'delivered', label: 'Delivered', icon: '📦', color: '#f97316' },
      { key: 'invoiced', label: 'Invoiced', icon: '💵', color: '#06b6d4' },
      { key: 'complete', label: 'Complete', icon: '✅', color: '#22c55e' },
    ];

    return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Gig Board</title>
  <style>
    :root {
      --bg-primary: var(--vscode-editor-background);
      --bg-secondary: var(--vscode-sideBar-background);
      --bg-card: var(--vscode-input-background);
      --text-primary: var(--vscode-editor-foreground);
      --text-secondary: var(--vscode-descriptionForeground);
      --border: var(--vscode-panel-border);
      --accent: var(--vscode-focusBorder);
      --hover: var(--vscode-list-hoverBackground);
    }

    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }

    body {
      font-family: var(--vscode-font-family);
      font-size: var(--vscode-font-size);
      background: var(--bg-primary);
      color: var(--text-primary);
      overflow-x: auto;
      height: 100vh;
    }

    .header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 16px 24px;
      border-bottom: 1px solid var(--border);
      background: var(--bg-secondary);
    }

    .header h1 {
      font-size: 20px;
      font-weight: 600;
      display: flex;
      align-items: center;
      gap: 8px;
    }

    .header-actions {
      display: flex;
      gap: 8px;
    }

    .btn {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 8px 16px;
      border-radius: 6px;
      border: none;
      cursor: pointer;
      font-size: 13px;
      font-weight: 500;
      transition: all 0.15s ease;
    }

    .btn-primary {
      background: var(--accent);
      color: white;
    }

    .btn-primary:hover {
      filter: brightness(1.1);
    }

    .btn-secondary {
      background: var(--bg-card);
      color: var(--text-primary);
      border: 1px solid var(--border);
    }

    .btn-secondary:hover {
      background: var(--hover);
    }

    .board {
      display: flex;
      gap: 16px;
      padding: 24px;
      min-height: calc(100vh - 70px);
      overflow-x: auto;
    }

    .column {
      flex: 0 0 280px;
      background: var(--bg-secondary);
      border-radius: 12px;
      display: flex;
      flex-direction: column;
      max-height: calc(100vh - 110px);
    }

    .column-header {
      padding: 16px;
      border-bottom: 1px solid var(--border);
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .column-title {
      display: flex;
      align-items: center;
      gap: 8px;
      font-weight: 600;
    }

    .column-count {
      background: var(--bg-card);
      padding: 2px 8px;
      border-radius: 10px;
      font-size: 12px;
      color: var(--text-secondary);
    }

    .column-cards {
      flex: 1;
      padding: 12px;
      overflow-y: auto;
      display: flex;
      flex-direction: column;
      gap: 10px;
    }

    .card {
      background: var(--bg-card);
      border-radius: 8px;
      padding: 12px;
      cursor: grab;
      border: 1px solid transparent;
      transition: all 0.2s ease;
    }

    .card:hover {
      border-color: var(--accent);
      transform: translateY(-2px);
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    }

    .card.dragging {
      opacity: 0.5;
      transform: rotate(3deg);
    }

    .card-name {
      font-weight: 500;
      margin-bottom: 6px;
      word-break: break-word;
    }

    .card-client {
      font-size: 12px;
      color: var(--text-secondary);
      margin-bottom: 8px;
    }

    .card-meta {
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-size: 11px;
      color: var(--text-secondary);
    }

    .card-hours {
      display: flex;
      align-items: center;
      gap: 4px;
    }

    .card-type {
      padding: 2px 6px;
      background: var(--bg-secondary);
      border-radius: 4px;
    }

    .card-actions {
      display: flex;
      gap: 4px;
      margin-top: 8px;
      opacity: 0;
      transition: opacity 0.15s ease;
    }

    .card:hover .card-actions {
      opacity: 1;
    }

    .card-btn {
      padding: 4px 8px;
      border-radius: 4px;
      border: none;
      background: var(--bg-secondary);
      color: var(--text-secondary);
      cursor: pointer;
      font-size: 11px;
      transition: all 0.15s ease;
    }

    .card-btn:hover {
      background: var(--hover);
      color: var(--text-primary);
    }

    .card-btn.danger:hover {
      background: #ef444420;
      color: #ef4444;
    }

    .add-card {
      padding: 10px;
      border: 1px dashed var(--border);
      border-radius: 8px;
      text-align: center;
      color: var(--text-secondary);
      cursor: pointer;
      transition: all 0.15s ease;
      font-size: 13px;
    }

    .add-card:hover {
      border-color: var(--accent);
      color: var(--accent);
      background: var(--hover);
    }

    .add-form {
      padding: 12px;
      background: var(--bg-card);
      border-radius: 8px;
    }

    .add-form input {
      width: 100%;
      padding: 8px 12px;
      border: 1px solid var(--border);
      border-radius: 6px;
      background: var(--bg-primary);
      color: var(--text-primary);
      font-size: 13px;
      margin-bottom: 8px;
    }

    .add-form input:focus {
      outline: none;
      border-color: var(--accent);
    }

    .add-form-actions {
      display: flex;
      gap: 6px;
    }

    .add-form-actions button {
      flex: 1;
      padding: 6px;
      border-radius: 4px;
      border: none;
      cursor: pointer;
      font-size: 12px;
    }

    .add-form-actions .save {
      background: var(--accent);
      color: white;
    }

    .add-form-actions .cancel {
      background: var(--bg-secondary);
      color: var(--text-secondary);
    }

    .drop-zone {
      min-height: 60px;
      border: 2px dashed transparent;
      border-radius: 8px;
      transition: all 0.2s ease;
    }

    .drop-zone.drag-over {
      border-color: var(--accent);
      background: var(--hover);
    }

    .empty-state {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      padding: 40px;
      color: var(--text-secondary);
      text-align: center;
    }

    .empty-state span {
      font-size: 48px;
      margin-bottom: 16px;
    }

    .status-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      display: inline-block;
    }

    .toast {
      position: fixed;
      bottom: 20px;
      right: 20px;
      padding: 12px 20px;
      background: var(--vscode-editorError-foreground);
      color: white;
      border-radius: 8px;
      font-size: 13px;
      animation: slideIn 0.3s ease;
      z-index: 1000;
    }

    @keyframes slideIn {
      from {
        transform: translateY(20px);
        opacity: 0;
      }
      to {
        transform: translateY(0);
        opacity: 1;
      }
    }
  </style>
</head>
<body>
  <div class="header">
    <h1>📋 Gig Board</h1>
    <div class="header-actions">
      <button class="btn btn-secondary" onclick="refresh()">
        ↻ Refresh
      </button>
    </div>
  </div>

  <div class="board" id="board">
    ${columns
      .map((col) => {
        const columnGigs = gigs.filter((g) => g.status === col.key);
        return `
      <div class="column" data-status="${col.key}">
        <div class="column-header">
          <div class="column-title">
            <span class="status-dot" style="background: ${col.color}"></span>
            ${col.icon} ${col.label}
          </div>
          <span class="column-count">${columnGigs.length}</span>
        </div>
        <div class="column-cards drop-zone"
             data-status="${col.key}"
             ondragover="handleDragOver(event)"
             ondragleave="handleDragLeave(event)"
             ondrop="handleDrop(event)">
          ${
            columnGigs.length === 0
              ? '<div class="empty-state"><span>📭</span><p>No gigs yet</p></div>'
              : columnGigs
                  .map(
                    (gig) => `
            <div class="card"
                 draggable="true"
                 data-gig-id="${gig.id}"
                 ondragstart="handleDragStart(event)"
                 ondragend="handleDragEnd(event)">
              <div class="card-name">${escapeHtml(gig.name)}</div>
              <div class="card-client">${escapeHtml(gig.client === '-' ? 'No client' : gig.client)}</div>
              <div class="card-meta">
                <span class="card-hours">⏱ ${gig.hours.toFixed(1)}h</span>
                <span class="card-type">${gig.type}</span>
              </div>
              <div class="card-actions">
                <button class="card-btn" onclick="showMoveMenu(${gig.id})">Move →</button>
                <button class="card-btn danger" onclick="deleteGig(${gig.id})">Delete</button>
              </div>
            </div>
          `
                  )
                  .join('')
          }
          <div class="add-card" data-status="${col.key}" onclick="showAddForm(this, '${col.key}')">
            + Add Gig
          </div>
        </div>
      </div>
    `;
      })
      .join('')}
  </div>

  <script>
    const vscode = acquireVsCodeApi();
    let draggedGigId = null;

    function escapeHtml(text) {
      const div = document.createElement('div');
      div.textContent = text;
      return div.innerHTML;
    }

    function handleDragStart(e) {
      draggedGigId = parseInt(e.target.dataset.gigId);
      e.target.classList.add('dragging');
      e.dataTransfer.effectAllowed = 'move';
    }

    function handleDragEnd(e) {
      e.target.classList.remove('dragging');
      document.querySelectorAll('.drop-zone').forEach(zone => {
        zone.classList.remove('drag-over');
      });
    }

    function handleDragOver(e) {
      e.preventDefault();
      e.currentTarget.classList.add('drag-over');
    }

    function handleDragLeave(e) {
      e.currentTarget.classList.remove('drag-over');
    }

    function handleDrop(e) {
      e.preventDefault();
      e.currentTarget.classList.remove('drag-over');

      const newStatus = e.currentTarget.dataset.status;
      if (draggedGigId && newStatus) {
        vscode.postMessage({
          command: 'moveGig',
          gigId: draggedGigId,
          newStatus: newStatus
        });
      }
    }

    function showAddForm(element, status) {
      element.outerHTML = \`
        <div class="add-form">
          <input type="text" id="newGigName-\${status}" placeholder="Gig name..." autofocus onkeydown="handleAddKeydown(event, '\${status}')">
          <div class="add-form-actions">
            <button class="save" onclick="createGig('\${status}')">Add</button>
            <button class="cancel" onclick="cancelAdd('\${status}')">Cancel</button>
          </div>
        </div>
      \`;
      document.getElementById('newGigName-' + status).focus();
    }

    function handleAddKeydown(e, status) {
      if (e.key === 'Enter') {
        createGig(status);
      } else if (e.key === 'Escape') {
        cancelAdd(status);
      }
    }

    function createGig(status) {
      const input = document.getElementById('newGigName-' + status);
      const name = input.value.trim();
      if (name) {
        vscode.postMessage({
          command: 'createGig',
          name: name,
          status: status
        });
      }
    }

    function cancelAdd(status) {
      const form = document.querySelector('.add-form');
      if (form) {
        form.outerHTML = \`<div class="add-card" data-status="\${status}" onclick="showAddForm(this, '\${status}')">+ Add Gig</div>\`;
      }
    }

    function deleteGig(gigId) {
      if (confirm('Delete this gig? This cannot be undone!')) {
        vscode.postMessage({
          command: 'deleteGig',
          gigId: gigId
        });
      }
    }

    function showMoveMenu(gigId) {
      const statuses = ['pipeline', 'negotiating', 'active', 'delivered', 'invoiced', 'complete'];
      const status = prompt('Move to: ' + statuses.join(', '));
      if (status && statuses.includes(status)) {
        vscode.postMessage({
          command: 'moveGig',
          gigId: gigId,
          newStatus: status
        });
      }
    }

    function refresh() {
      vscode.postMessage({ command: 'refresh' });
    }

    // Handle messages from extension
    window.addEventListener('message', event => {
      const message = event.data;
      if (message.command === 'error') {
        showToast(message.message);
      }
    });

    function showToast(message) {
      const toast = document.createElement('div');
      toast.className = 'toast';
      toast.textContent = message;
      document.body.appendChild(toast);
      setTimeout(() => toast.remove(), 3000);
    }
  </script>
</body>
</html>`;
  }

  public dispose(): void {
    GigPanel.currentPanel = undefined;
    this._panel.dispose();
    while (this._disposables.length) {
      const d = this._disposables.pop();
      if (d) {
        d.dispose();
      }
    }
  }
}

// Export helper function for escaping HTML
function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}
