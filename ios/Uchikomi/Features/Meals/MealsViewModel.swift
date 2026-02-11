import Foundation

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

        // 食事データを取得
        do {
            dailyMeals = try await repository.getDailyMeals(date: selectedDate)
        } catch let error as APIError {
            errorMessage = error.localizedDescription
            isLoading = false
            return
        } catch {
            errorMessage = "食事データの取得に失敗しました"
            isLoading = false
            return
        }

        // 体重データと栄養目標を取得（失敗しても食事データは表示）
        let currentWeight: Double?
        let goalWeight: Double?

        let calendar = Calendar.current
        let endDate = selectedDate
        let startDate = calendar.date(byAdding: .day, value: -7, to: endDate) ?? endDate

        do {
            let weightResponse = try await weightRepository.getRecords(
                from: startDate,
                to: endDate
            )
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
            // ログは出力できるが、ユーザーへの表示は不要
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
            nutritionGoal = nil
        }

        isLoading = false
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
