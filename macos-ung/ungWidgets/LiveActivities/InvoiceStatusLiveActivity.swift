//
//  InvoiceStatusLiveActivity.swift
//  ungWidgets
//
//  Premium Live Activity for tracking invoice status with urgency indicators
//  "Never miss a payment deadline" - Marketing hook
//

#if os(iOS)
import ActivityKit
import SwiftUI
import WidgetKit

// MARK: - Invoice Status Activity Attributes
@available(iOS 16.1, *)
struct InvoiceStatusActivityAttributes: ActivityAttributes {
    public struct ContentState: Codable, Hashable {
        var status: String // "sent", "viewed", "paid", "overdue"
        var viewedAt: Date?
        var paidAt: Date?
        var daysUntilDue: Int
    }

    var invoiceNumber: String
    var clientName: String
    var amount: Double
    var currency: String
    var dueDate: Date
    var sentAt: Date
}

// MARK: - Invoice Status Live Activity Widget
@available(iOS 16.1, *)
struct InvoiceStatusLiveActivity: Widget {
    var body: some WidgetConfiguration {
        ActivityConfiguration(for: InvoiceStatusActivityAttributes.self) { context in
            // Lock Screen / Banner view
            InvoiceLockScreenView(context: context)
                .activityBackgroundTint(backgroundTint(for: context.state.status))
                .activitySystemActionForegroundColor(Color.white)
        } dynamicIsland: { context in
            DynamicIsland {
                DynamicIslandExpandedRegion(.leading) {
                    InvoiceExpandedLeading(context: context)
                }

                DynamicIslandExpandedRegion(.trailing) {
                    InvoiceExpandedTrailing(context: context)
                }

                DynamicIslandExpandedRegion(.center) {
                    InvoiceExpandedCenter(context: context)
                }

                DynamicIslandExpandedRegion(.bottom) {
                    InvoiceExpandedBottom(context: context)
                }
            } compactLeading: {
                InvoiceCompactLeading(context: context)
            } compactTrailing: {
                InvoiceCompactTrailing(context: context)
            } minimal: {
                InvoiceMinimalView(context: context)
            }
            .widgetURL(URL(string: "ung://invoices"))
            .keylineTint(keylineTint(for: context.state.status))
        }
    }

    private func backgroundTint(for status: String) -> Color {
        switch status {
        case "paid": return Color.green.opacity(0.85)
        case "overdue": return Color.red.opacity(0.85)
        case "viewed": return Color.blue.opacity(0.85)
        default: return Color.black.opacity(0.85)
        }
    }

    private func keylineTint(for status: String) -> Color {
        switch status {
        case "paid": return .green
        case "overdue": return .red
        case "viewed": return .blue
        default: return .cyan
        }
    }
}

// MARK: - Invoice Theme
@available(iOS 16.1, *)
private struct InvoiceTheme {
    let status: String
    let daysUntilDue: Int

    var primaryColor: Color {
        switch status {
        case "paid": return .green
        case "overdue": return .red
        case "viewed": return .blue
        default:
            // Urgency based on days remaining
            if daysUntilDue <= 1 { return .red }
            if daysUntilDue <= 3 { return .orange }
            return .cyan
        }
    }

    var gradient: LinearGradient {
        switch status {
        case "paid":
            return LinearGradient(colors: [.green, .mint], startPoint: .leading, endPoint: .trailing)
        case "overdue":
            return LinearGradient(colors: [.red, .orange], startPoint: .leading, endPoint: .trailing)
        case "viewed":
            return LinearGradient(colors: [.blue, .cyan], startPoint: .leading, endPoint: .trailing)
        default:
            if daysUntilDue <= 1 {
                return LinearGradient(colors: [.red, .orange], startPoint: .leading, endPoint: .trailing)
            }
            if daysUntilDue <= 3 {
                return LinearGradient(colors: [.orange, .yellow], startPoint: .leading, endPoint: .trailing)
            }
            return LinearGradient(colors: [.cyan, .blue], startPoint: .leading, endPoint: .trailing)
        }
    }

    var icon: String {
        switch status {
        case "paid": return "checkmark.circle.fill"
        case "overdue": return "exclamationmark.triangle.fill"
        case "viewed": return "eye.fill"
        default: return "doc.text.fill"
        }
    }

    var statusLabel: String {
        switch status {
        case "paid": return "PAID"
        case "overdue": return "OVERDUE"
        case "viewed": return "VIEWED"
        default: return "SENT"
        }
    }

    var statusEmoji: String {
        switch status {
        case "paid": return "✓"
        case "overdue": return "!"
        case "viewed": return "👁"
        default: return "📤"
        }
    }
}

// MARK: - Lock Screen View
@available(iOS 16.1, *)
private struct InvoiceLockScreenView: View {
    let context: ActivityViewContext<InvoiceStatusActivityAttributes>

    private var theme: InvoiceTheme {
        InvoiceTheme(status: context.state.status, daysUntilDue: context.state.daysUntilDue)
    }

    var body: some View {
        HStack(spacing: 16) {
            // Status indicator
            InvoiceStatusIndicator(status: context.state.status, daysUntilDue: context.state.daysUntilDue)

            // Invoice info
            VStack(alignment: .leading, spacing: 6) {
                HStack(spacing: 8) {
                    Text("Invoice #\(context.attributes.invoiceNumber)")
                        .font(.system(size: 17, weight: .semibold))
                        .foregroundColor(.white)

                    StatusBadge(status: context.state.status, daysUntilDue: context.state.daysUntilDue)
                }

                HStack(spacing: 8) {
                    Image(systemName: "building.2.fill")
                        .font(.system(size: 10))
                        .foregroundColor(.gray)
                    Text(context.attributes.clientName)
                        .font(.system(size: 13))
                        .foregroundColor(.gray)
                        .lineLimit(1)
                }
            }

            Spacer()

            // Amount and due info
            VStack(alignment: .trailing, spacing: 4) {
                Text("\(context.attributes.currency)\(context.attributes.amount, specifier: "%.2f")")
                    .font(.system(size: 24, weight: .bold, design: .rounded))
                    .foregroundStyle(theme.gradient)

                if context.state.status == "paid" {
                    Text("Payment received")
                        .font(.system(size: 11))
                        .foregroundColor(.green)
                } else if context.state.status == "overdue" {
                    Text("\(abs(context.state.daysUntilDue)) days overdue")
                        .font(.system(size: 11))
                        .foregroundColor(.red)
                } else if context.state.daysUntilDue == 0 {
                    Text("Due today!")
                        .font(.system(size: 11))
                        .foregroundColor(.orange)
                } else if context.state.daysUntilDue == 1 {
                    Text("Due tomorrow")
                        .font(.system(size: 11))
                        .foregroundColor(.orange)
                } else {
                    Text("Due in \(context.state.daysUntilDue) days")
                        .font(.system(size: 11))
                        .foregroundColor(.gray)
                }
            }
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 14)
    }
}

// MARK: - Status Indicator
@available(iOS 16.1, *)
private struct InvoiceStatusIndicator: View {
    let status: String
    let daysUntilDue: Int

    private var theme: InvoiceTheme {
        InvoiceTheme(status: status, daysUntilDue: daysUntilDue)
    }

    var body: some View {
        ZStack {
            // Glow effect for urgent statuses
            if status == "overdue" || status == "paid" {
                Circle()
                    .fill(
                        RadialGradient(
                            colors: [theme.primaryColor.opacity(0.4), theme.primaryColor.opacity(0)],
                            center: .center,
                            startRadius: 15,
                            endRadius: 35
                        )
                    )
                    .frame(width: 70, height: 70)
            }

            // Background circle
            Circle()
                .fill(theme.primaryColor.opacity(0.2))
                .frame(width: 56, height: 56)

            // Progress ring (for non-paid)
            if status != "paid" {
                Circle()
                    .stroke(theme.primaryColor.opacity(0.3), lineWidth: 4)
                    .frame(width: 56, height: 56)

                // Urgency progress
                let urgencyProgress: Double
                if status == "overdue" {
                    urgencyProgress = 1.0
                } else {
                    urgencyProgress = max(0, 1.0 - (Double(daysUntilDue) / 14.0))
                }

                Circle()
                    .trim(from: 0, to: urgencyProgress)
                    .stroke(theme.primaryColor, style: StrokeStyle(lineWidth: 4, lineCap: .round))
                    .frame(width: 56, height: 56)
                    .rotationEffect(.degrees(-90))
            }

            // Center icon
            Image(systemName: theme.icon)
                .font(.system(size: 22))
                .foregroundColor(theme.primaryColor)
        }
    }
}

// MARK: - Status Badge
@available(iOS 16.1, *)
private struct StatusBadge: View {
    let status: String
    let daysUntilDue: Int

    private var theme: InvoiceTheme {
        InvoiceTheme(status: status, daysUntilDue: daysUntilDue)
    }

    var body: some View {
        Text(theme.statusLabel)
            .font(.system(size: 9, weight: .bold))
            .foregroundColor(theme.primaryColor)
            .padding(.horizontal, 6)
            .padding(.vertical, 2)
            .background(theme.primaryColor.opacity(0.2))
            .cornerRadius(4)
    }
}

// MARK: - Dynamic Island Expanded Views
@available(iOS 16.1, *)
private struct InvoiceExpandedLeading: View {
    let context: ActivityViewContext<InvoiceStatusActivityAttributes>
    private var theme: InvoiceTheme {
        InvoiceTheme(status: context.state.status, daysUntilDue: context.state.daysUntilDue)
    }

    var body: some View {
        HStack(spacing: 8) {
            ZStack {
                Circle()
                    .fill(theme.primaryColor.opacity(0.2))
                    .frame(width: 28, height: 28)

                Image(systemName: theme.icon)
                    .font(.system(size: 14))
                    .foregroundColor(theme.primaryColor)
            }

            VStack(alignment: .leading, spacing: 2) {
                Text("#\(context.attributes.invoiceNumber)")
                    .font(.system(size: 14, weight: .semibold))

                Text(context.attributes.clientName)
                    .font(.system(size: 10))
                    .foregroundColor(.secondary)
                    .lineLimit(1)
            }
        }
    }
}

@available(iOS 16.1, *)
private struct InvoiceExpandedTrailing: View {
    let context: ActivityViewContext<InvoiceStatusActivityAttributes>
    private var theme: InvoiceTheme {
        InvoiceTheme(status: context.state.status, daysUntilDue: context.state.daysUntilDue)
    }

    var body: some View {
        VStack(alignment: .trailing, spacing: 2) {
            Text("\(context.attributes.currency)\(context.attributes.amount, specifier: "%.0f")")
                .font(.system(size: 20, weight: .bold, design: .rounded))
                .foregroundStyle(theme.gradient)

            Text(theme.statusLabel)
                .font(.system(size: 9, weight: .bold))
                .foregroundColor(theme.primaryColor)
        }
    }
}

@available(iOS 16.1, *)
private struct InvoiceExpandedCenter: View {
    let context: ActivityViewContext<InvoiceStatusActivityAttributes>
    private var theme: InvoiceTheme {
        InvoiceTheme(status: context.state.status, daysUntilDue: context.state.daysUntilDue)
    }

    var body: some View {
        if context.state.status == "viewed" {
            HStack(spacing: 4) {
                Image(systemName: "eye.fill")
                    .font(.system(size: 10))
                Text("Client viewed invoice")
                    .font(.system(size: 10))
            }
            .foregroundColor(.blue)
        } else if context.state.status == "overdue" {
            HStack(spacing: 4) {
                Image(systemName: "exclamationmark.triangle.fill")
                    .font(.system(size: 10))
                Text("PAYMENT OVERDUE")
                    .font(.system(size: 10, weight: .bold))
            }
            .foregroundColor(.red)
        }
    }
}

@available(iOS 16.1, *)
private struct InvoiceExpandedBottom: View {
    let context: ActivityViewContext<InvoiceStatusActivityAttributes>
    private var theme: InvoiceTheme {
        InvoiceTheme(status: context.state.status, daysUntilDue: context.state.daysUntilDue)
    }

    var body: some View {
        VStack(spacing: 8) {
            // Timeline
            InvoiceTimeline(context: context)

            // Due date info
            HStack {
                HStack(spacing: 4) {
                    Image(systemName: "calendar")
                        .font(.system(size: 10))
                    Text("Sent \(context.attributes.sentAt, style: .date)")
                        .font(.system(size: 10))
                }
                .foregroundColor(.secondary)

                Spacer()

                HStack(spacing: 4) {
                    Image(systemName: "clock.fill")
                        .font(.system(size: 10))
                    if context.state.status == "paid" {
                        Text("Paid")
                            .font(.system(size: 10))
                            .foregroundColor(.green)
                    } else {
                        Text("Due \(context.attributes.dueDate, style: .date)")
                            .font(.system(size: 10))
                            .foregroundColor(context.state.daysUntilDue <= 3 ? .orange : .secondary)
                    }
                }
            }
            .padding(.horizontal, 4)
        }
    }
}

// MARK: - Invoice Timeline
@available(iOS 16.1, *)
private struct InvoiceTimeline: View {
    let context: ActivityViewContext<InvoiceStatusActivityAttributes>

    private var steps: [(icon: String, label: String, completed: Bool, active: Bool)] {
        let status = context.state.status
        return [
            ("paperplane.fill", "Sent", true, status == "sent"),
            ("eye.fill", "Viewed", status == "viewed" || status == "paid", status == "viewed"),
            ("checkmark.circle.fill", "Paid", status == "paid", status == "paid")
        ]
    }

    var body: some View {
        HStack(spacing: 0) {
            ForEach(Array(steps.enumerated()), id: \.offset) { index, step in
                // Step indicator
                VStack(spacing: 4) {
                    ZStack {
                        Circle()
                            .fill(step.completed ? (step.active ? Color.green : Color.gray) : Color.gray.opacity(0.3))
                            .frame(width: 20, height: 20)

                        Image(systemName: step.icon)
                            .font(.system(size: 10))
                            .foregroundColor(step.completed ? .white : .gray)
                    }

                    Text(step.label)
                        .font(.system(size: 8))
                        .foregroundColor(step.active ? .primary : .secondary)
                }

                // Connector line
                if index < steps.count - 1 {
                    Rectangle()
                        .fill(steps[index + 1].completed ? Color.gray : Color.gray.opacity(0.3))
                        .frame(height: 2)
                        .frame(maxWidth: .infinity)
                        .padding(.horizontal, 4)
                        .offset(y: -8)
                }
            }
        }
        .padding(.horizontal, 8)
    }
}

// MARK: - Compact Views
@available(iOS 16.1, *)
private struct InvoiceCompactLeading: View {
    let context: ActivityViewContext<InvoiceStatusActivityAttributes>
    private var theme: InvoiceTheme {
        InvoiceTheme(status: context.state.status, daysUntilDue: context.state.daysUntilDue)
    }

    var body: some View {
        ZStack {
            Circle()
                .fill(theme.primaryColor.opacity(0.3))
                .frame(width: 20, height: 20)

            Image(systemName: theme.icon)
                .font(.system(size: 10))
                .foregroundColor(theme.primaryColor)
        }
    }
}

@available(iOS 16.1, *)
private struct InvoiceCompactTrailing: View {
    let context: ActivityViewContext<InvoiceStatusActivityAttributes>
    private var theme: InvoiceTheme {
        InvoiceTheme(status: context.state.status, daysUntilDue: context.state.daysUntilDue)
    }

    var body: some View {
        if context.state.status == "paid" {
            Image(systemName: "checkmark")
                .font(.system(size: 12, weight: .bold))
                .foregroundColor(.green)
        } else if context.state.status == "overdue" {
            Text("!")
                .font(.system(size: 13, weight: .bold))
                .foregroundColor(.red)
        } else {
            Text("\(context.state.daysUntilDue)d")
                .font(.system(size: 12, weight: .bold))
                .foregroundStyle(theme.gradient)
        }
    }
}

// MARK: - Minimal View
@available(iOS 16.1, *)
private struct InvoiceMinimalView: View {
    let context: ActivityViewContext<InvoiceStatusActivityAttributes>
    private var theme: InvoiceTheme {
        InvoiceTheme(status: context.state.status, daysUntilDue: context.state.daysUntilDue)
    }

    var body: some View {
        ZStack {
            // Background
            Circle()
                .fill(
                    RadialGradient(
                        colors: [theme.primaryColor.opacity(0.3), Color.clear],
                        center: .center,
                        startRadius: 2,
                        endRadius: 14
                    )
                )

            // Icon based on status
            if context.state.status == "paid" {
                Image(systemName: "checkmark")
                    .font(.system(size: 12, weight: .bold))
                    .foregroundColor(.green)
            } else if context.state.status == "overdue" {
                Image(systemName: "exclamationmark")
                    .font(.system(size: 12, weight: .bold))
                    .foregroundColor(.red)
            } else {
                Image(systemName: "doc.text.fill")
                    .font(.system(size: 10))
                    .foregroundColor(theme.primaryColor)
            }
        }
    }
}
#endif
