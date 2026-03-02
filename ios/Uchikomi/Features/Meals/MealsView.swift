import os
import SwiftUI

private let logger = Logger(
    subsystem: Bundle.main.bundleIdentifier ?? "Uchikomi",
    category: "MealsView"
)

// MARK: - MealsView

struct MealsView: View {
    @State private var viewModel = MealsViewModel()
    @State private var selectedMealTypeForInput: MealType?
    @State private var isExerciseInputPresented = false

    private let dateFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd"
        formatter.timeZone = TimeZone.current
        return formatter
    }()

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
                                carbohydrates: dailyMeals.dailyTotal.totalCarbohydrates,
                                micronutrients: dailyMeals.dailyTotal.totalMicronutrients,
                                burnedCalories: dailyMeals.totalBurnedCaloriesKcal,
                                goal: viewModel.nutritionGoal
                            )

                            ForEach(MealType.allCases, id: \.self) { mealType in
                                MealTypeSection(
                                    mealType: mealType,
                                    meals: dailyMeals.meals.meals(for: mealType),
                                    pendingAnalyses: dailyMeals.pendingAnalyses(for: mealType),
                                    isSkipped: dailyMeals.meals.isSkipped(for: mealType),
                                    loadingPendingEntryId: viewModel.loadingPendingEntryId,
                                    onTapped: {
                                        selectedMealTypeForInput = mealType
                                    },
                                    onPendingTapped: { entry in
                                        Task {
                                            await viewModel.openPendingEditor(entry: entry)
                                        }
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
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    HStack(spacing: 16) {
                        Button {
                            isExerciseInputPresented = true
                        } label: {
                            Image(systemName: "flame")
                        }

                        NavigationLink {
                            SuggestionRequestView()
                        } label: {
                            Image(systemName: "sparkles")
                        }
                    }
                }
            }
            .sheet(item: $selectedMealTypeForInput) { mealType in
                MealInputView(
                    mealDate: viewModel.selectedDate,
                    initialMealType: mealType,
                    existingMeals: viewModel.dailyMeals?.meals.meals(for: mealType) ?? []
                ) {
                    Task {
                        await viewModel.loadMeals()
                        await cancelNotificationIfToday(for: mealType)
                    }
                }
            }
            .sheet(item: $viewModel.pendingEditorEntry) { entry in
                NutritionEditorView(
                    historyId: entry.id,
                    foods: viewModel.pendingEditorFoods
                ) {
                    Task {
                        await viewModel.loadMeals()
                    }
                }
            }
            .overlay {
                if viewModel.loadingPendingEntryId != nil {
                    Color.black.opacity(0.3)
                        .ignoresSafeArea()
                        .overlay {
                            ProgressView("読み込み中...")
                                .padding()
                                .background(Color(.systemBackground))
                                .clipShape(RoundedRectangle(cornerRadius: 12))
                        }
                }
            }
            .sheet(isPresented: $isExerciseInputPresented) {
                ExerciseInputView(recordedDate: dateFormatter.string(from: viewModel.selectedDate)) {
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

    /// 当日の食事記録に対して、該当する通知をキャンセルし翌日に再スケジュールする。
    /// リマインダー対象外の食事タイプ、当日以外の日付、または通知がグローバル無効の場合は何もしない。
    private func cancelNotificationIfToday(for mealType: MealType) async {
        guard Calendar.current.isDateInToday(viewModel.selectedDate),
              MealType.reminderTargets.contains(mealType) else { return }

        let store = NotificationSettingsStore()
        let settings = store.load()
        guard settings.isGlobalEnabled else { return }

        let manager = NotificationManager()
        await manager.handleMealRecorded(mealType: mealType, settings: settings)

        if let error = manager.lastSchedulingError {
            logger.error("食事記録後の通知再スケジュールに失敗: \(mealType.rawValue): \(error.localizedDescription)")
        } else {
            logger.info("食事記録後に通知をキャンセルし翌日に再スケジュール: \(mealType.rawValue)")
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
    let pendingAnalyses: [PendingAnalysisEntry]
    let isSkipped: Bool
    let loadingPendingEntryId: String?
    let onTapped: () -> Void
    let onPendingTapped: (PendingAnalysisEntry) -> Void

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
            } else {
                if meals.isEmpty, pendingAnalyses.isEmpty {
                    Text("記録なし")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .padding(.vertical, 4)
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

                        ForEach(pendingAnalyses) { entry in
                            PendingAnalysisRow(
                                entry: entry,
                                isLoading: loadingPendingEntryId == entry.id,
                                onTapped: { onPendingTapped(entry) }
                            )
                        }
                    }
                }
            }
        }
        .padding()
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

// MARK: - PendingAnalysisRow

private struct PendingAnalysisRow: View {
    let entry: PendingAnalysisEntry
    let isLoading: Bool
    let onTapped: () -> Void

    var body: some View {
        if entry.isAnalyzing {
            HStack(spacing: 8) {
                ProgressView()
                    .progressViewStyle(.circular)
                    .scaleEffect(0.8)
                Text("分析中...")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                Spacer()
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 10)
            .background(Color(.secondarySystemBackground))
            .clipShape(RoundedRectangle(cornerRadius: 8))
        } else if entry.isReadyToConfirm {
            Button(action: onTapped) {
                HStack {
                    Image(systemName: "checkmark.circle")
                        .foregroundStyle(.green)
                    Text("確認待ち - タップして確認")
                        .font(.subheadline)
                        .foregroundStyle(.primary)
                    Spacer()
                    if isLoading {
                        ProgressView()
                            .scaleEffect(0.7)
                    } else {
                        Image(systemName: "chevron.right")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
                .padding(.horizontal, 12)
                .padding(.vertical, 10)
                .background(Color.green.opacity(0.1))
                .clipShape(RoundedRectangle(cornerRadius: 8))
            }
            .buttonStyle(.plain)
            .disabled(isLoading)
        } else if entry.isFailed {
            HStack {
                Image(systemName: "exclamationmark.circle")
                    .foregroundStyle(.red)
                VStack(alignment: .leading, spacing: 2) {
                    Text("分析失敗")
                        .font(.subheadline)
                        .foregroundStyle(.primary)
                    if let msg = entry.errorMessage, !msg.isEmpty {
                        Text(msg)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
                Spacer()
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 10)
            .background(Color.red.opacity(0.08))
            .clipShape(RoundedRectangle(cornerRadius: 8))
        }
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
