---
id: quickstart
title: Quick Start
---

# Quick Start

Get started with UNG in 5 minutes.

## 1. Set Up Your Company

First, add your company details:

```bash
ung company add
```

This interactive wizard will prompt for:
- Company name
- Email
- Address
- Tax ID
- Bank details (for invoices)

## 2. Add a Client

```bash
ung client add
```

Or with flags:
```bash
ung client add --name "Acme Corp" --email "billing@acme.com"
```

## 3. Create a Contract

```bash
ung contract add
```

Choose between:
- **Hourly** - Set an hourly rate (e.g., $100/hour)
- **Fixed price** - Set a project price (e.g., $5000)

## 4. Track Time

### Start a session
```bash
ung track start --client "Acme"
```

### Stop the session
```bash
ung track stop
```

### Or log hours manually
```bash
ung track log --client "Acme" --hours 8 --description "Feature development"
```

## 5. Generate Invoice

Create an invoice from tracked time:
```bash
ung invoice acme
```

This will:
1. Find unbilled time for "Acme Corp"
2. Calculate the amount based on contract rate
3. Create an invoice
4. Generate a PDF

## 6. Send the Invoice

Export to your email client:
```bash
ung invoice email 1
```

Choose from:
- Apple Mail (with auto-attachment)
- Outlook
- Gmail (browser)

## Common Commands

| Command | Description |
|---------|-------------|
| `ung company ls` | List companies |
| `ung client ls` | List clients |
| `ung contract ls` | List contracts |
| `ung invoice ls` | List invoices |
| `ung track ls` | List time sessions |
| `ung dashboard` | View revenue dashboard |
| `ung expense add` | Log an expense |

## Next Steps

- Read the [Configuration Guide](./configuration) to customize labels and settings
- Explore the [CLI Reference](./cli/ung) for all commands
