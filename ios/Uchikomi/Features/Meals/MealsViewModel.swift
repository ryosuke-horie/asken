import Foundation
import os

private let logger = Logger(subsystem: Bundle.main.bundleIdentifier ?? "Uchikomi", category: "MealsViewModel")

@Observable
final class MealsViewModel {
    var selectedDate = Date()
    var dailyMeals: DailyMeals?
    var nutritionGoal: NutritionGoal?
    var isLoading = false
    var errorMessage: String?

    private let repository: MealRepositoryProtocol
    private let nutritionGoalRepository: NutritionGoalRepositoryProtocol
    private let weightRepository: WeightRepositoryProtocol

    init(
        repository: MealRepositoryProtocol = MealRepository(),
        nutritionGoalRepository: NutritionGoalRepositoryProtocol = NutritionGoalRepository(),
        weightRepository: WeightRepositoryProtocol = WeightRepository()
    ) {
        self.repository = repository
        self.nutritionGoalRepository = nutritionGoalRepository
        self.weightRepository = weightRepository
    }

    var formattedDate: String {
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "ja_JP")
        formatter.dateFormat = "M月d日(E)"
        return formatter.string(from: selectedDate)
    }

    var isToday: Bool {
        Calendar.current.isDateInToday(selectedDate)
    }

    func loadMeals() async {
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        let calendar = Calendar.current
        let endDate = selectedDate
        let startDate = calendar.date(byAdding: .day, value: -7, to: endDate) ?? endDate

        // 食事データと体重データを並列取得
        // 注: async letはStructured Concurrencyにより、早期リターン時も自動でキャンセル・awaitされる
        async let mealsResult = repository.getDailyMeals(date: selectedDate)
        async let weightResult = weightRepository.getRecords(from: startDate, to: endDate)

        // 食事データを取得（必須）
        do {
            dailyMeals = try await mealsResult
        } catch let error as APIError {
            logger.error("食事データ取得でAPIエラー: \(error.localizedDescription)")
            errorMessage = error.localizedDescription
            return
        } catch {
            logger.error("食事データ取得で予期しないエラー: \(error.localizedDescription)")
            errorMessage = "食事データの取得に失敗しました"
            return
        }

        // 体重データを処理（失敗しても食事データは表示）
        let currentWeight: Double?
        let goalWeight: Double?

        do {
            let weightResponse = try await weightResult
            let dateFormatter = DateFormatter()
            dateFormatter.dateFormat = "yyyy-MM-dd"
            let todayString = dateFormatter.string(from: selectedDate)

            if let todaySummary = weightResponse.dailySummary[todayString] {
                currentWeight = todaySummary.latestWeight
            } else {
                currentWeight = weightResponse.records.first?.weightKg
            }

            goalWeight = weightResponse.goal?.targetWeightKg
        } catch {
            // 体重取得に失敗した場合はnilを使用（フェーズは維持期になる）
            logger.warning("体重データ取得に失敗（維持期にフォールバック）: \(error.localizedDescription)")
            currentWeight = nil
            goalWeight = nil
        }

        do {
            nutritionGoal = try await nutritionGoalRepository.getGoal(
                currentWeight: currentWeight,
                goalWeight: goalWeight
            )
        } catch {
            // 栄養目標取得に失敗した場合はnilを使用（目標なし表示になる）
            logger.warning("栄養目標取得に失敗（目標なし表示にフォールバック）: \(error.localizedDescription)")
            nutritionGoal = nil
        }
    }

    func goToPreviousDay() {
        selectedDate = Calendar.current.date(byAdding: .day, value: -1, to: selectedDate) ?? selectedDate
        Task {
            await loadMeals()
        }
    }

    func goToNextDay() {
        guard !isToday else { return }
        selectedDate = Calendar.current.date(byAdding: .day, value: 1, to: selectedDate) ?? selectedDate
        Task {
            await loadMeals()
        }
    }

    func goToToday() {
        selectedDate = Date()
        Task {
            await loadMeals()
        }
    }
}
