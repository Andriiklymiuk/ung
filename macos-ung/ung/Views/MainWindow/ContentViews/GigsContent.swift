//
//  GigsContent.swift
//  ung
//
//  Beautiful kanban board for managing gigs with drag-and-drop.
//  Responsive design for iOS, iPad, and macOS.
//

import SwiftUI

// MARK: - Gig Board State

@MainActor
class GigBoardState: ObservableObject {
    @Published var gigs: [GigWithTasks] = []
    @Published var isLoading = true
    @Published var showNewGigSheet = false
    @Published var selectedGig: GigWithTasks?
    @Published var clients: [ClientModel] = []
    @Published var expandedColumns: Set<GigStatus> = Set(GigStatus.allCases)

    struct GigWithTasks: Identifiable, Equatable {
        let gig: Gig
        var tasks: [GigTask]
        let client: ClientModel?

        var id: Int64 { gig.id ?? 0 }

        static func == (lhs: GigWithTasks, rhs: GigWithTasks) -> Bool {
            lhs.id == rhs.id
        }

        var completedTasksCount: Int {
            tasks.filter { $0.completed }.count
        }

        var progress: Double {
            guard !tasks.isEmpty else { return 0 }
            return Double(completedTasksCount) / Double(tasks.count)
        }
    }

    let visibleStatuses: [GigStatus] = [.pipeline, .negotiating, .active, .delivered, .invoiced, .complete]

    func loadData() async {
        do {
            let data = try await DatabaseService.shared.getGigsWithTasks()
            withAnimation(.spring(response: 0.35, dampingFraction: 0.8)) {
                gigs = data.map { GigWithTasks(gig: $0.gig, tasks: $0.tasks, client: $0.client) }
            }
            clients = try await DatabaseService.shared.getClients()
        } catch {
            print("[GigBoard] Error loading: \(error)")
        }
        withAnimation(.easeOut(duration: 0.2)) {
            isLoading = false
        }
    }

    func gigsForStatus(_ status: GigStatus) -> [GigWithTasks] {
        gigs.filter { $0.gig.gigStatus == status }
    }

    func moveGig(_ gigId: Int64, to status: GigStatus) async {
        // Optimistic update with animation
        withAnimation(.spring(response: 0.4, dampingFraction: 0.75)) {
            if let index = gigs.firstIndex(where: { $0.id == gigId }) {
                var updatedGig = gigs[index].gig
                updatedGig.status = status.rawValue
                gigs[index] = GigWithTasks(gig: updatedGig, tasks: gigs[index].tasks, client: gigs[index].client)
            }
        }

        do {
            try await DatabaseService.shared.updateGigStatus(id: gigId, status: status.rawValue)
        } catch {
            print("[GigBoard] Error moving gig: \(error)")
            await loadData()
        }
    }

    func createGig(name: String, clientId: Int64?, status: GigStatus = .pipeline) async {
        let gig = Gig(
            name: name,
            clientId: clientId,
            contractId: nil,
            applicationId: nil,
            status: status.rawValue,
            gigType: "hourly",
            priority: 0,
            estimatedHours: nil,
            estimatedAmount: nil,
            hourlyRate: nil,
            currency: "USD",
            totalHoursTracked: 0,
            lastTrackedAt: nil,
            totalInvoiced: 0,
            lastInvoicedAt: nil,
            startDate: nil,
            dueDate: nil,
            completedAt: nil,
            description: nil,
            notes: nil,
            createdAt: nil,
            updatedAt: nil
        )
        do {
            _ = try await DatabaseService.shared.createGig(gig)
            await loadData()
        } catch {
            print("[GigBoard] Error creating gig: \(error)")
        }
    }

    func deleteGig(_ gigId: Int64) async {
        withAnimation(.spring(response: 0.3, dampingFraction: 0.8)) {
            gigs.removeAll { $0.id == gigId }
        }
        do {
            try await DatabaseService.shared.deleteGig(id: gigId)
        } catch {
            print("[GigBoard] Error deleting gig: \(error)")
            await loadData()
        }
    }

    func addTask(to gigId: Int64, title: String) async {
        let task = GigTask(gigId: gigId, title: title)
        do {
            _ = try await DatabaseService.shared.createGigTask(task)
            await loadData()
        } catch {
            print("[GigBoard] Error adding task: \(error)")
        }
    }

    func toggleTask(_ taskId: Int64) async {
        do {
            try await DatabaseService.shared.toggleGigTaskCompleted(id: taskId)
            await loadData()
        } catch {
            print("[GigBoard] Error toggling task: \(error)")
        }
    }

    func deleteTask(_ taskId: Int64) async {
        do {
            try await DatabaseService.shared.deleteGigTask(id: taskId)
            await loadData()
        } catch {
            print("[GigBoard] Error deleting task: \(error)")
        }
    }
}

// MARK: - Main Content View

struct GigsContent: View {
    @StateObject private var state = GigBoardState()
    @State private var draggedGig: GigBoardState.GigWithTasks?
    @Environment(\.colorScheme) private var colorScheme
    #if os(iOS)
    @Environment(\.horizontalSizeClass) private var horizontalSizeClass
    #endif

    private var isCompact: Bool {
        #if os(iOS)
        return horizontalSizeClass == .compact
        #else
        return false
        #endif
    }

    var body: some View {
        ZStack {
            // Background
            backgroundGradient
                .ignoresSafeArea()

            VStack(spacing: 0) {
                if state.isLoading {
                    loadingView
                } else if isCompact {
                    compactLayout
                } else {
                    kanbanLayout
                }
            }
        }
        .task {
            await state.loadData()
        }
        .sheet(isPresented: $state.showNewGigSheet) {
            NewGigSheet(state: state)
        }
        .sheet(item: $state.selectedGig) { gig in
            GigDetailSheet(gig: gig, state: state)
        }
    }

    // MARK: - Background

    private var backgroundGradient: some View {
        LinearGradient(
            colors: colorScheme == .dark
                ? [Color(white: 0.08), Color(white: 0.12)]
                : [Color(white: 0.96), Color(white: 0.98)],
            startPoint: .topLeading,
            endPoint: .bottomTrailing
        )
    }

    // MARK: - Loading View

    private var loadingView: some View {
        VStack(spacing: 20) {
            ProgressView()
                .scaleEffect(1.2)

            Text("Loading gigs...")
                .font(.subheadline)
                .foregroundColor(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    // MARK: - Compact Layout (iOS Phone)

    private var compactLayout: some View {
        ScrollView(.vertical, showsIndicators: false) {
            LazyVStack(spacing: 16) {
                // Quick Add Button
                quickAddButton
                    .padding(.horizontal, 16)
                    .padding(.top, 8)

                // Status sections
                ForEach(state.visibleStatuses, id: \.self) { status in
                    CompactStatusSection(
                        status: status,
                        gigs: state.gigsForStatus(status),
                        state: state
                    )
                }
            }
            .padding(.bottom, 24)
        }
    }

    // MARK: - Kanban Layout (iPad/Mac)

    private var kanbanLayout: some View {
        VStack(spacing: 0) {
            // Header Bar
            headerBar
                .padding(.horizontal, 20)
                .padding(.vertical, 12)

            // Kanban Board
            ScrollView(.horizontal, showsIndicators: false) {
                HStack(alignment: .top, spacing: 14) {
                    ForEach(state.visibleStatuses, id: \.self) { status in
                        KanbanColumn(
                            status: status,
                            gigs: state.gigsForStatus(status),
                            state: state,
                            draggedGig: $draggedGig
                        )
                        .transition(.asymmetric(
                            insertion: .scale(scale: 0.9).combined(with: .opacity),
                            removal: .scale(scale: 0.9).combined(with: .opacity)
                        ))
                    }
                }
                .padding(.horizontal, 20)
                .padding(.vertical, 16)
            }
        }
    }

    // MARK: - Header Bar

    private var headerBar: some View {
        HStack(spacing: 16) {
            // Stats pills
            HStack(spacing: 8) {
                StatPill(
                    value: "\(state.gigs.count)",
                    label: "Total",
                    color: .blue
                )

                let activeCount = state.gigsForStatus(.active).count
                if activeCount > 0 {
                    StatPill(
                        value: "\(activeCount)",
                        label: "Active",
                        color: .green
                    )
                }
            }

            Spacer()

            // Add button
            Button {
                withAnimation(.spring(response: 0.3, dampingFraction: 0.7)) {
                    state.showNewGigSheet = true
                }
            } label: {
                HStack(spacing: 6) {
                    Image(systemName: "plus")
                        .font(.system(size: 12, weight: .bold))
                    Text("New Gig")
                        .font(.system(size: 13, weight: .semibold))
                }
                .foregroundColor(.white)
                .padding(.horizontal, 14)
                .padding(.vertical, 8)
                .background(
                    Capsule()
                        .fill(
                            LinearGradient(
                                colors: [Color.blue, Color.blue.opacity(0.8)],
                                startPoint: .top,
                                endPoint: .bottom
                            )
                        )
                        .shadow(color: .blue.opacity(0.3), radius: 8, y: 4)
                )
            }
            .buttonStyle(ScaleButtonStyle())
        }
    }

    // MARK: - Quick Add Button (Compact)

    private var quickAddButton: some View {
        Button {
            withAnimation(.spring(response: 0.3, dampingFraction: 0.7)) {
                state.showNewGigSheet = true
            }
        } label: {
            HStack {
                Image(systemName: "plus.circle.fill")
                    .font(.title3)
                Text("Add New Gig")
                    .font(.system(size: 15, weight: .medium))
                Spacer()
                Image(systemName: "chevron.right")
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
            .foregroundColor(.blue)
            .padding(14)
            .background(
                RoundedRectangle(cornerRadius: 12)
                    .fill(Color.blue.opacity(0.08))
                    .overlay(
                        RoundedRectangle(cornerRadius: 12)
                            .stroke(Color.blue.opacity(0.2), lineWidth: 1)
                    )
            )
        }
        .buttonStyle(ScaleButtonStyle())
    }
}

// MARK: - Stat Pill

struct StatPill: View {
    let value: String
    let label: String
    let color: Color

    var body: some View {
        HStack(spacing: 4) {
            Text(value)
                .font(.system(size: 13, weight: .bold, design: .rounded))
                .foregroundColor(color)
            Text(label)
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.secondary)
        }
        .padding(.horizontal, 10)
        .padding(.vertical, 5)
        .background(
            Capsule()
                .fill(color.opacity(0.1))
        )
    }
}

// MARK: - Compact Status Section

struct CompactStatusSection: View {
    let status: GigStatus
    let gigs: [GigBoardState.GigWithTasks]
    @ObservedObject var state: GigBoardState
    @State private var isExpanded = true

    var body: some View {
        VStack(spacing: 10) {
            // Section Header
            Button {
                withAnimation(.spring(response: 0.3, dampingFraction: 0.8)) {
                    isExpanded.toggle()
                }
            } label: {
                HStack(spacing: 10) {
                    Circle()
                        .fill(status.color)
                        .frame(width: 8, height: 8)

                    Text(status.displayName)
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundColor(.primary)

                    Text("\(gigs.count)")
                        .font(.system(size: 12, weight: .medium, design: .rounded))
                        .foregroundColor(.secondary)
                        .padding(.horizontal, 6)
                        .padding(.vertical, 2)
                        .background(
                            Capsule()
                                .fill(Color.secondary.opacity(0.12))
                        )

                    Spacer()

                    Image(systemName: "chevron.right")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundColor(.secondary)
                        .rotationEffect(.degrees(isExpanded ? 90 : 0))
                }
                .padding(.horizontal, 16)
                .padding(.vertical, 10)
            }
            .buttonStyle(.plain)

            // Cards
            if isExpanded && !gigs.isEmpty {
                VStack(spacing: 10) {
                    ForEach(gigs) { gig in
                        CompactGigCard(gig: gig, state: state)
                            .transition(.asymmetric(
                                insertion: .scale(scale: 0.95).combined(with: .opacity),
                                removal: .scale(scale: 0.95).combined(with: .opacity)
                            ))
                    }
                }
                .padding(.horizontal, 16)
            }

            // Empty state
            if isExpanded && gigs.isEmpty {
                Text("No gigs")
                    .font(.caption)
                    .foregroundColor(.secondary)
                    .padding(.vertical, 8)
            }
        }
    }
}

// MARK: - Compact Gig Card

struct CompactGigCard: View {
    let gig: GigBoardState.GigWithTasks
    @ObservedObject var state: GigBoardState
    @State private var showActions = false
    @Environment(\.colorScheme) private var colorScheme

    var body: some View {
        Button {
            state.selectedGig = gig
        } label: {
            HStack(spacing: 12) {
                // Progress indicator
                ZStack {
                    Circle()
                        .stroke(Color.secondary.opacity(0.2), lineWidth: 3)

                    Circle()
                        .trim(from: 0, to: gig.progress)
                        .stroke(gig.gig.gigStatus.color, style: StrokeStyle(lineWidth: 3, lineCap: .round))
                        .rotationEffect(.degrees(-90))
                        .animation(.spring(response: 0.5, dampingFraction: 0.7), value: gig.progress)

                    if !gig.tasks.isEmpty {
                        Text("\(gig.completedTasksCount)")
                            .font(.system(size: 10, weight: .bold, design: .rounded))
                            .foregroundColor(gig.gig.gigStatus.color)
                    }
                }
                .frame(width: 36, height: 36)

                // Content
                VStack(alignment: .leading, spacing: 4) {
                    Text(gig.gig.name)
                        .font(.system(size: 15, weight: .medium))
                        .foregroundColor(.primary)
                        .lineLimit(1)

                    HStack(spacing: 8) {
                        if let client = gig.client {
                            Text(client.name)
                                .font(.caption)
                                .foregroundColor(.secondary)
                        }

                        if gig.gig.totalHoursTracked > 0 {
                            Label(String(format: "%.1fh", gig.gig.totalHoursTracked), systemImage: "clock")
                                .font(.caption2)
                                .foregroundColor(.secondary)
                        }
                    }
                }

                Spacer()

                // Chevron
                Image(systemName: "chevron.right")
                    .font(.system(size: 12, weight: .medium))
                    .foregroundColor(Color.secondary.opacity(0.5))
            }
            .padding(12)
            .background(
                RoundedRectangle(cornerRadius: 14)
                    .fill(colorScheme == .dark ? Color(white: 0.15) : .white)
                    .shadow(color: .black.opacity(colorScheme == .dark ? 0.3 : 0.06), radius: 8, y: 2)
            )
        }
        .buttonStyle(ScaleButtonStyle())
    }
}

// MARK: - Kanban Column

struct KanbanColumn: View {
    let status: GigStatus
    let gigs: [GigBoardState.GigWithTasks]
    @ObservedObject var state: GigBoardState
    @Binding var draggedGig: GigBoardState.GigWithTasks?
    @State private var isTargeted = false
    @Environment(\.colorScheme) private var colorScheme

    var body: some View {
        VStack(spacing: 0) {
            // Column Header
            HStack(spacing: 8) {
                Circle()
                    .fill(status.color)
                    .frame(width: 8, height: 8)

                Text(status.displayName)
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundColor(.primary)

                Spacer()

                Text("\(gigs.count)")
                    .font(.system(size: 11, weight: .bold, design: .rounded))
                    .foregroundColor(status.color)
                    .padding(.horizontal, 7)
                    .padding(.vertical, 3)
                    .background(
                        Capsule()
                            .fill(status.color.opacity(0.15))
                    )
            }
            .padding(.horizontal, 14)
            .padding(.vertical, 12)

            // Cards
            ScrollView(.vertical, showsIndicators: false) {
                LazyVStack(spacing: 10) {
                    ForEach(gigs) { gig in
                        GigCard(gig: gig, state: state)
                            .draggable(gig.id) {
                                GigCard(gig: gig, state: state)
                                    .frame(width: 240)
                                    .opacity(0.9)
                                    .scaleEffect(1.02)
                            }
                            .onDrag {
                                draggedGig = gig
                                return NSItemProvider(object: "\(gig.id)" as NSString)
                            }
                            .transition(.asymmetric(
                                insertion: .scale(scale: 0.9).combined(with: .opacity).combined(with: .offset(y: -10)),
                                removal: .scale(scale: 0.9).combined(with: .opacity)
                            ))
                    }

                    // Empty state
                    if gigs.isEmpty {
                        emptyColumnView
                    }
                }
                .padding(.horizontal, 10)
                .padding(.bottom, 16)
            }
        }
        .frame(width: 260)
        .background(
            RoundedRectangle(cornerRadius: 16)
                .fill(colorScheme == .dark ? Color(white: 0.12) : Color(white: 0.97))
                .overlay(
                    RoundedRectangle(cornerRadius: 16)
                        .stroke(
                            isTargeted ? status.color.opacity(0.5) : Color.clear,
                            lineWidth: 2
                        )
                )
                .shadow(
                    color: isTargeted ? status.color.opacity(0.2) : .black.opacity(colorScheme == .dark ? 0.3 : 0.05),
                    radius: isTargeted ? 12 : 8,
                    y: 2
                )
        )
        .scaleEffect(isTargeted ? 1.02 : 1.0)
        .animation(.spring(response: 0.25, dampingFraction: 0.7), value: isTargeted)
        .dropDestination(for: Int64.self) { items, _ in
            guard let gigId = items.first else { return false }
            Task {
                await state.moveGig(gigId, to: status)
            }
            return true
        } isTargeted: { targeted in
            withAnimation(.spring(response: 0.2, dampingFraction: 0.8)) {
                isTargeted = targeted
            }
        }
    }

    private var emptyColumnView: some View {
        VStack(spacing: 8) {
            Image(systemName: status.emptyIcon)
                .font(.system(size: 24))
                .foregroundColor(.secondary.opacity(0.5))

            Text("No gigs")
                .font(.caption)
                .foregroundColor(.secondary)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 32)
    }
}

// MARK: - Gig Card

struct GigCard: View {
    let gig: GigBoardState.GigWithTasks
    @ObservedObject var state: GigBoardState
    @State private var newTaskTitle = ""
    @State private var isAddingTask = false
    @State private var isHovered = false
    @FocusState private var taskFieldFocused: Bool
    @Environment(\.colorScheme) private var colorScheme

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            // Header
            HStack(alignment: .top) {
                VStack(alignment: .leading, spacing: 3) {
                    Text(gig.gig.name)
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundColor(.primary)
                        .lineLimit(2)
                        .fixedSize(horizontal: false, vertical: true)

                    if let client = gig.client {
                        Text(client.name)
                            .font(.system(size: 11))
                            .foregroundColor(.secondary)
                    }
                }

                Spacer()

                Menu {
                    Button { state.selectedGig = gig } label: {
                        Label("Edit", systemImage: "pencil")
                    }
                    Divider()
                    Button(role: .destructive) {
                        Task { await state.deleteGig(gig.id) }
                    } label: {
                        Label("Delete", systemImage: "trash")
                    }
                } label: {
                    Image(systemName: "ellipsis")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundColor(.secondary)
                        .frame(width: 24, height: 24)
                        .contentShape(Rectangle())
                }
                .buttonStyle(.plain)
                .opacity(isHovered ? 1 : 0.5)
            }

            // Progress bar (if has tasks)
            if !gig.tasks.isEmpty {
                VStack(spacing: 4) {
                    GeometryReader { geo in
                        ZStack(alignment: .leading) {
                            Capsule()
                                .fill(Color.secondary.opacity(0.15))

                            Capsule()
                                .fill(
                                    LinearGradient(
                                        colors: [gig.gig.gigStatus.color, gig.gig.gigStatus.color.opacity(0.7)],
                                        startPoint: .leading,
                                        endPoint: .trailing
                                    )
                                )
                                .frame(width: geo.size.width * gig.progress)
                                .animation(.spring(response: 0.4, dampingFraction: 0.7), value: gig.progress)
                        }
                    }
                    .frame(height: 4)

                    HStack {
                        Text("\(gig.completedTasksCount)/\(gig.tasks.count) tasks")
                            .font(.system(size: 10, weight: .medium))
                            .foregroundColor(.secondary)
                        Spacer()
                    }
                }
            }

            // Stats
            if gig.gig.totalHoursTracked > 0 || gig.gig.hourlyRate != nil {
                HStack(spacing: 10) {
                    if gig.gig.totalHoursTracked > 0 {
                        HStack(spacing: 3) {
                            Image(systemName: "clock")
                                .font(.system(size: 9))
                            Text(String(format: "%.1fh", gig.gig.totalHoursTracked))
                                .font(.system(size: 10, weight: .medium))
                        }
                        .foregroundColor(.secondary)
                    }

                    if let rate = gig.gig.hourlyRate {
                        HStack(spacing: 2) {
                            Text("$\(Int(rate))")
                                .font(.system(size: 10, weight: .semibold))
                            Text("/h")
                                .font(.system(size: 9))
                        }
                        .foregroundColor(.green)
                    }
                }
            }

            // Task input
            if isAddingTask {
                HStack(spacing: 6) {
                    Circle()
                        .stroke(Color.secondary.opacity(0.3), lineWidth: 1.5)
                        .frame(width: 14, height: 14)

                    TextField("Add task...", text: $newTaskTitle)
                        .textFieldStyle(.plain)
                        .font(.system(size: 12))
                        .focused($taskFieldFocused)
                        .onSubmit {
                            submitTask()
                        }
                }
                .padding(8)
                .background(
                    RoundedRectangle(cornerRadius: 8)
                        .fill(Color.secondary.opacity(0.08))
                )
                .transition(.scale(scale: 0.95).combined(with: .opacity))
            }

            // Add task button
            if !isAddingTask {
                Button {
                    withAnimation(.spring(response: 0.25, dampingFraction: 0.8)) {
                        isAddingTask = true
                    }
                    taskFieldFocused = true
                } label: {
                    HStack(spacing: 4) {
                        Image(systemName: "plus")
                            .font(.system(size: 10, weight: .medium))
                        Text("Add task")
                            .font(.system(size: 11, weight: .medium))
                    }
                    .foregroundColor(.secondary)
                }
                .buttonStyle(.plain)
                .opacity(isHovered ? 1 : 0.6)
            }
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 12)
                .fill(colorScheme == .dark ? Color(white: 0.18) : .white)
                .shadow(
                    color: .black.opacity(colorScheme == .dark ? 0.4 : 0.08),
                    radius: isHovered ? 8 : 4,
                    y: isHovered ? 4 : 2
                )
        )
        .scaleEffect(isHovered ? 1.01 : 1.0)
        .animation(.spring(response: 0.2, dampingFraction: 0.8), value: isHovered)
        .onHover { hovering in
            isHovered = hovering
        }
        .onTapGesture {
            state.selectedGig = gig
        }
    }

    private func submitTask() {
        guard !newTaskTitle.isEmpty else {
            withAnimation(.spring(response: 0.25, dampingFraction: 0.8)) {
                isAddingTask = false
            }
            return
        }
        Task {
            await state.addTask(to: gig.id, title: newTaskTitle)
            newTaskTitle = ""
            withAnimation(.spring(response: 0.25, dampingFraction: 0.8)) {
                isAddingTask = false
            }
        }
    }
}

// MARK: - New Gig Sheet

struct NewGigSheet: View {
    @ObservedObject var state: GigBoardState
    @Environment(\.dismiss) private var dismiss
    @State private var name = ""
    @State private var selectedClientId: Int64?
    @State private var selectedStatus: GigStatus = .pipeline
    @FocusState private var nameFocused: Bool

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("What's the gig?", text: $name)
                        .font(.system(size: 16))
                        .focused($nameFocused)
                } header: {
                    Text("Name")
                }

                Section {
                    Picker("Client", selection: $selectedClientId) {
                        Text("None").tag(nil as Int64?)
                        ForEach(state.clients) { client in
                            Text(client.name).tag(client.id as Int64?)
                        }
                    }

                    Picker("Start in", selection: $selectedStatus) {
                        ForEach(state.visibleStatuses, id: \.self) { status in
                            HStack {
                                Circle()
                                    .fill(status.color)
                                    .frame(width: 8, height: 8)
                                Text(status.displayName)
                            }
                            .tag(status)
                        }
                    }
                }
            }
            .formStyle(.grouped)
            .navigationTitle("New Gig")
            #if os(iOS)
            .navigationBarTitleDisplayMode(.inline)
            #endif
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Create") {
                        Task {
                            await state.createGig(name: name, clientId: selectedClientId, status: selectedStatus)
                            dismiss()
                        }
                    }
                    .fontWeight(.semibold)
                    .disabled(name.isEmpty)
                }
            }
            .onAppear {
                nameFocused = true
            }
        }
        #if os(macOS)
        .frame(minWidth: 380, minHeight: 280)
        #endif
    }
}

// MARK: - Gig Detail Sheet

struct GigDetailSheet: View {
    let gig: GigBoardState.GigWithTasks
    @ObservedObject var state: GigBoardState
    @Environment(\.dismiss) private var dismiss
    @State private var name: String = ""
    @State private var status: GigStatus = .pipeline
    @State private var newTaskTitle = ""

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Name", text: $name)
                        .font(.system(size: 15))

                    if let client = gig.client {
                        LabeledContent("Client", value: client.name)
                    }

                    Picker("Status", selection: $status) {
                        ForEach(state.visibleStatuses, id: \.self) { s in
                            HStack {
                                Circle().fill(s.color).frame(width: 8, height: 8)
                                Text(s.displayName)
                            }
                            .tag(s)
                        }
                    }
                } header: {
                    Text("Details")
                }

                Section {
                    LabeledContent("Hours") {
                        Text(String(format: "%.1f h", gig.gig.totalHoursTracked))
                            .foregroundColor(.secondary)
                    }
                    if let rate = gig.gig.hourlyRate {
                        LabeledContent("Rate") {
                            Text("$\(Int(rate))/h")
                                .foregroundColor(.green)
                        }
                    }
                    LabeledContent("Invoiced") {
                        Text(String(format: "$%.0f", gig.gig.totalInvoiced))
                            .foregroundColor(.secondary)
                    }
                } header: {
                    Text("Stats")
                }

                Section {
                    ForEach(gig.tasks) { task in
                        TaskDetailRow(task: task, state: state)
                    }

                    HStack {
                        TextField("New task...", text: $newTaskTitle)
                            .onSubmit { addTask() }

                        Button(action: addTask) {
                            Image(systemName: "plus.circle.fill")
                                .foregroundColor(.blue)
                        }
                        .buttonStyle(.plain)
                        .disabled(newTaskTitle.isEmpty)
                    }
                } header: {
                    Text("Tasks (\(gig.tasks.count))")
                }
            }
            .formStyle(.grouped)
            .navigationTitle(gig.gig.name)
            #if os(iOS)
            .navigationBarTitleDisplayMode(.inline)
            #endif
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Close") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        Task {
                            var updated = gig.gig
                            updated.name = name
                            updated.status = status.rawValue
                            try? await DatabaseService.shared.updateGig(updated)
                            await state.loadData()
                            dismiss()
                        }
                    }
                    .fontWeight(.semibold)
                }
            }
            .onAppear {
                name = gig.gig.name
                status = gig.gig.gigStatus
            }
        }
        #if os(macOS)
        .frame(minWidth: 450, minHeight: 450)
        #endif
    }

    private func addTask() {
        guard !newTaskTitle.isEmpty else { return }
        Task {
            await state.addTask(to: gig.id, title: newTaskTitle)
            newTaskTitle = ""
        }
    }
}

// MARK: - Task Detail Row

struct TaskDetailRow: View {
    let task: GigTask
    @ObservedObject var state: GigBoardState
    @State private var isHovered = false

    var body: some View {
        HStack(spacing: 10) {
            Button {
                Task { await state.toggleTask(task.id!) }
            } label: {
                Image(systemName: task.completed ? "checkmark.circle.fill" : "circle")
                    .font(.system(size: 18))
                    .foregroundColor(task.completed ? .green : .secondary)
                    .contentTransition(.symbolEffect(.replace))
            }
            .buttonStyle(.plain)

            Text(task.title)
                .strikethrough(task.completed)
                .foregroundColor(task.completed ? .secondary : .primary)

            Spacer()

            if isHovered {
                Button {
                    Task { await state.deleteTask(task.id!) }
                } label: {
                    Image(systemName: "xmark.circle.fill")
                        .font(.system(size: 14))
                        .foregroundColor(.secondary)
                }
                .buttonStyle(.plain)
                .transition(.scale.combined(with: .opacity))
            }
        }
        .padding(.vertical, 2)
        .onHover { isHovered = $0 }
        .animation(.easeInOut(duration: 0.15), value: isHovered)
    }
}

// MARK: - Scale Button Style

struct ScaleButtonStyle: ButtonStyle {
    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .scaleEffect(configuration.isPressed ? 0.96 : 1.0)
            .opacity(configuration.isPressed ? 0.9 : 1.0)
            .animation(.spring(response: 0.2, dampingFraction: 0.7), value: configuration.isPressed)
    }
}

// MARK: - GigStatus Extensions

extension GigStatus {
    var color: Color {
        switch self {
        case .pipeline: return .gray
        case .negotiating: return .purple
        case .active: return .blue
        case .delivered: return .orange
        case .invoiced: return .cyan
        case .complete: return .green
        case .onHold: return .yellow
        case .cancelled: return .red
        }
    }

    var emptyIcon: String {
        switch self {
        case .pipeline: return "tray"
        case .negotiating: return "bubble.left.and.bubble.right"
        case .active: return "bolt"
        case .delivered: return "paperplane"
        case .invoiced: return "doc.text"
        case .complete: return "checkmark.seal"
        case .onHold: return "pause.circle"
        case .cancelled: return "xmark.circle"
        }
    }
}

// MARK: - Transferable

extension Int64: Transferable {
    public static var transferRepresentation: some TransferRepresentation {
        CodableRepresentation(contentType: .plainText)
    }
}

#Preview {
    GigsContent()
}
