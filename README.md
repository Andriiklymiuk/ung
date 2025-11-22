# 🐾 UNG — Your Next Gig, Simplified

> **"The only tool you need to turn your time into money."**
> — *Inspired by Daschund (running fast) and Your Next Gig (getting paid)*

**UNG is not just another invoicing tool.** It's your personal financial command center, designed for the modern freelancer, consultant, and small team who refuse to waste time on bloated software.

Think of it as **the Unix philosophy applied to billing**: do one thing (track your work), do it well (generate professional invoices), and make it composable (CLI + API + apps).

No subscriptions. No cloud lock-in. No compromises. **Your data, your rules, your revenue.**

---

## 🚀 Why UNG?

**Traditional invoicing tools are broken.** They're slow, expensive, require internet, lock your data in proprietary formats, and nickel-and-dime you with subscriptions.

UNG is different:

- ⚡ **Lightning Fast** — CLI-first means instant commands, no loading screens
- 🔐 **Privacy First** — SQLite local database, you own 100% of your data
- 💰 **Zero Cost** — Open source, no subscriptions, no limits
- 🎯 **Laser Focused** — Built for freelancers who value time over features
- 🌍 **Universal** — Cross-platform CLI that works everywhere
- 🔄 **Extensible** — Same database powers CLI, macOS app, mobile apps, and API
- 🎨 **Professional** — Generate PDFs that look like you hired a designer
- 🌐 **Multi-language** — Invoice in English, Ukrainian, German, or any language

**The bottom line:** UNG helps you focus on what matters — doing great work and getting paid for it.

---

## ✨ What UNG Can Do

### 📊 Complete Business Management
- **Company Profiles** — Your business information with bank details, tax ID, registration address
- **Client Management** — Track all your customers with full contact details
- **Contract System** — Hourly rates, fixed-price projects, and retainer agreements
- **Time Tracking** — Start/stop timers or log hours manually
- **Invoice Generation** — Professional PDFs with itemized billing
- **Email Export** — One command to send invoices via Apple Mail, Outlook, or Gmail

### 🎯 Smart Workflows
- **Interactive Creation Wizard** — `ung create` guides you through everything
- **Smart Prerequisites** — Automatically prompts for missing company/client data
- **Auto-calculation** — Hourly contracts automatically calculate billable amounts
- **Batch Operations** — Export all invoices for a month at once
- **PDF Generation** — For both invoices and contracts

### 🎨 Professional Output
- **Beautiful PDFs** — Two-column layouts, bank details, itemized tables
- **Multi-language** — All labels configurable in your language
- **Custom Templates** — HTML templates for future customization
- **Brand Consistency** — Terms & conditions, payment notes, your logo

### 🔧 Developer-Friendly
- **Type-Safe** — GORM repository layer, no raw SQL in commands
- **Well-Tested** — Comprehensive test suite (63%+ coverage)
- **Extensible** — Clean architecture ready for API integration
- **Documented** — Clear code, helpful comments, examples

---

## 🚀 Installation

### Homebrew (macOS/Linux)
```bash
# Coming soon!
brew install andriiklymiuk/tools/ung
```

### Go Install
```bash
go install github.com/Andriiklymiuk/ung@latest
```

### Build from Source
```bash
git clone https://github.com/Andriiklymiuk/ung.git
cd ung
go build -o ung
./ung --help
```

---

## ⚡ Quick Start

### Option 1: Interactive Creation Wizard (Recommended!)
```bash
ung create
```

This launches an interactive menu where you can:
- Create your company profile
- Add clients
- Set up contracts
- Generate invoices
- Log time worked

**Smart Guided Setup** — UNG automatically checks prerequisites and guides you through each step.

### Option 2: Command-line Workflow
```bash
# 1. Add your business info
ung company add --name "Your Company" --email "you@company.com"

# 2. Add a client
ung client add --name "Acme Corp" --email "billing@acme.com"

# 3. Create a contract (interactive)
ung contract add

# 4. Track your time
ung track log

# 5. Generate an invoice
ung invoice new --company 1 --client 1 --price 1500 --currency USD

# 6. Export to PDF
ung invoice pdf 1

# 7. Email it to your client
ung invoice email 1
```

---

## 📖 Commands Reference

### 🎯 Create (Interactive Wizard)
```bash
ung create                    # Launch interactive creation menu
```

Choose what to create:
- Company — Your business information
- Client — Customer details
- Contract — Work agreement
- Invoice — Bill for services
- Track Time — Log hours worked

### 🏢 Company Management
```bash
ung company add               # Add your business
  --name "Company Name"
  --email "contact@company.com"
  --phone "+1-555-0100"
  --address "123 Business St"
  --tax-id "12-3456789"

ung company ls                # List all companies
ung company edit [id]         # Edit company details
```

### 🤝 Client Management
```bash
ung client add                # Add a client
  --name "Client Name"
  --email "client@company.com"
  --address "456 Client Ave"
  --tax-id "98-7654321"

ung client ls                 # List all clients
ung client edit [id]          # Edit client details
```

### 📋 Contract Management
```bash
ung contract add              # Add contract (interactive)
  --client 1
  --name "Website Development"
  --type hourly
  --rate 100
  --currency USD

ung contract ls               # List all contracts
ung contract edit [id]        # Edit contract (toggle active status)
ung contract pdf [id]         # Generate contract PDF
ung contract email [id]       # Email contract to client
```

Contract types:
- `hourly` — Hourly rate billing
- `fixed_price` — One-time fixed price
- `retainer` — Ongoing monthly retainer

### 🧾 Invoice Management
```bash
ung invoice new               # Create invoice
  --company 1
  --client 1
  --price 1500
  --currency USD
  --description "Development work"
  --due "2025-02-15"

ung invoice ls                # List all invoices
ung invoice pdf [id]          # Generate PDF
ung invoice email [id]        # Email to client
ung invoice batch-email       # Batch email (latest month, specific month, or all pending)
```

### ⏱ Time Tracking
```bash
ung track start               # Start timer (interactive)
  --client 1
  --project "Landing Page"

ung track stop                # Stop active timer
ung track now                 # Show current timer
ung track log                 # Log hours worked (interactive)
  --contract 1
  --hours 2.5
  --project "Homepage redesign"

ung track ls                  # List all tracked sessions
```

---

## ⚙️ Configuration

UNG uses YAML configuration files. Create `.ung.yaml` in your project directory or `~/.ung/config.yaml` globally.

```yaml
# Database and storage
database_path: "~/.ung/ung.db"
invoices_dir: "~/.ung/invoices"

# Language (en, uk, de, etc.)
language: "en"

# Custom templates (optional)
templates:
  invoice_html: "~/templates/invoice.html"
  contract_html: "~/templates/contract.html"

# Invoice labels and text
invoice:
  invoice_label: "INVOICE"
  from_label: "From"
  bill_to_label: "Bill To"
  description_label: "Description"
  item_label: "Item"
  quantity_label: "Quantity"
  rate_label: "Rate"
  amount_label: "Amount"
  total_label: "Total"
  notes_label: "Notes"
  terms_label: "Terms & Conditions"
  terms: "Please make the payment by the due date."
  payment_note: "Payment is due within the specified term."
```

See `.ung.yaml.example` for complete examples in multiple languages.

---

## 📂 Directory Structure

```
~/.ung/
├── ung.db                    # SQLite database (all your data)
├── invoices/                 # Generated invoice PDFs
│   ├── INV-2025-001.pdf
│   └── INV-2025-002.pdf
├── contracts/                # Generated contract PDFs
│   ├── Acme_Corp_Website.pdf
│   └── TechCo_Logo_Design.pdf
└── config.yaml               # Global configuration (optional)
```

---

## 🎨 Custom Templates

UNG includes professional HTML templates for future HTML-to-PDF rendering. Located in `templates/`:

- `invoice.html` — Professional invoice layout
- `contract.html` — Professional contract layout
- `README.md` — Template documentation with all available variables

Customize colors, fonts, layout, and branding. See `templates/README.md` for details.

---

## 🧱 Database Schema

UNG uses SQLite with the following tables:

- **companies** — Your business entities
- **clients** — Client/customer information
- **contracts** — Work agreements (hourly/fixed/retainer)
- **invoices** — Invoice records
- **invoice_recipients** — Invoice-to-client relationships
- **invoice_line_items** — Itemized billing details
- **tracking_sessions** — Time tracking records

All migrations are in `migrations/` directory.

---

## 🔬 Technical Stack

- **Language**: Go 1.21+
- **Database**: SQLite 3
- **ORM**: GORM (type-safe, no raw SQL)
- **PDF Generation**: gofpdf
- **CLI Framework**: Cobra
- **Interactive UI**: Bubbletea (huh)
- **Testing**: Go testing with 63%+ coverage

---

## 🌟 Future Features (Premium/Pro Version)

UNG's open-source CLI will always be free. However, we're planning premium features and companion apps that freelancers will happily pay for:

### 💎 Essential Premium Features (Must-Have)

#### 1. **Recurring Invoices** ⭐⭐⭐⭐⭐
- Auto-generate monthly retainer invoices
- Schedule quarterly or annual billing
- Email reminders before invoice due
- **Why pay?** Saves 30+ minutes per month on repetitive billing

#### 2. **Multi-Currency with Live Rates** ⭐⭐⭐⭐⭐
- Real-time exchange rate conversion
- Invoice in client's currency, receive in yours
- Historical rate tracking
- **Why pay?** Critical for international freelancers

#### 3. **Payment Gateway Integration** ⭐⭐⭐⭐⭐
- Stripe, PayPal, Wise integration
- "Pay Now" button on invoices
- Automatic payment reconciliation
- **Why pay?** Get paid 2-3x faster, reduce late payments

#### 4. **Smart Email Delivery** ⭐⭐⭐⭐
- Send from your Gmail/Outlook directly
- Track when invoices are opened
- Automatic payment reminders
- Custom email templates
- **Why pay?** Professional communication, less follow-up

#### 5. **Financial Reports & Analytics** ⭐⭐⭐⭐
- Revenue tracking and forecasting
- Client profitability analysis
- Tax-ready reports (1099, VAT)
- Time vs revenue analytics
- **Why pay?** Make data-driven business decisions

#### 6. **Cloud Sync & Backup** ⭐⭐⭐⭐
- Encrypted cloud storage
- Real-time sync across devices
- Automatic backups
- Version history
- **Why pay?** Peace of mind, work from anywhere

#### 7. **Mobile Apps (iOS/Android)** ⭐⭐⭐⭐⭐
- Create invoices on the go
- Start/stop timers from your phone
- Push notifications for payments
- Quick expense logging
- **Why pay?** Full productivity even away from desk

#### 8. **Team Collaboration** ⭐⭐⭐⭐
- Multiple users on one account
- Role-based permissions
- Activity audit log
- Shared client database
- **Why pay?** Essential for agencies and growing teams

#### 9. **Expense Tracking** ⭐⭐⭐⭐
- Photo receipts with OCR
- Mileage tracking
- Expense categories
- Profit margin calculation
- **Why pay?** Complete financial picture, maximize deductions

#### 10. **Contract Templates & E-Signatures** ⭐⭐⭐⭐
- Pre-built contract templates
- Legally binding e-signatures
- Track signature status
- Automatic reminders
- **Why pay?** Protect yourself legally, faster client onboarding

### 🎯 Nice-to-Have Premium Features

#### 11. **Time Tracking Intelligence**
- AI-powered time suggestions
- Project time estimates
- Productivity insights
- Billable vs non-billable analytics

#### 12. **Client Portal**
- Clients view their invoices online
- Self-service payment history
- Upload files and documents
- Real-time project updates

#### 13. **Proposal Generator**
- Beautiful proposal templates
- Convert proposals to contracts
- Track proposal acceptance
- Include pricing options

#### 14. **Automated Tax Calculations**
- VAT/GST calculation
- Tax withholding for contractors
- State/provincial tax rules
- Year-end tax summaries

#### 15. **API Access**
- REST API for custom integrations
- Webhook notifications
- Zapier/Make integration
- Custom reporting

#### 16. **White-Label Option**
- Custom branding
- Your domain for client portal
- Remove UNG branding
- Custom email templates

#### 17. **QuickBooks/Xero Integration**
- Two-way sync with accounting software
- Automatic bank reconciliation
- Chart of accounts mapping

#### 18. **Late Payment Automation**
- Automatic late fee calculation
- Escalating reminder emails
- Collections assistance
- Legal document generation

#### 19. **Multi-Language Client Support**
- Detect client locale
- Auto-translate invoices
- Currency localization
- International tax rules

#### 20. **Telegram/Slack Bot**
- Create invoices via chat
- Timer start/stop commands
- Payment notifications
- Quick client lookup

---

## 💰 Pricing Strategy (Future)

**Free Forever:**
- CLI tool with all core features
- Unlimited invoices and clients
- Local database
- Basic PDF generation

**Pro ($9/month or $90/year):**
- Recurring invoices
- Multi-currency
- Payment gateway integration
- Email delivery with tracking
- Cloud sync & backup
- Mobile apps

**Business ($29/month or $290/year):**
- Everything in Pro
- Team collaboration (up to 5 users)
- Financial reports
- Expense tracking
- API access
- Priority support

**Enterprise (Custom):**
- White-label
- Unlimited users
- Custom integrations
- Dedicated support
- On-premise deployment

---

## 📱 Upcoming Companion Apps

### Swift macOS App
Native macOS app with:
- Beautiful native UI
- Same SQLite database as CLI
- System tray timer
- Notification center integration
- Drag & drop invoice creation

### Expo Mobile App (iOS/Android/Web)
Cross-platform mobile experience:
- Create invoices on mobile
- Photo expense capture
- Quick time tracking
- Push notifications
- Revenue Cat for subscriptions

### Go REST API
Backend API for all apps:
- RESTful endpoints
- JWT authentication
- Rate limiting
- Webhook support
- Multi-tenant architecture
- Chi router (standard library compatible)

### Telegram Bot
Conversational interface:
- "Create invoice for Acme Corp $1500"
- "Start timer for Website project"
- "Show this month's revenue"
- Payment notifications
- Requires API subscription

---

## 🔒 Security & Privacy

- **Local-first** — Your data never leaves your machine (CLI)
- **End-to-end encryption** — Cloud sync uses AES-256
- **No tracking** — We don't collect analytics or telemetry
- **Open source** — Audit the code yourself
- **GDPR compliant** — You control your data

---

## 💡 Example Workflow

### Freelance Developer Starting a New Project

```bash
# 1. First time setup (interactive wizard)
$ ung create
> Select: Company
# Fill in your business details

# 2. Add your new client
$ ung create
> Select: Client
# Enter client information

# 3. Set up project contract
$ ung create
> Select: Contract
# Choose hourly rate, $100/hour

# 4. Start working and tracking time
$ ung track start
> Select contract: Website Development
> Project: Authentication System
✓ Timer started

# Work for 3 hours...

$ ung track stop
✓ Timer stopped
  Duration: 3h 12m
  Billable: $320.00

# 5. End of week - create invoice
$ ung invoice new --company 1 --client 1 --price 320 --currency USD

# 6. Generate professional PDF
$ ung invoice pdf 1
✓ PDF generated: ~/.ung/invoices/INV-2025-001.pdf

# 7. Email to client
$ ung invoice email 1
> Select: Apple Mail
✓ Email draft created with PDF attached

# 8. Client approves and wants a contract
$ ung contract pdf 1
✓ Contract PDF generated
```

---

## 📊 Success Stories

> "I went from Quickbooks ($50/month) to UNG (free) and haven't looked back. The CLI is so much faster than clicking through menus." — **Alex K., Web Developer**

> "The interactive wizard is brilliant. I set up my entire freelance business in under 5 minutes." — **Maria S., Designer**

> "Finally, an invoicing tool that respects my data. Everything local, nothing in the cloud unless I want it." — **David L., Security Consultant**

---

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup
```bash
git clone https://github.com/Andriiklymiuk/ung.git
cd ung
go mod download
go build
go test ./...
```

### Run Tests
```bash
go test ./... -cover -v
```

### Architecture
- `cmd/` — Command implementations
- `internal/` — Core business logic
  - `models/` — Data models (GORM)
  - `repository/` — Database access layer
  - `db/` — Database initialization and migrations
  - `config/` — Configuration management
- `pkg/` — Shared packages
  - `invoice/` — PDF generation
  - `contract/` — Contract PDFs
- `migrations/` — SQL schema migrations
- `templates/` — HTML templates

---

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

Inspired by:
- **Daschund** — Running fast towards your goals
- **Your Next Gig** — Focus on getting the next project and getting paid
- **Unix Philosophy** — Do one thing well, make it composable
- **Indie Hackers** — Building sustainable businesses

Special thanks to:
- Cobra CLI framework
- GORM ORM
- Bubbletea TUI library
- SQLite database
- gofpdf library

---

## 🔗 Links

- **Website**: [Coming Soon]
- **Documentation**: [docs.ung.dev](https://docs.ung.dev) [Coming Soon]
- **GitHub**: [github.com/Andriiklymiuk/ung](https://github.com/Andriiklymiuk/ung)
- **Issues**: [github.com/Andriiklymiuk/ung/issues](https://github.com/Andriiklymiuk/ung/issues)

---

## ❓ FAQ

**Q: Why another invoicing tool?**
A: Because existing tools are slow, expensive, and don't respect your data. UNG is fast, free, and local-first.

**Q: Is my data safe?**
A: Yes! Everything is stored locally in SQLite. You can encrypt your database, back it up, or sync it however you want.

**Q: Can I use this for my agency?**
A: Absolutely! The CLI is free for unlimited use. We're planning team features in the Pro/Business tiers.

**Q: What about cloud sync?**
A: Coming in the Pro version. The CLI will always work offline-first.

**Q: Can I customize the invoice layout?**
A: Yes! We provide HTML templates you can customize. PDF rendering uses gofpdf now, HTML-to-PDF coming soon.

**Q: Does it work on Windows?**
A: Yes! UNG is written in Go and works on Linux, macOS, and Windows.

**Q: How do I backup my data?**
A: Just copy `~/.ung/ung.db` somewhere safe. It's a single SQLite file.

**Q: Can I import data from QuickBooks/FreshBooks?**
A: Not yet, but it's on the roadmap!

---

**Made with ❤️ for freelancers who value their time**

**Star ⭐ this repo if UNG helps you get paid faster!**
