//
//  WelcomeWalkthroughView.swift
//  ung
//
//  Classic multi-page walkthrough for first-time users
//

import SwiftUI

// MARK: - Welcome Walkthrough View
struct WelcomeWalkthroughView: View {
  @EnvironmentObject var appState: AppState
  @State private var currentPage = 0
  @State private var isAnimating = false
  @Namespace private var animation

  // Streamlined 3-page walkthrough (research shows 3-4 pages optimal)
  private let pages: [WalkthroughPage] = [
    WalkthroughPage(
      icon: "lightbulb.max.fill",
      iconColors: [.yellow, .orange],
      title: "Dig Into Your Ideas",
      subtitle: "Validate before you build",
      features: [
        FeatureItem(icon: "brain.head.profile", text: "AI analyzes from 5+ expert perspectives", color: .purple),
        FeatureItem(icon: "chart.line.uptrend.xyaxis", text: "Get revenue projections & market fit", color: .green),
        FeatureItem(icon: "doc.text.fill", text: "LLM-ready prompt to help you build", color: .blue),
      ]
    ),
    WalkthroughPage(
      icon: "clock.badge.checkmark.fill",
      iconColors: [.blue, .purple],
      title: "Track Your Time",
      subtitle: "Log billable hours effortlessly",
      features: [
        FeatureItem(icon: "play.circle.fill", text: "One-click time tracking", color: .green),
        FeatureItem(icon: "chart.bar.fill", text: "Weekly & monthly reports", color: .blue),
        FeatureItem(icon: "target", text: "Set income goals", color: .orange),
      ]
    ),
    WalkthroughPage(
      icon: "doc.text.fill",
      iconColors: [.teal, .cyan],
      title: "Invoice & Get Paid",
      subtitle: "Professional invoices in seconds",
      features: [
        FeatureItem(icon: "wand.and.stars", text: "Auto-generate from tracked time", color: .purple),
        FeatureItem(icon: "checkmark.circle.fill", text: "Track payment status", color: .green),
        FeatureItem(icon: "rectangle.3.group.fill", text: "Organize gigs on kanban board", color: .pink),
      ]
    ),
    WalkthroughPage(
      icon: "sparkles",
      iconColors: [.green, .teal],
      title: "You're All Set!",
      subtitle: "Your freelance toolkit is ready",
      features: [
        FeatureItem(icon: "lock.shield.fill", text: "Your data stays private", color: .green),
        FeatureItem(icon: "icloud.fill", text: "Optional iCloud sync", color: .blue),
        FeatureItem(icon: "brain.head.profile", text: "Focus timer included", color: .orange),
      ],
      isLastPage: true
    ),
  ]

  var body: some View {
    GeometryReader { geometry in
      ZStack {
        // Background gradient
        LinearGradient(
          colors: [
            Color(pages[currentPage].iconColors[0]).opacity(0.1),
            Color(pages[currentPage].iconColors[1]).opacity(0.05),
          ],
          startPoint: .topLeading,
          endPoint: .bottomTrailing
        )
        .animation(.easeInOut(duration: 0.5), value: currentPage)
        .ignoresSafeArea()

        VStack(spacing: 0) {
          // Page content
          TabView(selection: $currentPage) {
            ForEach(0..<pages.count, id: \.self) { index in
              WalkthroughPageView(page: pages[index], isAnimating: currentPage == index)
                .tag(index)
            }
          }
          #if os(macOS)
          .tabViewStyle(.automatic)
          #else
          .tabViewStyle(.page(indexDisplayMode: .never))
          #endif

          // Bottom controls
          VStack(spacing: 20) {
            // Page indicator with counter
            VStack(spacing: 8) {
              HStack(spacing: 8) {
                ForEach(0..<pages.count, id: \.self) { index in
                  Circle()
                    .fill(currentPage == index ? pages[currentPage].iconColors[0] : Color.secondary.opacity(0.3))
                    .frame(width: currentPage == index ? 10 : 8, height: currentPage == index ? 10 : 8)
                    .animation(.spring(response: 0.3), value: currentPage)
                    .onTapGesture {
                      withAnimation { currentPage = index }
                    }
                    .accessibilityLabel("Page \(index + 1) of \(pages.count)")
                }
              }
              Text("\(currentPage + 1) of \(pages.count)")
                .font(.caption)
                .foregroundColor(.secondary)
            }
            .padding(.bottom, 8)

            // Buttons
            HStack(spacing: 16) {
              if currentPage > 0 {
                Button(action: {
                  withAnimation(.spring(response: 0.4)) {
                    currentPage -= 1
                  }
                }) {
                  HStack {
                    Image(systemName: "chevron.left")
                    Text("Back")
                  }
                  .frame(width: 100)
                }
                .buttonStyle(.plain)
                .foregroundColor(.secondary)
              } else {
                Spacer().frame(width: 100)
              }

              Spacer()

              if pages[currentPage].isLastPage {
                Button(action: {
                  appState.completeWelcomeWalkthrough()
                }) {
                  HStack {
                    Text("Get Started")
                    Image(systemName: "arrow.right")
                  }
                  .font(.headline)
                  .foregroundColor(.white)
                  .frame(width: 160, height: 44)
                  .background(
                    LinearGradient(
                      colors: pages[currentPage].iconColors,
                      startPoint: .leading,
                      endPoint: .trailing
                    )
                  )
                  .cornerRadius(22)
                  .shadow(color: pages[currentPage].iconColors[0].opacity(0.4), radius: 8, y: 4)
                }
                .buttonStyle(.plain)
              } else {
                Button(action: {
                  withAnimation(.spring(response: 0.4)) {
                    currentPage += 1
                  }
                }) {
                  HStack {
                    Text("Next")
                    Image(systemName: "chevron.right")
                  }
                  .font(.headline)
                  .foregroundColor(.white)
                  .frame(width: 120, height: 44)
                  .background(
                    LinearGradient(
                      colors: pages[currentPage].iconColors,
                      startPoint: .leading,
                      endPoint: .trailing
                    )
                  )
                  .cornerRadius(22)
                  .shadow(color: pages[currentPage].iconColors[0].opacity(0.4), radius: 8, y: 4)
                }
                .buttonStyle(.plain)
              }
            }
            .padding(.horizontal, 40)

            // Skip button (except on last page)
            if !pages[currentPage].isLastPage {
              Button(action: {
                appState.completeWelcomeWalkthrough()
              }) {
                Text("Skip")
                  .font(.subheadline)
                  .foregroundColor(.secondary)
              }
              .buttonStyle(.plain)
              .padding(.top, 8)
            }
          }
          .padding(.bottom, 40)
        }
      }
    }
    .frame(minWidth: 500, minHeight: 600)
  }
}

// MARK: - Walkthrough Page Model
struct WalkthroughPage {
  let icon: String
  let iconColors: [Color]
  let title: String
  let subtitle: String
  let features: [FeatureItem]
  var isLastPage: Bool = false
}

struct FeatureItem {
  let icon: String
  let text: String
  let color: Color
}

// MARK: - Walkthrough Page View
struct WalkthroughPageView: View {
  let page: WalkthroughPage
  let isAnimating: Bool
  @State private var showFeatures = false

  var body: some View {
    VStack(spacing: 24) {
      Spacer()

      // Animated icon
      ZStack {
        // Outer glow
        Circle()
          .fill(
            RadialGradient(
              colors: [page.iconColors[0].opacity(0.3), .clear],
              center: .center,
              startRadius: 0,
              endRadius: 80
            )
          )
          .frame(width: 160, height: 160)
          .scaleEffect(isAnimating ? 1.2 : 1.0)
          .animation(.easeInOut(duration: 2).repeatForever(autoreverses: true), value: isAnimating)

        // Icon background
        Circle()
          .fill(
            LinearGradient(
              colors: page.iconColors,
              startPoint: .topLeading,
              endPoint: .bottomTrailing
            )
          )
          .frame(width: 100, height: 100)
          .shadow(color: page.iconColors[0].opacity(0.5), radius: 20, y: 10)

        // Icon
        Image(systemName: page.icon)
          .font(.system(size: 44, weight: .medium))
          .foregroundColor(.white)
          .symbolEffect(.pulse, options: .repeating, value: isAnimating)
      }
      .padding(.bottom, 16)

      // Title
      Text(page.title)
        .font(.system(size: 32, weight: .bold, design: .rounded))
        .foregroundColor(.primary)
        .multilineTextAlignment(.center)

      // Subtitle
      Text(page.subtitle)
        .font(.system(size: 17))
        .foregroundColor(.secondary)
        .multilineTextAlignment(.center)
        .padding(.horizontal, 40)

      // Feature list
      VStack(spacing: 16) {
        ForEach(Array(page.features.enumerated()), id: \.offset) { index, feature in
          HStack(spacing: 16) {
            ZStack {
              Circle()
                .fill(feature.color.opacity(0.15))
                .frame(width: 40, height: 40)

              Image(systemName: feature.icon)
                .font(.system(size: 18))
                .foregroundColor(feature.color)
            }

            Text(feature.text)
              .font(.system(size: 15))
              .foregroundColor(.primary)

            Spacer()
          }
          .padding(.horizontal, 20)
          .padding(.vertical, 8)
          .background(
            RoundedRectangle(cornerRadius: 12)
              .fill(Color.primary.opacity(0.03))
          )
          .offset(x: showFeatures ? 0 : 50)
          .opacity(showFeatures ? 1 : 0)
          .animation(
            .spring(response: 0.5, dampingFraction: 0.8).delay(Double(index) * 0.1),
            value: showFeatures
          )
        }
      }
      .frame(maxWidth: 400)
      .padding(.top, 24)

      Spacer()
      Spacer()
    }
    .padding()
    .onAppear {
      DispatchQueue.main.asyncAfter(deadline: .now() + 0.2) {
        showFeatures = true
      }
    }
    .onChange(of: isAnimating) { _, newValue in
      if newValue {
        showFeatures = false
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.2) {
          showFeatures = true
        }
      }
    }
  }
}

#Preview("Welcome Walkthrough") {
  WelcomeWalkthroughView()
    .environmentObject(AppState())
    .frame(width: 600, height: 700)
}
