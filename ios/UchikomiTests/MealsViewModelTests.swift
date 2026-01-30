import XCTest
@testable import Uchikomi

final class MealsViewModelTests: XCTestCase {

    func testFormattedDateInJapanese() {
        let viewModel = MealsViewModel(repository: MockMealRepository())

        // 特定の日付を設定
        let calendar = Calendar.current
        let components = DateComponents(year: 2024, month: 1, day: 15)
        if let date = calendar.date(from: components) {
            viewModel.selectedDate = date
            XCTAssertEqual(viewModel.formattedDate, "1月15日(月)")
        }
    }

    func testIsTodayCheck() {
        let viewModel = MealsViewModel(repository: MockMealRepository())

        // デフォルトは今日
        XCTAssertTrue(viewModel.isToday)

        // 昨日に設定
        if let yesterday = Calendar.current.date(byAdding: .day, value: -1, to: Date()) {
            viewModel.selectedDate = yesterday
            XCTAssertFalse(viewModel.isToday)
        }
    }

    func testGoToPreviousDay() {
        let viewModel = MealsViewModel(repository: MockMealRepository())
        let originalDate = viewModel.selectedDate

        viewModel.goToPreviousDay()

        let expectedDate = Calendar.current.date(byAdding: .day, value: -1, to: originalDate)!
        XCTAssertEqual(
            Calendar.current.startOfDay(for: viewModel.selectedDate),
            Calendar.current.startOfDay(for: expectedDate)
        )
    }

    func testGoToNextDayNotAllowedWhenToday() {
        let viewModel = MealsViewModel(repository: MockMealRepository())
        let originalDate = viewModel.selectedDate

        viewModel.goToNextDay()

        // 今日の場合は日付が変わらない
        XCTAssertEqual(
            Calendar.current.startOfDay(for: viewModel.selectedDate),
            Calendar.current.startOfDay(for: originalDate)
        )
    }

    func testLoadMealsSuccess() async {
        let mockRepo = MockMealRepository()
        mockRepo.dailyMealsResult = .success(DailyMeals(
            date: "2024-01-15",
            meals: MealsByType(breakfast: [], lunch: [], dinner: [], snack: []),
            dailyTotal: DailyTotal(
                totalCalories: 1500,
                totalProtein: 60,
                totalFat: 50,
                totalCarbohydrates: 180
            )
        ))

        let viewModel = MealsViewModel(repository: mockRepo)

        await viewModel.loadMeals()

        XCTAssertNotNil(viewModel.dailyMeals)
        XCTAssertEqual(viewModel.dailyMeals?.dailyTotal.totalCalories, 1500)
        XCTAssertNil(viewModel.errorMessage)
    }

    func testLoadMealsFailure() async {
        let mockRepo = MockMealRepository()
        mockRepo.dailyMealsResult = .failure(APIError.networkError(NSError(domain: "", code: -1)))

        let viewModel = MealsViewModel(repository: mockRepo)

        await viewModel.loadMeals()

        XCTAssertNil(viewModel.dailyMeals)
        XCTAssertNotNil(viewModel.errorMessage)
    }
}

// MARK: - Mock

private class MockMealRepository: MealRepositoryProtocol {
    var dailyMealsResult: Result<DailyMeals, Error> = .failure(APIError.notFound)
    var uploadImageResult: Result<String, Error> = .failure(APIError.notFound)
    var analysisStatusResult: Result<AnalysisStatusResponse, Error> = .failure(APIError.notFound)
    var analysisResultResult: Result<AnalysisResultResponse, Error> = .failure(APIError.notFound)

    func getDailyMeals(date: Date) async throws -> DailyMeals {
        switch dailyMealsResult {
        case .success(let response):
            return response
        case .failure(let error):
            throw error
        }
    }

    func uploadImage(data: Data, filename: String, mealType: MealType, mealDate: Date) async throws -> String {
        switch uploadImageResult {
        case .success(let id):
            return id
        case .failure(let error):
            throw error
        }
    }

    func checkAnalysisStatus(id: String) async throws -> AnalysisStatusResponse {
        switch analysisStatusResult {
        case .success(let response):
            return response
        case .failure(let error):
            throw error
        }
    }

    func getAnalysisResult(id: String) async throws -> AnalysisResultResponse {
        switch analysisResultResult {
        case .success(let response):
            return response
        case .failure(let error):
            throw error
        }
    }
}
