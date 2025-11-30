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

  // Testimonials - funky, feature-specific
  private let testimonials: [Testimonial] = [
    Testimonial(
      quote: "Tracked myself for a week. Turns out I spend 11 hours on 'quick emails.' I'm basically an unpaid mailman with a design degree.",
      name: "Jake Rivera",
      role: "Freelance Designer",
      feature: "Tracking",
      color: .blue
    ),
    Testimonial(
      quote: "Sent an invoice at 2am in my underwear. Got paid by breakfast. This is the future our ancestors dreamed of.",
      name: "Priya K.",
      role: "Developer",
      feature: "Invoicing",
      color: .teal
    ),
    Testimonial(
      quote: "Dig told me my 'revolutionary' idea was just Uber for laundry. Again. Fourth time. I might have a problem.",
      name: "Marcus Chen",
      role: "Serial Idea-Haver",
      feature: "Dig",
      color: .yellow
    ),
    Testimonial(
      quote: "Hunt found me 3 gigs while I was doom-scrolling. It's hunting while I'm procrastinating. Terrifying. Also: rent paid.",
      name: "Nina Volkov",
      role: "Freelance Writer",
      feature: "Hunt",
      color: .orange
    ),
    Testimonial(
      quote: "Moved a card to 'Done' and felt more dopamine than my entire childhood. Is this addiction? Don't care. Shipping.",
      name: "David Morrison",
      role: "Solo Developer",
      feature: "Kanban",
      color: .purple
    ),
    Testimonial(
      quote: "Set a $10k monthly goal as a joke. Hit it by accident because the red progress bar made me feel broke every morning.",
      name: "Sarah Okonkwo",
      role: "Consultant",
      feature: "Goals",
      color: .green
    ),
    Testimonial(
      quote: "25 minutes of Focus mode and I wrote more code than the entire previous week. My phone might be the actual enemy.",
      name: "Raj Patel",
      role: "Engineer",
      feature: "Focus",
      color: .red
    ),
    Testimonial(
      quote: "Client said 'we'll pay eventually.' I showed them my Invoices dashboard. They paid in 4 hours. Fear is a motivator.",
      name: "Emma Lindqvist",
      role: "Photographer",
      feature: "Invoicing",
      color: .pink
    ),
  ]

  // 6-page walkthrough covering all major features
  private let pages: [WalkthroughPage] = [
    // 1. Time Tracking + Goals
    WalkthroughPage(
      icon: "clock.badge.checkmark.fill",
      iconColors: [.blue, .purple],
      title: "Track Time. Get Paid.",
      subtitle: "Every minute counts. Now they're all counted.",
      features: [
        FeatureItem(icon: "play.circle.fill", text: "One-click tracking - start in a second", color: .green),
        FeatureItem(icon: "chart.bar.fill", text: "See where your time actually goes", color: .blue),
        FeatureItem(icon: "target", text: "Set income goals, watch the bar fill", color: .orange),
      ],
      testimonial: "\"Found 11 hours/week I was giving away for free.\" - Designer"
    ),
    // 2. Gigs & Kanban
    WalkthroughPage(
      icon: "rectangle.3.group.fill",
      iconColors: [.pink, .purple],
      title: "Gigs on a Board",
      subtitle: "Drag. Drop. Ship. Repeat.",
      features: [
        FeatureItem(icon: "arrow.left.arrow.right", text: "Kanban board - move gigs through stages", color: .purple),
        FeatureItem(icon: "person.2.fill", text: "Client management in one place", color: .blue),
        FeatureItem(icon: "doc.badge.clock", text: "Contracts & deadlines auto-tracked", color: .orange),
      ],
      testimonial: "\"Moved a card to Done. Felt dopamine. Shipping.\" - Solo Dev"
    ),
    // 3. Invoicing
    WalkthroughPage(
      icon: "doc.text.fill",
      iconColors: [.teal, .cyan],
      title: "Invoice in Seconds",
      subtitle: "Your time. Your money. On autopilot.",
      features: [
        FeatureItem(icon: "wand.and.stars", text: "Auto-generate from tracked hours", color: .purple),
        FeatureItem(icon: "checkmark.circle.fill", text: "Know who paid, who owes", color: .green),
        FeatureItem(icon: "arrow.clockwise", text: "Recurring invoices on autopilot", color: .teal),
      ],
      testimonial: "\"2am invoice. Paid by breakfast.\" - Developer"
    ),
    // 4. Hunt - Job Finder
    WalkthroughPage(
      icon: "binoculars.fill",
      iconColors: [.orange, .red],
      title: "Hunt: Find Work",
      subtitle: "It hunts while you sleep. Or doom-scroll.",
      features: [
        FeatureItem(icon: "magnifyingglass", text: "Aggregates gigs from everywhere", color: .orange),
        FeatureItem(icon: "bell.badge.fill", text: "Alerts when perfect matches appear", color: .red),
        FeatureItem(icon: "doc.richtext", text: "One-click proposals with AI", color: .purple),
      ],
      testimonial: "\"Found 3 gigs while I was procrastinating.\" - Writer"
    ),
    // 5. Focus Mode / Pomodoro
    WalkthroughPage(
      icon: "timer",
      iconColors: [.red, .orange],
      title: "Focus: Deep Work",
      subtitle: "25 minutes of silence. A lifetime of productivity.",
      features: [
        FeatureItem(icon: "bell.slash.fill", text: "Kill notifications, enter flow state", color: .red),
        FeatureItem(icon: "clock.badge.checkmark", text: "Pomodoro timer built-in", color: .orange),
        FeatureItem(icon: "chart.xyaxis.line", text: "Track focus streaks over time", color: .green),
      ],
      testimonial: "\"25 mins Focus > my entire previous week.\" - Engineer"
    ),
    // 6. Dig - Idea Validation
    WalkthroughPage(
      icon: "lightbulb.max.fill",
      iconColors: [.yellow, .orange],
      title: "Dig: Validate Ideas",
      subtitle: "10 minutes of truth beats 10 months of hope.",
      features: [
        FeatureItem(icon: "brain.head.profile", text: "5 experts tear your idea apart", color: .purple),
        FeatureItem(icon: "chart.line.uptrend.xyaxis", text: "Revenue projections before code", color: .green),
        FeatureItem(icon: "exclamationmark.triangle.fill", text: "Devil's Advocate finds blind spots", color: .red),
      ],
      testimonial: "\"Told me it was Uber for laundry. Again.\" - Idea-Haver"
    ),
    // 7. Final page - Everything together
    WalkthroughPage(
      icon: "sparkles",
      iconColors: [.green, .teal],
      title: "Your Freelance OS",
      subtitle: "Track. Invoice. Hunt. Focus. Build.",
      features: [
        FeatureItem(icon: "lock.shield.fill", text: "Your data never leaves your device", color: .green),
        FeatureItem(icon: "icloud.fill", text: "Sync across Mac, iPhone, iPad", color: .blue),
        FeatureItem(icon: "chart.pie.fill", text: "Reports that actually make sense", color: .purple),
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
