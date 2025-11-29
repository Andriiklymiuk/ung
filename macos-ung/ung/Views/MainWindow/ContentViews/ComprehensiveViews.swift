//
//  ComprehensiveViews.swift
//  ung
//
//  Full-featured views with animations and micro-interactions
//

import Combine
import SwiftUI

// MARK: - Animated Card Wrapper
struct AnimatedCard<Content: View>: View {
  let content: Content
  @State private var isHovered = false
  @Environment(\.colorScheme) var colorScheme

  init(@ViewBuilder content: @escaping () -> Content) {
    self.content = content()
  }

  var body: some View {
    content
      .background(
        RoundedRectangle(cornerRadius: Design.Radius.md)
          .fill(Design.Colors.surfaceElevated(colorScheme))
          .shadow(
            color: isHovered ? Design.Shadow.lg.color : Design.Shadow.sm.color,
            radius: isHovered ? Design.Shadow.lg.radius : Design.Shadow.sm.radius,
            y: isHovered ? Design.Shadow.lg.y : Design.Shadow.sm.y
          )
      )
      .scaleEffect(isHovered ? 1.01 : 1.0)
      .animation(Design.Animation.smooth, value: isHovered)
      .onHover { hovering in
        isHovered = hovering
      }
  }
}

// MARK: - Animated List Row
struct AnimatedListRow<Leading: View, Trailing: View>: View {
  let title: String
  var subtitle: String?
  let leading: Leading
  let trailing: Trailing
  var action: (() -> Void)?

  @State private var isHovered = false
  @State private var isPressed = false
  @Environment(\.colorScheme) var colorScheme

  init(
    title: String,
    subtitle: String? = nil,
    @ViewBuilder leading: @escaping () -> Leading,
    @ViewBuilder trailing: @escaping () -> Trailing,
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

        if action != nil {
          Image(systemName: "chevron.right")
            .font(.system(size: 12, weight: .semibold))
            .foregroundColor(Design.Colors.textTertiary)
            .opacity(isHovered ? 1 : 0.5)
        }
      }
      .padding(Design.Spacing.sm)
      .background(
        RoundedRectangle(cornerRadius: Design.Radius.sm)
          .fill(
            isHovered
              ? Design.Colors.backgroundTertiary(colorScheme)
              : Design.Colors.backgroundSecondary(colorScheme))
      )
      .scaleEffect(isPressed ? 0.98 : 1.0)
    }
    .buttonStyle(.plain)
    .disabled(action == nil)
    .onHover { hovering in
      withAnimation(Design.Animation.quick) {
        isHovered = hovering
      }
    }
    .simultaneousGesture(
      DragGesture(minimumDistance: 0)
        .onChanged { _ in isPressed = true }
        .onEnded { _ in isPressed = false }
    )
  }
}

// MARK: - Animated Icon Button
struct AnimatedIconButton: View {
  let icon: String
  let color: Color
  var size: CGFloat = 32
  let action: () -> Void

  @State private var isHovered = false
  @State private var isPressed = false

  var body: some View {
    Button(action: action) {
      Image(systemName: icon)
        .font(.system(size: size * 0.45, weight: .medium))
        .foregroundColor(isHovered ? .white : color)
        .frame(width: size, height: size)
        .background(
          Circle()
            .fill(isHovered ? color : color.opacity(0.1))
        )
        .scaleEffect(isPressed ? 0.9 : 1.0)
    }
    .buttonStyle(.plain)
    .animation(Design.Animation.quick, value: isHovered)
    .animation(Design.Animation.quick, value: isPressed)
    .onHover { hovering in isHovered = hovering }
    .simultaneousGesture(
      DragGesture(minimumDistance: 0)
        .onChanged { _ in isPressed = true }
        .onEnded { _ in isPressed = false }
    )
  }
}

// MARK: - Search Bar
struct SearchBar: View {
  @Binding var text: String
  var placeholder: String = "Search..."
  var onSubmit: (() -> Void)?
  @FocusState private var isFocused: Bool
  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    HStack(spacing: Design.Spacing.xs) {
      Image(systemName: "magnifyingglass")
        .font(.system(size: 14))
        .foregroundColor(isFocused ? Design.Colors.primary : Design.Colors.textTertiary)

      TextField(placeholder, text: $text)
        .textFieldStyle(.plain)
        .font(Design.Typography.bodyMedium)
        .focused($isFocused)
        .onSubmit { onSubmit?() }
        .onKeyPress(.escape) {
          isFocused = false
          return .handled
        }

      if !text.isEmpty {
        Button(action: { text = "" }) {
          Image(systemName: "xmark.circle.fill")
            .font(.system(size: 14))
            .foregroundColor(Design.Colors.textTertiary)
        }
        .buttonStyle(.plain)
        .transition(.scale.combined(with: .opacity))
      }
    }
    .padding(Design.Spacing.sm)
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.sm)
        .fill(Design.Colors.backgroundSecondary(colorScheme))
        .overlay(
          RoundedRectangle(cornerRadius: Design.Radius.sm)
            .stroke(isFocused ? Design.Colors.primary : Design.Colors.border, lineWidth: 1)
        )
    )
    .animation(Design.Animation.quick, value: isFocused)
    .animation(Design.Animation.quick, value: text.isEmpty)
  }

  // Allow external focus control
  func focused(_ condition: Bool) -> some View {
    self.onAppear {
      if condition {
        isFocused = true
      }
    }
  }
}

// MARK: - Segmented Control
struct AnimatedSegmentedControl<T: Hashable>: View {
  @Binding var selection: T
  let options: [(T, String)]
  @Namespace private var animation
  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    HStack(spacing: 4) {
      ForEach(options, id: \.0) { option in
        Button(action: {
          withAnimation(Design.Animation.smooth) {
            selection = option.0
          }
        }) {
          Text(option.1)
            .font(Design.Typography.labelMedium)
            .foregroundColor(
              selection == option.0 ? Design.Colors.textPrimary : Design.Colors.textSecondary
            )
            .padding(.horizontal, Design.Spacing.md)
            .padding(.vertical, Design.Spacing.xs)
            .background(
              Group {
                if selection == option.0 {
                  RoundedRectangle(cornerRadius: Design.Radius.xs)
                    .fill(Design.Colors.surfaceElevated(colorScheme))
                    .shadow(
                      color: Design.Shadow.sm.color, radius: Design.Shadow.sm.radius,
                      y: Design.Shadow.sm.y
                    )
                    .matchedGeometryEffect(id: "segment", in: animation)
                }
              }
            )
        }
        .buttonStyle(.plain)
      }
    }
    .padding(4)
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.sm)
        .fill(Design.Colors.backgroundTertiary(colorScheme))
    )
  }
}

// MARK: - Floating Action Button
struct FloatingActionButton: View {
  let icon: String
  let color: Color
  let action: () -> Void

  @State private var isHovered = false
  @State private var isPressed = false

  var body: some View {
    Button(action: action) {
      Image(systemName: icon)
        .font(.system(size: 20, weight: .semibold))
        .foregroundColor(.white)
        .frame(width: 56, height: 56)
        .background(
          Circle()
            .fill(
              LinearGradient(
                colors: [color, color.opacity(0.8)],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
              )
            )
            .shadow(color: color.opacity(0.4), radius: isHovered ? 16 : 8, y: isHovered ? 8 : 4)
        )
        .scaleEffect(isPressed ? 0.9 : (isHovered ? 1.1 : 1.0))
    }
    .buttonStyle(.plain)
    .animation(Design.Animation.bounce, value: isHovered)
    .animation(Design.Animation.quick, value: isPressed)
    .onHover { hovering in isHovered = hovering }
    .simultaneousGesture(
      DragGesture(minimumDistance: 0)
        .onChanged { _ in isPressed = true }
        .onEnded { _ in isPressed = false }
    )
  }
}

// MARK: - Empty State with Animation
struct AnimatedEmptyState: View {
  let icon: String
  let title: String
  let message: String
  var actionTitle: String?
  var action: (() -> Void)?

  @State private var isAnimating = false

  var body: some View {
    VStack(spacing: Design.Spacing.lg) {
      ZStack {
        Circle()
          .fill(Color.gray.opacity(0.08))
          .frame(width: 100, height: 100)
          .scaleEffect(isAnimating ? 1.1 : 1.0)

        Circle()
          .fill(Color.gray.opacity(0.05))
          .frame(width: 80, height: 80)

        Image(systemName: icon)
          .font(.system(size: 36, weight: .light))
          .foregroundColor(.secondary.opacity(0.5))
          .offset(y: isAnimating ? -2 : 2)
      }
      .animation(.easeInOut(duration: 2).repeatForever(autoreverses: true), value: isAnimating)

      VStack(spacing: Design.Spacing.xs) {
        Text(title)
          .font(Design.Typography.headingSmall)
          .foregroundColor(Design.Colors.textPrimary)

        Text(message)
          .font(Design.Typography.bodyMedium)
          .foregroundColor(Design.Colors.textSecondary)
          .multilineTextAlignment(.center)
          .fixedSize(horizontal: false, vertical: true)
          .frame(maxWidth: 400)
      }

      if let actionTitle = actionTitle, let action = action {
        Button(action: action) {
          Text(actionTitle)
        }
        .buttonStyle(DSPrimaryButtonStyle())
        .padding(.top, Design.Spacing.xs)
      }
    }
    .frame(maxWidth: .infinity, maxHeight: .infinity)
    .onAppear { isAnimating = true }
  }
}

// MARK: - Shimmer Loading Effect
struct ShimmerView: View {
  @State private var isAnimating = false

  var body: some View {
    LinearGradient(
      colors: [
        Color.gray.opacity(0.1),
        Color.gray.opacity(0.2),
        Color.gray.opacity(0.1),
      ],
      startPoint: .leading,
      endPoint: .trailing
    )
    .mask(Rectangle())
    .offset(x: isAnimating ? 200 : -200)
    .animation(.linear(duration: 1.5).repeatForever(autoreverses: false), value: isAnimating)
    .onAppear { isAnimating = true }
  }
}

struct SkeletonRow: View {
  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    HStack(spacing: Design.Spacing.sm) {
      Circle()
        .fill(Design.Colors.backgroundTertiary(colorScheme))
        .frame(width: 40, height: 40)
        .overlay(ShimmerView())
        .clipShape(Circle())

      VStack(alignment: .leading, spacing: Design.Spacing.xs) {
        RoundedRectangle(cornerRadius: 4)
          .fill(Design.Colors.backgroundTertiary(colorScheme))
          .frame(width: 150, height: 14)
          .overlay(ShimmerView())
          .clipShape(RoundedRectangle(cornerRadius: 4))

        RoundedRectangle(cornerRadius: 4)
          .fill(Design.Colors.backgroundTertiary(colorScheme))
          .frame(width: 100, height: 10)
          .overlay(ShimmerView())
          .clipShape(RoundedRectangle(cornerRadius: 4))
      }

      Spacer()
    }
    .padding(Design.Spacing.sm)
  }
}

// MARK: - Action Sheet Menu
struct ActionMenu<Content: View>: View {
  let title: String
  let content: Content
  @Binding var isPresented: Bool
  var onSubmit: (() -> Void)?
  @Environment(\.colorScheme) var colorScheme

  init(
    title: String,
    isPresented: Binding<Bool>,
    onSubmit: (() -> Void)? = nil,
    @ViewBuilder content: @escaping () -> Content
  ) {
    self.title = title
    self._isPresented = isPresented
    self.onSubmit = onSubmit
    self.content = content()
  }

  var body: some View {
    VStack(spacing: 0) {
      // Header
      HStack {
        Text(title)
          .font(Design.Typography.headingSmall)

        Spacer()

        Button(action: { isPresented = false }) {
          Image(systemName: "xmark.circle.fill")
            .font(.system(size: 24))
            .foregroundColor(Design.Colors.textTertiary)
        }
        .buttonStyle(.plain)
        .keyboardShortcut(.escape, modifiers: [])
      }
      .padding(Design.Spacing.md)

      Divider()

      // Content
      content
        .padding(Design.Spacing.md)
    }
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.lg)
        .fill(Design.Colors.surfaceElevated(colorScheme))
        .shadow(
          color: Design.Shadow.lg.color, radius: Design.Shadow.lg.radius, y: Design.Shadow.lg.y)
    )
    .frame(width: 360)
    .onExitCommand { isPresented = false }
  }
}

// MARK: - Confirmation Dialog
struct ConfirmationDialog: View {
  let title: String
  let message: String
  var destructive: Bool = false
  let confirmTitle: String
  let onConfirm: () -> Void
  let onCancel: () -> Void
  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    VStack(spacing: Design.Spacing.lg) {
      // Icon
      ZStack {
        Circle()
          .fill(destructive ? Design.Colors.errorLight : Design.Colors.warningLight)
          .frame(width: 64, height: 64)

        Image(systemName: destructive ? "trash.fill" : "exclamationmark.triangle.fill")
          .font(.system(size: 28))
          .foregroundColor(destructive ? Design.Colors.error : Design.Colors.warning)
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
      }

      // Actions
      HStack(spacing: Design.Spacing.sm) {
        Button("Cancel", action: onCancel)
          .buttonStyle(DSSecondaryButtonStyle())
          .keyboardShortcut(.escape, modifiers: [])

        Button(confirmTitle, action: onConfirm)
          .buttonStyle(
            DSPrimaryButtonStyle(color: destructive ? Design.Colors.error : Design.Colors.primary)
          )
          .keyboardShortcut(.return, modifiers: .command)
      }
    }
    .padding(Design.Spacing.lg)
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.lg)
        .fill(Design.Colors.surfaceElevated(colorScheme))
    )
    .frame(width: 320)
    .onExitCommand(perform: onCancel)
  }
}

// MARK: - Toast Notification Manager
class ToastManager: ObservableObject {
  static let shared = ToastManager()

  @Published var currentToast: ToastData?

  struct ToastData: Identifiable {
    let id = UUID()
    let message: String
    let type: DSToast.ToastType
  }

  func show(_ message: String, type: DSToast.ToastType = .info) {
    withAnimation(Design.Animation.smooth) {
      currentToast = ToastData(message: message, type: type)
    }

    DispatchQueue.main.asyncAfter(deadline: .now() + 3) {
      withAnimation(Design.Animation.smooth) {
        if self.currentToast?.message == message {
          self.currentToast = nil
        }
      }
    }
  }
}

struct ToastContainer: View {
  @ObservedObject var manager = ToastManager.shared

  var body: some View {
    VStack {
      if let toast = manager.currentToast {
        DSToast(
          message: toast.message,
          type: toast.type,
          isPresented: Binding(
            get: { manager.currentToast != nil },
            set: { if !$0 { manager.currentToast = nil } }
          )
        )
        .transition(.move(edge: .top).combined(with: .opacity))
        .padding()
      }

      Spacer()
    }
    .animation(Design.Animation.smooth, value: manager.currentToast?.id)
  }
}

// MARK: - Pull to Refresh
struct RefreshableScrollView<Content: View>: View {
  let onRefresh: () async -> Void
  let content: Content

  init(onRefresh: @escaping () async -> Void, @ViewBuilder content: @escaping () -> Content) {
    self.onRefresh = onRefresh
    self.content = content()
  }

  var body: some View {
    ScrollView {
      content
    }
    .refreshable {
      await onRefresh()
    }
  }
}

// MARK: - Stat Card with Animation
struct AnimatedStatCard: View {
  let title: String
  let value: String
  let icon: String
  let color: Color
  var trend: String?
  var trendUp: Bool = true

  @State private var appeared = false
  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    VStack(alignment: .leading, spacing: Design.Spacing.sm) {
      HStack {
        Image(systemName: icon)
          .font(.system(size: 16))
          .foregroundColor(color)

        Spacer()

        if let trend = trend {
          HStack(spacing: 2) {
            Image(systemName: trendUp ? "arrow.up.right" : "arrow.down.right")
              .font(.system(size: 10, weight: .bold))
            Text(trend)
              .font(Design.Typography.labelSmall)
          }
          .foregroundColor(trendUp ? Design.Colors.success : Design.Colors.error)
          .padding(.horizontal, 6)
          .padding(.vertical, 3)
          .background(
            Capsule()
              .fill(trendUp ? Design.Colors.successLight : Design.Colors.errorLight)
          )
        }
      }

      Text(value)
        .font(Design.Typography.displaySmall)
        .foregroundColor(Design.Colors.textPrimary)
        .opacity(appeared ? 1 : 0)
        .offset(y: appeared ? 0 : 10)

      Text(title)
        .font(Design.Typography.labelSmall)
        .foregroundColor(Design.Colors.textTertiary)
        .textCase(.uppercase)
        .tracking(0.5)
    }
    .padding(Design.Spacing.md)
    .frame(maxWidth: .infinity, alignment: .leading)
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.md)
        .fill(Design.Colors.surfaceElevated(colorScheme))
        .shadow(
          color: Design.Shadow.sm.color, radius: Design.Shadow.sm.radius, y: Design.Shadow.sm.y)
    )
    .drawingGroup() // Performance: Flatten view hierarchy for faster rendering
    .onAppear {
      withAnimation(Design.Animation.smooth.delay(0.1)) {
        appeared = true
      }
    }
  }
}

// MARK: - Form Section
struct FormSection<Content: View>: View {
  let title: String
  let content: Content
  @Environment(\.colorScheme) var colorScheme

  init(
    title: String,
    @ViewBuilder content: @escaping () -> Content
  ) {
    self.title = title
    self.content = content()
  }

  var body: some View {
    VStack(alignment: .leading, spacing: Design.Spacing.sm) {
      Text(title)
        .font(Design.Typography.labelSmall)
        .foregroundColor(Design.Colors.textTertiary)
        .textCase(.uppercase)
        .tracking(0.5)

      content
    }
  }
}

// MARK: - Form Field
struct FormField: View {
  let label: String
  @Binding var text: String
  var placeholder: String = ""
  var isSecure: Bool = false
  @Environment(\.colorScheme) var colorScheme
  @FocusState private var isFocused: Bool

  var body: some View {
    VStack(alignment: .leading, spacing: Design.Spacing.xxs) {
      Text(label)
        .font(Design.Typography.labelSmall)
        .foregroundColor(Design.Colors.textSecondary)

      Group {
        if isSecure {
          SecureField(placeholder, text: $text)
        } else {
          TextField(placeholder, text: $text)
        }
      }
      .textFieldStyle(.plain)
      .font(Design.Typography.bodyMedium)
      .padding(Design.Spacing.sm)
      .background(
        RoundedRectangle(cornerRadius: Design.Radius.sm)
          .fill(Design.Colors.backgroundSecondary(colorScheme))
          .overlay(
            RoundedRectangle(cornerRadius: Design.Radius.sm)
              .stroke(isFocused ? Design.Colors.primary : Design.Colors.border, lineWidth: 1)
          )
      )
      .focused($isFocused)
    }
    .animation(Design.Animation.quick, value: isFocused)
  }
}

// MARK: - Number Stepper
struct NumberStepper: View {
  let label: String
  @Binding var value: Double
  var range: ClosedRange<Double> = 0...1000
  var step: Double = 1
  var format: String = "%.0f"
  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    HStack {
      Text(label)
        .font(Design.Typography.bodyMedium)
        .foregroundColor(Design.Colors.textSecondary)

      Spacer()

      HStack(spacing: Design.Spacing.xs) {
        Button(action: { value = max(range.lowerBound, value - step) }) {
          Image(systemName: "minus")
            .font(.system(size: 12, weight: .bold))
            .frame(width: 28, height: 28)
            .background(Circle().fill(Design.Colors.backgroundTertiary(colorScheme)))
        }
        .buttonStyle(.plain)
        .disabled(value <= range.lowerBound)

        Text(String(format: format, value))
          .font(Design.Typography.monoSmall)
          .frame(width: 60)

        Button(action: { value = min(range.upperBound, value + step) }) {
          Image(systemName: "plus")
            .font(.system(size: 12, weight: .bold))
            .frame(width: 28, height: 28)
            .background(Circle().fill(Design.Colors.backgroundTertiary(colorScheme)))
        }
        .buttonStyle(.plain)
        .disabled(value >= range.upperBound)
      }
    }
  }
}
