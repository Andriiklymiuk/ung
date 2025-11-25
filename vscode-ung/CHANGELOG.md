# Changelog

All notable changes to the UNG VSCode extension will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-11-25

### Added

#### Core Features
- Complete VSCode extension for UNG CLI tool
- Integration with UNG CLI for all billing and time tracking operations
- Dedicated activity bar with 5 organized views

#### Company & Client Management
- Company creation with full details (tax ID, bank info, registration address)
- Client management with add, edit, delete operations
- Company and client list views

#### Contract Management
- Contract tree view provider showing all contracts
- Contract status indicators (active/inactive)
- Generate contract PDFs
- Email contracts to clients
- View contract details

#### Invoice Management
- Invoice tree view with status badges (pending, paid, overdue)
- Create invoices with interactive forms
- Generate invoices from tracked time
- Invoice detail webview panel
- Export invoices to PDF
- Email invoices directly from VSCode
- Context menu actions for invoices

#### Expense Tracking
- Expense tree view provider
- Log expenses with categories
- Expense report generation
- Category-based organization (software, hardware, travel, meals, etc.)
- Monthly expense summaries

#### Time Tracking
- Time tracking tree view showing all sessions
- Start/stop timer commands
- Manual time logging
- Active session display in status bar
- Real-time status bar updates (every 5 seconds)
- Billable/non-billable tracking
- Project and client association

#### UI/UX Features
- Context menus for all tree items
- Status bar integration for active tracking
- Progress indicators for long operations
- User-friendly error messages
- Confirmation prompts for destructive actions
- Interactive forms with validation
- Output channel for detailed logging

#### Configuration
- Configurable CLI path
- Default currency setting
- Date format preferences
- Auto-refresh toggle
- Status bar visibility control

#### Developer Features
- TypeScript with strict mode
- Comprehensive JSDoc comments
- Unit and integration tests
- CLI wrapper for command execution
- Modular architecture with separation of concerns
- Tree view providers for data display
- Webview panels for rich UI
- Event-driven refresh mechanism

### Technical Implementation

#### Architecture
- **src/cli/ungCli.ts** - CLI wrapper for executing commands
- **src/commands/** - Command handlers for all operations
- **src/views/** - Tree view providers for sidebar
- **src/webview/** - Webview panels for rich UI
- **src/utils/** - Configuration, formatting, and status bar utilities
- **src/extension.ts** - Main extension entry point

#### Commands Registered
- 30+ VSCode commands for all UNG operations
- Company: create, edit
- Client: create, edit, delete, list
- Contract: create, view, edit, delete, generate PDF
- Invoice: create, generate from time, view, edit, delete, export, email
- Expense: log, edit, delete, view report
- Tracking: start, stop, log manually, view active

#### Tree Views
- ungInvoices - Invoice list with status
- ungContracts - Contract list with rates
- ungClients - Client list with emails
- ungExpenses - Expense list with categories
- ungTracking - Tracking sessions with durations

#### Webviews
- Invoice detail panel with export/email actions
- Export wizard for batch operations (placeholder)

### Dependencies
- VSCode Engine: ^1.80.0
- TypeScript 5.0
- @vscode/test-electron 2.3.0
- ESLint & TypeScript ESLint

### Requirements
- VSCode 1.80.0 or higher
- UNG CLI installed and in PATH
- Node.js for development

### Known Issues
- Some CLI operations require interactive terminal (noted in commands)
- Invoice/expense inline editing not yet implemented (use CLI)
- Limited batch operations support

### Documentation
- Comprehensive README with usage guide
- Installation instructions
- Configuration reference
- Workflow examples
- Troubleshooting guide

---

## [Unreleased]

### Planned Features
- Full inline editing for invoices
- Advanced filtering and search
- Dashboard with analytics
- Custom PDF templates
- Multi-workspace support
- Integration with accounting software

---

**Note**: This is the initial release of the UNG VSCode extension. Future updates will focus on enhancing UI capabilities and reducing reliance on CLI for interactive operations.
