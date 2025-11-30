//
//  LogExpenseSheet.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI

struct LogExpenseSheet: View {
    @EnvironmentObject var appState: AppState
    @Environment(\.dismiss) var dismiss

    @State private var description: String = ""
    @State private var amountText: String = ""
    @State private var selectedCategory: ExpenseCategory = .other
    @State private var isLogging = false

    enum ExpenseCategory: String, CaseIterable {
        case software = "Software"
        case hardware = "Hardware"
        case travel = "Travel"
        case meals = "Meals"
        case office = "Office Supplies"
        case utilities = "Utilities"
        case marketing = "Marketing"
        case other = "Other"

        var icon: String {
            switch self {
            case .software: return "app.badge.fill"
            case .hardware: return "desktopcomputer"
            case .travel: return "airplane"
            case .meals: return "fork.knife"
            case .office: return "paperclip"
            case .utilities: return "bolt.fill"
            case .marketing: return "megaphone.fill"
            case .other: return "tag.fill"
            }
        }

        var color: Color {
            switch self {
            case .software: return .blue
            case .hardware: return .purple
            case .travel: return .cyan
            case .meals: return .orange
            case .office: return .gray
            case .utilities: return .yellow
            case .marketing: return .pink
            case .other: return .secondary
            }
        }
    }

    var amount: Double {
        Double(amountText.replacingOccurrences(of: ",", with: ".")) ?? 0
    }

    var isValid: Bool {
        !description.isEmpty && amount > 0
    }

    var body: some View {
        VStack(spacing: 0) {
            // Header
            headerSection

            Divider()

            // Content
            ScrollView {
                VStack(spacing: 16) {
                    // Description input
                    descriptionInputSection

                    // Amount input
                    amountInputSection

                    // Category selection
                    categorySelectionSection
                }
                .padding(16)
            }

            Divider()

            // Footer actions
            footerSection
        }
        .frame(width: 320, height: 420)
        .background(Design.Colors.windowBackground)
    }

    // MARK: - Header
    private var headerSection: some View {
        HStack {
            VStack(alignment: .leading, spacing: 2) {
                Text("Log Expense")
                    .font(Design.Typography.headingSmall)
                Text("Track your business expenses")
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

    // MARK: - Description Input
    private var descriptionInputSection: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.xs) {
            Text("Description")
                .font(Design.Typography.labelSmall)
                .foregroundColor(Design.Colors.textSecondary)

            HStack(spacing: Design.Spacing.xs) {
                Image(systemName: "text.alignleft")
                    .font(.system(size: Design.IconSize.xs))
                    .foregroundColor(Design.Colors.primary)

                TextField("What did you spend on?", text: $description)
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

    // MARK: - Amount Input
    private var amountInputSection: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.xs) {
            Text("Amount")
                .font(Design.Typography.labelSmall)
                .foregroundColor(Design.Colors.textSecondary)

            HStack(spacing: Design.Spacing.xs) {
                Image(systemName: "dollarsign.circle.fill")
                    .font(.system(size: Design.IconSize.xs))
                    .foregroundColor(Design.Colors.success)

                TextField("0.00", text: $amountText)
                    .textFieldStyle(.plain)
                    .font(.system(size: 16, weight: .semibold, design: .rounded))
            }
            .padding(Design.Spacing.sm)
            .background(
                RoundedRectangle(cornerRadius: Design.Radius.sm)
                    .fill(Design.Colors.controlBackground)
                    .overlay(
                        RoundedRectangle(cornerRadius: Design.Radius.sm)
                            .stroke(Design.Colors.success.opacity(0.3), lineWidth: 1)
                    )
            )
        }
    }

    // MARK: - Category Selection
    private var categorySelectionSection: some View {
        VStack(alignment: .leading, spacing: Design.Spacing.xs) {
            Text("Category")
                .font(Design.Typography.labelSmall)
                .foregroundColor(Design.Colors.textSecondary)

            LazyVGrid(columns: [
                GridItem(.flexible()),
                GridItem(.flexible())
            ], spacing: Design.Spacing.xs) {
                ForEach(ExpenseCategory.allCases, id: \.self) { category in
                    CategoryButton(
                        category: category,
                        isSelected: selectedCategory == category,
                        action: { withAnimation(Design.Animation.snappy) { selectedCategory = category } }
                    )
                }
            }
        }
    }

    // MARK: - Footer
    private var footerSection: some View {
        HStack(spacing: Design.Spacing.sm) {
            Button("Cancel") {
                dismiss()
            }
            .buttonStyle(DSGhostButtonStyle())

            Spacer()

            Button(action: logExpense) {
                HStack(spacing: 6) {
                    if isLogging {
                        ProgressView()
                            .scaleEffect(0.7)
                            .tint(.white)
                    } else {
                        Image(systemName: "plus.circle.fill")
                            .font(.system(size: 10))
                    }
                    Text("Log Expense")
                }
            }
            .buttonStyle(DSCompactButtonStyle(isDisabled: !isValid))
            .disabled(!isValid || isLogging)
        }
        .padding(Design.Spacing.md)
    }

    // MARK: - Actions
    private func logExpense() {
        isLogging = true
        Task {
            let newExpense = Expense(
                description: description,
                amount: amount,
                currency: "USD",
                category: selectedCategory.rawValue.lowercased(),
                date: Date()
            )
            _ = try? await appState.database.createExpense(newExpense)
            await appState.refreshDashboard()
            isLogging = false
            dismiss()
        }
    }
}

// MARK: - Category Button
struct CategoryButton: View {
    let category: LogExpenseSheet.ExpenseCategory
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: Design.Spacing.xs) {
                Image(systemName: category.icon)
                    .font(.system(size: 11))
                    .foregroundColor(isSelected ? Design.Colors.primary : Design.Colors.textSecondary)

                Text(category.rawValue)
                    .font(Design.Typography.labelSmall)
                    .foregroundColor(isSelected ? Design.Colors.textPrimary : Design.Colors.textSecondary)
                    .lineLimit(1)

                Spacer()
            }
            .padding(Design.Spacing.xs)
            .background(
                RoundedRectangle(cornerRadius: Design.Radius.xs)
                    .fill(isSelected ? Design.Colors.primary.opacity(0.15) : Design.Colors.controlBackground)
                    .overlay(
                        RoundedRectangle(cornerRadius: Design.Radius.xs)
                            .stroke(isSelected ? Design.Colors.primary.opacity(0.5) : Color.clear, lineWidth: 1)
                    )
            )
        }
        .buttonStyle(DSInteractiveStyle())
    }
}

#Preview {
    LogExpenseSheet()
        .environmentObject(AppState())
}
