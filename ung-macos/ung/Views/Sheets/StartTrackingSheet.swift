//
//  StartTrackingSheet.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI

struct StartTrackingSheet: View {
    @EnvironmentObject var appState: AppState
    @Environment(\.dismiss) var dismiss

    @State private var projectName: String = ""
    @State private var selectedClientId: Int?
    @State private var isStarting = false

    // Recent projects for quick selection
    private let recentProjects = ["Development", "Design", "Meeting", "Research", "Admin"]

    var body: some View {
        VStack(spacing: 0) {
            // Header
            headerSection

            Divider()

            // Content
            ScrollView {
                VStack(spacing: 16) {
                    // Project name input
                    projectInputSection

                    // Quick select recent projects
                    if projectName.isEmpty {
                        quickSelectSection
                    }

                    // Client selection
                    clientSelectionSection
                }
                .padding(16)
            }

            Divider()

            // Footer actions
            footerSection
        }
        .frame(width: 320, height: 400)
        .background(Color(nsColor: .windowBackgroundColor))
        .onAppear {
            Task {
                await appState.refreshDashboard()
            }
        }
    }

    // MARK: - Header
    private var headerSection: some View {
        HStack {
            VStack(alignment: .leading, spacing: 2) {
                Text("Start Tracking")
                    .font(.system(size: 14, weight: .semibold))
                Text("What are you working on?")
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
            }

            Spacer()

            Button(action: { dismiss() }) {
                Image(systemName: "xmark.circle.fill")
                    .font(.system(size: 18))
                    .foregroundColor(.secondary)
            }
            .buttonStyle(.plain)
        }
        .padding(16)
    }

    // MARK: - Project Input
    private var projectInputSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Project / Task")
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.secondary)

            HStack(spacing: 8) {
                Image(systemName: "folder.fill")
                    .font(.system(size: 12))
                    .foregroundColor(.blue)

                TextField("Enter project name...", text: $projectName)
                    .textFieldStyle(.plain)
                    .font(.system(size: 13))
            }
            .padding(10)
            .background(
                RoundedRectangle(cornerRadius: 8)
                    .fill(Color(nsColor: .controlBackgroundColor))
                    .overlay(
                        RoundedRectangle(cornerRadius: 8)
                            .stroke(Color.blue.opacity(0.3), lineWidth: 1)
                    )
            )
        }
    }

    // MARK: - Quick Select
    private var quickSelectSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Quick Select")
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.secondary)

            FlowLayout(spacing: 6) {
                ForEach(recentProjects, id: \.self) { project in
                    Button(action: { projectName = project }) {
                        Text(project)
                            .font(.system(size: 11))
                            .padding(.horizontal, 10)
                            .padding(.vertical, 6)
                            .background(Color(nsColor: .controlBackgroundColor))
                            .cornerRadius(6)
                    }
                    .buttonStyle(.plain)
                }
            }
        }
    }

    // MARK: - Client Selection
    private var clientSelectionSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Client (Optional)")
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.secondary)

            if appState.clients.isEmpty {
                HStack(spacing: 8) {
                    Image(systemName: "person.fill.questionmark")
                        .font(.system(size: 12))
                        .foregroundColor(.secondary)

                    Text("No clients yet")
                        .font(.system(size: 12))
                        .foregroundColor(.secondary)
                }
                .padding(10)
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(
                    RoundedRectangle(cornerRadius: 8)
                        .fill(Color(nsColor: .controlBackgroundColor))
                )
            } else {
                VStack(spacing: 4) {
                    // None option
                    ClientOptionRow(
                        client: nil,
                        isSelected: selectedClientId == nil,
                        action: { selectedClientId = nil }
                    )

                    ForEach(appState.clients) { client in
                        ClientOptionRow(
                            client: client,
                            isSelected: selectedClientId == client.id,
                            action: { selectedClientId = client.id }
                        )
                    }
                }
            }
        }
    }

    // MARK: - Footer
    private var footerSection: some View {
        HStack(spacing: 12) {
            Button("Cancel") {
                dismiss()
            }
            .buttonStyle(.plain)
            .foregroundColor(.secondary)

            Spacer()

            Button(action: startTracking) {
                HStack(spacing: 6) {
                    if isStarting {
                        ProgressView()
                            .scaleEffect(0.7)
                    } else {
                        Image(systemName: "play.fill")
                            .font(.system(size: 10))
                    }
                    Text("Start")
                        .font(.system(size: 12, weight: .semibold))
                }
                .foregroundColor(.white)
                .padding(.horizontal, 16)
                .padding(.vertical, 8)
                .background(
                    RoundedRectangle(cornerRadius: 8)
                        .fill(projectName.isEmpty ? Color.gray : Color.green)
                )
            }
            .buttonStyle(.plain)
            .disabled(projectName.isEmpty || isStarting)
        }
        .padding(16)
    }

    // MARK: - Actions
    private func startTracking() {
        isStarting = true
        Task {
            await appState.startTracking(project: projectName, clientId: selectedClientId)
            isStarting = false
            dismiss()
        }
    }
}

// MARK: - Client Option Row
struct ClientOptionRow: View {
    let client: Client?
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: 8) {
                Image(systemName: isSelected ? "checkmark.circle.fill" : "circle")
                    .font(.system(size: 12))
                    .foregroundColor(isSelected ? .blue : .secondary)

                if let client = client {
                    Image(systemName: "person.fill")
                        .font(.system(size: 10))
                        .foregroundColor(.secondary)
                    Text(client.name)
                        .font(.system(size: 12))
                        .foregroundColor(.primary)
                } else {
                    Text("No Client")
                        .font(.system(size: 12))
                        .foregroundColor(.secondary)
                }

                Spacer()
            }
            .padding(8)
            .background(
                RoundedRectangle(cornerRadius: 6)
                    .fill(isSelected ? Color.blue.opacity(0.1) : Color.clear)
            )
        }
        .buttonStyle(.plain)
    }
}

// MARK: - Flow Layout
struct FlowLayout: Layout {
    var spacing: CGFloat = 8

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        let result = FlowResult(in: proposal.width ?? 0, subviews: subviews, spacing: spacing)
        return result.size
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        let result = FlowResult(in: bounds.width, subviews: subviews, spacing: spacing)
        for (index, subview) in subviews.enumerated() {
            subview.place(at: CGPoint(x: bounds.minX + result.positions[index].x,
                                      y: bounds.minY + result.positions[index].y),
                         proposal: .unspecified)
        }
    }

    struct FlowResult {
        var size: CGSize = .zero
        var positions: [CGPoint] = []

        init(in maxWidth: CGFloat, subviews: Subviews, spacing: CGFloat) {
            var x: CGFloat = 0
            var y: CGFloat = 0
            var rowHeight: CGFloat = 0

            for subview in subviews {
                let size = subview.sizeThatFits(.unspecified)

                if x + size.width > maxWidth && x > 0 {
                    x = 0
                    y += rowHeight + spacing
                    rowHeight = 0
                }

                positions.append(CGPoint(x: x, y: y))
                rowHeight = max(rowHeight, size.height)
                x += size.width + spacing

                self.size.width = max(self.size.width, x - spacing)
            }

            self.size.height = y + rowHeight
        }
    }
}

#Preview {
    StartTrackingSheet()
        .environmentObject(AppState())
}
