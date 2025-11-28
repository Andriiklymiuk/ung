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
        .background(Color(nsColor: .windowBackgroundColor))
    }

    // MARK: - Header
    private var headerSection: some View {
        HStack {
            VStack(alignment: .leading, spacing: 2) {
                Text("Log Expense")
                    .font(.system(size: 14, weight: .semibold))
                Text("Track your business expenses")
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

    // MARK: - Description Input
    private var descriptionInputSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Description")
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.secondary)

            HStack(spacing: 8) {
                Image(systemName: "text.alignleft")
                    .font(.system(size: 12))
                    .foregroundColor(.orange)

                TextField("What did you spend on?", text: $description)
                    .textFieldStyle(.plain)
                    .font(.system(size: 13))
            }
            .padding(10)
            .background(
                RoundedRectangle(cornerRadius: 8)
                    .fill(Color(nsColor: .controlBackgroundColor))
                    .overlay(
                        RoundedRectangle(cornerRadius: 8)
                            .stroke(Color.orange.opacity(0.3), lineWidth: 1)
                    )
            )
        }
    }

    // MARK: - Amount Input
    private var amountInputSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Amount")
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.secondary)

            HStack(spacing: 8) {
                Image(systemName: "dollarsign.circle.fill")
                    .font(.system(size: 12))
                    .foregroundColor(.green)

                TextField("0.00", text: $amountText)
                    .textFieldStyle(.plain)
                    .font(.system(size: 16, weight: .semibold, design: .rounded))
            }
            .padding(10)
            .background(
                RoundedRectangle(cornerRadius: 8)
                    .fill(Color(nsColor: .controlBackgroundColor))
                    .overlay(
                        RoundedRectangle(cornerRadius: 8)
                            .stroke(Color.green.opacity(0.3), lineWidth: 1)
                    )
            )
        }
    }

    // MARK: - Category Selection
    private var categorySelectionSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Category")
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.secondary)

            LazyVGrid(columns: [
                GridItem(.flexible()),
                GridItem(.flexible())
            ], spacing: 8) {
                ForEach(ExpenseCategory.allCases, id: \.self) { category in
                    CategoryButton(
                        category: category,
                        isSelected: selectedCategory == category,
                        action: { selectedCategory = category }
                    )
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

            Button(action: logExpense) {
                HStack(spacing: 6) {
                    if isLogging {
                        ProgressView()
                            .scaleEffect(0.7)
                    } else {
                        Image(systemName: "plus.circle.fill")
                            .font(.system(size: 10))
                    }
                    Text("Log Expense")
                        .font(.system(size: 12, weight: .semibold))
                }
                .foregroundColor(.white)
                .padding(.horizontal, 16)
                .padding(.vertical, 8)
                .background(
                    RoundedRectangle(cornerRadius: 8)
                        .fill(isValid ? Color.orange : Color.gray)
                )
            }
            .buttonStyle(.plain)
            .disabled(!isValid || isLogging)
        }
        .padding(16)
    }

    // MARK: - Actions
    private func logExpense() {
        isLogging = true
        Task {
            let success = await appState.cliService.logExpense(
                description: description,
                amount: amount,
                category: selectedCategory.rawValue.lowercased()
            )
            if success {
                await appState.refreshDashboard()
            }
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
            HStack(spacing: 6) {
                Image(systemName: category.icon)
                    .font(.system(size: 11))
                    .foregroundColor(isSelected ? category.color : .secondary)

                Text(category.rawValue)
                    .font(.system(size: 11))
                    .foregroundColor(isSelected ? .primary : .secondary)
                    .lineLimit(1)

                Spacer()
            }
            .padding(8)
            .background(
                RoundedRectangle(cornerRadius: 6)
                    .fill(isSelected ? category.color.opacity(0.15) : Color(nsColor: .controlBackgroundColor))
                    .overlay(
                        RoundedRectangle(cornerRadius: 6)
                            .stroke(isSelected ? category.color.opacity(0.5) : Color.clear, lineWidth: 1)
                    )
            )
        }
        .buttonStyle(.plain)
    }
}

#Preview {
    LogExpenseSheet()
        .environmentObject(AppState())
}
