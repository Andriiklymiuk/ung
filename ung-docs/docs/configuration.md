---
id: configuration
title: Configuration
---

# Configuration

UNG stores all data in `~/.ung/` directory.

## Data Location

| Platform | Path |
|----------|------|
| macOS | `~/.ung/` |
| Linux | `~/.ung/` |
| Windows | `%USERPROFILE%\.ung\` |

## Directory Structure

```
~/.ung/
├── ung.db           # Main SQLite database
├── invoices/        # Generated invoice PDFs
├── contracts/       # Generated contract PDFs
└── config.yaml      # Configuration file (optional)
```

## Configuration File

Create `~/.ung/config.yaml` to customize settings:

```yaml
# Invoice settings
invoice:
  invoice_label: "INVOICE"
  bill_to_label: "Bill To"
  item_label: "Item"
  quantity_label: "Quantity"
  rate_label: "Rate"
  amount_label: "Amount"
  total_label: "Total"
  notes_label: "Notes"
  terms_label: "Terms & Conditions"
  terms: "Please make the payment by the due date."

# Default currency
currency: "USD"

# Date format (Go time format)
date_format: "02 Jan 2006"
```

## Invoice PDF Labels

You can customize invoice labels for localization:

```yaml
invoice:
  invoice_label: "FACTURE"      # French
  bill_to_label: "Facturer à"
  total_label: "Total"
```

## Database

UNG uses SQLite for local storage. The database is created automatically on first run.

### Backup

```bash
cp ~/.ung/ung.db ~/.ung/ung.db.backup
```

### Reset

```bash
rm ~/.ung/ung.db
# Next run will create a fresh database
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `UNG_HOME` | Override data directory | `~/.ung` |
| `UNG_CONFIG` | Override config file path | `~/.ung/config.yaml` |

## PDF Output

Generated PDFs are saved to:
- Invoices: `~/.ung/invoices/`
- Contracts: `~/.ung/contracts/`

Example filenames:
- `inv.acme.2025-01-15.pdf`
- `Acme_Software_Development_Contract.pdf`
