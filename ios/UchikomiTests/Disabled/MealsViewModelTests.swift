import Foundation
import Testing
@testable import Uchikomi

@Suite
struct MealsViewModelTests {
    private let emptyDailyMeals = DailyMeals(
        date: "2024-01-15",
        meals: MealsByType(breakfast: [], lunch: [], dinner: [], snack: []),
        dailyTotal: DailyTotal(
            totalCalories: 0,
            totalProtein: 0,
            totalFat: 0,
            totalCarbohydrates: 0
        )
    )

    private func createMockRepository() -> MealRepositoryProtocolMock {
        let mock = MealRepositoryProtocolMock()
        mock.getDailyMealsHandler = { _ in emptyDailyMeals }
        return mock
    }

    @Test
    func 日付が日本語形式でフォーマットされるべき() {
        let viewModel = MealsViewModel(repository: createMockRepository())

        let calendar = Calendar.current
        let components = DateComponents(year: 2_024, month: 1, day: 15)
        if let date = calendar.date(from: components) {
            viewModel.selectedDate = date
            #expect(viewModel.formattedDate == "1月15日(月)")
        }
    }

    @Test
    func 今日が選択されている場合isTodayがtrueになるべき() {
        let viewModel = MealsViewModel(repository: createMockRepository())
        #expect(viewModel.isToday == true)
    }

    @Test
    func 昨日が選択されている場合isTodayがfalseになるべき() {
        let viewModel = MealsViewModel(repository: createMockRepository())

        if let yesterday = Calendar.current.date(byAdding: .day, value: -1, to: Date()) {
            viewModel.selectedDate = yesterday
            #expect(viewModel.isToday == false)
        }
    }

    @Test
    func 前日へ移動すると日付が1日前になるべき() async throws {
        let mockRepo = createMockRepository()
        let viewModel = MealsViewModel(repository: mockRepo)
        let originalDate = viewModel.selectedDate

        viewModel.goToPreviousDay()

        try await Task.sleep(nanoseconds: 100_000_000)

        let expectedDate = try #require(Calendar.current.date(byAdding: .day, value: -1, to: originalDate))
        #expect(
            Calendar.current.startOfDay(for: viewModel.selectedDate) ==
                Calendar.current.startOfDay(for: expectedDate)
        )
    }

    @Test
    func 今日の場合は翌日へ移動できないべき() async throws {
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

    @Test
    func 食事データの取得が成功した場合dailyMealsが設定されるべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        mockRepo.getDailyMealsHandler = { _ in
            DailyMeals(
                date: "2024-01-15",
                meals: MealsByType(breakfast: [], lunch: [], dinner: [], snack: []),
                dailyTotal: DailyTotal(
                    totalCalories: 1_500,
                    totalProtein: 60,
                    totalFat: 50,
                    totalCarbohydrates: 180
                )
            )
        }

        let viewModel = MealsViewModel(repository: mockRepo)

        await viewModel.loadMeals()

        #expect(viewModel.dailyMeals != nil)
        #expect(viewModel.dailyMeals?.dailyTotal.totalCalories == 1_500)
        #expect(viewModel.errorMessage == nil)
    }

    @Test
    func 食事データの取得が失敗した場合エラーメッセージが設定されるべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        mockRepo.getDailyMealsHandler = { _ in
            throw APIError.networkError(NSError(domain: "", code: -1))
        }

        let viewModel = MealsViewModel(repository: mockRepo)

        await viewModel.loadMeals()

        #expect(viewModel.dailyMeals == nil)
        #expect(viewModel.errorMessage != nil)
    }

    @Test
    func goToTodayで今日の日付に戻るべき() async throws {
        let mockRepo = createMockRepository()
        let viewModel = MealsViewModel(repository: mockRepo)

        if let threeDaysAgo = Calendar.current.date(byAdding: .day, value: -3, to: Date()) {
            viewModel.selectedDate = threeDaysAgo
        }

        viewModel.goToToday()

        try await Task.sleep(nanoseconds: 100_000_000)

        #expect(viewModel.isToday == true)
    }

    @Test
    func 履歴削除が成功した場合食事データがリロードされるべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        mockRepo.deleteHistoryHandler = { _ in }
        mockRepo.getDailyMealsHandler = { _ in emptyDailyMeals }

        let viewModel = MealsViewModel(repository: mockRepo)

        await viewModel.deleteHistory(id: "test-id")

        #expect(mockRepo.deleteHistoryCallCount == 1)
        #expect(mockRepo.getDailyMealsCallCount == 1)
        #expect(viewModel.errorMessage == nil)
        #expect(viewModel.isDeleting == false)
    }

    @Test
    func MealsByTypeのisSkippedがskippedレコードを正しく判定するべき() {
        let skippedMeal = HistoryDetail(
            id: "skip-1",
            inputType: .skipped,
            imagePath: nil,
            inputText: nil,
            createdAt: "2024-01-15T00:00:00Z",
            mealType: .breakfast,
            mealDate: "2024-01-15",
            totalCalories: 0,
            totalProtein: 0,
            totalFat: 0,
            totalCarbohydrates: 0,
            foods: []
        )
        let normalMeal = HistoryDetail(
            id: "meal-1",
            inputType: .text,
            imagePath: nil,
            inputText: "サラダ",
            createdAt: "2024-01-15T00:00:00Z",
            mealType: .lunch,
            mealDate: "2024-01-15",
            totalCalories: 200,
            totalProtein: 10,
            totalFat: 5,
            totalCarbohydrates: 20,
            foods: []
        )

        let mealsByType = MealsByType(
            breakfast: [skippedMeal],
            lunch: [normalMeal],
            dinner: [],
            snack: []
        )

        #expect(mealsByType.isSkipped(for: .breakfast) == true)
        #expect(mealsByType.isSkipped(for: .lunch) == false)
        #expect(mealsByType.isSkipped(for: .dinner) == false)
        #expect(mealsByType.isSkipped(for: .snack) == false)
    }

    @Test
    func 履歴削除が失敗した場合エラーメッセージが設定されるべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        mockRepo.deleteHistoryHandler = { _ in
            throw APIError.networkError(NSError(domain: "", code: -1))
        }
        mockRepo.getDailyMealsHandler = { _ in emptyDailyMeals }

        let viewModel = MealsViewModel(repository: mockRepo)

        await viewModel.deleteHistory(id: "test-id")

        #expect(mockRepo.deleteHistoryCallCount == 1)
        #expect(viewModel.errorMessage != nil)
        #expect(viewModel.isDeleting == false)
    }
}
