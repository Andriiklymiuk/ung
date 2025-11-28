import * as vscode from 'vscode';
import type { UngCli } from '../cli/ungCli';

// Template block types that can be used in the editor
type BlockType =
  | 'header'
  | 'company_info'
  | 'client_info'
  | 'invoice_meta'
  | 'line_items'
  | 'totals'
  | 'notes'
  | 'terms'
  | 'spacer';

interface TemplateBlock {
  id: string;
  type: BlockType;
  position: {
    x: number;
    y: number;
    width: number;
    height: number;
    align: 'left' | 'center' | 'right';
  };
  style: {
    backgroundColor?: string;
    textColor?: string;
    borderColor?: string;
    borderWidth?: number;
    padding?: number;
    fontSize?: number;
    fontStyle?: string;
  };
  options?: Record<string, unknown>;
}

interface TemplateDefinition {
  name: string;
  description: string;
  type: 'invoice' | 'contract';
  page_size: string;
  margins: {
    top: number;
    bottom: number;
    left: number;
    right: number;
  };
  colors: {
    primary: string;
    secondary: string;
    text: string;
    muted: string;
    accent: string;
  };
  fonts: {
    family: string;
    size_title: number;
    size_body: number;
    size_small: number;
  };
  blocks: TemplateBlock[];
}

/**
 * Template Editor Webview Panel with drag-and-drop functionality
 */
export class TemplateEditorPanel {
  public static currentPanel: TemplateEditorPanel | undefined;
  private readonly panel: vscode.WebviewPanel;
  private disposables: vscode.Disposable[] = [];
  private cli: UngCli;
  private currentTemplate: TemplateDefinition | null = null;
  private templatePath: string | null = null;

  private constructor(
    panel: vscode.WebviewPanel,
    cli: UngCli,
    templatePath?: string
  ) {
    this.cli = cli;
    this.panel = panel;
    this.templatePath = templatePath || null;

    // Set the webview's HTML content
    this.update();

    // Handle messages from the webview
    this.panel.webview.onDidReceiveMessage(
      async (message) => {
        switch (message.command) {
          case 'ready':
            await this.loadTemplate();
            break;
          case 'save':
            await this.saveTemplate(message.template);
            break;
          case 'preview':
            await this.generatePreview();
            break;
          case 'updateBlock':
            this.updateBlock(message.blockId, message.updates);
            break;
          case 'addBlock':
            this.addBlock(message.blockType, message.beforeIndex);
            break;
          case 'removeBlock':
            this.removeBlock(message.blockId);
            break;
          case 'reorderBlocks':
            this.reorderBlocks(message.blockIds);
            break;
          case 'updateColors':
            this.updateColors(message.colors);
            break;
          case 'updateFonts':
            this.updateFonts(message.fonts);
            break;
          case 'updateMargins':
            this.updateMargins(message.margins);
            break;
          case 'newTemplate':
            await this.createNewTemplate();
            break;
          case 'openTemplate':
            await this.openTemplateDialog();
            break;
        }
      },
      null,
      this.disposables
    );

    // Handle panel disposed
    this.panel.onDidDispose(() => this.dispose(), null, this.disposables);
  }

  /**
   * Create or show the template editor panel
   */
  public static createOrShow(cli: UngCli, templatePath?: string) {
    const column = vscode.window.activeTextEditor
      ? vscode.window.activeTextEditor.viewColumn
      : undefined;

    // If we already have a panel, show it
    if (TemplateEditorPanel.currentPanel) {
      TemplateEditorPanel.currentPanel.panel.reveal(column);
      if (templatePath) {
        TemplateEditorPanel.currentPanel.templatePath = templatePath;
        TemplateEditorPanel.currentPanel.loadTemplate();
      }
      return;
    }

    // Otherwise, create a new panel
    const panel = vscode.window.createWebviewPanel(
      'ungTemplateEditor',
      'Template Editor',
      column || vscode.ViewColumn.One,
      {
        enableScripts: true,
        retainContextWhenHidden: true,
      }
    );

    TemplateEditorPanel.currentPanel = new TemplateEditorPanel(
      panel,
      cli,
      templatePath
    );
  }

  /**
   * Load template from file or create default
   */
  private async loadTemplate() {
    if (this.templatePath) {
      const result = await this.cli.exec([
        'template',
        'show',
        this.templatePath,
      ]);
      if (result.success && result.stdout) {
        try {
          this.currentTemplate = JSON.parse(result.stdout);
        } catch {
          this.currentTemplate = this.getDefaultTemplate();
        }
      } else {
        this.currentTemplate = this.getDefaultTemplate();
      }
    } else {
      this.currentTemplate = this.getDefaultTemplate();
    }

    this.sendTemplateToWebview();
  }

  /**
   * Get default template definition
   */
  private getDefaultTemplate(): TemplateDefinition {
    return {
      name: 'New Template',
      description: 'Custom invoice template',
      type: 'invoice',
      page_size: 'A4',
      margins: { top: 15, bottom: 15, left: 15, right: 15 },
      colors: {
        primary: '#E87722',
        secondary: '#505050',
        text: '#3C3C3C',
        muted: '#808080',
        accent: '#2C3E50',
      },
      fonts: {
        family: 'Arial',
        size_title: 28,
        size_body: 10,
        size_small: 9,
      },
      blocks: [
        {
          id: 'block-1',
          type: 'header',
          position: { x: 0, y: 0, width: 100, height: 0, align: 'left' },
          style: { fontSize: 28, fontStyle: 'B' },
          options: { show_logo: true, show_title: true },
        },
        {
          id: 'block-2',
          type: 'spacer',
          position: { x: 0, y: 0, width: 100, height: 10, align: 'left' },
          style: {},
        },
        {
          id: 'block-3',
          type: 'company_info',
          position: { x: 0, y: 0, width: 50, height: 0, align: 'left' },
          style: { fontSize: 9 },
        },
        {
          id: 'block-4',
          type: 'invoice_meta',
          position: { x: 50, y: 0, width: 50, height: 0, align: 'right' },
          style: { fontSize: 10 },
        },
        {
          id: 'block-5',
          type: 'spacer',
          position: { x: 0, y: 0, width: 100, height: 10, align: 'left' },
          style: {},
        },
        {
          id: 'block-6',
          type: 'client_info',
          position: { x: 0, y: 0, width: 50, height: 0, align: 'left' },
          style: { fontSize: 9 },
        },
        {
          id: 'block-7',
          type: 'spacer',
          position: { x: 0, y: 0, width: 100, height: 15, align: 'left' },
          style: {},
        },
        {
          id: 'block-8',
          type: 'line_items',
          position: { x: 0, y: 0, width: 100, height: 0, align: 'left' },
          style: { fontSize: 9 },
          options: { show_header: true, show_borders: true },
        },
        {
          id: 'block-9',
          type: 'totals',
          position: { x: 0, y: 0, width: 100, height: 0, align: 'right' },
          style: { fontSize: 10, fontStyle: 'B' },
        },
        {
          id: 'block-10',
          type: 'spacer',
          position: { x: 0, y: 0, width: 100, height: 15, align: 'left' },
          style: {},
        },
        {
          id: 'block-11',
          type: 'notes',
          position: { x: 0, y: 0, width: 100, height: 0, align: 'left' },
          style: { fontSize: 9 },
        },
        {
          id: 'block-12',
          type: 'terms',
          position: { x: 0, y: 0, width: 100, height: 0, align: 'left' },
          style: { fontSize: 9 },
        },
      ],
    };
  }

  /**
   * Send current template to webview
   */
  private sendTemplateToWebview() {
    this.panel.webview.postMessage({
      command: 'loadTemplate',
      template: this.currentTemplate,
    });
  }

  /**
   * Save template to file
   */
  private async saveTemplate(template: TemplateDefinition) {
    if (!this.templatePath) {
      // Ask for file path
      const uri = await vscode.window.showSaveDialog({
        defaultUri: vscode.Uri.file(`${template.name}.json`),
        filters: { 'JSON files': ['json'] },
      });

      if (!uri) {
        return;
      }

      this.templatePath = uri.fsPath;
    }

    try {
      const fs = await import('node:fs');
      fs.writeFileSync(this.templatePath, JSON.stringify(template, null, 2));
      this.currentTemplate = template;
      vscode.window.showInformationMessage(
        `Template saved to ${this.templatePath}`
      );
    } catch (error) {
      vscode.window.showErrorMessage(
        `Failed to save template: ${error instanceof Error ? error.message : 'Unknown error'}`
      );
    }
  }

  /**
   * Generate PDF preview
   */
  private async generatePreview() {
    if (!this.currentTemplate) {
      vscode.window.showWarningMessage('No template loaded');
      return;
    }

    // Save to temp file and generate preview
    const tempPath = `/tmp/ung-template-preview-${Date.now()}.json`;
    const previewPath = `/tmp/ung-template-preview-${Date.now()}.pdf`;

    try {
      const fs = await import('node:fs');
      fs.writeFileSync(tempPath, JSON.stringify(this.currentTemplate, null, 2));

      // Use --file flag to specify external file path
      // Use useGlobal: false to prevent --global flag from being added
      // (otherwise CLI interprets the path as a template name in ~/.ung/templates/)
      const result = await this.cli.exec(
        ['template', 'preview', '--file', tempPath, '--output', previewPath],
        { useGlobal: false }
      );

      if (result.success) {
        // Open the PDF
        await vscode.env.openExternal(vscode.Uri.file(previewPath));
      } else {
        vscode.window.showErrorMessage(
          `Preview failed: ${result.error || 'Unknown error'}`
        );
      }

      // Clean up temp file
      try {
        fs.unlinkSync(tempPath);
      } catch {
        // Ignore cleanup errors
      }
    } catch (error) {
      vscode.window.showErrorMessage(
        `Preview failed: ${error instanceof Error ? error.message : 'Unknown error'}`
      );
    }
  }

  /**
   * Update a block in the template
   */
  private updateBlock(blockId: string, updates: Partial<TemplateBlock>) {
    if (!this.currentTemplate) return;

    const blockIndex = this.currentTemplate.blocks.findIndex(
      (b) => b.id === blockId
    );
    if (blockIndex !== -1) {
      const block = this.currentTemplate.blocks[blockIndex];
      // Deep merge nested objects (position, style, options)
      this.currentTemplate.blocks[blockIndex] = {
        ...block,
        ...updates,
        position: { ...block.position, ...updates.position },
        style: { ...block.style, ...updates.style },
        options: { ...block.options, ...updates.options },
      };
      this.sendTemplateToWebview();
    }
  }

  /**
   * Add a new block to the template
   * @param blockType Type of block to add
   * @param beforeIndex Optional index to insert before (appends to end if not provided)
   */
  private addBlock(blockType: BlockType, beforeIndex?: number) {
    if (!this.currentTemplate) return;

    const newBlock: TemplateBlock = {
      id: `block-${Date.now()}`,
      type: blockType,
      position: { x: 0, y: 0, width: 100, height: 0, align: 'left' },
      style: { fontSize: 9 },
    };

    if (beforeIndex !== undefined && beforeIndex >= 0) {
      this.currentTemplate.blocks.splice(beforeIndex, 0, newBlock);
    } else {
      this.currentTemplate.blocks.push(newBlock);
    }
    this.sendTemplateToWebview();
  }

  /**
   * Remove a block from the template
   */
  private removeBlock(blockId: string) {
    if (!this.currentTemplate) return;

    this.currentTemplate.blocks = this.currentTemplate.blocks.filter(
      (b) => b.id !== blockId
    );
    this.sendTemplateToWebview();
  }

  /**
   * Reorder blocks in the template
   */
  private reorderBlocks(blockIds: string[]) {
    if (!this.currentTemplate) return;

    const reorderedBlocks: TemplateBlock[] = [];
    for (const id of blockIds) {
      const block = this.currentTemplate.blocks.find((b) => b.id === id);
      if (block) {
        reorderedBlocks.push(block);
      }
    }
    this.currentTemplate.blocks = reorderedBlocks;
    this.sendTemplateToWebview();
  }

  /**
   * Update template colors
   */
  private updateColors(colors: TemplateDefinition['colors']) {
    if (!this.currentTemplate) return;
    this.currentTemplate.colors = colors;
    this.sendTemplateToWebview();
  }

  /**
   * Update template fonts
   */
  private updateFonts(fonts: TemplateDefinition['fonts']) {
    if (!this.currentTemplate) return;
    this.currentTemplate.fonts = fonts;
    this.sendTemplateToWebview();
  }

  /**
   * Update template margins
   */
  private updateMargins(margins: TemplateDefinition['margins']) {
    if (!this.currentTemplate) return;
    this.currentTemplate.margins = margins;
    this.sendTemplateToWebview();
  }

  /**
   * Create a new template
   */
  private async createNewTemplate() {
    const name = await vscode.window.showInputBox({
      prompt: 'Enter template name',
      value: 'My Custom Template',
    });

    if (!name) return;

    this.currentTemplate = this.getDefaultTemplate();
    this.currentTemplate.name = name;
    this.templatePath = null;
    this.sendTemplateToWebview();
  }

  /**
   * Open template file dialog
   */
  private async openTemplateDialog() {
    const uri = await vscode.window.showOpenDialog({
      canSelectMany: false,
      filters: { 'JSON files': ['json'] },
    });

    if (!uri || uri.length === 0) return;

    this.templatePath = uri[0].fsPath;
    await this.loadTemplate();
  }

  /**
   * Update webview content
   */
  private update() {
    this.panel.webview.html = this.getHtmlForWebview();
  }

  /**
   * Get HTML content for the webview
   */
  private getHtmlForWebview(): string {
    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Template Editor</title>
    <style>
        * {
            box-sizing: border-box;
        }
        body {
            font-family: var(--vscode-font-family);
            color: var(--vscode-foreground);
            background: var(--vscode-editor-background);
            margin: 0;
            padding: 0;
            height: 100vh;
            display: flex;
            flex-direction: column;
        }
        .toolbar {
            display: flex;
            gap: 8px;
            padding: 12px;
            background: var(--vscode-titleBar-activeBackground);
            border-bottom: 1px solid var(--vscode-panel-border);
            align-items: center;
        }
        .toolbar-group {
            display: flex;
            gap: 4px;
        }
        .toolbar-separator {
            width: 1px;
            height: 24px;
            background: var(--vscode-panel-border);
            margin: 0 8px;
        }
        button {
            background: var(--vscode-button-background);
            color: var(--vscode-button-foreground);
            border: none;
            padding: 6px 12px;
            border-radius: 3px;
            cursor: pointer;
            font-size: 12px;
        }
        button:hover {
            background: var(--vscode-button-hoverBackground);
        }
        button.secondary {
            background: var(--vscode-button-secondaryBackground);
            color: var(--vscode-button-secondaryForeground);
        }
        button.secondary:hover {
            background: var(--vscode-button-secondaryHoverBackground);
        }
        .main-content {
            flex: 1;
            display: flex;
            overflow: hidden;
        }
        .sidebar {
            width: 280px;
            background: var(--vscode-sideBar-background);
            border-right: 1px solid var(--vscode-panel-border);
            overflow-y: auto;
            display: flex;
            flex-direction: column;
        }
        .sidebar-section {
            padding: 12px;
            border-bottom: 1px solid var(--vscode-panel-border);
        }
        .sidebar-section h3 {
            margin: 0 0 12px 0;
            font-size: 11px;
            text-transform: uppercase;
            color: var(--vscode-descriptionForeground);
            font-weight: 600;
        }
        .component-list {
            display: flex;
            flex-direction: column;
            gap: 4px;
        }
        .component-item {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 8px;
            background: var(--vscode-input-background);
            border: 1px solid var(--vscode-input-border);
            border-radius: 4px;
            cursor: grab;
            font-size: 12px;
        }
        .component-item:hover {
            background: var(--vscode-list-hoverBackground);
        }
        .component-item:active {
            cursor: grabbing;
        }
        .component-icon {
            width: 20px;
            text-align: center;
        }
        .canvas-container {
            flex: 1;
            padding: 20px;
            overflow: auto;
            display: flex;
            justify-content: center;
        }
        .canvas {
            width: 595px; /* A4 width in pixels at 72dpi */
            min-height: 842px; /* A4 height */
            background: white;
            box-shadow: 0 2px 10px rgba(0,0,0,0.2);
            position: relative;
        }
        .canvas-page {
            padding: 40px;
        }
        .block {
            position: relative;
            padding: 8px;
            margin-bottom: 4px;
            border: 2px dashed transparent;
            transition: all 0.2s;
            min-height: 30px;
        }
        .block:hover {
            border-color: var(--vscode-focusBorder);
            background: rgba(0, 120, 212, 0.05);
        }
        .block.selected {
            border-color: var(--vscode-focusBorder);
            border-style: solid;
        }
        .block.dragging {
            opacity: 0.5;
        }
        .block.drop-target {
            border-color: #4CAF50;
            border-style: solid;
        }
        .block-handle {
            position: absolute;
            top: 4px;
            right: 4px;
            cursor: move;
            opacity: 0;
            transition: opacity 0.2s;
            display: flex;
            gap: 4px;
        }
        .block:hover .block-handle {
            opacity: 1;
        }
        .block-handle button {
            padding: 2px 6px;
            font-size: 10px;
        }
        .block-label {
            position: absolute;
            top: -10px;
            left: 8px;
            font-size: 10px;
            background: var(--vscode-badge-background);
            color: var(--vscode-badge-foreground);
            padding: 2px 6px;
            border-radius: 2px;
            opacity: 0;
            transition: opacity 0.2s;
        }
        .block:hover .block-label {
            opacity: 1;
        }
        .block-content {
            color: #333;
            font-size: 11px;
        }
        .properties-panel {
            width: 300px;
            background: var(--vscode-sideBar-background);
            border-left: 1px solid var(--vscode-panel-border);
            overflow-y: auto;
        }
        .property-group {
            padding: 12px;
            border-bottom: 1px solid var(--vscode-panel-border);
        }
        .property-group h4 {
            margin: 0 0 12px 0;
            font-size: 12px;
            font-weight: 600;
        }
        .property-row {
            display: flex;
            align-items: center;
            margin-bottom: 8px;
            gap: 8px;
        }
        .property-row label {
            width: 80px;
            font-size: 11px;
            color: var(--vscode-descriptionForeground);
        }
        .property-row input,
        .property-row select {
            flex: 1;
            padding: 4px 8px;
            background: var(--vscode-input-background);
            color: var(--vscode-input-foreground);
            border: 1px solid var(--vscode-input-border);
            border-radius: 3px;
            font-size: 12px;
        }
        .property-row input[type="color"] {
            width: 40px;
            height: 26px;
            padding: 2px;
        }
        .color-preview {
            width: 20px;
            height: 20px;
            border-radius: 3px;
            border: 1px solid var(--vscode-input-border);
        }
        .drop-zone {
            min-height: 60px;
            border: 2px dashed var(--vscode-input-border);
            border-radius: 8px;
            display: flex;
            align-items: center;
            justify-content: center;
            color: var(--vscode-descriptionForeground);
            font-size: 12px;
            margin: 12px 0;
            padding: 16px;
            transition: all 0.2s;
            background: rgba(128, 128, 128, 0.05);
        }
        .drop-zone.active, .drop-zone.drop-target {
            border-color: var(--vscode-focusBorder);
            border-style: solid;
            background: rgba(0, 120, 212, 0.15);
            color: var(--vscode-textLink-foreground);
            font-weight: 500;
        }

        /* Preview blocks styling */
        .preview-header {
            display: flex;
            justify-content: space-between;
            margin-bottom: 15px;
        }
        .preview-company {
            font-weight: bold;
            font-size: 16px;
            color: #333;
        }
        .preview-title {
            font-weight: bold;
            font-size: 24px;
        }
        .preview-meta {
            text-align: right;
            font-size: 11px;
        }
        .preview-section {
            margin: 10px 0;
        }
        .preview-section-title {
            font-weight: bold;
            font-size: 10px;
            margin-bottom: 4px;
        }
        .preview-table {
            width: 100%;
            border-collapse: collapse;
            font-size: 10px;
        }
        .preview-table th {
            padding: 6px;
            text-align: left;
            color: white;
        }
        .preview-table td {
            padding: 6px;
            border-bottom: 1px solid #ddd;
        }
        .preview-total {
            text-align: right;
            font-weight: bold;
            font-size: 14px;
            padding: 10px 0;
        }
        .preview-notes {
            font-size: 10px;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="toolbar">
        <div class="toolbar-group">
            <button onclick="newTemplate()">New</button>
            <button onclick="openTemplate()" class="secondary">Open</button>
            <button onclick="saveTemplate()">Save</button>
        </div>
        <div class="toolbar-separator"></div>
        <div class="toolbar-group">
            <button onclick="previewTemplate()">Preview PDF</button>
        </div>
        <div class="toolbar-separator"></div>
        <span style="font-size: 12px; color: var(--vscode-descriptionForeground);">
            <span id="templateName">New Template</span>
        </span>
    </div>

    <div class="main-content">
        <div class="sidebar">
            <div class="sidebar-section">
                <h3>Components</h3>
                <div class="component-list">
                    <div class="component-item" draggable="true" data-type="header">
                        <span class="component-icon">📄</span>
                        <span>Header</span>
                    </div>
                    <div class="component-item" draggable="true" data-type="company_info">
                        <span class="component-icon">🏢</span>
                        <span>Company Info</span>
                    </div>
                    <div class="component-item" draggable="true" data-type="client_info">
                        <span class="component-icon">👤</span>
                        <span>Client Info</span>
                    </div>
                    <div class="component-item" draggable="true" data-type="invoice_meta">
                        <span class="component-icon">📋</span>
                        <span>Invoice Details</span>
                    </div>
                    <div class="component-item" draggable="true" data-type="line_items">
                        <span class="component-icon">📊</span>
                        <span>Line Items</span>
                    </div>
                    <div class="component-item" draggable="true" data-type="totals">
                        <span class="component-icon">💰</span>
                        <span>Totals</span>
                    </div>
                    <div class="component-item" draggable="true" data-type="notes">
                        <span class="component-icon">📝</span>
                        <span>Notes</span>
                    </div>
                    <div class="component-item" draggable="true" data-type="terms">
                        <span class="component-icon">📜</span>
                        <span>Terms</span>
                    </div>
                    <div class="component-item" draggable="true" data-type="spacer">
                        <span class="component-icon">↕️</span>
                        <span>Spacer</span>
                    </div>
                </div>
            </div>

            <div class="sidebar-section">
                <h3>Colors</h3>
                <div class="property-row">
                    <label>Primary</label>
                    <input type="color" id="colorPrimary" value="#E87722" onchange="updateColor('primary', this.value)">
                </div>
                <div class="property-row">
                    <label>Secondary</label>
                    <input type="color" id="colorSecondary" value="#505050" onchange="updateColor('secondary', this.value)">
                </div>
                <div class="property-row">
                    <label>Text</label>
                    <input type="color" id="colorText" value="#3C3C3C" onchange="updateColor('text', this.value)">
                </div>
            </div>

            <div class="sidebar-section">
                <h3>Margins (mm)</h3>
                <div class="property-row">
                    <label>Top</label>
                    <input type="number" id="marginTop" value="15" min="0" max="50" onchange="updateMargin('top', this.value)">
                </div>
                <div class="property-row">
                    <label>Bottom</label>
                    <input type="number" id="marginBottom" value="15" min="0" max="50" onchange="updateMargin('bottom', this.value)">
                </div>
                <div class="property-row">
                    <label>Left</label>
                    <input type="number" id="marginLeft" value="15" min="0" max="50" onchange="updateMargin('left', this.value)">
                </div>
                <div class="property-row">
                    <label>Right</label>
                    <input type="number" id="marginRight" value="15" min="0" max="50" onchange="updateMargin('right', this.value)">
                </div>
            </div>
        </div>

        <div class="canvas-container">
            <div class="canvas">
                <div class="canvas-page" id="canvasPage">
                    <div class="drop-zone" id="dropZone">
                        Drag components here to build your template
                    </div>
                </div>
            </div>
        </div>

        <div class="properties-panel" id="propertiesPanel">
            <div class="property-group">
                <h4>Select a block to edit its properties</h4>
                <p style="font-size: 11px; color: var(--vscode-descriptionForeground);">
                    Click on any block in the canvas to view and edit its properties.
                </p>
            </div>
        </div>
    </div>

    <script>
        const vscode = acquireVsCodeApi();
        let currentTemplate = null;
        let selectedBlockId = null;
        let draggedType = null;
        let draggedBlockId = null;

        // Initialize
        window.addEventListener('load', () => {
            vscode.postMessage({ command: 'ready' });
            setupDragAndDrop();
            setupContextMenu();
        });

        // Handle messages from extension
        window.addEventListener('message', event => {
            const message = event.data;
            switch (message.command) {
                case 'loadTemplate':
                    currentTemplate = message.template;
                    renderTemplate();
                    break;
            }
        });

        function setupDragAndDrop() {
            // Component items - use event delegation on sidebar
            const sidebar = document.querySelector('.sidebar');
            if (sidebar) {
                sidebar.addEventListener('dragstart', (e) => {
                    const item = e.target.closest('.component-item');
                    if (item) {
                        draggedType = item.dataset.type;
                        e.dataTransfer.effectAllowed = 'copy';
                        e.dataTransfer.setData('text/plain', item.dataset.type);
                    }
                });
                sidebar.addEventListener('dragend', () => {
                    draggedType = null;
                });
            }
        }

        function setupContextMenu() {
            // Right-click context menu for blocks
            document.addEventListener('contextmenu', (e) => {
                const block = e.target.closest('.block');
                if (block) {
                    e.preventDefault();
                    const blockId = block.dataset.blockId;
                    if (blockId && confirm('Remove this block?')) {
                        vscode.postMessage({ command: 'removeBlock', blockId });
                    }
                }
            });
        }

        function renderTemplate() {
            if (!currentTemplate) return;

            document.getElementById('templateName').textContent = currentTemplate.name;
            document.getElementById('colorPrimary').value = currentTemplate.colors.primary;
            document.getElementById('colorSecondary').value = currentTemplate.colors.secondary;
            document.getElementById('colorText').value = currentTemplate.colors.text;
            document.getElementById('marginTop').value = currentTemplate.margins.top;
            document.getElementById('marginBottom').value = currentTemplate.margins.bottom;
            document.getElementById('marginLeft').value = currentTemplate.margins.left;
            document.getElementById('marginRight').value = currentTemplate.margins.right;

            const canvasPage = document.getElementById('canvasPage');
            canvasPage.innerHTML = '';

            if (currentTemplate.blocks.length === 0) {
                canvasPage.innerHTML = '<div class="drop-zone" id="dropZone">Drag components here to build your template</div>';
            } else {
                currentTemplate.blocks.forEach((block, index) => {
                    const blockEl = createBlockElement(block, index);
                    canvasPage.appendChild(blockEl);
                });
            }

            // Add drop zone at the end
            const dropZone = document.createElement('div');
            dropZone.className = 'drop-zone';
            dropZone.id = 'dropZone';
            dropZone.textContent = 'Drop here to add';
            dropZone.addEventListener('dragover', handleDragOver);
            dropZone.addEventListener('drop', handleDropEnd);
            dropZone.addEventListener('dragleave', handleDragLeave);
            canvasPage.appendChild(dropZone);
        }

        function createBlockElement(block, index) {
            const el = document.createElement('div');
            el.className = 'block' + (block.id === selectedBlockId ? ' selected' : '');
            el.dataset.blockId = block.id;
            el.draggable = true;

            // Block label
            const label = document.createElement('div');
            label.className = 'block-label';
            label.textContent = getBlockLabel(block.type);
            el.appendChild(label);

            // Block content (preview)
            const content = document.createElement('div');
            content.className = 'block-content';
            content.innerHTML = getBlockPreview(block);
            el.appendChild(content);

            // Block handle
            const handle = document.createElement('div');
            handle.className = 'block-handle';
            handle.innerHTML = '<button onclick="removeBlock(event, \\'' + block.id + '\\')">×</button>';
            el.appendChild(handle);

            // Event listeners
            el.addEventListener('click', () => selectBlock(block.id));
            el.addEventListener('dragstart', (e) => {
                draggedBlockId = block.id;
                el.classList.add('dragging');
                e.dataTransfer.effectAllowed = 'move';
            });
            el.addEventListener('dragend', () => {
                draggedBlockId = null;
                el.classList.remove('dragging');
            });
            el.addEventListener('dragover', handleDragOver);
            el.addEventListener('drop', (e) => handleDropOnBlock(e, block.id));
            el.addEventListener('dragleave', handleDragLeave);

            return el;
        }

        function getBlockLabel(type) {
            const labels = {
                'header': 'Header',
                'company_info': 'Company Info',
                'client_info': 'Client Info',
                'invoice_meta': 'Invoice Details',
                'line_items': 'Line Items',
                'totals': 'Totals',
                'notes': 'Notes',
                'terms': 'Terms',
                'spacer': 'Spacer'
            };
            return labels[type] || type;
        }

        function getBlockPreview(block) {
            const primaryColor = currentTemplate?.colors?.primary || '#E87722';

            switch (block.type) {
                case 'header':
                    return '<div class="preview-header"><span class="preview-company">Company Name</span><span class="preview-title" style="color: ' + primaryColor + '">INVOICE</span></div>';
                case 'company_info':
                    return '<div class="preview-section"><div style="font-weight: bold;">Your Company Name</div><div>Tax ID: 12-3456789</div><div>123 Business St</div></div>';
                case 'client_info':
                    return '<div class="preview-section"><div class="preview-section-title" style="color: ' + primaryColor + '">Bill To</div><div style="font-weight: bold;">Client Name</div><div>Tax ID: 98-7654321</div></div>';
                case 'invoice_meta':
                    return '<div class="preview-meta"><div><strong>Invoice#</strong> INV-2024-001</div><div><strong>Date</strong> 27 Nov 2024</div><div><strong>Due</strong> 27 Dec 2024</div></div>';
                case 'line_items':
                    return '<table class="preview-table"><thead><tr style="background: ' + primaryColor + '"><th>Item</th><th>Qty</th><th>Rate</th><th>Amount</th></tr></thead><tbody><tr><td>Service</td><td>10</td><td>$100</td><td>$1,000</td></tr></tbody></table>';
                case 'totals':
                    return '<div class="preview-total">Total: $1,000.00</div>';
                case 'notes':
                    return '<div class="preview-notes"><strong style="color: ' + primaryColor + '">Notes</strong><br>Thank you for your business!</div>';
                case 'terms':
                    return '<div class="preview-notes"><strong style="color: ' + primaryColor + '">Terms</strong><br>Payment due within 30 days.</div>';
                case 'spacer':
                    return '<div style="height: ' + (block.position?.height || 10) + 'px; background: #f0f0f0; opacity: 0.3;"></div>';
                default:
                    return '<div>[' + block.type + ']</div>';
            }
        }

        function handleDragOver(e) {
            e.preventDefault();
            e.stopPropagation();
            e.dataTransfer.dropEffect = draggedType ? 'copy' : 'move';
            e.currentTarget.classList.add('drop-target', 'active');
        }

        function handleDragLeave(e) {
            e.preventDefault();
            // Only remove class if we're actually leaving the element
            const rect = e.currentTarget.getBoundingClientRect();
            const x = e.clientX;
            const y = e.clientY;
            if (x < rect.left || x >= rect.right || y < rect.top || y >= rect.bottom) {
                e.currentTarget.classList.remove('drop-target', 'active');
            }
        }

        function handleDropOnBlock(e, targetBlockId) {
            e.preventDefault();
            e.stopPropagation();
            e.currentTarget.classList.remove('drop-target', 'active');

            // Get dragged type from dataTransfer if draggedType is null
            const droppedType = draggedType || e.dataTransfer.getData('text/plain');

            if (droppedType) {
                // Adding new block before target
                const targetIndex = currentTemplate.blocks.findIndex(b => b.id === targetBlockId);
                vscode.postMessage({ command: 'addBlock', blockType: droppedType, beforeIndex: targetIndex });
                draggedType = null;
            } else if (draggedBlockId && draggedBlockId !== targetBlockId) {
                // Reordering blocks
                const newOrder = [...currentTemplate.blocks.map(b => b.id)];
                const draggedIndex = newOrder.indexOf(draggedBlockId);
                const targetIndex = newOrder.indexOf(targetBlockId);
                newOrder.splice(draggedIndex, 1);
                newOrder.splice(targetIndex, 0, draggedBlockId);
                vscode.postMessage({ command: 'reorderBlocks', blockIds: newOrder });
            }
        }

        function handleDropEnd(e) {
            e.preventDefault();
            e.stopPropagation();
            e.currentTarget.classList.remove('drop-target', 'active');

            // Get dragged type from dataTransfer if draggedType is null
            const droppedType = draggedType || e.dataTransfer.getData('text/plain');

            if (droppedType) {
                vscode.postMessage({ command: 'addBlock', blockType: droppedType });
                draggedType = null;
            }
        }

        function selectBlock(blockId) {
            selectedBlockId = blockId;
            renderTemplate();
            showBlockProperties(blockId);
        }

        function showBlockProperties(blockId) {
            const block = currentTemplate.blocks.find(b => b.id === blockId);
            if (!block) return;

            const panel = document.getElementById('propertiesPanel');
            panel.innerHTML = '<div class="property-group"><h4>' + getBlockLabel(block.type) + ' Properties</h4>' +
                '<div class="property-row"><label>Font Size</label><input type="number" value="' + (block.style?.fontSize || 9) + '" onchange="updateBlockStyle(\\'' + blockId + '\\', \\'fontSize\\', Number(this.value))"></div>' +
                '<div class="property-row"><label>Align</label><select onchange="updateBlockPosition(\\'' + blockId + '\\', \\'align\\', this.value)"><option value="left"' + (block.position?.align === 'left' ? ' selected' : '') + '>Left</option><option value="center"' + (block.position?.align === 'center' ? ' selected' : '') + '>Center</option><option value="right"' + (block.position?.align === 'right' ? ' selected' : '') + '>Right</option></select></div>' +
                (block.type === 'spacer' ? '<div class="property-row"><label>Height</label><input type="number" value="' + (block.position?.height || 10) + '" min="5" max="100" onchange="updateBlockPosition(\\'' + blockId + '\\', \\'height\\', Number(this.value))"></div>' : '') +
                '</div>';
        }

        function updateBlockStyle(blockId, property, value) {
            const block = currentTemplate.blocks.find(b => b.id === blockId);
            if (block) {
                block.style = block.style || {};
                block.style[property] = value;
                vscode.postMessage({ command: 'updateBlock', blockId, updates: { style: block.style } });
            }
        }

        function updateBlockPosition(blockId, property, value) {
            const block = currentTemplate.blocks.find(b => b.id === blockId);
            if (block) {
                block.position = block.position || {};
                block.position[property] = value;
                vscode.postMessage({ command: 'updateBlock', blockId, updates: { position: block.position } });
            }
        }

        function removeBlock(e, blockId) {
            e.stopPropagation();
            vscode.postMessage({ command: 'removeBlock', blockId });
        }

        function updateColor(key, value) {
            if (currentTemplate) {
                currentTemplate.colors[key] = value;
                vscode.postMessage({ command: 'updateColors', colors: currentTemplate.colors });
            }
        }

        function updateMargin(key, value) {
            if (currentTemplate) {
                currentTemplate.margins[key] = Number(value);
                vscode.postMessage({ command: 'updateMargins', margins: currentTemplate.margins });
            }
        }

        function newTemplate() {
            vscode.postMessage({ command: 'newTemplate' });
        }

        function openTemplate() {
            vscode.postMessage({ command: 'openTemplate' });
        }

        function saveTemplate() {
            vscode.postMessage({ command: 'save', template: currentTemplate });
        }

        function previewTemplate() {
            vscode.postMessage({ command: 'preview' });
        }
    </script>
</body>
</html>`;
  }

  /**
   * Dispose resources
   */
  public dispose() {
    TemplateEditorPanel.currentPanel = undefined;

    this.panel.dispose();

    while (this.disposables.length) {
      const disposable = this.disposables.pop();
      if (disposable) {
        disposable.dispose();
      }
    }
  }
}
