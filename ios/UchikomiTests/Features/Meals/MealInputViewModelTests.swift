import Foundation
import Testing
@testable import Uchikomi

@Suite
struct MealInputViewModelTests {
    @Test
    @MainActor
    func 空文字でテキスト分析するとエラーになるべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        let viewModel = MealInputViewModel(repository: mockRepo)
        viewModel.inputText = "   "

        await viewModel.analyzeText()

        #expect(viewModel.errorMessage == "食事内容を入力してください")
        #expect(viewModel.isAnalyzing == false)
        #expect(mockRepo.analyzeTextCallCount == 0)
    }

    @Test
    @MainActor
    func 1000文字超過でエラーになるべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        let viewModel = MealInputViewModel(repository: mockRepo)
        viewModel.inputText = String(repeating: "あ", count: 1_001)

        await viewModel.analyzeText()

        #expect(viewModel.errorMessage == "入力は1000文字以内にしてください")
        #expect(viewModel.isAnalyzing == false)
        #expect(mockRepo.analyzeTextCallCount == 0)
    }

    @Test
    @MainActor
    func テキスト分析成功時にエディタを表示すべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        mockRepo.analyzeTextHandler = { _, _, _ in "test-id" }
        mockRepo.checkAnalysisStatusHandler = { _ in
            AnalysisStatusResponse(status: "completed", error: nil, message: nil)
        }
        mockRepo.getAnalysisResultHandler = { _ in
            AnalysisResultResponse(
                status: "completed",
                result: AnalysisResult(
                    foods: [
                        NutritionInfo(
                            name: "鶏むね肉",
                            estimatedAmount: "100g",
                            caloriesKcal: 108,
                            proteinG: 22.3,
                            fatG: 1.5,
                            carbohydratesG: 0
                        ),
                    ],
                    totalCalories: 108,
                    totalProtein: 22.3,
                    totalFat: 1.5,
                    totalCarbohydrates: 0
                )
            )
        }

        let viewModel = MealInputViewModel(repository: mockRepo)
        viewModel.inputText = "鶏むね肉100g"

        await viewModel.analyzeText()

        #expect(viewModel.showEditor == true)
        #expect(viewModel.analysisId == "test-id")
        #expect(viewModel.analysisResult != nil)
        #expect(mockRepo.analyzeTextCallCount == 1)
    }

    @Test
    @MainActor
    func テキスト分析失敗時にエラーメッセージを表示すべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        mockRepo.analyzeTextHandler = { _, _, _ in
            throw APIError.serverError("サーバーエラーが発生しました")
        }

        let viewModel = MealInputViewModel(repository: mockRepo)
        viewModel.inputText = "鶏むね肉100g"

        await viewModel.analyzeText()

        #expect(viewModel.showEditor == false)
        #expect(viewModel.errorMessage != nil)
        #expect(viewModel.isAnalyzing == false)
    }

    @Test
    @MainActor
    func 分析待機中にタイムアウトした場合エラーメッセージを表示すべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        mockRepo.analyzeTextHandler = { _, _, _ in "test-id" }

        var callCount = 0
        mockRepo.checkAnalysisStatusHandler = { _ in
            callCount += 1
            if callCount >= 3 {
                throw APIError.serverError("分析がタイムアウトしました（120秒経過）")
            }
            return AnalysisStatusResponse(status: "processing", error: nil, message: nil)
        }

        let viewModel = MealInputViewModel(repository: mockRepo)
        viewModel.inputText = "鶏むね肉100g"

        await viewModel.analyzeText()

        #expect(viewModel.showEditor == false)
        #expect(viewModel.errorMessage != nil)
    }

    // MARK: - 追加テストケース

    @Test
    @MainActor
    func 分析ステータスがfailedの場合エラーメッセージを表示すべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        mockRepo.analyzeTextHandler = { _, _, _ in "test-id" }
        mockRepo.checkAnalysisStatusHandler = { _ in
            AnalysisStatusResponse(status: "failed", error: "画像の解析に失敗しました", message: nil)
        }

        let viewModel = MealInputViewModel(repository: mockRepo)
        viewModel.inputText = "鶏むね肉100g"

        await viewModel.analyzeText()

        #expect(viewModel.showEditor == false)
        #expect(viewModel.errorMessage != nil)
        #expect(viewModel.errorMessage?.contains("画像の解析に失敗しました") == true)
    }

    @Test
    @MainActor
    func 分析ステータスが不明な場合エラーメッセージを表示すべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        mockRepo.analyzeTextHandler = { _, _, _ in "test-id" }
        mockRepo.checkAnalysisStatusHandler = { _ in
            AnalysisStatusResponse(status: "unknown_status", error: nil, message: nil)
        }

        let viewModel = MealInputViewModel(repository: mockRepo)
        viewModel.inputText = "鶏むね肉100g"

        await viewModel.analyzeText()

        #expect(viewModel.showEditor == false)
        #expect(viewModel.errorMessage != nil)
        #expect(viewModel.errorMessage?.contains("分析ステータスが不明です") == true)
    }

    @Test
    @MainActor
    func 文字数1000文字ちょうどで分析が成功すべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        mockRepo.analyzeTextHandler = { _, _, _ in "test-id" }
        mockRepo.checkAnalysisStatusHandler = { _ in
            AnalysisStatusResponse(status: "completed", error: nil, message: nil)
        }
        mockRepo.getAnalysisResultHandler = { _ in
            AnalysisResultResponse(
                status: "completed",
                result: AnalysisResult(
                    foods: [],
                    totalCalories: 0,
                    totalProtein: 0,
                    totalFat: 0,
                    totalCarbohydrates: 0
                )
            )
        }

        let viewModel = MealInputViewModel(repository: mockRepo)
        viewModel.inputText = String(repeating: "あ", count: 1_000)

        await viewModel.analyzeText()

        #expect(viewModel.errorMessage == nil)
        #expect(viewModel.showEditor == true)
        #expect(mockRepo.analyzeTextCallCount == 1)
    }

    @Test
    @MainActor
    func 分析結果取得失敗時にエラーメッセージを表示すべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        mockRepo.analyzeTextHandler = { _, _, _ in "test-id" }
        mockRepo.checkAnalysisStatusHandler = { _ in
            AnalysisStatusResponse(status: "completed", error: nil, message: nil)
        }
        mockRepo.getAnalysisResultHandler = { _ in
            throw APIError.serverError("結果の取得に失敗しました")
        }

        let viewModel = MealInputViewModel(repository: mockRepo)
        viewModel.inputText = "鶏むね肉100g"

        await viewModel.analyzeText()

        #expect(viewModel.showEditor == false)
        #expect(viewModel.errorMessage != nil)
    }

    @Test
    @MainActor
    func 非APIErrorでもエラーメッセージを表示すべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        mockRepo.analyzeTextHandler = { _, _, _ in
            throw NSError(domain: "TestError", code: 999, userInfo: [NSLocalizedDescriptionKey: "テストエラー"])
        }

        let viewModel = MealInputViewModel(repository: mockRepo)
        viewModel.inputText = "鶏むね肉100g"

        await viewModel.analyzeText()

        #expect(viewModel.showEditor == false)
        #expect(viewModel.errorMessage != nil)
        #expect(viewModel.errorMessage?.contains("テキスト分析に失敗しました") == true)
    }

    @Test
    @MainActor
    func deleteHistory成功時にtrueを返すべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        mockRepo.deleteHistoryHandler = { _ in }

        let viewModel = MealInputViewModel(repository: mockRepo)

        let result = await viewModel.deleteHistory(id: "test-id")

        #expect(result == true)
        #expect(viewModel.errorMessage == nil)
        #expect(mockRepo.deleteHistoryCallCount == 1)
    }

    @Test
    @MainActor
    func deleteHistory失敗時にfalseを返すべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        mockRepo.deleteHistoryHandler = { _ in
            throw APIError.serverError("削除に失敗しました")
        }

        let viewModel = MealInputViewModel(repository: mockRepo)

        let result = await viewModel.deleteHistory(id: "test-id")

        #expect(result == false)
        #expect(viewModel.errorMessage != nil)
        #expect(viewModel.errorMessage?.contains("削除に失敗しました") == true)
    }

    // MARK: - skipMeal

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
