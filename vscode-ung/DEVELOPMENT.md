# Development Guide

## Project Structure

```
vscode-ung/
├── package.json              # Extension manifest
├── tsconfig.json            # TypeScript configuration
├── .gitignore              # Git ignore patterns
├── .vscodeignore           # VSIX packaging exclusions
├── README.md               # User documentation
├── CHANGELOG.md            # Version history
├── DEVELOPMENT.md          # This file
├── media/
│   ├── icon.svg            # Extension icon (SVG source)
│   ├── icon.png            # Extension icon (128x128 PNG - needs conversion)
│   ├── README.md           # Icon creation guide
│   └── icons/              # Additional icons (if needed)
└── src/
    ├── extension.ts        # Main entry point
    ├── cli/
    │   └── ungCli.ts       # CLI wrapper for executing commands
    ├── commands/
    │   ├── company.ts      # Company command handlers
    │   ├── client.ts       # Client command handlers
    │   ├── contract.ts     # Contract command handlers
    │   ├── invoice.ts      # Invoice command handlers
    │   ├── expense.ts      # Expense command handlers
    │   └── tracking.ts     # Time tracking command handlers
    ├── views/
    │   ├── invoiceProvider.ts   # Invoice tree view provider
    │   ├── contractProvider.ts  # Contract tree view provider
    │   ├── clientProvider.ts    # Client tree view provider
    │   ├── expenseProvider.ts   # Expense tree view provider
    │   └── trackingProvider.ts  # Time tracking tree view provider
    ├── webview/
    │   ├── invoicePanel.ts      # Invoice detail webview
    │   └── exportPanel.ts       # Export wizard webview
    ├── utils/
    │   ├── config.ts           # Configuration management
    │   ├── formatting.ts       # Date/currency formatting
    │   └── statusBar.ts        # Status bar manager
    └── test/
        ├── runTest.ts          # Test runner
        └── suite/
            ├── index.ts        # Test suite index
            ├── cli.test.ts     # CLI wrapper tests
            ├── commands.test.ts # Command tests
            └── providers.test.ts # Provider tests
```

## Setup Development Environment

### Prerequisites
- Node.js 18+ and npm
- VSCode 1.80+
- UNG CLI installed (`go install github.com/Andriiklymiuk/ung@latest`)

### Installation

1. Clone/navigate to the extension directory:
   ```bash
   cd vscode-ung
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Generate icon.png from icon.svg:
   ```bash
   cd media
   # Use ImageMagick, Inkscape, or online converter
   convert -background none -size 128x128 icon.svg icon.png
   cd ..
   ```

4. Compile TypeScript:
   ```bash
   npm run compile
   ```

## Development Workflow

### Run Extension in Debug Mode

1. Open project in VSCode: `code .`
2. Press `F5` or click "Run > Start Debugging"
3. A new VSCode window opens with the extension loaded
4. Test commands via Command Palette (`Cmd/Ctrl + Shift + P`)
5. Check Output panel: "UNG Operations" for logs

### Watch Mode

Auto-compile on file changes:
```bash
npm run watch
```

Then press `F5` to debug. Changes will recompile automatically.

### Running Tests

```bash
# Run all tests
npm test

# Compile before testing
npm run pretest
```

## Architecture

### Extension Activation Flow

1. **extension.ts** - `activate()` called when extension loads
2. Creates output channel for logging
3. Initializes `UngCli` wrapper
4. Checks if CLI is installed
5. Initializes `StatusBarManager`
6. Registers tree view providers
7. Initializes command handlers
8. Registers all commands with VSCode

### CLI Wrapper (`UngCli`)

The `UngCli` class wraps all CLI commands:
- Executes shell commands via `child_process.exec`
- Handles stdout/stderr
- Parses JSON output (if requested)
- Provides type-safe methods for each UNG command
- Logs all operations to output channel

### Command Handlers

Each command handler class:
- Accepts `UngCli` instance
- Optional refresh callback for auto-refresh
- Methods for each user command
- Shows progress indicators
- Displays success/error messages
- Validates user input

### Tree View Providers

Each provider:
- Implements `vscode.TreeDataProvider<T>`
- Fetches data via `UngCli`
- Parses CLI text output into tree items
- Provides refresh capability
- Returns themed icons
- Handles empty states

### Status Bar Manager

- Polls active session every 5 seconds
- Parses CLI output
- Calculates elapsed time
- Updates status bar text
- Shows/hides based on settings

### Webview Panels

- Singleton pattern (one panel at a time)
- HTML/CSS/JS interface
- Message passing to extension
- Themed with VSCode CSS variables
- Disposes resources properly

## Key Design Decisions

### 1. CLI Wrapper Pattern
**Decision**: Wrap CLI instead of direct database access

**Rationale**:
- Maintains single source of truth (CLI)
- No database schema coupling
- CLI handles all business logic
- Easier maintenance

**Trade-off**:
- Some operations require text parsing
- Limited by CLI's output format
- Can't provide all features in UI

### 2. Text Parsing vs JSON
**Decision**: Parse text output for list commands

**Rationale**:
- CLI doesn't yet provide --json flag
- Text output is stable and predictable
- Future: migrate to JSON when available

**Implementation**:
```typescript
private parseInvoiceOutput(output: string): InvoiceItem[] {
    const lines = output.split('\n');
    // Parse tabular output
}
```

### 3. Interactive Commands
**Decision**: Delegate some commands to CLI

**Rationale**:
- CLI has better interactive forms (huh library)
- Complex multi-step workflows
- Avoid duplicating validation logic

**User Experience**:
- Show informative message
- Provide CLI command to run
- Future: implement in extension

### 4. Status Bar Polling
**Decision**: Poll every 5 seconds for active session

**Rationale**:
- Real-time updates for time tracking
- Low performance impact
- Alternative would require CLI daemon

### 5. Single Output Channel
**Decision**: One output channel for all operations

**Rationale**:
- Centralized logging
- Easy troubleshooting
- Follows VSCode best practices

## Adding New Features

### Add a New Command

1. **Define in package.json**:
```json
{
  "command": "ung.myCommand",
  "title": "My Command",
  "category": "UNG"
}
```

2. **Add CLI method** (if needed):
```typescript
// src/cli/ungCli.ts
async myCliOperation(params): Promise<CliResult> {
    return this.exec(['my', 'command', ...]);
}
```

3. **Create command handler**:
```typescript
// src/commands/myFeature.ts
async myCommand(): Promise<void> {
    const result = await this.cli.myCliOperation();
    if (result.success) {
        vscode.window.showInformationMessage('Success!');
    }
}
```

4. **Register in extension.ts**:
```typescript
context.subscriptions.push(
    vscode.commands.registerCommand('ung.myCommand',
        () => myFeatureCommands.myCommand())
);
```

### Add a New Tree View

1. **Define in package.json**:
```json
{
  "views": {
    "ung": [
      {
        "id": "ungMyView",
        "name": "My View"
      }
    ]
  }
}
```

2. **Create provider**:
```typescript
// src/views/myViewProvider.ts
export class MyViewProvider implements vscode.TreeDataProvider<MyItem> {
    // Implement getChildren() and getTreeItem()
}
```

3. **Register in extension.ts**:
```typescript
const myViewProvider = new MyViewProvider(cli);
const myViewTree = vscode.window.createTreeView('ungMyView', {
    treeDataProvider: myViewProvider
});
```

## Testing Strategy

### Unit Tests
- Test CLI wrapper methods
- Mock exec calls
- Test parsing logic
- Test formatting utilities

### Integration Tests
- Test command registration
- Test provider data fetching
- Use @vscode/test-electron

### Manual Testing Checklist
- [ ] All commands show in Command Palette
- [ ] Tree views load data
- [ ] Context menus work
- [ ] Status bar updates
- [ ] Webviews open and function
- [ ] Error messages are helpful
- [ ] Progress indicators show
- [ ] Settings are respected

## Packaging

### Build VSIX

```bash
# Install vsce
npm install -g @vscode/vsce

# Package extension
vsce package

# Output: ung-1.0.0.vsix
```

### Install Locally

```bash
code --install-extension ung-1.0.0.vsix
```

### Publish to Marketplace

```bash
# Get Personal Access Token from Azure DevOps
# https://dev.azure.com/[your-org]/_usersSettings/tokens

vsce login <publisher>
vsce publish
```

## Troubleshooting

### Extension Not Activating
- Check activation events in package.json
- Look for errors in Developer Tools (Help > Toggle Developer Tools)

### Commands Not Showing
- Verify command IDs match between package.json and extension.ts
- Reload window: Developer: Reload Window

### Tree Views Empty
- Check UNG CLI is installed: `ung --version`
- Check output channel for errors
- Verify database exists: `~/.ung/ung.db`

### TypeScript Errors
- Run `npm run compile` to see errors
- Check tsconfig.json strictness settings

## Contributing

### Code Style
- Use TypeScript strict mode
- Add JSDoc comments for public APIs
- Follow VSCode extension conventions
- Use async/await (no callbacks)
- Handle all error cases

### Git Workflow
1. Create feature branch
2. Make changes
3. Test thoroughly
4. Update CHANGELOG.md
5. Create pull request

### Commit Messages
- Format: `feat: add invoice filtering`
- Types: feat, fix, docs, style, refactor, test, chore

## Resources

- [VSCode Extension API](https://code.visualstudio.com/api)
- [TreeView Guide](https://code.visualstudio.com/api/extension-guides/tree-view)
- [Webview Guide](https://code.visualstudio.com/api/extension-guides/webview)
- [Testing Extensions](https://code.visualstudio.com/api/working-with-extensions/testing-extension)
- [Publishing Extensions](https://code.visualstudio.com/api/working-with-extensions/publishing-extension)
- [UNG CLI Repository](https://github.com/Andriiklymiuk/ung)

## Future Enhancements

### Short-term
- Add --json flag support in CLI
- Implement inline editing
- Add advanced filtering
- Improve error messages

### Medium-term
- Dashboard with charts
- Custom themes for PDFs
- Multi-workspace support
- Integration with accounting software

### Long-term
- Cloud sync integration
- Mobile app connection
- AI-powered insights
- Team collaboration features
