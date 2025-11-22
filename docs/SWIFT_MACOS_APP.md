# UNG Swift macOS App

> **Native macOS application for UNG using the same SQLite database as the CLI**

## Overview

The Swift macOS app provides a beautiful native interface for UNG, targeting macOS users who prefer GUI over CLI. It uses the **exact same SQLite database** and **same migration files** as the CLI, ensuring perfect compatibility.

## Core Architecture

### Database Layer
- **Shared SQLite Database**: `~/.ung/ung.db`
- **Same Migrations**: Use migrations from `../migrations/` directory
- **SQLite.swift**: Recommended library for type-safe database access
- **Migration Runner**: Implement migration tracking matching CLI's schema_migrations table

### Technology Stack
- **Language**: Swift 5.9+
- **Framework**: SwiftUI for UI
- **Database**: SQLite.swift
- **Minimum OS**: macOS 13.0 (Ventura)
- **Architecture**: MVVM (Model-View-ViewModel)

## Key Features

### 1. Dashboard View
- **Revenue Overview**: Monthly/yearly revenue charts
- **Active Contracts**: Quick access to ongoing projects
- **Recent Invoices**: Last 5-10 invoices with status
- **Quick Actions**: "Create Invoice", "Start Timer", "Add Client"

### 2. Company Management
- Single company profile screen
- Edit business details
- Upload company logo (store path in database)
- Bank details management

### 3. Client Management
- List view with search
- Detail view with contract history
- Add/Edit client forms
- Quick actions (Create Invoice, New Contract)

### 4. Contract Management
- Active/Inactive contract lists
- Contract detail view with:
  - Client information
  - Contract terms
  - Time tracking summary
  - Generate PDF button
  - Email contract button
- Contract creation wizard

### 5. Invoice Management
- Invoice list with filters (status, date range, client)
- Invoice detail view
- PDF generation (native macOS rendering or reuse gofpdf logic)
- Email invoice directly from app
- Batch operations (export monthly invoices)

### 6. Time Tracking
- **Menu Bar Timer**: System tray integration
  - Start/stop timer from menu bar
  - Show current duration
  - Quick project selection
- **Timer View**: Full timer interface in app
- **Session History**: List all tracked sessions
- **Manual Entry**: Log hours worked

### 7. Settings
- Database location
- Invoice directory
- Language preferences
- Template customization
- Email configuration

## Database Schema Implementation

### Migration System
```swift
// MigrationRunner.swift
class MigrationRunner {
    static func runMigrations(db: Connection) throws {
        // Read migration files from ../migrations/
        // Track applied migrations in schema_migrations table
        // Apply migrations in order
    }
}
```

### Models (Matching CLI)
```swift
// Company.swift
struct Company: Codable {
    let id: Int64
    var name: String
    var email: String
    var phone: String?
    var address: String?
    var registrationAddress: String?
    var taxID: String?
    var bankName: String?
    var bankAccount: String?
    var bankSWIFT: String?
    var logoPath: String?
    let createdAt: Date
    var updatedAt: Date
}

// Client.swift
struct Client: Codable {
    let id: Int64
    var name: String
    var email: String
    var address: String?
    var taxID: String?
    let createdAt: Date
    var updatedAt: Date
}

// Contract.swift
struct Contract: Codable {
    let id: Int64
    let clientID: Int64
    var name: String
    var contractType: ContractType
    var hourlyRate: Double?
    var fixedPrice: Double?
    var currency: String
    var startDate: Date
    var endDate: Date?
    var active: Bool
    var notes: String?
    var pdfPath: String?
    let createdAt: Date
    var updatedAt: Date
}

enum ContractType: String, Codable {
    case hourly
    case fixedPrice = "fixed_price"
    case retainer
}

// Invoice.swift
struct Invoice: Codable {
    let id: Int64
    var invoiceNum: String
    let companyID: Int64
    var amount: Double
    var currency: String
    var description: String
    var status: InvoiceStatus
    var issuedDate: Date
    var dueDate: Date
    var pdfPath: String?
    let createdAt: Date
    var updatedAt: Date
}

enum InvoiceStatus: String, Codable {
    case pending, sent, paid, overdue
}

// InvoiceLineItem.swift
struct InvoiceLineItem: Codable {
    let id: Int64
    let invoiceID: Int64
    var itemName: String
    var description: String?
    var quantity: Double
    var rate: Double
    var amount: Double
    let createdAt: Date
}

// TrackingSession.swift
struct TrackingSession: Codable {
    let id: Int64
    let contractID: Int64?
    var projectName: String
    var startTime: Date
    var endTime: Date?
    var notes: String?
    var billableAmount: Double?
    let createdAt: Date
}
```

### Repository Pattern
```swift
// CompanyRepository.swift
class CompanyRepository {
    let db: Connection

    func getFirst() throws -> Company?
    func create(_ company: Company) throws -> Company
    func update(_ company: Company) throws
}

// ClientRepository.swift
class ClientRepository {
    func list() throws -> [Client]
    func get(id: Int64) throws -> Client?
    func create(_ client: Client) throws -> Client
    func update(_ client: Client) throws
    func delete(id: Int64) throws
}

// Similar repositories for Contract, Invoice, TrackingSession
```

## UI/UX Design

### Window Structure
```
┌─────────────────────────────────────────┐
│ UNG                            ⚙️  👤    │ <- Title Bar
├─────────────────────────────────────────┤
│ [Sidebar]  │  [Main Content Area]      │
│            │                            │
│ Dashboard  │  Dashboard with charts    │
│ Clients    │  and quick actions        │
│ Contracts  │                            │
│ Invoices   │                            │
│ Time Track │                            │
│ Settings   │                            │
│            │                            │
│            │                            │
└─────────────────────────────────────────┘
```

### Menu Bar Integration
```
┌─────────────────┐
│ ⏱️ UNG Timer    │
├─────────────────┤
│ 01:23:45        │ <- Current timer
│ Website Dev     │ <- Project name
├─────────────────┤
│ ▶️ Start Timer  │
│ ⏹️ Stop Timer   │
│ 📊 Open App     │
│ ⚙️ Settings     │
│ ❌ Quit         │
└─────────────────┘
```

## PDF Generation

### Option 1: Use gofpdf via CLI
```swift
// Call CLI binary to generate PDF
let process = Process()
process.executableURL = URL(fileURLWithPath: "/usr/local/bin/ung")
process.arguments = ["invoice", "pdf", "\(invoiceID)"]
try process.run()
process.waitUntilExit()
```

### Option 2: Native PDF Generation
```swift
// Use PDFKit and HTML templates
import PDFKit

class PDFGenerator {
    static func generateInvoice(invoice: Invoice, company: Company, client: Client) -> URL? {
        // Render HTML template
        // Convert to PDF using WKWebView or PDFKit
        // Save to ~/.ung/invoices/
    }
}
```

### Option 3: Hybrid Approach
- Use gofpdf logic ported to Swift
- Or embed Go code as Swift package

## Email Integration

### macOS Mail Integration
```swift
import AppKit

class EmailHelper {
    static func composeEmail(to: String, subject: String, body: String, attachments: [URL]) {
        let service = NSSharingService(named: .composeEmail)
        service?.recipients = [to]
        service?.subject = subject

        // Attach PDFs
        service?.perform(withItems: attachments)
    }
}
```

### Direct SMTP (Future)
```swift
// Use SwiftSMTP library
import SwiftSMTP

class SMTPService {
    func sendInvoice(invoice: Invoice, to: String, attachment: URL) async throws {
        // Send via user's configured SMTP
    }
}
```

## Notifications

### Local Notifications
```swift
import UserNotifications

class NotificationManager {
    // Timer reminders
    func scheduleTimerReminder(after duration: TimeInterval)

    // Invoice due date reminders
    func scheduleInvoiceReminder(invoice: Invoice)

    // Payment received notifications
    func notifyPaymentReceived(invoice: Invoice)
}
```

## Syncing with CLI

### Database Observation
```swift
import Combine

class DatabaseObserver {
    // Watch for file changes to ~/.ung/ung.db
    // Reload data when CLI makes changes

    func observeChanges() {
        // Use DispatchSource.makeFileSystemObjectSource
        // Or FileSystemEvent API
    }
}
```

## Distribution

### App Store
- **Pros**: Discovery, auto-updates, trust
- **Cons**: Review process, 30% fee (if paid)
- **Strategy**: Free version on App Store, Pro features via in-app purchase

### Direct Download
- **Pros**: No fees, faster updates
- **Cons**: Code signing required, less discovery
- **Strategy**: Notarized DMG on website

### Homebrew Cask
```ruby
cask "ung-app" do
  version "1.0.0"
  sha256 "..."

  url "https://github.com/Andriiklymiuk/ung/releases/download/v#{version}/UNG.dmg"
  name "UNG"
  desc "Your Next Gig billing and tracking app"
  homepage "https://ung.dev"

  app "UNG.app"
end
```

## Testing Strategy

### Unit Tests
```swift
import XCTest

class CompanyRepositoryTests: XCTestCase {
    var testDB: Connection!
    var repo: CompanyRepository!

    override func setUp() {
        // Create in-memory database
        testDB = try! Connection(.inMemory)
        MigrationRunner.runMigrations(db: testDB)
        repo = CompanyRepository(db: testDB)
    }

    func testCreateCompany() {
        // Test company creation
    }
}
```

### UI Tests
```swift
import XCTest

class UNGUITests: XCTestCase {
    func testCreateInvoiceFlow() {
        let app = XCUIApplication()
        app.launch()

        // Test invoice creation workflow
    }
}
```

## Performance Optimization

### Database Indexing
- Same indexes as CLI migrations
- Query optimization for large datasets

### Lazy Loading
```swift
// Load invoices on-demand
@State private var invoices: [Invoice] = []

func loadInvoices() async {
    invoices = try await repository.listInvoices(limit: 50, offset: offset)
}
```

### Caching
- Cache dashboard data
- Invalidate on database changes

## Security

### Keychain Integration
```swift
import Security

class KeychainService {
    // Store email credentials
    func saveEmailPassword(email: String, password: String)

    // Store API tokens (future)
    func saveAPIToken(_ token: String)
}
```

### Database Encryption (Future)
```swift
// Use SQLCipher for encrypted database
import SQLCipher

let db = try Connection("~/.ung/ung.db", key: userPassword)
```

## Key Implementation Steps

### Phase 1: Foundation (Week 1-2)
1. Set up Xcode project with SwiftUI
2. Implement migration runner
3. Create all model structs
4. Build repository layer
5. Set up database connection

### Phase 2: Core Features (Week 3-4)
1. Dashboard view
2. Client list and detail views
3. Invoice list and detail views
4. Contract management
5. Basic PDF generation (via CLI)

### Phase 3: Advanced Features (Week 5-6)
1. Time tracking with menu bar
2. Email integration
3. Settings screen
4. Search and filters
5. Keyboard shortcuts

### Phase 4: Polish (Week 7-8)
1. Animations and transitions
2. Error handling and validation
3. Onboarding flow
4. Documentation
5. App icon and branding

### Phase 5: Distribution (Week 9-10)
1. Code signing
2. Notarization
3. DMG creation
4. App Store submission (optional)
5. Website and marketing

## Future Enhancements

### Widgets (macOS 14+)
- Revenue widget
- Active timer widget
- Upcoming invoices widget

### Shortcuts Integration
```swift
import AppIntents

struct CreateInvoiceIntent: AppIntent {
    static var title: LocalizedStringResource = "Create Invoice"

    func perform() async throws -> some IntentResult {
        // Create invoice via Shortcuts
    }
}
```

### iCloud Sync (Paid Feature)
- Sync database via CloudKit
- Conflict resolution
- Backup and restore

## Monetization

### Free Tier
- All core features
- Unlimited invoices and clients
- PDF generation
- Time tracking

### Pro Tier ($9/month or $90/year)
- Cloud sync via iCloud
- Recurring invoices
- Advanced reports
- Email delivery
- Priority support

### Activation
```swift
// Use RevenueCat or StoreKit 2
import StoreKit

class SubscriptionManager {
    func checkSubscription() async -> Bool
    func purchase(product: Product) async throws
    func restore() async throws
}
```

## Resources

### Libraries
- **SQLite.swift**: https://github.com/stephencelis/SQLite.swift
- **SwiftUI Charts**: Built-in (macOS 13+)
- **PDFKit**: Built-in
- **RevenueCat**: https://www.revenuecat.com

### Design
- **SF Symbols**: System icons
- **Human Interface Guidelines**: Apple's design guidelines
- **AppKit**: For advanced menu bar integration

### Community
- **Swift Forums**: https://forums.swift.org
- **r/swift**: Reddit community
- **Stack Overflow**: swift + swiftui tags

## Migration Checklist

- [ ] Same SQLite schema as CLI
- [ ] Same migration files (000001, 000002, 000003, 000004)
- [ ] Migration tracking table (schema_migrations)
- [ ] All models match CLI models exactly
- [ ] Repository pattern for type safety
- [ ] PDF generation compatible with CLI
- [ ] Email templates reusable
- [ ] Config file compatibility (.ung.yaml)

## Notes

- **CLI Compatibility**: App must work seamlessly alongside CLI
- **Data Integrity**: Never break database compatibility
- **Migration Order**: Apply migrations in exact order as CLI
- **Testing**: Test with CLI-created data extensively
- **Performance**: Handle large datasets (1000+ invoices)
- **Localization**: Support same languages as CLI (English, Ukrainian, German, etc.)
