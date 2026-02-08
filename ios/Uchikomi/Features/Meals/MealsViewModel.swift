import Foundation

@Observable
final class MealsViewModel {
    var selectedDate = Date()
    var dailyMeals: DailyMeals?
    var isLoading = false
    var errorMessage: String?
    var isDeleting = false
    var isSkipping = false
    var actionError: String?

    private let repository: MealRepositoryProtocol

    init(repository: MealRepositoryProtocol = MealRepository()) {
        self.repository = repository
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

    func skipMeal(mealType: MealType) async {
        guard !isSkipping else { return }
        isSkipping = true
        actionError = nil

        do {
            try await repository.skipMeal(mealType: mealType, mealDate: selectedDate)
            await loadMeals()
        } catch let error as APIError {
            actionError = error.localizedDescription
        } catch {
            actionError = "スキップに失敗しました"
        }

        isSkipping = false
    }

    func deleteHistory(id: String) async {
        isDeleting = true
        errorMessage = nil

        do {
            try await repository.deleteHistory(historyId: id)
            await loadMeals()
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "削除に失敗しました"
        }

        isDeleting = false
    }
}
