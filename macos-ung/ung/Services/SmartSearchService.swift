//
//  SmartSearchService.swift
//  ung
//
//  Smart search using Apple's Natural Language framework
//  Parses natural language queries like "invoices from last month over $500"
//

import Foundation
import NaturalLanguage

// MARK: - Search Query Types

enum SearchCategory: String, CaseIterable {
    case invoices = "invoices"
    case sessions = "sessions"
    case clients = "clients"
    case expenses = "expenses"
    case contracts = "contracts"
    case all = "all"

    var keywords: [String] {
        switch self {
        case .invoices: return ["invoice", "invoices", "bill", "bills", "payment", "payments", "unpaid", "paid", "overdue"]
        case .sessions: return ["session", "sessions", "time", "hours", "worked", "tracking", "work"]
        case .clients: return ["client", "clients", "customer", "customers", "company", "companies"]
        case .expenses: return ["expense", "expenses", "cost", "costs", "spending", "spent", "purchase"]
        case .contracts: return ["contract", "contracts", "agreement", "project", "projects"]
        case .all: return []
        }
    }
}

enum AmountOperator {
    case greaterThan
    case lessThan
    case equals
    case between(Double, Double)
}

struct AmountFilter {
    let op: AmountOperator
    let value: Double
}

enum StatusFilter: String {
    case paid
    case unpaid
    case pending
    case overdue
    case draft
    case sent
    case active
    case inactive
}

struct ParsedQuery {
    var category: SearchCategory = .all
    var searchText: String = ""
    var dateRange: (start: Date, end: Date)?
    var amountFilter: AmountFilter?
    var statusFilter: StatusFilter?
    var clientName: String?
    var projectName: String?
    var limit: Int?

    var isEmpty: Bool {
        searchText.isEmpty && dateRange == nil && amountFilter == nil &&
        statusFilter == nil && clientName == nil && projectName == nil
    }
}

// MARK: - Smart Search Service

class SmartSearchService {
    static let shared = SmartSearchService()

    private let tagger = NLTagger(tagSchemes: [.nameType, .lexicalClass])
    private let dateDetector: NSDataDetector?

    // Common date keywords
    private let relativeDateKeywords: [String: () -> (Date, Date)] = [
        "today": {
            let start = Calendar.current.startOfDay(for: Date())
            let end = Calendar.current.date(byAdding: .day, value: 1, to: start)!
            return (start, end)
        },
        "yesterday": {
            let today = Calendar.current.startOfDay(for: Date())
            let start = Calendar.current.date(byAdding: .day, value: -1, to: today)!
            return (start, today)
        },
        "this week": {
            let now = Date()
            let calendar = Calendar.current
            let weekday = calendar.component(.weekday, from: now)
            let start = calendar.date(byAdding: .day, value: -(weekday - 1), to: calendar.startOfDay(for: now))!
            let end = calendar.date(byAdding: .day, value: 7, to: start)!
            return (start, end)
        },
        "last week": {
            let now = Date()
            let calendar = Calendar.current
            let weekday = calendar.component(.weekday, from: now)
            let thisWeekStart = calendar.date(byAdding: .day, value: -(weekday - 1), to: calendar.startOfDay(for: now))!
            let start = calendar.date(byAdding: .day, value: -7, to: thisWeekStart)!
            return (start, thisWeekStart)
        },
        "this month": {
            let now = Date()
            let calendar = Calendar.current
            let start = calendar.date(from: calendar.dateComponents([.year, .month], from: now))!
            let end = calendar.date(byAdding: .month, value: 1, to: start)!
            return (start, end)
        },
        "last month": {
            let now = Date()
            let calendar = Calendar.current
            let thisMonth = calendar.date(from: calendar.dateComponents([.year, .month], from: now))!
            let start = calendar.date(byAdding: .month, value: -1, to: thisMonth)!
            return (start, thisMonth)
        },
        "this year": {
            let now = Date()
            let calendar = Calendar.current
            let start = calendar.date(from: calendar.dateComponents([.year], from: now))!
            let end = calendar.date(byAdding: .year, value: 1, to: start)!
            return (start, end)
        },
        "last year": {
            let now = Date()
            let calendar = Calendar.current
            let thisYear = calendar.date(from: calendar.dateComponents([.year], from: now))!
            let start = calendar.date(byAdding: .year, value: -1, to: thisYear)!
            return (start, thisYear)
        }
    ]

    private init() {
        dateDetector = try? NSDataDetector(types: NSTextCheckingResult.CheckingType.date.rawValue)
    }

    // MARK: - Main Parse Function

    func parse(_ query: String) -> ParsedQuery {
        var result = ParsedQuery()
        let lowercased = query.lowercased().trimmingCharacters(in: .whitespacesAndNewlines)

        guard !lowercased.isEmpty else { return result }

        // 1. Detect category
        result.category = detectCategory(lowercased)

        // 2. Extract date range
        result.dateRange = extractDateRange(lowercased)

        // 3. Extract amount filter
        result.amountFilter = extractAmountFilter(lowercased)

        // 4. Extract status filter
        result.statusFilter = extractStatusFilter(lowercased)

        // 5. Extract entity names (clients, projects)
        let entities = extractEntities(query)
        result.clientName = entities.client
        result.projectName = entities.project

        // 6. Extract limit
        result.limit = extractLimit(lowercased)

        // 7. Clean search text (remove parsed parts)
        result.searchText = cleanSearchText(lowercased, parsed: result)

        return result
    }

    // MARK: - Category Detection

    private func detectCategory(_ query: String) -> SearchCategory {
        for category in SearchCategory.allCases where category != .all {
            for keyword in category.keywords {
                if query.contains(keyword) {
                    return category
                }
            }
        }
        return .all
    }

    // MARK: - Date Extraction

    private func extractDateRange(_ query: String) -> (Date, Date)? {
        // Check relative date keywords first
        for (keyword, dateGenerator) in relativeDateKeywords {
            if query.contains(keyword) {
                return dateGenerator()
            }
        }

        // Check for month names
        let months = ["january", "february", "march", "april", "may", "june",
                      "july", "august", "september", "october", "november", "december"]
        for (index, month) in months.enumerated() {
            if query.contains(month) || query.contains(month.prefix(3)) {
                let calendar = Calendar.current
                let year = calendar.component(.year, from: Date())
                var components = DateComponents()
                components.year = year
                components.month = index + 1
                components.day = 1
                if let start = calendar.date(from: components),
                   let end = calendar.date(byAdding: .month, value: 1, to: start) {
                    return (start, end)
                }
            }
        }

        // Check for "last X days/weeks/months"
        let patterns = [
            (pattern: #"last\s+(\d+)\s+days?"#, unit: Calendar.Component.day),
            (pattern: #"last\s+(\d+)\s+weeks?"#, unit: Calendar.Component.weekOfYear),
            (pattern: #"last\s+(\d+)\s+months?"#, unit: Calendar.Component.month)
        ]

        for (pattern, unit) in patterns {
            if let regex = try? NSRegularExpression(pattern: pattern, options: .caseInsensitive),
               let match = regex.firstMatch(in: query, range: NSRange(query.startIndex..., in: query)),
               let numberRange = Range(match.range(at: 1), in: query),
               let number = Int(query[numberRange]) {
                let end = Date()
                let start = Calendar.current.date(byAdding: unit, value: -number, to: end)!
                return (start, end)
            }
        }

        return nil
    }

    // MARK: - Amount Extraction

    private func extractAmountFilter(_ query: String) -> AmountFilter? {
        // Patterns: "over $500", "more than 100", "under $50", "less than 200", "between 100 and 500"
        let patterns: [(String, AmountOperator)] = [
            (#"(?:over|more than|greater than|above|>\s*)\$?(\d+(?:\.\d{2})?)"#, .greaterThan),
            (#"(?:under|less than|below|<\s*)\$?(\d+(?:\.\d{2})?)"#, .lessThan),
            (#"(?:exactly|equals?|=\s*)\$?(\d+(?:\.\d{2})?)"#, .equals)
        ]

        for (pattern, op) in patterns {
            if let regex = try? NSRegularExpression(pattern: pattern, options: .caseInsensitive),
               let match = regex.firstMatch(in: query, range: NSRange(query.startIndex..., in: query)),
               let valueRange = Range(match.range(at: 1), in: query),
               let value = Double(query[valueRange]) {
                return AmountFilter(op: op, value: value)
            }
        }

        // Between pattern
        let betweenPattern = #"between\s+\$?(\d+(?:\.\d{2})?)\s+and\s+\$?(\d+(?:\.\d{2})?)"#
        if let regex = try? NSRegularExpression(pattern: betweenPattern, options: .caseInsensitive),
           let match = regex.firstMatch(in: query, range: NSRange(query.startIndex..., in: query)),
           let lowRange = Range(match.range(at: 1), in: query),
           let highRange = Range(match.range(at: 2), in: query),
           let low = Double(query[lowRange]),
           let high = Double(query[highRange]) {
            return AmountFilter(op: .between(low, high), value: low)
        }

        return nil
    }

    // MARK: - Status Extraction

    private func extractStatusFilter(_ query: String) -> StatusFilter? {
        let statusKeywords: [String: StatusFilter] = [
            "unpaid": .unpaid,
            "not paid": .unpaid,
            "pending": .pending,
            "paid": .paid,
            "overdue": .overdue,
            "late": .overdue,
            "draft": .draft,
            "sent": .sent,
            "active": .active,
            "inactive": .inactive
        ]

        for (keyword, status) in statusKeywords {
            if query.contains(keyword) {
                return status
            }
        }

        return nil
    }

    // MARK: - Entity Extraction

    private func extractEntities(_ query: String) -> (client: String?, project: String?) {
        var clientName: String?
        var projectName: String?

        // Use NLTagger to find organization/person names
        tagger.string = query

        tagger.enumerateTags(in: query.startIndex..<query.endIndex, unit: .word, scheme: .nameType) { tag, range in
            if let tag = tag {
                let entity = String(query[range])
                switch tag {
                case .organizationName, .personalName:
                    if clientName == nil {
                        clientName = entity
                    }
                default:
                    break
                }
            }
            return true
        }

        // Check for "for [client]" or "from [client]" patterns
        let forPattern = #"(?:for|from|with|by)\s+([A-Z][a-zA-Z\s]+?)(?:\s+(?:in|on|last|this|over|under)|$)"#
        if let regex = try? NSRegularExpression(pattern: forPattern, options: []),
           let match = regex.firstMatch(in: query, range: NSRange(query.startIndex..., in: query)),
           let nameRange = Range(match.range(at: 1), in: query) {
            let name = String(query[nameRange]).trimmingCharacters(in: .whitespaces)
            if !name.isEmpty && name.count > 2 {
                clientName = name
            }
        }

        // Check for project patterns
        let projectPattern = #"(?:project|on)\s+[\"']?([^\"']+)[\"']?"#
        if let regex = try? NSRegularExpression(pattern: projectPattern, options: .caseInsensitive),
           let match = regex.firstMatch(in: query, range: NSRange(query.startIndex..., in: query)),
           let nameRange = Range(match.range(at: 1), in: query) {
            projectName = String(query[nameRange]).trimmingCharacters(in: .whitespaces)
        }

        return (clientName, projectName)
    }

    // MARK: - Limit Extraction

    private func extractLimit(_ query: String) -> Int? {
        let patterns = [
            #"(?:top|first|last|show)\s+(\d+)"#,
            #"(\d+)\s+(?:results?|items?|records?)"#
        ]

        for pattern in patterns {
            if let regex = try? NSRegularExpression(pattern: pattern, options: .caseInsensitive),
               let match = regex.firstMatch(in: query, range: NSRange(query.startIndex..., in: query)),
               let numberRange = Range(match.range(at: 1), in: query),
               let number = Int(query[numberRange]) {
                return min(number, 100) // Cap at 100
            }
        }

        return nil
    }

    // MARK: - Clean Search Text

    private func cleanSearchText(_ query: String, parsed: ParsedQuery) -> String {
        var cleaned = query

        // Remove category keywords
        for keyword in parsed.category.keywords {
            cleaned = cleaned.replacingOccurrences(of: keyword, with: "")
        }

        // Remove date keywords
        for keyword in relativeDateKeywords.keys {
            cleaned = cleaned.replacingOccurrences(of: keyword, with: "")
        }

        // Remove status keywords
        if let status = parsed.statusFilter {
            cleaned = cleaned.replacingOccurrences(of: status.rawValue, with: "")
        }

        // Remove common filler words
        let fillers = ["show", "me", "all", "the", "my", "find", "get", "list", "from", "in", "on", "for", "with"]
        for filler in fillers {
            cleaned = cleaned.replacingOccurrences(of: "\\b\(filler)\\b", with: "", options: .regularExpression)
        }

        // Clean up whitespace
        cleaned = cleaned.components(separatedBy: .whitespaces)
            .filter { !$0.isEmpty }
            .joined(separator: " ")

        return cleaned
    }
}

// MARK: - Search Result Types

struct SearchResult: Identifiable {
    let id: String
    let type: SearchCategory
    let title: String
    let subtitle: String
    let detail: String?
    let date: Date?
    let amount: Double?
    let icon: String
    let deepLink: String

    init(type: SearchCategory, title: String, subtitle: String, detail: String? = nil,
         date: Date? = nil, amount: Double? = nil, deepLink: String) {
        self.id = UUID().uuidString
        self.type = type
        self.title = title
        self.subtitle = subtitle
        self.detail = detail
        self.date = date
        self.amount = amount
        self.deepLink = deepLink

        switch type {
        case .invoices: self.icon = "doc.text.fill"
        case .sessions: self.icon = "clock.fill"
        case .clients: self.icon = "person.fill"
        case .expenses: self.icon = "dollarsign.circle.fill"
        case .contracts: self.icon = "doc.badge.ellipsis"
        case .all: self.icon = "magnifyingglass"
        }
    }
}

// MARK: - Query Suggestions

extension SmartSearchService {
    func getSuggestions(for partial: String) -> [String] {
        let lowercased = partial.lowercased()

        var suggestions: [String] = []

        // Category suggestions
        if lowercased.isEmpty || "invoices".hasPrefix(lowercased) {
            suggestions.append("invoices this month")
            suggestions.append("unpaid invoices")
            suggestions.append("invoices over $500")
        }

        if lowercased.isEmpty || "sessions".hasPrefix(lowercased) || "hours".hasPrefix(lowercased) {
            suggestions.append("hours this week")
            suggestions.append("sessions last month")
        }

        if lowercased.isEmpty || "expenses".hasPrefix(lowercased) {
            suggestions.append("expenses this month")
            suggestions.append("expenses over $100")
        }

        // Time-based suggestions
        if lowercased.contains("last") {
            suggestions.append("\(partial) week")
            suggestions.append("\(partial) month")
            suggestions.append("\(partial) 30 days")
        }

        return Array(suggestions.prefix(5))
    }
}
