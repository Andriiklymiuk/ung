# 🐾 UNG — Universal Next-Gen Billing & Tracking (CLI)

UNG is a fast, cross-platform command-line tool written in Go for managing company details, clients, invoices, and time tracking — all in one place.
It uses SQLite as the backend and works both as a standalone CLI and the data foundation for an optional future macOS Swift app.

UNG's goal is to make invoicing, billing, and tracking simple, scriptable, and universal for freelancers and small teams.

---

## ✨ Features

### 🧾 Company & Client Management
- Add your own business information
- Add multiple clients
- Edit & list all entities

### 📄 Invoice Generation
- Create invoices tied to company + client
- Auto-generate invoice numbers
- Store PDFs locally
- Track statuses: pending, sent, paid, overdue

### 📋 Contract Management
- Track hourly rate, fixed price, and retainer contracts
- Link contracts to clients
- Automatically calculate billable amounts
- Interactive prompts for easy data entry

### ⏱ Time Tracking
- Start/stop timers
- Manually log hours worked
- Assign sessions to contracts with hourly rates
- Automatic billable amount calculation
- Convert tracked time → invoice line items

### 📂 Unified SQLite Database
- All data stored in `~/.ung/ung.db`
- CLI and macOS Swift app share the same database

### 🔐 Secure
- User data stored locally — never transmitted
- You own everything

---

## 🚀 Installation

```bash
go install github.com/Andriiklymiuk/ung@latest
```

Or build from source:

```bash
git clone https://github.com/Andriiklymiuk/ung.git
cd ung
go build -o ung
```

---

## ⚡ Quick Start

New to UNG? Just run this command and follow the prompts:

```bash
ung contract add
```

UNG will automatically guide you through:
1. Adding your company information (if not set up)
2. Adding your first client (if not set up)
3. Creating your first contract with hourly rate or fixed price

Then you can immediately start tracking time:

```bash
ung track log
```

Select your contract, enter hours worked, and UNG automatically calculates your billable amount!

---

## ⚙️ Usage Overview

```bash
ung [group] [command] [flags]
```

**Groups:**
- `company` — manage your own business details
- `client` — manage clients
- `contract` — manage contracts with clients
- `invoice` — create & list invoices
- `track` — time tracking utilities
- `config` — global settings

---

## 🏢 COMPANY COMMANDS

### Add your business (you)

```bash
ung company add \
  --name "John Doe Studio" \
  --email "hello@doe.dev" \
  --address "Kyiv, Ukraine" \
  --tax-id "UA12345"
```

### List companies

```bash
ung company ls
```

### Edit company

```bash
ung company edit 1 --address "New Address 202"
```

---

## 🤝 CLIENT COMMANDS

### Add a client

```bash
ung client add \
  --name "Acme Inc." \
  --email "billing@acme.com" \
  --address "Business Street 12" \
  --tax-id "EU88774411"
```

### List clients

```bash
ung client ls
```

### Edit client

```bash
ung client edit 3 --email "new@acme.com"
```

---

## 📋 CONTRACT COMMANDS

Contracts link clients to specific work agreements with hourly rates or fixed prices.

### Add a contract (interactive mode - no flags needed!)

```bash
ung contract add
```

**✨ Smart Guided Setup:** If you haven't added your company info or any clients yet, UNG will guide you through the process step-by-step! Just run the command and it will:
1. Check if you've added your company → prompt to add it if not
2. Check if you've added any clients → prompt to add one if not
3. Then help you create your first contract

This will then prompt you to:
- Select a client
- Enter contract name
- Choose contract type (hourly, fixed_price, retainer)
- Enter rate or price
- Set currency

### Add a contract (with flags)

```bash
ung contract add \
  --client 1 \
  --name "Website Development Q1 2025" \
  --type hourly \
  --rate 75 \
  --currency USD
```

### List contracts

```bash
ung contract ls
```

Output shows client, type, and rates:
```
ID  NAME                    CLIENT     TYPE    RATE/PRICE    ACTIVE
1   Website Development     Acme Corp  hourly  75.00 USD/hr  ✓
2   Logo Design Project     TechCo     fixed   5000.00 USD   ✓
```

### Edit contract (mark inactive)

```bash
ung contract edit 1
```

---

## 🧾 INVOICE COMMANDS

### Create an invoice

```bash
ung invoice new \
  --company 1 \
  --client 3 \
  --price 600 \
  --currency EUR \
  --description "Development work for Jan 2026" \
  --due "2026-02-10"
```

### Generate & export PDF

```bash
ung invoice pdf 12
```

### Send via email (opens email client auto-filled)

```bash
ung invoice send 12
```

### List invoices

```bash
ung invoice ls
```

---

## ⏱ TIME TRACKING

The track module lets you record billable or non-billable time.

### Log hours worked (interactive mode - recommended!)

```bash
ung track log
```

**✨ Smart Guided Setup:** If you haven't set up contracts yet, UNG will guide you! It will prompt you to add your company, clients, and create a contract before you can track time.

This will prompt you to:
- Select a contract (shows hourly rates for easy reference)
- Enter hours worked
- Add project/task name
- Add notes

**Automatically calculates billable amount** based on contract hourly rate!

Example output:
```
✓ Time logged successfully (Session ID: 1)
  Client: Acme Corp
  Contract: Website Development
  Hours: 2.50
  Billable Amount: 187.50 USD
  Project: Homepage redesign
```

### Log hours (with flags)

```bash
ung track log --contract 1 --hours 2.5 --project "Homepage redesign"
```

### Start a timer

```bash
ung track start --client 3 --project "Landing Page"
```

### Stop current timer

```bash
ung track stop
```

### Show ongoing timer

```bash
ung track now
```

### Show all tracked sessions

```bash
ung track ls
```

---

## 📂 Directory Structure

UNG stores everything here:

```
~/.ung/
    ung.db
    invoices/
        INV-2025-001.pdf
        INV-2025-002.pdf
```

You can back it up, sync it, or point a Swift app to it.

---

## 🧱 Database Schema

- **companies** — your business entities
- **clients** — client/customer information
- **contracts** — work agreements with hourly rates or fixed prices
- **invoices** — invoice records
- **invoice_recipients** — links between invoices and clients
- **tracking_sessions** — time tracking records linked to contracts

---

## 🌱 Roadmap (future features)

- macOS SwiftUI app using same DB
- Email templates
- Invoice branding templates
- Recurring invoices
- Export to CSV, JSON, or Markdown
- Cloud sync (optional)

---

## 💡 Example Workflow

### 1. Add yourself

```bash
ung company add --name "Andrii" --email "andrii@example.com"
```

### 2. Add a client

```bash
ung client add --name "TechCorp" --email "billing@techcorp.com"
```

### 3. Create a contract

```bash
ung contract add --client 1 --name "Web Dev Q1" --type hourly --rate 75 --currency USD
```

### 4. Log your work (interactive)

```bash
ung track log
# Select contract, enter hours (e.g., 2.5), add notes
# Billable amount calculated automatically!
```

### 5. Create invoice

```bash
ung invoice new --company 1 --client 1 --price 187.50 --currency USD
```

### 6. Generate PDF

```bash
ung invoice pdf 1
```

---

## 📝 License

MIT

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
