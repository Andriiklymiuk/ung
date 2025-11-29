//
//  NotificationSettingsView.swift
//  ung
//
//  Settings UI for notification preferences
//

import SwiftUI

struct NotificationSettingsView: View {
    @ObservedObject var notificationService = NotificationService.shared
    @State private var showAuthAlert = false

    var body: some View {
        ScrollView {
            VStack(spacing: 20) {
                // Authorization Status
                authorizationSection

                if notificationService.isAuthorized {
                    // Pomodoro Notifications
                    pomodoroSection

                    // Tracking Notifications
                    trackingSection

                    // Invoice Notifications
                    invoiceSection

                    // Weekly Goal Notifications
                    weeklyGoalSection

                    // Daily Summary
                    dailySummarySection

                    // Quiet Hours
                    quietHoursSection
                }
            }
            .padding(20)
        }
        .navigationTitle("Notifications")
        .onChange(of: notificationService.settings) { _, _ in
            notificationService.saveSettings()
        }
    }

    // MARK: - Authorization Section
    private var authorizationSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Image(systemName: notificationService.isAuthorized ? "bell.badge.fill" : "bell.slash.fill")
                    .font(.system(size: 24))
                    .foregroundColor(notificationService.isAuthorized ? .green : .orange)

                VStack(alignment: .leading, spacing: 2) {
                    Text("Notification Status")
                        .font(.system(size: 14, weight: .semibold))

                    Text(notificationService.isAuthorized ? "Notifications are enabled" : "Notifications are disabled")
                        .font(.system(size: 12))
                        .foregroundColor(.secondary)
                }

                Spacer()

                if !notificationService.isAuthorized {
                    Button("Enable") {
                        Task {
                            let granted = await notificationService.requestAuthorization()
                            if !granted {
                                showAuthAlert = true
                            }
                        }
                    }
                    .buttonStyle(.borderedProminent)
                }
            }
            .padding(16)
            .background(
                RoundedRectangle(cornerRadius: 12)
                    .fill(notificationService.isAuthorized ? Color.green.opacity(0.1) : Color.orange.opacity(0.1))
            )
        }
        .alert("Notifications Disabled", isPresented: $showAuthAlert) {
            Button("Open Settings") {
                #if os(iOS)
                if let url = URL(string: UIApplication.openSettingsURLString) {
                    UIApplication.shared.open(url)
                }
                #elseif os(macOS)
                if let url = URL(string: "x-apple.systempreferences:com.apple.preference.notifications") {
                    NSWorkspace.shared.open(url)
                }
                #endif
            }
            Button("Cancel", role: .cancel) {}
        } message: {
            Text("Please enable notifications in System Settings to receive alerts.")
        }
    }

    // MARK: - Pomodoro Section
    private var pomodoroSection: some View {
        SettingsSection(title: "Pomodoro Timer", icon: "brain.head.profile", color: .orange) {
            Toggle("Enable Pomodoro Notifications", isOn: $notificationService.settings.pomodoroEnabled)
                .font(.system(size: 13))

            if notificationService.settings.pomodoroEnabled {
                Toggle("Play Sound", isOn: $notificationService.settings.pomodoroSound)
                    .font(.system(size: 13))
                    .foregroundColor(.secondary)
            }

            Text("Get notified when focus sessions and breaks complete.")
                .font(.system(size: 11))
                .foregroundColor(.secondary)
        }
    }

    // MARK: - Tracking Section
    private var trackingSection: some View {
        SettingsSection(title: "Time Tracking", icon: "clock.fill", color: .blue) {
            Toggle("Enable Tracking Reminders", isOn: $notificationService.settings.trackingRemindersEnabled)
                .font(.system(size: 13))

            if notificationService.settings.trackingRemindersEnabled {
                HStack {
                    Text("Reminder interval")
                        .font(.system(size: 13))
                        .foregroundColor(.secondary)
                    Spacer()
                    Picker("", selection: $notificationService.settings.trackingReminderInterval) {
                        Text("30 min").tag(30)
                        Text("1 hour").tag(60)
                        Text("2 hours").tag(120)
                        Text("4 hours").tag(240)
                    }
                    .pickerStyle(.menu)
                    .frame(width: 100)
                }

                HStack {
                    Text("Long session alert after")
                        .font(.system(size: 13))
                        .foregroundColor(.secondary)
                    Spacer()
                    Picker("", selection: $notificationService.settings.longSessionAlertMinutes) {
                        Text("1 hour").tag(60)
                        Text("2 hours").tag(120)
                        Text("3 hours").tag(180)
                        Text("4 hours").tag(240)
                    }
                    .pickerStyle(.menu)
                    .frame(width: 100)
                }
            }
        }
    }

    // MARK: - Invoice Section
    private var invoiceSection: some View {
        SettingsSection(title: "Invoices", icon: "doc.text.fill", color: .green) {
            Toggle("Enable Invoice Reminders", isOn: $notificationService.settings.invoiceRemindersEnabled)
                .font(.system(size: 13))

            if notificationService.settings.invoiceRemindersEnabled {
                HStack {
                    Text("Remind before due date")
                        .font(.system(size: 13))
                        .foregroundColor(.secondary)
                    Spacer()
                    Picker("", selection: $notificationService.settings.invoiceDueDaysBefore) {
                        Text("1 day").tag(1)
                        Text("3 days").tag(3)
                        Text("5 days").tag(5)
                        Text("7 days").tag(7)
                    }
                    .pickerStyle(.menu)
                    .frame(width: 100)
                }
            }

            Text("Get notified about upcoming and overdue invoices.")
                .font(.system(size: 11))
                .foregroundColor(.secondary)
        }
    }

    // MARK: - Weekly Goal Section
    private var weeklyGoalSection: some View {
        SettingsSection(title: "Weekly Goals", icon: "target", color: .purple) {
            Toggle("Enable Goal Notifications", isOn: $notificationService.settings.weeklyGoalRemindersEnabled)
                .font(.system(size: 13))

            Text("Get notified when you're close to or achieve your weekly hours goal.")
                .font(.system(size: 11))
                .foregroundColor(.secondary)
        }
    }

    // MARK: - Daily Summary Section
    private var dailySummarySection: some View {
        SettingsSection(title: "Daily Summary", icon: "chart.bar.fill", color: .cyan) {
            Toggle("Enable Daily Summary", isOn: $notificationService.settings.dailySummaryEnabled)
                .font(.system(size: 13))

            if notificationService.settings.dailySummaryEnabled {
                HStack {
                    Text("Summary time")
                        .font(.system(size: 13))
                        .foregroundColor(.secondary)
                    Spacer()
                    DatePicker("", selection: $notificationService.settings.dailySummaryTime, displayedComponents: .hourAndMinute)
                        .labelsHidden()
                        .frame(width: 100)
                }
            }

            Text("Receive a summary of your tracked time at the end of each day.")
                .font(.system(size: 11))
                .foregroundColor(.secondary)
        }
    }

    // MARK: - Quiet Hours Section
    private var quietHoursSection: some View {
        SettingsSection(title: "Quiet Hours", icon: "moon.fill", color: .indigo) {
            Toggle("Enable Quiet Hours", isOn: $notificationService.settings.quietHoursEnabled)
                .font(.system(size: 13))

            if notificationService.settings.quietHoursEnabled {
                HStack {
                    Text("From")
                        .font(.system(size: 13))
                        .foregroundColor(.secondary)
                    Spacer()
                    DatePicker("", selection: $notificationService.settings.quietHoursStart, displayedComponents: .hourAndMinute)
                        .labelsHidden()
                        .frame(width: 80)

                    Text("to")
                        .font(.system(size: 13))
                        .foregroundColor(.secondary)

                    DatePicker("", selection: $notificationService.settings.quietHoursEnd, displayedComponents: .hourAndMinute)
                        .labelsHidden()
                        .frame(width: 80)
                }
            }

            Text("No notifications will be sent during quiet hours.")
                .font(.system(size: 11))
                .foregroundColor(.secondary)
        }
    }
}

// MARK: - Settings Section Component
struct SettingsSection<Content: View>: View {
    let title: String
    let icon: String
    let color: Color
    @ViewBuilder let content: () -> Content

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack(spacing: 8) {
                Image(systemName: icon)
                    .font(.system(size: 14))
                    .foregroundColor(color)

                Text(title)
                    .font(.system(size: 14, weight: .semibold))
            }

            VStack(alignment: .leading, spacing: 10) {
                content()
            }
            .padding(.leading, 22)
        }
        .padding(16)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(
            RoundedRectangle(cornerRadius: 12)
                .fill(Design.Colors.controlBackground)
        )
    }
}

#Preview {
    NotificationSettingsView()
}
