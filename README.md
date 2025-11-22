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

### ⏱ Time Tracking
- Start/stop timers
- Assign sessions to a client/project
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

## ⚙️ Usage Overview

```bash
ung [group] [command] [flags]
```

**Groups:**
- `company` — manage your own business details
- `client` — manage clients
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

### Convert tracked time → invoice item

```bash
ung track invoice --rate 40 --invoice 12
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
- **invoices** — invoice records
- **invoice_recipients** — links between invoices and clients
- **tracking_sessions** — time tracking records

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

### 3. Track time

```bash
ung track start --client 1
ung track stop
```

### 4. Create invoice

```bash
ung invoice new --company 1 --client 1 --price 120
```

### 5. Generate PDF

```bash
ung invoice pdf 2
```

---

## 📝 License

MIT

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
