import SwiftUI

// MARK: - PantryListView

struct PantryListView: View {
    @State private var viewModel = PantryViewModel()
    @State private var showingAddSheet = false
    @State private var showingReceiptScan = false
    @State private var editingIngredient: Ingredient?

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isLoading {
                    ProgressView()
                        .frame(maxWidth: .infinity, maxHeight: .infinity)
                } else if let error = viewModel.errorMessage {
                    PantryErrorView(message: error) {
                        Task { await viewModel.loadIngredients() }
                    }
                } else if viewModel.ingredients.isEmpty {
                    emptyStateView
                } else {
                    ingredientList
                }
            }
            .navigationTitle("食材")
            .navigationBarTitleDisplayMode(.large)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Menu {
                        Button {
                            showingAddSheet = true
                        } label: {
                            Label("手動で追加", systemImage: "pencil")
                        }
                        Button {
                            showingReceiptScan = true
                        } label: {
                            Label("レシートを撮影", systemImage: "camera")
                        }
                    } label: {
                        Image(systemName: "plus")
                    }
                }
            }
            .sheet(isPresented: $showingAddSheet) {
                IngredientEditView(ingredient: nil) { saved in
                    viewModel.addIngredient(saved)
                }
            }
            .sheet(isPresented: $showingReceiptScan) {
                ReceiptScanView { savedIngredients in
                    savedIngredients.forEach { viewModel.addIngredient($0) }
                }
            }
            .sheet(item: $editingIngredient) { ingredient in
                IngredientEditView(ingredient: ingredient) { updated in
                    viewModel.updateIngredient(updated)
                }
            }
            .task {
                await viewModel.loadIngredients()
            }
        }
    }

    // MARK: - Subviews

    private var emptyStateView: some View {
        ContentUnavailableView {
            Label("食材がありません", systemImage: "cart")
        } description: {
            Text("レシートを撮影するか、手動で食材を追加してください")
        } actions: {
            Button("レシートを撮影") {
                showingReceiptScan = true
            }
            .buttonStyle(.borderedProminent)
            Button("手動で追加") {
                showingAddSheet = true
            }
        }
    }

    private var ingredientList: some View {
        List {
            ForEach(viewModel.groupedIngredients, id: \.category) { group in
                Section(group.category.displayName) {
                    ForEach(group.items) { ingredient in
                        IngredientRow(ingredient: ingredient)
                            .contentShape(Rectangle())
                            .onTapGesture {
                                editingIngredient = ingredient
                            }
                    }
                    .onDelete { offsets in
                        deleteIngredients(in: group.items, at: offsets)
                    }
                }
            }
        }
        .listStyle(.insetGrouped)
    }

    private func deleteIngredients(in items: [Ingredient], at offsets: IndexSet) {
        let idsToDelete = offsets.map { items[$0].id }
        Task {
            for id in idsToDelete {
                await viewModel.deleteIngredient(id: id)
            }
        }
    }
}

// MARK: - IngredientRow

private struct IngredientRow: View {
    let ingredient: Ingredient

    var body: some View {
        HStack {
            VStack(alignment: .leading, spacing: 4) {
                Text(ingredient.name)
                    .font(.body)
                if let expiryDate = ingredient.expiryDate {
                    ExpiryLabel(
                        expiryDate: expiryDate,
                        isExpired: ingredient.isExpired,
                        isExpiring: ingredient.isExpiringWithinThreeDays
                    )
                }
            }

            Spacer()

            Text("\(quantityText(ingredient.quantity)) \(ingredient.unit)")
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
        .padding(.vertical, 2)
    }

    private func quantityText(_ quantity: Double) -> String {
        if quantity.truncatingRemainder(dividingBy: 1) == 0 {
            return String(Int(quantity))
        }
        return String(format: "%.1f", quantity)
    }
}

// MARK: - ExpiryLabel

private struct ExpiryLabel: View {
    let expiryDate: Date
    let isExpired: Bool
    let isExpiring: Bool

    var body: some View {
        HStack(spacing: 4) {
            Image(systemName: isExpired ? "exclamationmark.circle.fill" : "clock")
                .font(.caption)
            Text(expiryText)
                .font(.caption)
        }
        .foregroundStyle(labelColor)
    }

    private var expiryText: String {
        if isExpired {
            return "期限切れ"
        }
        let formatter = DateFormatter()
        formatter.dateStyle = .short
        formatter.timeStyle = .none
        return "期限: \(formatter.string(from: expiryDate))"
    }

    private var labelColor: Color {
        if isExpired { return .red }
        if isExpiring { return .orange }
        return .secondary
    }
}

// MARK: - PantryErrorView

private struct PantryErrorView: View {
    let message: String
    let onRetry: () -> Void

    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "exclamationmark.triangle")
                .font(.largeTitle)
                .foregroundStyle(.orange)

            Text(message)
                .multilineTextAlignment(.center)

            Button("再試行", action: onRetry)
                .buttonStyle(.bordered)
        }
        .padding()
    }
}

// MARK: - Preview

#Preview {
    PantryListView()
}
