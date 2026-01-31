import Foundation
import Testing
@testable import Uchikomi

@Suite struct MealsViewModelTests {

    private func createMockRepository() -> MealRepositoryProtocolMock {
        let mock = MealRepositoryProtocolMock()
        mock.getDailyMealsHandler = { _ in
            DailyMeals(
                date: "2024-01-15",
                meals: MealsByType(breakfast: [], lunch: [], dinner: [], snack: []),
                dailyTotal: DailyTotal(
                    totalCalories: 0,
                    totalProtein: 0,
                    totalFat: 0,
                    totalCarbohydrates: 0
                )
            )
        }
        return mock
    }

    @Test func 日付が日本語形式でフォーマットされるべき() {
        let viewModel = MealsViewModel(repository: createMockRepository())

        let calendar = Calendar.current
        let components = DateComponents(year: 2024, month: 1, day: 15)
        if let date = calendar.date(from: components) {
            viewModel.selectedDate = date
            #expect(viewModel.formattedDate == "1月15日(月)")
        }
    }

    @Test func 今日が選択されている場合isTodayがtrueになるべき() {
        let viewModel = MealsViewModel(repository: createMockRepository())
        #expect(viewModel.isToday == true)
    }

    @Test func 昨日が選択されている場合isTodayがfalseになるべき() {
        let viewModel = MealsViewModel(repository: createMockRepository())

        if let yesterday = Calendar.current.date(byAdding: .day, value: -1, to: Date()) {
            viewModel.selectedDate = yesterday
            #expect(viewModel.isToday == false)
        }
    }

    @Test func 前日へ移動すると日付が1日前になるべき() async throws {
        let mockRepo = createMockRepository()
        let viewModel = MealsViewModel(repository: mockRepo)
        let originalDate = viewModel.selectedDate

        viewModel.goToPreviousDay()

        try await Task.sleep(nanoseconds: 100_000_000)

        let expectedDate = Calendar.current.date(byAdding: .day, value: -1, to: originalDate)!
        #expect(
            Calendar.current.startOfDay(for: viewModel.selectedDate) ==
            Calendar.current.startOfDay(for: expectedDate)
        )
    }

    @Test func 今日の場合は翌日へ移動できないべき() async throws {
        let mockRepo = createMockRepository()
        let viewModel = MealsViewModel(repository: mockRepo)
        let originalDate = viewModel.selectedDate

        viewModel.goToNextDay()

        try await Task.sleep(nanoseconds: 100_000_000)

        #expect(
            Calendar.current.startOfDay(for: viewModel.selectedDate) ==
            Calendar.current.startOfDay(for: originalDate)
        )
    }

    @Test func 食事データの取得が成功した場合dailyMealsが設定されるべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        mockRepo.getDailyMealsHandler = { _ in
            DailyMeals(
                date: "2024-01-15",
                meals: MealsByType(breakfast: [], lunch: [], dinner: [], snack: []),
                dailyTotal: DailyTotal(
                    totalCalories: 1500,
                    totalProtein: 60,
                    totalFat: 50,
                    totalCarbohydrates: 180
                )
            )
        }

        let viewModel = MealsViewModel(repository: mockRepo)

        await viewModel.loadMeals()

        #expect(viewModel.dailyMeals != nil)
        #expect(viewModel.dailyMeals?.dailyTotal.totalCalories == 1500)
        #expect(viewModel.errorMessage == nil)
    }

    @Test func 食事データの取得が失敗した場合エラーメッセージが設定されるべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        mockRepo.getDailyMealsHandler = { _ in
            throw APIError.networkError(NSError(domain: "", code: -1))
        }

        let viewModel = MealsViewModel(repository: mockRepo)

        await viewModel.loadMeals()

        #expect(viewModel.dailyMeals == nil)
        #expect(viewModel.errorMessage != nil)
    }
}
