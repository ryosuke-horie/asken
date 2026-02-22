import SwiftUI

// MARK: - SuggestionListView

struct SuggestionListView: View {
    @Bindable var viewModel: CookingSuggestionViewModel

    var body: some View {
        Group {
            if viewModel.suggestions.isEmpty {
                emptyStateView
            } else {
                suggestionScrollView
            }
        }
        .navigationTitle("サジェスト一覧")
        .navigationDestination(for: MenuSuggestion.self) { suggestion in
            RecipeDetailView(suggestion: suggestion)
        }
    }

    // MARK: - Subviews

    private var emptyStateView: some View {
        ContentUnavailableView {
            Label("サジェストがありません", systemImage: "lightbulb")
        } description: {
            Text("メニューサジェストを生成してください")
        }
    }

    private var suggestionScrollView: some View {
        ScrollView {
            LazyVStack(spacing: 12) {
                ForEach(viewModel.suggestions) { suggestion in
                    NavigationLink(value: suggestion) {
                        SuggestionCard(suggestion: suggestion)
                    }
                    .buttonStyle(.plain)
                    .contextMenu {
                        Button(role: .destructive) {
                            Task {
                                await viewModel.dismissSuggestion(id: suggestion.id)
                            }
                        } label: {
                            Label("却下", systemImage: "xmark")
                        }
                    }
                }
            }
            .padding()
        }
    }
}

// MARK: - SuggestionCard

struct SuggestionCard: View {
    let suggestion: MenuSuggestion

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Label(suggestion.mealType.displayName, systemImage: suggestion.mealType.icon)
                    .font(.caption)
                    .foregroundStyle(.secondary)
                Spacer()
                Image(systemName: "chevron.right")
                    .font(.caption)
                    .foregroundStyle(.tertiary)
            }

            Text(suggestion.title)
                .font(.headline)
                .lineLimit(2)
                .multilineTextAlignment(.leading)

            Text(suggestion.description)
                .font(.subheadline)
                .foregroundStyle(.secondary)
                .lineLimit(2)
                .multilineTextAlignment(.leading)

            HStack(spacing: 16) {
                NutritionBadge(label: "kcal", value: suggestion.estimatedNutrition.calories)
                NutritionBadge(label: "P", value: suggestion.estimatedNutrition.protein)
                NutritionBadge(label: "F", value: suggestion.estimatedNutrition.fat)
                NutritionBadge(label: "C", value: suggestion.estimatedNutrition.carbohydrates)
            }

            HStack(spacing: 4) {
                Image(systemName: "leaf")
                    .font(.caption)
                Text("食材 \(suggestion.ingredientsUsed.count)品目使用")
                    .font(.caption)
            }
            .foregroundStyle(.secondary)
        }
        .padding()
        .background(.background)
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .shadow(color: .black.opacity(0.08), radius: 4, y: 2)
    }
}

// MARK: - NutritionBadge

private struct NutritionBadge: View {
    let label: String
    let value: Double

    var body: some View {
        VStack(spacing: 2) {
            Text(formattedValue)
                .font(.subheadline)
                .fontWeight(.medium)
            Text(label)
                .font(.caption2)
                .foregroundStyle(.secondary)
        }
    }

    private var formattedValue: String {
        if value >= 100 {
            return String(Int(value))
        }
        return String(format: "%.1f", value)
    }
}
