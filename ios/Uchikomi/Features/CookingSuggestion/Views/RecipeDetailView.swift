import SwiftUI

// MARK: - RecipeDetailView

struct RecipeDetailView: View {
    @State private var viewModel: RecipeDetailViewModel
    @Environment(\.dismiss) private var dismiss
    @State private var showingAcceptConfirmation = false
    @State private var showingAcceptResult = false
    @State private var showingAcceptError = false

    init(suggestion: MenuSuggestion) {
        _viewModel = State(initialValue: RecipeDetailViewModel(suggestion: suggestion))
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 20) {
                headerSection
                nutritionSection
                ingredientsSection
                recipeSection
                acceptButton
            }
            .padding()
        }
        .navigationTitle("メニュー詳細")
        .navigationBarTitleDisplayMode(.inline)
        .task {
            await viewModel.loadRecipe()
        }
        .alert("このメニューで記録しますか？", isPresented: $showingAcceptConfirmation) {
            Button("記録する") {
                Task {
                    let success = await viewModel.accept()
                    if success {
                        showingAcceptResult = true
                    } else {
                        showingAcceptError = true
                    }
                }
            }
            Button("キャンセル", role: .cancel) {}
        } message: {
            Text("食事記録が作成され、使用した食材の在庫が自動的に控除されます。")
        }
        .alert("記録しました", isPresented: $showingAcceptResult) {
            Button("OK") {
                dismiss()
            }
        } message: {
            if let result = viewModel.acceptResult {
                Text(acceptResultMessage(result))
            }
        }
        .alert("記録に失敗しました", isPresented: $showingAcceptError) {
            Button("OK") {}
        } message: {
            if let error = viewModel.acceptErrorMessage {
                Text(error)
            }
        }
    }

    // MARK: - Sections

    private var headerSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Label(viewModel.suggestion.mealType.displayName, systemImage: viewModel.suggestion.mealType.icon)
                .font(.subheadline)
                .foregroundStyle(Theme.primary)

            Text(viewModel.suggestion.title)
                .font(.title2)
                .fontWeight(.bold)

            Text(viewModel.suggestion.reason)
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
    }

    private var nutritionSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("推定栄養素")
                .font(.headline)

            HStack(spacing: 0) {
                NutritionItem(
                    label: "カロリー",
                    value: "\(Int(viewModel.suggestion.estimatedNutrition.calories))",
                    unit: "kcal"
                )
                Spacer()
                NutritionItem(
                    label: "タンパク質",
                    value: String(format: "%.1f", viewModel.suggestion.estimatedNutrition.protein),
                    unit: "g"
                )
                Spacer()
                NutritionItem(
                    label: "脂質",
                    value: String(format: "%.1f", viewModel.suggestion.estimatedNutrition.fat),
                    unit: "g"
                )
                Spacer()
                NutritionItem(
                    label: "炭水化物",
                    value: String(format: "%.1f", viewModel.suggestion.estimatedNutrition.carbohydrates),
                    unit: "g"
                )
            }
            .padding()
            .background(Color(.systemGray6))
            .clipShape(RoundedRectangle(cornerRadius: 12))
        }
    }

    private var ingredientsSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("使用食材")
                .font(.headline)

            ForEach(viewModel.suggestion.ingredientsUsed) { ingredient in
                HStack {
                    Text(ingredient.name)
                        .font(.body)
                    Spacer()
                    Text("\(formattedQuantity(ingredient.quantity)) \(ingredient.unit)")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }
                .padding(.vertical, 4)
            }
        }
    }

    private var recipeSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("調理手順")
                .font(.headline)

            if let recipe = viewModel.suggestion.recipe {
                Text(recipe)
                    .font(.body)
                    .lineSpacing(4)
            } else if viewModel.isLoadingRecipe {
                HStack {
                    ProgressView()
                    Text("レシピを生成中...")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }
                .padding(.vertical, 8)
            } else if let error = viewModel.recipeErrorMessage {
                VStack(spacing: 8) {
                    Text(error)
                        .font(.subheadline)
                        .foregroundStyle(.red)
                    Button("再試行") {
                        Task {
                            await viewModel.loadRecipe()
                        }
                    }
                    .buttonStyle(.bordered)
                }
            }
        }
    }

    private var acceptButton: some View {
        Button {
            showingAcceptConfirmation = true
        } label: {
            HStack {
                Spacer()
                if viewModel.isAccepting {
                    ProgressView()
                        .padding(.trailing, 8)
                    Text("記録中...")
                } else {
                    Image(systemName: "checkmark.circle")
                    Text("このメニューで記録する")
                }
                Spacer()
            }
            .padding(.vertical, 4)
        }
        .buttonStyle(.borderedProminent)
        .tint(Theme.primary)
        .disabled(viewModel.isAccepting || viewModel.suggestion.status != .suggested)
        .padding(.top, 8)
    }

    // MARK: - Helpers

    private func formattedQuantity(_ quantity: Double) -> String {
        if quantity.truncatingRemainder(dividingBy: 1) == 0 {
            return String(Int(quantity))
        }
        return String(format: "%.1f", quantity)
    }

    private func acceptResultMessage(_ result: AcceptMenuSuggestionResult) -> String {
        if result.deductedIngredients.isEmpty {
            return "食事記録が作成されました。"
        }
        let deductedNames = result.deductedIngredients.map(\.name).joined(separator: "、")
        return "食事記録が作成されました。\n控除された食材: \(deductedNames)"
    }
}

// MARK: - NutritionItem

private struct NutritionItem: View {
    let label: String
    let value: String
    let unit: String

    var body: some View {
        VStack(spacing: 4) {
            Text(value)
                .font(.title3)
                .fontWeight(.semibold)
            Text(unit)
                .font(.caption2)
                .foregroundStyle(.secondary)
            Text(label)
                .font(.caption)
                .foregroundStyle(.secondary)
        }
    }
}
