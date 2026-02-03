import SwiftUI

// MARK: - MealsView

struct MealsView: View {
    @State private var viewModel = MealsViewModel()
    @State private var selectedMealTypeForInput: MealType?

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
                                    onTapped: {
                                        selectedMealTypeForInput = mealType
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
                    initialMealType: mealType,
                    existingMeals: viewModel.dailyMeals?.meals.meals(for: mealType) ?? []
                ) {
                    Task {
                        await viewModel.loadMeals()
                    }
                }
            }
        }
        .task {
            await viewModel.loadMeals()
        }
    }
}

// MARK: - DateNavigationBar

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

// MARK: - MealTypeSection

private struct MealTypeSection: View {
    let mealType: MealType
    let meals: [HistoryDetail]
    let onTapped: () -> Void

    var body: some View {
        Button(action: onTapped) {
            VStack(alignment: .leading, spacing: 8) {
                HStack {
                    Image(systemName: mealType.icon)
                        .foregroundStyle(Theme.primary)
                    Text(mealType.displayName)
                        .font(.headline)
                        .foregroundStyle(.primary)
                    Spacer()
                    if !meals.isEmpty {
                        Text("\(Int(meals.reduce(0) { $0 + $1.totalCalories })) kcal")
                            .font(.subheadline)
                            .foregroundStyle(Theme.primary)
                    }
                    Image(systemName: "square.and.pencil")
                        .font(.title3)
                        .foregroundStyle(Theme.primary)
                }

                if meals.isEmpty {
                    Text("記録なし")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .padding(.vertical, 8)
                } else {
                    let allFoods = meals.flatMap(\.foods)
                    VStack(alignment: .leading, spacing: 4) {
                        ForEach(Array(allFoods.enumerated()), id: \.offset) { _, food in
                            HStack {
                                Text(food.name)
                                    .font(.subheadline)
                                    .foregroundStyle(.primary)
                                Spacer()
                                Text("\(Int(food.caloriesKcal)) kcal")
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                            }
                            .padding(.horizontal, 12)
                            .padding(.vertical, 8)
                            .background(Color(.secondarySystemBackground))
                            .clipShape(RoundedRectangle(cornerRadius: 8))
                        }
                    }
                }
            }
            .padding()
            .background(Color(.systemBackground))
            .clipShape(RoundedRectangle(cornerRadius: 12))
        }
        .buttonStyle(.plain)
    }
}

// MARK: - ErrorView

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
