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

    init(
        repository: MealRepositoryProtocol = MealRepository(),
        nutritionGoalRepository: NutritionGoalRepositoryProtocol = NutritionGoalRepository()
    ) {
        self.repository = repository
        self.nutritionGoalRepository = nutritionGoalRepository
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

        do {
            dailyMeals = try await repository.getDailyMeals(date: selectedDate)
            // 栄養目標も取得
            nutritionGoal = try await nutritionGoalRepository.getGoal()
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "食事データの取得に失敗しました"
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
