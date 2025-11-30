//
//  DigService.swift
//  ung
//
//  Dig - Idea Analysis & Incubation Service
//  Analyze ideas from multiple perspectives and get actionable insights
//

import Foundation
import SwiftUI
import Security

// MARK: - Dig Models

struct DigSession: Codable, Identifiable {
    let id: Int64?
    var title: String?
    let rawIdea: String
    var refinedIdea: String?
    var status: DigSessionStatus
    var overallScore: Double?
    var recommendation: DigRecommendation?
    var currentStage: String?
    var stagesCompleted: String?
    var startedAt: Date?
    var completedAt: Date?
    let createdAt: Date?
    var updatedAt: Date?

    // Relationships (when loaded)
    var analyses: [DigAnalysis]?
    var executionPlan: DigExecutionPlan?
    var marketing: DigMarketing?
    var revenueProjection: DigRevenueProjection?
    var alternatives: [DigAlternative]?

    enum CodingKeys: String, CodingKey {
        case id, title, status, recommendation
        case rawIdea = "raw_idea"
        case refinedIdea = "refined_idea"
        case overallScore = "overall_score"
        case currentStage = "current_stage"
        case stagesCompleted = "stages_completed"
        case startedAt = "started_at"
        case completedAt = "completed_at"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
        case analyses
        case executionPlan = "execution_plan"
        case marketing
        case revenueProjection = "revenue_projection"
        case alternatives
    }

    var stagesCompletedArray: [String] {
        guard let stagesCompleted = stagesCompleted,
              let data = stagesCompleted.data(using: .utf8),
              let stages = try? JSONDecoder().decode([String].self, from: data) else {
            return []
        }
        return stages
    }

    var progressPercentage: Double {
        let totalStages = 9.0 // 5 perspectives + 4 generation stages
        return (Double(stagesCompletedArray.count) / totalStages) * 100
    }
}

enum DigSessionStatus: String, Codable {
    case pending
    case analyzing
    case completed
    case failed
}

enum DigRecommendation: String, Codable {
    case proceed
    case pivot
    case refine
    case abandon

    var displayName: String {
        switch self {
        case .proceed: return "Proceed"
        case .pivot: return "Pivot"
        case .refine: return "Refine"
        case .abandon: return "Abandon"
        }
    }

    var color: Color {
        switch self {
        case .proceed: return .green
        case .pivot: return .orange
        case .refine: return .yellow
        case .abandon: return .red
        }
    }

    var icon: String {
        switch self {
        case .proceed: return "checkmark.circle.fill"
        case .pivot: return "arrow.triangle.turn.up.right.circle.fill"
        case .refine: return "slider.horizontal.3"
        case .abandon: return "xmark.circle.fill"
        }
    }
}

enum DigPerspective: String, Codable, CaseIterable {
    case firstPrinciples = "first_principles"
    case designer
    case marketing
    case technical
    case financial

    var displayName: String {
        switch self {
        case .firstPrinciples: return "First Principles"
        case .designer: return "Designer"
        case .marketing: return "Marketing"
        case .technical: return "Technical"
        case .financial: return "Financial"
        }
    }

    var icon: String {
        switch self {
        case .firstPrinciples: return "atom"
        case .designer: return "paintbrush.fill"
        case .marketing: return "megaphone.fill"
        case .technical: return "gearshape.2.fill"
        case .financial: return "chart.line.uptrend.xyaxis"
        }
    }

    var color: Color {
        switch self {
        case .firstPrinciples: return .purple
        case .designer: return .pink
        case .marketing: return .orange
        case .technical: return .blue
        case .financial: return .green
        }
    }
}

struct DigAnalysis: Codable, Identifiable {
    let id: Int64?
    let sessionId: Int64?
    let perspective: DigPerspective
    var summary: String?
    var strengths: String?
    var weaknesses: String?
    var opportunities: String?
    var threats: String?
    var recommendations: String?
    var score: Double?
    var detailedAnalysis: String?
    let createdAt: Date?

    enum CodingKeys: String, CodingKey {
        case id, perspective, summary, strengths, weaknesses, opportunities, threats, recommendations, score
        case sessionId = "session_id"
        case detailedAnalysis = "detailed_analysis"
        case createdAt = "created_at"
    }

    func parseJSONArray(_ json: String?) -> [String] {
        guard let json = json,
              let data = json.data(using: .utf8),
              let array = try? JSONDecoder().decode([String].self, from: data) else {
            return []
        }
        return array
    }

    var strengthsArray: [String] { parseJSONArray(strengths) }
    var weaknessesArray: [String] { parseJSONArray(weaknesses) }
    var opportunitiesArray: [String] { parseJSONArray(opportunities) }
    var threatsArray: [String] { parseJSONArray(threats) }
    var recommendationsArray: [String] { parseJSONArray(recommendations) }
}

struct DigExecutionPlan: Codable, Identifiable {
    let id: Int64?
    let sessionId: Int64?
    var summary: String?
    var mvpScope: String?
    var fullScope: String?
    var architecture: String?
    var techStack: String?
    var integrations: String?
    var phases: String?
    var milestones: String?
    var teamRequirements: String?
    var estimatedCost: String?
    var llmPrompt: String?
    let createdAt: Date?

    enum CodingKeys: String, CodingKey {
        case id, summary, architecture, integrations, phases, milestones
        case sessionId = "session_id"
        case mvpScope = "mvp_scope"
        case fullScope = "full_scope"
        case techStack = "tech_stack"
        case teamRequirements = "team_requirements"
        case estimatedCost = "estimated_cost"
        case llmPrompt = "llm_prompt"
        case createdAt = "created_at"
    }
}

struct DigMarketing: Codable, Identifiable {
    let id: Int64?
    let sessionId: Int64?
    var valueProposition: String?
    var targetAudience: String?
    var positioningStatement: String?
    var taglines: String?
    var elevatorPitch: String?
    var headlines: String?
    var descriptions: String?
    var colorSuggestions: String?
    var imageryPrompts: String?
    var generatedImages: String?
    var channelStrategy: String?
    var launchStrategy: String?
    let createdAt: Date?

    enum CodingKeys: String, CodingKey {
        case id, taglines, headlines, descriptions
        case sessionId = "session_id"
        case valueProposition = "value_proposition"
        case targetAudience = "target_audience"
        case positioningStatement = "positioning_statement"
        case elevatorPitch = "elevator_pitch"
        case colorSuggestions = "color_suggestions"
        case imageryPrompts = "imagery_prompts"
        case generatedImages = "generated_images"
        case channelStrategy = "channel_strategy"
        case launchStrategy = "launch_strategy"
        case createdAt = "created_at"
    }

    var taglinesArray: [String] {
        guard let taglines = taglines,
              let data = taglines.data(using: .utf8),
              let array = try? JSONDecoder().decode([String].self, from: data) else {
            return []
        }
        return array
    }

    var headlinesArray: [String] {
        guard let headlines = headlines,
              let data = headlines.data(using: .utf8),
              let array = try? JSONDecoder().decode([String].self, from: data) else {
            return []
        }
        return array
    }
}

struct DigRevenueProjection: Codable, Identifiable {
    let id: Int64?
    let sessionId: Int64?
    var marketSize: String?
    var marketGrowth: String?
    var competitors: String?
    var pricingModels: String?
    var recommendedPrice: String?
    var pricingRationale: String?
    var year1Revenue: String?
    var year2Revenue: String?
    var year3Revenue: String?
    var keyMetrics: String?
    var breakEvenAnalysis: String?
    var assumptions: String?
    var risks: String?
    let createdAt: Date?

    enum CodingKeys: String, CodingKey {
        case id, competitors, assumptions, risks
        case sessionId = "session_id"
        case marketSize = "market_size"
        case marketGrowth = "market_growth"
        case pricingModels = "pricing_models"
        case recommendedPrice = "recommended_price"
        case pricingRationale = "pricing_rationale"
        case year1Revenue = "year1_revenue"
        case year2Revenue = "year2_revenue"
        case year3Revenue = "year3_revenue"
        case keyMetrics = "key_metrics"
        case breakEvenAnalysis = "break_even_analysis"
        case createdAt = "created_at"
    }
}

struct DigAlternative: Codable, Identifiable {
    let id: Int64?
    let sessionId: Int64?
    let alternativeIdea: String
    var rationale: String?
    var comparison: String?
    var viabilityScore: Double?
    var effortLevel: String?
    var potential: String?
    let createdAt: Date?

    enum CodingKeys: String, CodingKey {
        case id, rationale, comparison, potential
        case sessionId = "session_id"
        case alternativeIdea = "alternative_idea"
        case viabilityScore = "viability_score"
        case effortLevel = "effort_level"
        case createdAt = "created_at"
    }

    var potentialColor: Color {
        switch potential?.lowercased() {
        case "very_high": return .green
        case "high": return .blue
        case "medium": return .yellow
        default: return .gray
        }
    }
}

struct DigProgressResponse: Codable {
    let sessionId: Int64
    let status: DigSessionStatus
    let currentStage: String
    let stagesCompleted: [String]
    let progress: Int
    let message: String

    enum CodingKeys: String, CodingKey {
        case status, progress, message
        case sessionId = "session_id"
        case currentStage = "current_stage"
        case stagesCompleted = "stages_completed"
    }
}

// MARK: - Keychain Service for API Keys

class KeychainService {
    static let shared = KeychainService()

    private let openAIKeyName = "com.ung.openai.apikey"
    private let claudeKeyName = "com.ung.claude.apikey"

    private init() {}

    func saveOpenAIKey(_ key: String) -> Bool {
        return save(key: key, account: openAIKeyName)
    }

    func getOpenAIKey() -> String? {
        return get(account: openAIKeyName)
    }

    func saveClaudeKey(_ key: String) -> Bool {
        return save(key: key, account: claudeKeyName)
    }

    func getClaudeKey() -> String? {
        return get(account: claudeKeyName)
    }

    func deleteOpenAIKey() -> Bool {
        return delete(account: openAIKeyName)
    }

    func deleteClaudeKey() -> Bool {
        return delete(account: claudeKeyName)
    }

    private func save(key: String, account: String) -> Bool {
        // Delete existing key first
        _ = delete(account: account)

        guard let data = key.data(using: .utf8) else { return false }

        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: account,
            kSecValueData as String: data,
            kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlocked
        ]

        let status = SecItemAdd(query as CFDictionary, nil)
        return status == errSecSuccess
    }

    private func get(account: String) -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: account,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)

        guard status == errSecSuccess,
              let data = result as? Data,
              let key = String(data: data, encoding: .utf8) else {
            return nil
        }

        return key
    }

    private func delete(account: String) -> Bool {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: account
        ]

        let status = SecItemDelete(query as CFDictionary)
        return status == errSecSuccess || status == errSecItemNotFound
    }
}

// MARK: - Dig State

@MainActor
class DigState: ObservableObject {
    @Published var sessions: [DigSession] = []
    @Published var currentSession: DigSession?
    @Published var isLoading = false
    @Published var isAnalyzing = false
    @Published var error: String?
    @Published var progress: DigProgressResponse?

    // API Key settings
    @Published var hasOpenAIKey: Bool = false
    @Published var hasClaudeKey: Bool = false
    @Published var preferredProvider: AIProvider = .openai

    private let service = DigAPIService()
    private var progressTimer: Timer?

    enum AIProvider: String, CaseIterable {
        case openai = "OpenAI"
        case claude = "Claude"
    }

    init() {
        checkAPIKeys()
    }

    func checkAPIKeys() {
        hasOpenAIKey = KeychainService.shared.getOpenAIKey() != nil
        hasClaudeKey = KeychainService.shared.getClaudeKey() != nil
    }

    func saveOpenAIKey(_ key: String) {
        if KeychainService.shared.saveOpenAIKey(key) {
            hasOpenAIKey = true
        }
    }

    func saveClaudeKey(_ key: String) {
        if KeychainService.shared.saveClaudeKey(key) {
            hasClaudeKey = true
        }
    }

    func deleteOpenAIKey() {
        if KeychainService.shared.deleteOpenAIKey() {
            hasOpenAIKey = false
        }
    }

    func deleteClaudeKey() {
        if KeychainService.shared.deleteClaudeKey() {
            hasClaudeKey = false
        }
    }

    // MARK: - API Operations

    func loadSessions(appState: AppState) async {
        isLoading = true
        error = nil

        do {
            sessions = try await service.listSessions(appState: appState)
        } catch {
            self.error = error.localizedDescription
        }

        isLoading = false
    }

    func startAnalysis(idea: String, appState: AppState) async {
        isAnalyzing = true
        error = nil

        do {
            let session = try await service.startSession(idea: idea, appState: appState)
            currentSession = session
            sessions.insert(session, at: 0)

            // Start polling for progress
            startProgressPolling(sessionId: session.id!, appState: appState)
        } catch {
            self.error = error.localizedDescription
            isAnalyzing = false
        }
    }

    func loadSession(id: Int64, appState: AppState) async {
        isLoading = true
        error = nil

        do {
            currentSession = try await service.getSession(id: id, appState: appState)
        } catch {
            self.error = error.localizedDescription
        }

        isLoading = false
    }

    func deleteSession(id: Int64, appState: AppState) async {
        do {
            try await service.deleteSession(id: id, appState: appState)
            sessions.removeAll { $0.id == id }
            if currentSession?.id == id {
                currentSession = nil
            }
        } catch {
            self.error = error.localizedDescription
        }
    }

    func generateImages(appState: AppState) async {
        guard let sessionId = currentSession?.id else { return }

        do {
            let images = try await service.generateImages(sessionId: sessionId, appState: appState)
            // Update current session's marketing images
            if var marketing = currentSession?.marketing {
                marketing.generatedImages = try? String(data: JSONEncoder().encode(images), encoding: .utf8)
                currentSession?.marketing = marketing
            }
        } catch {
            self.error = error.localizedDescription
        }
    }

    func exportSession(format: String, appState: AppState) async -> Data? {
        guard let sessionId = currentSession?.id else { return nil }

        do {
            return try await service.exportSession(id: sessionId, format: format, appState: appState)
        } catch {
            self.error = error.localizedDescription
            return nil
        }
    }

    // MARK: - Progress Polling

    private func startProgressPolling(sessionId: Int64, appState: AppState) {
        progressTimer?.invalidate()

        progressTimer = Timer.scheduledTimer(withTimeInterval: 2.0, repeats: true) { [weak self] timer in
            Task { @MainActor [weak self] in
                guard let self = self else {
                    timer.invalidate()
                    return
                }

                do {
                    let progress = try await self.service.getProgress(sessionId: sessionId, appState: appState)
                    self.progress = progress

                    if progress.status == .completed || progress.status == .failed {
                        timer.invalidate()
                        self.isAnalyzing = false

                        // Load full session data
                        await self.loadSession(id: sessionId, appState: appState)
                    }
                } catch {
                    // Continue polling on error
                }
            }
        }
    }

    func stopProgressPolling() {
        progressTimer?.invalidate()
        progressTimer = nil
    }
}

// MARK: - Dig API Service

actor DigAPIService {
    private let baseURL: String

    init(baseURL: String = "http://localhost:8080/api/v1") {
        self.baseURL = baseURL
    }

    func listSessions(appState: AppState) async throws -> [DigSession] {
        let url = URL(string: "\(baseURL)/dig")!
        var request = URLRequest(url: url)
        request.setValue("Bearer \(appState.authToken ?? "")", forHTTPHeaderField: "Authorization")

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 else {
            return []
        }

        let result = try JSONDecoder.digDecoder.decode(APIResponse<[DigSession]>.self, from: data)
        return result.data ?? []
    }

    func startSession(idea: String, appState: AppState) async throws -> DigSession {
        let url = URL(string: "\(baseURL)/dig")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("Bearer \(appState.authToken ?? "")", forHTTPHeaderField: "Authorization")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        let body = ["idea": idea]
        request.httpBody = try JSONEncoder().encode(body)

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 || httpResponse.statusCode == 201 else {
            throw DigError.startFailed
        }

        let result = try JSONDecoder.digDecoder.decode(APIResponse<DigSession>.self, from: data)
        guard let session = result.data else {
            throw DigError.invalidResponse
        }
        return session
    }

    func getSession(id: Int64, appState: AppState) async throws -> DigSession {
        let url = URL(string: "\(baseURL)/dig/\(id)")!
        var request = URLRequest(url: url)
        request.setValue("Bearer \(appState.authToken ?? "")", forHTTPHeaderField: "Authorization")

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 else {
            throw DigError.notFound
        }

        let result = try JSONDecoder.digDecoder.decode(APIResponse<DigSession>.self, from: data)
        guard let session = result.data else {
            throw DigError.invalidResponse
        }
        return session
    }

    func getProgress(sessionId: Int64, appState: AppState) async throws -> DigProgressResponse {
        let url = URL(string: "\(baseURL)/dig/\(sessionId)/progress")!
        var request = URLRequest(url: url)
        request.setValue("Bearer \(appState.authToken ?? "")", forHTTPHeaderField: "Authorization")

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 else {
            throw DigError.progressFailed
        }

        let result = try JSONDecoder.digDecoder.decode(APIResponse<DigProgressResponse>.self, from: data)
        guard let progress = result.data else {
            throw DigError.invalidResponse
        }
        return progress
    }

    func deleteSession(id: Int64, appState: AppState) async throws {
        let url = URL(string: "\(baseURL)/dig/\(id)")!
        var request = URLRequest(url: url)
        request.httpMethod = "DELETE"
        request.setValue("Bearer \(appState.authToken ?? "")", forHTTPHeaderField: "Authorization")

        let (_, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 else {
            throw DigError.deleteFailed
        }
    }

    func generateImages(sessionId: Int64, appState: AppState) async throws -> [String] {
        let url = URL(string: "\(baseURL)/dig/\(sessionId)/images")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("Bearer \(appState.authToken ?? "")", forHTTPHeaderField: "Authorization")
        request.timeoutInterval = 120 // Image generation can take time

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 else {
            throw DigError.imageGenerationFailed
        }

        let result = try JSONDecoder.digDecoder.decode(APIResponse<[String]>.self, from: data)
        return result.data ?? []
    }

    func exportSession(id: Int64, format: String, appState: AppState) async throws -> Data {
        let url = URL(string: "\(baseURL)/dig/\(id)/export?format=\(format)")!
        var request = URLRequest(url: url)
        request.setValue("Bearer \(appState.authToken ?? "")", forHTTPHeaderField: "Authorization")

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 else {
            throw DigError.exportFailed
        }

        return data
    }
}

// MARK: - Dig Errors

enum DigError: LocalizedError {
    case invalidResponse
    case notFound
    case startFailed
    case progressFailed
    case deleteFailed
    case imageGenerationFailed
    case exportFailed

    var errorDescription: String? {
        switch self {
        case .invalidResponse: return "Invalid response from server"
        case .notFound: return "Session not found"
        case .startFailed: return "Failed to start analysis"
        case .progressFailed: return "Failed to get progress"
        case .deleteFailed: return "Failed to delete session"
        case .imageGenerationFailed: return "Failed to generate images"
        case .exportFailed: return "Failed to export session"
        }
    }
}

// MARK: - JSON Decoder Extension

extension JSONDecoder {
    static var digDecoder: JSONDecoder {
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .custom { decoder in
            let container = try decoder.singleValueContainer()
            let dateString = try container.decode(String.self)

            let iso8601Formatter = ISO8601DateFormatter()
            iso8601Formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
            if let date = iso8601Formatter.date(from: dateString) {
                return date
            }

            iso8601Formatter.formatOptions = [.withInternetDateTime]
            if let date = iso8601Formatter.date(from: dateString) {
                return date
            }

            let dateFormatter = DateFormatter()
            dateFormatter.locale = Locale(identifier: "en_US_POSIX")

            for format in ["yyyy-MM-dd'T'HH:mm:ssZ", "yyyy-MM-dd'T'HH:mm:ss", "yyyy-MM-dd HH:mm:ss", "yyyy-MM-dd"] {
                dateFormatter.dateFormat = format
                if let date = dateFormatter.date(from: dateString) {
                    return date
                }
            }

            throw DecodingError.dataCorruptedError(in: container, debugDescription: "Cannot decode date: \(dateString)")
        }
        return decoder
    }
}
