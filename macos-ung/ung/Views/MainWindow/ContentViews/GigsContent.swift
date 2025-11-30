//
//  GigsContent.swift
//  ung
//
//  Kanban board for managing gigs with drag-and-drop.
//  Shows pipeline flow: Pipeline -> Negotiating -> Active -> Delivered -> Invoiced -> Complete
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

    struct GigWithTasks: Identifiable {
        let gig: Gig
        var tasks: [GigTask]
        let client: ClientModel?

        var id: Int64 { gig.id ?? 0 }
    }

    // Visible columns - focused flow
    let visibleStatuses: [GigStatus] = [.pipeline, .negotiating, .active, .delivered, .invoiced, .complete]

    func loadData() async {
        isLoading = true
        do {
            let data = try await DatabaseService.shared.getGigsWithTasks()
            gigs = data.map { GigWithTasks(gig: $0.gig, tasks: $0.tasks, client: $0.client) }
            clients = try await DatabaseService.shared.getClients()
        } catch {
            print("[GigBoard] Error loading: \(error)")
        }
        isLoading = false
    }

    func gigsForStatus(_ status: GigStatus) -> [GigWithTasks] {
        gigs.filter { $0.gig.gigStatus == status }
    }

    func moveGig(_ gigId: Int64, to status: GigStatus) async {
        do {
            try await DatabaseService.shared.updateGigStatus(id: gigId, status: status.rawValue)
            await loadData()
        } catch {
            print("[GigBoard] Error moving gig: \(error)")
        }
    }

    func createGig(name: String, clientId: Int64?, status: GigStatus = .pipeline) async {
        var gig = Gig(
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
        do {
            try await DatabaseService.shared.deleteGig(id: gigId)
            await loadData()
        } catch {
            print("[GigBoard] Error deleting gig: \(error)")
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

    var body: some View {
        VStack(spacing: 0) {
            // Header
            HStack {
                Text("Gigs Board")
                    .font(.title2)
                    .fontWeight(.semibold)

                Spacer()

                Button {
                    state.showNewGigSheet = true
                } label: {
                    Label("New Gig", systemImage: "plus")
                }
                .buttonStyle(.borderedProminent)
            }
            .padding()

            Divider()

            // Kanban Board
            if state.isLoading {
                ProgressView()
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                ScrollView(.horizontal, showsIndicators: true) {
                    HStack(alignment: .top, spacing: 16) {
                        ForEach(state.visibleStatuses, id: \.self) { status in
                            KanbanColumn(
                                status: status,
                                gigs: state.gigsForStatus(status),
                                state: state,
                                draggedGig: $draggedGig
                            )
                        }
                    }
                    .padding()
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
}

// MARK: - Kanban Column

struct KanbanColumn: View {
    let status: GigStatus
    let gigs: [GigBoardState.GigWithTasks]
    @ObservedObject var state: GigBoardState
    @Binding var draggedGig: GigBoardState.GigWithTasks?

    @State private var isTargeted = false

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            // Column Header
            HStack {
                Circle()
                    .fill(statusColor)
                    .frame(width: 10, height: 10)

                Text(status.displayName)
                    .font(.headline)
                    .foregroundColor(.primary)

                Text("\(gigs.count)")
                    .font(.caption)
                    .foregroundColor(.secondary)
                    .padding(.horizontal, 6)
                    .padding(.vertical, 2)
                    .background(Color.secondary.opacity(0.2))
                    .clipShape(Capsule())

                Spacer()
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 8)

            // Cards
            ScrollView(.vertical, showsIndicators: false) {
                LazyVStack(spacing: 10) {
                    ForEach(gigs) { gig in
                        GigCard(gig: gig, state: state)
                            .draggable(gig.id) {
                                GigCard(gig: gig, state: state)
                                    .frame(width: 260)
                                    .opacity(0.8)
                            }
                            .onDrag {
                                draggedGig = gig
                                return NSItemProvider(object: "\(gig.id)" as NSString)
                            }
                    }
                }
                .padding(.horizontal, 8)
                .padding(.bottom, 12)
            }

            Spacer()
        }
        .frame(width: 280)
        .background(
            RoundedRectangle(cornerRadius: 12)
                .fill(isTargeted ? statusColor.opacity(0.15) : Color(.controlBackgroundColor))
                .overlay(
                    RoundedRectangle(cornerRadius: 12)
                        .stroke(isTargeted ? statusColor : Color.clear, lineWidth: 2)
                )
        )
        .dropDestination(for: Int64.self) { items, _ in
            guard let gigId = items.first else { return false }
            Task {
                await state.moveGig(gigId, to: status)
            }
            return true
        } isTargeted: { targeted in
            isTargeted = targeted
        }
    }

    private var statusColor: Color {
        switch status {
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
}

// MARK: - Gig Card

struct GigCard: View {
    let gig: GigBoardState.GigWithTasks
    @ObservedObject var state: GigBoardState
    @State private var newTaskTitle = ""
    @State private var isAddingTask = false
    @FocusState private var taskFieldFocused: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            // Card Header
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    Text(gig.gig.name)
                        .font(.headline)
                        .lineLimit(2)

                    if let client = gig.client {
                        Text(client.name)
                            .font(.caption)
                            .foregroundColor(.secondary)
                    }
                }

                Spacer()

                Menu {
                    Button("Edit") {
                        state.selectedGig = gig
                    }
                    Divider()
                    Button("Delete", role: .destructive) {
                        Task {
                            await state.deleteGig(gig.id)
                        }
                    }
                } label: {
                    Image(systemName: "ellipsis")
                        .foregroundColor(.secondary)
                        .frame(width: 24, height: 24)
                }
                .buttonStyle(.plain)
            }

            // Stats Row
            HStack(spacing: 12) {
                if gig.gig.totalHoursTracked > 0 {
                    Label(String(format: "%.1fh", gig.gig.totalHoursTracked), systemImage: "clock")
                        .font(.caption2)
                        .foregroundColor(.secondary)
                }

                if let rate = gig.gig.hourlyRate {
                    Label("$\(Int(rate))/h", systemImage: "dollarsign.circle")
                        .font(.caption2)
                        .foregroundColor(.secondary)
                }

                if gig.gig.priority > 0 {
                    Image(systemName: gig.gig.priority == 2 ? "exclamationmark.2" : "exclamationmark")
                        .font(.caption2)
                        .foregroundColor(gig.gig.priority == 2 ? .red : .orange)
                }
            }

            // Tasks Section
            if !gig.tasks.isEmpty || isAddingTask {
                Divider()

                VStack(alignment: .leading, spacing: 6) {
                    ForEach(gig.tasks) { task in
                        TaskRow(task: task, state: state)
                    }

                    // New task input
                    if isAddingTask {
                        HStack(spacing: 6) {
                            Image(systemName: "circle")
                                .font(.caption)
                                .foregroundColor(.secondary)

                            TextField("Task...", text: $newTaskTitle)
                                .textFieldStyle(.plain)
                                .font(.caption)
                                .focused($taskFieldFocused)
                                .onSubmit {
                                    if !newTaskTitle.isEmpty {
                                        Task {
                                            await state.addTask(to: gig.id, title: newTaskTitle)
                                            newTaskTitle = ""
                                            isAddingTask = false
                                        }
                                    }
                                }
                        }
                    }
                }
            }

            // Add Task Button
            Button {
                isAddingTask = true
                taskFieldFocused = true
            } label: {
                HStack {
                    Image(systemName: "plus")
                    Text("Add task")
                }
                .font(.caption)
                .foregroundColor(.secondary)
            }
            .buttonStyle(.plain)
            .opacity(isAddingTask ? 0 : 1)
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color(.textBackgroundColor))
                .shadow(color: .black.opacity(0.08), radius: 3, y: 1)
        )
        .onTapGesture {
            state.selectedGig = gig
        }
    }
}

// MARK: - Task Row

struct TaskRow: View {
    let task: GigTask
    @ObservedObject var state: GigBoardState

    var body: some View {
        HStack(spacing: 6) {
            Button {
                Task {
                    await state.toggleTask(task.id!)
                }
            } label: {
                Image(systemName: task.completed ? "checkmark.circle.fill" : "circle")
                    .font(.caption)
                    .foregroundColor(task.completed ? .green : .secondary)
            }
            .buttonStyle(.plain)

            Text(task.title)
                .font(.caption)
                .strikethrough(task.completed)
                .foregroundColor(task.completed ? .secondary : .primary)
                .lineLimit(1)

            Spacer()

            Button {
                Task {
                    await state.deleteTask(task.id!)
                }
            } label: {
                Image(systemName: "xmark")
                    .font(.system(size: 8))
                    .foregroundColor(.secondary)
            }
            .buttonStyle(.plain)
            .opacity(0.5)
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

    var body: some View {
        NavigationStack {
            Form {
                Section("Gig Details") {
                    TextField("Name", text: $name)

                    Picker("Client", selection: $selectedClientId) {
                        Text("None").tag(nil as Int64?)
                        ForEach(state.clients) { client in
                            Text(client.name).tag(client.id as Int64?)
                        }
                    }

                    Picker("Status", selection: $selectedStatus) {
                        ForEach(state.visibleStatuses, id: \.self) { status in
                            Text(status.displayName).tag(status)
                        }
                    }
                }
            }
            .formStyle(.grouped)
            .navigationTitle("New Gig")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        dismiss()
                    }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Create") {
                        Task {
                            await state.createGig(
                                name: name,
                                clientId: selectedClientId,
                                status: selectedStatus
                            )
                            dismiss()
                        }
                    }
                    .disabled(name.isEmpty)
                }
            }
        }
        .frame(minWidth: 400, minHeight: 300)
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
                Section("Details") {
                    TextField("Name", text: $name)

                    if let client = gig.client {
                        LabeledContent("Client", value: client.name)
                    }

                    Picker("Status", selection: $status) {
                        ForEach(state.visibleStatuses, id: \.self) { s in
                            Text(s.displayName).tag(s)
                        }
                    }
                }

                Section("Stats") {
                    LabeledContent("Hours Tracked") {
                        Text(String(format: "%.1f h", gig.gig.totalHoursTracked))
                    }
                    if let rate = gig.gig.hourlyRate {
                        LabeledContent("Hourly Rate") {
                            Text("$\(Int(rate))/h")
                        }
                    }
                    LabeledContent("Total Invoiced") {
                        Text(String(format: "$%.0f", gig.gig.totalInvoiced))
                    }
                }

                Section("Tasks (\(gig.tasks.count))") {
                    ForEach(gig.tasks) { task in
                        HStack {
                            Button {
                                Task { await state.toggleTask(task.id!) }
                            } label: {
                                Image(systemName: task.completed ? "checkmark.circle.fill" : "circle")
                                    .foregroundColor(task.completed ? .green : .secondary)
                            }
                            .buttonStyle(.plain)

                            Text(task.title)
                                .strikethrough(task.completed)

                            Spacer()
                        }
                    }

                    HStack {
                        TextField("New task...", text: $newTaskTitle)
                            .onSubmit {
                                guard !newTaskTitle.isEmpty else { return }
                                Task {
                                    await state.addTask(to: gig.id, title: newTaskTitle)
                                    newTaskTitle = ""
                                }
                            }
                        Button("Add") {
                            guard !newTaskTitle.isEmpty else { return }
                            Task {
                                await state.addTask(to: gig.id, title: newTaskTitle)
                                newTaskTitle = ""
                            }
                        }
                        .disabled(newTaskTitle.isEmpty)
                    }
                }
            }
            .formStyle(.grouped)
            .navigationTitle(gig.gig.name)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Close") {
                        dismiss()
                    }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        Task {
                            // Update gig
                            var updated = gig.gig
                            updated.name = name
                            updated.status = status.rawValue
                            try? await DatabaseService.shared.updateGig(updated)
                            await state.loadData()
                            dismiss()
                        }
                    }
                }
            }
            .onAppear {
                name = gig.gig.name
                status = gig.gig.gigStatus
            }
        }
        .frame(minWidth: 500, minHeight: 500)
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
