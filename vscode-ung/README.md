# UNG — Freelance Billing Made Simple (VS Code Extension)

A visual interface for invoicing, time tracking, and client management right inside VS Code. Works offline. Your data stays local.

## Install

### 1. Install the Extension

Search for **"UNG - Billing & Time Tracking"** in VS Code Extensions, or:

```
ext install andriiklymiuk.ung
```

### 2. Install the UNG CLI

The extension requires the UNG CLI tool:

```bash
# macOS/Linux
brew install andriiklymiuk/tools/ung

# Windows (Scoop)
scoop bucket add ung https://github.com/Andriiklymiuk/scoop-ung
scoop install ung

# Or with Go (any platform)
go install github.com/Andriiklymiuk/ung@latest
```

**Tip:** Use `UNG: Install UNG CLI` from the Command Palette if you need help installing.

## Quick Start

### 1. Initialize UNG

After installing, click the UNG icon in the Activity Bar. The Welcome screen will guide you:

- **Initialize (Global)** — Data stored in `~/.ung/` (recommended)
- **Initialize (Workspace)** — Data stored in `./.ung/` (project-specific)

### 2. Set Up Your Business

```
Command Palette (Cmd/Ctrl + Shift + P) → UNG: Create Company
```

Fill in your company name, email, address, and tax details.

### 3. Add Your First Client

```
Command Palette → UNG: Create Client
```

Or click the **+** button in the Clients panel.

### 4. Create a Contract

```
Command Palette → UNG: Create Contract
```

Choose from:
- **Hourly** — Bill by the hour ($100/hr)
- **Fixed Price** — One-time project ($2,500)
- **Retainer** — Monthly recurring ($3,000/mo)

### 5. Track Your Time

```
Cmd/Ctrl + Alt + T  →  Quick Start Time Tracking
```

Select a contract, add a description, and the timer starts. Watch it tick in the status bar.

### 6. Create an Invoice

```
Cmd/Ctrl + Alt + I  →  Create Invoice
```

Or generate from tracked time:
```
Cmd/Ctrl + Alt + G  →  Generate Invoice from Time
```

## Sidebar Views

Click the UNG icon in the Activity Bar to access:

| View | Description |
|------|-------------|
| **Dashboard** | Revenue overview, monthly projections, quick stats |
| **Invoices** | All invoices with status (pending/sent/paid/overdue) |
| **Contracts** | Active contracts with rates and types |
| **Clients** | Client list with contact info |
| **Expenses** | Business expenses by category |
| **Time Tracking** | Time entries and active sessions |

## Keyboard Shortcuts

| Action | Windows/Linux | macOS |
|--------|--------------|-------|
| Command Center | `Ctrl+Shift+U` | `Cmd+Shift+U` |
| Quick Start Tracking | `Ctrl+Alt+T` | `Cmd+Alt+T` |
| Stop Tracking | `Ctrl+Alt+S` | `Cmd+Alt+S` |
| Create Invoice | `Ctrl+Alt+I` | `Cmd+Alt+I` |
| Generate from Time | `Ctrl+Alt+G` | `Cmd+Alt+G` |
| Search Everything | `Ctrl+Alt+F` | `Cmd+Alt+F` |
| Quick Actions | `Ctrl+Alt+U` | `Cmd+Alt+U` |
| Statistics & Reports | `Ctrl+Alt+D` | `Cmd+Alt+D` |
| Refresh Dashboard | `Ctrl+Alt+R` | `Cmd+Alt+R` |

## All Commands

### Core Commands
| Command | Description |
|---------|-------------|
| `UNG: Command Center` | Central hub for all UNG actions |
| `UNG: Quick Actions` | Fast access to common tasks |
| `UNG: Search Everything` | Search across invoices, clients, contracts |
| `UNG: Open Dashboard` | Open the main dashboard view |
| `UNG: Refresh All Views` | Refresh all sidebar panels |

### Company & Clients
| Command | Description |
|---------|-------------|
| `UNG: Create Company` | Add your business profile |
| `UNG: Edit Company` | Update company details |
| `UNG: Create Client` | Add a new client |
| `UNG: Edit Client` | Modify client info |
| `UNG: Delete Client` | Remove a client |
| `UNG: List Clients` | Show all clients |
| `UNG: View Client Details` | See full client profile |
| `UNG: Search Clients` | Find clients by name |

### Contracts
| Command | Description |
|---------|-------------|
| `UNG: Create Contract` | New hourly/fixed/retainer contract |
| `UNG: Edit Contract` | Modify contract terms |
| `UNG: Delete Contract` | Remove a contract |
| `UNG: View Contract` | See contract details |
| `UNG: Generate Contract PDF` | Export contract to PDF |
| `UNG: Search Contracts` | Find contracts |

### Invoices
| Command | Description |
|---------|-------------|
| `UNG: Create Invoice` | Create a new invoice |
| `UNG: Generate Invoice from Time` | Auto-create from tracked hours |
| `UNG: Generate All Invoices` | Batch create invoices |
| `UNG: View Invoice` | Preview invoice |
| `UNG: Edit Invoice` | Modify invoice |
| `UNG: Duplicate Invoice` | Copy an existing invoice |
| `UNG: Delete Invoice` | Remove invoice |
| `UNG: Export Invoice to PDF` | Generate PDF |
| `UNG: Email Invoice` | Send invoice to client |
| `UNG: Mark Invoice as Paid` | Update status to paid |
| `UNG: Mark Invoice as Sent` | Update status to sent |
| `UNG: Change Invoice Status` | Set any status |
| `UNG: Send All Pending Invoices` | Batch email invoices |
| `UNG: Manage Recurring Invoices` | Set up recurring billing |
| `UNG: Search Invoices` | Find invoices |

### Time Tracking
| Command | Description |
|---------|-------------|
| `UNG: Start Time Tracking` | Start the timer |
| `UNG: Stop Time Tracking` | Stop and save time |
| `UNG: Toggle Time Tracking` | Start/stop toggle |
| `UNG: Quick Start Time Tracking` | Fast start with contract picker |
| `UNG: Log Time Manually` | Add time entry without timer |
| `UNG: View Active Tracking Session` | See current timer |
| `UNG: Edit Tracking Session` | Modify time entry |
| `UNG: Delete Tracking Session` | Remove time entry |
| `UNG: Start Pomodoro Timer` | 25-min focused work session |

### Expenses
| Command | Description |
|---------|-------------|
| `UNG: Log Expense` | Add a business expense |
| `UNG: Edit Expense` | Modify expense |
| `UNG: Delete Expense` | Remove expense |
| `UNG: View Expense Report` | See expense summary |

### Reports & Analytics
| Command | Description |
|---------|-------------|
| `UNG: Open Statistics & Reports` | Full analytics view |
| `UNG: Open Profit Dashboard` | Revenue vs expenses |
| `UNG: View Weekly Report` | This week's summary |
| `UNG: View Monthly Report` | This month's summary |
| `UNG: Business Insights` | Trends and projections |
| `UNG: View Goal Progress` | Track income goals |

### Planning & Rates
| Command | Description |
|---------|-------------|
| `UNG: Set Income Goal` | Monthly/yearly targets |
| `UNG: Calculate Hourly Rate` | Rate calculator tool |
| `UNG: Analyze Actual Rates` | Compare planned vs actual |

### Data Management
| Command | Description |
|---------|-------------|
| `UNG: Export Data for Accounting` | Export for tax prep |
| `UNG: Open Export Wizard` | Guided data export |
| `UNG: Import Data from CSV` | Import existing data |
| `UNG: Backup & Sync Data` | Backup your database |
| `UNG: Open Template Editor` | Customize invoice templates |

### Setup & Maintenance
| Command | Description |
|---------|-------------|
| `UNG: Install UNG CLI` | Install CLI tool |
| `UNG: Check for Updates` | Check for CLI updates |
| `UNG: Recheck CLI Installation` | Verify CLI is working |
| `UNG: Initialize UNG (Global)` | Setup global config |
| `UNG: Initialize UNG (Workspace)` | Setup local config |
| `UNG: Open Documentation` | Open online docs |

## Configuration

Access via `Settings → Extensions → UNG`:

| Setting | Default | Description |
|---------|---------|-------------|
| `ung.useGlobalConfig` | `true` | Use `~/.ung/` for data. Set `false` for workspace `.ung/` |
| `ung.autoRefresh` | `true` | Auto-refresh views after operations |
| `ung.showStatusBar` | `true` | Show active timer in status bar |
| `ung.roundHours` | `true` | Round hours up (9.5 → 10) for billing |
| `ung.roundRevenue` | `true` | Round revenue to nearest dollar |

## Data Storage

Your data is stored locally:

```
~/.ung/                    # Global (default)
├── ung.db                 # SQLite database
├── config.yaml            # Settings
├── invoices/              # Generated PDFs
└── contracts/             # Contract PDFs

.ung/                      # Workspace (if configured)
└── ...same structure...
```

Toggle between global/local with `ung.useGlobalConfig` setting.

## Typical Workflow

```
┌─────────────────────────────────────────────────────────────┐
│  1. START DAY                                               │
│     Cmd+Alt+T → Select contract → Start tracking            │
├─────────────────────────────────────────────────────────────┤
│  2. WORK                                                    │
│     Status bar shows elapsed time                           │
│     Switch contracts as needed                              │
├─────────────────────────────────────────────────────────────┤
│  3. END DAY                                                 │
│     Cmd+Alt+S → Stop tracking                               │
│     Time saved automatically                                │
├─────────────────────────────────────────────────────────────┤
│  4. INVOICE (weekly/monthly)                                │
│     Cmd+Alt+G → Generate from tracked time                  │
│     Review → Export PDF → Email to client                   │
├─────────────────────────────────────────────────────────────┤
│  5. GET PAID                                                │
│     Right-click invoice → Mark as Paid                      │
│     Dashboard updates automatically                         │
└─────────────────────────────────────────────────────────────┘
```

## Right-Click Context Menus

### On Invoices
- View Invoice
- Export to PDF
- Email Invoice
- Mark as Paid/Sent
- Change Status
- Edit/Duplicate/Delete

### On Contracts
- View Contract
- Generate PDF
- Edit/Delete

### On Clients
- Edit/Delete

### On Expenses
- Edit/Delete

### On Time Entries
- View Session
- Edit/Delete

## Status Bar

When tracking time, the status bar shows:

```
⏱ 2:34:15 | Project Name
```

Click to see session details or stop tracking.

## Troubleshooting

### CLI Not Found
1. Verify installation: Open terminal, run `ung --version`
2. If not installed, use `UNG: Install UNG CLI` command
3. Restart VS Code after installation
4. If still not found, ensure CLI is in your PATH

### Views Not Loading
1. Check Output panel: `View → Output → UNG Operations`
2. Try `UNG: Refresh All Views`
3. Run `UNG: Recheck CLI Installation`

### Data Not Syncing
- The extension and CLI share the same database
- Check `ung.useGlobalConfig` setting matches your CLI setup
- Default is global (`~/.ung/`), same as CLI

### Time Tracking Issues
- Only one timer can run at a time
- Use `UNG: View Active Tracking Session` to see current timer
- Status bar shows active timer when running

## Links

- **Documentation**: https://andriiklymiuk.github.io/ung/docs/intro
- **Issues**: https://github.com/Andriiklymiuk/ung/issues
- **CLI Tool**: https://github.com/Andriiklymiuk/ung

---

MIT License | Made for freelancers who value their time
