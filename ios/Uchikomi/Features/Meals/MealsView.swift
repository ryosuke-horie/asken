import SwiftUI

// MARK: - MealsView

struct MealsView: View {
    @State private var viewModel = MealsViewModel()
    @State private var selectedMealTypeForInput: MealType?
    @State private var skippingMealType: MealType?

    var body: some View {
        NavigationStack {
            VStack(spacing: 0) {
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
                            NutritionSummaryCard(
                                calories: dailyMeals.dailyTotal.totalCalories,
                                protein: dailyMeals.dailyTotal.totalProtein,
                                fat: dailyMeals.dailyTotal.totalFat,
                                carbohydrates: dailyMeals.dailyTotal.totalCarbohydrates
                            )

                            ForEach(MealType.allCases, id: \.self) { mealType in
                                MealTypeSection(
                                    mealType: mealType,
                                    meals: dailyMeals.meals.meals(for: mealType),
                                    isSkipped: dailyMeals.meals.isSkipped(for: mealType),
                                    isSkipping: viewModel.isSkipping,
                                    onTapped: {
                                        selectedMealTypeForInput = mealType
                                    },
                                    onSkipped: {
                                        skippingMealType = mealType
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
            .alert("確認", isPresented: Binding(
                get: { skippingMealType != nil },
                set: { if !$0 { skippingMealType = nil } }
            )) {
                Button("キャンセル", role: .cancel) {}
                Button("食べなかった") {
                    if let mealType = skippingMealType {
                        Task {
                            await viewModel.skipMeal(mealType: mealType)
                        }
                    }
                    skippingMealType = nil
                }
            } message: {
                if let mealType = skippingMealType {
                    Text("\(mealType.displayName)を「食べなかった」として記録しますか？")
                }
            }
            .alert("エラー", isPresented: Binding(
                get: { viewModel.actionError != nil },
                set: { if !$0 { viewModel.actionError = nil } }
            )) {
                Button("OK") {}
            } message: {
                if let error = viewModel.actionError {
                    Text(error)
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
    let isSkipped: Bool
    let isSkipping: Bool
    let onTapped: () -> Void
    let onSkipped: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Button(action: onTapped) {
                HStack {
                    Image(systemName: isSkipped ? "moon.zzz" : mealType.icon)
                        .foregroundStyle(isSkipped ? .secondary : Theme.primary)
                    Text(mealType.displayName)
                        .font(.headline)
                        .foregroundStyle(.primary)
                    Spacer()
                    if !isSkipped, !meals.isEmpty {
                        Text("\(Int(meals.reduce(0) { $0 + $1.totalCalories })) kcal")
                            .font(.subheadline)
                            .foregroundStyle(Theme.primary)
                    }
                    Image(systemName: "square.and.pencil")
                        .font(.title3)
                        .foregroundStyle(Theme.primary)
                }
            }
            .buttonStyle(.plain)

            if isSkipped {
                Text("食べませんでした")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .padding(.vertical, 8)
            } else if meals.isEmpty {
                Text("記録なし")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .padding(.vertical, 4)

                Button(action: onSkipped) {
                    Label("食べなかった", systemImage: "moon.zzz")
                        .font(.subheadline)
                }
                .buttonStyle(.bordered)
                .tint(.secondary)
                .disabled(isSkipping)
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
