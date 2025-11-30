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

  @State private var projectName: String = "Software"
  @State private var selectedClientId: Int?
  @State private var isStarting = false

  // Recent projects for quick selection
  private let recentProjects = ["Software", "Development", "Design", "Meeting", "Research"]

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
          quickSelectSection

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
    .background(Design.Colors.windowBackground)
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
          .font(Design.Typography.headingSmall)
        Text("What are you working on?")
          .font(Design.Typography.bodySmall)
          .foregroundColor(Design.Colors.textSecondary)
      }

      Spacer()

      Button(action: { dismiss() }) {
        Image(systemName: "xmark.circle.fill")
          .font(.system(size: 18))
          .foregroundColor(Design.Colors.textTertiary)
      }
      .buttonStyle(DSInteractiveStyle())
    }
    .padding(Design.Spacing.md)
  }

  // MARK: - Project Input
  private var projectInputSection: some View {
    VStack(alignment: .leading, spacing: Design.Spacing.xs) {
      Text("Project / Task")
        .font(Design.Typography.labelSmall)
        .foregroundColor(Design.Colors.textSecondary)

      HStack(spacing: Design.Spacing.xs) {
        Image(systemName: "folder.fill")
          .font(.system(size: Design.IconSize.xs))
          .foregroundColor(Design.Colors.primary)

        TextField("Enter project name...", text: $projectName)
          .textFieldStyle(.plain)
          .font(Design.Typography.bodyMedium)
      }
      .padding(Design.Spacing.sm)
      .background(
        RoundedRectangle(cornerRadius: Design.Radius.sm)
          .fill(Design.Colors.controlBackground)
          .overlay(
            RoundedRectangle(cornerRadius: Design.Radius.sm)
              .stroke(Design.Colors.primary.opacity(0.3), lineWidth: 1)
          )
      )
    }
  }

  // MARK: - Quick Select
  private var quickSelectSection: some View {
    VStack(alignment: .leading, spacing: Design.Spacing.xs) {
      Text("Quick Select")
        .font(Design.Typography.labelSmall)
        .foregroundColor(Design.Colors.textSecondary)

      FlowLayout(spacing: 6) {
        ForEach(recentProjects, id: \.self) { project in
          Button(action: {
            withAnimation(Design.Animation.snappy) {
              projectName = project
            }
          }) {
            Text(project)
          }
          .buttonStyle(DSPillButtonStyle(color: Design.Colors.primary, isSelected: projectName == project))
        }
      }
    }
  }

  // MARK: - Client Selection
  private var clientSelectionSection: some View {
    VStack(alignment: .leading, spacing: Design.Spacing.xs) {
      Text("Client (Optional)")
        .font(Design.Typography.labelSmall)
        .foregroundColor(Design.Colors.textSecondary)

      if appState.clients.isEmpty {
        HStack(spacing: Design.Spacing.xs) {
          Image(systemName: "person.fill.questionmark")
            .font(.system(size: Design.IconSize.xs))
            .foregroundColor(Design.Colors.textTertiary)

          Text("No clients yet")
            .font(Design.Typography.bodySmall)
            .foregroundColor(Design.Colors.textTertiary)
        }
        .padding(Design.Spacing.sm)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(
          RoundedRectangle(cornerRadius: Design.Radius.sm)
            .fill(Design.Colors.controlBackground)
        )
      } else {
        VStack(spacing: Design.Spacing.xxs) {
          // None option
          ClientOptionRow(
            client: nil,
            isSelected: selectedClientId == nil,
            action: { withAnimation(Design.Animation.snappy) { selectedClientId = nil } }
          )

          ForEach(appState.clients) { client in
            ClientOptionRow(
              client: client,
              isSelected: selectedClientId == client.id,
              action: { withAnimation(Design.Animation.snappy) { selectedClientId = client.id } }
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
      .buttonStyle(DSGhostButtonStyle())

      Spacer()

      Button(action: startTracking) {
        HStack(spacing: 6) {
          if isStarting {
            ProgressView()
              .scaleEffect(0.7)
              .tint(.white)
          } else {
            Image(systemName: "play.fill")
              .font(.system(size: 10))
          }
          Text("Start")
        }
      }
      .buttonStyle(DSCompactButtonStyle(isDisabled: projectName.isEmpty))
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
      HStack(spacing: Design.Spacing.xs) {
        Image(systemName: isSelected ? "checkmark.circle.fill" : "circle")
          .font(.system(size: Design.IconSize.xs))
          .foregroundColor(isSelected ? Design.Colors.primary : Design.Colors.textTertiary)

        if let client = client {
          Image(systemName: "person.fill")
            .font(.system(size: 10))
            .foregroundColor(Design.Colors.textSecondary)
          Text(client.name)
            .font(Design.Typography.bodySmall)
            .foregroundColor(Design.Colors.textPrimary)
        } else {
          Text("No Client")
            .font(Design.Typography.bodySmall)
            .foregroundColor(Design.Colors.textSecondary)
        }

        Spacer()
      }
      .padding(Design.Spacing.xs)
      .background(
        RoundedRectangle(cornerRadius: Design.Radius.xs)
          .fill(isSelected ? Design.Colors.primary.opacity(0.1) : Color.clear)
      )
    }
    .buttonStyle(DSInteractiveStyle())
  }
}

// MARK: - Flow Layout
struct FlowLayout: Layout {
  var spacing: CGFloat = 8

  func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
    let result = FlowResult(in: proposal.width ?? 0, subviews: subviews, spacing: spacing)
    return result.size
  }

  func placeSubviews(
    in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()
  ) {
    let result = FlowResult(in: bounds.width, subviews: subviews, spacing: spacing)
    for (index, subview) in subviews.enumerated() {
      subview.place(
        at: CGPoint(
          x: bounds.minX + result.positions[index].x,
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
