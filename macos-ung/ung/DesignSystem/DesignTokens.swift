//
//  DesignTokens.swift
//  ung
//
//  Design System inspired by Uber's Base Design System
//  Clean, professional, accessible, and consistent
//

import SwiftUI

// MARK: - Design Tokens
enum Design {

  // MARK: - Spacing (8pt Grid)
  enum Spacing {
    static let xxxs: CGFloat = 2
    static let xxs: CGFloat = 4
    static let xs: CGFloat = 8
    static let sm: CGFloat = 12
    static let md: CGFloat = 16
    static let lg: CGFloat = 24
    static let xl: CGFloat = 32
    static let xxl: CGFloat = 48
    static let xxxl: CGFloat = 64
  }

  // MARK: - Corner Radius
  enum Radius {
    static let xs: CGFloat = 4
    static let sm: CGFloat = 8
    static let md: CGFloat = 12
    static let lg: CGFloat = 16
    static let xl: CGFloat = 24
    static let full: CGFloat = 9999
  }

  // MARK: - Typography
  enum Typography {
    // Display
    static let displayLarge = Font.system(size: 40, weight: .bold, design: .rounded)
    static let displayMedium = Font.system(size: 32, weight: .bold, design: .rounded)
    static let displaySmall = Font.system(size: 28, weight: .bold, design: .rounded)

    // Headings
    static let headingLarge = Font.system(size: 24, weight: .semibold, design: .default)
    static let headingMedium = Font.system(size: 20, weight: .semibold, design: .default)
    static let headingSmall = Font.system(size: 16, weight: .semibold, design: .default)

    // Body
    static let bodyLarge = Font.system(size: 16, weight: .regular, design: .default)
    static let bodyMedium = Font.system(size: 14, weight: .regular, design: .default)
    static let bodySmall = Font.system(size: 12, weight: .regular, design: .default)

    // Labels
    static let labelLarge = Font.system(size: 14, weight: .medium, design: .default)
    static let labelMedium = Font.system(size: 12, weight: .medium, design: .default)
    static let labelSmall = Font.system(size: 10, weight: .medium, design: .default)

    // Mono (for numbers/timers)
    static let monoLarge = Font.system(size: 32, weight: .bold, design: .monospaced)
    static let monoMedium = Font.system(size: 20, weight: .semibold, design: .monospaced)
    static let monoSmall = Font.system(size: 14, weight: .medium, design: .monospaced)
  }

  // MARK: - Colors (Semantic)
  enum Colors {
    // Brand - Professional blue
    static let brand = Color(red: 0.20, green: 0.45, blue: 0.90)
    static let brandLight = Color(red: 0.20, green: 0.45, blue: 0.90).opacity(0.15)
    static let brandDark = Color(red: 0.15, green: 0.35, blue: 0.75)

    // Primary (using brand color)
    static let primary = brand
    static let primaryLight = brandLight
    static let primaryMuted = brand.opacity(0.6)

    // Accent colors - refined, warmer palette
    static let success = Color(red: 0.20, green: 0.65, blue: 0.45)
    static let successLight = Color(red: 0.20, green: 0.65, blue: 0.45).opacity(0.1)
    static let warning = Color(red: 0.95, green: 0.60, blue: 0.25)
    static let warningLight = Color(red: 0.95, green: 0.60, blue: 0.25).opacity(0.1)
    static let error = Color(red: 0.90, green: 0.35, blue: 0.35)
    static let errorLight = Color(red: 0.90, green: 0.35, blue: 0.35).opacity(0.1)

    // Neutrals
    static func backgroundPrimary(_ scheme: ColorScheme) -> Color {
      scheme == .dark ? Color(white: 0.11) : Color.white
    }

    static func backgroundSecondary(_ scheme: ColorScheme) -> Color {
      scheme == .dark ? Color(white: 0.15) : Color(white: 0.97)
    }

    static func backgroundTertiary(_ scheme: ColorScheme) -> Color {
      scheme == .dark ? Color(white: 0.18) : Color(white: 0.94)
    }

    static func surfaceElevated(_ scheme: ColorScheme) -> Color {
      scheme == .dark ? Color(white: 0.14) : Color.white
    }

    static let textPrimary = Color.primary
    static let textSecondary = Color.secondary
    static let textTertiary = Color.secondary.opacity(0.6)

    static let border = Color.gray.opacity(0.2)
    static let divider = Color.gray.opacity(0.15)

    // Additional accent colors - refined palette
    static let purple = Color(red: 0.55, green: 0.35, blue: 0.70)
    static let indigo = Color(red: 0.35, green: 0.40, blue: 0.75)
    static let teal = brand

    // Feature colors - cohesive with brand
    static let tracking = success
    static let pomodoro = warning
    static let invoice = brand
    static let expense = warning
    static let client = purple
    static let contract = indigo

    // Cross-platform system colors
    static var windowBackground: Color {
      #if os(macOS)
      Color(nsColor: .windowBackgroundColor)
      #else
      Color(uiColor: .systemBackground)
      #endif
    }

    static var controlBackground: Color {
      #if os(macOS)
      Color(nsColor: .controlBackgroundColor)
      #else
      Color(uiColor: .secondarySystemBackground)
      #endif
    }

    static var tertiaryBackground: Color {
      #if os(macOS)
      Color(nsColor: .underPageBackgroundColor)
      #else
      Color(uiColor: .tertiarySystemBackground)
      #endif
    }
  }

  // MARK: - Shadows
  enum Shadow {
    static let sm = (color: Color.black.opacity(0.04), radius: CGFloat(4), y: CGFloat(2))
    static let md = (color: Color.black.opacity(0.06), radius: CGFloat(8), y: CGFloat(4))
    static let lg = (color: Color.black.opacity(0.08), radius: CGFloat(16), y: CGFloat(8))
  }

  // MARK: - Animation
  enum Animation {
    static let quick = SwiftUI.Animation.easeOut(duration: 0.15)
    static let standard = SwiftUI.Animation.easeInOut(duration: 0.25)
    static let smooth = SwiftUI.Animation.spring(response: 0.35, dampingFraction: 0.8)
    static let bounce = SwiftUI.Animation.spring(response: 0.4, dampingFraction: 0.6)
  }

  // MARK: - Icon Sizes
  enum IconSize {
    static let xs: CGFloat = 12
    static let sm: CGFloat = 16
    static let md: CGFloat = 20
    static let lg: CGFloat = 24
    static let xl: CGFloat = 32
  }
}

// MARK: - Reusable Components

// MARK: - Card Component
struct DSCard<Content: View>: View {
  let content: Content
  var padding: CGFloat = Design.Spacing.md
  @Environment(\.colorScheme) var colorScheme

  init(padding: CGFloat = Design.Spacing.md, @ViewBuilder content: () -> Content) {
    self.padding = padding
    self.content = content()
  }

  var body: some View {
    content
      .padding(padding)
      .background(
        RoundedRectangle(cornerRadius: Design.Radius.md)
          .fill(Design.Colors.surfaceElevated(colorScheme))
          .shadow(
            color: Design.Shadow.md.color,
            radius: Design.Shadow.md.radius,
            y: Design.Shadow.md.y
          )
      )
  }
}

// MARK: - Button Styles
struct DSPrimaryButtonStyle: ButtonStyle {
  let color: Color
  let isDisabled: Bool

  init(color: Color = Design.Colors.primary, isDisabled: Bool = false) {
    self.color = color
    self.isDisabled = isDisabled
  }

  func makeBody(configuration: Configuration) -> some View {
    configuration.label
      .font(Design.Typography.labelLarge)
      .foregroundColor(.white)
      .padding(.horizontal, Design.Spacing.lg)
      .padding(.vertical, Design.Spacing.sm)
      .background(
        RoundedRectangle(cornerRadius: Design.Radius.sm)
          .fill(isDisabled ? Color.gray : color)
      )
      .opacity(configuration.isPressed ? 0.85 : 1)
      .scaleEffect(configuration.isPressed ? 0.98 : 1)
      .animation(Design.Animation.quick, value: configuration.isPressed)
  }
}

struct DSSecondaryButtonStyle: ButtonStyle {
  @Environment(\.colorScheme) var colorScheme

  func makeBody(configuration: Configuration) -> some View {
    configuration.label
      .font(Design.Typography.labelLarge)
      .foregroundColor(Design.Colors.textPrimary)
      .padding(.horizontal, Design.Spacing.lg)
      .padding(.vertical, Design.Spacing.sm)
      .background(
        RoundedRectangle(cornerRadius: Design.Radius.sm)
          .fill(Design.Colors.backgroundSecondary(colorScheme))
      )
      .opacity(configuration.isPressed ? 0.85 : 1)
      .animation(Design.Animation.quick, value: configuration.isPressed)
  }
}

struct DSGhostButtonStyle: ButtonStyle {
  let color: Color

  func makeBody(configuration: Configuration) -> some View {
    configuration.label
      .font(Design.Typography.labelMedium)
      .foregroundColor(color)
      .opacity(configuration.isPressed ? 0.6 : 1)
      .animation(Design.Animation.quick, value: configuration.isPressed)
  }
}

// MARK: - Text Field Style
struct DSTextFieldStyle: TextFieldStyle {
  @Environment(\.colorScheme) var colorScheme

  func _body(configuration: TextField<Self._Label>) -> some View {
    configuration
      .font(Design.Typography.bodyMedium)
      .padding(Design.Spacing.sm)
      .background(
        RoundedRectangle(cornerRadius: Design.Radius.sm)
          .fill(Design.Colors.backgroundSecondary(colorScheme))
          .overlay(
            RoundedRectangle(cornerRadius: Design.Radius.sm)
              .stroke(Design.Colors.border, lineWidth: 1)
          )
      )
  }
}

// MARK: - Badge Component
struct DSBadge: View {
  let text: String
  let color: Color
  var size: BadgeSize = .medium

  enum BadgeSize {
    case small, medium, large

    var font: Font {
      switch self {
      case .small: return Design.Typography.labelSmall
      case .medium: return Design.Typography.labelMedium
      case .large: return Design.Typography.labelLarge
      }
    }

    var padding: (h: CGFloat, v: CGFloat) {
      switch self {
      case .small: return (6, 2)
      case .medium: return (8, 4)
      case .large: return (12, 6)
      }
    }
  }

  var body: some View {
    Text(text)
      .font(size.font)
      .foregroundColor(color)
      .padding(.horizontal, size.padding.h)
      .padding(.vertical, size.padding.v)
      .background(
        Capsule()
          .fill(color.opacity(0.15))
      )
  }
}

// MARK: - Empty State Component
struct DSEmptyState: View {
  let icon: String
  let title: String
  let message: String
  var action: (() -> Void)?
  var actionTitle: String?

  var body: some View {
    VStack(spacing: Design.Spacing.lg) {
      // Icon
      ZStack {
        Circle()
          .fill(Color.gray.opacity(0.1))
          .frame(width: 80, height: 80)

        Image(systemName: icon)
          .font(.system(size: 32, weight: .light))
          .foregroundColor(.secondary.opacity(0.6))
      }

      // Text
      VStack(spacing: Design.Spacing.xs) {
        Text(title)
          .font(Design.Typography.headingSmall)
          .foregroundColor(Design.Colors.textPrimary)

        Text(message)
          .font(Design.Typography.bodyMedium)
          .foregroundColor(Design.Colors.textSecondary)
          .multilineTextAlignment(.center)
          .frame(maxWidth: 280)
      }

      // Action button
      if let action = action, let title = actionTitle {
        Button(action: action) {
          Text(title)
        }
        .buttonStyle(DSPrimaryButtonStyle())
        .padding(.top, Design.Spacing.xs)
      }
    }
    .frame(maxWidth: .infinity, maxHeight: .infinity)
    .padding(Design.Spacing.xxl)
  }
}

// MARK: - Loading Indicator
struct DSLoadingIndicator: View {
  var size: CGFloat = 24
  var color: Color = Design.Colors.primary

  @State private var isAnimating = false

  var body: some View {
    Circle()
      .trim(from: 0, to: 0.7)
      .stroke(color, style: StrokeStyle(lineWidth: 3, lineCap: .round))
      .frame(width: size, height: size)
      .rotationEffect(.degrees(isAnimating ? 360 : 0))
      .animation(.linear(duration: 0.8).repeatForever(autoreverses: false), value: isAnimating)
      .onAppear { isAnimating = true }
  }
}

// MARK: - Metric Display
struct DSMetricDisplay: View {
  let value: String
  let label: String
  var icon: String?
  var color: Color = Design.Colors.textPrimary
  var trend: String?
  var trendPositive: Bool = true
  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    VStack(alignment: .leading, spacing: Design.Spacing.xs) {
      // Header
      HStack(spacing: Design.Spacing.xxs) {
        if let icon = icon {
          Image(systemName: icon)
            .font(.system(size: Design.IconSize.sm))
            .foregroundColor(color.opacity(0.8))
        }

        Text(label)
          .font(Design.Typography.labelSmall)
          .foregroundColor(Design.Colors.textTertiary)
          .textCase(.uppercase)
          .tracking(0.5)
      }

      // Value
      HStack(alignment: .firstTextBaseline, spacing: Design.Spacing.xs) {
        Text(value)
          .font(Design.Typography.headingLarge)
          .foregroundColor(color)

        if let trend = trend {
          HStack(spacing: 2) {
            Image(systemName: trendPositive ? "arrow.up.right" : "arrow.down.right")
              .font(.system(size: 10, weight: .semibold))
            Text(trend)
              .font(Design.Typography.labelSmall)
          }
          .foregroundColor(trendPositive ? Design.Colors.success : Design.Colors.error)
        }
      }
    }
  }
}

// MARK: - Section Header
struct DSSectionHeader: View {
  let title: String
  var action: (() -> Void)?
  var actionTitle: String?

  var body: some View {
    HStack {
      Text(title)
        .font(Design.Typography.labelMedium)
        .foregroundColor(Design.Colors.textSecondary)
        .textCase(.uppercase)
        .tracking(0.5)

      Spacer()

      if let action = action, let title = actionTitle {
        Button(action: action) {
          Text(title)
            .font(Design.Typography.labelMedium)
            .foregroundColor(Design.Colors.primary)
        }
        .buttonStyle(.plain)
      }
    }
  }
}

// MARK: - List Row
struct DSListRow<Leading: View, Trailing: View>: View {
  let title: String
  var subtitle: String?
  let leading: Leading
  let trailing: Trailing
  var action: (() -> Void)?
  @Environment(\.colorScheme) var colorScheme

  init(
    title: String,
    subtitle: String? = nil,
    @ViewBuilder leading: () -> Leading,
    @ViewBuilder trailing: () -> Trailing,
    action: (() -> Void)? = nil
  ) {
    self.title = title
    self.subtitle = subtitle
    self.leading = leading()
    self.trailing = trailing()
    self.action = action
  }

  var body: some View {
    Button(action: { action?() }) {
      HStack(spacing: Design.Spacing.sm) {
        leading

        VStack(alignment: .leading, spacing: Design.Spacing.xxxs) {
          Text(title)
            .font(Design.Typography.bodyMedium)
            .foregroundColor(Design.Colors.textPrimary)

          if let subtitle = subtitle {
            Text(subtitle)
              .font(Design.Typography.bodySmall)
              .foregroundColor(Design.Colors.textSecondary)
          }
        }

        Spacer()

        trailing
      }
      .padding(Design.Spacing.sm)
      .background(
        RoundedRectangle(cornerRadius: Design.Radius.sm)
          .fill(Design.Colors.backgroundSecondary(colorScheme))
      )
    }
    .buttonStyle(.plain)
    .disabled(action == nil)
  }
}

// MARK: - Circular Progress
struct DSCircularProgress: View {
  let progress: Double
  var lineWidth: CGFloat = 8
  var size: CGFloat = 120
  var gradientColors: [Color] = [Design.Colors.primary, Design.Colors.primary.opacity(0.6)]
  var showPercentage: Bool = true

  var body: some View {
    ZStack {
      // Background circle
      Circle()
        .stroke(Color.gray.opacity(0.2), lineWidth: lineWidth)

      // Progress circle
      Circle()
        .trim(from: 0, to: min(progress, 1.0))
        .stroke(
          LinearGradient(
            colors: gradientColors, startPoint: .topLeading, endPoint: .bottomTrailing),
          style: StrokeStyle(lineWidth: lineWidth, lineCap: .round)
        )
        .rotationEffect(.degrees(-90))
        .animation(Design.Animation.smooth, value: progress)

      // Percentage
      if showPercentage {
        Text("\(Int(progress * 100))%")
          .font(Design.Typography.headingMedium)
          .foregroundColor(Design.Colors.textPrimary)
      }
    }
    .frame(width: size, height: size)
  }
}

// MARK: - Toast/Snackbar
struct DSToast: View {
  let message: String
  let type: ToastType
  @Binding var isPresented: Bool
  var autoDismiss: Bool = true
  var dismissAfter: Double = 4.0

  enum ToastType {
    case success, error, warning, info

    var color: Color {
      switch self {
      case .success: return Design.Colors.success
      case .error: return Design.Colors.error
      case .warning: return Design.Colors.warning
      case .info: return Design.Colors.primary
      }
    }

    var icon: String {
      switch self {
      case .success: return "checkmark.circle.fill"
      case .error: return "xmark.circle.fill"
      case .warning: return "exclamationmark.triangle.fill"
      case .info: return "info.circle.fill"
      }
    }

    var accessibilityLabel: String {
      switch self {
      case .success: return "Success"
      case .error: return "Error"
      case .warning: return "Warning"
      case .info: return "Information"
      }
    }
  }

  var body: some View {
    HStack(spacing: Design.Spacing.xs) {
      Image(systemName: type.icon)
        .font(Design.Typography.labelMedium)
        .foregroundColor(type.color)

      Text(message)
        .font(Design.Typography.labelMedium)
        .foregroundColor(Design.Colors.textPrimary)
        .lineLimit(2)

      Button(action: { withAnimation(Design.Animation.quick) { isPresented = false } }) {
        Image(systemName: "xmark")
          .font(Design.Typography.labelSmall)
          .foregroundColor(Design.Colors.textTertiary)
          .padding(Design.Spacing.xxs)
      }
      .buttonStyle(.plain)
      .accessibilityLabel("Dismiss notification")
    }
    .padding(.horizontal, Design.Spacing.sm)
    .padding(.vertical, Design.Spacing.xs)
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.sm)
        .fill(.ultraThinMaterial)
        .overlay(
          RoundedRectangle(cornerRadius: Design.Radius.sm)
            .stroke(type.color.opacity(0.5), lineWidth: 1.5)
        )
        .shadow(color: Design.Shadow.md.color, radius: Design.Shadow.md.radius, y: Design.Shadow.md.y)
    )
    .fixedSize()
    .transition(.move(edge: .bottom).combined(with: .opacity))
    .accessibilityElement(children: .combine)
    .accessibilityLabel("\(type.accessibilityLabel): \(message)")
    .accessibilityAddTraits(.isStaticText)
    .onAppear {
      if autoDismiss {
        DispatchQueue.main.asyncAfter(deadline: .now() + dismissAfter) {
          withAnimation(Design.Animation.smooth) {
            isPresented = false
          }
        }
      }
    }
  }
}

// MARK: - Skeleton Loading View
struct DSSkeletonView: View {
  var width: CGFloat? = nil
  var height: CGFloat = 16
  var cornerRadius: CGFloat = Design.Radius.xs
  @State private var isAnimating = false
  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    RoundedRectangle(cornerRadius: cornerRadius)
      .fill(
        LinearGradient(
          colors: [
            Design.Colors.backgroundSecondary(colorScheme),
            Design.Colors.backgroundTertiary(colorScheme),
            Design.Colors.backgroundSecondary(colorScheme)
          ],
          startPoint: .leading,
          endPoint: .trailing
        )
      )
      .frame(width: width, height: height)
      .mask(
        Rectangle()
          .fill(
            LinearGradient(
              colors: [.clear, .white.opacity(0.5), .clear],
              startPoint: .leading,
              endPoint: .trailing
            )
          )
          .offset(x: isAnimating ? 200 : -200)
      )
      .onAppear {
        withAnimation(.linear(duration: 1.5).repeatForever(autoreverses: false)) {
          isAnimating = true
        }
      }
  }
}

// MARK: - Skeleton Card Loading
struct DSSkeletonCard: View {
  var lines: Int = 3
  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    VStack(alignment: .leading, spacing: Design.Spacing.sm) {
      DSSkeletonView(width: 120, height: 12)
      ForEach(0..<lines, id: \.self) { index in
        DSSkeletonView(
          width: index == lines - 1 ? 180 : nil,
          height: 14
        )
      }
    }
    .padding(Design.Spacing.md)
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.md)
        .fill(Design.Colors.surfaceElevated(colorScheme))
        .shadow(
          color: Design.Shadow.sm.color,
          radius: Design.Shadow.sm.radius,
          y: Design.Shadow.sm.y
        )
    )
  }
}

// MARK: - Shimmer Effect Modifier
struct ShimmerModifier: ViewModifier {
  @State private var phase: CGFloat = 0
  var duration: Double = 1.5

  func body(content: Content) -> some View {
    content
      .overlay(
        GeometryReader { geometry in
          LinearGradient(
            colors: [
              .clear,
              .white.opacity(0.2),
              .clear
            ],
            startPoint: .leading,
            endPoint: .trailing
          )
          .frame(width: geometry.size.width * 2)
          .offset(x: -geometry.size.width + (geometry.size.width * 2 * phase))
        }
        .mask(content)
      )
      .onAppear {
        withAnimation(.linear(duration: duration).repeatForever(autoreverses: false)) {
          phase = 1
        }
      }
  }
}

// MARK: - Accessibility Helpers
extension View {
  /// Adds standard accessibility for interactive elements
  func accessibleButton(label: String, hint: String? = nil) -> some View {
    self
      .accessibilityLabel(label)
      .accessibilityHint(hint ?? "")
      .accessibilityAddTraits(.isButton)
  }

  /// Adds accessibility for static text elements
  func accessibleText(label: String) -> some View {
    self
      .accessibilityLabel(label)
      .accessibilityAddTraits(.isStaticText)
  }

  /// Adds accessibility for header elements
  func accessibleHeader(label: String) -> some View {
    self
      .accessibilityLabel(label)
      .accessibilityAddTraits(.isHeader)
  }

  /// Adds shimmer loading effect
  func shimmer(duration: Double = 1.5) -> some View {
    modifier(ShimmerModifier(duration: duration))
  }
}

// MARK: - View Extensions
extension View {
  func dsCard(padding: CGFloat = Design.Spacing.md) -> some View {
    modifier(DSCardModifier(padding: padding))
  }
}

struct DSCardModifier: ViewModifier {
  let padding: CGFloat
  @Environment(\.colorScheme) var colorScheme

  func body(content: Content) -> some View {
    content
      .padding(padding)
      .background(
        RoundedRectangle(cornerRadius: Design.Radius.md)
          .fill(Design.Colors.surfaceElevated(colorScheme))
          .shadow(
            color: Design.Shadow.md.color,
            radius: Design.Shadow.md.radius,
            y: Design.Shadow.md.y
          )
      )
  }
}
