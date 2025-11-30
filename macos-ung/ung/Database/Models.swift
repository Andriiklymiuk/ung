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

// MARK: - Recurring Invoice

enum RecurringFrequency: String, Codable, CaseIterable {
    case weekly = "weekly"
    case biweekly = "biweekly"
    case monthly = "monthly"
    case quarterly = "quarterly"
    case yearly = "yearly"

    var displayName: String {
        switch self {
        case .weekly: return "Weekly"
        case .biweekly: return "Bi-weekly"
        case .monthly: return "Monthly"
        case .quarterly: return "Quarterly"
        case .yearly: return "Yearly"
        }
    }
}

struct RecurringInvoice: Codable, FetchableRecord, PersistableRecord, Identifiable {
    var id: Int64?
    var clientId: Int64
    var contractId: Int64?
    var amount: Double
    var currency: String
    var description: String?
    var frequency: String  // RecurringFrequency raw value
    var dayOfMonth: Int
    var dayOfWeek: Int
    var nextGenerationDate: Date
    var lastGeneratedDate: Date?
    var lastInvoiceId: Int64?
    var active: Bool
    var autoPdf: Bool
    var autoSend: Bool
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
        case autoPdf = "auto_pdf"
        case autoSend = "auto_send"
        case emailApp = "email_app"
        case generatedCount = "generated_count"
        case notes
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }

    mutating func didInsert(_ inserted: InsertionSuccess) {
        id = inserted.rowID
    }

    var frequencyType: RecurringFrequency {
        RecurringFrequency(rawValue: frequency) ?? .monthly
    }

    // Calculate the next generation date based on frequency
    func calculateNextGenerationDate(from date: Date = Date()) -> Date {
        let calendar = Calendar.current

        switch frequencyType {
        case .weekly:
            return calendar.date(byAdding: .weekOfYear, value: 1, to: date) ?? date
        case .biweekly:
            return calendar.date(byAdding: .weekOfYear, value: 2, to: date) ?? date
        case .monthly:
            guard var nextDate = calendar.date(byAdding: .month, value: 1, to: date) else { return date }
            var components = calendar.dateComponents([.year, .month], from: nextDate)
            components.day = min(dayOfMonth, 28) // Cap at 28 for safety
            return calendar.date(from: components) ?? nextDate
        case .quarterly:
            guard var nextDate = calendar.date(byAdding: .month, value: 3, to: date) else { return date }
            var components = calendar.dateComponents([.year, .month], from: nextDate)
            components.day = min(dayOfMonth, 28)
            return calendar.date(from: components) ?? nextDate
        case .yearly:
            return calendar.date(byAdding: .year, value: 1, to: date) ?? date
        }
    }

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

// MARK: - Job Hunter Profile

struct HunterProfile: Codable, FetchableRecord, PersistableRecord, Identifiable {
    var id: Int64?
    var name: String
    var title: String?
    var bio: String?
    var skills: String?      // JSON array of skills
    var experience: Int?     // Years of experience
    var rate: Double?        // Desired hourly rate
    var currency: String?
    var location: String?
    var remote: Bool
    var languages: String?   // JSON array of languages
    var education: String?   // JSON array of education
    var projects: String?    // JSON array of notable projects
    var links: String?       // JSON: github, linkedin, portfolio
    var pdfPath: String?     // Original CV path
    var pdfContent: String?  // Extracted text from PDF
    var createdAt: Date?
    var updatedAt: Date?

    static let databaseTableName = "profiles"

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case title
        case bio
        case skills
        case experience
        case rate
        case currency
        case location
        case remote
        case languages
        case education
        case projects
        case links
        case pdfPath = "pdf_path"
        case pdfContent = "pdf_content"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }

    mutating func didInsert(_ inserted: InsertionSuccess) {
        id = inserted.rowID
    }

    // Helper to parse skills from JSON
    var skillsList: [String] {
        guard let skills = skills,
              let data = skills.data(using: .utf8),
              let array = try? JSONDecoder().decode([String].self, from: data) else {
            return []
        }
        return array
    }
}

// MARK: - Job Source

enum JobSource: String, Codable, CaseIterable {
    case hackernews
    case remoteok
    case weworkremotely
    case jobicy
    case arbeitnow
    case djinni       // Ukrainian IT jobs
    case dou          // Ukrainian IT community
    case netherlands  // Dutch job boards
    case eurojobs     // European job boards
    case upwork
    case linkedin
    case manual

    var displayName: String {
        switch self {
        case .hackernews: return "Hacker News"
        case .remoteok: return "RemoteOK"
        case .weworkremotely: return "WeWorkRemotely"
        case .jobicy: return "Jobicy"
        case .arbeitnow: return "Arbeitnow"
        case .djinni: return "Djinni (UA)"
        case .dou: return "DOU (UA)"
        case .netherlands: return "Netherlands"
        case .eurojobs: return "Euro Jobs"
        case .upwork: return "Upwork"
        case .linkedin: return "LinkedIn"
        case .manual: return "Manual"
        }
    }

    var iconName: String {
        switch self {
        case .hackernews: return "y.square.fill"
        case .remoteok: return "globe"
        case .weworkremotely: return "laptopcomputer"
        case .jobicy: return "briefcase.fill"
        case .arbeitnow: return "building.2"
        case .djinni: return "flag.fill"
        case .dou: return "text.bubble.fill"
        case .netherlands: return "windmill.fill"
        case .eurojobs: return "globe.europe.africa.fill"
        case .upwork: return "person.2.fill"
        case .linkedin: return "person.crop.square.fill"
        case .manual: return "square.and.pencil"
        }
    }

    var country: String? {
        switch self {
        case .djinni, .dou: return "Ukraine"
        case .netherlands: return "Netherlands"
        case .eurojobs: return "Europe"
        default: return nil
        }
    }
}

// MARK: - Job

struct HunterJob: Codable, FetchableRecord, PersistableRecord, Identifiable {
    var id: Int64?
    var source: String       // JobSource raw value
    var sourceId: String?
    var sourceUrl: String?
    var title: String
    var company: String?
    var description: String?
    var skills: String?      // JSON array of required skills
    var rateMin: Double?
    var rateMax: Double?
    var rateType: String?    // hourly, monthly, yearly
    var currency: String?
    var remote: Bool
    var location: String?
    var jobType: String?     // contract, fulltime, parttime
    var matchScore: Double?  // 0-100 match with profile
    var postedAt: Date?
    var expiresAt: Date?
    var createdAt: Date?
    var updatedAt: Date?

    static let databaseTableName = "jobs"

    enum CodingKeys: String, CodingKey {
        case id
        case source
        case sourceId = "source_id"
        case sourceUrl = "source_url"
        case title
        case company
        case description
        case skills
        case rateMin = "rate_min"
        case rateMax = "rate_max"
        case rateType = "rate_type"
        case currency
        case remote
        case location
        case jobType = "job_type"
        case matchScore = "match_score"
        case postedAt = "posted_at"
        case expiresAt = "expires_at"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }

    mutating func didInsert(_ inserted: InsertionSuccess) {
        id = inserted.rowID
    }

    var jobSource: JobSource {
        JobSource(rawValue: source) ?? .manual
    }

    var skillsList: [String] {
        guard let skills = skills,
              let data = skills.data(using: .utf8),
              let array = try? JSONDecoder().decode([String].self, from: data) else {
            return []
        }
        return array
    }

    var formattedRate: String? {
        guard let min = rateMin, min > 0 else { return nil }
        let curr = currency ?? "USD"
        let type = rateType ?? "yearly"

        if let max = rateMax, max > 0 && max != min {
            return "\(curr) \(Int(min))-\(Int(max))/\(type)"
        }
        return "\(curr) \(Int(min))/\(type)"
    }
}

// MARK: - Application Status

enum ApplicationStatus: String, Codable, CaseIterable {
    case draft
    case applied
    case viewed
    case response
    case interview
    case offer
    case rejected
    case withdrawn

    var displayName: String {
        switch self {
        case .draft: return "Draft"
        case .applied: return "Applied"
        case .viewed: return "Viewed"
        case .response: return "Response"
        case .interview: return "Interview"
        case .offer: return "Offer"
        case .rejected: return "Rejected"
        case .withdrawn: return "Withdrawn"
        }
    }

    var color: String {
        switch self {
        case .draft: return "gray"
        case .applied: return "blue"
        case .viewed: return "purple"
        case .response: return "cyan"
        case .interview: return "orange"
        case .offer: return "green"
        case .rejected: return "red"
        case .withdrawn: return "gray"
        }
    }

    var iconName: String {
        switch self {
        case .draft: return "doc"
        case .applied: return "paperplane.fill"
        case .viewed: return "eye.fill"
        case .response: return "envelope.fill"
        case .interview: return "person.2.fill"
        case .offer: return "checkmark.seal.fill"
        case .rejected: return "xmark.circle.fill"
        case .withdrawn: return "arrow.uturn.backward"
        }
    }
}

// MARK: - Application

struct HunterApplication: Codable, FetchableRecord, PersistableRecord, Identifiable {
    var id: Int64?
    var jobId: Int64
    var profileId: Int64?
    var proposal: String?
    var proposalPdf: String?
    var coverLetter: String?
    var status: String       // ApplicationStatus raw value
    var notes: String?
    var appliedAt: Date?
    var responseAt: Date?
    var createdAt: Date?
    var updatedAt: Date?

    static let databaseTableName = "applications"

    enum CodingKeys: String, CodingKey {
        case id
        case jobId = "job_id"
        case profileId = "profile_id"
        case proposal
        case proposalPdf = "proposal_pdf"
        case coverLetter = "cover_letter"
        case status
        case notes
        case appliedAt = "applied_at"
        case responseAt = "response_at"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }

    mutating func didInsert(_ inserted: InsertionSuccess) {
        id = inserted.rowID
    }

    var applicationStatus: ApplicationStatus {
        ApplicationStatus(rawValue: status) ?? .draft
    }

    // Relationships
    static let job = belongsTo(HunterJob.self)
    static let profile = belongsTo(HunterProfile.self)
}

// MARK: - Hunter Statistics

struct HunterStats: Codable {
    var totalJobs: Int
    var totalApplications: Int
    var statusCounts: [String: Int]
    var topSkills: [String]
    var averageMatchScore: Double
    var recentJobs: [HunterJob]

    init(totalJobs: Int = 0, totalApplications: Int = 0, statusCounts: [String: Int] = [:], topSkills: [String] = [], averageMatchScore: Double = 0, recentJobs: [HunterJob] = []) {
        self.totalJobs = totalJobs
        self.totalApplications = totalApplications
        self.statusCounts = statusCounts
        self.topSkills = topSkills
        self.averageMatchScore = averageMatchScore
        self.recentJobs = recentJobs
    }
}

// MARK: - Gig Status

// Simplified flow: todo → in_progress → sent → done
enum GigStatus: String, Codable, CaseIterable {
    case todo
    case inProgress = "in_progress"
    case sent
    case done
    case onHold = "on_hold"
    case cancelled

    var displayName: String {
        switch self {
        case .todo: return "Todo"
        case .inProgress: return "In Progress"
        case .sent: return "Sent"
        case .done: return "Done"
        case .onHold: return "On Hold"
        case .cancelled: return "Cancelled"
        }
    }

    var color: String {
        switch self {
        case .todo: return "gray"
        case .inProgress: return "blue"
        case .sent: return "orange"
        case .done: return "green"
        case .onHold: return "yellow"
        case .cancelled: return "red"
        }
    }

    // Main workflow statuses (excludes on_hold and cancelled)
    static var workflowStatuses: [GigStatus] {
        [.todo, .inProgress, .sent, .done]
    }
}

// MARK: - Gig

struct Gig: Codable, FetchableRecord, PersistableRecord, Identifiable {
    var id: Int64?
    var name: String
    var clientId: Int64?
    var contractId: Int64?
    var applicationId: Int64?
    var status: String
    var gigType: String?
    var priority: Int
    var estimatedHours: Double?
    var estimatedAmount: Double?
    var hourlyRate: Double?
    var currency: String?
    var totalHoursTracked: Double
    var lastTrackedAt: Date?
    var totalInvoiced: Double
    var lastInvoicedAt: Date?
    var startDate: Date?
    var dueDate: Date?
    var completedAt: Date?
    var description: String?
    var notes: String?
    var project: String?
    var createdAt: Date?
    var updatedAt: Date?

    static let databaseTableName = "gigs"

    enum CodingKeys: String, CodingKey {
        case id, name, status, priority, currency, description, notes, project
        case clientId = "client_id"
        case contractId = "contract_id"
        case applicationId = "application_id"
        case gigType = "gig_type"
        case estimatedHours = "estimated_hours"
        case estimatedAmount = "estimated_amount"
        case hourlyRate = "hourly_rate"
        case totalHoursTracked = "total_hours_tracked"
        case lastTrackedAt = "last_tracked_at"
        case totalInvoiced = "total_invoiced"
        case lastInvoicedAt = "last_invoiced_at"
        case startDate = "start_date"
        case dueDate = "due_date"
        case completedAt = "completed_at"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }

    mutating func didInsert(_ inserted: InsertionSuccess) {
        id = inserted.rowID
    }

    var gigStatus: GigStatus {
        GigStatus(rawValue: status) ?? .todo
    }

    static let client = belongsTo(ClientModel.self)
}

// MARK: - Work Log

struct WorkLog: Codable, FetchableRecord, PersistableRecord, Identifiable {
    var id: Int64?
    var gigId: Int64?
    var clientId: Int64?
    var trackingSessionId: Int64?
    var content: String
    var logType: String
    var createdAt: Date?
    var updatedAt: Date?

    static let databaseTableName = "work_logs"

    enum CodingKeys: String, CodingKey {
        case id, content
        case gigId = "gig_id"
        case clientId = "client_id"
        case trackingSessionId = "tracking_session_id"
        case logType = "log_type"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }

    mutating func didInsert(_ inserted: InsertionSuccess) {
        id = inserted.rowID
    }
}

// MARK: - Gig Task

struct GigTask: Codable, FetchableRecord, PersistableRecord, Identifiable {
    var id: Int64?
    var gigId: Int64
    var title: String
    var description: String?
    var completed: Bool
    var completedAt: Date?
    var dueDate: Date?
    var sortOrder: Int

    static let databaseTableName = "gig_tasks"

    enum CodingKeys: String, CodingKey {
        case id, title, description, completed
        case gigId = "gig_id"
        case completedAt = "completed_at"
        case dueDate = "due_date"
        case sortOrder = "sort_order"
    }

    init(id: Int64? = nil, gigId: Int64, title: String, description: String? = nil, completed: Bool = false, completedAt: Date? = nil, dueDate: Date? = nil, sortOrder: Int = 0) {
        self.id = id
        self.gigId = gigId
        self.title = title
        self.description = description
        self.completed = completed
        self.completedAt = completedAt
        self.dueDate = dueDate
        self.sortOrder = sortOrder
    }

    mutating func didInsert(_ inserted: InsertionSuccess) {
        id = inserted.rowID
    }

    static let gig = belongsTo(Gig.self)
}
