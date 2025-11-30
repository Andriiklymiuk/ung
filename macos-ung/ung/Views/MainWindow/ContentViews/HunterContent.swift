//
//  HunterContent.swift
//  ung
//
//  Job Hunter feature - find and apply to freelance jobs
//

import SwiftUI
import UniformTypeIdentifiers

struct HunterContent: View {
    @EnvironmentObject var appState: AppState
    @StateObject private var hunterState = HunterState()
    @State private var selectedTab: HunterTab = .jobs
    @State private var showImportCV = false
    @State private var showEditProfile = false

    private let compactWidthThreshold: CGFloat = 800

    var body: some View {
        GeometryReader { geometry in
            let isCompact = geometry.size.width < compactWidthThreshold

            VStack(spacing: 0) {
                // Tab bar
                hunterTabBar

                Divider()

                // Content based on selected tab
                ScrollView(showsIndicators: false) {
                    switch selectedTab {
                    case .jobs:
                        jobsContent(isCompact: isCompact)
                    case .applications:
                        applicationsContent(isCompact: isCompact)
                    case .profile:
                        profileContent(isCompact: isCompact)
                    }
                }
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        }
        .onAppear {
            Task {
                await hunterState.loadData(appState: appState)
                // Auto-prompt CV import if no profile
                if hunterState.profile == nil || hunterState.profile?.skillsList.isEmpty == true {
                    showImportCV = true
                }
            }
        }
        .sheet(isPresented: $showImportCV) {
            ImportCVSheet(hunterState: hunterState, appState: appState, autoHuntAfterImport: hunterState.profile == nil)
        }
        .sheet(isPresented: $showEditProfile) {
            EditProfileSheet(hunterState: hunterState, appState: appState)
        }
    }

    // MARK: - Tab Bar

    private var hunterTabBar: some View {
        HStack(spacing: 0) {
            ForEach(HunterTab.allCases, id: \.self) { tab in
                Button(action: { selectedTab = tab }) {
                    HStack(spacing: 8) {
                        Image(systemName: tab.icon)
                            .font(.system(size: 14))
                        Text(tab.title)
                            .font(.system(size: 14, weight: .medium))
                        if tab == .jobs {
                            Text("\(hunterState.jobs.count)")
                                .font(.system(size: 12, weight: .semibold))
                                .padding(.horizontal, 6)
                                .padding(.vertical, 2)
                                .background(Color.blue.opacity(0.2))
                                .foregroundColor(.blue)
                                .clipShape(Capsule())
                        } else if tab == .applications {
                            Text("\(hunterState.applications.count)")
                                .font(.system(size: 12, weight: .semibold))
                                .padding(.horizontal, 6)
                                .padding(.vertical, 2)
                                .background(Color.green.opacity(0.2))
                                .foregroundColor(.green)
                                .clipShape(Capsule())
                        }
                    }
                    .foregroundColor(selectedTab == tab ? .primary : .secondary)
                    .padding(.horizontal, 16)
                    .padding(.vertical, 10)
                    .background(
                        selectedTab == tab
                            ? Color.accentColor.opacity(0.1)
                            : Color.clear
                    )
                    .clipShape(RoundedRectangle(cornerRadius: 8))
                }
                .buttonStyle(.plain)
            }

            Spacer()

            // Hunt button - shows import CV if no profile
            Button(action: {
                if hunterState.profile == nil || hunterState.profile?.skillsList.isEmpty == true {
                    showImportCV = true
                } else {
                    Task { await hunterState.hunt(appState: appState) }
                }
            }) {
                HStack(spacing: 6) {
                    if hunterState.isHunting {
                        ProgressView()
                            .scaleEffect(0.7)
                    } else {
                        Image(systemName: "magnifyingglass")
                    }
                    Text("Hunt Jobs")
                        .font(.system(size: 14, weight: .semibold))
                }
                .foregroundColor(.white)
                .padding(.horizontal, 16)
                .padding(.vertical, 8)
                .background(
                    LinearGradient(
                        colors: [.blue, .purple],
                        startPoint: .leading,
                        endPoint: .trailing
                    )
                )
                .clipShape(RoundedRectangle(cornerRadius: 8))
            }
            .buttonStyle(.plain)
            .disabled(hunterState.isHunting)
        }
        .padding(.horizontal, 20)
        .padding(.vertical, 12)
    }

    // MARK: - Jobs Content

    private func jobsContent(isCompact: Bool) -> some View {
        VStack(spacing: 16) {
            if hunterState.profile == nil {
                noProfileBanner
            }

            if hunterState.jobs.isEmpty {
                emptyJobsState
            } else {
                // Filter bar
                jobsFilterBar

                // Jobs grid/list
                LazyVStack(spacing: 12) {
                    ForEach(hunterState.filteredJobs) { job in
                        JobCard(job: job, hunterState: hunterState, appState: appState)
                    }
                }
            }
        }
        .padding(20)
    }

    private var noProfileBanner: some View {
        HStack(spacing: 12) {
            Image(systemName: "doc.badge.plus")
                .font(.system(size: 20))
                .foregroundColor(.orange)

            Text("Upload CV to find matching jobs")
                .font(.system(size: 14))

            Spacer()

            Button("Import CV") {
                showImportCV = true
            }
            .buttonStyle(.borderedProminent)
            .controlSize(.small)
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 8)
                .fill(Color.orange.opacity(0.08))
        )
    }

    private var emptyJobsState: some View {
        VStack(spacing: 12) {
            Image(systemName: "briefcase")
                .font(.system(size: 36))
                .foregroundColor(.secondary)

            Text("No jobs yet")
                .font(.system(size: 16, weight: .medium))

            Text("Click Hunt Jobs to start")
                .font(.system(size: 13))
                .foregroundColor(.secondary)
        }
        .frame(maxWidth: .infinity)
        .padding(32)
    }

    private var jobsFilterBar: some View {
        VStack(spacing: 12) {
            HStack(spacing: 12) {
                // Search
                HStack(spacing: 8) {
                    Image(systemName: "magnifyingglass")
                        .foregroundColor(.secondary)
                    TextField("Search jobs...", text: $hunterState.searchQuery)
                        .textFieldStyle(.plain)
                }
                .padding(8)
                .background(Color.gray.opacity(0.1))
                .clipShape(RoundedRectangle(cornerRadius: 8))
                .frame(maxWidth: 300)

                // Region filter
                Picker("Region", selection: $hunterState.selectedRegion) {
                    Text("All Regions").tag(nil as JobRegion?)
                    ForEach(JobRegion.allCases, id: \.self) { region in
                        HStack {
                            Text(region.flag)
                            Text(region.displayName)
                        }.tag(region as JobRegion?)
                    }
                }
                .pickerStyle(.menu)
                .frame(width: 140)

                // Source filter
                Picker("Source", selection: $hunterState.selectedSource) {
                    Text("All Sources").tag(nil as JobSource?)
                    ForEach(JobSource.allCases, id: \.self) { source in
                        HStack {
                            Image(systemName: source.iconName)
                            Text(source.displayName)
                        }.tag(source as JobSource?)
                    }
                }
                .pickerStyle(.menu)
                .frame(width: 170)

                // Remote filter
                Toggle("Remote only", isOn: $hunterState.remoteOnly)

                Spacer()

                // Sort
                Picker("Sort", selection: $hunterState.sortBy) {
                    Text("Match Score").tag(JobSortBy.matchScore)
                    Text("Posted Date").tag(JobSortBy.postedDate)
                    Text("Company").tag(JobSortBy.company)
                }
                .pickerStyle(.menu)
                .frame(width: 130)
            }

            // Quick region filters
            HStack(spacing: 8) {
                Text("Quick filter:")
                    .font(.system(size: 12))
                    .foregroundColor(.secondary)

                ForEach(JobRegion.allCases, id: \.self) { region in
                    Button(action: {
                        hunterState.selectedRegion = hunterState.selectedRegion == region ? nil : region
                    }) {
                        HStack(spacing: 4) {
                            Text(region.flag)
                            Text(region.shortName)
                                .font(.system(size: 12, weight: .medium))
                        }
                        .padding(.horizontal, 10)
                        .padding(.vertical, 5)
                        .background(
                            hunterState.selectedRegion == region
                                ? Color.accentColor.opacity(0.2)
                                : Color.gray.opacity(0.1)
                        )
                        .foregroundColor(
                            hunterState.selectedRegion == region
                                ? .accentColor
                                : .secondary
                        )
                        .clipShape(Capsule())
                    }
                    .buttonStyle(.plain)
                }

                Spacer()

                Text("\(hunterState.filteredJobs.count) jobs")
                    .font(.system(size: 12))
                    .foregroundColor(.secondary)
            }
        }
    }

    // MARK: - Applications Content

    private func applicationsContent(isCompact: Bool) -> some View {
        VStack(spacing: 16) {
            if hunterState.applications.isEmpty {
                emptyApplicationsState
            } else {
                // Stats overview
                applicationsStats

                // Applications list
                LazyVStack(spacing: 12) {
                    ForEach(hunterState.applications) { application in
                        ApplicationCard(application: application, hunterState: hunterState)
                    }
                }
            }
        }
        .padding(20)
    }

    private var emptyApplicationsState: some View {
        VStack(spacing: 12) {
            Image(systemName: "paperplane")
                .font(.system(size: 36))
                .foregroundColor(.secondary)

            Text("No applications yet")
                .font(.system(size: 16, weight: .medium))

            Text("Apply to jobs to track progress")
                .font(.system(size: 13))
                .foregroundColor(.secondary)
        }
        .frame(maxWidth: .infinity)
        .padding(32)
    }

    private var applicationsStats: some View {
        HStack(spacing: 16) {
            ForEach(ApplicationStatus.allCases, id: \.self) { status in
                let count = hunterState.applications.filter { $0.applicationStatus == status }.count
                if count > 0 {
                    HStack(spacing: 6) {
                        Image(systemName: status.iconName)
                            .foregroundColor(statusColor(status))
                        Text("\(count)")
                            .font(.system(size: 14, weight: .semibold))
                        Text(status.displayName)
                            .font(.system(size: 12))
                            .foregroundColor(.secondary)
                    }
                    .padding(.horizontal, 12)
                    .padding(.vertical, 6)
                    .background(statusColor(status).opacity(0.1))
                    .clipShape(Capsule())
                }
            }
            Spacer()
        }
    }

    // MARK: - Profile Content

    private func profileContent(isCompact: Bool) -> some View {
        VStack(spacing: 20) {
            if let profile = hunterState.profile {
                profileCard(profile: profile, isCompact: isCompact)
            } else {
                noProfileState
            }
        }
        .padding(20)
    }

    private func profileCard(profile: HunterProfile, isCompact: Bool) -> some View {
        VStack(spacing: 24) {
            // Header
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    Text(profile.name)
                        .font(.system(size: 24, weight: .bold))
                    if let title = profile.title {
                        Text(title)
                            .font(.system(size: 16))
                            .foregroundColor(.secondary)
                    }
                }

                Spacer()

                Button("Edit Profile") {
                    showEditProfile = true
                }
                .buttonStyle(.bordered)

                Button("Import New CV") {
                    showImportCV = true
                }
                .buttonStyle(.borderedProminent)
            }

            Divider()

            // Info grid
            LazyVGrid(columns: [
                GridItem(.flexible()),
                GridItem(.flexible()),
                GridItem(.flexible())
            ], spacing: 16) {
                if let experience = profile.experience {
                    InfoTile(title: "Experience", value: "\(experience) years", icon: "briefcase")
                }
                if let rate = profile.rate {
                    InfoTile(title: "Hourly Rate", value: "$\(Int(rate))", icon: "dollarsign.circle")
                }
                if let location = profile.location {
                    InfoTile(title: "Location", value: location, icon: "location")
                }
                InfoTile(title: "Remote", value: profile.remote ? "Yes" : "No", icon: "house")
            }

            // Skills
            if !profile.skillsList.isEmpty {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Skills")
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundColor(.secondary)

                    FlowLayout(spacing: 8) {
                        ForEach(profile.skillsList, id: \.self) { skill in
                            Text(skill)
                                .font(.system(size: 13))
                                .padding(.horizontal, 10)
                                .padding(.vertical, 5)
                                .background(Color.blue.opacity(0.1))
                                .foregroundColor(.blue)
                                .clipShape(Capsule())
                        }
                    }
                }
            }

            // Bio
            if let bio = profile.bio, !bio.isEmpty {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Bio")
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundColor(.secondary)
                    Text(bio)
                        .font(.system(size: 14))
                        .foregroundColor(.primary)
                }
            }
        }
        .padding(24)
        .background(
            RoundedRectangle(cornerRadius: 16)
                .fill(Color(.windowBackgroundColor))
                .shadow(color: .black.opacity(0.1), radius: 8, y: 2)
        )
    }

    private var noProfileState: some View {
        VStack(spacing: 16) {
            Image(systemName: "person.crop.circle.badge.plus")
                .font(.system(size: 48))
                .foregroundColor(.accentColor)

            Text("No profile yet")
                .font(.system(size: 18, weight: .semibold))

            Text("Import your CV to extract skills")
                .font(.system(size: 13))
                .foregroundColor(.secondary)

            Button("Import CV") {
                showImportCV = true
            }
            .buttonStyle(.borderedProminent)
        }
        .frame(maxWidth: .infinity)
        .padding(32)
    }

    // MARK: - Helpers

    private func statusColor(_ status: ApplicationStatus) -> Color {
        switch status {
        case .draft: return .gray
        case .applied: return .blue
        case .viewed: return .purple
        case .response: return .cyan
        case .interview: return .orange
        case .offer: return .green
        case .rejected: return .red
        case .withdrawn: return .gray
        }
    }
}

// MARK: - Supporting Views

struct InfoTile: View {
    let title: String
    let value: String
    let icon: String

    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: icon)
                .font(.system(size: 20))
                .foregroundColor(.accentColor)
                .frame(width: 32)

            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                    .font(.system(size: 12))
                    .foregroundColor(.secondary)
                Text(value)
                    .font(.system(size: 14, weight: .medium))
            }

            Spacer()
        }
        .padding(12)
        .background(Color.gray.opacity(0.05))
        .clipShape(RoundedRectangle(cornerRadius: 8))
    }
}

struct JobCard: View {
    let job: HunterJob
    @ObservedObject var hunterState: HunterState
    var appState: AppState
    @State private var showApply = false

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                // Source badge
                HStack(spacing: 4) {
                    Image(systemName: job.jobSource.iconName)
                        .font(.system(size: 10))
                    Text(job.jobSource.displayName)
                        .font(.system(size: 11, weight: .medium))
                }
                .foregroundColor(.secondary)
                .padding(.horizontal, 8)
                .padding(.vertical, 4)
                .background(Color.gray.opacity(0.1))
                .clipShape(Capsule())

                if job.remote {
                    HStack(spacing: 4) {
                        Image(systemName: "house")
                            .font(.system(size: 10))
                        Text("Remote")
                            .font(.system(size: 11, weight: .medium))
                    }
                    .foregroundColor(.green)
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(Color.green.opacity(0.1))
                    .clipShape(Capsule())
                }

                Spacer()

                // Match score
                if let score = job.matchScore, score > 0 {
                    HStack(spacing: 4) {
                        Image(systemName: "star.fill")
                            .font(.system(size: 10))
                        Text("\(Int(score))% match")
                            .font(.system(size: 11, weight: .semibold))
                    }
                    .foregroundColor(scoreColor(score))
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(scoreColor(score).opacity(0.1))
                    .clipShape(Capsule())
                }
            }

            // Title and company
            VStack(alignment: .leading, spacing: 4) {
                Text(job.title)
                    .font(.system(size: 16, weight: .semibold))
                    .lineLimit(2)

                if let company = job.company, !company.isEmpty {
                    Text(company)
                        .font(.system(size: 14))
                        .foregroundColor(.secondary)
                }
            }

            // Rate if available
            if let formattedRate = job.formattedRate {
                Text(formattedRate)
                    .font(.system(size: 13, weight: .medium))
                    .foregroundColor(.green)
            }

            // Skills
            if !job.skillsList.isEmpty {
                FlowLayout(spacing: 6) {
                    ForEach(job.skillsList.prefix(6), id: \.self) { skill in
                        Text(skill)
                            .font(.system(size: 11))
                            .padding(.horizontal, 8)
                            .padding(.vertical, 3)
                            .background(Color.blue.opacity(0.1))
                            .foregroundColor(.blue)
                            .clipShape(Capsule())
                    }
                    if job.skillsList.count > 6 {
                        Text("+\(job.skillsList.count - 6)")
                            .font(.system(size: 11))
                            .foregroundColor(.secondary)
                    }
                }
            }

            Divider()

            // Actions
            HStack {
                if let postedAt = job.postedAt {
                    Text(postedAt.formatted(date: .abbreviated, time: .omitted))
                        .font(.system(size: 12))
                        .foregroundColor(.secondary)
                }

                if let location = job.location, !location.isEmpty {
                    Text("• \(location)")
                        .font(.system(size: 12))
                        .foregroundColor(.secondary)
                }

                Spacer()

                if let sourceUrl = job.sourceUrl, let url = URL(string: sourceUrl) {
                    Link(destination: url) {
                        HStack(spacing: 4) {
                            Text("View")
                            Image(systemName: "arrow.up.right.square")
                        }
                        .font(.system(size: 12))
                    }
                }

                Button("Apply") {
                    showApply = true
                }
                .buttonStyle(.borderedProminent)
                .controlSize(.small)
            }
        }
        .padding(16)
        .background(
            RoundedRectangle(cornerRadius: 12)
                .fill(Color(.windowBackgroundColor))
                .shadow(color: .black.opacity(0.05), radius: 4, y: 2)
        )
        .sheet(isPresented: $showApply) {
            ApplyToJobSheet(job: job, hunterState: hunterState, appState: appState)
        }
    }

    private func scoreColor(_ score: Double) -> Color {
        if score >= 70 { return .green }
        if score >= 40 { return .orange }
        return .gray
    }
}

struct ApplicationCard: View {
    let application: HunterApplication
    @ObservedObject var hunterState: HunterState

    var body: some View {
        HStack(spacing: 16) {
            // Status indicator
            Circle()
                .fill(statusColor(application.applicationStatus))
                .frame(width: 12, height: 12)

            VStack(alignment: .leading, spacing: 4) {
                Text("Application #\(application.id ?? 0)")
                    .font(.system(size: 14, weight: .medium))

                Text(application.applicationStatus.displayName)
                    .font(.system(size: 12))
                    .foregroundColor(.secondary)
            }

            Spacer()

            if let appliedAt = application.appliedAt {
                Text(appliedAt.formatted(date: .abbreviated, time: .omitted))
                    .font(.system(size: 12))
                    .foregroundColor(.secondary)
            }

            // Status picker
            Picker("Status", selection: Binding(
                get: { application.status },
                set: { newStatus in
                    Task {
                        await hunterState.updateApplicationStatus(
                            applicationId: application.id ?? 0,
                            status: newStatus
                        )
                    }
                }
            )) {
                ForEach(ApplicationStatus.allCases, id: \.self) { status in
                    Text(status.displayName).tag(status.rawValue)
                }
            }
            .pickerStyle(.menu)
            .frame(width: 120)
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 8)
                .fill(Color(.windowBackgroundColor))
        )
    }

    private func statusColor(_ status: ApplicationStatus) -> Color {
        switch status {
        case .draft: return .gray
        case .applied: return .blue
        case .viewed: return .purple
        case .response: return .cyan
        case .interview: return .orange
        case .offer: return .green
        case .rejected: return .red
        case .withdrawn: return .gray
        }
    }
}

// MARK: - Import CV Sheet

struct ImportCVSheet: View {
    @ObservedObject var hunterState: HunterState
    var appState: AppState
    @Environment(\.dismiss) var dismiss
    @State private var selectedFileURL: URL?
    @State private var isImporting = false
    @State private var errorMessage: String?
    @State private var importComplete = false
    var autoHuntAfterImport: Bool = true

    var body: some View {
        VStack(spacing: 20) {
            if importComplete {
                // Success state
                VStack(spacing: 12) {
                    Image(systemName: "checkmark.circle.fill")
                        .font(.system(size: 40))
                        .foregroundColor(.green)

                    Text("Done!")
                        .font(.system(size: 18, weight: .semibold))

                    if let profile = hunterState.profile {
                        Text("\(profile.skillsList.count) skills extracted")
                            .font(.system(size: 13))
                            .foregroundColor(.secondary)
                    }

                    if autoHuntAfterImport {
                        HStack(spacing: 8) {
                            ProgressView()
                                .scaleEffect(0.7)
                            Text("Hunting jobs...")
                                .font(.system(size: 13))
                                .foregroundColor(.secondary)
                        }
                    }
                }
                .padding(20)
            } else {
                // Upload state
                Text("Import CV")
                    .font(.system(size: 18, weight: .semibold))

                VStack(spacing: 12) {
                    Image(systemName: "doc.badge.plus")
                        .font(.system(size: 36))
                        .foregroundColor(.accentColor)

                    Button("Choose PDF...") {
                        let panel = NSOpenPanel()
                        panel.allowedContentTypes = [.pdf]
                        panel.allowsMultipleSelection = false
                        if panel.runModal() == .OK {
                            selectedFileURL = panel.url
                        }
                    }
                    .buttonStyle(.bordered)

                    if let url = selectedFileURL {
                        Text(url.lastPathComponent)
                            .font(.system(size: 12))
                            .foregroundColor(.secondary)
                            .lineLimit(1)
                    }

                    if let error = errorMessage {
                        Text(error)
                            .font(.system(size: 12))
                            .foregroundColor(.red)
                    }
                }
                .frame(maxWidth: .infinity)
                .padding(20)
                .background(
                    RoundedRectangle(cornerRadius: 10)
                        .stroke(style: StrokeStyle(lineWidth: 1, dash: [6]))
                        .foregroundColor(.gray.opacity(0.3))
                )

                HStack {
                    Button("Cancel") { dismiss() }
                        .buttonStyle(.bordered)

                    Spacer()

                    Button(isImporting ? "Importing..." : "Import") {
                        guard let url = selectedFileURL else { return }
                        isImporting = true
                        errorMessage = nil

                        Task {
                            do {
                                try await hunterState.importCV(from: url, appState: appState)
                                importComplete = true

                                if autoHuntAfterImport {
                                    try? await Task.sleep(nanoseconds: 400_000_000)
                                    await hunterState.hunt(appState: appState)
                                    dismiss()
                                }
                            } catch {
                                errorMessage = error.localizedDescription
                                isImporting = false
                            }
                        }
                    }
                    .buttonStyle(.borderedProminent)
                    .disabled(selectedFileURL == nil || isImporting)
                }
            }
        }
        .padding(20)
        .frame(width: 360)
    }
}

// MARK: - Edit Profile Sheet

struct EditProfileSheet: View {
    @ObservedObject var hunterState: HunterState
    var appState: AppState
    @Environment(\.dismiss) var dismiss

    @State private var name: String = ""
    @State private var title: String = ""
    @State private var bio: String = ""
    @State private var experience: String = ""
    @State private var rate: String = ""
    @State private var location: String = ""
    @State private var remote: Bool = true
    @State private var skillsText: String = ""

    var body: some View {
        VStack(spacing: 20) {
            Text("Edit Profile")
                .font(.system(size: 20, weight: .bold))

            Form {
                TextField("Name", text: $name)
                TextField("Title", text: $title)
                TextField("Years of Experience", text: $experience)
                TextField("Hourly Rate (USD)", text: $rate)
                TextField("Location", text: $location)
                Toggle("Available for Remote", isOn: $remote)

                VStack(alignment: .leading) {
                    Text("Skills (comma separated)")
                        .font(.caption)
                    TextField("go, swift, typescript...", text: $skillsText)
                }

                VStack(alignment: .leading) {
                    Text("Bio")
                        .font(.caption)
                    TextEditor(text: $bio)
                        .frame(height: 100)
                }
            }

            HStack {
                Button("Cancel") {
                    dismiss()
                }
                .buttonStyle(.bordered)

                Spacer()

                Button("Save") {
                    Task {
                        await hunterState.updateProfile(
                            name: name,
                            title: title,
                            bio: bio,
                            experience: Int(experience) ?? 0,
                            rate: Double(rate) ?? 0,
                            location: location,
                            remote: remote,
                            skills: skillsText.split(separator: ",").map { $0.trimmingCharacters(in: .whitespaces) },
                            appState: appState
                        )
                        dismiss()
                    }
                }
                .buttonStyle(.borderedProminent)
            }
        }
        .padding(24)
        .frame(width: 500, height: 500)
        .onAppear {
            if let profile = hunterState.profile {
                name = profile.name
                title = profile.title ?? ""
                bio = profile.bio ?? ""
                experience = profile.experience.map { String($0) } ?? ""
                rate = profile.rate.map { String(Int($0)) } ?? ""
                location = profile.location ?? ""
                remote = profile.remote
                skillsText = profile.skillsList.joined(separator: ", ")
            }
        }
    }
}

// MARK: - Apply to Job Sheet

struct ApplyToJobSheet: View {
    let job: HunterJob
    @ObservedObject var hunterState: HunterState
    var appState: AppState
    @Environment(\.dismiss) var dismiss

    @State private var proposal: String = ""
    @State private var isGenerating = false
    @State private var isSubmitting = false

    var body: some View {
        VStack(spacing: 20) {
            Text("Apply to \(job.title)")
                .font(.system(size: 18, weight: .bold))

            if let company = job.company {
                Text("at \(company)")
                    .font(.system(size: 14))
                    .foregroundColor(.secondary)
            }

            VStack(alignment: .leading, spacing: 8) {
                HStack {
                    Text("Proposal")
                        .font(.system(size: 14, weight: .medium))

                    Spacer()

                    Button(action: {
                        isGenerating = true
                        Task {
                            await hunterState.generateProposal(for: job, appState: appState)
                            proposal = hunterState.generatedProposal ?? ""
                            isGenerating = false
                        }
                    }) {
                        HStack(spacing: 4) {
                            if isGenerating {
                                ProgressView()
                                    .scaleEffect(0.7)
                            } else {
                                Image(systemName: "sparkles")
                            }
                            Text("Generate with AI")
                                .font(.system(size: 12))
                        }
                    }
                    .buttonStyle(.bordered)
                    .disabled(isGenerating || hunterState.profile == nil)
                }

                TextEditor(text: $proposal)
                    .font(.system(size: 13))
                    .frame(height: 200)
                    .border(Color.gray.opacity(0.3), width: 1)
            }

            HStack {
                Button("Cancel") {
                    dismiss()
                }
                .buttonStyle(.bordered)

                Spacer()

                Button("Submit Application") {
                    isSubmitting = true
                    Task {
                        await hunterState.submitApplication(
                            jobId: job.id ?? 0,
                            proposal: proposal,
                            appState: appState
                        )
                        isSubmitting = false
                        dismiss()
                    }
                }
                .buttonStyle(.borderedProminent)
                .disabled(proposal.isEmpty || isSubmitting)
            }
        }
        .padding(24)
        .frame(width: 550, height: 400)
    }
}

// MARK: - Flow Layout

struct FlowLayout: Layout {
    var spacing: CGFloat = 8

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        let sizes = subviews.map { $0.sizeThatFits(.unspecified) }
        return layout(sizes: sizes, proposal: proposal).size
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        let sizes = subviews.map { $0.sizeThatFits(.unspecified) }
        let offsets = layout(sizes: sizes, proposal: proposal).offsets

        for (offset, subview) in zip(offsets, subviews) {
            subview.place(at: CGPoint(x: bounds.minX + offset.x, y: bounds.minY + offset.y), proposal: .unspecified)
        }
    }

    private func layout(sizes: [CGSize], proposal: ProposedViewSize) -> (offsets: [CGPoint], size: CGSize) {
        let maxWidth = proposal.width ?? .infinity
        var offsets: [CGPoint] = []
        var currentX: CGFloat = 0
        var currentY: CGFloat = 0
        var lineHeight: CGFloat = 0
        var maxX: CGFloat = 0

        for size in sizes {
            if currentX + size.width > maxWidth, currentX > 0 {
                currentX = 0
                currentY += lineHeight + spacing
                lineHeight = 0
            }

            offsets.append(CGPoint(x: currentX, y: currentY))
            lineHeight = max(lineHeight, size.height)
            currentX += size.width + spacing
            maxX = max(maxX, currentX)
        }

        return (offsets, CGSize(width: maxX, height: currentY + lineHeight))
    }
}

// MARK: - Hunter Tab

enum HunterTab: String, CaseIterable {
    case jobs
    case applications
    case profile

    var title: String {
        switch self {
        case .jobs: return "Jobs"
        case .applications: return "Applications"
        case .profile: return "Profile"
        }
    }

    var icon: String {
        switch self {
        case .jobs: return "briefcase"
        case .applications: return "paperplane"
        case .profile: return "person.crop.circle"
        }
    }
}

// MARK: - Sort Options

enum JobSortBy: String, CaseIterable {
    case matchScore
    case postedDate
    case company
}

// MARK: - Job Region

enum JobRegion: String, CaseIterable {
    case global
    case ukraine
    case netherlands
    case europe
    case usa

    var displayName: String {
        switch self {
        case .global: return "Global"
        case .ukraine: return "Ukraine"
        case .netherlands: return "Netherlands"
        case .europe: return "Europe"
        case .usa: return "USA"
        }
    }

    var shortName: String {
        switch self {
        case .global: return "Global"
        case .ukraine: return "UA"
        case .netherlands: return "NL"
        case .europe: return "EU"
        case .usa: return "US"
        }
    }

    var flag: String {
        switch self {
        case .global: return "🌍"
        case .ukraine: return "🇺🇦"
        case .netherlands: return "🇳🇱"
        case .europe: return "🇪🇺"
        case .usa: return "🇺🇸"
        }
    }

    var sources: [JobSource] {
        switch self {
        case .global:
            return JobSource.allCases
        case .ukraine:
            return [.djinni, .dou]
        case .netherlands:
            return [.netherlands]
        case .europe:
            return [.netherlands, .eurojobs, .arbeitnow]
        case .usa:
            return [.hackernews, .remoteok, .weworkremotely]
        }
    }
}

#Preview {
    HunterContent()
        .environmentObject(AppState())
        .frame(width: 900, height: 700)
}
