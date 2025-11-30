//
//  WelcomeWalkthroughView.swift
//  ung
//
//  Premium animated multi-page walkthrough for first-time users
//  Inspired by Uber, Headspace, and top-tier app onboarding experiences
//

import SwiftUI

// MARK: - Haptic Feedback Manager
#if os(iOS)
import UIKit

class HapticManager {
    static let shared = HapticManager()

    private let lightImpact = UIImpactFeedbackGenerator(style: .light)
    private let mediumImpact = UIImpactFeedbackGenerator(style: .medium)
    private let heavyImpact = UIImpactFeedbackGenerator(style: .heavy)
    private let softImpact = UIImpactFeedbackGenerator(style: .soft)
    private let rigidImpact = UIImpactFeedbackGenerator(style: .rigid)
    private let selectionFeedback = UISelectionFeedbackGenerator()
    private let notificationFeedback = UINotificationFeedbackGenerator()

    private init() {
        // Prepare all generators for immediate response
        prepareAll()
    }

    func prepareAll() {
        lightImpact.prepare()
        mediumImpact.prepare()
        heavyImpact.prepare()
        softImpact.prepare()
        rigidImpact.prepare()
        selectionFeedback.prepare()
        notificationFeedback.prepare()
    }

    func impact(_ style: UIImpactFeedbackGenerator.FeedbackStyle) {
        switch style {
        case .light: lightImpact.impactOccurred()
        case .medium: mediumImpact.impactOccurred()
        case .heavy: heavyImpact.impactOccurred()
        case .soft: softImpact.impactOccurred()
        case .rigid: rigidImpact.impactOccurred()
        @unknown default: lightImpact.impactOccurred()
        }
    }

    func impactWithIntensity(_ intensity: CGFloat) {
        mediumImpact.impactOccurred(intensity: intensity)
    }

    func selection() {
        selectionFeedback.selectionChanged()
    }

    func notification(_ type: UINotificationFeedbackGenerator.FeedbackType) {
        notificationFeedback.notificationOccurred(type)
    }
}
#endif

// MARK: - Reduce Motion Check
struct MotionManager {
    static var reduceMotionEnabled: Bool {
        #if os(iOS)
        return UIAccessibility.isReduceMotionEnabled
        #else
        return NSWorkspace.shared.accessibilityDisplayShouldReduceMotion
        #endif
    }
}

// MARK: - Testimonial Model
struct Testimonial {
    let quote: String
    let name: String
    let role: String
    let feature: String
    let color: Color
}

// MARK: - Floating Orb Model
struct FloatingOrb: Identifiable {
    let id = UUID()
    var x: CGFloat
    var y: CGFloat
    var z: CGFloat // Depth for parallax
    var size: CGFloat
    var color: Color
    var opacity: Double
    var speed: Double
    var phase: Double
}

// MARK: - Sparkle Particle (for icon effects)
struct SparkleParticle: Identifiable {
    let id = UUID()
    var x: CGFloat
    var y: CGFloat
    var scale: CGFloat
    var rotation: Double
    var opacity: Double
    var delay: Double
}

// MARK: - Confetti Particle
struct ConfettiParticle: Identifiable {
    let id = UUID()
    var x: CGFloat
    var y: CGFloat
    var rotation: Double
    var rotationAxis: (x: CGFloat, y: CGFloat, z: CGFloat)
    var scale: CGFloat
    var color: Color
    var speed: Double
    var shape: ConfettiShape
    var wobble: Double
}

enum ConfettiShape: CaseIterable {
    case rectangle, circle, star, triangle
}

// MARK: - Welcome Walkthrough View
struct WelcomeWalkthroughView: View {
  @EnvironmentObject var appState: AppState
  @State private var currentPage = 0
  @State private var isAnimating = false
  @State private var floatingOrbs: [FloatingOrb] = []
  @State private var confettiParticles: [ConfettiParticle] = []
  @State private var sparkleParticles: [SparkleParticle] = []
  @State private var showConfetti = false
  @State private var animationTimer: Timer?
  @State private var dragOffset: CGFloat = 0
  @State private var dragOffset3D: CGSize = .zero  // For 3D tilt effect
  @State private var progressAnimation: CGFloat = 0
  @State private var buttonGlowPhase: Double = 0
  @State private var isButtonPressed = false
  @Namespace private var animation

  // Reduce motion preference
  private var reduceMotion: Bool { MotionManager.reduceMotionEnabled }

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
  // iOS: Optimized 6-page version (mobile-first design, different visuals)
  // macOS: Full 6-page version (power user features)
  #if os(iOS)
  private let pages: [WalkthroughPage] = [
    // 1. START: Your freelance companion
    WalkthroughPage(
      icon: "sparkles",
      iconColors: [.purple, .pink],
      title: "Your Gig, Your Way",
      subtitle: "From spark to paycheck. All in your pocket.",
      features: [
        FeatureItem(icon: "lightbulb.max.fill", text: "Validate ideas before you build", color: .yellow),
        FeatureItem(icon: "clock.fill", text: "Track every hour, effortlessly", color: .blue),
        FeatureItem(icon: "dollarsign.circle.fill", text: "Get paid, on time, every time", color: .green),
      ],
      testimonial: "\"This app is my entire business.\" - Priya"
    ),
    // 2. FIND: Gigs find you
    WalkthroughPage(
      icon: "binoculars.fill",
      iconColors: [.orange, .red],
      title: "Gigs Find You",
      subtitle: "Push alerts. Perfect matches. Zero effort.",
      features: [
        FeatureItem(icon: "bell.badge.fill", text: "Alerts when your gig drops", color: .red),
        FeatureItem(icon: "hand.tap.fill", text: "Quick apply from notification", color: .orange),
        FeatureItem(icon: "sparkle.magnifyingglass", text: "AI matches you to opportunities", color: .purple),
      ],
      testimonial: "\"$8k contract. Alert hit at the gym.\" - Nina"
    ),
    // 3. TRACK: Every minute counts
    WalkthroughPage(
      icon: "clock.badge.checkmark.fill",
      iconColors: [.blue, .cyan],
      title: "Track Everything",
      subtitle: "One tap. Zero friction. Pure accountability.",
      features: [
        FeatureItem(icon: "play.circle.fill", text: "Start tracking in one tap", color: .green),
        FeatureItem(icon: "chart.bar.fill", text: "See where your time actually goes", color: .blue),
        FeatureItem(icon: "applewatch", text: "Track from your Apple Watch", color: .purple),
      ],
      testimonial: "\"Found out I worked 9hrs, billed 4.\" - Jake"
    ),
    // 4. FOCUS: Deep work mode
    WalkthroughPage(
      icon: "timer",
      iconColors: [.red, .orange],
      title: "Stay Focused",
      subtitle: "Pomodoro. Focus mode. No distractions.",
      features: [
        FeatureItem(icon: "moon.fill", text: "Focus mode silences everything", color: .indigo),
        FeatureItem(icon: "flame.fill", text: "Build focus streaks", color: .orange),
        FeatureItem(icon: "bell.slash.fill", text: "DND syncs automatically", color: .red),
      ],
      testimonial: "\"Built my MVP in 6 weeks. Focus mode.\" - Raj"
    ),
    // 5. GET PAID: Invoice instantly
    WalkthroughPage(
      icon: "dollarsign.circle.fill",
      iconColors: [.green, .teal],
      title: "Get Paid Fast",
      subtitle: "Track hours → Generate invoice → Collect money.",
      features: [
        FeatureItem(icon: "wand.and.stars", text: "Auto-invoices from tracked time", color: .purple),
        FeatureItem(icon: "arrow.clockwise", text: "Recurring invoices on autopilot", color: .teal),
        FeatureItem(icon: "bell.badge.fill", text: "Payment alerts on your phone", color: .orange),
      ],
      testimonial: "\"Sent invoice at 2am. Paid by breakfast.\" - Priya"
    ),
    // 6. FINAL: Everything synced
    WalkthroughPage(
      icon: "icloud.fill",
      iconColors: [.purple, .blue],
      title: "Always In Sync",
      subtitle: "iPhone. iPad. Mac. Watch. All connected.",
      features: [
        FeatureItem(icon: "lock.shield.fill", text: "Your data stays on your devices", color: .green),
        FeatureItem(icon: "arrow.triangle.2.circlepath", text: "Real-time sync across everything", color: .blue),
        FeatureItem(icon: "hand.raised.fill", text: "No account required", color: .purple),
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
        // MARK: - Animated Background
        AnimatedGradientBackground(
          colors: pages[currentPage].iconColors,
          geometry: geometry,
          reduceMotion: reduceMotion
        )
        .ignoresSafeArea()

        // MARK: - Floating Orbs (Parallax Background with Depth)
        ForEach(floatingOrbs) { orb in
          Circle()
            .fill(
              RadialGradient(
                colors: [orb.color.opacity(orb.opacity), orb.color.opacity(0)],
                center: .center,
                startRadius: 0,
                endRadius: orb.size / 2
              )
            )
            .frame(width: orb.size, height: orb.size)
            .position(x: orb.x, y: orb.y)
            .blur(radius: orb.size * 0.1)
            // 3D Parallax based on depth (z) - farther objects move less
            .offset(
              x: reduceMotion ? 0 : dragOffset3D.width * orb.z * 0.02,
              y: reduceMotion ? 0 : dragOffset3D.height * orb.z * 0.02
            )
            .scaleEffect(1.0 + (orb.z * 0.001))
        }

        // MARK: - Sparkle Effects
        if !reduceMotion {
          ForEach(sparkleParticles) { sparkle in
            SparkleView(particle: sparkle, colors: pages[currentPage].iconColors)
          }
        }

        // MARK: - Confetti (Final Page)
        if showConfetti {
          ForEach(confettiParticles) { particle in
            ConfettiView(particle: particle, reduceMotion: reduceMotion)
          }
        }

        VStack(spacing: 0) {
          // MARK: - Progress Bar
          ProgressBarView(
            currentPage: currentPage,
            totalPages: pages.count,
            colors: pages[currentPage].iconColors,
            progress: progressAnimation,
            reduceMotion: reduceMotion
          )
          .padding(.top, 20)
          .padding(.horizontal, 40)

          // MARK: - Page Content with 3D Perspective
          TabView(selection: $currentPage) {
            ForEach(0..<pages.count, id: \.self) { index in
              WalkthroughPageView(
                page: pages[index],
                isAnimating: currentPage == index,
                testimonials: testimonials,
                geometry: geometry,
                pageIndex: index,
                reduceMotion: reduceMotion
              )
              .tag(index)
            }
          }
          #if os(macOS)
          .tabViewStyle(.automatic)
          #else
          .tabViewStyle(.page(indexDisplayMode: .never))
          #endif
          // 3D tilt effect on drag (subtle)
          .rotation3DEffect(
            .degrees(reduceMotion ? 0 : Double(dragOffset3D.width) * 0.01),
            axis: (x: 0, y: 1, z: 0),
            perspective: 0.5
          )
          .rotation3DEffect(
            .degrees(reduceMotion ? 0 : Double(-dragOffset3D.height) * 0.008),
            axis: (x: 1, y: 0, z: 0),
            perspective: 0.5
          )
          .onChange(of: currentPage) { oldValue, newValue in
            // Rich haptic feedback on iOS
            #if os(iOS)
            HapticManager.shared.impact(.light)
            // Additional subtle haptic as page settles
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.15) {
              HapticManager.shared.impact(.soft)
            }
            #endif

            // Animate progress bar
            withAnimation(.spring(response: 0.6, dampingFraction: 0.8)) {
              progressAnimation = CGFloat(newValue) / CGFloat(pages.count - 1)
            }

            // Regenerate sparkles for new page
            if !reduceMotion {
              setupSparkles(geometry: geometry)
            }

            // Show confetti on last page
            if pages[newValue].isLastPage && !showConfetti {
              triggerConfetti(geometry: geometry)
            }
          }

          // MARK: - Bottom Navigation
          BottomNavigationView(
            currentPage: $currentPage,
            pages: pages,
            onComplete: {
              #if os(iOS)
              HapticManager.shared.notification(.success)
              #endif
              appState.completeWelcomeWalkthrough()
            },
            reduceMotion: reduceMotion
          )
          .padding(.bottom, 30)
        }
      }
      // Track drag for 3D effect
      .gesture(
        DragGesture(minimumDistance: 0)
          .onChanged { value in
            if !reduceMotion {
              withAnimation(.interactiveSpring(response: 0.15, dampingFraction: 0.8)) {
                dragOffset3D = value.translation
              }
            }
          }
          .onEnded { _ in
            withAnimation(.spring(response: 0.5, dampingFraction: 0.7)) {
              dragOffset3D = .zero
            }
          }
      )
    }
    .frame(minWidth: 500, minHeight: 600)
    .onAppear {
      setupFloatingOrbs()
      if !reduceMotion {
        startAnimations()
      }
      #if os(iOS)
      HapticManager.shared.prepareAll()
      #endif
    }
    .onDisappear {
      animationTimer?.invalidate()
    }
  }

  // MARK: - Setup Floating Orbs
  private func setupFloatingOrbs() {
    floatingOrbs = (0..<8).map { i in
      FloatingOrb(
        x: CGFloat.random(in: 50...450),
        y: CGFloat.random(in: 50...550),
        z: CGFloat.random(in: 0.5...2.0), // Depth for parallax
        size: CGFloat.random(in: 60...200),
        color: [Color.purple, .blue, .pink, .orange, .teal, .yellow].randomElement()!,
        opacity: Double.random(in: 0.1...0.25),
        speed: Double.random(in: 0.3...0.8),
        phase: Double.random(in: 0...2 * .pi)
      )
    }
  }

  // MARK: - Setup Sparkles
  private func setupSparkles(geometry: GeometryProxy) {
    sparkleParticles = (0..<12).map { i in
      SparkleParticle(
        x: CGFloat.random(in: geometry.size.width * 0.2...geometry.size.width * 0.8),
        y: CGFloat.random(in: geometry.size.height * 0.1...geometry.size.height * 0.5),
        scale: CGFloat.random(in: 0.3...0.8),
        rotation: Double.random(in: 0...360),
        opacity: Double.random(in: 0.3...0.7),
        delay: Double(i) * 0.1
      )
    }
  }

  // MARK: - Start Animations
  private func startAnimations() {
    var time: Double = 0
    animationTimer = Timer.scheduledTimer(withTimeInterval: 1/30, repeats: true) { _ in
      time += 0.02
      withAnimation(.linear(duration: 0.03)) {
        for i in floatingOrbs.indices {
          let orb = floatingOrbs[i]
          floatingOrbs[i].x = orb.x + CGFloat(sin(time * orb.speed + orb.phase)) * 0.5
          floatingOrbs[i].y = orb.y + CGFloat(cos(time * orb.speed * 0.7 + orb.phase)) * 0.3
        }
      }
    }
  }

  // MARK: - Trigger Confetti
  private func triggerConfetti(geometry: GeometryProxy) {
    showConfetti = true

    // Create varied confetti with different shapes and physics
    confettiParticles = (0..<60).map { i in
      ConfettiParticle(
        x: geometry.size.width * CGFloat.random(in: 0.2...0.8),
        y: CGFloat.random(in: -50...(-20)),
        rotation: Double.random(in: 0...360),
        rotationAxis: (
          x: CGFloat.random(in: 0...1),
          y: CGFloat.random(in: 0...1),
          z: CGFloat.random(in: 0...1)
        ),
        scale: CGFloat.random(in: 0.5...1.2),
        color: [.purple, .pink, .orange, .yellow, .green, .blue, .red, .cyan, .mint].randomElement()!,
        speed: Double.random(in: 2...5),
        shape: ConfettiShape.allCases.randomElement()!,
        wobble: Double.random(in: 0...2 * .pi)
      )
    }

    // Animate confetti falling with wobble
    if reduceMotion {
      // Simple fade for reduce motion
      withAnimation(.easeOut(duration: 2)) {
        for i in confettiParticles.indices {
          confettiParticles[i].y = geometry.size.height + 50
        }
      }
    } else {
      withAnimation(.easeOut(duration: 4)) {
        for i in confettiParticles.indices {
          confettiParticles[i].y = geometry.size.height + 50
          confettiParticles[i].x += CGFloat.random(in: -250...250)
          confettiParticles[i].rotation += Double.random(in: 720...1440)
          confettiParticles[i].wobble += Double.random(in: 4...8) * .pi
        }
      }
    }

    // Rich haptic feedback cascade
    #if os(iOS)
    HapticManager.shared.notification(.success)
    // Cascade haptics for celebration feel
    for i in 1...3 {
      DispatchQueue.main.asyncAfter(deadline: .now() + Double(i) * 0.15) {
        HapticManager.shared.impact(.light)
      }
    }
    #endif
  }
}

// MARK: - Sparkle View
struct SparkleView: View {
  let particle: SparkleParticle
  let colors: [Color]
  @State private var isAnimating = false

  var body: some View {
    Image(systemName: "sparkle")
      .font(.system(size: 12 * particle.scale))
      .foregroundStyle(
        LinearGradient(colors: colors, startPoint: .topLeading, endPoint: .bottomTrailing)
      )
      .opacity(isAnimating ? particle.opacity : 0)
      .scaleEffect(isAnimating ? 1.0 : 0.3)
      .rotationEffect(.degrees(isAnimating ? particle.rotation + 45 : particle.rotation))
      .position(x: particle.x, y: particle.y)
      .onAppear {
        withAnimation(
          .easeInOut(duration: 1.5)
          .repeatForever(autoreverses: true)
          .delay(particle.delay)
        ) {
          isAnimating = true
        }
      }
  }
}

// MARK: - Animated Gradient Background
struct AnimatedGradientBackground: View {
  let colors: [Color]
  let geometry: GeometryProxy
  var reduceMotion: Bool = false
  @State private var animateGradient = false
  @State private var meshPhase: CGFloat = 0

  var body: some View {
    ZStack {
      // Base gradient
      LinearGradient(
        colors: [
          colors[0].opacity(0.15),
          colors[1].opacity(0.08),
          Color.clear
        ],
        startPoint: reduceMotion ? .topLeading : (animateGradient ? .topLeading : .topTrailing),
        endPoint: reduceMotion ? .bottomTrailing : (animateGradient ? .bottomTrailing : .bottomLeading)
      )

      // Animated mesh-like overlay
      EllipticalGradient(
        colors: [colors[0].opacity(0.12), Color.clear],
        center: reduceMotion ? .topLeading : (animateGradient ? .topLeading : .bottomTrailing),
        startRadiusFraction: 0,
        endRadiusFraction: 0.8
      )

      // Second animated overlay
      EllipticalGradient(
        colors: [colors[1].opacity(0.1), Color.clear],
        center: reduceMotion ? .bottomTrailing : (animateGradient ? .bottomTrailing : .topLeading),
        startRadiusFraction: 0,
        endRadiusFraction: 0.6
      )

      // Third layer for more depth (premium feel)
      if !reduceMotion {
        EllipticalGradient(
          colors: [colors[0].opacity(0.05), Color.clear],
          center: .center,
          startRadiusFraction: 0,
          endRadiusFraction: 0.5 + meshPhase * 0.2
        )
        .blendMode(.plusLighter)
      }
    }
    .animation(reduceMotion ? nil : .easeInOut(duration: 4).repeatForever(autoreverses: true), value: animateGradient)
    .onAppear {
      if !reduceMotion {
        animateGradient = true
        withAnimation(.easeInOut(duration: 6).repeatForever(autoreverses: true)) {
          meshPhase = 1
        }
      }
    }
    .animation(.easeInOut(duration: 0.8), value: colors)
  }
}

// MARK: - Progress Bar View
struct ProgressBarView: View {
  let currentPage: Int
  let totalPages: Int
  let colors: [Color]
  let progress: CGFloat
  var reduceMotion: Bool = false
  @State private var glowPhase: CGFloat = 0

  var body: some View {
    VStack(spacing: 8) {
      // Step indicators with glow effect
      HStack(spacing: 4) {
        ForEach(0..<totalPages, id: \.self) { index in
          Capsule()
            .fill(index <= currentPage ?
              LinearGradient(colors: colors, startPoint: .leading, endPoint: .trailing) :
              LinearGradient(colors: [Color.secondary.opacity(0.2)], startPoint: .leading, endPoint: .trailing)
            )
            .frame(height: 4)
            .shadow(
              color: index <= currentPage ? colors[0].opacity(reduceMotion ? 0.3 : 0.3 + glowPhase * 0.2) : .clear,
              radius: reduceMotion ? 4 : 4 + glowPhase * 2
            )
            .animation(.spring(response: 0.4, dampingFraction: 0.7), value: currentPage)
        }
      }

      // Page label
      HStack {
        Text("Step \(currentPage + 1) of \(totalPages)")
          .font(.caption)
          .fontWeight(.medium)
          .foregroundColor(.secondary)

        Spacer()

        // Percentage complete with animated appearance
        Text("\(Int(progress * 100))% complete")
          .font(.caption)
          .fontWeight(.semibold)
          .foregroundColor(colors[0])
          .contentTransition(.numericText())
      }
    }
    .onAppear {
      if !reduceMotion {
        withAnimation(.easeInOut(duration: 1.5).repeatForever(autoreverses: true)) {
          glowPhase = 1
        }
      }
    }
  }
}

// MARK: - Confetti View
struct ConfettiView: View {
  let particle: ConfettiParticle
  var reduceMotion: Bool = false
  @State private var appear = false

  var body: some View {
    confettiShape
      .fill(particle.color)
      .frame(width: shapeSize.width * particle.scale, height: shapeSize.height * particle.scale)
      .rotation3DEffect(
        .degrees(reduceMotion ? 0 : particle.rotation),
        axis: (x: particle.rotationAxis.x, y: particle.rotationAxis.y, z: particle.rotationAxis.z)
      )
      .offset(x: reduceMotion ? 0 : sin(particle.wobble) * 20)
      .position(x: particle.x, y: particle.y)
      .opacity(appear ? 0.9 : 0)
      .shadow(color: particle.color.opacity(0.5), radius: 2)
      .onAppear {
        withAnimation(.easeOut(duration: 0.3)) {
          appear = true
        }
      }
  }

  private var shapeSize: CGSize {
    switch particle.shape {
    case .rectangle: return CGSize(width: 8, height: 14)
    case .circle: return CGSize(width: 10, height: 10)
    case .star: return CGSize(width: 12, height: 12)
    case .triangle: return CGSize(width: 10, height: 12)
    }
  }

  @ViewBuilder
  private var confettiShape: some Shape {
    switch particle.shape {
    case .rectangle:
      RoundedRectangle(cornerRadius: 2)
    case .circle:
      Circle()
    case .star:
      StarShape()
    case .triangle:
      Triangle()
    }
  }
}

// MARK: - Star Shape
struct StarShape: Shape {
  func path(in rect: CGRect) -> Path {
    let center = CGPoint(x: rect.midX, y: rect.midY)
    let outerRadius = min(rect.width, rect.height) / 2
    let innerRadius = outerRadius * 0.4
    var path = Path()

    for i in 0..<5 {
      let outerAngle = (CGFloat(i) * 72 - 90) * .pi / 180
      let innerAngle = (CGFloat(i) * 72 + 36 - 90) * .pi / 180

      let outerPoint = CGPoint(
        x: center.x + outerRadius * cos(outerAngle),
        y: center.y + outerRadius * sin(outerAngle)
      )
      let innerPoint = CGPoint(
        x: center.x + innerRadius * cos(innerAngle),
        y: center.y + innerRadius * sin(innerAngle)
      )

      if i == 0 {
        path.move(to: outerPoint)
      } else {
        path.addLine(to: outerPoint)
      }
      path.addLine(to: innerPoint)
    }
    path.closeSubpath()
    return path
  }
}

// MARK: - Triangle Shape
struct Triangle: Shape {
  func path(in rect: CGRect) -> Path {
    var path = Path()
    path.move(to: CGPoint(x: rect.midX, y: rect.minY))
    path.addLine(to: CGPoint(x: rect.maxX, y: rect.maxY))
    path.addLine(to: CGPoint(x: rect.minX, y: rect.maxY))
    path.closeSubpath()
    return path
  }
}

// MARK: - Bottom Navigation View
struct BottomNavigationView: View {
  @Binding var currentPage: Int
  let pages: [WalkthroughPage]
  let onComplete: () -> Void
  var reduceMotion: Bool = false
  @State private var glowPhase: CGFloat = 0
  @State private var isButtonPressed = false

  var body: some View {
    VStack(spacing: 16) {
      // Main action button
      HStack(spacing: 20) {
        // Back button
        if currentPage > 0 {
          Button(action: {
            #if os(iOS)
            HapticManager.shared.impact(.light)
            #endif
            withAnimation(.spring(response: 0.5, dampingFraction: 0.8)) {
              currentPage -= 1
            }
          }) {
            HStack(spacing: 6) {
              Image(systemName: "chevron.left")
                .font(.system(size: 14, weight: .semibold))
              Text("Back")
                .font(.system(size: 15, weight: .medium))
            }
            .foregroundColor(.secondary)
            .frame(width: 90, height: 44)
            .background(
              RoundedRectangle(cornerRadius: 22)
                .fill(.ultraThinMaterial)  // Glass morphism
            )
            .overlay(
              RoundedRectangle(cornerRadius: 22)
                .stroke(Color.primary.opacity(0.1), lineWidth: 1)
            )
          }
          .buttonStyle(GlassButtonStyle())
          .transition(.opacity.combined(with: .move(edge: .leading)))
        }

        Spacer()

        // Next / Get Started button with glow
        Button(action: {
          #if os(iOS)
          HapticManager.shared.impact(.medium)
          #endif
          if pages[currentPage].isLastPage {
            onComplete()
          } else {
            withAnimation(.spring(response: 0.5, dampingFraction: 0.8)) {
              currentPage += 1
            }
          }
        }) {
          HStack(spacing: 8) {
            Text(pages[currentPage].isLastPage ? "Get Started" : "Continue")
              .font(.system(size: 16, weight: .semibold))

            Image(systemName: pages[currentPage].isLastPage ? "arrow.right" : "chevron.right")
              .font(.system(size: 14, weight: .semibold))
              .symbolEffect(.bounce, value: pages[currentPage].isLastPage)
          }
          .foregroundColor(.white)
          .frame(width: pages[currentPage].isLastPage ? 160 : 140, height: 50)
          .background(
            ZStack {
              // Gradient background
              LinearGradient(
                colors: pages[currentPage].iconColors,
                startPoint: .leading,
                endPoint: .trailing
              )

              // Animated shine effect
              if !reduceMotion {
                LinearGradient(
                  colors: [.clear, .white.opacity(0.3), .clear],
                  startPoint: .leading,
                  endPoint: .trailing
                )
                .offset(x: -200 + glowPhase * 400)
                .mask(Capsule())
              }
            }
          )
          .clipShape(Capsule())
          // Animated glow
          .shadow(
            color: pages[currentPage].iconColors[0].opacity(reduceMotion ? 0.4 : 0.4 + glowPhase * 0.2),
            radius: reduceMotion ? 12 : 12 + glowPhase * 4,
            y: 6
          )
          .scaleEffect(pages[currentPage].isLastPage ? 1.05 : 1.0)
          .animation(.spring(response: 0.3), value: pages[currentPage].isLastPage)
        }
        .buttonStyle(PremiumButtonStyle())
      }
      .padding(.horizontal, 30)

      // Skip button with subtle interaction
      if !pages[currentPage].isLastPage {
        Button(action: {
          #if os(iOS)
          HapticManager.shared.selection()
          #endif
          onComplete()
        }) {
          Text("Skip for now")
            .font(.system(size: 14))
            .foregroundColor(.secondary.opacity(0.7))
        }
        .buttonStyle(SubtleButtonStyle())
        .transition(.opacity)
      }
    }
    .animation(.easeInOut(duration: 0.3), value: currentPage)
    .onAppear {
      if !reduceMotion {
        withAnimation(.easeInOut(duration: 2).repeatForever(autoreverses: true)) {
          glowPhase = 1
        }
      }
    }
  }
}

// MARK: - Premium Button Style (for main CTA)
struct PremiumButtonStyle: ButtonStyle {
  func makeBody(configuration: Configuration) -> some View {
    configuration.label
      .scaleEffect(configuration.isPressed ? 0.95 : 1.0)
      .brightness(configuration.isPressed ? -0.05 : 0)
      .animation(.spring(response: 0.2, dampingFraction: 0.6), value: configuration.isPressed)
      .onChange(of: configuration.isPressed) { _, isPressed in
        if isPressed {
          #if os(iOS)
          HapticManager.shared.impact(.soft)
          #endif
        }
      }
  }
}

// MARK: - Glass Button Style (for secondary buttons)
struct GlassButtonStyle: ButtonStyle {
  func makeBody(configuration: Configuration) -> some View {
    configuration.label
      .scaleEffect(configuration.isPressed ? 0.97 : 1.0)
      .opacity(configuration.isPressed ? 0.8 : 1.0)
      .animation(.spring(response: 0.2, dampingFraction: 0.7), value: configuration.isPressed)
  }
}

// MARK: - Subtle Button Style (for skip)
struct SubtleButtonStyle: ButtonStyle {
  func makeBody(configuration: Configuration) -> some View {
    configuration.label
      .opacity(configuration.isPressed ? 0.5 : 1.0)
      .animation(.easeOut(duration: 0.15), value: configuration.isPressed)
  }
}

// MARK: - Bounce Button Style
struct BounceButtonStyle: ButtonStyle {
  func makeBody(configuration: Configuration) -> some View {
    configuration.label
      .scaleEffect(configuration.isPressed ? 0.95 : 1.0)
      .animation(.spring(response: 0.2, dampingFraction: 0.6), value: configuration.isPressed)
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
  var testimonial: String? = nil
  var showTestimonials: Bool = false
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
  let geometry: GeometryProxy
  let pageIndex: Int
  var reduceMotion: Bool = false

  @State private var showIcon = false
  @State private var showTitle = false
  @State private var showFeatures = false
  @State private var currentTestimonialIndex = 0
  @State private var iconRotation: Double = 0
  @State private var iconScale: CGFloat = 0.3
  @State private var iconPulse: CGFloat = 1.0
  @State private var testimonialTimer: Timer?

  var body: some View {
    VStack(spacing: 24) {
      Spacer()

      // MARK: - Animated Icon with 3D Effects
      ZStack {
        // Outer pulsing rings with 3D perspective
        ForEach(0..<3) { i in
          Circle()
            .stroke(
              LinearGradient(
                colors: [page.iconColors[0].opacity(0.4), page.iconColors[1].opacity(0.1)],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
              ),
              lineWidth: 2
            )
            .frame(width: 100 + CGFloat(i * 30), height: 100 + CGFloat(i * 30))
            .scaleEffect(showIcon ? 1.0 + CGFloat(i) * 0.1 : 0.5)
            .opacity(showIcon ? 0.6 - Double(i) * 0.2 : 0)
            .rotation3DEffect(
              .degrees(reduceMotion ? 0 : Double(i) * 5),
              axis: (x: 1, y: 0, z: 0),
              perspective: 0.5
            )
            .animation(
              reduceMotion ? .none : .easeOut(duration: 1.2).delay(Double(i) * 0.15),
              value: showIcon
            )
        }

        // Outer glow with pulse
        Circle()
          .fill(
            RadialGradient(
              colors: [page.iconColors[0].opacity(0.4), .clear],
              center: .center,
              startRadius: 0,
              endRadius: 70
            )
          )
          .frame(width: 140, height: 140)
          .scaleEffect(reduceMotion ? 1.0 : (isAnimating ? 1.15 * iconPulse : 1.0))
          .animation(reduceMotion ? nil : .easeInOut(duration: 2.5).repeatForever(autoreverses: true), value: isAnimating)

        // Icon background with enhanced 3D effect
        ZStack {
          // Deep shadow layer
          Circle()
            .fill(Color.black.opacity(0.3))
            .frame(width: 88, height: 88)
            .offset(y: 8)
            .blur(radius: 12)

          // Mid shadow layer
          Circle()
            .fill(
              LinearGradient(
                colors: [page.iconColors[0].opacity(0.8), page.iconColors[1]],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
              )
            )
            .frame(width: 88, height: 88)
            .offset(y: 4)
            .blur(radius: 8)
            .opacity(0.5)

          // Main circle with glass effect
          Circle()
            .fill(
              LinearGradient(
                colors: page.iconColors,
                startPoint: .topLeading,
                endPoint: .bottomTrailing
              )
            )
            .frame(width: 88, height: 88)
            .overlay(
              // Top highlight (glass reflection)
              Circle()
                .fill(
                  LinearGradient(
                    colors: [.white.opacity(0.5), .white.opacity(0.1), .clear],
                    startPoint: .topLeading,
                    endPoint: .center
                  )
                )
                .frame(width: 88, height: 88)
            )
            .overlay(
              // Inner edge highlight
              Circle()
                .stroke(
                  LinearGradient(
                    colors: [.white.opacity(0.4), .clear],
                    startPoint: .topLeading,
                    endPoint: .bottomTrailing
                  ),
                  lineWidth: 1.5
                )
                .frame(width: 86, height: 86)
            )
            .shadow(color: page.iconColors[0].opacity(0.6), radius: 20, y: 10)
        }
        .scaleEffect(iconScale * (reduceMotion ? 1.0 : iconPulse))
        .rotation3DEffect(
          .degrees(reduceMotion ? 0 : iconRotation * 0.1),
          axis: (x: 0.5, y: 1, z: 0),
          perspective: 0.3
        )
        .rotationEffect(.degrees(iconRotation))

        // Icon with symbol effect
        Image(systemName: page.icon)
          .font(.system(size: 38, weight: .medium))
          .foregroundColor(.white)
          .scaleEffect(iconScale)
          .shadow(color: .black.opacity(0.2), radius: 2, y: 2)
          .symbolEffect(reduceMotion ? .pulse : .bounce, options: .repeating.speed(0.3), value: isAnimating)
      }
      .frame(height: 180)

      // MARK: - Title with stagger animation
      VStack(spacing: 8) {
        Text(page.title)
          .font(.system(size: 32, weight: .bold, design: .rounded))
          .foregroundColor(.primary)
          .multilineTextAlignment(.center)
          .opacity(showTitle ? 1 : 0)
          .offset(y: showTitle ? 0 : 20)

        Text(page.subtitle)
          .font(.system(size: 17, weight: .medium))
          .foregroundColor(.secondary)
          .multilineTextAlignment(.center)
          .padding(.horizontal, 40)
          .opacity(showTitle ? 1 : 0)
          .offset(y: showTitle ? 0 : 15)
          .animation(.spring(response: 0.6).delay(0.1), value: showTitle)
      }
      .animation(.spring(response: 0.6, dampingFraction: 0.8), value: showTitle)

      // MARK: - Feature List with stagger
      VStack(spacing: 14) {
        ForEach(Array(page.features.enumerated()), id: \.offset) { index, feature in
          FeatureRow(feature: feature, index: index, isVisible: showFeatures)
        }
      }
      .frame(maxWidth: 400)
      .padding(.top, 8)

      // MARK: - Mini Testimonial
      if let testimonial = page.testimonial {
        Text(testimonial)
          .font(.system(size: 14, weight: .medium, design: .serif))
          .italic()
          .foregroundColor(.secondary)
          .multilineTextAlignment(.center)
          .padding(.horizontal, 50)
          .padding(.top, 8)
          .opacity(showFeatures ? 1 : 0)
          .offset(y: showFeatures ? 0 : 10)
          .animation(.spring(response: 0.6).delay(0.5), value: showFeatures)
      }

      // MARK: - Full Testimonials Carousel (Final Page)
      if page.showTestimonials && !testimonials.isEmpty {
        VStack(spacing: 12) {
          TestimonialCard(testimonial: testimonials[currentTestimonialIndex])
            .frame(maxWidth: 420)
            .transition(.asymmetric(
              insertion: .move(edge: .trailing).combined(with: .opacity),
              removal: .move(edge: .leading).combined(with: .opacity)
            ))
            .id(currentTestimonialIndex)

          // Testimonial indicators
          HStack(spacing: 8) {
            ForEach(0..<min(testimonials.count, 5), id: \.self) { index in
              Capsule()
                .fill(currentTestimonialIndex == index ?
                  testimonials[index].color :
                  Color.secondary.opacity(0.3)
                )
                .frame(width: currentTestimonialIndex == index ? 20 : 8, height: 8)
                .animation(.spring(response: 0.3), value: currentTestimonialIndex)
                .onTapGesture {
                  withAnimation(.spring(response: 0.4)) {
                    currentTestimonialIndex = index
                  }
                }
            }
          }
        }
        .padding(.top, 8)
        .opacity(showFeatures ? 1 : 0)
        .offset(y: showFeatures ? 0 : 20)
        .animation(.spring(response: 0.6).delay(0.4), value: showFeatures)
        .onAppear {
          startTestimonialRotation()
        }
        .onDisappear {
          testimonialTimer?.invalidate()
        }
      }

      Spacer()
    }
    .padding()
    .onAppear {
      triggerEntranceAnimation()
    }
    .onChange(of: isAnimating) { _, newValue in
      if newValue {
        resetAndTriggerAnimation()
      }
    }
  }

  // MARK: - Animation Triggers
  private func triggerEntranceAnimation() {
    if reduceMotion {
      // Instant appearance for reduce motion
      showIcon = true
      iconScale = 1.0
      showTitle = true
      showFeatures = true
      return
    }

    // Icon entrance with spring
    withAnimation(.spring(response: 0.8, dampingFraction: 0.6).delay(0.1)) {
      showIcon = true
      iconScale = 1.0
    }

    // 3D rotation for premium feel
    withAnimation(.spring(response: 0.8, dampingFraction: 0.5).delay(0.1)) {
      iconRotation = 360
    }

    // Start subtle pulse animation
    withAnimation(.easeInOut(duration: 2).repeatForever(autoreverses: true).delay(0.8)) {
      iconPulse = 1.03
    }

    // Haptic on icon appearance
    #if os(iOS)
    DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) {
      HapticManager.shared.impact(.soft)
    }
    #endif

    // Title entrance
    DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) {
      withAnimation {
        showTitle = true
      }
    }

    // Features entrance with haptic per feature
    DispatchQueue.main.asyncAfter(deadline: .now() + 0.5) {
      withAnimation {
        showFeatures = true
      }
      #if os(iOS)
      // Subtle haptics as features appear
      for i in 0..<page.features.count {
        DispatchQueue.main.asyncAfter(deadline: .now() + Double(i) * 0.12) {
          HapticManager.shared.impactWithIntensity(0.3)
        }
      }
      #endif
    }
  }

  private func resetAndTriggerAnimation() {
    showIcon = false
    showTitle = false
    showFeatures = false
    iconScale = 0.3
    iconRotation = 0
    iconPulse = 1.0

    DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) {
      triggerEntranceAnimation()
    }
  }

  private func startTestimonialRotation() {
    testimonialTimer = Timer.scheduledTimer(withTimeInterval: 5.0, repeats: true) { _ in
      withAnimation(.spring(response: 0.5)) {
        currentTestimonialIndex = (currentTestimonialIndex + 1) % min(testimonials.count, 5)
      }
    }
  }
}

// MARK: - Feature Row
struct FeatureRow: View {
  let feature: FeatureItem
  let index: Int
  let isVisible: Bool

  var body: some View {
    HStack(spacing: 16) {
      // Icon with animated background
      ZStack {
        Circle()
          .fill(feature.color.opacity(0.15))
          .frame(width: 42, height: 42)

        Circle()
          .stroke(feature.color.opacity(0.3), lineWidth: 1.5)
          .frame(width: 42, height: 42)

        Image(systemName: feature.icon)
          .font(.system(size: 18, weight: .medium))
          .foregroundColor(feature.color)
      }

      Text(feature.text)
        .font(.system(size: 15, weight: .medium))
        .foregroundColor(.primary)

      Spacer()

      // Checkmark that appears
      Image(systemName: "checkmark.circle.fill")
        .font(.system(size: 18))
        .foregroundColor(feature.color.opacity(0.6))
        .opacity(isVisible ? 1 : 0)
        .scaleEffect(isVisible ? 1 : 0.5)
        .animation(.spring(response: 0.4).delay(Double(index) * 0.1 + 0.3), value: isVisible)
    }
    .padding(.horizontal, 18)
    .padding(.vertical, 12)
    .background(
      RoundedRectangle(cornerRadius: 14)
        .fill(Color.primary.opacity(0.03))
        .overlay(
          RoundedRectangle(cornerRadius: 14)
            .stroke(Color.primary.opacity(0.06), lineWidth: 1)
        )
    )
    .offset(x: isVisible ? 0 : 60)
    .opacity(isVisible ? 1 : 0)
    .animation(
      .spring(response: 0.6, dampingFraction: 0.75).delay(Double(index) * 0.12),
      value: isVisible
    )
  }
}

// MARK: - Testimonial Card
struct TestimonialCard: View {
  let testimonial: Testimonial
  @State private var isHovered = false

  private var cardBackgroundColor: Color {
    #if os(iOS)
    return Color(uiColor: .systemBackground).opacity(0.5)
    #else
    return Color(nsColor: .windowBackgroundColor).opacity(0.5)
    #endif
  }

  var body: some View {
    VStack(alignment: .leading, spacing: 12) {
      // Quote with feature badge
      HStack(alignment: .top, spacing: 10) {
        Image(systemName: "quote.opening")
          .font(.system(size: 20, weight: .bold))
          .foregroundColor(testimonial.color.opacity(0.5))

        VStack(alignment: .leading, spacing: 8) {
          Text(testimonial.quote)
            .font(.system(size: 14, weight: .medium, design: .serif))
            .italic()
            .foregroundColor(.primary.opacity(0.85))
            .lineLimit(3)
            .fixedSize(horizontal: false, vertical: true)
        }
      }

      // Attribution with feature tag
      HStack {
        // Feature tag
        Text(testimonial.feature)
          .font(.system(size: 11, weight: .semibold))
          .foregroundColor(testimonial.color)
          .padding(.horizontal, 10)
          .padding(.vertical, 4)
          .background(
            Capsule()
              .fill(testimonial.color.opacity(0.15))
          )

        Spacer()

        VStack(alignment: .trailing, spacing: 2) {
          Text(testimonial.name)
            .font(.system(size: 13, weight: .semibold))
            .foregroundColor(.primary)
          Text(testimonial.role)
            .font(.system(size: 11))
            .foregroundColor(.secondary)
        }
      }
    }
    .padding(18)
    .background(
      RoundedRectangle(cornerRadius: 16)
        .fill(cardBackgroundColor)
        .overlay(
          RoundedRectangle(cornerRadius: 16)
            .stroke(
              LinearGradient(
                colors: [testimonial.color.opacity(0.3), testimonial.color.opacity(0.1)],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
              ),
              lineWidth: 1.5
            )
        )
        .shadow(color: testimonial.color.opacity(0.1), radius: 12, y: 4)
    )
    .scaleEffect(isHovered ? 1.02 : 1.0)
    .animation(.spring(response: 0.3), value: isHovered)
    #if os(macOS)
    .onHover { hovering in
      isHovered = hovering
    }
    #endif
  }
}

#Preview("Welcome Walkthrough") {
  WelcomeWalkthroughView()
    .environmentObject(AppState())
    .frame(width: 600, height: 750)
}
