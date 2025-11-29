//
//  HunterService.swift
//  ung
//
//  API service for Job Hunter feature
//

import Foundation
import SwiftUI

// MARK: - Hunter State

@MainActor
class HunterState: ObservableObject {
    @Published var profile: HunterProfile?
    @Published var jobs: [HunterJob] = []
    @Published var applications: [HunterApplication] = []

    @Published var isLoading = false
    @Published var isHunting = false
    @Published var error: String?

    // Filter state
    @Published var searchQuery = ""
    @Published var selectedSource: JobSource?
    @Published var remoteOnly = false
    @Published var sortBy: JobSortBy = .matchScore

    // Generated proposal
    @Published var generatedProposal: String?

    private let service = HunterAPIService()

    var filteredJobs: [HunterJob] {
        var result = jobs

        // Search filter
        if !searchQuery.isEmpty {
            result = result.filter { job in
                job.title.localizedCaseInsensitiveContains(searchQuery) ||
                (job.company?.localizedCaseInsensitiveContains(searchQuery) ?? false) ||
                (job.description?.localizedCaseInsensitiveContains(searchQuery) ?? false)
            }
        }

        // Source filter
        if let source = selectedSource {
            result = result.filter { $0.jobSource == source }
        }

        // Remote filter
        if remoteOnly {
            result = result.filter { $0.remote }
        }

        // Sort
        switch sortBy {
        case .matchScore:
            result.sort { ($0.matchScore ?? 0) > ($1.matchScore ?? 0) }
        case .postedDate:
            result.sort { ($0.postedAt ?? .distantPast) > ($1.postedAt ?? .distantPast) }
        case .company:
            result.sort { ($0.company ?? "") < ($1.company ?? "") }
        }

        return result
    }

    // MARK: - Load Data

    func loadData(appState: AppState) async {
        isLoading = true
        error = nil

        do {
            async let profileTask = service.getProfile(appState: appState)
            async let jobsTask = service.getJobs(appState: appState)
            async let applicationsTask = service.getApplications(appState: appState)

            let (fetchedProfile, fetchedJobs, fetchedApplications) = await (
                try profileTask,
                try jobsTask,
                try applicationsTask
            )

            self.profile = fetchedProfile
            self.jobs = fetchedJobs
            self.applications = fetchedApplications
        } catch {
            self.error = error.localizedDescription
        }

        isLoading = false
    }

    // MARK: - Import CV

    func importCV(from url: URL, appState: AppState) async throws {
        isLoading = true
        error = nil

        defer { isLoading = false }

        let fileData = try Data(contentsOf: url)
        profile = try await service.importProfile(pdfData: fileData, filename: url.lastPathComponent, appState: appState)
    }

    // MARK: - Update Profile

    func updateProfile(
        name: String,
        title: String,
        bio: String,
        experience: Int,
        rate: Double,
        location: String,
        remote: Bool,
        skills: [String],
        appState: AppState
    ) async {
        isLoading = true
        error = nil

        do {
            profile = try await service.updateProfile(
                name: name,
                title: title,
                bio: bio,
                experience: experience,
                rate: rate,
                location: location,
                remote: remote,
                skills: skills,
                appState: appState
            )
        } catch {
            self.error = error.localizedDescription
        }

        isLoading = false
    }

    // MARK: - Hunt Jobs

    func hunt(appState: AppState) async {
        isHunting = true
        error = nil

        do {
            jobs = try await service.huntJobs(appState: appState)
        } catch {
            self.error = error.localizedDescription
        }

        isHunting = false
    }

    // MARK: - Generate Proposal

    func generateProposal(for job: HunterJob, appState: AppState) async {
        guard let jobId = job.id else { return }

        do {
            generatedProposal = try await service.generateProposal(jobId: jobId, appState: appState)
        } catch {
            self.error = error.localizedDescription
        }
    }

    // MARK: - Submit Application

    func submitApplication(jobId: Int64, proposal: String, appState: AppState) async {
        do {
            let application = try await service.createApplication(
                jobId: jobId,
                proposal: proposal,
                appState: appState
            )
            applications.append(application)
        } catch {
            self.error = error.localizedDescription
        }
    }

    // MARK: - Update Application Status

    func updateApplicationStatus(applicationId: Int64, status: String) async {
        // Find and update locally first
        if let index = applications.firstIndex(where: { $0.id == applicationId }) {
            applications[index].status = status
        }

        // TODO: Sync with API
    }
}

// MARK: - Hunter API Service

actor HunterAPIService {
    private let baseURL: String

    init(baseURL: String = "http://localhost:8080/api/v1") {
        self.baseURL = baseURL
    }

    // MARK: - Profile

    func getProfile(appState: AppState) async throws -> HunterProfile? {
        let url = URL(string: "\(baseURL)/hunter/profile")!
        var request = URLRequest(url: url)
        request.setValue("Bearer \(appState.authToken ?? "")", forHTTPHeaderField: "Authorization")

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw HunterError.invalidResponse
        }

        if httpResponse.statusCode == 404 {
            return nil
        }

        guard httpResponse.statusCode == 200 else {
            throw HunterError.serverError(httpResponse.statusCode)
        }

        let result = try JSONDecoder.hunterDecoder.decode(APIResponse<HunterProfile>.self, from: data)
        return result.data
    }

    func importProfile(pdfData: Data, filename: String, appState: AppState) async throws -> HunterProfile {
        let url = URL(string: "\(baseURL)/hunter/profile/import")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("Bearer \(appState.authToken ?? "")", forHTTPHeaderField: "Authorization")

        let boundary = UUID().uuidString
        request.setValue("multipart/form-data; boundary=\(boundary)", forHTTPHeaderField: "Content-Type")

        var body = Data()
        body.append("--\(boundary)\r\n".data(using: .utf8)!)
        body.append("Content-Disposition: form-data; name=\"file\"; filename=\"\(filename)\"\r\n".data(using: .utf8)!)
        body.append("Content-Type: application/pdf\r\n\r\n".data(using: .utf8)!)
        body.append(pdfData)
        body.append("\r\n--\(boundary)--\r\n".data(using: .utf8)!)

        request.httpBody = body

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 || httpResponse.statusCode == 201 else {
            throw HunterError.importFailed
        }

        let result = try JSONDecoder.hunterDecoder.decode(APIResponse<HunterProfile>.self, from: data)
        guard let profile = result.data else {
            throw HunterError.invalidResponse
        }
        return profile
    }

    func updateProfile(
        name: String,
        title: String,
        bio: String,
        experience: Int,
        rate: Double,
        location: String,
        remote: Bool,
        skills: [String],
        appState: AppState
    ) async throws -> HunterProfile {
        let url = URL(string: "\(baseURL)/hunter/profile")!
        var request = URLRequest(url: url)
        request.httpMethod = "PUT"
        request.setValue("Bearer \(appState.authToken ?? "")", forHTTPHeaderField: "Authorization")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        let skillsJSON = try JSONEncoder().encode(skills)
        let updateData: [String: Any] = [
            "name": name,
            "title": title,
            "bio": bio,
            "experience": experience,
            "rate": rate,
            "location": location,
            "remote": remote,
            "skills": String(data: skillsJSON, encoding: .utf8) ?? "[]"
        ]

        request.httpBody = try JSONSerialization.data(withJSONObject: updateData)

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 else {
            throw HunterError.updateFailed
        }

        let result = try JSONDecoder.hunterDecoder.decode(APIResponse<HunterProfile>.self, from: data)
        guard let profile = result.data else {
            throw HunterError.invalidResponse
        }
        return profile
    }

    // MARK: - Jobs

    func getJobs(appState: AppState) async throws -> [HunterJob] {
        let url = URL(string: "\(baseURL)/hunter/jobs")!
        var request = URLRequest(url: url)
        request.setValue("Bearer \(appState.authToken ?? "")", forHTTPHeaderField: "Authorization")

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 else {
            return []
        }

        let result = try JSONDecoder.hunterDecoder.decode(APIResponse<[HunterJob]>.self, from: data)
        return result.data ?? []
    }

    func huntJobs(appState: AppState) async throws -> [HunterJob] {
        let url = URL(string: "\(baseURL)/hunter/hunt")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("Bearer \(appState.authToken ?? "")", forHTTPHeaderField: "Authorization")
        request.timeoutInterval = 120 // Scraping can take a while

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 else {
            throw HunterError.huntFailed
        }

        let result = try JSONDecoder.hunterDecoder.decode(APIResponse<HuntResult>.self, from: data)
        return result.data?.jobs ?? []
    }

    // MARK: - Applications

    func getApplications(appState: AppState) async throws -> [HunterApplication] {
        let url = URL(string: "\(baseURL)/hunter/applications")!
        var request = URLRequest(url: url)
        request.setValue("Bearer \(appState.authToken ?? "")", forHTTPHeaderField: "Authorization")

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 else {
            return []
        }

        let result = try JSONDecoder.hunterDecoder.decode(APIResponse<[HunterApplication]>.self, from: data)
        return result.data ?? []
    }

    func createApplication(jobId: Int64, proposal: String, appState: AppState) async throws -> HunterApplication {
        let url = URL(string: "\(baseURL)/hunter/applications")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("Bearer \(appState.authToken ?? "")", forHTTPHeaderField: "Authorization")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        let applicationData: [String: Any] = [
            "job_id": jobId,
            "proposal": proposal,
            "generate_proposal": false
        ]

        request.httpBody = try JSONSerialization.data(withJSONObject: applicationData)

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 || httpResponse.statusCode == 201 else {
            throw HunterError.applicationFailed
        }

        let result = try JSONDecoder.hunterDecoder.decode(APIResponse<HunterApplication>.self, from: data)
        guard let application = result.data else {
            throw HunterError.invalidResponse
        }
        return application
    }

    func generateProposal(jobId: Int64, appState: AppState) async throws -> String {
        let url = URL(string: "\(baseURL)/hunter/applications")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("Bearer \(appState.authToken ?? "")", forHTTPHeaderField: "Authorization")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        let applicationData: [String: Any] = [
            "job_id": jobId,
            "generate_proposal": true
        ]

        request.httpBody = try JSONSerialization.data(withJSONObject: applicationData)

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 || httpResponse.statusCode == 201 else {
            throw HunterError.proposalGenerationFailed
        }

        let result = try JSONDecoder.hunterDecoder.decode(APIResponse<HunterApplication>.self, from: data)
        return result.data?.proposal ?? ""
    }
}

// MARK: - API Response

struct APIResponse<T: Decodable>: Decodable {
    let success: Bool
    let data: T?
    let message: String?
    let error: String?
}

struct HuntResult: Decodable {
    let jobs: [HunterJob]
    let newCount: Int
    let totalCount: Int

    enum CodingKeys: String, CodingKey {
        case jobs
        case newCount = "new_count"
        case totalCount = "total_count"
    }
}

// MARK: - Hunter Errors

enum HunterError: LocalizedError {
    case invalidResponse
    case serverError(Int)
    case importFailed
    case updateFailed
    case huntFailed
    case applicationFailed
    case proposalGenerationFailed

    var errorDescription: String? {
        switch self {
        case .invalidResponse:
            return "Invalid response from server"
        case .serverError(let code):
            return "Server error: \(code)"
        case .importFailed:
            return "Failed to import CV/Resume"
        case .updateFailed:
            return "Failed to update profile"
        case .huntFailed:
            return "Failed to hunt for jobs"
        case .applicationFailed:
            return "Failed to create application"
        case .proposalGenerationFailed:
            return "Failed to generate proposal"
        }
    }
}

// MARK: - JSON Decoder Extension

extension JSONDecoder {
    static var hunterDecoder: JSONDecoder {
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .custom { decoder in
            let container = try decoder.singleValueContainer()
            let dateString = try container.decode(String.self)

            // Try ISO8601 with fractional seconds
            let iso8601Formatter = ISO8601DateFormatter()
            iso8601Formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
            if let date = iso8601Formatter.date(from: dateString) {
                return date
            }

            // Try ISO8601 without fractional seconds
            iso8601Formatter.formatOptions = [.withInternetDateTime]
            if let date = iso8601Formatter.date(from: dateString) {
                return date
            }

            // Try common date formats
            let dateFormatter = DateFormatter()
            dateFormatter.locale = Locale(identifier: "en_US_POSIX")

            let formats = [
                "yyyy-MM-dd'T'HH:mm:ssZ",
                "yyyy-MM-dd'T'HH:mm:ss",
                "yyyy-MM-dd HH:mm:ss",
                "yyyy-MM-dd"
            ]

            for format in formats {
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

// MARK: - AppState Extension for Auth Token

extension AppState {
    var authToken: String? {
        // Return the stored auth token if available
        // For now, return nil - implement when authentication is added
        return UserDefaults.standard.string(forKey: "authToken")
    }
}
