//
//  ungWidgetsBundle.swift
//  ungWidgets
//
//  Widget extension for iOS home screen widgets
//

import SwiftUI
import WidgetKit

@main
struct ungWidgetsBundle: WidgetBundle {
    var body: some Widget {
        TrackingStatusWidget()
        WeeklyStatsWidget()
        PomodoroWidget()
        QuickActionsWidget()
    }
}
