//
//  WidgetDesignSystem.swift
//  ungWidgets
//
//  Premium design system inspired by Telegram's aesthetic
//

import SwiftUI
import WidgetKit

// MARK: - Widget Color Palette
enum WidgetColors {
    // Primary gradients
    static let trackingGradient = LinearGradient(
        colors: [Color(hex: "FF6B6B"), Color(hex: "EE5A5A")],
        startPoint: .topLeading,
        endPoint: .bottomTrailing
    )

    static let focusGradient = LinearGradient(
        colors: [Color(hex: "FF9F43"), Color(hex: "F7B731")],
        startPoint: .topLeading,
        endPoint: .bottomTrailing
    )

    static let breakGradient = LinearGradient(
        colors: [Color(hex: "26DE81"), Color(hex: "20BF6B")],
        startPoint: .topLeading,
        endPoint: .bottomTrailing
    )

    static let statsGradient = LinearGradient(
        colors: [Color(hex: "4834D4"), Color(hex: "686DE0")],
        startPoint: .topLeading,
        endPoint: .bottomTrailing
    )

    static let invoiceGradient = LinearGradient(
        colors: [Color(hex: "00D2D3"), Color(hex: "54A0FF")],
        startPoint: .topLeading,
        endPoint: .bottomTrailing
    )

    // Accent colors
    static let trackingRed = Color(hex: "FF6B6B")
    static let focusOrange = Color(hex: "FF9F43")
    static let breakGreen = Color(hex: "26DE81")
    static let statsPurple = Color(hex: "4834D4")
    static let invoiceCyan = Color(hex: "00D2D3")

    // Neutral palette
    static let cardBackground = Color(hex: "1C1C1E")
    static let cardBackgroundLight = Color(hex: "F2F2F7")
    static let textPrimary = Color.primary
    static let textSecondary = Color.secondary
    static let textTertiary = Color(hex: "8E8E93")

    // Glassmorphism
    static func glassBackground(_ scheme: ColorScheme) -> some View {
        RoundedRectangle(cornerRadius: 16)
            .fill(.ultraThinMaterial)
            .overlay(
                RoundedRectangle(cornerRadius: 16)
                    .stroke(
                        LinearGradient(
                            colors: [
                                Color.white.opacity(scheme == .dark ? 0.15 : 0.3),
                                Color.white.opacity(scheme == .dark ? 0.05 : 0.1)
                            ],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        ),
                        lineWidth: 0.5
                    )
            )
    }
}

// MARK: - Color Hex Extension
extension Color {
    init(hex: String) {
        let hex = hex.trimmingCharacters(in: CharacterSet.alphanumerics.inverted)
        var int: UInt64 = 0
        Scanner(string: hex).scanHexInt64(&int)
        let a, r, g, b: UInt64
        switch hex.count {
        case 3: // RGB (12-bit)
            (a, r, g, b) = (255, (int >> 8) * 17, (int >> 4 & 0xF) * 17, (int & 0xF) * 17)
        case 6: // RGB (24-bit)
            (a, r, g, b) = (255, int >> 16, int >> 8 & 0xFF, int & 0xFF)
        case 8: // ARGB (32-bit)
            (a, r, g, b) = (int >> 24, int >> 16 & 0xFF, int >> 8 & 0xFF, int & 0xFF)
        default:
            (a, r, g, b) = (1, 1, 1, 0)
        }
        self.init(
            .sRGB,
            red: Double(r) / 255,
            green: Double(g) / 255,
            blue: Double(b) / 255,
            opacity: Double(a) / 255
        )
    }
}

// MARK: - Premium Widgets Components

struct PremiumProgressRing: View {
    let progress: Double
    let lineWidth: CGFloat
    let gradient: LinearGradient
    let size: CGFloat

    var body: some View {
        ZStack {
            // Background ring
            Circle()
                .stroke(Color.gray.opacity(0.15), lineWidth: lineWidth)
                .frame(width: size, height: size)

            // Progress ring
            Circle()
                .trim(from: 0, to: min(progress, 1.0))
                .stroke(
                    gradient,
                    style: StrokeStyle(lineWidth: lineWidth, lineCap: .round)
                )
                .frame(width: size, height: size)
                .rotationEffect(.degrees(-90))

            // Glow effect
            Circle()
                .trim(from: 0, to: min(progress, 1.0))
                .stroke(
                    gradient,
                    style: StrokeStyle(lineWidth: lineWidth * 2, lineCap: .round)
                )
                .frame(width: size, height: size)
                .rotationEffect(.degrees(-90))
                .blur(radius: lineWidth)
                .opacity(0.4)
        }
    }
}

struct PremiumIconBadge: View {
    let icon: String
    let gradient: LinearGradient
    let size: CGFloat
    var isActive: Bool = false

    var body: some View {
        ZStack {
            // Glow when active
            if isActive {
                Circle()
                    .fill(gradient)
                    .frame(width: size * 1.5, height: size * 1.5)
                    .blur(radius: size * 0.4)
                    .opacity(0.5)
            }

            // Background circle
            Circle()
                .fill(gradient)
                .frame(width: size, height: size)

            // Icon
            Image(systemName: icon)
                .font(.system(size: size * 0.45, weight: .semibold))
                .foregroundColor(.white)
        }
    }
}

struct PremiumStatCard: View {
    let icon: String
    let title: String
    let value: String
    let gradient: LinearGradient
    let compact: Bool

    init(icon: String, title: String, value: String, gradient: LinearGradient, compact: Bool = false) {
        self.icon = icon
        self.title = title
        self.value = value
        self.gradient = gradient
        self.compact = compact
    }

    var body: some View {
        HStack(spacing: compact ? 8 : 12) {
            // Icon
            ZStack {
                Circle()
                    .fill(gradient.opacity(0.15))
                    .frame(width: compact ? 28 : 36, height: compact ? 28 : 36)

                Image(systemName: icon)
                    .font(.system(size: compact ? 12 : 14, weight: .semibold))
                    .foregroundStyle(gradient)
            }

            VStack(alignment: .leading, spacing: 1) {
                Text(title)
                    .font(.system(size: compact ? 10 : 11, weight: .medium))
                    .foregroundColor(WidgetColors.textTertiary)

                Text(value)
                    .font(.system(size: compact ? 14 : 16, weight: .bold))
                    .foregroundColor(WidgetColors.textPrimary)
            }
        }
    }
}

struct PulsingDot: View {
    let color: Color
    let size: CGFloat

    var body: some View {
        ZStack {
            // Outer glow
            Circle()
                .fill(color.opacity(0.3))
                .frame(width: size * 2, height: size * 2)

            // Inner dot
            Circle()
                .fill(color)
                .frame(width: size, height: size)
        }
    }
}

struct MonospacedTimer: View {
    let time: String
    let size: CGFloat
    let color: Color

    var body: some View {
        Text(time)
            .font(.system(size: size, weight: .bold, design: .monospaced))
            .foregroundColor(color)
            .monospacedDigit()
            .shadow(color: color.opacity(0.3), radius: 4, x: 0, y: 2)
    }
}

// MARK: - Session Dots
struct SessionDots: View {
    let completed: Int
    let total: Int
    let activeColor: Color

    var body: some View {
        HStack(spacing: 4) {
            ForEach(0..<total, id: \.self) { index in
                RoundedRectangle(cornerRadius: 2)
                    .fill(index < completed % total ? activeColor : Color.gray.opacity(0.2))
                    .frame(width: 16, height: 4)
            }
        }
    }
}
