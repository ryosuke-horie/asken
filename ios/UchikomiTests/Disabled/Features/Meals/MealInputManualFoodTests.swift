import Foundation
import Testing
@testable import Uchikomi

@Suite
struct MealInputManualFoodTests {
    // MARK: - 手動入力テスト

    @Test
    @MainActor
    func 手動入力を追加できるべき() {
        let mockRepo = MealRepositoryProtocolMock()
        let viewModel = MealInputViewModel(repository: mockRepo)

        #expect(viewModel.manualFoods.count == 1)

        viewModel.addManualFood()

        #expect(viewModel.manualFoods.count == 2)
    }

    @Test
    @MainActor
    func 手動入力を削除できるべき() {
        let mockRepo = MealRepositoryProtocolMock()
        let viewModel = MealInputViewModel(repository: mockRepo)
        viewModel.addManualFood()

        #expect(viewModel.manualFoods.count == 2)

        let foodToRemove = viewModel.manualFoods[0]
        viewModel.removeManualFood(foodToRemove)

        #expect(viewModel.manualFoods.count == 1)
    }

    @Test
    @MainActor
    func 最後の手動入力を削除しても空の行が残るべき() {
        let mockRepo = MealRepositoryProtocolMock()
        let viewModel = MealInputViewModel(repository: mockRepo)

        #expect(viewModel.manualFoods.count == 1)

        let foodToRemove = viewModel.manualFoods[0]
        viewModel.removeManualFood(foodToRemove)

        #expect(viewModel.manualFoods.count == 1)
    }

    @Test
    @MainActor
    func 有効な手動入力がある場合hasValidManualInputがtrueになるべき() {
        let mockRepo = MealRepositoryProtocolMock()
        let viewModel = MealInputViewModel(repository: mockRepo)

        viewModel.manualFoods[0].name = "鶏むね肉"

        #expect(viewModel.hasValidManualInput == true)
    }

    @Test
    @MainActor
    func 空の手動入力の場合hasValidManualInputがfalseになるべき() {
        let mockRepo = MealRepositoryProtocolMock()
        let viewModel = MealInputViewModel(repository: mockRepo)

        #expect(viewModel.hasValidManualInput == false)

        viewModel.manualFoods[0].name = "   "

        #expect(viewModel.hasValidManualInput == false)
    }

    @Test
    @MainActor
    func 手動入力も画像もない場合analyzeでエラーになるべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        let viewModel = MealInputViewModel(repository: mockRepo)

        await viewModel.analyze()

        #expect(viewModel.errorMessage == "食事内容を入力するか、画像を選択してください")
    }

    @Test
    @MainActor
    func 手動入力がある場合analyzeでテキスト分析が呼ばれるべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        mockRepo.analyzeTextHandler = { inputText, _, _ in
            #expect(inputText == "鶏むね肉 100g")
            return "test-id"
        }
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
        viewModel.manualFoods[0].name = "鶏むね肉"
        viewModel.manualFoods[0].quantity = "100g"

        await viewModel.analyze()

        #expect(mockRepo.analyzeTextCallCount == 1)
        #expect(viewModel.showEditor == true)
    }

    @Test
    @MainActor
    func 複数の手動入力がカンマ区切りでテキストに変換されるべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        var capturedInputText = ""
        mockRepo.analyzeTextHandler = { inputText, _, _ in
            capturedInputText = inputText
            return "test-id"
        }
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
        viewModel.manualFoods[0].name = "鶏むね肉"
        viewModel.manualFoods[0].quantity = "100g"
        viewModel.addManualFood()
        viewModel.manualFoods[1].name = "ご飯"
        viewModel.manualFoods[1].quantity = "1杯"

        await viewModel.analyze()

        #expect(capturedInputText == "鶏むね肉 100g, ご飯 1杯")
    }

    @Test
    @MainActor
    func 量が空でも食品名のみでテキストに変換されるべき() async {
        let mockRepo = MealRepositoryProtocolMock()
        var capturedInputText = ""
        mockRepo.analyzeTextHandler = { inputText, _, _ in
            capturedInputText = inputText
            return "test-id"
        }
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
        viewModel.manualFoods[0].name = "鶏むね肉"
        viewModel.manualFoods[0].quantity = ""

        await viewModel.analyze()

        #expect(capturedInputText == "鶏むね肉")
    }

    @Test
    @MainActor
    func resetで手動入力がクリアされるべき() {
        let mockRepo = MealRepositoryProtocolMock()
        let viewModel = MealInputViewModel(repository: mockRepo)
        viewModel.manualFoods[0].name = "鶏むね肉"
        viewModel.addManualFood()

        viewModel.reset()

        #expect(viewModel.manualFoods.count == 1)
        #expect(viewModel.manualFoods[0].name == "")
        #expect(viewModel.isSkipping == false)
    }
}
