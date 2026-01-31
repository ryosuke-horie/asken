import SwiftUI

struct MealsView: View {
    @State private var viewModel = MealsViewModel()
    @State private var selectedMealTypeForInput: MealType?
    @State private var editingMeal: HistoryDetail?
    @State private var deletingMeal: HistoryDetail?

    var body: some View {
        NavigationStack {
            VStack(spacing: 0) {
                // Date Navigation
                DateNavigationBar(
                    formattedDate: viewModel.formattedDate,
                    isToday: viewModel.isToday,
                    onPrevious: viewModel.goToPreviousDay,
                    onNext: viewModel.goToNextDay,
                    onToday: viewModel.goToToday
                )

                if viewModel.isLoading {
                    Spacer()
                    ProgressView()
                    Spacer()
                } else if let error = viewModel.errorMessage {
                    Spacer()
                    ErrorView(message: error) {
                        Task {
                            await viewModel.loadMeals()
                        }
                    }
                    Spacer()
                } else if let dailyMeals = viewModel.dailyMeals {
                    ScrollView {
                        VStack(spacing: 16) {
                            // Daily Total
                            NutritionSummaryCard(
                                calories: dailyMeals.dailyTotal.totalCalories,
                                protein: dailyMeals.dailyTotal.totalProtein,
                                fat: dailyMeals.dailyTotal.totalFat,
                                carbohydrates: dailyMeals.dailyTotal.totalCarbohydrates
                            )

                            // Meals by Type
                            ForEach(MealType.allCases, id: \.self) { mealType in
                                MealTypeSection(
                                    mealType: mealType,
                                    meals: dailyMeals.meals.meals(for: mealType),
                                    onAddTapped: {
                                        selectedMealTypeForInput = mealType
                                    },
                                    onEditTapped: { meal in
                                        editingMeal = meal
                                    },
                                    onDeleteTapped: { meal in
                                        deletingMeal = meal
                                    }
                                )
                            }
                        }
                        .padding()
                    }
                } else {
                    Spacer()
                    Text("データがありません")
                        .foregroundStyle(.secondary)
                    Spacer()
                }
            }
            .navigationTitle("食事記録")
            .sheet(item: $selectedMealTypeForInput) { mealType in
                MealInputView(
                    mealDate: viewModel.selectedDate,
                    initialMealType: mealType
                ) {
                    Task {
                        await viewModel.loadMeals()
                    }
                }
            }
            .sheet(item: $editingMeal) { meal in
                NutritionEditorView(
                    historyId: meal.id,
                    foods: meal.foods
                ) {
                    Task {
                        await viewModel.loadMeals()
                    }
                }
            }
            .alert("削除の確認", isPresented: Binding(
                get: { deletingMeal != nil },
                set: { if !$0 { deletingMeal = nil } }
            )) {
                Button("キャンセル", role: .cancel) {
                    deletingMeal = nil
                }
                Button("削除", role: .destructive) {
                    if let meal = deletingMeal {
                        Task {
                            await viewModel.deleteHistory(id: meal.id)
                        }
                    }
                    deletingMeal = nil
                }
            } message: {
                Text("この食事記録を削除しますか？")
            }
        }
        .task {
            await viewModel.loadMeals()
        }
    }
}

// MARK: - Subviews

private struct DateNavigationBar: View {
    let formattedDate: String
    let isToday: Bool
    let onPrevious: () -> Void
    let onNext: () -> Void
    let onToday: () -> Void

    var body: some View {
        HStack {
            Button(action: onPrevious) {
                Image(systemName: "chevron.left")
                    .font(.title3)
            }

            Spacer()

            VStack(spacing: 2) {
                Text(formattedDate)
                    .font(.headline)

                if !isToday {
                    Button("今日に戻る", action: onToday)
                        .font(.caption)
                }
            }

            Spacer()

            Button(action: onNext) {
                Image(systemName: "chevron.right")
                    .font(.title3)
            }
            .disabled(isToday)
            .opacity(isToday ? 0.3 : 1)
        }
        .padding()
        .background(Color(.systemGroupedBackground))
    }
}

private struct MealTypeSection: View {
    let mealType: MealType
    let meals: [HistoryDetail]
    let onAddTapped: () -> Void
    let onEditTapped: (HistoryDetail) -> Void
    let onDeleteTapped: (HistoryDetail) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: mealType.icon)
                    .foregroundStyle(Theme.primary)
                Text(mealType.displayName)
                    .font(.headline)
                Spacer()
                if !meals.isEmpty {
                    Text("\(Int(meals.reduce(0) { $0 + $1.totalCalories })) kcal")
                        .font(.subheadline)
                        .foregroundStyle(Theme.primary)
                }
                Button(action: onAddTapped) {
                    Image(systemName: "square.and.pencil")
                        .font(.title3)
                        .foregroundStyle(Theme.primary)
                }
            }

            if meals.isEmpty {
                Text("記録なし")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .padding(.vertical, 8)
            } else {
                ForEach(meals) { meal in
                    MealCard(
                        meal: meal,
                        onEdit: { onEditTapped(meal) },
                        onDelete: { onDeleteTapped(meal) }
                    )
                }
            }
        }
        .padding()
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

private struct MealCard: View {
    let meal: HistoryDetail
    let onEdit: () -> Void
    let onDelete: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            ForEach(meal.foods) { food in
                HStack {
                    Text(food.name)
                        .font(.subheadline)
                    Spacer()
                    Text("\(Int(food.caloriesKcal)) kcal")
                        .font(.caption)
                        .foregroundStyle(Theme.primary)
                }
            }

            HStack {
                Spacer()
                Button(action: onEdit) {
                    Label("編集", systemImage: "pencil")
                        .font(.caption)
                }
                .buttonStyle(.bordered)
                .tint(Theme.primary)

                Button(action: onDelete) {
                    Label("削除", systemImage: "trash")
                        .font(.caption)
                }
                .buttonStyle(.bordered)
                .tint(.red)
            }
        }
        .padding()
        .background(Color(.secondarySystemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 8))
    }
}

private struct ErrorView: View {
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

#Preview {
    MealsView()
}
