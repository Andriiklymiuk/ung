//
//  Theme.swift
//  ung
//
//  Theme system for customizable app appearance
//

import SwiftUI

// MARK: - Theme Definition
enum AppTheme: String, CaseIterable, Identifiable, Codable {
    case `default` = "Default"
    case suit = "Suit"
    case unicorn = "Unicorn"
    case greenGoblin = "Green Goblin"

    var id: String { rawValue }

    var displayName: String { rawValue }

    var description: String {
        switch self {
        case .default:
            return "Professional blue palette"
        case .suit:
            return "Elegant black, white & grey"
        case .unicorn:
            return "Colorful rainbow palette"
        case .greenGoblin:
            return "Fresh green tones"
        }
    }

    var icon: String {
        switch self {
        case .default:
            return "paintpalette"
        case .suit:
            return "circle.lefthalf.filled"
        case .unicorn:
            return "sparkles"
        case .greenGoblin:
            return "leaf.fill"
        }
    }

    // MARK: - Color Palette
    var colors: ThemeColors {
        switch self {
        case .default:
            return DefaultThemeColors()
        case .suit:
            return SuitThemeColors()
        case .unicorn:
            return UnicornThemeColors()
        case .greenGoblin:
            return GreenGoblinThemeColors()
        }
    }
}

// MARK: - Theme Colors Protocol
protocol ThemeColors {
    // Brand colors
    var brand: Color { get }
    var brandLight: Color { get }
    var brandDark: Color { get }

    // Primary colors
    var primary: Color { get }
    var primaryLight: Color { get }
    var primaryMuted: Color { get }

    // Semantic colors
    var success: Color { get }
    var successLight: Color { get }
    var warning: Color { get }
    var warningLight: Color { get }
    var error: Color { get }
    var errorLight: Color { get }

    // Accent colors
    var purple: Color { get }
    var indigo: Color { get }
    var teal: Color { get }

    // Feature colors
    var tracking: Color { get }
    var pomodoro: Color { get }
    var invoice: Color { get }
    var expense: Color { get }
    var client: Color { get }
    var contract: Color { get }

    // Background variations
    func backgroundPrimary(_ scheme: ColorScheme) -> Color
    func backgroundSecondary(_ scheme: ColorScheme) -> Color
    func backgroundTertiary(_ scheme: ColorScheme) -> Color
    func surfaceElevated(_ scheme: ColorScheme) -> Color
}

// MARK: - Default Theme (Professional Blue)
struct DefaultThemeColors: ThemeColors {
    let brand = Color(red: 0.20, green: 0.45, blue: 0.90)
    var brandLight: Color { brand.opacity(0.15) }
    let brandDark = Color(red: 0.15, green: 0.35, blue: 0.75)

    var primary: Color { brand }
    var primaryLight: Color { brandLight }
    var primaryMuted: Color { brand.opacity(0.6) }

    let success = Color(red: 0.20, green: 0.65, blue: 0.45)
    var successLight: Color { success.opacity(0.1) }
    let warning = Color(red: 0.95, green: 0.60, blue: 0.25)
    var warningLight: Color { warning.opacity(0.1) }
    let error = Color(red: 0.90, green: 0.35, blue: 0.35)
    var errorLight: Color { error.opacity(0.1) }

    let purple = Color(red: 0.55, green: 0.35, blue: 0.70)
    let indigo = Color(red: 0.35, green: 0.40, blue: 0.75)
    var teal: Color { brand }

    var tracking: Color { success }
    var pomodoro: Color { warning }
    var invoice: Color { brand }
    var expense: Color { warning }
    var client: Color { purple }
    var contract: Color { indigo }

    func backgroundPrimary(_ scheme: ColorScheme) -> Color {
        scheme == .dark ? Color(white: 0.11) : Color.white
    }

    func backgroundSecondary(_ scheme: ColorScheme) -> Color {
        scheme == .dark ? Color(white: 0.15) : Color(white: 0.97)
    }

    func backgroundTertiary(_ scheme: ColorScheme) -> Color {
        scheme == .dark ? Color(white: 0.18) : Color(white: 0.94)
    }

    func surfaceElevated(_ scheme: ColorScheme) -> Color {
        scheme == .dark ? Color(white: 0.14) : Color.white
    }
}

// MARK: - Suit Theme (Black, White, Grey - Uber-like)
struct SuitThemeColors: ThemeColors {
    let brand = Color(white: 0.10)
    var brandLight: Color { Color(white: 0.95) }
    let brandDark = Color.black

    var primary: Color { brand }
    var primaryLight: Color { brandLight }
    var primaryMuted: Color { Color(white: 0.40) }

    let success = Color(white: 0.20)
    var successLight: Color { Color(white: 0.92) }
    let warning = Color(white: 0.35)
    var warningLight: Color { Color(white: 0.90) }
    let error = Color(red: 0.85, green: 0.15, blue: 0.15)
    var errorLight: Color { error.opacity(0.1) }

    let purple = Color(white: 0.30)
    let indigo = Color(white: 0.25)
    var teal: Color { Color(white: 0.20) }

    var tracking: Color { Color(white: 0.15) }
    var pomodoro: Color { Color(white: 0.25) }
    var invoice: Color { brand }
    var expense: Color { Color(white: 0.30) }
    var client: Color { Color(white: 0.20) }
    var contract: Color { Color(white: 0.25) }

    func backgroundPrimary(_ scheme: ColorScheme) -> Color {
        scheme == .dark ? Color(white: 0.08) : Color.white
    }

    func backgroundSecondary(_ scheme: ColorScheme) -> Color {
        scheme == .dark ? Color(white: 0.12) : Color(white: 0.96)
    }

    func backgroundTertiary(_ scheme: ColorScheme) -> Color {
        scheme == .dark ? Color(white: 0.15) : Color(white: 0.92)
    }

    func surfaceElevated(_ scheme: ColorScheme) -> Color {
        scheme == .dark ? Color(white: 0.10) : Color.white
    }
}

// MARK: - Unicorn Theme (Rainbow Palette)
struct UnicornThemeColors: ThemeColors {
    let brand = Color(red: 0.85, green: 0.35, blue: 0.65) // Pink
    var brandLight: Color { brand.opacity(0.15) }
    let brandDark = Color(red: 0.70, green: 0.25, blue: 0.55)

    var primary: Color { brand }
    var primaryLight: Color { brandLight }
    var primaryMuted: Color { brand.opacity(0.6) }

    let success = Color(red: 0.30, green: 0.80, blue: 0.55) // Mint green
    var successLight: Color { success.opacity(0.15) }
    let warning = Color(red: 1.0, green: 0.75, blue: 0.30) // Golden yellow
    var warningLight: Color { warning.opacity(0.15) }
    let error = Color(red: 1.0, green: 0.40, blue: 0.45) // Coral red
    var errorLight: Color { error.opacity(0.15) }

    let purple = Color(red: 0.70, green: 0.45, blue: 0.95) // Violet
    let indigo = Color(red: 0.45, green: 0.50, blue: 0.95) // Indigo
    let teal = Color(red: 0.30, green: 0.80, blue: 0.85) // Cyan

    var tracking: Color { teal }
    var pomodoro: Color { warning }
    var invoice: Color { purple }
    var expense: Color { Color(red: 1.0, green: 0.55, blue: 0.40) } // Coral
    var client: Color { indigo }
    var contract: Color { Color(red: 0.55, green: 0.75, blue: 0.95) } // Sky blue

    func backgroundPrimary(_ scheme: ColorScheme) -> Color {
        scheme == .dark ? Color(red: 0.10, green: 0.08, blue: 0.14) : Color.white
    }

    func backgroundSecondary(_ scheme: ColorScheme) -> Color {
        scheme == .dark ? Color(red: 0.14, green: 0.12, blue: 0.18) : Color(red: 0.98, green: 0.97, blue: 1.0)
    }

    func backgroundTertiary(_ scheme: ColorScheme) -> Color {
        scheme == .dark ? Color(red: 0.18, green: 0.15, blue: 0.22) : Color(red: 0.96, green: 0.95, blue: 0.98)
    }

    func surfaceElevated(_ scheme: ColorScheme) -> Color {
        scheme == .dark ? Color(red: 0.12, green: 0.10, blue: 0.16) : Color.white
    }
}

// MARK: - Green Goblin Theme (Green Tones)
struct GreenGoblinThemeColors: ThemeColors {
    let brand = Color(red: 0.20, green: 0.70, blue: 0.45) // Emerald
    var brandLight: Color { brand.opacity(0.15) }
    let brandDark = Color(red: 0.15, green: 0.55, blue: 0.35)

    var primary: Color { brand }
    var primaryLight: Color { brandLight }
    var primaryMuted: Color { brand.opacity(0.6) }

    let success = Color(red: 0.30, green: 0.75, blue: 0.50) // Bright green
    var successLight: Color { success.opacity(0.15) }
    let warning = Color(red: 0.85, green: 0.70, blue: 0.30) // Olive yellow
    var warningLight: Color { warning.opacity(0.15) }
    let error = Color(red: 0.85, green: 0.35, blue: 0.35)
    var errorLight: Color { error.opacity(0.15) }

    let purple = Color(red: 0.35, green: 0.55, blue: 0.45) // Forest teal
    let indigo = Color(red: 0.25, green: 0.50, blue: 0.55) // Deep teal
    let teal = Color(red: 0.20, green: 0.65, blue: 0.55) // Teal green

    var tracking: Color { Color(red: 0.25, green: 0.80, blue: 0.50) } // Lime
    var pomodoro: Color { Color(red: 0.50, green: 0.70, blue: 0.30) } // Chartreuse
    var invoice: Color { brand }
    var expense: Color { Color(red: 0.45, green: 0.65, blue: 0.40) } // Sage
    var client: Color { teal }
    var contract: Color { Color(red: 0.30, green: 0.60, blue: 0.50) } // Sea green

    func backgroundPrimary(_ scheme: ColorScheme) -> Color {
        scheme == .dark ? Color(red: 0.08, green: 0.12, blue: 0.10) : Color.white
    }

    func backgroundSecondary(_ scheme: ColorScheme) -> Color {
        scheme == .dark ? Color(red: 0.10, green: 0.16, blue: 0.13) : Color(red: 0.97, green: 0.99, blue: 0.97)
    }

    func backgroundTertiary(_ scheme: ColorScheme) -> Color {
        scheme == .dark ? Color(red: 0.13, green: 0.19, blue: 0.15) : Color(red: 0.94, green: 0.98, blue: 0.95)
    }

    func surfaceElevated(_ scheme: ColorScheme) -> Color {
        scheme == .dark ? Color(red: 0.09, green: 0.14, blue: 0.11) : Color.white
    }
}

// MARK: - Theme Manager
@MainActor
class ThemeManager: ObservableObject {
    static let shared = ThemeManager()

    @Published var currentTheme: AppTheme {
        didSet {
            saveTheme()
        }
    }

    private let themeKey = "selectedAppTheme"

    private init() {
        if let savedTheme = UserDefaults.standard.string(forKey: themeKey),
           let theme = AppTheme(rawValue: savedTheme) {
            self.currentTheme = theme
        } else {
            self.currentTheme = .default
        }
    }

    private func saveTheme() {
        UserDefaults.standard.set(currentTheme.rawValue, forKey: themeKey)
    }

    var colors: ThemeColors {
        currentTheme.colors
    }
}

// MARK: - Environment Key
struct ThemeEnvironmentKey: EnvironmentKey {
    static let defaultValue: ThemeColors = DefaultThemeColors()
}

extension EnvironmentValues {
    var themeColors: ThemeColors {
        get { self[ThemeEnvironmentKey.self] }
        set { self[ThemeEnvironmentKey.self] = newValue }
    }
}

// MARK: - Theme Preview Colors
extension AppTheme {
    var previewColors: [Color] {
        let colors = self.colors
        return [colors.brand, colors.success, colors.warning, colors.purple, colors.teal]
    }
}
