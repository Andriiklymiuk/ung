# UNG — Freelance Billing Made Simple

A fast CLI tool for invoicing, time tracking, and client management. Works offline. Your data stays local.

## Install

```bash
# macOS/Linux
brew install andriiklymiuk/tools/ung

# Or with Go
go install github.com/Andriiklymiuk/ung@latest
```

## Quick Start

### 1. Initialize UNG

```bash
# Global setup (recommended) - data stored in ~/.ung/
ung config init --global

# OR project-specific - data stored in ./.ung/
ung config init
```

### 2. Set Up Your Business

```bash
# Add your company
ung company add --name "Your Company" --email "you@email.com"

# Add a client
ung client add --name "Acme Corp" --email "billing@acme.com"
```

### 3. Create a Contract

```bash
# Hourly contract ($100/hour)
ung contract add --client 1 --name "Web Development" --type hourly --rate 100

# Fixed-price contract
ung contract add --client 1 --name "Logo Design" --type fixed_price --price 2500
```

### 4. Track Time

```bash
# Start timer
ung track start --contract 1 --project "Homepage"

# ... work ...

# Stop timer
ung track stop

# Or log hours manually
ung track log --contract 1 --hours 3.5 --project "API integration"
```

### 5. Create Invoice

```bash
# From tracked time (auto-calculate from unbilled hours)
ung invoice --client "Acme Corp"

# Fixed amount
ung invoice new --company 1 --client 1 --price 1500 --currency USD

# Generate PDF
ung invoice pdf 1

# Email to client
ung invoice email 1
```

## Common Commands

| Command | Description |
|---------|-------------|
| `ung create` | Interactive wizard for everything |
| `ung company ls` | List companies |
| `ung client ls` | List clients |
| `ung contract ls` | List contracts |
| `ung invoice ls` | List invoices |
| `ung track ls` | List time entries |
| `ung track now` | Show active timer |
| `ung dashboard` | Revenue overview |
| `ung doctor` | Health check |

## Data Storage

```
~/.ung/                    # Global (default)
├── ung.db                 # SQLite database
├── config.yaml            # Settings
├── invoices/              # PDF invoices
└── contracts/             # PDF contracts

.ung/                      # Local (project-specific)
└── ...same structure...
```

Use `--global` flag with any command to force global storage.

## Configuration

Create `~/.ung/config.yaml` (global) or `.ung/config.yaml` (local):

```yaml
database_path: "~/.ung/ung.db"
invoices_dir: "~/.ung/invoices"
contracts_dir: "~/.ung/contracts"
language: "en"

invoice:
  invoice_label: "INVOICE"
  terms: "Payment due within 30 days."
```

## VS Code Extension

Install **UNG - Billing & Time Tracking** from the marketplace for a visual interface.

## Links

- **Docs**: https://andriiklymiuk.github.io/ung
- **Issues**: https://github.com/Andriiklymiuk/ung/issues

---

MIT License | Made for freelancers who value their time
