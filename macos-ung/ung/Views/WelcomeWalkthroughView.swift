//
//  WelcomeWalkthroughView.swift
//  ung
//
//  Classic multi-page walkthrough for first-time users
//

import SwiftUI

// MARK: - Testimonial Model
struct Testimonial {
    let quote: String
    let name: String
    let role: String
    let feature: String
    let color: Color
}

// MARK: - Welcome Walkthrough View
struct WelcomeWalkthroughView: View {
  @EnvironmentObject var appState: AppState
  @State private var currentPage = 0
  @State private var isAnimating = false
  @Namespace private var animation

  // Testimonials from real user personas
  private let testimonials: [Testimonial] = [
    Testimonial(
      quote: "Dig told me my habit tracker idea scored 34. It hurt. But it would have hurt more to waste another four months.",
      name: "Marcus Chen",
      role: "Senior Engineer, SF",
      feature: "Dig",
      color: .yellow
    ),
    Testimonial(
      quote: "I ran my idea through the Scalability Stress Test - it said my architecture would break at 1,000 users. Found out in minute 7, not month 3.",
      name: "Raj Patel",
      role: "Engineering Manager",
      feature: "Dig",
      color: .orange
    ),
    Testimonial(
      quote: "Finally, one app that tracks my time AND generates invoices from it. No more spreadsheet juggling.",
      name: "Emma L.",
      role: "Freelance Designer",
      feature: "Tracking",
      color: .blue
    ),
    Testimonial(
      quote: "The Copycat Analysis showed me 85% of billion-dollar companies are copycats. It gave me permission to stop seeking novelty and start seeking execution.",
      name: "Sarah Okonkwo",
      role: "Ex-McKinsey Founder",
      feature: "Dig",
      color: .purple
    ),
  ]

  // Balanced 4-page walkthrough - core features first, Dig as powerful bonus
  private let pages: [WalkthroughPage] = [
    // Core workflow first - Time Tracking
    WalkthroughPage(
      icon: "clock.badge.checkmark.fill",
      iconColors: [.blue, .purple],
      title: "Track Time. Get Paid.",
      subtitle: "Every minute counts. Now they're all counted.",
      features: [
        FeatureItem(icon: "play.circle.fill", text: "One-click tracking - start in a second", color: .green),
        FeatureItem(icon: "chart.bar.fill", text: "See where your time actually goes", color: .blue),
        FeatureItem(icon: "target", text: "Set income goals, watch them happen", color: .orange),
      ],
      testimonial: "\"I found 12 hours/week I was giving away for free.\" - Designer"
    ),
    // Invoicing
    WalkthroughPage(
      icon: "doc.text.fill",
      iconColors: [.teal, .cyan],
      title: "Invoice in Seconds",
      subtitle: "Your time. Your money. On autopilot.",
      features: [
        FeatureItem(icon: "wand.and.stars", text: "Auto-generate from tracked hours", color: .purple),
        FeatureItem(icon: "checkmark.circle.fill", text: "Know who paid, who owes", color: .green),
        FeatureItem(icon: "rectangle.3.group.fill", text: "Kanban board for all your gigs", color: .pink),
      ],
      testimonial: "\"No more spreadsheet chaos. Just send and get paid.\" - Freelancer"
    ),
    // Dig as the powerful "bonus" feature
    WalkthroughPage(
      icon: "lightbulb.max.fill",
      iconColors: [.yellow, .orange],
      title: "Dig: Validate First",
      subtitle: "10 minutes of truth beats 10 months of hope.",
      features: [
        FeatureItem(icon: "brain.head.profile", text: "5 expert perspectives tear your idea apart", color: .purple),
        FeatureItem(icon: "chart.line.uptrend.xyaxis", text: "Revenue projections before line one of code", color: .green),
        FeatureItem(icon: "exclamationmark.triangle.fill", text: "Devil's Advocate finds your blind spots", color: .red),
      ],
      testimonial: "\"Saved me from building 3 things nobody wanted.\" - Indie Dev"
    ),
    // Final page with testimonials
    WalkthroughPage(
      icon: "sparkles",
      iconColors: [.green, .teal],
      title: "Your Freelance OS",
      subtitle: "Track. Invoice. Build. All in one place.",
      features: [
        FeatureItem(icon: "lock.shield.fill", text: "Your data never leaves your device", color: .green),
        FeatureItem(icon: "icloud.fill", text: "Sync across Mac, iPhone, iPad", color: .blue),
        FeatureItem(icon: "timer", text: "Focus mode when you need flow", color: .orange),
      ],
      isLastPage: true,
      showTestimonials: true
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
              WalkthroughPageView(
                page: pages[index],
                isAnimating: currentPage == index,
                testimonials: testimonials
              )
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
  var testimonial: String? = nil  // Mini testimonial for feature pages
  var showTestimonials: Bool = false  // Show full testimonials carousel
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
  var testimonials: [Testimonial] = []
  @State private var showFeatures = false
  @State private var currentTestimonialIndex = 0

  var body: some View {
    VStack(spacing: 20) {
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
          .frame(width: 140, height: 140)
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
          .frame(width: 90, height: 90)
          .shadow(color: page.iconColors[0].opacity(0.5), radius: 20, y: 10)

        // Icon
        Image(systemName: page.icon)
          .font(.system(size: 40, weight: .medium))
          .foregroundColor(.white)
          .symbolEffect(.pulse, options: .repeating, value: isAnimating)
      }
      .padding(.bottom, 12)

      // Title
      Text(page.title)
        .font(.system(size: 28, weight: .bold, design: .rounded))
        .foregroundColor(.primary)
        .multilineTextAlignment(.center)

      // Subtitle
      Text(page.subtitle)
        .font(.system(size: 16))
        .foregroundColor(.secondary)
        .multilineTextAlignment(.center)
        .padding(.horizontal, 40)

      // Feature list
      VStack(spacing: 12) {
        ForEach(Array(page.features.enumerated()), id: \.offset) { index, feature in
          HStack(spacing: 14) {
            ZStack {
              Circle()
                .fill(feature.color.opacity(0.15))
                .frame(width: 36, height: 36)

              Image(systemName: feature.icon)
                .font(.system(size: 16))
                .foregroundColor(feature.color)
            }

            Text(feature.text)
              .font(.system(size: 14))
              .foregroundColor(.primary)

            Spacer()
          }
          .padding(.horizontal, 16)
          .padding(.vertical, 6)
          .background(
            RoundedRectangle(cornerRadius: 10)
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
      .frame(maxWidth: 380)
      .padding(.top, 16)

      // Mini testimonial (for feature pages like Dig)
      if let testimonial = page.testimonial {
        Text(testimonial)
          .font(.system(size: 13, weight: .medium, design: .serif))
          .italic()
          .foregroundColor(.secondary)
          .multilineTextAlignment(.center)
          .padding(.horizontal, 40)
          .padding(.top, 8)
          .offset(y: showFeatures ? 0 : 20)
          .opacity(showFeatures ? 1 : 0)
          .animation(.spring(response: 0.5).delay(0.4), value: showFeatures)
      }

      // Full testimonials carousel (for final page)
      if page.showTestimonials && !testimonials.isEmpty {
        TestimonialCard(testimonial: testimonials[currentTestimonialIndex])
          .frame(maxWidth: 400)
          .padding(.top, 8)
          .offset(y: showFeatures ? 0 : 20)
          .opacity(showFeatures ? 1 : 0)
          .animation(.spring(response: 0.5).delay(0.3), value: showFeatures)
          .onAppear {
            // Auto-rotate testimonials
            Timer.scheduledTimer(withTimeInterval: 4.0, repeats: true) { _ in
              withAnimation(.easeInOut(duration: 0.5)) {
                currentTestimonialIndex = (currentTestimonialIndex + 1) % testimonials.count
              }
            }
          }

        // Testimonial indicators
        HStack(spacing: 6) {
          ForEach(0..<testimonials.count, id: \.self) { index in
            Circle()
              .fill(currentTestimonialIndex == index ? testimonials[index].color : Color.secondary.opacity(0.3))
              .frame(width: 6, height: 6)
              .onTapGesture {
                withAnimation { currentTestimonialIndex = index }
              }
          }
        }
        .padding(.top, 4)
      }

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

// MARK: - Testimonial Card
struct TestimonialCard: View {
  let testimonial: Testimonial

  var body: some View {
    VStack(alignment: .leading, spacing: 10) {
      // Quote
      HStack(alignment: .top, spacing: 8) {
        Image(systemName: "quote.opening")
          .font(.system(size: 16))
          .foregroundColor(testimonial.color.opacity(0.6))

        Text(testimonial.quote)
          .font(.system(size: 13, design: .serif))
          .italic()
          .foregroundColor(.primary.opacity(0.9))
          .lineLimit(3)
      }

      // Attribution
      HStack {
        Spacer()
        VStack(alignment: .trailing, spacing: 2) {
          Text(testimonial.name)
            .font(.system(size: 12, weight: .semibold))
            .foregroundColor(.primary)
          Text(testimonial.role)
            .font(.system(size: 11))
            .foregroundColor(.secondary)
        }
      }
    }
    .padding(16)
    .background(
      RoundedRectangle(cornerRadius: 12)
        .fill(testimonial.color.opacity(0.08))
        .overlay(
          RoundedRectangle(cornerRadius: 12)
            .stroke(testimonial.color.opacity(0.2), lineWidth: 1)
        )
    )
  }
}

#Preview("Welcome Walkthrough") {
  WelcomeWalkthroughView()
    .environmentObject(AppState())
    .frame(width: 600, height: 700)
}
