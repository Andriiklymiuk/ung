//
//  CLIService.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import Foundation

actor CLIService {
  private let cliPath: String

  init() {
    // Try common installation paths
    let paths = [
      "/usr/local/bin/ung",
      "/opt/homebrew/bin/ung",
      "\(NSHomeDirectory())/go/bin/ung",
      "/usr/bin/ung",
    ]

    self.cliPath = paths.first { FileManager.default.fileExists(atPath: $0) } ?? "ung"
  }

  // MARK: - CLI Execution
  private func execute(_ args: [String], useGlobal: Bool = true) async -> (
    output: String, error: String, exitCode: Int32
  ) {
    let process = Process()
    let outputPipe = Pipe()
    let errorPipe = Pipe()

    var arguments = args
    if useGlobal {
      arguments.insert("--global", at: 0)
    }

    let cliPath = findCliPath()
    process.executableURL = URL(fileURLWithPath: cliPath)
    process.arguments = arguments
    process.standardOutput = outputPipe
    process.standardError = errorPipe

    // Set environment with comprehensive PATH
    var env = ProcessInfo.processInfo.environment
    env["PATH"] =
      "/opt/homebrew/bin:/usr/local/bin:\(NSHomeDirectory())/go/bin:/usr/bin:/bin:/sbin:/usr/sbin"
    env["HOME"] = NSHomeDirectory()
    process.environment = env

    // Set current directory to home to avoid sandbox issues
    process.currentDirectoryURL = URL(fileURLWithPath: NSHomeDirectory())

    do {
      try process.run()
      process.waitUntilExit()

      let outputData = outputPipe.fileHandleForReading.readDataToEndOfFile()
      let errorData = errorPipe.fileHandleForReading.readDataToEndOfFile()

      let output = String(data: outputData, encoding: .utf8) ?? ""
      let error = String(data: errorData, encoding: .utf8) ?? ""

      return (output, error, process.terminationStatus)
    } catch {
      print("[CLIService] Execute error: \(error)")
      return ("", error.localizedDescription, -1)
    }
  }

  private func findCliPath() -> String {
    let paths = [
      "/opt/homebrew/bin/ung",
      "/usr/local/bin/ung",
      "\(NSHomeDirectory())/go/bin/ung",
      "/usr/bin/ung",
    ]

    print("[CLIService] Checking CLI paths:")
    for path in paths {
      let exists = FileManager.default.fileExists(atPath: path)
      print("  - \(path): \(exists ? "✓ FOUND" : "✗ not found")")
    }

    // Try to find existing installation
    if let found = paths.first(where: { FileManager.default.fileExists(atPath: $0) }) {
      print("[CLIService] Using CLI at: \(found)")
      return found
    }

    // Fallback to PATH lookup
    print("[CLIService] No CLI found in standard paths, using PATH")
    return "ung"
  }

  // MARK: - Status Checks
  func isCliInstalled() async -> Bool {
    // First check if the binary exists at known paths
    let paths = [
      "/opt/homebrew/bin/ung",
      "/usr/local/bin/ung",
      "\(NSHomeDirectory())/go/bin/ung",
      "/usr/bin/ung",
    ]

    print("[CLIService] isCliInstalled - checking paths:")
    for path in paths {
      let exists = FileManager.default.fileExists(atPath: path)
      print("  - \(path): \(exists ? "EXISTS" : "not found")")
      if exists {
        // Verify it's executable
        let isExecutable = FileManager.default.isExecutableFile(atPath: path)
        print("    executable: \(isExecutable)")
        if isExecutable {
          // Try to run it
          let result = await execute(["--version"], useGlobal: false)
          print("[CLIService] Version check result:")
          print("  - Output: \(result.output)")
          print("  - Error: \(result.error)")
          print("  - Exit code: \(result.exitCode)")
          return result.exitCode == 0
        }
      }
    }

    // Fallback: try to execute anyway (in case it's in PATH)
    let result = await execute(["--version"], useGlobal: false)
    print("[CLIService] Fallback version check:")
    print("  - Output: \(result.output)")
    print("  - Error: \(result.error)")
    print("  - Exit code: \(result.exitCode)")
    return result.exitCode == 0
  }

  func isInitialized() async -> Bool {
    // Check if global database exists
    let globalDbPath = "\(NSHomeDirectory())/.ung/ung.db"
    return FileManager.default.fileExists(atPath: globalDbPath)
  }

  // MARK: - Initialization
  func initializeGlobal() async -> Bool {
    let result = await execute(["init", "--global"], useGlobal: false)
    return result.exitCode == 0
  }

  func initializeLocal() async -> Bool {
    let result = await execute(["init"], useGlobal: false)
    return result.exitCode == 0
  }

  // MARK: - Dashboard Data
  func getDashboardMetrics() async -> DashboardMetrics? {
    var metrics = await MainActor.run { DashboardMetrics() }

    // Get invoice totals
    let invoiceResult = await execute(["invoice", "list"])
    if invoiceResult.exitCode == 0 {
      let (total, pending, overdue) = parseInvoiceTotals(invoiceResult.output)
      metrics.totalRevenue = total
      metrics.pendingAmount = pending
      metrics.overdueAmount = overdue
    }

    // Get weekly hours
    let trackingResult = await execute(["track", "list", "--week"])
    if trackingResult.exitCode == 0 {
      metrics.weeklyHours = parseWeeklyHours(trackingResult.output)
    }

    return metrics
  }

  func getSetupStatus() async -> SetupStatus {
    var status = await MainActor.run { SetupStatus() }

    // Check company
    let companyResult = await execute(["company", "view"])
    status.hasCompany = companyResult.exitCode == 0 && !companyResult.output.contains("No company")

    // Check clients
    let clientResult = await execute(["client", "list"])
    status.hasClient = clientResult.exitCode == 0 && clientResult.output.contains("│")

    // Check contracts
    let contractResult = await execute(["contract", "list"])
    status.hasContract = contractResult.exitCode == 0 && contractResult.output.contains("│")

    return status
  }

  func getActiveSession() async -> ActiveSession? {
    let result = await execute(["track", "status"])

    if result.exitCode == 0 && result.output.contains("Active") {
      return parseActiveSession(result.output)
    }
    return nil
  }

  func getRecentInvoices() async -> [RecentInvoice] {
    let result = await execute(["invoice", "list"])
    if result.exitCode == 0 {
      return parseInvoices(result.output)
    }
    return []
  }

  func getRecentSessions() async -> [RecentSession] {
    let result = await execute(["track", "list", "--limit", "5"])
    if result.exitCode == 0 {
      return parseSessions(result.output)
    }
    return []
  }

  func getRecentExpenses() async -> [RecentExpense] {
    let result = await execute(["expense", "list", "--limit", "5"])
    if result.exitCode == 0 {
      return parseExpenses(result.output)
    }
    return []
  }

  func getEntityCounts() async -> (clients: Int, contracts: Int, invoices: Int) {
    var counts = (clients: 0, contracts: 0, invoices: 0)

    let clientResult = await execute(["client", "list"])
    if clientResult.exitCode == 0 {
      counts.clients = countTableRows(clientResult.output)
    }

    let contractResult = await execute(["contract", "list"])
    if contractResult.exitCode == 0 {
      counts.contracts = countTableRows(contractResult.output)
    }

    let invoiceResult = await execute(["invoice", "list"])
    if invoiceResult.exitCode == 0 {
      counts.invoices = countTableRows(invoiceResult.output)
    }

    return counts
  }

  func getClients() async -> [Client] {
    let result = await execute(["client", "list"])
    if result.exitCode == 0 {
      return parseClients(result.output)
    }
    return []
  }

  func getContracts() async -> [Contract] {
    let result = await execute(["contract", "list"])
    if result.exitCode == 0 {
      return parseContracts(result.output)
    }
    return []
  }

  // MARK: - Tracking
  func startTracking(project: String, clientId: Int?) async -> Bool {
    var args = ["track", "start", project]
    if let clientId = clientId {
      args.append(contentsOf: ["--client", "\(clientId)"])
    }
    let result = await execute(args)
    return result.exitCode == 0
  }

  func stopTracking() async -> Bool {
    let result = await execute(["track", "stop"])
    return result.exitCode == 0
  }

  // MARK: - Company
  func createCompany(name: String, email: String?) async -> Bool {
    var args = ["company", "create", name]
    if let email = email {
      args.append(contentsOf: ["--email", email])
    }
    let result = await execute(args)
    return result.exitCode == 0
  }

  func updateCompany(
    name: String?, email: String?, address: String?, phone: String?, taxId: String?
  ) async -> Bool {
    var args = ["company", "update"]
    if let name = name {
      args.append(contentsOf: ["--name", name])
    }
    if let email = email {
      args.append(contentsOf: ["--email", email])
    }
    if let address = address {
      args.append(contentsOf: ["--address", address])
    }
    if let phone = phone {
      args.append(contentsOf: ["--phone", phone])
    }
    if let taxId = taxId {
      args.append(contentsOf: ["--tax-id", taxId])
    }
    let result = await execute(args)
    return result.exitCode == 0
  }

  func getCompanyDetails() async -> (
    name: String, email: String, address: String, phone: String, taxId: String
  )? {
    let result = await execute(["company", "view"])
    if result.exitCode == 0 && !result.output.contains("No company") {
      return parseCompanyDetails(result.output)
    }
    return nil
  }

  // MARK: - Clients
  func createClient(name: String, email: String?) async -> Bool {
    var args = ["client", "create", name]
    if let email = email {
      args.append(contentsOf: ["--email", email])
    }
    let result = await execute(args)
    return result.exitCode == 0
  }

  func updateClient(id: Int, name: String?, email: String?) async -> Bool {
    var args = ["client", "update", "\(id)"]
    if let name = name {
      args.append(contentsOf: ["--name", name])
    }
    if let email = email {
      args.append(contentsOf: ["--email", email])
    }
    let result = await execute(args)
    return result.exitCode == 0
  }

  func deleteClient(id: Int) async -> Bool {
    let result = await execute(["client", "delete", "\(id)", "--force"])
    return result.exitCode == 0
  }

  func getClientDetails(id: Int) async -> (
    name: String, email: String, address: String, phone: String
  )? {
    let result = await execute(["client", "view", "\(id)"])
    if result.exitCode == 0 {
      return parseClientDetails(result.output)
    }
    return nil
  }

  // MARK: - Contracts
  func createContract(name: String, clientId: Int, rate: Double, type: String) async -> Bool {
    let args = [
      "contract", "create", name, "--client", "\(clientId)", "--rate", "\(rate)", "--type", type,
    ]
    let result = await execute(args)
    return result.exitCode == 0
  }

  func updateContract(id: Int, name: String?, rate: Double?, type: String?) async -> Bool {
    var args = ["contract", "update", "\(id)"]
    if let name = name {
      args.append(contentsOf: ["--name", name])
    }
    if let rate = rate {
      args.append(contentsOf: ["--rate", "\(rate)"])
    }
    if let type = type {
      args.append(contentsOf: ["--type", type])
    }
    let result = await execute(args)
    return result.exitCode == 0
  }

  func deleteContract(id: Int) async -> Bool {
    let result = await execute(["contract", "delete", "\(id)", "--force"])
    return result.exitCode == 0
  }

  // MARK: - Expenses
  func logExpense(description: String, amount: Double, category: String) async -> Bool {
    let args = ["expense", "log", description, "--amount", "\(amount)", "--category", category]
    let result = await execute(args)
    return result.exitCode == 0
  }

  func updateExpense(id: Int, description: String?, amount: Double?, category: String?) async
    -> Bool
  {
    var args = ["expense", "update", "\(id)"]
    if let description = description {
      args.append(contentsOf: ["--description", description])
    }
    if let amount = amount {
      args.append(contentsOf: ["--amount", "\(amount)"])
    }
    if let category = category {
      args.append(contentsOf: ["--category", category])
    }
    let result = await execute(args)
    return result.exitCode == 0
  }

  func deleteExpense(id: Int) async -> Bool {
    let result = await execute(["expense", "delete", "\(id)", "--force"])
    return result.exitCode == 0
  }

  func getExpenseCategories() async -> [String] {
    return ["Software", "Hardware", "Travel", "Meals", "Office", "Marketing", "Education", "Other"]
  }

  // MARK: - Invoices
  func createInvoice(clientId: Int) async -> Bool {
    let args = ["invoice", "create", "--client", "\(clientId)"]
    let result = await execute(args)
    return result.exitCode == 0
  }

  func createInvoiceFromSessions(contractId: Int, sessionIds: [Int]?) async -> Bool {
    var args = ["invoice", "create", "--contract", "\(contractId)"]
    if let sessionIds = sessionIds {
      args.append(contentsOf: ["--sessions", sessionIds.map { "\($0)" }.joined(separator: ",")])
    }
    let result = await execute(args)
    return result.exitCode == 0
  }

  func updateInvoiceStatus(invoiceId: Int, status: String) async -> Bool {
    let args = ["invoice", "status", "\(invoiceId)", status]
    let result = await execute(args)
    return result.exitCode == 0
  }

  func markInvoicePaid(invoiceId: Int) async -> Bool {
    return await updateInvoiceStatus(invoiceId: invoiceId, status: "paid")
  }

  func markInvoiceSent(invoiceId: Int) async -> Bool {
    return await updateInvoiceStatus(invoiceId: invoiceId, status: "sent")
  }

  func deleteInvoice(id: Int) async -> Bool {
    let result = await execute(["invoice", "delete", "\(id)", "--force"])
    return result.exitCode == 0
  }

  func generateInvoicePDF(invoiceId: Int, outputPath: String) async -> Bool {
    let args = ["invoice", "pdf", "\(invoiceId)", "--output", outputPath]
    let result = await execute(args)
    return result.exitCode == 0
  }

  func sendInvoiceEmail(invoiceId: Int) async -> Bool {
    let args = ["invoice", "send", "\(invoiceId)"]
    let result = await execute(args)
    return result.exitCode == 0
  }

  // MARK: - Sessions
  func deleteSession(id: Int) async -> Bool {
    let result = await execute(["track", "delete", "\(id)", "--force"])
    return result.exitCode == 0
  }

  func updateSession(id: Int, project: String?, notes: String?) async -> Bool {
    var args = ["track", "update", "\(id)"]
    if let project = project {
      args.append(contentsOf: ["--project", project])
    }
    if let notes = notes {
      args.append(contentsOf: ["--notes", notes])
    }
    let result = await execute(args)
    return result.exitCode == 0
  }

  // MARK: - Database Operations
  func importDatabase(from path: String) async -> Bool {
    // Check file extension
    if path.hasSuffix(".sql") {
      // Import SQL file
      let result = await execute(["db", "import", path], useGlobal: true)
      return result.exitCode == 0
    } else if path.hasSuffix(".db") {
      // Copy database file directly
      let globalDbPath = "\(NSHomeDirectory())/.ung/ung.db"
      do {
        // Backup existing database first
        let backupPath = "\(NSHomeDirectory())/.ung/ung_backup_\(Date().timeIntervalSince1970).db"
        if FileManager.default.fileExists(atPath: globalDbPath) {
          try FileManager.default.copyItem(atPath: globalDbPath, toPath: backupPath)
        }
        // Copy new database
        try FileManager.default.removeItem(atPath: globalDbPath)
        try FileManager.default.copyItem(atPath: path, toPath: globalDbPath)
        return true
      } catch {
        return false
      }
    }
    return false
  }

  func exportDatabase(to path: String) async -> Bool {
    let result = await execute(["db", "export", path], useGlobal: true)
    if result.exitCode != 0 {
      // Fallback: copy the database file directly
      let globalDbPath = "\(NSHomeDirectory())/.ung/ung.db"
      do {
        if FileManager.default.fileExists(atPath: path) {
          try FileManager.default.removeItem(atPath: path)
        }
        try FileManager.default.copyItem(atPath: globalDbPath, toPath: path)
        return true
      } catch {
        return false
      }
    }
    return true
  }

  func backupDatabase(to path: String) async -> Bool {
    let globalDbPath = "\(NSHomeDirectory())/.ung/ung.db"
    do {
      if FileManager.default.fileExists(atPath: path) {
        try FileManager.default.removeItem(atPath: path)
      }
      try FileManager.default.copyItem(atPath: globalDbPath, toPath: path)
      return true
    } catch {
      return false
    }
  }

  func resetDatabase() async -> Bool {
    let result = await execute(["db", "reset", "--force"], useGlobal: true)
    if result.exitCode != 0 {
      // Fallback: delete database file
      let globalDbPath = "\(NSHomeDirectory())/.ung/ung.db"
      do {
        if FileManager.default.fileExists(atPath: globalDbPath) {
          try FileManager.default.removeItem(atPath: globalDbPath)
        }
        return true
      } catch {
        return false
      }
    }
    return true
  }

  // MARK: - Parsing Helpers
  private func parseInvoiceTotals(_ output: String) -> (
    total: Double, pending: Double, overdue: Double
  ) {
    var total: Double = 0
    var pending: Double = 0
    var overdue: Double = 0

    let lines = output.components(separatedBy: "\n")
    for line in lines {
      guard line.contains("│") else { continue }
      let parts = line.components(separatedBy: "│").map { $0.trimmingCharacters(in: .whitespaces) }
      guard parts.count >= 5 else { continue }

      // Parse amount (remove $ and commas)
      let amountStr = parts[3].replacingOccurrences(of: "$", with: "").replacingOccurrences(
        of: ",", with: "")
      let amount = Double(amountStr) ?? 0
      let status = parts[4].lowercased()

      total += amount
      if status.contains("pending") || status.contains("sent") {
        pending += amount
      } else if status.contains("overdue") {
        overdue += amount
      }
    }

    return (total, pending, overdue)
  }

  private func parseWeeklyHours(_ output: String) -> Double {
    var totalMinutes: Double = 0

    let lines = output.components(separatedBy: "\n")
    for line in lines {
      guard line.contains("│") else { continue }

      // Look for duration pattern like "2:30" or "2:30:00"
      if let match = line.range(of: #"\d+:\d+"#, options: .regularExpression) {
        let duration = String(line[match])
        let parts = duration.components(separatedBy: ":")
        if parts.count >= 2 {
          let hours = Double(parts[0]) ?? 0
          let minutes = Double(parts[1]) ?? 0
          totalMinutes += hours * 60 + minutes
        }
      }
    }

    return totalMinutes / 60.0
  }

  private func parseActiveSession(_ output: String) -> ActiveSession? {
    // Parse output like "Active session: Project Name (Client) - 1:23:45"
    let lines = output.components(separatedBy: "\n")

    for line in lines {
      if line.contains("Active") || line.contains("Tracking") {
        // Extract project name and duration
        var project = "Project"
        var client = ""
        var elapsed = 0

        // Try to find duration pattern
        if let match = line.range(of: #"(\d+):(\d+):(\d+)"#, options: .regularExpression) {
          let duration = String(line[match])
          let parts = duration.components(separatedBy: ":")
          if parts.count == 3 {
            let hours = Int(parts[0]) ?? 0
            let minutes = Int(parts[1]) ?? 0
            let seconds = Int(parts[2]) ?? 0
            elapsed = hours * 3600 + minutes * 60 + seconds
          }
        }

        // Try to extract project name
        if let projectMatch = line.range(of: #":\s*(.+?)\s*(\(|$|-)"#, options: .regularExpression)
        {
          project = String(line[projectMatch]).trimmingCharacters(
            in: CharacterSet(charactersIn: ": ()-"))
        }

        // Try to extract client
        if let clientMatch = line.range(of: #"\(([^)]+)\)"#, options: .regularExpression) {
          client = String(line[clientMatch]).trimmingCharacters(
            in: CharacterSet(charactersIn: "()"))
        }

        return ActiveSession(
          id: 1,
          project: project.isEmpty ? "Active Session" : project,
          client: client,
          startTime: Date().addingTimeInterval(-Double(elapsed)),
          elapsedSeconds: elapsed
        )
      }
    }
    return nil
  }

  private func parseInvoices(_ output: String) -> [RecentInvoice] {
    var invoices: [RecentInvoice] = []

    let lines = output.components(separatedBy: "\n")
    for line in lines {
      guard line.contains("│") else { continue }
      let parts = line.components(separatedBy: "│").map { $0.trimmingCharacters(in: .whitespaces) }
      guard parts.count >= 5, let id = Int(parts[0]) else { continue }

      invoices.append(
        RecentInvoice(
          id: id,
          invoiceNum: parts[1],
          client: parts[2],
          amount: parts[3],
          status: parts[4].lowercased()
        ))

      if invoices.count >= 3 { break }
    }

    return invoices
  }

  private func parseSessions(_ output: String) -> [RecentSession] {
    var sessions: [RecentSession] = []

    let lines = output.components(separatedBy: "\n")
    for line in lines {
      guard line.contains("│") else { continue }
      let parts = line.components(separatedBy: "│").map { $0.trimmingCharacters(in: .whitespaces) }
      guard parts.count >= 4, let id = Int(parts[0]) else { continue }

      sessions.append(
        RecentSession(
          id: id,
          project: parts[1],
          duration: parts[2],
          date: parts[3]
        ))

      if sessions.count >= 5 { break }
    }

    return sessions
  }

  private func parseExpenses(_ output: String) -> [RecentExpense] {
    var expenses: [RecentExpense] = []

    let lines = output.components(separatedBy: "\n")
    for line in lines {
      guard line.contains("│") else { continue }
      let parts = line.components(separatedBy: "│").map { $0.trimmingCharacters(in: .whitespaces) }
      guard parts.count >= 4, let id = Int(parts[0]) else { continue }

      expenses.append(
        RecentExpense(
          id: id,
          description: parts[1],
          amount: parts[2],
          category: parts[3]
        ))

      if expenses.count >= 5 { break }
    }

    return expenses
  }

  private func parseClients(_ output: String) -> [Client] {
    var clients: [Client] = []

    let lines = output.components(separatedBy: "\n")
    for line in lines {
      guard line.contains("│") else { continue }
      let parts = line.components(separatedBy: "│").map { $0.trimmingCharacters(in: .whitespaces) }
      guard parts.count >= 2, let id = Int(parts[0]) else { continue }

      clients.append(Client(id: id, name: parts[1]))
    }

    return clients
  }

  private func parseContracts(_ output: String) -> [Contract] {
    var contracts: [Contract] = []

    let lines = output.components(separatedBy: "\n")
    for line in lines {
      guard line.contains("│") else { continue }
      let parts = line.components(separatedBy: "│").map { $0.trimmingCharacters(in: .whitespaces) }
      guard parts.count >= 3, let id = Int(parts[0]) else { continue }

      contracts.append(Contract(id: id, name: parts[1], clientName: parts[2]))
    }

    return contracts
  }

  private func countTableRows(_ output: String) -> Int {
    return output.components(separatedBy: "\n")
      .filter { $0.contains("│") && !$0.contains("ID") && !$0.contains("─") }
      .count
  }

  private func parseCompanyDetails(_ output: String) -> (
    name: String, email: String, address: String, phone: String, taxId: String
  )? {
    var name = ""
    var email = ""
    var address = ""
    var phone = ""
    var taxId = ""

    let lines = output.components(separatedBy: "\n")
    for line in lines {
      let trimmed = line.trimmingCharacters(in: .whitespaces)
      if trimmed.hasPrefix("Name:") {
        name = trimmed.replacingOccurrences(of: "Name:", with: "").trimmingCharacters(
          in: .whitespaces)
      } else if trimmed.hasPrefix("Email:") {
        email = trimmed.replacingOccurrences(of: "Email:", with: "").trimmingCharacters(
          in: .whitespaces)
      } else if trimmed.hasPrefix("Address:") {
        address = trimmed.replacingOccurrences(of: "Address:", with: "").trimmingCharacters(
          in: .whitespaces)
      } else if trimmed.hasPrefix("Phone:") {
        phone = trimmed.replacingOccurrences(of: "Phone:", with: "").trimmingCharacters(
          in: .whitespaces)
      } else if trimmed.hasPrefix("Tax ID:") || trimmed.hasPrefix("TaxID:") {
        taxId = trimmed.replacingOccurrences(of: "Tax ID:", with: "").replacingOccurrences(
          of: "TaxID:", with: ""
        ).trimmingCharacters(in: .whitespaces)
      }
    }

    if name.isEmpty { return nil }
    return (name, email, address, phone, taxId)
  }

  private func parseClientDetails(_ output: String) -> (
    name: String, email: String, address: String, phone: String
  )? {
    var name = ""
    var email = ""
    var address = ""
    var phone = ""

    let lines = output.components(separatedBy: "\n")
    for line in lines {
      let trimmed = line.trimmingCharacters(in: .whitespaces)
      if trimmed.hasPrefix("Name:") {
        name = trimmed.replacingOccurrences(of: "Name:", with: "").trimmingCharacters(
          in: .whitespaces)
      } else if trimmed.hasPrefix("Email:") {
        email = trimmed.replacingOccurrences(of: "Email:", with: "").trimmingCharacters(
          in: .whitespaces)
      } else if trimmed.hasPrefix("Address:") {
        address = trimmed.replacingOccurrences(of: "Address:", with: "").trimmingCharacters(
          in: .whitespaces)
      } else if trimmed.hasPrefix("Phone:") {
        phone = trimmed.replacingOccurrences(of: "Phone:", with: "").trimmingCharacters(
          in: .whitespaces)
      }
    }

    if name.isEmpty { return nil }
    return (name, email, address, phone)
  }
}
