---
id: intro
title: Introduction
sidebar_position: 1
---

# UNG - Universal Next-Gen Billing & Tracking

UNG is a fast, cross-platform CLI tool for freelancers and small businesses to manage:

- **Company details** - Store your business information
- **Clients** - Manage client contacts and details
- **Contracts** - Create and track contracts with PDF generation
- **Invoices** - Generate professional PDF invoices
- **Time tracking** - Track billable hours per client/project
- **Expenses** - Log and categorize business expenses

All data is stored locally in `~/.ung/ung.db` (SQLite database).

## Key Features

### Time Tracking
Track billable hours with automatic contract rate calculation:
```bash
ung track start --client "Acme Corp"
ung track stop
ung track log --client "Acme" --hours 8 --description "Feature development"
```

### Invoice Generation
Create professional PDF invoices from tracked time:
```bash
ung invoice acme           # Auto-generate from tracked time
ung invoice pdf 1          # Generate PDF for invoice #1
ung invoice email 1        # Export to email client
```

### Contract Management
Store contracts and generate PDFs:
```bash
ung contract add
ung contract pdf 1
```

### Expense Tracking
Log business expenses:
```bash
ung expense add --amount 50 --category "Software" --description "GitHub Pro"
ung expense report
```

### Dashboard
View revenue projections and business metrics:
```bash
ung dashboard
```

## Why UNG?

- **Privacy-first** - All data stored locally, no cloud required
- **Fast** - Native Go binary, instant startup
- **Simple** - Intuitive CLI interface
- **Flexible** - Works with hourly or fixed-price contracts
- **Cross-platform** - macOS, Linux, Windows support
