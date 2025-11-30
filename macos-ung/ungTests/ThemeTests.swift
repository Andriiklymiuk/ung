//
//  ThemeTests.swift
//  ungTests
//
//  Tests for theme system functionality
//

import XCTest
@testable import ung

final class ThemeTests: XCTestCase {

    // MARK: - Theme Enumeration Tests

    func testAllThemesExist() {
        let themes = AppTheme.allCases
        XCTAssertEqual(themes.count, 4)
        XCTAssertTrue(themes.contains(.default))
        XCTAssertTrue(themes.contains(.suit))
        XCTAssertTrue(themes.contains(.unicorn))
        XCTAssertTrue(themes.contains(.greenGoblin))
    }

    func testThemeDisplayNames() {
        XCTAssertEqual(AppTheme.default.displayName, "Default")
        XCTAssertEqual(AppTheme.suit.displayName, "Suit")
        XCTAssertEqual(AppTheme.unicorn.displayName, "Unicorn")
        XCTAssertEqual(AppTheme.greenGoblin.displayName, "Green Goblin")
    }

    func testThemeDescriptions() {
        XCTAssertFalse(AppTheme.default.description.isEmpty)
        XCTAssertFalse(AppTheme.suit.description.isEmpty)
        XCTAssertFalse(AppTheme.unicorn.description.isEmpty)
        XCTAssertFalse(AppTheme.greenGoblin.description.isEmpty)
    }

    func testThemeIcons() {
        XCTAssertEqual(AppTheme.default.icon, "paintpalette")
        XCTAssertEqual(AppTheme.suit.icon, "circle.lefthalf.filled")
        XCTAssertEqual(AppTheme.unicorn.icon, "sparkles")
        XCTAssertEqual(AppTheme.greenGoblin.icon, "leaf.fill")
    }

    func testThemeIdentifiable() {
        for theme in AppTheme.allCases {
            XCTAssertEqual(theme.id, theme.rawValue)
        }
    }

    // MARK: - Theme Colors Tests

    func testDefaultThemeHasColors() {
        let colors = DefaultThemeColors()
        XCTAssertNotNil(colors.brand)
        XCTAssertNotNil(colors.primary)
        XCTAssertNotNil(colors.success)
        XCTAssertNotNil(colors.warning)
        XCTAssertNotNil(colors.error)
    }

    func testSuitThemeHasColors() {
        let colors = SuitThemeColors()
        XCTAssertNotNil(colors.brand)
        XCTAssertNotNil(colors.primary)
        XCTAssertNotNil(colors.success)
        XCTAssertNotNil(colors.warning)
        XCTAssertNotNil(colors.error)
    }

    func testUnicornThemeHasColors() {
        let colors = UnicornThemeColors()
        XCTAssertNotNil(colors.brand)
        XCTAssertNotNil(colors.primary)
        XCTAssertNotNil(colors.success)
        XCTAssertNotNil(colors.warning)
        XCTAssertNotNil(colors.error)
    }

    func testGreenGoblinThemeHasColors() {
        let colors = GreenGoblinThemeColors()
        XCTAssertNotNil(colors.brand)
        XCTAssertNotNil(colors.primary)
        XCTAssertNotNil(colors.success)
        XCTAssertNotNil(colors.warning)
        XCTAssertNotNil(colors.error)
    }

    // MARK: - Theme Feature Colors Tests

    func testDefaultThemeFeatureColors() {
        let colors = DefaultThemeColors()
        XCTAssertNotNil(colors.tracking)
        XCTAssertNotNil(colors.pomodoro)
        XCTAssertNotNil(colors.invoice)
        XCTAssertNotNil(colors.expense)
        XCTAssertNotNil(colors.client)
        XCTAssertNotNil(colors.contract)
    }

    func testSuitThemeFeatureColors() {
        let colors = SuitThemeColors()
        XCTAssertNotNil(colors.tracking)
        XCTAssertNotNil(colors.pomodoro)
        XCTAssertNotNil(colors.invoice)
        XCTAssertNotNil(colors.expense)
        XCTAssertNotNil(colors.client)
        XCTAssertNotNil(colors.contract)
    }

    // MARK: - Preview Colors Tests

    func testThemePreviewColors() {
        for theme in AppTheme.allCases {
            let previewColors = theme.previewColors
            XCTAssertEqual(previewColors.count, 5, "Theme \(theme) should have 5 preview colors")
        }
    }

    // MARK: - Theme Codable Tests

    func testThemeCodable() throws {
        let theme = AppTheme.unicorn
        let encoder = JSONEncoder()
        let data = try encoder.encode(theme)

        let decoder = JSONDecoder()
        let decoded = try decoder.decode(AppTheme.self, from: data)

        XCTAssertEqual(theme, decoded)
    }

    func testAllThemesCodable() throws {
        for theme in AppTheme.allCases {
            let encoder = JSONEncoder()
            let data = try encoder.encode(theme)

            let decoder = JSONDecoder()
            let decoded = try decoder.decode(AppTheme.self, from: data)

            XCTAssertEqual(theme, decoded)
        }
    }

    // MARK: - Theme Colors Protocol Conformance

    func testThemeColorsProtocolConformance() {
        let themes: [ThemeColors] = [
            DefaultThemeColors(),
            SuitThemeColors(),
            UnicornThemeColors(),
            GreenGoblinThemeColors()
        ]

        for themeColors in themes {
            // All themes should have all required colors
            XCTAssertNotNil(themeColors.brand)
            XCTAssertNotNil(themeColors.brandLight)
            XCTAssertNotNil(themeColors.brandDark)
            XCTAssertNotNil(themeColors.primary)
            XCTAssertNotNil(themeColors.primaryLight)
            XCTAssertNotNil(themeColors.primaryMuted)
            XCTAssertNotNil(themeColors.success)
            XCTAssertNotNil(themeColors.successLight)
            XCTAssertNotNil(themeColors.warning)
            XCTAssertNotNil(themeColors.warningLight)
            XCTAssertNotNil(themeColors.error)
            XCTAssertNotNil(themeColors.errorLight)
            XCTAssertNotNil(themeColors.purple)
            XCTAssertNotNil(themeColors.indigo)
            XCTAssertNotNil(themeColors.teal)
            XCTAssertNotNil(themeColors.tracking)
            XCTAssertNotNil(themeColors.pomodoro)
            XCTAssertNotNil(themeColors.invoice)
            XCTAssertNotNil(themeColors.expense)
            XCTAssertNotNil(themeColors.client)
            XCTAssertNotNil(themeColors.contract)
        }
    }
}
