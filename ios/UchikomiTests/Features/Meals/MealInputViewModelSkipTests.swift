import Foundation
import Testing
@testable import Uchikomi

@Suite
struct MealInputViewModelSkipTests {
    @Test
    @MainActor
    func skipMeal成功時にtrueを返し正しい引数で呼ばれるべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        var capturedMealType: MealType?
        var capturedDate: Date?
        mockRepo.skipMealHandler = { mealType, mealDate in
            capturedMealType = mealType
            capturedDate = mealDate
        }

        let viewModel = MealInputViewModel(repository: mockRepo)
        viewModel.selectedMealType = .breakfast
        viewModel.mealDate = Date()

        let result = await viewModel.skipMeal()

        #expect(result == true)
        #expect(mockRepo.skipMealCallCount == 1)
        #expect(capturedMealType == .breakfast)
        #expect(capturedDate != nil)
        #expect(
            Calendar.current.startOfDay(for: capturedDate!) ==
                Calendar.current.startOfDay(for: viewModel.mealDate)
        )
        #expect(viewModel.errorMessage == nil)
        #expect(viewModel.isSkipping == false)
    }

    @Test
    @MainActor
    func skipMealがAPIError失敗時にfalseを返しエラーメッセージが設定されるべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        mockRepo.skipMealHandler = { _, _ in
            throw APIError.networkError(NSError(domain: "", code: -1))
        }

        let viewModel = MealInputViewModel(repository: mockRepo)

        let result = await viewModel.skipMeal()

        #expect(result == false)
        #expect(mockRepo.skipMealCallCount == 1)
        #expect(viewModel.errorMessage != nil)
        #expect(viewModel.isSkipping == false)
    }

    @Test
    @MainActor
    func skipMealが非APIError失敗時にフォールバックメッセージが設定されるべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        mockRepo.skipMealHandler = { _, _ in
            throw NSError(domain: "test", code: 999)
        }

        let viewModel = MealInputViewModel(repository: mockRepo)

        let result = await viewModel.skipMeal()

        #expect(result == false)
        #expect(viewModel.errorMessage == "スキップに失敗しました")
        #expect(viewModel.isSkipping == false)
    }

    @Test
    @MainActor
    func skipMeal中はisSkippingがtrueになるべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        let viewModel = MealInputViewModel(repository: mockRepo)
        var wasSkippingDuringCall = false

        mockRepo.skipMealHandler = { _, _ in
            wasSkippingDuringCall = viewModel.isSkipping
        }

        _ = await viewModel.skipMeal()

        #expect(wasSkippingDuringCall == true)
        #expect(viewModel.isSkipping == false)
    }

    @Test
    @MainActor
    func skipMealはisSkipping中に二重呼び出しを防止するべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        let viewModel = MealInputViewModel(repository: mockRepo)

        mockRepo.skipMealHandler = { _, _ in
            let secondResult = await viewModel.skipMeal()
            #expect(secondResult == false)
        }

        let result = await viewModel.skipMeal()

        #expect(result == true)
        #expect(mockRepo.skipMealCallCount == 1)
    }
}
