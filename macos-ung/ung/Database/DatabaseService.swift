//
//  DatabaseService.swift
//  ung
//
//  Native Swift database service using GRDB.
//  Reads the same SQL migration files as the Go CLI for consistency.
//  Supports iCloud sync for cross-device data sharing.
//

import Foundation
import GRDB

// MARK: - Database Error

enum DatabaseError: LocalizedError {
    case notInitialized
    case migrationFailed(String)
    case queryFailed(String)
    case notFound
    case invalidData(String)
    case iCloudNotAvailable

    var errorDescription: String? {
        switch self {
        case .notInitialized:
            return "Database not initialized"
        case .migrationFailed(let message):
            return "Migration failed: \(message)"
        case .queryFailed(let message):
            return "Query failed: \(message)"
        case .notFound:
            return "Record not found"
        case .invalidData(let message):
            return "Invalid data: \(message)"
        case .iCloudNotAvailable:
            return "iCloud is not available on this device"
        }
    }
}

// MARK: - Sync Status

enum SyncStatus: Equatable {
    case idle
    case syncing
    case completed
    case error(String)
}

// MARK: - Database Service

actor DatabaseService {
    private var dbPool: DatabasePool?
    private let fileManager = FileManager.default

    // iCloud sync state
    private(set) var iCloudEnabled: Bool = false
    private(set) var syncStatus: SyncStatus = .idle
    private var metadataQuery: NSMetadataQuery?

    // Singleton for app-wide access
    static let shared = DatabaseService()

    private init() {
        // Load iCloud preference
        iCloudEnabled = UserDefaults.standard.bool(forKey: "iCloudSyncEnabled")
    }

    // MARK: - iCloud Availability

    /// Check if iCloud is available on this device
    var isICloudAvailable: Bool {
        fileManager.ubiquityIdentityToken != nil
    }

    /// Get the iCloud container URL if available
    var iCloudContainerURL: URL? {
        fileManager.url(forUbiquityContainerIdentifier: nil)
    }

    /// Get the iCloud Documents directory
    var iCloudDocumentsURL: URL? {
        iCloudContainerURL?.appendingPathComponent("Documents")
    }

    // MARK: - Database Path

    /// Returns the path to the database file
    /// - On macOS with iCloud: ~/Library/Mobile Documents/iCloud~com~ung~ung/Documents/ung.db
    /// - On macOS without iCloud: ~/.ung/ung.db
    /// - On iOS: Always uses iCloud if available, otherwise local Documents
    var databasePath: String {
        #if os(iOS)
        // iOS: Use iCloud Documents if available, otherwise local Documents
        if let iCloudURL = iCloudDocumentsURL {
            try? fileManager.createDirectory(at: iCloudURL, withIntermediateDirectories: true)
            return iCloudURL.appendingPathComponent("ung.db").path
        }
        let documentsURL = fileManager.urls(for: .documentDirectory, in: .userDomainMask).first!
        return documentsURL.appendingPathComponent("ung.db").path
        #else
        // macOS: Use iCloud if enabled, otherwise ~/.ung/
        if iCloudEnabled, let iCloudURL = iCloudDocumentsURL {
            try? fileManager.createDirectory(at: iCloudURL, withIntermediateDirectories: true)
            return iCloudURL.appendingPathComponent("ung.db").path
        }
        let homeDir = fileManager.homeDirectoryForCurrentUser
        let ungDir = homeDir.appendingPathComponent(".ung")
        try? fileManager.createDirectory(at: ungDir, withIntermediateDirectories: true)
        return ungDir.appendingPathComponent("ung.db").path
        #endif
    }

    /// Returns the local database path (non-iCloud)
    var localDatabasePath: String {
        #if os(iOS)
        let documentsURL = fileManager.urls(for: .documentDirectory, in: .userDomainMask).first!
        return documentsURL.appendingPathComponent("ung.db").path
        #else
        let homeDir = fileManager.homeDirectoryForCurrentUser
        let ungDir = homeDir.appendingPathComponent(".ung")
        try? fileManager.createDirectory(at: ungDir, withIntermediateDirectories: true)
        return ungDir.appendingPathComponent("ung.db").path
        #endif
    }

    /// Returns the iCloud database path if available
    var iCloudDatabasePath: String? {
        guard let iCloudURL = iCloudDocumentsURL else { return nil }
        try? fileManager.createDirectory(at: iCloudURL, withIntermediateDirectories: true)
        return iCloudURL.appendingPathComponent("ung.db").path
    }

    /// Returns the directory for storing invoices
    var invoicesDirectory: URL {
        #if os(iOS)
        let documentsURL = fileManager.urls(for: .documentDirectory, in: .userDomainMask).first!
        let invoicesURL = documentsURL.appendingPathComponent("invoices")
        try? fileManager.createDirectory(at: invoicesURL, withIntermediateDirectories: true)
        return invoicesURL
        #else
        let homeDir = fileManager.homeDirectoryForCurrentUser
        let invoicesURL = homeDir.appendingPathComponent(".ung/invoices")
        try? fileManager.createDirectory(at: invoicesURL, withIntermediateDirectories: true)
        return invoicesURL
        #endif
    }

    // MARK: - Initialization

    /// Check if the database exists and is initialized
    var isInitialized: Bool {
        fileManager.fileExists(atPath: databasePath)
    }

    /// Initialize the database and run migrations
    func initialize() async throws {
        // Create the .ung directory if needed
        let dbURL = URL(fileURLWithPath: databasePath)
        let dbDir = dbURL.deletingLastPathComponent()
        try? fileManager.createDirectory(at: dbDir, withIntermediateDirectories: true)

        // Open the database
        var config = Configuration()
        config.prepareDatabase { db in
            // Enable foreign keys
            try db.execute(sql: "PRAGMA foreign_keys = ON")
        }

        dbPool = try DatabasePool(path: databasePath, configuration: config)

        // Run migrations
        try await runMigrations()
    }

    /// Close the database connection
    func close() {
        dbPool = nil
    }

    // MARK: - iCloud Sync

    /// Enable or disable iCloud sync
    /// When enabling, copies local database to iCloud
    /// When disabling, copies iCloud database to local
    func setICloudEnabled(_ enabled: Bool) async throws {
        guard enabled != iCloudEnabled else { return }

        if enabled {
            guard isICloudAvailable else {
                throw DatabaseError.iCloudNotAvailable
            }
            // Copy local database to iCloud
            try await migrateToICloud()
        } else {
            // Copy iCloud database to local
            try await migrateFromICloud()
        }

        iCloudEnabled = enabled
        UserDefaults.standard.set(enabled, forKey: "iCloudSyncEnabled")

        // Reinitialize with new path
        close()
        try await initialize()
    }

    /// Migrate database from local to iCloud
    private func migrateToICloud() async throws {
        guard let iCloudPath = iCloudDatabasePath else {
            throw DatabaseError.iCloudNotAvailable
        }

        syncStatus = .syncing

        let localPath = localDatabasePath
        let iCloudURL = URL(fileURLWithPath: iCloudPath)

        // Close current connection
        close()

        // Check if local database exists
        if fileManager.fileExists(atPath: localPath) {
            // Check if iCloud already has a database
            if fileManager.fileExists(atPath: iCloudPath) {
                // Use the more recent one
                let localAttrs = try? fileManager.attributesOfItem(atPath: localPath)
                let iCloudAttrs = try? fileManager.attributesOfItem(atPath: iCloudPath)

                let localDate = localAttrs?[.modificationDate] as? Date ?? .distantPast
                let iCloudDate = iCloudAttrs?[.modificationDate] as? Date ?? .distantPast

                if localDate > iCloudDate {
                    // Local is newer, overwrite iCloud
                    try fileManager.removeItem(atPath: iCloudPath)
                    try fileManager.copyItem(atPath: localPath, toPath: iCloudPath)
                }
                // Otherwise keep iCloud version
            } else {
                // Copy local to iCloud
                try fileManager.copyItem(atPath: localPath, toPath: iCloudPath)
            }
        }

        syncStatus = .completed
    }

    /// Migrate database from iCloud to local
    private func migrateFromICloud() async throws {
        guard let iCloudPath = iCloudDatabasePath else { return }

        syncStatus = .syncing

        let localPath = localDatabasePath

        // Close current connection
        close()

        // Check if iCloud database exists
        if fileManager.fileExists(atPath: iCloudPath) {
            // Backup local if it exists
            if fileManager.fileExists(atPath: localPath) {
                let backupPath = localPath + ".backup"
                try? fileManager.removeItem(atPath: backupPath)
                try? fileManager.copyItem(atPath: localPath, toPath: backupPath)
                try fileManager.removeItem(atPath: localPath)
            }
            // Copy from iCloud
            try fileManager.copyItem(atPath: iCloudPath, toPath: localPath)
        }

        syncStatus = .completed
    }

    /// Trigger iCloud sync check - call when app comes to foreground
    func triggerSync() async -> SyncStatus {
        guard iCloudEnabled, isICloudAvailable else {
            return .idle
        }

        syncStatus = .syncing

        // Force download if file exists in iCloud but not downloaded
        if let iCloudPath = iCloudDatabasePath {
            let url = URL(fileURLWithPath: iCloudPath)
            do {
                try fileManager.startDownloadingUbiquitousItem(at: url)
                // Wait a moment for sync to start
                try? await Task.sleep(nanoseconds: 500_000_000)
                syncStatus = .completed
            } catch {
                syncStatus = .error(error.localizedDescription)
            }
        } else {
            syncStatus = .idle
        }

        return syncStatus
    }

    // MARK: - Migrations

    /// Run SQL migrations from bundle or inline fallback
    private func runMigrations() async throws {
        guard let db = dbPool else {
            throw DatabaseError.notInitialized
        }

        try await db.write { db in
            // Create migrations tracking table
            try db.execute(sql: """
                CREATE TABLE IF NOT EXISTS schema_migrations (
                    version TEXT PRIMARY KEY,
                    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                )
                """)

            // Try to load migrations from bundle first
            if let migrationsURL = Bundle.main.url(forResource: "migrations", withExtension: nil) {
                try self.runMigrationsFromDirectory(db: db, directory: migrationsURL)
            } else {
                // Fallback to inline schema (same as Go CLI)
                try self.runInlineSchema(db: db)
            }
        }
    }

    /// Run migrations from a directory of SQL files
    private func runMigrationsFromDirectory(db: Database, directory: URL) throws {
        let files = try fileManager.contentsOfDirectory(at: directory, includingPropertiesForKeys: nil)
            .filter { $0.pathExtension == "sql" && $0.lastPathComponent.contains(".up.") }
            .sorted { $0.lastPathComponent < $1.lastPathComponent }

        for file in files {
            let fileName = file.lastPathComponent
            let version = String(fileName.dropLast(7))  // Remove ".up.sql"

            // Check if already applied
            let count = try Int.fetchOne(
                db,
                sql: "SELECT COUNT(*) FROM schema_migrations WHERE version = ?",
                arguments: [version]
            ) ?? 0

            if count > 0 {
                continue  // Already applied
            }

            // Read and execute migration
            let migrationSQL = try String(contentsOf: file, encoding: .utf8)

            // Split by semicolons and execute each statement
            let statements = migrationSQL.components(separatedBy: ";")
                .map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }
                .filter { !$0.isEmpty && !$0.hasPrefix("--") }

            for statement in statements {
                do {
                    try db.execute(sql: statement)
                } catch {
                    // Some statements may fail if column already exists, etc.
                    // This is expected behavior for idempotent migrations
                    print("[DatabaseService] Migration statement warning: \(error)")
                }
            }

            // Record migration
            try db.execute(
                sql: "INSERT INTO schema_migrations (version) VALUES (?)",
                arguments: [version]
            )

            print("[DatabaseService] Applied migration: \(version)")
        }
    }

    /// Inline schema fallback (matches Go CLI's runInlineSchema)
    private func runInlineSchema(db: Database) throws {
        let schema = """
            CREATE TABLE IF NOT EXISTS companies (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL,
                email TEXT NOT NULL,
                address TEXT,
                tax_id TEXT,
                phone TEXT,
                registration_address TEXT,
                bank_name TEXT,
                bank_account TEXT,
                bank_swift TEXT,
                logo_path TEXT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                deleted_at TIMESTAMP
            );

            CREATE TABLE IF NOT EXISTS clients (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL,
                email TEXT NOT NULL,
                address TEXT,
                tax_id TEXT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            );

            CREATE TABLE IF NOT EXISTS contracts (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                contract_num TEXT UNIQUE NOT NULL DEFAULT '',
                client_id INTEGER NOT NULL,
                name TEXT NOT NULL,
                contract_type TEXT NOT NULL,
                hourly_rate REAL,
                fixed_price REAL,
                currency TEXT DEFAULT 'USD',
                start_date TIMESTAMP NOT NULL,
                end_date TIMESTAMP,
                active BOOLEAN DEFAULT 1,
                notes TEXT,
                pdf_path TEXT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (client_id) REFERENCES clients(id)
            );

            CREATE TABLE IF NOT EXISTS invoices (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                invoice_num TEXT UNIQUE NOT NULL,
                company_id INTEGER NOT NULL,
                amount REAL NOT NULL,
                currency TEXT DEFAULT 'USD',
                description TEXT,
                status TEXT DEFAULT 'pending',
                issued_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                due_date TIMESTAMP,
                pdf_path TEXT,
                notes TEXT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (company_id) REFERENCES companies(id)
            );

            CREATE TABLE IF NOT EXISTS invoice_line_items (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                invoice_id INTEGER NOT NULL,
                item_name TEXT NOT NULL,
                description TEXT,
                quantity REAL NOT NULL DEFAULT 1,
                rate REAL NOT NULL,
                amount REAL NOT NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (invoice_id) REFERENCES invoices(id)
            );

            CREATE TABLE IF NOT EXISTS invoice_recipients (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                invoice_id INTEGER NOT NULL,
                client_id INTEGER NOT NULL,
                FOREIGN KEY (invoice_id) REFERENCES invoices(id),
                FOREIGN KEY (client_id) REFERENCES clients(id)
            );

            CREATE TABLE IF NOT EXISTS tracking_sessions (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                client_id INTEGER,
                contract_id INTEGER,
                project_name TEXT,
                start_time TIMESTAMP NOT NULL,
                end_time TIMESTAMP,
                duration INTEGER,
                hours REAL,
                billable BOOLEAN DEFAULT 1,
                notes TEXT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                deleted_at TIMESTAMP,
                FOREIGN KEY (client_id) REFERENCES clients(id),
                FOREIGN KEY (contract_id) REFERENCES contracts(id)
            );

            CREATE TABLE IF NOT EXISTS expenses (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                description TEXT NOT NULL,
                amount REAL NOT NULL,
                currency TEXT DEFAULT 'USD',
                category TEXT NOT NULL,
                date TIMESTAMP NOT NULL,
                vendor TEXT,
                receipt_path TEXT,
                notes TEXT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            );

            CREATE TABLE IF NOT EXISTS recurring_invoices (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                client_id INTEGER NOT NULL,
                contract_id INTEGER,
                amount REAL NOT NULL,
                currency TEXT DEFAULT 'USD',
                description TEXT,
                frequency TEXT NOT NULL,
                day_of_month INTEGER DEFAULT 1,
                day_of_week INTEGER DEFAULT 1,
                next_generation_date TIMESTAMP NOT NULL,
                last_generated_date TIMESTAMP,
                last_invoice_id INTEGER,
                active BOOLEAN DEFAULT 1,
                auto_send BOOLEAN DEFAULT 0,
                auto_pdf BOOLEAN DEFAULT 1,
                email_app TEXT,
                generated_count INTEGER DEFAULT 0,
                notes TEXT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (client_id) REFERENCES clients(id),
                FOREIGN KEY (contract_id) REFERENCES contracts(id),
                FOREIGN KEY (last_invoice_id) REFERENCES invoices(id)
            );

            CREATE INDEX IF NOT EXISTS idx_invoices_company ON invoices(company_id);
            CREATE INDEX IF NOT EXISTS idx_invoice_recipients_invoice ON invoice_recipients(invoice_id);
            CREATE INDEX IF NOT EXISTS idx_invoice_recipients_client ON invoice_recipients(client_id);
            CREATE INDEX IF NOT EXISTS idx_tracking_sessions_client ON tracking_sessions(client_id);
            CREATE INDEX IF NOT EXISTS idx_tracking_sessions_contract ON tracking_sessions(contract_id);
            CREATE INDEX IF NOT EXISTS idx_contracts_client ON contracts(client_id);
            CREATE INDEX IF NOT EXISTS idx_contracts_active ON contracts(active);
            CREATE INDEX IF NOT EXISTS idx_invoice_line_items_invoice ON invoice_line_items(invoice_id);
            CREATE INDEX IF NOT EXISTS idx_recurring_invoices_client ON recurring_invoices(client_id);
            CREATE INDEX IF NOT EXISTS idx_recurring_invoices_active ON recurring_invoices(active);
            CREATE INDEX IF NOT EXISTS idx_recurring_invoices_next_date ON recurring_invoices(next_generation_date);
            CREATE INDEX IF NOT EXISTS idx_expenses_date ON expenses(date);
            CREATE INDEX IF NOT EXISTS idx_expenses_category ON expenses(category);
            """

        // Execute each statement
        let statements = schema.components(separatedBy: ";")
            .map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }
            .filter { !$0.isEmpty }

        for statement in statements {
            try db.execute(sql: statement)
        }
    }

    // MARK: - Generic Database Access

    private func getDatabase() throws -> DatabasePool {
        guard let db = dbPool else {
            throw DatabaseError.notInitialized
        }
        return db
    }

    // MARK: - Company Operations

    func getCompany() async throws -> Company? {
        let db = try getDatabase()
        return try await db.read { db in
            try Company.fetchOne(db)
        }
    }

    func createCompany(_ company: Company) async throws -> Company {
        let db = try getDatabase()
        return try await db.write { db in
            var newCompany = company
            newCompany.createdAt = Date()
            newCompany.updatedAt = Date()
            try newCompany.insert(db)
            return newCompany
        }
    }

    func updateCompany(_ company: Company) async throws {
        let db = try getDatabase()
        try await db.write { db in
            var updated = company
            updated.updatedAt = Date()
            try updated.update(db)
        }
    }

    // MARK: - Client Operations

    func getClients() async throws -> [ClientModel] {
        let db = try getDatabase()
        return try await db.read { db in
            try ClientModel.order(Column("name")).fetchAll(db)
        }
    }

    func getClient(id: Int64) async throws -> ClientModel? {
        let db = try getDatabase()
        return try await db.read { db in
            try ClientModel.fetchOne(db, key: id)
        }
    }

    func createClient(_ client: ClientModel) async throws -> ClientModel {
        let db = try getDatabase()
        return try await db.write { db in
            var newClient = client
            newClient.createdAt = Date()
            newClient.updatedAt = Date()
            try newClient.insert(db)
            return newClient
        }
    }

    func updateClient(_ client: ClientModel) async throws {
        let db = try getDatabase()
        try await db.write { db in
            var updated = client
            updated.updatedAt = Date()
            try updated.update(db)
        }
    }

    func deleteClient(id: Int64) async throws {
        let db = try getDatabase()
        _ = try await db.write { db in
            try ClientModel.deleteOne(db, key: id)
        }
    }

    func getClientCount() async throws -> Int {
        let db = try getDatabase()
        return try await db.read { db in
            try ClientModel.fetchCount(db)
        }
    }

    // MARK: - Contract Operations

    func getContracts() async throws -> [ContractModel] {
        let db = try getDatabase()
        return try await db.read { db in
            try ContractModel.order(Column("name")).fetchAll(db)
        }
    }

    func getActiveContracts() async throws -> [ContractModel] {
        let db = try getDatabase()
        return try await db.read { db in
            try ContractModel
                .filter(Column("active") == true)
                .order(Column("name"))
                .fetchAll(db)
        }
    }

    func getContract(id: Int64) async throws -> ContractModel? {
        let db = try getDatabase()
        return try await db.read { db in
            try ContractModel.fetchOne(db, key: id)
        }
    }

    func createContract(_ contract: ContractModel) async throws -> ContractModel {
        let db = try getDatabase()
        return try await db.write { db in
            var newContract = contract

            // Generate contract number if empty
            if newContract.contractNum.isEmpty {
                let count = try ContractModel.fetchCount(db)
                newContract.contractNum = "C-\(String(format: "%04d", count + 1))"
            }

            newContract.createdAt = Date()
            newContract.updatedAt = Date()
            try newContract.insert(db)
            return newContract
        }
    }

    func updateContract(_ contract: ContractModel) async throws {
        let db = try getDatabase()
        try await db.write { db in
            var updated = contract
            updated.updatedAt = Date()
            try updated.update(db)
        }
    }

    func deleteContract(id: Int64) async throws {
        let db = try getDatabase()
        _ = try await db.write { db in
            try ContractModel.deleteOne(db, key: id)
        }
    }

    func getContractCount() async throws -> Int {
        let db = try getDatabase()
        return try await db.read { db in
            try ContractModel.fetchCount(db)
        }
    }

    // MARK: - Tracking Session Operations

    func getActiveSessions() async throws -> [TrackingSession] {
        let db = try getDatabase()
        return try await db.read { db in
            try TrackingSession
                .filter(Column("end_time") == nil)
                .filter(Column("deleted_at") == nil)
                .order(Column("start_time").desc)
                .fetchAll(db)
        }
    }

    func getActiveSession() async throws -> TrackingSession? {
        let db = try getDatabase()
        return try await db.read { db in
            try TrackingSession
                .filter(Column("end_time") == nil)
                .filter(Column("deleted_at") == nil)
                .order(Column("start_time").desc)
                .fetchOne(db)
        }
    }

    func getRecentSessions(limit: Int = 5) async throws -> [TrackingSession] {
        let db = try getDatabase()
        return try await db.read { db in
            try TrackingSession
                .filter(Column("deleted_at") == nil)
                .order(Column("start_time").desc)
                .limit(limit)
                .fetchAll(db)
        }
    }

    func getWeeklySessions() async throws -> [TrackingSession] {
        let db = try getDatabase()
        let weekAgo = Calendar.current.date(byAdding: .day, value: -7, to: Date())!
        return try await db.read { db in
            try TrackingSession
                .filter(Column("start_time") >= weekAgo)
                .filter(Column("deleted_at") == nil)
                .order(Column("start_time").desc)
                .fetchAll(db)
        }
    }

    func startTracking(projectName: String, clientId: Int64?, contractId: Int64? = nil) async throws -> TrackingSession {
        let db = try getDatabase()
        return try await db.write { db in
            var session = TrackingSession(
                clientId: clientId,
                contractId: contractId,
                projectName: projectName,
                startTime: Date(),
                billable: true
            )
            session.createdAt = Date()
            session.updatedAt = Date()
            try session.insert(db)
            return session
        }
    }

    func stopTracking(sessionId: Int64) async throws -> TrackingSession {
        let db = try getDatabase()
        return try await db.write { db in
            guard var session = try TrackingSession.fetchOne(db, key: sessionId) else {
                throw DatabaseError.notFound
            }

            let endTime = Date()
            session.endTime = endTime
            session.duration = Int(endTime.timeIntervalSince(session.startTime))
            session.hours = Double(session.duration ?? 0) / 3600.0
            session.updatedAt = Date()
            try session.update(db)
            return session
        }
    }

    func updateSession(_ session: TrackingSession) async throws {
        let db = try getDatabase()
        try await db.write { db in
            var updated = session
            updated.updatedAt = Date()
            try updated.update(db)
        }
    }

    func deleteSession(id: Int64, hard: Bool = false) async throws {
        let db = try getDatabase()
        try await db.write { db in
            if hard {
                _ = try TrackingSession.deleteOne(db, key: id)
            } else {
                // Soft delete
                guard var session = try TrackingSession.fetchOne(db, key: id) else {
                    throw DatabaseError.notFound
                }
                session.deletedAt = Date()
                session.updatedAt = Date()
                try session.update(db)
            }
        }
    }

    // MARK: - Invoice Operations

    func getInvoices() async throws -> [Invoice] {
        let db = try getDatabase()
        return try await db.read { db in
            try Invoice.order(Column("issued_date").desc).fetchAll(db)
        }
    }

    func getRecentInvoices(limit: Int = 5) async throws -> [Invoice] {
        let db = try getDatabase()
        return try await db.read { db in
            try Invoice
                .order(Column("issued_date").desc)
                .limit(limit)
                .fetchAll(db)
        }
    }

    func getInvoice(id: Int64) async throws -> Invoice? {
        let db = try getDatabase()
        return try await db.read { db in
            try Invoice.fetchOne(db, key: id)
        }
    }

    func createInvoice(_ invoice: Invoice) async throws -> Invoice {
        let db = try getDatabase()
        return try await db.write { db in
            var newInvoice = invoice

            // Generate invoice number if empty
            if newInvoice.invoiceNum.isEmpty {
                let count = try Invoice.fetchCount(db)
                let year = Calendar.current.component(.year, from: Date())
                newInvoice.invoiceNum = "INV-\(year)-\(String(format: "%04d", count + 1))"
            }

            newInvoice.createdAt = Date()
            newInvoice.updatedAt = Date()
            try newInvoice.insert(db)
            return newInvoice
        }
    }

    func updateInvoice(_ invoice: Invoice) async throws {
        let db = try getDatabase()
        try await db.write { db in
            var updated = invoice
            updated.updatedAt = Date()
            try updated.update(db)
        }
    }

    func updateInvoiceStatus(id: Int64, status: String) async throws {
        let db = try getDatabase()
        try await db.write { db in
            guard var invoice = try Invoice.fetchOne(db, key: id) else {
                throw DatabaseError.notFound
            }
            invoice.status = status
            invoice.updatedAt = Date()
            try invoice.update(db)
        }
    }

    func deleteInvoice(id: Int64) async throws {
        let db = try getDatabase()
        _ = try await db.write { db in
            try Invoice.deleteOne(db, key: id)
        }
    }

    func getInvoiceCount() async throws -> Int {
        let db = try getDatabase()
        return try await db.read { db in
            try Invoice.fetchCount(db)
        }
    }

    func getInvoiceTotals() async throws -> (total: Double, pending: Double, overdue: Double) {
        let db = try getDatabase()
        return try await db.read { db in
            let total = try Double.fetchOne(db, sql: "SELECT COALESCE(SUM(amount), 0) FROM invoices") ?? 0
            let pending = try Double.fetchOne(
                db,
                sql: "SELECT COALESCE(SUM(amount), 0) FROM invoices WHERE status IN ('pending', 'sent')"
            ) ?? 0
            let overdue = try Double.fetchOne(
                db,
                sql: "SELECT COALESCE(SUM(amount), 0) FROM invoices WHERE status = 'overdue'"
            ) ?? 0
            return (total, pending, overdue)
        }
    }

    // MARK: - Expense Operations

    func getExpenses() async throws -> [Expense] {
        let db = try getDatabase()
        return try await db.read { db in
            try Expense.order(Column("date").desc).fetchAll(db)
        }
    }

    func getRecentExpenses(limit: Int = 5) async throws -> [Expense] {
        let db = try getDatabase()
        return try await db.read { db in
            try Expense
                .order(Column("date").desc)
                .limit(limit)
                .fetchAll(db)
        }
    }

    func getExpense(id: Int64) async throws -> Expense? {
        let db = try getDatabase()
        return try await db.read { db in
            try Expense.fetchOne(db, key: id)
        }
    }

    func createExpense(_ expense: Expense) async throws -> Expense {
        let db = try getDatabase()
        return try await db.write { db in
            var newExpense = expense
            newExpense.createdAt = Date()
            newExpense.updatedAt = Date()
            try newExpense.insert(db)
            return newExpense
        }
    }

    func updateExpense(_ expense: Expense) async throws {
        let db = try getDatabase()
        try await db.write { db in
            var updated = expense
            updated.updatedAt = Date()
            try updated.update(db)
        }
    }

    func deleteExpense(id: Int64) async throws {
        let db = try getDatabase()
        _ = try await db.write { db in
            try Expense.deleteOne(db, key: id)
        }
    }

    // MARK: - Dashboard Metrics

    func getDashboardMetrics() async throws -> (
        totalRevenue: Double,
        pendingAmount: Double,
        overdueAmount: Double,
        weeklyHours: Double
    ) {
        let invoiceTotals = try await getInvoiceTotals()
        let weeklySessions = try await getWeeklySessions()

        let weeklyHours = weeklySessions.reduce(0.0) { total, session in
            total + (session.hours ?? Double(session.calculatedDuration) / 3600.0)
        }

        return (
            invoiceTotals.total,
            invoiceTotals.pending,
            invoiceTotals.overdue,
            weeklyHours
        )
    }

    // MARK: - Database Operations

    func resetDatabase() async throws {
        close()
        try fileManager.removeItem(atPath: databasePath)
        try await initialize()
    }

    func exportDatabase(to url: URL) async throws {
        let sourceURL = URL(fileURLWithPath: databasePath)
        if fileManager.fileExists(atPath: url.path) {
            try fileManager.removeItem(at: url)
        }
        try fileManager.copyItem(at: sourceURL, to: url)
    }

    func importDatabase(from url: URL) async throws {
        close()

        // Backup existing
        let backupPath = databasePath + ".backup"
        if fileManager.fileExists(atPath: databasePath) {
            try? fileManager.copyItem(atPath: databasePath, toPath: backupPath)
            try fileManager.removeItem(atPath: databasePath)
        }

        // Copy new database
        try fileManager.copyItem(at: url, to: URL(fileURLWithPath: databasePath))

        // Reinitialize
        try await initialize()
    }
}
