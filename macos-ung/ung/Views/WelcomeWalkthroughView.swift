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

  // Testimonials - real outcomes, specific help
  private let testimonials: [Testimonial] = [
    Testimonial(
      quote: "Look, I don't wanna write a testimonial. I'm too busy actually building my app now. Dig killed 6 of my ideas in one afternoon. The 7th scored 73. That's the one I'm shipping. Leave me alone.",
      name: "Marcus Chen",
      role: "Finally Building Something",
      feature: "Dig",
      color: .yellow
    ),
    Testimonial(
      quote: "Tracking showed me I was billing 4 hours but working 9. Did the math: I was earning $11/hour. Raised my rates 3x. Same clients stayed. I was just scared of nothing.",
      name: "Jake Rivera",
      role: "Designer, now $95/hr",
      feature: "Tracking",
      color: .blue
    ),
    Testimonial(
      quote: "Client ghosted for 47 days. I sent the invoice link with 'overdue' status visible. They paid in 6 hours with an apology. The red badge did more than my 5 follow-up emails.",
      name: "Emma Lindqvist",
      role: "Photographer",
      feature: "Invoicing",
      color: .teal
    ),
    Testimonial(
      quote: "Hunt matched me with a gig I never would've found - small startup, weird niche, perfect fit. $8k contract. They found ME because I had alerts on while sleeping.",
      name: "Nina Volkov",
      role: "Writer, now with retainer",
      feature: "Hunt",
      color: .orange
    ),
    Testimonial(
      quote: "I had 23 'in progress' projects. Kanban showed me 19 were actually dead. Archived them. Finished the other 4 in two weeks. The board doesn't lie.",
      name: "David Morrison",
      role: "Solo Dev, 4 shipped apps",
      feature: "Kanban",
      color: .purple
    ),
    Testimonial(
      quote: "Goal was $8k/month. The progress bar turned red at $2k by day 20. Panic mode. Sent 12 proposals that week. Hit $11k. The bar bullied me into success.",
      name: "Sarah Okonkwo",
      role: "Consultant",
      feature: "Goals",
      color: .green
    ),
    Testimonial(
      quote: "4.5 hours free time per week. Focus mode + tracking = I know exactly where every minute goes. Built my MVP in 6 weeks. Previously took 6 months and failed.",
      name: "Raj Patel",
      role: "Eng Manager, side project shipped",
      feature: "Focus",
      color: .red
    ),
    Testimonial(
      quote: "Recurring invoices. Set it once in January. It's November. I've collected $34k on autopilot. Literally forgot some clients existed. They just... pay.",
      name: "Priya K.",
      role: "Developer",
      feature: "Invoicing",
      color: .pink
    ),
  ]

  // Journey-based onboarding: Idea → Find Work → Manage → Do Work → Get Paid
  // iOS: Simplified 4-page version (mobile-focused)
  // macOS: Full 6-page version (power user features)
  #if os(iOS)
  private let pages: [WalkthroughPage] = [
    // 1. QUICK START: Track & Invoice
    WalkthroughPage(
      icon: "clock.badge.checkmark.fill",
      iconColors: [.blue, .purple],
      title: "Track. Invoice. Done.",
      subtitle: "Your time. Your money. One tap away.",
      features: [
        FeatureItem(icon: "play.circle.fill", text: "One-tap tracking - start anywhere", color: .green),
        FeatureItem(icon: "doc.text.fill", text: "Invoices auto-generated from hours", color: .teal),
        FeatureItem(icon: "bell.badge.fill", text: "Get paid alerts on your wrist", color: .orange),
      ],
      testimonial: "\"Sent invoice from my phone at 2am. Paid by breakfast.\" - Priya"
    ),
    // 2. FIND: Gigs on the go
    WalkthroughPage(
      icon: "binoculars.fill",
      iconColors: [.orange, .red],
      title: "Gigs Find You",
      subtitle: "Hunt alerts. Even when you're offline.",
      features: [
        FeatureItem(icon: "bell.badge.fill", text: "Push alerts when perfect gigs drop", color: .red),
        FeatureItem(icon: "hand.tap.fill", text: "Quick apply from notification", color: .orange),
        FeatureItem(icon: "icloud.fill", text: "Syncs with your Mac instantly", color: .blue),
      ],
      testimonial: "\"$8k contract. Alert hit while I was at the gym.\" - Nina"
    ),
    // 3. FOCUS: Stay in flow
    WalkthroughPage(
      icon: "timer",
      iconColors: [.red, .orange],
      title: "Stay Focused",
      subtitle: "Pomodoro timer. Focus mode. Deep work.",
      features: [
        FeatureItem(icon: "timer", text: "25-minute focus sessions", color: .red),
        FeatureItem(icon: "bell.slash.fill", text: "Silence notifications automatically", color: .orange),
        FeatureItem(icon: "applewatch", text: "Timer on your Apple Watch", color: .green),
      ],
      testimonial: "\"Focus mode from my watch. No excuses anymore.\" - Raj"
    ),
    // 4. FINAL: Everything synced
    WalkthroughPage(
      icon: "sparkles",
      iconColors: [.purple, .pink],
      title: "Your Pocket Office",
      subtitle: "Track. Invoice. Focus. All synced.",
      features: [
        FeatureItem(icon: "icloud.fill", text: "Changes sync to Mac instantly", color: .blue),
        FeatureItem(icon: "lock.shield.fill", text: "Your data stays on your devices", color: .green),
        FeatureItem(icon: "applewatch", text: "Works on iPhone, iPad, Watch", color: .purple),
      ],
      isLastPage: true,
      showTestimonials: true
    ),
  ]
  #else
  private let pages: [WalkthroughPage] = [
    // 1. START: Got an idea?
    WalkthroughPage(
      icon: "lightbulb.max.fill",
      iconColors: [.yellow, .orange],
      title: "Got an Idea?",
      subtitle: "Let's find out if it's worth your time.",
      features: [
        FeatureItem(icon: "brain.head.profile", text: "Dig: 5 experts tear your idea apart", color: .purple),
        FeatureItem(icon: "chart.line.uptrend.xyaxis", text: "Revenue projections before you code", color: .green),
        FeatureItem(icon: "exclamationmark.triangle.fill", text: "Find the blind spots early", color: .red),
      ],
      testimonial: "\"Killed 6 ideas in one afternoon. The 7th scored 73. Shipping that one.\" - Marcus"
    ),
    // 2. FIND: Get your next gig
    WalkthroughPage(
      icon: "binoculars.fill",
      iconColors: [.orange, .red],
      title: "Find Your Next Gig",
      subtitle: "Hunt searches while you sleep. Or procrastinate.",
      features: [
        FeatureItem(icon: "magnifyingglass", text: "Aggregates opportunities from everywhere", color: .orange),
        FeatureItem(icon: "bell.badge.fill", text: "Alerts when perfect matches drop", color: .red),
        FeatureItem(icon: "doc.richtext", text: "One-click proposals with AI assist", color: .purple),
      ],
      testimonial: "\"$8k contract. Alert hit while I slept. They found ME.\" - Nina"
    ),
    // 3. MANAGE: Organize the chaos
    WalkthroughPage(
      icon: "rectangle.3.group.fill",
      iconColors: [.pink, .purple],
      title: "Manage the Work",
      subtitle: "Clients. Contracts. Deadlines. One board.",
      features: [
        FeatureItem(icon: "arrow.left.arrow.right", text: "Kanban board - drag gigs through stages", color: .purple),
        FeatureItem(icon: "person.2.fill", text: "Clients & contracts in one place", color: .blue),
        FeatureItem(icon: "target", text: "Set income goals, watch them fill", color: .orange),
      ],
      testimonial: "\"23 projects 'in progress.' 19 were dead. Shipped 4 in 2 weeks.\" - David"
    ),
    // 4. DO: Track your time, stay focused
    WalkthroughPage(
      icon: "clock.badge.checkmark.fill",
      iconColors: [.blue, .purple],
      title: "Do the Work",
      subtitle: "Track time. Stay focused. Know where every hour goes.",
      features: [
        FeatureItem(icon: "play.circle.fill", text: "One-click tracking - start in a second", color: .green),
        FeatureItem(icon: "timer", text: "Focus mode + Pomodoro built-in", color: .red),
        FeatureItem(icon: "chart.bar.fill", text: "See where your time actually goes", color: .blue),
      ],
      testimonial: "\"Billing 4hrs, working 9. Did the math: $11/hr. Raised rates 3x.\" - Jake"
    ),
    // 5. GET PAID: Invoice and collect
    WalkthroughPage(
      icon: "dollarsign.circle.fill",
      iconColors: [.green, .teal],
      title: "Get Paid",
      subtitle: "Your time tracked. Your invoice sent. Your money collected.",
      features: [
        FeatureItem(icon: "wand.and.stars", text: "Auto-generate invoices from hours", color: .purple),
        FeatureItem(icon: "arrow.clockwise", text: "Recurring invoices on autopilot", color: .teal),
        FeatureItem(icon: "checkmark.circle.fill", text: "Track who paid, chase who didn't", color: .green),
      ],
      testimonial: "\"Client ghosted 47 days. Sent 'overdue' link. Paid in 6 hours.\" - Emma"
    ),
    // 6. FINAL: Your complete workflow
    WalkthroughPage(
      icon: "sparkles",
      iconColors: [.purple, .pink],
      title: "Your Next Gig Starts Here",
      subtitle: "Idea → Gig → Work → Money. All in ung.",
      features: [
        FeatureItem(icon: "lock.shield.fill", text: "Your data never leaves your device", color: .green),
        FeatureItem(icon: "icloud.fill", text: "Sync across Mac, iPhone, iPad", color: .blue),
        FeatureItem(icon: "chart.pie.fill", text: "Reports that actually make sense", color: .purple),
      ],
      isLastPage: true,
      showTestimonials: true
    ),
  ]
  #endif

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
