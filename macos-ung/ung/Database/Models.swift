//
//  Models.swift
//  ung
//
//  Database models matching the Go CLI models.
//  These use GRDB for SQLite access.
//

import Foundation
import GRDB

// MARK: - Company

struct Company: Codable, FetchableRecord, PersistableRecord, Identifiable {
    var id: Int64?
    var name: String
    var email: String
    var address: String?
    var taxId: String?
    var phone: String?
    var registrationAddress: String?
    var bankName: String?
    var bankAccount: String?
    var bankSwift: String?
    var logoPath: String?
    var createdAt: Date?
    var updatedAt: Date?
    var deletedAt: Date?

    static let databaseTableName = "companies"

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case email
        case address
        case taxId = "tax_id"
        case phone
        case registrationAddress = "registration_address"
        case bankName = "bank_name"
        case bankAccount = "bank_account"
        case bankSwift = "bank_swift"
        case logoPath = "logo_path"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
        case deletedAt = "deleted_at"
    }

    mutating func didInsert(_ inserted: InsertionSuccess) {
        id = inserted.rowID
    }
}

// MARK: - Client

struct ClientModel: Codable, FetchableRecord, PersistableRecord, Identifiable {
    var id: Int64?
    var name: String
    var email: String
    var address: String?
    var taxId: String?
    var createdAt: Date?
    var updatedAt: Date?

    static let databaseTableName = "clients"

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case email
        case address
        case taxId = "tax_id"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }

    mutating func didInsert(_ inserted: InsertionSuccess) {
        id = inserted.rowID
    }
}

// MARK: - Contract

enum ContractType: String, Codable {
    case hourly
    case fixedPrice = "fixed_price"
    case retainer
}

struct ContractModel: Codable, FetchableRecord, PersistableRecord, Identifiable {
    var id: Int64?
    var contractNum: String
    var clientId: Int64
    var name: String
    var contractType: String
    var hourlyRate: Double?
    var fixedPrice: Double?
    var currency: String
    var startDate: Date
    var endDate: Date?
    var active: Bool
    var notes: String?
    var pdfPath: String?
    var createdAt: Date?
    var updatedAt: Date?

    static let databaseTableName = "contracts"

    enum CodingKeys: String, CodingKey {
        case id
        case contractNum = "contract_num"
        case clientId = "client_id"
        case name
        case contractType = "contract_type"
        case hourlyRate = "hourly_rate"
        case fixedPrice = "fixed_price"
        case currency
        case startDate = "start_date"
        case endDate = "end_date"
        case active
        case notes
        case pdfPath = "pdf_path"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }

    mutating func didInsert(_ inserted: InsertionSuccess) {
        id = inserted.rowID
    }

    // Relationship
    static let client = belongsTo(ClientModel.self)
}

// MARK: - Invoice

enum InvoiceStatus: String, Codable {
    case pending
    case sent
    case paid
    case overdue
}

struct Invoice: Codable, FetchableRecord, PersistableRecord, Identifiable {
    var id: Int64?
    var invoiceNum: String
    var companyId: Int64
    var amount: Double
    var currency: String
    var description: String?
    var status: String
    var issuedDate: Date?
    var dueDate: Date?
    var pdfPath: String?
    var notes: String?
    var createdAt: Date?
    var updatedAt: Date?

    static let databaseTableName = "invoices"

    enum CodingKeys: String, CodingKey {
        case id
        case invoiceNum = "invoice_num"
        case companyId = "company_id"
        case amount
        case currency
        case description
        case status
        case issuedDate = "issued_date"
        case dueDate = "due_date"
        case pdfPath = "pdf_path"
        case notes
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }

    mutating func didInsert(_ inserted: InsertionSuccess) {
        id = inserted.rowID
    }

    // Relationships
    static let lineItems = hasMany(InvoiceLineItem.self)
    static let recipients = hasMany(InvoiceRecipient.self)
}

// MARK: - Invoice Line Item

struct InvoiceLineItem: Codable, FetchableRecord, PersistableRecord, Identifiable {
    var id: Int64?
    var invoiceId: Int64
    var itemName: String
    var description: String?
    var quantity: Double
    var rate: Double
    var amount: Double
    var createdAt: Date?

    static let databaseTableName = "invoice_line_items"

    enum CodingKeys: String, CodingKey {
        case id
        case invoiceId = "invoice_id"
        case itemName = "item_name"
        case description
        case quantity
        case rate
        case amount
        case createdAt = "created_at"
    }

    mutating func didInsert(_ inserted: InsertionSuccess) {
        id = inserted.rowID
    }

    static let invoice = belongsTo(Invoice.self)
}

// MARK: - Invoice Recipient

struct InvoiceRecipient: Codable, FetchableRecord, PersistableRecord, Identifiable {
    var id: Int64?
    var invoiceId: Int64
    var clientId: Int64

    static let databaseTableName = "invoice_recipients"

    enum CodingKeys: String, CodingKey {
        case id
        case invoiceId = "invoice_id"
        case clientId = "client_id"
    }

    mutating func didInsert(_ inserted: InsertionSuccess) {
        id = inserted.rowID
    }

    static let invoice = belongsTo(Invoice.self)
    static let client = belongsTo(ClientModel.self)
}

// MARK: - Tracking Session

struct TrackingSession: Codable, FetchableRecord, PersistableRecord, Identifiable {
    var id: Int64?
    var clientId: Int64?
    var contractId: Int64?
    var projectName: String?
    var startTime: Date
    var endTime: Date?
    var duration: Int?  // in seconds
    var hours: Double?
    var billable: Bool
    var notes: String?
    var createdAt: Date?
    var updatedAt: Date?
    var deletedAt: Date?  // soft delete

    static let databaseTableName = "tracking_sessions"

    enum CodingKeys: String, CodingKey {
        case id
        case clientId = "client_id"
        case contractId = "contract_id"
        case projectName = "project_name"
        case startTime = "start_time"
        case endTime = "end_time"
        case duration
        case hours
        case billable
        case notes
        case createdAt = "created_at"
        case updatedAt = "updated_at"
        case deletedAt = "deleted_at"
    }

    mutating func didInsert(_ inserted: InsertionSuccess) {
        id = inserted.rowID
    }

    // Computed properties
    var isActive: Bool {
        endTime == nil && deletedAt == nil
    }

    var calculatedDuration: Int {
        guard let end = endTime else {
            return Int(Date().timeIntervalSince(startTime))
        }
        return Int(end.timeIntervalSince(startTime))
    }

    var formattedDuration: String {
        let totalSeconds = calculatedDuration
        let hours = totalSeconds / 3600
        let minutes = (totalSeconds % 3600) / 60
        let seconds = totalSeconds % 60
        return String(format: "%d:%02d:%02d", hours, minutes, seconds)
    }

    // Relationships
    static let client = belongsTo(ClientModel.self)
    static let contract = belongsTo(ContractModel.self)
}

// MARK: - Expense

enum ExpenseCategory: String, Codable, CaseIterable {
    case software = "Software"
    case hardware = "Hardware"
    case travel = "Travel"
    case meals = "Meals"
    case office = "Office"
    case marketing = "Marketing"
    case education = "Education"
    case other = "Other"
}

struct Expense: Codable, FetchableRecord, PersistableRecord, Identifiable {
    var id: Int64?
    var description: String
    var amount: Double
    var currency: String
    var category: String
    var date: Date
    var vendor: String?
    var receiptPath: String?
    var notes: String?
    var createdAt: Date?
    var updatedAt: Date?

    static let databaseTableName = "expenses"

    enum CodingKeys: String, CodingKey {
        case id
        case description
        case amount
        case currency
        case category
        case date
        case vendor
        case receiptPath = "receipt_path"
        case notes
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }

    mutating func didInsert(_ inserted: InsertionSuccess) {
        id = inserted.rowID
    }
}

// MARK: - Recurring Invoice

enum RecurringFrequency: String, Codable {
    case weekly
    case biweekly
    case monthly
    case quarterly
    case yearly
}

struct RecurringInvoice: Codable, FetchableRecord, PersistableRecord, Identifiable {
    var id: Int64?
    var clientId: Int64
    var contractId: Int64?
    var amount: Double
    var currency: String
    var description: String?
    var frequency: String
    var dayOfMonth: Int
    var dayOfWeek: Int
    var nextGenerationDate: Date
    var lastGeneratedDate: Date?
    var lastInvoiceId: Int64?
    var active: Bool
    var autoSend: Bool
    var autoPdf: Bool
    var emailApp: String?
    var generatedCount: Int
    var notes: String?
    var createdAt: Date?
    var updatedAt: Date?

    static let databaseTableName = "recurring_invoices"

    enum CodingKeys: String, CodingKey {
        case id
        case clientId = "client_id"
        case contractId = "contract_id"
        case amount
        case currency
        case description
        case frequency
        case dayOfMonth = "day_of_month"
        case dayOfWeek = "day_of_week"
        case nextGenerationDate = "next_generation_date"
        case lastGeneratedDate = "last_generated_date"
        case lastInvoiceId = "last_invoice_id"
        case active
        case autoSend = "auto_send"
        case autoPdf = "auto_pdf"
        case emailApp = "email_app"
        case generatedCount = "generated_count"
        case notes
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }

    mutating func didInsert(_ inserted: InsertionSuccess) {
        id = inserted.rowID
    }

    // Relationships
    static let client = belongsTo(ClientModel.self)
    static let contract = belongsTo(ContractModel.self)
}

// MARK: - Schema Migrations Tracking

struct SchemaMigration: Codable, FetchableRecord, PersistableRecord {
    var version: String
    var appliedAt: Date?

    static let databaseTableName = "schema_migrations"

    enum CodingKeys: String, CodingKey {
        case version
        case appliedAt = "applied_at"
    }
}
