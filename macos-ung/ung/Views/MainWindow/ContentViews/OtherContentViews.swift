//
//  OtherContentViews.swift
//  ung
//
//  Created by Andrii Klymiuk on 28.11.2025.
//

import SwiftUI
#if os(macOS)
import AppKit
#endif
// MARK: - Database Document
import UniformTypeIdentifiers

// MARK: - Tracking Content
struct TrackingContent: View {
  @EnvironmentObject var appState: AppState
  @Binding var showAddSheet: Bool
  @State private var projectName = ""
  @State private var selectedClientId: Int?
  @State private var isStarting = false
  @State private var searchQuery = ""
  @State private var showDeleteConfirmation = false
  @State private var sessionToDelete: RecentSession?

  var filteredSessions: [RecentSession] {
    if searchQuery.isEmpty {
      return appState.recentSessions
    }
    return appState.recentSessions.filter {
      $0.project.localizedCaseInsensitiveContains(searchQuery)
    }
  }

  // Threshold for switching to vertical layout
  private let compactWidthThreshold: CGFloat = 750

  var body: some View {
    ZStack {
      GeometryReader { geometry in
        let isCompact = geometry.size.width < compactWidthThreshold

        ScrollView(showsIndicators: false) {
          VStack(spacing: Design.Spacing.lg) {
            // Active tracking banner
            if appState.isTracking, let session = appState.activeSession {
              activeTrackingBanner(session)
                .transition(.move(edge: .top).combined(with: .opacity))
            }

            if isCompact {
              // Vertical layout for narrow windows
              VStack(spacing: Design.Spacing.lg) {
                startTrackingForm
                recentSessionsList
              }
            } else {
              // Horizontal layout for wide windows
              HStack(alignment: .top, spacing: Design.Spacing.lg) {
                startTrackingForm
                  .frame(maxWidth: 400)
                recentSessionsList
              }
            }

            Spacer(minLength: Design.Spacing.lg)
          }
          .padding(Design.Spacing.lg)
        }
        .animation(Design.Animation.smooth, value: appState.isTracking)
      }

      // Delete confirmation overlay
      if showDeleteConfirmation, let session = sessionToDelete {
        Color.black.opacity(0.3)
          .ignoresSafeArea()
          .onTapGesture { showDeleteConfirmation = false }

        ConfirmationDialog(
          title: "Delete Session",
          message:
            "Are you sure you want to delete '\(session.project)'? This action cannot be undone.",
          destructive: true,
          confirmTitle: "Delete",
          onConfirm: {
            Task {
              try? await appState.database.deleteSession(id: Int64(session.id))
              await appState.refreshDashboard()
            }
            showDeleteConfirmation = false
          },
          onCancel: { showDeleteConfirmation = false }
        )
        .transition(.scale.combined(with: .opacity))
      }
    }
  }

  private func activeTrackingBanner(_ session: ActiveSession) -> some View {
    HStack(spacing: Design.Spacing.md) {
      // Pulsing indicator
      ZStack {
        Circle()
          .fill(Color.red.opacity(0.3))
          .frame(width: 20, height: 20)
          .scaleEffect(1.5)
          .opacity(0.5)

        Circle()
          .fill(Color.red)
          .frame(width: 12, height: 12)
      }

      VStack(alignment: .leading, spacing: Design.Spacing.xxxs) {
        Text(session.project)
          .font(Design.Typography.headingSmall)
          .foregroundColor(.white)
          .lineLimit(1)

        if !session.client.isEmpty {
          Text(session.client)
            .font(Design.Typography.bodySmall)
            .foregroundColor(.white.opacity(0.8))
            .lineLimit(1)
        }
      }

      Spacer()

      Text(session.formattedDuration)
        .font(.system(size: 32, weight: .bold, design: .monospaced))
        .foregroundColor(.white)

      Button(action: { Task { await appState.stopTracking() } }) {
        HStack(spacing: Design.Spacing.xs) {
          Image(systemName: "stop.fill")
          Text("Stop")
        }
        .font(Design.Typography.labelMedium)
        .foregroundColor(.white)
        .padding(.horizontal, Design.Spacing.md)
        .padding(.vertical, Design.Spacing.sm)
        .background(
          RoundedRectangle(cornerRadius: Design.Radius.sm)
            .fill(Color.white.opacity(0.2))
        )
      }
      .buttonStyle(.plain)
    }
    .padding(Design.Spacing.lg)
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.lg)
        .fill(
          LinearGradient(
            colors: [Design.Colors.error, Design.Colors.error.opacity(0.8)],
            startPoint: .leading,
            endPoint: .trailing
          )
        )
        .shadow(color: Design.Colors.error.opacity(0.3), radius: 16, y: 8)
    )
  }

  private var startTrackingForm: some View {
    AnimatedCard {
      VStack(alignment: .leading, spacing: Design.Spacing.md) {
        HStack {
          Image(systemName: "play.fill")
            .foregroundColor(Design.Colors.success)
          Text("Start Tracking")
            .font(Design.Typography.headingSmall)
        }

        FormField(
          label: "Project / Task", text: $projectName, placeholder: "What are you working on?")

        VStack(alignment: .leading, spacing: Design.Spacing.xxs) {
          Text("Client (Optional)")
            .font(Design.Typography.labelSmall)
            .foregroundColor(Design.Colors.textSecondary)

          Picker("", selection: $selectedClientId) {
            Text("No Client").tag(nil as Int?)
            ForEach(appState.clients) { client in
              Text(client.name).tag(client.id as Int?)
            }
          }
          .pickerStyle(.menu)
        }

        Button(action: startTracking) {
          HStack(spacing: Design.Spacing.xs) {
            if isStarting {
              ProgressView()
                .scaleEffect(0.8)
                .tint(.white)
            } else {
              Image(systemName: "play.fill")
            }
            Text("Start Tracking")
          }
        }
        .buttonStyle(
          DSPrimaryButtonStyle(color: projectName.isEmpty ? .gray : Design.Colors.success)
        )
        .disabled(projectName.isEmpty || isStarting || appState.isTracking)
      }
      .padding(Design.Spacing.md)
    }
  }

  private var recentSessionsList: some View {
    AnimatedCard {
      VStack(alignment: .leading, spacing: Design.Spacing.md) {
        HStack {
          Image(systemName: "clock.fill")
            .foregroundColor(Design.Colors.primary)
          Text("Recent Sessions")
            .font(Design.Typography.headingSmall)
          Spacer()
        }

        SearchBar(text: $searchQuery, placeholder: "Search sessions...")

        if filteredSessions.isEmpty {
          AnimatedEmptyState(
            icon: "clock.badge.questionmark",
            title: searchQuery.isEmpty ? "No Sessions Yet" : "No Results",
            message: searchQuery.isEmpty
              ? "Start tracking to see your sessions here" : "Try a different search term"
          )
          .frame(height: 200)
        } else {
          ScrollView {
            LazyVStack(spacing: Design.Spacing.xs) {
              ForEach(filteredSessions) { session in
                sessionRow(session)
              }
            }
          }
          .frame(maxHeight: 400)
        }
      }
      .padding(Design.Spacing.md)
    }
  }

  private func sessionRow(_ session: RecentSession) -> some View {
    AnimatedListRow(
      title: session.project,
      subtitle: session.date,
      leading: {
        Image(systemName: "clock.fill")
          .font(.system(size: 14))
          .foregroundColor(Design.Colors.primary)
          .frame(width: 32, height: 32)
          .background(Circle().fill(Design.Colors.primary.opacity(0.1)))
      },
      trailing: {
        HStack(spacing: Design.Spacing.sm) {
          Text(appState.secureMode ? "**:**" : session.duration)
            .font(Design.Typography.monoSmall)
            .foregroundColor(Design.Colors.primary)

          Menu {
            Button(role: .destructive) {
              sessionToDelete = session
              showDeleteConfirmation = true
            } label: {
              Label("Delete", systemImage: "trash")
            }
          } label: {
            Image(systemName: "ellipsis.circle")
              .font(.system(size: 16))
              .foregroundColor(Design.Colors.textTertiary)
          }
          .menuStyle(.borderlessButton)
          .frame(width: 24)
        }
      }
    )
  }

  private func startTracking() {
    isStarting = true
    Task {
      await appState.startTracking(project: projectName, clientId: selectedClientId)
      projectName = ""
      isStarting = false
    }
  }
}

// MARK: - Clients Content
struct ClientsContent: View {
  @EnvironmentObject var appState: AppState
  @Binding var showAddSheet: Bool
  @State private var showEditClient = false
  @State private var showDeleteConfirmation = false
  @State private var selectedClient: Client?
  @State private var searchQuery = ""

  // Form fields
  @State private var clientName = ""
  @State private var clientEmail = ""
  @State private var clientAddress = ""
  @State private var clientTaxId = ""

  var filteredClients: [Client] {
    if searchQuery.isEmpty {
      return appState.clients
    }
    return appState.clients.filter {
      $0.name.localizedCaseInsensitiveContains(searchQuery)
        || $0.email.localizedCaseInsensitiveContains(searchQuery)
    }
  }

  var body: some View {
    ZStack {
      VStack(spacing: Design.Spacing.md) {
        // Search bar only - no duplicate header
        if !appState.clients.isEmpty {
          HStack {
            SearchBar(text: $searchQuery, placeholder: "Search clients...")
            Spacer()
          }
          .padding(.horizontal, Design.Spacing.lg)
          .padding(.top, Design.Spacing.sm)
        }

        if filteredClients.isEmpty {
          Spacer()
          AnimatedEmptyState(
            icon: "person.2.fill",
            title: searchQuery.isEmpty ? "No Clients Yet" : "No Results",
            message: searchQuery.isEmpty
              ? "Add your first client to start tracking time and creating invoices"
              : "Try a different search term",
            actionTitle: searchQuery.isEmpty ? "Add Client" : nil,
            action: searchQuery.isEmpty ? { showAddSheet = true } : nil
          )
          Spacer()
        } else {
          ScrollView {
            LazyVGrid(columns: [GridItem(.adaptive(minimum: 300))], spacing: Design.Spacing.md) {
              ForEach(filteredClients) { client in
                ClientCardView(
                  client: client,
                  onEdit: {
                    selectedClient = client
                    clientName = client.name
                    clientEmail = client.email
                    clientAddress = client.address
                    clientTaxId = client.taxId
                    showEditClient = true
                  },
                  onDelete: {
                    selectedClient = client
                    showDeleteConfirmation = true
                  }
                )
              }
            }
            .padding(.horizontal, Design.Spacing.lg)
            .padding(.bottom, Design.Spacing.lg)
          }
        }
      }

      // Modals
      if showAddSheet || showEditClient {
        Color.black.opacity(0.3)
          .ignoresSafeArea()
          .onTapGesture {
            showAddSheet = false
            showEditClient = false
          }

        clientFormSheet
          .transition(.scale.combined(with: .opacity))
      }

      if showDeleteConfirmation, let client = selectedClient {
        Color.black.opacity(0.3)
          .ignoresSafeArea()
          .onTapGesture { showDeleteConfirmation = false }

        ConfirmationDialog(
          title: "Delete Client",
          message:
            "Are you sure you want to delete '\(client.name)'? This will also delete associated contracts and invoices.",
          destructive: true,
          confirmTitle: "Delete",
          onConfirm: {
            Task {
              try? await appState.database.deleteClient(id: Int64(client.id))
              await appState.refreshDashboard()
            }
            showDeleteConfirmation = false
          },
          onCancel: { showDeleteConfirmation = false }
        )
        .transition(.scale.combined(with: .opacity))
      }
    }
    .animation(Design.Animation.smooth, value: showAddSheet)
    .animation(Design.Animation.smooth, value: showEditClient)
    .animation(Design.Animation.smooth, value: showDeleteConfirmation)
    .onChange(of: showAddSheet) { _, newValue in
      if newValue {
        clientName = ""
        clientEmail = ""
        clientAddress = ""
        clientTaxId = ""
      }
    }
  }

  private var clientFormSheet: some View {
    ActionMenu(
      title: showEditClient ? "Edit Client" : "Add Client",
      isPresented: showEditClient ? $showEditClient : $showAddSheet
    ) {
      VStack(spacing: Design.Spacing.md) {
        FormField(label: "Name", text: $clientName, placeholder: "Client name")
        FormField(label: "Email", text: $clientEmail, placeholder: "client@example.com")
        FormField(label: "Address", text: $clientAddress, placeholder: "123 Business St, City")
        FormField(label: "Tax ID", text: $clientTaxId, placeholder: "XX-XXXXXXX")

        HStack(spacing: Design.Spacing.sm) {
          Button("Cancel") {
            showAddSheet = false
            showEditClient = false
          }
          .buttonStyle(DSSecondaryButtonStyle())
          .keyboardShortcut(.escape, modifiers: [])

          Button(showEditClient ? "Save" : "Add Client") {
            submitClient()
          }
          .buttonStyle(DSPrimaryButtonStyle())
          .disabled(clientName.isEmpty)
          .keyboardShortcut(.return, modifiers: .command)
        }
      }
    }
  }

  private func submitClient() {
    guard !clientName.isEmpty else { return }
    Task {
      if showEditClient, let client = selectedClient {
        var updatedClient = ClientModel(
          id: Int64(client.id),
          name: clientName,
          email: clientEmail
        )
        updatedClient.address = clientAddress.isEmpty ? nil : clientAddress
        updatedClient.taxId = clientTaxId.isEmpty ? nil : clientTaxId
        try? await appState.database.updateClient(updatedClient)
      } else {
        var newClient = ClientModel(
          name: clientName,
          email: clientEmail.isEmpty ? "" : clientEmail
        )
        newClient.address = clientAddress.isEmpty ? nil : clientAddress
        newClient.taxId = clientTaxId.isEmpty ? nil : clientTaxId
        _ = try? await appState.database.createClient(newClient)
      }
      await appState.refreshDashboard()
      showAddSheet = false
      showEditClient = false
    }
  }
}

struct ClientCardView: View {
  let client: Client
  let onEdit: () -> Void
  let onDelete: () -> Void
  @State private var isHovered = false
  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    HStack(spacing: Design.Spacing.sm) {
      // Avatar
      ZStack {
        Circle()
          .fill(Design.Colors.purple.opacity(0.15))
          .frame(width: 48, height: 48)

        Text(String(client.name.prefix(1)).uppercased())
          .font(Design.Typography.headingSmall)
          .foregroundColor(Design.Colors.purple)
      }

      VStack(alignment: .leading, spacing: Design.Spacing.xxxs) {
        Text(client.name)
          .font(Design.Typography.labelMedium)
          .foregroundColor(Design.Colors.textPrimary)

        if !client.email.isEmpty {
          Text(client.email)
            .font(Design.Typography.bodySmall)
            .foregroundColor(Design.Colors.textSecondary)
        }
      }

      Spacer()

      if isHovered {
        HStack(spacing: Design.Spacing.xs) {
          AnimatedIconButton(icon: "pencil", color: Design.Colors.primary, size: 28, action: onEdit)
          AnimatedIconButton(icon: "trash", color: Design.Colors.error, size: 28, action: onDelete)
        }
        .transition(.scale.combined(with: .opacity))
      }
    }
    .padding(Design.Spacing.md)
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.md)
        .fill(Design.Colors.surfaceElevated(colorScheme))
        .shadow(
          color: isHovered ? Design.Shadow.md.color : Design.Shadow.sm.color,
          radius: isHovered ? Design.Shadow.md.radius : Design.Shadow.sm.radius,
          y: isHovered ? Design.Shadow.md.y : Design.Shadow.sm.y
        )
    )
    .scaleEffect(isHovered ? 1.01 : 1.0)
    .animation(Design.Animation.smooth, value: isHovered)
    .onHover { hovering in isHovered = hovering }
  }
}

// MARK: - Contracts Content
struct ContractsContent: View {
  @EnvironmentObject var appState: AppState
  @Binding var showAddSheet: Bool
  @State private var showEditContract = false
  @State private var showDeleteConfirmation = false
  @State private var selectedContract: Contract?
  @State private var searchQuery = ""

  // Form fields
  @State private var contractName = ""
  @State private var selectedClientId: Int?
  @State private var rate: Double = 50
  @State private var price: Double = 0
  @State private var contractType = "hourly"
  @State private var contractCurrency = "USD"
  @State private var contractNotes = ""

  let contractTypes = ["hourly", "fixed_price", "retainer"]
  let currencies = ["USD", "EUR", "GBP", "CAD", "AUD", "CHF", "JPY", "PLN", "UAH"]

  var filteredContracts: [Contract] {
    if searchQuery.isEmpty {
      return appState.contracts
    }
    return appState.contracts.filter {
      $0.name.localizedCaseInsensitiveContains(searchQuery)
        || $0.clientName.localizedCaseInsensitiveContains(searchQuery)
    }
  }

  var body: some View {
    ZStack {
      VStack(spacing: Design.Spacing.md) {
        // Search bar only - no duplicate header
        if !appState.contracts.isEmpty {
          HStack {
            SearchBar(text: $searchQuery, placeholder: "Search contracts...")
            Spacer()
          }
          .padding(.horizontal, Design.Spacing.lg)
          .padding(.top, Design.Spacing.sm)
        }

        if appState.clients.isEmpty {
          Spacer()
          AnimatedEmptyState(
            icon: "person.badge.plus",
            title: "Add a Client First",
            message: "You need to add a client before creating contracts",
            actionTitle: "Add Client",
            action: { appState.selectedTab = .clients }
          )
          Spacer()
        } else if filteredContracts.isEmpty {
          Spacer()
          AnimatedEmptyState(
            icon: "doc.text.fill",
            title: searchQuery.isEmpty ? "No Contracts Yet" : "No Results",
            message: searchQuery.isEmpty
              ? "Create contracts to define rates and terms with your clients"
              : "Try a different search term",
            actionTitle: searchQuery.isEmpty ? "Create Contract" : nil,
            action: searchQuery.isEmpty ? { showAddSheet = true } : nil
          )
          Spacer()
        } else {
          ScrollView {
            LazyVGrid(columns: [GridItem(.adaptive(minimum: 320))], spacing: Design.Spacing.md) {
              ForEach(filteredContracts) { contract in
                ContractCardView(
                  contract: contract,
                  onEdit: {
                    selectedContract = contract
                    contractName = contract.name
                    rate = contract.rate
                    price = contract.price
                    contractType = contract.type
                    contractCurrency = contract.currency
                    contractNotes = contract.notes
                    showEditContract = true
                  },
                  onDelete: {
                    selectedContract = contract
                    showDeleteConfirmation = true
                  }
                )
              }
            }
            .padding(.horizontal, Design.Spacing.lg)
            .padding(.bottom, Design.Spacing.lg)
          }
        }
      }

      // Modals
      if showAddSheet || showEditContract {
        Color.black.opacity(0.3)
          .ignoresSafeArea()
          .onTapGesture {
            showAddSheet = false
            showEditContract = false
          }

        contractFormSheet
          .transition(.scale.combined(with: .opacity))
      }

      if showDeleteConfirmation, let contract = selectedContract {
        Color.black.opacity(0.3)
          .ignoresSafeArea()
          .onTapGesture { showDeleteConfirmation = false }

        ConfirmationDialog(
          title: "Delete Contract",
          message: "Are you sure you want to delete '\(contract.name)'?",
          destructive: true,
          confirmTitle: "Delete",
          onConfirm: {
            Task {
              try? await appState.database.deleteContract(id: Int64(contract.id))
              await appState.refreshDashboard()
            }
            showDeleteConfirmation = false
          },
          onCancel: { showDeleteConfirmation = false }
        )
        .transition(.scale.combined(with: .opacity))
      }
    }
    .animation(Design.Animation.smooth, value: showAddSheet)
    .animation(Design.Animation.smooth, value: showEditContract)
    .animation(Design.Animation.smooth, value: showDeleteConfirmation)
    .onChange(of: showAddSheet) { _, newValue in
      if newValue {
        contractName = ""
        selectedClientId = appState.clients.first?.id
        rate = 50
        price = 0
        contractType = "hourly"
        contractCurrency = "USD"
        contractNotes = ""
      }
    }
  }

  private var contractFormSheet: some View {
    ActionMenu(
      title: showEditContract ? "Edit Contract" : "New Contract",
      isPresented: showEditContract ? $showEditContract : $showAddSheet
    ) {
      VStack(spacing: Design.Spacing.md) {
        FormField(
          label: "Contract Name", text: $contractName, placeholder: "e.g., Development Services")

        if !showEditContract {
          VStack(alignment: .leading, spacing: Design.Spacing.xxs) {
            Text("Client")
              .font(Design.Typography.labelSmall)
              .foregroundColor(Design.Colors.textSecondary)

            Picker("", selection: $selectedClientId) {
              ForEach(appState.clients) { client in
                Text(client.name).tag(client.id as Int?)
              }
            }
            .pickerStyle(.menu)
          }
        }

        VStack(alignment: .leading, spacing: Design.Spacing.xxs) {
          Text("Contract Type")
            .font(Design.Typography.labelSmall)
            .foregroundColor(Design.Colors.textSecondary)

          Picker("", selection: $contractType) {
            Text("Hourly").tag("hourly")
            Text("Fixed Price").tag("fixed_price")
            Text("Retainer").tag("retainer")
          }
          .pickerStyle(.segmented)
        }

        if contractType == "hourly" || contractType == "retainer" {
          NumberStepper(
            label: "Hourly Rate", value: $rate, range: 1...1000, step: 5, format: "%.0f")
        }

        if contractType == "fixed_price" {
          NumberStepper(
            label: "Fixed Price", value: $price, range: 0...100000, step: 100, format: "%.0f")
        }

        VStack(alignment: .leading, spacing: Design.Spacing.xxs) {
          Text("Currency")
            .font(Design.Typography.labelSmall)
            .foregroundColor(Design.Colors.textSecondary)

          Picker("", selection: $contractCurrency) {
            ForEach(currencies, id: \.self) { currency in
              Text(currency).tag(currency)
            }
          }
          .pickerStyle(.menu)
        }

        if showEditContract {
          FormField(label: "Notes", text: $contractNotes, placeholder: "Contract notes...")
        }

        HStack(spacing: Design.Spacing.sm) {
          Button("Cancel") {
            showAddSheet = false
            showEditContract = false
          }
          .buttonStyle(DSSecondaryButtonStyle())
          .keyboardShortcut(.escape, modifiers: [])

          Button(showEditContract ? "Save" : "Create") {
            submitContract()
          }
          .buttonStyle(DSPrimaryButtonStyle())
          .disabled(contractName.isEmpty || (!showEditContract && selectedClientId == nil))
          .keyboardShortcut(.return, modifiers: .command)
        }
      }
    }
  }

  private func submitContract() {
    guard !contractName.isEmpty else { return }
    guard showEditContract || selectedClientId != nil else { return }
    Task {
      if showEditContract, let contract = selectedContract {
        var updatedContract = ContractModel(
          id: Int64(contract.id),
          contractNum: "",
          clientId: Int64(selectedClientId ?? 0),
          name: contractName,
          contractType: contractType,
          currency: contractCurrency,
          startDate: Date(),
          active: true
        )
        updatedContract.hourlyRate = (contractType == "hourly" || contractType == "retainer") ? rate : nil
        updatedContract.fixedPrice = contractType == "fixed_price" ? price : nil
        updatedContract.notes = contractNotes.isEmpty ? nil : contractNotes
        try? await appState.database.updateContract(updatedContract)
      } else if let clientId = selectedClientId {
        var newContract = ContractModel(
          contractNum: "",
          clientId: Int64(clientId),
          name: contractName,
          contractType: contractType,
          currency: contractCurrency,
          startDate: Date(),
          active: true
        )
        newContract.hourlyRate = (contractType == "hourly" || contractType == "retainer") ? rate : nil
        newContract.fixedPrice = contractType == "fixed_price" ? price : nil
        _ = try? await appState.database.createContract(newContract)
      }
      await appState.refreshDashboard()
      showAddSheet = false
      showEditContract = false
    }
  }
}

struct ContractCardView: View {
  let contract: Contract
  let onEdit: () -> Void
  let onDelete: () -> Void
  @State private var isHovered = false
  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    VStack(alignment: .leading, spacing: Design.Spacing.sm) {
      HStack {
        Image(systemName: "doc.text.fill")
          .foregroundColor(Design.Colors.indigo)
        Text(contract.name)
          .font(Design.Typography.labelMedium)
          .lineLimit(1)
        Spacer()

        if isHovered {
          HStack(spacing: Design.Spacing.xs) {
            AnimatedIconButton(
              icon: "pencil", color: Design.Colors.primary, size: 24, action: onEdit)
            AnimatedIconButton(
              icon: "trash", color: Design.Colors.error, size: 24, action: onDelete)
          }
          .transition(.scale.combined(with: .opacity))
        }
      }

      Text(contract.clientName)
        .font(Design.Typography.bodySmall)
        .foregroundColor(Design.Colors.textSecondary)

      HStack {
        DSBadge(text: contract.type.capitalized, color: Design.Colors.indigo)

        Spacer()

        Text("$\(Int(contract.rate))/hr")
          .font(Design.Typography.labelMedium)
          .foregroundColor(Design.Colors.success)
      }
    }
    .padding(Design.Spacing.md)
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.md)
        .fill(Design.Colors.surfaceElevated(colorScheme))
        .shadow(
          color: isHovered ? Design.Shadow.md.color : Design.Shadow.sm.color,
          radius: isHovered ? Design.Shadow.md.radius : Design.Shadow.sm.radius,
          y: isHovered ? Design.Shadow.md.y : Design.Shadow.sm.y
        )
    )
    .scaleEffect(isHovered ? 1.01 : 1.0)
    .animation(Design.Animation.smooth, value: isHovered)
    .onHover { hovering in isHovered = hovering }
  }
}

// MARK: - Invoices Content
struct InvoicesContent: View {
  @EnvironmentObject var appState: AppState
  @Binding var showAddSheet: Bool
  @State private var showDeleteConfirmation = false
  @State private var showRecurringSheet = false
  @State private var selectedInvoice: RecentInvoice?
  @State private var searchQuery = ""
  @State private var filterStatus = "all"
  @State private var selectedClientId: Int?

  let statusFilters = ["all", "pending", "sent", "paid", "overdue"]

  var filteredInvoices: [RecentInvoice] {
    var result = appState.recentInvoices

    if !searchQuery.isEmpty {
      result = result.filter {
        $0.invoiceNum.localizedCaseInsensitiveContains(searchQuery)
          || $0.client.localizedCaseInsensitiveContains(searchQuery)
      }
    }

    if filterStatus != "all" {
      result = result.filter { $0.status.lowercased() == filterStatus }
    }

    return result
  }

  // Threshold for switching to compact layout
  private let compactWidthThreshold: CGFloat = 750

  var body: some View {
    ZStack {
      GeometryReader { geometry in
        let isCompact = geometry.size.width < compactWidthThreshold

        VStack(spacing: Design.Spacing.lg) {
          // Stats row - responsive grid
          LazyVGrid(
            columns: [GridItem(.adaptive(minimum: isCompact ? 140 : 180))],
            spacing: Design.Spacing.md
          ) {
            AnimatedStatCard(
              title: "Total Invoices",
              value: "\(appState.invoiceCount)",
              icon: "doc.plaintext.fill",
              color: Design.Colors.primary
            )

            AnimatedStatCard(
              title: "Pending",
              value: appState.formatCurrency(appState.metrics.pendingAmount),
              icon: "clock.fill",
              color: Design.Colors.warning
            )

            AnimatedStatCard(
              title: "Overdue",
              value: appState.formatCurrency(appState.metrics.overdueAmount),
              icon: "exclamationmark.triangle.fill",
              color: Design.Colors.error
            )

            AnimatedStatCard(
              title: "Total Revenue",
              value: appState.formatCurrency(appState.metrics.totalRevenue),
              icon: "chart.line.uptrend.xyaxis",
              color: Design.Colors.success
            )
          }

          // Filter bar - responsive (no button - moved to header)
          if isCompact {
            VStack(spacing: Design.Spacing.sm) {
              AnimatedSegmentedControl(
                selection: $filterStatus,
                options: statusFilters.map { ($0, $0.capitalized) }
              )

              SearchBar(text: $searchQuery, placeholder: "Search invoices...")
            }
          } else {
            HStack(spacing: Design.Spacing.md) {
              AnimatedSegmentedControl(
                selection: $filterStatus,
                options: statusFilters.map { ($0, $0.capitalized) }
              )

              Spacer()

              // Recurring invoices button
              Button(action: { showRecurringSheet = true }) {
                HStack(spacing: 4) {
                  Image(systemName: "arrow.triangle.2.circlepath")
                    .font(.system(size: 11))
                  Text("Recurring")
                    .font(.system(size: 11, weight: .medium))
                }
                .foregroundColor(Design.Colors.primary)
                .padding(.horizontal, 10)
                .padding(.vertical, 6)
                .background(
                  RoundedRectangle(cornerRadius: 6)
                    .fill(Design.Colors.primary.opacity(0.1))
                )
              }
              .buttonStyle(.plain)

              SearchBar(text: $searchQuery, placeholder: "Search invoices...")
                .frame(maxWidth: 250)
            }
          }

          // Invoice list
          if filteredInvoices.isEmpty {
            Spacer()
            invoiceEmptyState
            Spacer()
          } else {
            ScrollView {
              LazyVStack(spacing: Design.Spacing.xs) {
                ForEach(filteredInvoices) { invoice in
                  InvoiceRowView(
                    invoice: invoice,
                    onMarkPaid: {
                      Task {
                        try? await appState.database.updateInvoiceStatus(id: Int64(invoice.id), status: "paid")
                        await appState.refreshDashboard()
                      }
                    },
                    onMarkSent: {
                      Task {
                        try? await appState.database.updateInvoiceStatus(id: Int64(invoice.id), status: "sent")
                        await appState.refreshDashboard()
                      }
                    },
                    onDelete: {
                      selectedInvoice = invoice
                      showDeleteConfirmation = true
                    }
                  )
                }
              }
            }
          }
        }
        .padding(Design.Spacing.lg)
      }

      // Modals
      if showAddSheet {
        Color.black.opacity(0.3)
          .ignoresSafeArea()
          .onTapGesture { showAddSheet = false }

        createInvoiceSheet
          .transition(.scale.combined(with: .opacity))
      }

      if showDeleteConfirmation, let invoice = selectedInvoice {
        Color.black.opacity(0.3)
          .ignoresSafeArea()
          .onTapGesture { showDeleteConfirmation = false }

        ConfirmationDialog(
          title: "Delete Invoice",
          message: "Are you sure you want to delete invoice '\(invoice.invoiceNum)'?",
          destructive: true,
          confirmTitle: "Delete",
          onConfirm: {
            Task {
              try? await appState.database.deleteInvoice(id: Int64(invoice.id))
              await appState.refreshDashboard()
            }
            showDeleteConfirmation = false
          },
          onCancel: { showDeleteConfirmation = false }
        )
        .transition(.scale.combined(with: .opacity))
      }
    }
    .animation(Design.Animation.smooth, value: showAddSheet)
    .animation(Design.Animation.smooth, value: showDeleteConfirmation)
    .animation(Design.Animation.smooth, value: filterStatus)
    .sheet(isPresented: $showRecurringSheet) {
      RecurringInvoicesSheet()
        .environmentObject(appState)
    }
  }

  private var createInvoiceSheet: some View {
    ActionMenu(title: "Create Invoice", isPresented: $showAddSheet) {
      VStack(spacing: Design.Spacing.md) {
        VStack(alignment: .leading, spacing: Design.Spacing.xxs) {
          Text("Client")
            .font(Design.Typography.labelSmall)
            .foregroundColor(Design.Colors.textSecondary)

          Picker("", selection: $selectedClientId) {
            Text("Select Client").tag(nil as Int?)
            ForEach(appState.clients) { client in
              Text(client.name).tag(client.id as Int?)
            }
          }
          .pickerStyle(.menu)
        }

        Text("This will create an invoice from unbilled sessions for the selected client.")
          .font(Design.Typography.bodySmall)
          .foregroundColor(Design.Colors.textSecondary)

        HStack(spacing: Design.Spacing.sm) {
          Button("Cancel") {
            showAddSheet = false
          }
          .buttonStyle(DSSecondaryButtonStyle())
          .keyboardShortcut(.escape, modifiers: [])

          Button("Create Invoice") {
            submitInvoice()
          }
          .buttonStyle(DSPrimaryButtonStyle())
          .disabled(selectedClientId == nil)
          .keyboardShortcut(.return, modifiers: .command)
        }
      }
    }
  }

  private func submitInvoice() {
    guard let clientId = selectedClientId else { return }
    Task {
      guard let company = try? await appState.database.getCompany(), let companyId = company?.id else {
        return
      }
      let year = Calendar.current.component(.year, from: Date())
      let count = try? await appState.database.getInvoiceCount()
      let invoiceNum = "INV-\(year)-\(String(format: "%04d", (count ?? 0) + 1))"

      var invoice = Invoice(
        invoiceNum: invoiceNum,
        companyId: companyId,
        amount: 0,
        currency: "USD",
        status: "pending"
      )
      invoice.issuedDate = Date()
      _ = try? await appState.database.createInvoice(invoice)
      await appState.refreshDashboard()
      showAddSheet = false
    }
  }

  // MARK: - Invoice Empty State with Prerequisites
  @ViewBuilder
  private var invoiceEmptyState: some View {
    if searchQuery.isEmpty && filterStatus == "all" {
      // Check prerequisites and show appropriate guidance
      if !appState.setupStatus.hasCompany {
        AnimatedEmptyState(
          icon: "building.2.fill",
          title: "Set Up Your Company",
          message: "Add your company details first to create professional invoices",
          actionTitle: "Go to Settings",
          action: { appState.selectedTab = .settings }
        )
      } else if appState.clients.isEmpty {
        AnimatedEmptyState(
          icon: "person.2.fill",
          title: "Add a Client First",
          message: "You need at least one client before you can create invoices",
          actionTitle: "Add Client",
          action: { appState.selectedTab = .clients }
        )
      } else if appState.contracts.isEmpty {
        AnimatedEmptyState(
          icon: "doc.text.fill",
          title: "Create a Contract First",
          message: "Contracts define your rates and terms. Create one to invoice your clients",
          actionTitle: "Add Contract",
          action: { appState.selectedTab = .contracts }
        )
      } else {
        AnimatedEmptyState(
          icon: "doc.plaintext.fill",
          title: "No Invoices Yet",
          message: "Create your first invoice from tracked time sessions",
          actionTitle: "Create Invoice",
          action: { showAddSheet = true }
        )
      }
    } else {
      AnimatedEmptyState(
        icon: "doc.plaintext.fill",
        title: "No Results",
        message: "Try different filters or search terms",
        actionTitle: nil,
        action: nil
      )
    }
  }
}

struct InvoiceRowView: View {
  @EnvironmentObject var appState: AppState
  let invoice: RecentInvoice
  let onMarkPaid: () -> Void
  let onMarkSent: () -> Void
  let onDelete: () -> Void
  @State private var isHovered = false
  @State private var isExporting = false
  @State private var showExportSuccess = false
  @State private var showExportError = false
  @State private var exportedURL: URL?
  @Environment(\.colorScheme) var colorScheme

  var statusColor: Color {
    switch invoice.status.lowercased() {
    case "paid": return Design.Colors.success
    case "sent", "pending": return Design.Colors.warning
    case "overdue": return Design.Colors.error
    default: return Design.Colors.textTertiary
    }
  }

  var body: some View {
    HStack(spacing: Design.Spacing.md) {
      // Invoice icon
      ZStack {
        RoundedRectangle(cornerRadius: Design.Radius.xs)
          .fill(statusColor.opacity(0.1))
          .frame(width: 40, height: 40)

        Image(systemName: "doc.plaintext.fill")
          .font(.system(size: 16))
          .foregroundColor(statusColor)
      }

      VStack(alignment: .leading, spacing: Design.Spacing.xxxs) {
        Text(invoice.invoiceNum)
          .font(Design.Typography.labelMedium)
        Text(invoice.client)
          .font(Design.Typography.bodySmall)
          .foregroundColor(Design.Colors.textSecondary)
      }

      Spacer()

      Text(appState.secureMode ? "****" : invoice.amount)
        .font(Design.Typography.headingSmall)

      DSBadge(text: invoice.status.capitalized, color: statusColor)

      if isHovered {
        HStack(spacing: Design.Spacing.xs) {
          // Export PDF button
          if isExporting {
            ProgressView()
              .scaleEffect(0.6)
              .frame(width: 28, height: 28)
          } else {
            AnimatedIconButton(
              icon: "arrow.down.doc", color: Design.Colors.primary, size: 28, action: exportPDF)
          }

          if invoice.status.lowercased() != "paid" {
            AnimatedIconButton(
              icon: "checkmark.circle", color: Design.Colors.success, size: 28, action: onMarkPaid)
          }
          if invoice.status.lowercased() == "pending" {
            AnimatedIconButton(
              icon: "paperplane", color: Design.Colors.primary, size: 28, action: onMarkSent)
          }
          AnimatedIconButton(icon: "trash", color: Design.Colors.error, size: 28, action: onDelete)
        }
        .transition(.scale.combined(with: .opacity))
      }
    }
    .padding(Design.Spacing.md)
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.md)
        .fill(Design.Colors.surfaceElevated(colorScheme))
        .shadow(
          color: isHovered ? Design.Shadow.md.color : Design.Shadow.sm.color,
          radius: isHovered ? Design.Shadow.md.radius : Design.Shadow.sm.radius,
          y: isHovered ? Design.Shadow.md.y : Design.Shadow.sm.y
        )
    )
    .animation(Design.Animation.smooth, value: isHovered)
    .onHover { hovering in isHovered = hovering }
    .alert("PDF Exported", isPresented: $showExportSuccess) {
      Button("Open") {
        if let url = exportedURL {
          #if os(macOS)
          NSWorkspace.shared.open(url)
          #else
          // On iOS, we could use share sheet instead
          #endif
        }
      }
      Button("OK", role: .cancel) {}
    } message: {
      Text("Invoice PDF saved successfully.")
    }
    .alert("Export Failed", isPresented: $showExportError) {
      Button("OK", role: .cancel) {}
    } message: {
      Text("Failed to generate PDF. Please try again.")
    }
  }

  private func exportPDF() {
    isExporting = true
    Task {
      do {
        // Get the full invoice data
        guard let invoiceData = try await appState.database.getInvoice(id: Int64(invoice.id)) else {
          showExportError = true
          isExporting = false
          return
        }

        // Get company
        guard let company = try await appState.database.getCompany() else {
          showExportError = true
          isExporting = false
          return
        }

        // Get client for this invoice
        guard let client = try await appState.database.getInvoiceClient(invoiceId: Int64(invoice.id)) else {
          // Create a placeholder client from invoice data if no recipient linked
          let placeholderClient = ClientModel(
            name: invoice.client,
            email: ""
          )
          let lineItems = try await appState.database.getInvoiceLineItems(invoiceId: Int64(invoice.id))

          let pdfGenerator = PDFGenerator()
          if let url = pdfGenerator.saveInvoicePDF(
            invoice: invoiceData,
            company: company,
            client: placeholderClient,
            lineItems: lineItems
          ) {
            exportedURL = url
            try await appState.database.updateInvoicePDFPath(id: Int64(invoice.id), pdfPath: url.path)
            showExportSuccess = true
          } else {
            showExportError = true
          }
          isExporting = false
          return
        }

        // Get line items
        let lineItems = try await appState.database.getInvoiceLineItems(invoiceId: Int64(invoice.id))

        // Generate PDF
        let pdfGenerator = PDFGenerator()
        if let url = pdfGenerator.saveInvoicePDF(
          invoice: invoiceData,
          company: company,
          client: client,
          lineItems: lineItems
        ) {
          exportedURL = url
          try await appState.database.updateInvoicePDFPath(id: Int64(invoice.id), pdfPath: url.path)
          showExportSuccess = true
        } else {
          showExportError = true
        }
      } catch {
        print("PDF export error: \(error)")
        showExportError = true
      }
      isExporting = false
    }
  }
}

// MARK: - Expenses Content
struct ExpensesContent: View {
  @EnvironmentObject var appState: AppState
  @Binding var showAddSheet: Bool
  @State private var showEditExpense = false
  @State private var showDeleteConfirmation = false
  @State private var selectedExpense: RecentExpense?
  @State private var searchQuery = ""

  // Form fields
  @State private var expenseDescription = ""
  @State private var expenseAmount: Double = 0
  @State private var expenseCategory = "Software"

  let categories = [
    "Software", "Hardware", "Travel", "Meals", "Office", "Marketing", "Education", "Other",
  ]

  var filteredExpenses: [RecentExpense] {
    if searchQuery.isEmpty {
      return appState.recentExpenses
    }
    return appState.recentExpenses.filter {
      $0.description.localizedCaseInsensitiveContains(searchQuery)
        || $0.category.localizedCaseInsensitiveContains(searchQuery)
    }
  }

  var body: some View {
    ZStack {
      VStack(spacing: Design.Spacing.lg) {
        // Search bar only - no duplicate header
        if !appState.recentExpenses.isEmpty {
          HStack {
            SearchBar(text: $searchQuery, placeholder: "Search expenses...")
            Spacer()
          }
          .padding(.horizontal, Design.Spacing.lg)
          .padding(.top, Design.Spacing.sm)
        }

        if filteredExpenses.isEmpty {
          Spacer()
          AnimatedEmptyState(
            icon: "dollarsign.circle.fill",
            title: searchQuery.isEmpty ? "No Expenses Yet" : "No Results",
            message: searchQuery.isEmpty
              ? "Track your business expenses for tax deductions" : "Try a different search term",
            actionTitle: searchQuery.isEmpty ? "Log Expense" : nil,
            action: searchQuery.isEmpty ? { showAddSheet = true } : nil
          )
          Spacer()
        } else {
          ScrollView {
            LazyVStack(spacing: Design.Spacing.xs) {
              ForEach(filteredExpenses) { expense in
                ExpenseRowView(
                  expense: expense,
                  onEdit: {
                    selectedExpense = expense
                    expenseDescription = expense.description
                    expenseAmount =
                      Double(
                        expense.amount.replacingOccurrences(of: "$", with: "").replacingOccurrences(
                          of: ",", with: "")) ?? 0
                    expenseCategory = expense.category
                    showEditExpense = true
                  },
                  onDelete: {
                    selectedExpense = expense
                    showDeleteConfirmation = true
                  }
                )
              }
            }
          }
        }
      }
      .padding(Design.Spacing.lg)

      // Modals
      if showAddSheet || showEditExpense {
        Color.black.opacity(0.3)
          .ignoresSafeArea()
          .onTapGesture {
            showAddSheet = false
            showEditExpense = false
          }

        expenseFormSheet
          .transition(.scale.combined(with: .opacity))
      }

      if showDeleteConfirmation, let expense = selectedExpense {
        Color.black.opacity(0.3)
          .ignoresSafeArea()
          .onTapGesture { showDeleteConfirmation = false }

        ConfirmationDialog(
          title: "Delete Expense",
          message: "Are you sure you want to delete '\(expense.description)'?",
          destructive: true,
          confirmTitle: "Delete",
          onConfirm: {
            Task {
              try? await appState.database.deleteExpense(id: Int64(expense.id))
              await appState.refreshDashboard()
            }
            showDeleteConfirmation = false
          },
          onCancel: { showDeleteConfirmation = false }
        )
        .transition(.scale.combined(with: .opacity))
      }
    }
    .animation(Design.Animation.smooth, value: showAddSheet)
    .animation(Design.Animation.smooth, value: showEditExpense)
    .animation(Design.Animation.smooth, value: showDeleteConfirmation)
    .onChange(of: showAddSheet) { _, newValue in
      if newValue {
        expenseDescription = ""
        expenseAmount = 0
        expenseCategory = "Software"
      }
    }
  }

  private var expenseFormSheet: some View {
    ActionMenu(
      title: showEditExpense ? "Edit Expense" : "Log Expense",
      isPresented: showEditExpense ? $showEditExpense : $showAddSheet
    ) {
      VStack(spacing: Design.Spacing.md) {
        FormField(label: "Description", text: $expenseDescription, placeholder: "What did you buy?")

        NumberStepper(
          label: "Amount", value: $expenseAmount, range: 0...100000, step: 10, format: "$%.2f")

        VStack(alignment: .leading, spacing: Design.Spacing.xxs) {
          Text("Category")
            .font(Design.Typography.labelSmall)
            .foregroundColor(Design.Colors.textSecondary)

          Picker("", selection: $expenseCategory) {
            ForEach(categories, id: \.self) { category in
              Text(category).tag(category)
            }
          }
          .pickerStyle(.menu)
        }

        HStack(spacing: Design.Spacing.sm) {
          Button("Cancel") {
            showAddSheet = false
            showEditExpense = false
          }
          .buttonStyle(DSSecondaryButtonStyle())
          .keyboardShortcut(.escape, modifiers: [])

          Button(showEditExpense ? "Save" : "Log Expense") {
            submitExpense()
          }
          .buttonStyle(DSPrimaryButtonStyle(color: Design.Colors.warning))
          .disabled(expenseDescription.isEmpty || expenseAmount <= 0)
          .keyboardShortcut(.return, modifiers: .command)
        }
      }
    }
  }

  private func submitExpense() {
    guard !expenseDescription.isEmpty && expenseAmount > 0 else { return }
    Task {
      if showEditExpense, let expense = selectedExpense {
        var updatedExpense = Expense(
          id: Int64(expense.id),
          description: expenseDescription,
          amount: expenseAmount,
          currency: "USD",
          category: expenseCategory,
          date: Date()
        )
        try? await appState.database.updateExpense(updatedExpense)
      } else {
        let newExpense = Expense(
          description: expenseDescription,
          amount: expenseAmount,
          currency: "USD",
          category: expenseCategory,
          date: Date()
        )
        _ = try? await appState.database.createExpense(newExpense)
      }
      await appState.refreshDashboard()
      showAddSheet = false
      showEditExpense = false
    }
  }
}

struct ExpenseRowView: View {
  @EnvironmentObject var appState: AppState
  let expense: RecentExpense
  let onEdit: () -> Void
  let onDelete: () -> Void
  @State private var isHovered = false
  @Environment(\.colorScheme) var colorScheme

  var categoryIcon: String {
    switch expense.category.lowercased() {
    case "software": return "app.badge.fill"
    case "hardware": return "desktopcomputer"
    case "travel": return "airplane"
    case "meals": return "fork.knife"
    case "office": return "building.2.fill"
    case "marketing": return "megaphone.fill"
    case "education": return "book.fill"
    default: return "tag.fill"
    }
  }

  var body: some View {
    HStack(spacing: Design.Spacing.md) {
      ZStack {
        RoundedRectangle(cornerRadius: Design.Radius.xs)
          .fill(Design.Colors.warning.opacity(0.1))
          .frame(width: 40, height: 40)

        Image(systemName: categoryIcon)
          .font(.system(size: 16))
          .foregroundColor(Design.Colors.warning)
      }

      VStack(alignment: .leading, spacing: Design.Spacing.xxxs) {
        Text(expense.description)
          .font(Design.Typography.labelMedium)
        Text(expense.category)
          .font(Design.Typography.bodySmall)
          .foregroundColor(Design.Colors.textSecondary)
      }

      Spacer()

      Text(appState.secureMode ? "****" : expense.amount)
        .font(Design.Typography.headingSmall)
        .foregroundColor(Design.Colors.warning)

      if isHovered {
        HStack(spacing: Design.Spacing.xs) {
          AnimatedIconButton(icon: "pencil", color: Design.Colors.primary, size: 28, action: onEdit)
          AnimatedIconButton(icon: "trash", color: Design.Colors.error, size: 28, action: onDelete)
        }
        .transition(.scale.combined(with: .opacity))
      }
    }
    .padding(Design.Spacing.md)
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.md)
        .fill(Design.Colors.surfaceElevated(colorScheme))
        .shadow(
          color: isHovered ? Design.Shadow.md.color : Design.Shadow.sm.color,
          radius: isHovered ? Design.Shadow.md.radius : Design.Shadow.sm.radius,
          y: isHovered ? Design.Shadow.md.y : Design.Shadow.sm.y
        )
    )
    .animation(Design.Animation.smooth, value: isHovered)
    .onHover { hovering in isHovered = hovering }
  }
}

// MARK: - Reports Section Picker (Responsive)
struct ReportsSectionPicker: View {
  @Binding var selection: String
  let isCompact: Bool
  @Environment(\.colorScheme) var colorScheme

  private let sections: [(id: String, label: String, icon: String)] = [
    ("overview", "Overview", "chart.bar.xaxis"),
    ("calculator", "Rate Calculator", "dollarsign.circle"),
    ("goals", "Income Goals", "target"),
  ]

  var body: some View {
    HStack(spacing: isCompact ? Design.Spacing.xs : Design.Spacing.xxs) {
      ForEach(sections, id: \.id) { section in
        Button(action: { selection = section.id }) {
          if isCompact {
            // Icon-only for compact mode
            Image(systemName: section.icon)
              .font(.system(size: 14, weight: selection == section.id ? .semibold : .regular))
              .foregroundColor(selection == section.id ? .white : Design.Colors.textSecondary)
              .frame(width: 36, height: 28)
              .background(
                RoundedRectangle(cornerRadius: Design.Radius.sm)
                  .fill(selection == section.id ? Design.Colors.primary : Color.clear)
              )
          } else {
            // Full label for wide mode
            HStack(spacing: Design.Spacing.xxs) {
              Image(systemName: section.icon)
                .font(.system(size: 12))
              Text(section.label)
                .font(Design.Typography.labelSmall)
            }
            .foregroundColor(selection == section.id ? .white : Design.Colors.textSecondary)
            .padding(.horizontal, Design.Spacing.sm)
            .padding(.vertical, Design.Spacing.xs)
            .background(
              RoundedRectangle(cornerRadius: Design.Radius.sm)
                .fill(selection == section.id ? Design.Colors.primary : Color.clear)
            )
          }
        }
        .buttonStyle(.plain)
      }
    }
    .padding(Design.Spacing.xxxs)
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.md)
        .fill(Design.Colors.surfaceElevated(colorScheme))
    )
    .animation(Design.Animation.smooth, value: selection)
  }
}

// MARK: - Reports Content
struct ReportsContent: View {
  @EnvironmentObject var appState: AppState
  @State private var selectedPeriod = "week"
  @State private var selectedSection = "overview"

  let periods = [("week", "Week"), ("month", "Month"), ("year", "Year")]

  // Threshold for compact mode
  private let compactWidthThreshold: CGFloat = 700

  var body: some View {
    GeometryReader { geometry in
      let isCompact = geometry.size.width < compactWidthThreshold

      ScrollView(showsIndicators: false) {
        VStack(spacing: Design.Spacing.lg) {
          // Header - responsive
          if isCompact {
            VStack(alignment: .center, spacing: Design.Spacing.sm) {
              Text("Analytics & Reports")
                .font(Design.Typography.displaySmall)

              // Icon-only tabs for compact - centered
              ReportsSectionPicker(selection: $selectedSection, isCompact: true)
            }
            .frame(maxWidth: .infinity)
          } else {
            HStack {
              Text("Analytics & Reports")
                .font(Design.Typography.displaySmall)

              Spacer()

              // Full text tabs for wide
              ReportsSectionPicker(selection: $selectedSection, isCompact: false)
            }
          }

          switch selectedSection {
          case "calculator":
            RateCalculatorView()
          case "goals":
            IncomeGoalsView()
          default:
            overviewContent(isCompact: isCompact)
          }
        }
        .padding(Design.Spacing.lg)
      }
    }
    .animation(Design.Animation.smooth, value: selectedSection)
  }

  private func overviewContent(isCompact: Bool) -> some View {
    VStack(spacing: Design.Spacing.lg) {
      // Period filter - centered in compact mode
      if isCompact {
        VStack(spacing: Design.Spacing.sm) {
          Text("Period")
            .font(Design.Typography.labelMedium)
            .foregroundColor(Design.Colors.textSecondary)
          AnimatedSegmentedControl(selection: $selectedPeriod, options: periods)
        }
      } else {
        HStack {
          Text("Period")
            .font(Design.Typography.labelMedium)
            .foregroundColor(Design.Colors.textSecondary)
          AnimatedSegmentedControl(selection: $selectedPeriod, options: periods)
          Spacer()
        }
      }

      // Metrics overview - 2x2 grid in compact mode
      LazyVGrid(
        columns: [GridItem(.adaptive(minimum: isCompact ? 140 : 160))],
        spacing: Design.Spacing.md
      ) {
        AnimatedStatCard(
          title: "Total Revenue",
          value: appState.formatCurrency(appState.metrics.totalRevenue),
          icon: "chart.line.uptrend.xyaxis",
          color: Design.Colors.success
        )

        AnimatedStatCard(
          title: "Hours Tracked",
          value: appState.formatHours(appState.metrics.weeklyHours * 4),
          icon: "clock.fill",
          color: Design.Colors.primary
        )

        AnimatedStatCard(
          title: "Active Clients",
          value: "\(appState.clientCount)",
          icon: "person.2.fill",
          color: Design.Colors.purple
        )

        AnimatedStatCard(
          title: "Invoices",
          value: "\(appState.invoiceCount)",
          icon: "doc.plaintext.fill",
          color: Design.Colors.teal
        )
      }

      // Coming soon placeholder - smaller in compact mode
      VStack(spacing: Design.Spacing.md) {
        ZStack {
          Circle()
            .fill(Color.gray.opacity(0.08))
            .frame(width: isCompact ? 70 : 100, height: isCompact ? 70 : 100)

          Image(systemName: "chart.bar.xaxis")
            .font(.system(size: isCompact ? 24 : 36, weight: .light))
            .foregroundColor(.secondary.opacity(0.5))
        }

        VStack(spacing: Design.Spacing.xs) {
          Text("Detailed Reports Coming Soon")
            .font(Design.Typography.headingSmall)
            .foregroundColor(Design.Colors.textPrimary)

          Text(
            "Advanced analytics, charts, and exportable reports will be available in a future update"
          )
          .font(Design.Typography.bodySmall)
          .foregroundColor(Design.Colors.textSecondary)
          .multilineTextAlignment(.center)
          .fixedSize(horizontal: false, vertical: true)
        }
      }
      .padding(.vertical, Design.Spacing.xl)
    }
  }
}

// MARK: - Rate Calculator View
struct RateCalculatorView: View {
  @State private var desiredAnnualIncome: Double = 100000
  @State private var workWeeksPerYear: Double = 48
  @State private var billableHoursPerWeek: Double = 30
  @State private var businessExpenses: Double = 10000
  @State private var taxRate: Double = 30
  @State private var profitMargin: Double = 20

  var grossIncomeNeeded: Double {
    let afterTaxNeeded = desiredAnnualIncome + businessExpenses
    return afterTaxNeeded / (1 - taxRate / 100)
  }

  var withProfitMargin: Double {
    return grossIncomeNeeded * (1 + profitMargin / 100)
  }

  var totalBillableHours: Double {
    return workWeeksPerYear * billableHoursPerWeek
  }

  var calculatedRate: Double {
    guard totalBillableHours > 0 else { return 0 }
    return withProfitMargin / totalBillableHours
  }

  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    ScrollView {
      VStack(spacing: Design.Spacing.lg) {
        // Result Card
        AnimatedCard {
          VStack(spacing: Design.Spacing.md) {
            HStack {
              Image(systemName: "dollarsign.circle.fill")
                .font(.system(size: 24))
                .foregroundColor(Design.Colors.success)
              Text("Recommended Hourly Rate")
                .font(Design.Typography.headingSmall)
              Spacer()
            }

            HStack(alignment: .firstTextBaseline, spacing: Design.Spacing.xxs) {
              Text("$")
                .font(Design.Typography.displaySmall)
                .foregroundColor(Design.Colors.success)
              Text(String(format: "%.0f", calculatedRate))
                .font(.system(size: 56, weight: .bold, design: .rounded))
                .foregroundColor(Design.Colors.success)
              Text("/hour")
                .font(Design.Typography.bodyMedium)
                .foregroundColor(Design.Colors.textSecondary)
            }

            Divider()

            HStack(spacing: Design.Spacing.lg) {
              VStack(spacing: Design.Spacing.xxxs) {
                Text(String(format: "%.0f", totalBillableHours))
                  .font(Design.Typography.headingSmall)
                Text("Billable Hours/Year")
                  .font(Design.Typography.labelSmall)
                  .foregroundColor(Design.Colors.textSecondary)
              }

              Divider().frame(height: 40)

              VStack(spacing: Design.Spacing.xxxs) {
                Text(String(format: "$%.0f", grossIncomeNeeded))
                  .font(Design.Typography.headingSmall)
                Text("Gross Needed")
                  .font(Design.Typography.labelSmall)
                  .foregroundColor(Design.Colors.textSecondary)
              }

              Divider().frame(height: 40)

              VStack(spacing: Design.Spacing.xxxs) {
                Text(String(format: "$%.0f", withProfitMargin))
                  .font(Design.Typography.headingSmall)
                Text("Target Revenue")
                  .font(Design.Typography.labelSmall)
                  .foregroundColor(Design.Colors.textSecondary)
              }
            }
          }
          .padding(Design.Spacing.lg)
        }

        // Input Cards - responsive layout
        LazyVGrid(
          columns: [GridItem(.adaptive(minimum: 280), alignment: .top)],
          spacing: Design.Spacing.md
        ) {
          // Income & Expenses
          AnimatedCard {
            VStack(alignment: .leading, spacing: Design.Spacing.md) {
              HStack {
                Image(systemName: "banknote")
                  .foregroundColor(Design.Colors.primary)
                Text("Income & Expenses")
                  .font(Design.Typography.labelMedium)
              }

              NumberStepper(
                label: "Desired Annual Income",
                value: $desiredAnnualIncome,
                range: 10000...1_000_000,
                step: 5000,
                format: "$%.0f"
              )

              NumberStepper(
                label: "Business Expenses",
                value: $businessExpenses,
                range: 0...500000,
                step: 1000,
                format: "$%.0f"
              )

              NumberStepper(
                label: "Tax Rate",
                value: $taxRate,
                range: 0...60,
                step: 5,
                format: "%.0f%%"
              )

              NumberStepper(
                label: "Profit Margin",
                value: $profitMargin,
                range: 0...100,
                step: 5,
                format: "%.0f%%"
              )
            }
            .padding(Design.Spacing.md)
          }

          // Time
          AnimatedCard {
            VStack(alignment: .leading, spacing: Design.Spacing.md) {
              HStack {
                Image(systemName: "clock")
                  .foregroundColor(Design.Colors.warning)
                Text("Time")
                  .font(Design.Typography.labelMedium)
              }

              NumberStepper(
                label: "Work Weeks/Year",
                value: $workWeeksPerYear,
                range: 20...52,
                step: 1,
                format: "%.0f weeks"
              )

              NumberStepper(
                label: "Billable Hours/Week",
                value: $billableHoursPerWeek,
                range: 10...60,
                step: 5,
                format: "%.0f hours"
              )

              Divider()

              VStack(alignment: .leading, spacing: Design.Spacing.xxs) {
                Text("Tips")
                  .font(Design.Typography.labelSmall)
                  .foregroundColor(Design.Colors.textTertiary)

                Text("• Most freelancers bill 60-70% of their time")
                  .font(Design.Typography.bodySmall)
                  .foregroundColor(Design.Colors.textSecondary)

                Text("• Account for vacation, sick days, admin work")
                  .font(Design.Typography.bodySmall)
                  .foregroundColor(Design.Colors.textSecondary)
              }
            }
            .padding(Design.Spacing.md)
          }
        }
      }
    }
  }
}

// MARK: - Income Goals View
struct IncomeGoalsView: View {
  @EnvironmentObject var appState: AppState
  @State private var monthlyGoal: Double = 10000
  @State private var yearlyGoal: Double = 120000
  @AppStorage("incomeGoalMonthly") private var savedMonthlyGoal: Double = 10000
  @AppStorage("incomeGoalYearly") private var savedYearlyGoal: Double = 120000

  var monthlyProgress: Double {
    guard monthlyGoal > 0 else { return 0 }
    return min(appState.metrics.monthlyRevenue / monthlyGoal, 1.0)
  }

  var yearlyProgress: Double {
    guard yearlyGoal > 0 else { return 0 }
    return min(appState.metrics.totalRevenue / yearlyGoal, 1.0)
  }

  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    ScrollView {
      VStack(spacing: Design.Spacing.lg) {
        // Progress Cards - responsive layout
        LazyVGrid(
          columns: [GridItem(.adaptive(minimum: 280))],
          spacing: Design.Spacing.md
        ) {
          // Monthly Goal
          GoalProgressCard(
            title: "Monthly Goal",
            current: appState.metrics.monthlyRevenue,
            goal: savedMonthlyGoal,
            progress: monthlyProgress,
            color: Design.Colors.primary,
            icon: "calendar"
          )

          // Yearly Goal
          GoalProgressCard(
            title: "Yearly Goal",
            current: appState.metrics.totalRevenue,
            goal: savedYearlyGoal,
            progress: yearlyProgress,
            color: Design.Colors.success,
            icon: "chart.line.uptrend.xyaxis"
          )
        }

        // Goal Settings
        AnimatedCard {
          VStack(alignment: .leading, spacing: Design.Spacing.md) {
            HStack {
              Image(systemName: "target")
                .foregroundColor(Design.Colors.warning)
              Text("Set Your Goals")
                .font(Design.Typography.headingSmall)
              Spacer()
            }

            NumberStepper(
              label: "Monthly Income Goal",
              value: $monthlyGoal,
              range: 1000...100000,
              step: 500,
              format: "$%.0f"
            )

            NumberStepper(
              label: "Yearly Income Goal",
              value: $yearlyGoal,
              range: 10000...1_000_000,
              step: 5000,
              format: "$%.0f"
            )

            HStack(spacing: Design.Spacing.sm) {
              Button("Sync Monthly × 12") {
                yearlyGoal = monthlyGoal * 12
              }
              .buttonStyle(DSSecondaryButtonStyle())

              Spacer()

              Button("Save Goals") {
                savedMonthlyGoal = monthlyGoal
                savedYearlyGoal = yearlyGoal
                ToastManager.shared.show("Goals saved!", type: .success)
              }
              .buttonStyle(DSPrimaryButtonStyle())
            }
          }
          .padding(Design.Spacing.md)
        }
        .onAppear {
          monthlyGoal = savedMonthlyGoal
          yearlyGoal = savedYearlyGoal
        }

        // Projections
        AnimatedCard {
          VStack(alignment: .leading, spacing: Design.Spacing.md) {
            HStack {
              Image(systemName: "sparkles")
                .foregroundColor(Design.Colors.purple)
              Text("Projections")
                .font(Design.Typography.headingSmall)
              Spacer()
            }

            HStack(spacing: Design.Spacing.lg) {
              ProjectionItem(
                title: "At Current Pace",
                value: appState.formatCurrency(appState.metrics.monthlyRevenue * 12),
                subtitle: "Projected yearly",
                icon: "arrow.right",
                color: Design.Colors.primary
              )

              ProjectionItem(
                title: "To Hit Monthly Goal",
                value: appState.formatCurrency(
                  max(0, savedMonthlyGoal - appState.metrics.monthlyRevenue)),
                subtitle: "Still needed",
                icon: "target",
                color: Design.Colors.warning
              )

              ProjectionItem(
                title: "Hours This Month",
                value: appState.formatHours(appState.metrics.weeklyHours * 4),
                subtitle: "Logged time",
                icon: "clock.fill",
                color: Design.Colors.success
              )
            }
          }
          .padding(Design.Spacing.md)
        }
      }
    }
  }
}

struct GoalProgressCard: View {
  let title: String
  let current: Double
  let goal: Double
  let progress: Double
  let color: Color
  let icon: String
  @EnvironmentObject var appState: AppState
  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    VStack(alignment: .leading, spacing: Design.Spacing.md) {
      HStack {
        Image(systemName: icon)
          .foregroundColor(color)
        Text(title)
          .font(Design.Typography.labelMedium)
        Spacer()
        Text(String(format: "%.0f%%", progress * 100))
          .font(Design.Typography.headingSmall)
          .foregroundColor(color)
      }

      // Progress bar
      GeometryReader { geometry in
        ZStack(alignment: .leading) {
          RoundedRectangle(cornerRadius: 6)
            .fill(color.opacity(0.15))
            .frame(height: 12)

          RoundedRectangle(cornerRadius: 6)
            .fill(
              LinearGradient(
                colors: [color, color.opacity(0.8)],
                startPoint: .leading,
                endPoint: .trailing
              )
            )
            .frame(width: geometry.size.width * progress, height: 12)
        }
      }
      .frame(height: 12)

      HStack {
        VStack(alignment: .leading, spacing: Design.Spacing.xxxs) {
          Text("Current")
            .font(Design.Typography.labelSmall)
            .foregroundColor(Design.Colors.textTertiary)
          Text(appState.formatCurrency(current))
            .font(Design.Typography.headingSmall)
        }

        Spacer()

        VStack(alignment: .trailing, spacing: Design.Spacing.xxxs) {
          Text("Goal")
            .font(Design.Typography.labelSmall)
            .foregroundColor(Design.Colors.textTertiary)
          Text(appState.formatCurrency(goal))
            .font(Design.Typography.headingSmall)
            .foregroundColor(Design.Colors.textSecondary)
        }
      }
    }
    .padding(Design.Spacing.lg)
    .frame(maxWidth: .infinity)
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.md)
        .fill(Design.Colors.surfaceElevated(colorScheme))
        .shadow(
          color: Design.Shadow.sm.color, radius: Design.Shadow.sm.radius, y: Design.Shadow.sm.y)
    )
  }
}

struct ProjectionItem: View {
  let title: String
  let value: String
  let subtitle: String
  let icon: String
  let color: Color

  var body: some View {
    VStack(alignment: .leading, spacing: Design.Spacing.xs) {
      HStack(spacing: Design.Spacing.xxs) {
        Image(systemName: icon)
          .font(.system(size: 12))
          .foregroundColor(color)
        Text(title)
          .font(Design.Typography.labelSmall)
          .foregroundColor(Design.Colors.textSecondary)
      }

      Text(value)
        .font(Design.Typography.headingSmall)

      Text(subtitle)
        .font(Design.Typography.bodySmall)
        .foregroundColor(Design.Colors.textTertiary)
    }
    .frame(maxWidth: .infinity, alignment: .leading)
  }
}

// MARK: - Settings Content
struct SettingsContent: View {
  @EnvironmentObject var appState: AppState
  @State private var showCompanyEditor = false

  // Company form fields
  @State private var companyName = ""
  @State private var companyEmail = ""
  @State private var companyAddress = ""
  @State private var companyTaxId = ""
  @State private var companyBankName = ""
  @State private var companyBankAccount = ""
  @State private var companyBankSwift = ""

  var body: some View {
    ZStack {
      ScrollView {
        VStack(spacing: Design.Spacing.lg) {
          // Company Profile
          CompanyProfileSection(
            hasCompany: appState.setupStatus.hasCompany,
            onEdit: {
              Task {
                if let details = try? await appState.database.getCompany() {
                  companyName = details.name
                  companyEmail = details.email
                  companyAddress = details.address ?? ""
                  companyTaxId = details.taxId ?? ""
                  companyBankName = details.bankName ?? ""
                  companyBankAccount = details.bankAccount ?? ""
                  companyBankSwift = details.bankSwift ?? ""
                } else {
                  companyName = ""
                  companyEmail = ""
                  companyAddress = ""
                  companyTaxId = ""
                  companyBankName = ""
                  companyBankAccount = ""
                  companyBankSwift = ""
                }
                showCompanyEditor = true
              }
            }
          )

          // Settings sections from SettingsSection component
          SettingsSection()
        }
        .padding(Design.Spacing.lg)
      }

      // Company Editor Modal
      if showCompanyEditor {
        Color.black.opacity(0.3)
          .ignoresSafeArea()
          .onTapGesture { showCompanyEditor = false }

        companyEditorSheet
          .transition(.scale.combined(with: .opacity))
      }
    }
    .animation(Design.Animation.smooth, value: showCompanyEditor)
  }

  private var companyEditorSheet: some View {
    ActionMenu(
      title: appState.setupStatus.hasCompany ? "Edit Company" : "Create Company",
      isPresented: $showCompanyEditor
    ) {
      VStack(spacing: Design.Spacing.md) {
        FormField(label: "Company Name", text: $companyName, placeholder: "Your Company LLC")
        FormField(label: "Email", text: $companyEmail, placeholder: "contact@company.com")
        FormField(label: "Address", text: $companyAddress, placeholder: "123 Business St, City")
        FormField(label: "Tax ID / VAT", text: $companyTaxId, placeholder: "XX-XXXXXXX")

        Divider()

        Text("Bank Details")
          .font(Design.Typography.labelSmall)
          .foregroundColor(Design.Colors.textSecondary)
          .frame(maxWidth: .infinity, alignment: .leading)

        FormField(label: "Bank Name", text: $companyBankName, placeholder: "Bank of America")
        FormField(
          label: "Account / IBAN", text: $companyBankAccount, placeholder: "DE89370400440532013000")
        FormField(label: "SWIFT / BIC", text: $companyBankSwift, placeholder: "COBADEFFXXX")

        HStack(spacing: Design.Spacing.sm) {
          Button("Cancel") {
            showCompanyEditor = false
          }
          .buttonStyle(DSSecondaryButtonStyle())
          .keyboardShortcut(.escape, modifiers: [])

          Button(appState.setupStatus.hasCompany ? "Save" : "Create") {
            submitCompany()
          }
          .buttonStyle(DSPrimaryButtonStyle())
          .disabled(companyName.isEmpty)
          .keyboardShortcut(.return, modifiers: .command)
        }
      }
    }
  }

  private func submitCompany() {
    guard !companyName.isEmpty else { return }
    Task {
      if appState.setupStatus.hasCompany {
        // Get existing company and update it
        if var existingCompany = try? await appState.database.getCompany() {
          existingCompany.name = companyName
          existingCompany.email = companyEmail
          existingCompany.address = companyAddress.isEmpty ? nil : companyAddress
          existingCompany.taxId = companyTaxId.isEmpty ? nil : companyTaxId
          existingCompany.bankName = companyBankName.isEmpty ? nil : companyBankName
          existingCompany.bankAccount = companyBankAccount.isEmpty ? nil : companyBankAccount
          existingCompany.bankSwift = companyBankSwift.isEmpty ? nil : companyBankSwift
          try? await appState.database.updateCompany(existingCompany)
        }
      } else {
        var newCompany = Company(
          name: companyName,
          email: companyEmail.isEmpty ? "" : companyEmail
        )
        newCompany.address = companyAddress.isEmpty ? nil : companyAddress
        newCompany.taxId = companyTaxId.isEmpty ? nil : companyTaxId
        newCompany.bankName = companyBankName.isEmpty ? nil : companyBankName
        newCompany.bankAccount = companyBankAccount.isEmpty ? nil : companyBankAccount
        newCompany.bankSwift = companyBankSwift.isEmpty ? nil : companyBankSwift
        _ = try? await appState.database.createCompany(newCompany)
      }
      await appState.refreshDashboard()
      showCompanyEditor = false
    }
  }
}

// MARK: - Company Profile Section
struct CompanyProfileSection: View {
  let hasCompany: Bool
  let onEdit: () -> Void
  @State private var isHovered = false
  @Environment(\.colorScheme) var colorScheme

  var body: some View {
    VStack(alignment: .leading, spacing: Design.Spacing.md) {
      HStack {
        Image(systemName: "building.2.fill")
          .foregroundColor(Design.Colors.primary)
        Text("Company Profile")
          .font(Design.Typography.headingSmall)
        Spacer()

        if hasCompany {
          DSBadge(text: "Configured", color: Design.Colors.success)
        } else {
          DSBadge(text: "Not Set", color: Design.Colors.warning)
        }
      }

      if hasCompany {
        Text("Your company profile is used on invoices and contracts.")
          .font(Design.Typography.bodySmall)
          .foregroundColor(Design.Colors.textSecondary)
      } else {
        VStack(alignment: .leading, spacing: Design.Spacing.xs) {
          Text("Set up your company profile")
            .font(Design.Typography.labelMedium)
          Text("Add your company details to appear on invoices and contracts.")
            .font(Design.Typography.bodySmall)
            .foregroundColor(Design.Colors.textSecondary)
        }
      }

      Button(action: onEdit) {
        HStack(spacing: Design.Spacing.xs) {
          Image(systemName: hasCompany ? "pencil" : "plus")
          Text(hasCompany ? "Edit Profile" : "Create Profile")
        }
      }
      .buttonStyle(
        DSPrimaryButtonStyle(color: hasCompany ? Design.Colors.primary : Design.Colors.success))
    }
    .padding(Design.Spacing.lg)
    .frame(maxWidth: .infinity, alignment: .leading)
    .background(
      RoundedRectangle(cornerRadius: Design.Radius.md)
        .fill(Design.Colors.surfaceElevated(colorScheme))
        .overlay(
          RoundedRectangle(cornerRadius: Design.Radius.md)
            .stroke(
              hasCompany ? Color.clear : Design.Colors.warning.opacity(0.5),
              lineWidth: hasCompany ? 0 : 2)
        )
        .shadow(
          color: isHovered ? Design.Shadow.md.color : Design.Shadow.sm.color,
          radius: isHovered ? Design.Shadow.md.radius : Design.Shadow.sm.radius,
          y: isHovered ? Design.Shadow.md.y : Design.Shadow.sm.y
        )
    )
    .scaleEffect(isHovered ? 1.005 : 1.0)
    .animation(Design.Animation.smooth, value: isHovered)
    .onHover { hovering in isHovered = hovering }
  }
}

struct DatabaseDocument: FileDocument {
  static var readableContentTypes: [UTType] { [.data] }

  var data: Data

  init(data: Data) {
    self.data = data
  }

  init(configuration: ReadConfiguration) throws {
    data = configuration.file.regularFileContents ?? Data()
  }

  func fileWrapper(configuration: WriteConfiguration) throws -> FileWrapper {
    FileWrapper(regularFileWithContents: data)
  }
}
