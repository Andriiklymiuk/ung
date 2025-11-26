# UNG - Billing & Time Tracking

A VS Code extension for managing billing, invoicing, and time tracking using the UNG CLI.

## Features

### Dashboard
- Monthly revenue projections
- Contract breakdown by type (hourly, retainer, fixed-price)
- Quick stats for clients, invoices, and expenses

### Client & Contract Management
- Create and manage company profiles and clients
- View contracts with rates and status
- Generate contract PDFs

### Invoice Management
- Create invoices or generate from tracked time
- Export to PDF and email directly
- Status tracking (pending, paid, overdue)

### Time Tracking
- Start/stop timers from VS Code
- Status bar shows elapsed time
- Log time manually for contracts

### Expense Tracking
- Log and categorize business expenses
- View expense reports by category

## Requirements

- VS Code 1.80.0+
- UNG CLI installed: `brew install andriiklymiuk/homebrew-ung/ung`

## Quick Start

1. Open Command Palette (`Cmd/Ctrl + Shift + P`)
2. Type "UNG" to see commands
3. Create your company: `UNG: Create Company`
4. Add clients and start tracking time

## Keyboard Shortcuts

| Command | Windows/Linux | macOS |
|---------|--------------|-------|
| Command Center | `Ctrl+Shift+U` | `Cmd+Shift+U` |
| Quick Start Tracking | `Ctrl+Alt+T` | `Cmd+Alt+T` |
| Create Invoice | `Ctrl+Alt+I` | `Cmd+Alt+I` |
| Search | `Ctrl+Alt+F` | `Cmd+Alt+F` |

## Configuration

Works out of the box. Optional settings:

| Setting | Default | Description |
|---------|---------|-------------|
| `ung.autoRefresh` | `true` | Auto-refresh views after operations |
| `ung.showStatusBar` | `true` | Show tracking in status bar |
| `ung.roundHours` | `true` | Round hours up for billing |
| `ung.roundRevenue` | `true` | Round revenue projections |

## Troubleshooting

**CLI Not Found**: Run `ung --version` in terminal. Restart VS Code after installation.

**Views Not Loading**: Check Output panel (`View > Output > UNG Operations`).

## Documentation

Full documentation: https://andriiklymiuk.github.io/ung/docs/intro

## License

MIT
