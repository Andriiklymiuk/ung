//
//  RecentItemsViews.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI

// MARK: - Recent Sessions Card
struct RecentSessionsCard: View {
    @EnvironmentObject var appState: AppState

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: "clock.fill")
                    .font(.system(size: 11))
                    .foregroundColor(.blue)
                Text("Recent Sessions")
                    .font(.system(size: 12, weight: .semibold))
                Spacer()
            }

            if appState.recentSessions.isEmpty {
                EmptyStateView(
                    icon: "clock",
                    message: "No sessions yet"
                )
            } else {
                VStack(spacing: 4) {
                    ForEach(appState.recentSessions) { session in
                        SessionRow(session: session)
                    }
                }
            }
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color(nsColor: .controlBackgroundColor))
        )
    }
}

struct SessionRow: View {
    @EnvironmentObject var appState: AppState
    let session: RecentSession

    var body: some View {
        HStack(spacing: 8) {
            Circle()
                .fill(Color.blue.opacity(0.3))
                .frame(width: 6, height: 6)

            Text(session.project)
                .font(.system(size: 11))
                .foregroundColor(.primary)
                .lineLimit(1)

            Spacer()

            Text(appState.secureMode ? "**:**" : session.duration)
                .font(.system(size: 10, weight: .medium, design: .monospaced))
                .foregroundColor(.secondary)

            Text(session.date)
                .font(.system(size: 9))
                .foregroundColor(.secondary)
        }
        .padding(.vertical, 4)
    }
}

// MARK: - Recent Invoices Card
struct RecentInvoicesCard: View {
    @EnvironmentObject var appState: AppState

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: "doc.text.fill")
                    .font(.system(size: 11))
                    .foregroundColor(.green)
                Text("Recent Invoices")
                    .font(.system(size: 12, weight: .semibold))
                Spacer()
            }

            if appState.recentInvoices.isEmpty {
                EmptyStateView(
                    icon: "doc.text",
                    message: "No invoices yet"
                )
            } else {
                VStack(spacing: 4) {
                    ForEach(appState.recentInvoices) { invoice in
                        InvoiceRow(invoice: invoice)
                    }
                }
            }
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color(nsColor: .controlBackgroundColor))
        )
    }
}

struct InvoiceRow: View {
    @EnvironmentObject var appState: AppState
    let invoice: RecentInvoice

    var statusColor: Color {
        switch invoice.status.lowercased() {
        case "paid": return .green
        case "sent", "pending": return .orange
        case "overdue": return .red
        default: return .gray
        }
    }

    var body: some View {
        HStack(spacing: 8) {
            // Status dot
            Circle()
                .fill(statusColor)
                .frame(width: 6, height: 6)

            // Invoice number
            Text(invoice.invoiceNum)
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.primary)

            // Client
            Text(invoice.client)
                .font(.system(size: 10))
                .foregroundColor(.secondary)
                .lineLimit(1)

            Spacer()

            // Amount
            Text(appState.secureMode ? "****" : invoice.amount)
                .font(.system(size: 10, weight: .semibold))
                .foregroundColor(statusColor)

            // Status badge
            Text(invoice.status.capitalized)
                .font(.system(size: 8, weight: .medium))
                .foregroundColor(statusColor)
                .padding(.horizontal, 5)
                .padding(.vertical, 2)
                .background(statusColor.opacity(0.15))
                .cornerRadius(4)
        }
        .padding(.vertical, 4)
    }
}

// MARK: - Recent Expenses Card
struct RecentExpensesCard: View {
    @EnvironmentObject var appState: AppState

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: "dollarsign.circle.fill")
                    .font(.system(size: 11))
                    .foregroundColor(.orange)
                Text("Recent Expenses")
                    .font(.system(size: 12, weight: .semibold))
                Spacer()
            }

            if appState.recentExpenses.isEmpty {
                EmptyStateView(
                    icon: "dollarsign.circle",
                    message: "No expenses yet"
                )
            } else {
                VStack(spacing: 4) {
                    ForEach(appState.recentExpenses) { expense in
                        ExpenseRow(expense: expense)
                    }
                }
            }
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color(nsColor: .controlBackgroundColor))
        )
    }
}

struct ExpenseRow: View {
    @EnvironmentObject var appState: AppState
    let expense: RecentExpense

    var categoryIcon: String {
        switch expense.category.lowercased() {
        case "software": return "app.badge.fill"
        case "hardware": return "desktopcomputer"
        case "travel": return "airplane"
        case "meals": return "fork.knife"
        case "office": return "building.2.fill"
        case "utilities": return "bolt.fill"
        case "marketing": return "megaphone.fill"
        default: return "tag.fill"
        }
    }

    var body: some View {
        HStack(spacing: 8) {
            Image(systemName: categoryIcon)
                .font(.system(size: 10))
                .foregroundColor(.orange)
                .frame(width: 16)

            Text(expense.description)
                .font(.system(size: 11))
                .foregroundColor(.primary)
                .lineLimit(1)

            Spacer()

            Text(appState.secureMode ? "****" : expense.amount)
                .font(.system(size: 10, weight: .semibold))
                .foregroundColor(.orange)
        }
        .padding(.vertical, 4)
    }
}

// MARK: - Empty State View
struct EmptyStateView: View {
    let icon: String
    let message: String

    var body: some View {
        HStack(spacing: 8) {
            Image(systemName: icon)
                .font(.system(size: 14))
                .foregroundColor(.secondary.opacity(0.5))

            Text(message)
                .font(.system(size: 11))
                .foregroundColor(.secondary)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 12)
    }
}

#Preview {
    VStack(spacing: 12) {
        RecentSessionsCard()
        RecentInvoicesCard()
        RecentExpensesCard()
    }
    .padding()
    .environmentObject(AppState())
}
