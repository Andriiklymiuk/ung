//
//  ungWidgetsBundle.swift
//  ungWidgets
//
//  Widget extension for iOS and macOS widgets
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
            TrackingLiveActivity()
            PomodoroLiveActivity()
        }
        #endif
    }
}
