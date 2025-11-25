# UNG - Billing & Time Tracking Extension

A comprehensive VSCode extension for managing billing, invoicing, and time tracking using the UNG CLI tool.

## Features

### Company & Client Management
- Create and manage company profiles
- Add and manage client information
- View client list directly in VSCode sidebar

### Contract Management
- View all contracts in a tree view
- Generate contract PDFs
- Email contracts to clients
- Toggle contract active status

### Invoice Management
- Create new invoices with interactive forms
- Generate invoices from tracked time
- View invoice details in webview panels
- Export invoices to PDF
- Email invoices directly from VSCode
- View all invoices with status badges (pending, paid, overdue)

### Expense Tracking
- Log business expenses
- Categorize expenses (software, hardware, travel, meals, etc.)
- View expense reports
- Track expenses by category and date

### Time Tracking
- Start/stop timers directly from VSCode
- Log time manually for contracts
- View active tracking session in status bar
- See all tracking sessions in sidebar
- Real-time status bar updates showing elapsed time

### Tree View Sidebar
Dedicated activity bar with 5 organized views:
- **Invoices** - All invoices with amount, status, and due dates
- **Contracts** - Active and inactive contracts with rates
- **Clients** - All clients with contact information
- **Expenses** - Business expenses with categories
- **Time Tracking** - All tracked sessions with durations

### Context Menus
Right-click on any item for quick actions:
- **Invoices**: View, Export PDF, Email, Edit, Delete
- **Contracts**: View, Generate PDF, Edit, Delete
- **Clients**: Edit, Delete
- **Expenses**: Edit, Delete

### Status Bar Integration
- Shows active time tracking session with elapsed time
- Click to view session details
- Updates every 5 seconds
- Format: `⏱️ 2h 15m - Project Name`

## Requirements

- VSCode 1.80.0 or higher
- UNG CLI installed and available in PATH
  - Install via: `go install github.com/Andriiklymiuk/ung@latest`
  - Or use Homebrew: `brew install andriiklymiuk/tools/ung`

## Installation

1. Install from VSCode Marketplace (coming soon)
2. Or install from VSIX:
   ```bash
   code --install-extension ung-1.0.0.vsix
   ```

## Usage

### Getting Started

1. Open the Command Palette (`Cmd/Ctrl + Shift + P`)
2. Type "UNG" to see all available commands
3. Start by creating your company: `UNG: Create Company`
4. Add clients: `UNG: Create Client`
5. Create contracts and start tracking time!

### Command Palette Commands

**Company:**
- `UNG: Create Company` - Set up your business information
- `UNG: Edit Company` - Update company details

**Client:**
- `UNG: Create Client` - Add a new client
- `UNG: List Clients` - View all clients
- Context menu: Edit, Delete

**Contract:**
- `UNG: Create Contract` - Create a new contract (interactive)
- Context menu: View, Generate PDF, Email, Edit, Delete

**Invoice:**
- `UNG: Create Invoice` - Create a new invoice
- `UNG: Generate Invoice from Time` - Auto-generate from tracked hours
- Context menu: View, Export PDF, Email, Edit, Delete

**Expense:**
- `UNG: Log Expense` - Record a business expense (interactive)
- `UNG: View Expense Report` - See expense summary
- Context menu: Edit, Delete

**Time Tracking:**
- `UNG: Start Time Tracking` - Start tracking time
- `UNG: Stop Time Tracking` - Stop active timer
- `UNG: Log Time Manually` - Manually log hours (interactive)
- `UNG: View Active Session` - Show current tracking session

### Sidebar Usage

1. Click the UNG icon in the activity bar (left sidebar)
2. Expand any view (Invoices, Contracts, Clients, etc.)
3. Click toolbar icons to create new items
4. Right-click items for context menu actions
5. Use refresh buttons to update data

### Keyboard Shortcuts

No default keyboard shortcuts are set. You can customize them in VSCode:
1. Open Keyboard Shortcuts (`Cmd/Ctrl + K, Cmd/Ctrl + S`)
2. Search for "UNG"
3. Assign your preferred shortcuts

## Configuration

Access settings via `Preferences > Settings > Extensions > UNG`

### Available Settings

```json
{
  "ung.cliPath": "ung",
  "ung.defaultCurrency": "USD",
  "ung.autoRefresh": true,
  "ung.dateFormat": "YYYY-MM-DD",
  "ung.showStatusBar": true
}
```

**ung.cliPath** (string)
- Path to UNG CLI executable
- Default: `ung` (assumes it's in PATH)
- Example: `/usr/local/bin/ung` or `C:\\tools\\ung.exe`

**ung.defaultCurrency** (string)
- Default currency for new invoices
- Default: `USD`
- Options: USD, EUR, GBP, UAH, CAD, AUD

**ung.autoRefresh** (boolean)
- Auto-refresh views after operations
- Default: `true`

**ung.dateFormat** (string)
- Date format for display
- Default: `YYYY-MM-DD`
- Options: `YYYY-MM-DD`, `MM/DD/YYYY`, `DD/MM/YYYY`

**ung.showStatusBar** (boolean)
- Show active time tracking in status bar
- Default: `true`

## Workflows

### Freelance Project Workflow

1. **Setup** (one time):
   - Create your company profile
   - Add your client
   - Create a contract (hourly or fixed-price)

2. **Working**:
   - Start time tracking: Click play button in Time Tracking view
   - Work on your project
   - Status bar shows elapsed time
   - Stop tracking when done

3. **Billing**:
   - Use "Generate Invoice from Time" command
   - CLI will automatically calculate amount based on tracked hours
   - Export to PDF
   - Email to client

4. **Expenses**:
   - Log expenses as they occur
   - View monthly reports
   - Include in your accounting

### Agency Workflow

1. Manage multiple clients in the Clients view
2. Create separate contracts for each project
3. Team members can track time independently
4. Generate invoices per client/contract
5. Export batch invoices at month-end

## Troubleshooting

### UNG CLI Not Found

If you see "UNG CLI is not installed":
1. Verify installation: `ung --version` in terminal
2. If not in PATH, set full path in settings: `ung.cliPath`
3. Restart VSCode after installation

### Tree Views Not Loading

1. Check Output panel (`View > Output > UNG Operations`)
2. Ensure database exists: `~/.ung/ung.db`
3. Try manual refresh (click refresh icon)
4. Verify CLI works: Run `ung invoice ls` in terminal

### Commands Not Working

1. Check if commands are registered: `Cmd/Ctrl + Shift + P > UNG`
2. View Output channel for errors
3. Ensure you have required data (company, clients, contracts)
4. Some commands are interactive - they'll guide you through the CLI

## Known Limitations

- Some advanced features require using the CLI directly (indicated in command descriptions)
- Invoice/expense editing not yet fully implemented in UI (use CLI)
- No database migration support within extension
- Batch operations limited to those provided by CLI

## Roadmap

- [ ] Full inline editing for invoices and expenses
- [ ] Advanced filtering and search
- [ ] Dashboard with charts and analytics
- [ ] Integration with accounting software
- [ ] Custom themes for PDF exports
- [ ] Multi-workspace support
- [ ] Cloud sync integration (when available)

## Contributing

This extension wraps the UNG CLI tool. For CLI issues or feature requests:
- GitHub: https://github.com/Andriiklymiuk/ung
- Issues: https://github.com/Andriiklymiuk/ung/issues

## License

MIT License - See LICENSE file for details

## Support

- UNG CLI Documentation: https://github.com/Andriiklymiuk/ung
- Report Issues: https://github.com/Andriiklymiuk/ung/issues
- VSCode Extension Guide: https://code.visualstudio.com/docs/editor/extension-marketplace

---

**Made for freelancers who value their time ⏱️**

**Star ⭐ the project if this extension helps you get paid faster!**
