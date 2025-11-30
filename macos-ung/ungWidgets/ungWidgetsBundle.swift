//
//  ungWidgetsBundle.swift
//  ungWidgets
//
//  Widget extension for iOS and macOS widgets
//  Enhanced with premium Live Activities for Dynamic Island
//

import SwiftUI
import WidgetKit

#if canImport(ActivityKit)
import ActivityKit
#endif

@main
struct ungWidgetsBundle: WidgetBundle {
    var body: some Widget {
        // Standard widgets - work on iOS 14+ and macOS 14+
        TrackingStatusWidget()
        WeeklyStatsWidget()
        PomodoroWidget()
        QuickActionsWidget()

        // Live Activities - iOS 16.1+ only
        #if os(iOS)
        if #available(iOS 16.1, *) {
            // Core tracking activities
            TrackingLiveActivity()
            PomodoroLiveActivity()

            // Premium feature activities
            MilestoneLiveActivity()      // Earnings milestone celebrations
            InvoiceStatusLiveActivity()  // Invoice tracking with urgency
            DailySummaryLiveActivity()   // Daily/weekly achievements
            BreakReminderLiveActivity()  // Health-focused break reminders
        }
        #endif
    }
}
