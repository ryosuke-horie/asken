import UIKit
import Foundation

@Observable
final class MyMenuEditViewModel {
    // MARK: - Constants

    private enum Constants {
        static let pollingIntervalNanoseconds: UInt64 = 2_000_000_000
        static let maxPollingAttempts = 60
        static let pollingTimeoutSeconds = 120
    }

    // MARK: - Properties

    var menuName: String = ""
    var foodItems: [FoodEditItem] = []
    var isLoading = false
    var isSaving = false
    var errorMessage: String?
    var shouldDismiss = false

    // Analysis properties
    var selectedImage: UIImage?
    var manualFoods: [FoodEditItem] = [FoodEditItem()]
    var isAnalyzing = false
    var analysisResult: AnalysisResultResponse?

    private let repository: MyMenuRepositoryProtocol
    private let mealRepository: MealRepositoryProtocol
    private let existingMenuItem: MyMenuItem?

    init(
        repository: MyMenuRepositoryProtocol = MyMenuRepository(),
        mealRepository: MealRepositoryProtocol = MealRepository(),
        menuItem: MyMenuItem? = nil
    ) {
        self.repository = repository
        self.mealRepository = mealRepository
        self.existingMenuItem = menuItem

        if let item = menuItem {
            menuName = item.name
            foodItems = item.foods.map { FoodEditItem(from: $0) }
        }
    }

    var isEditMode: Bool {
        existingMenuItem != nil
    }

    var totalCalories: Double {
        foodItems.reduce(0) { $0 + $1.calories }
    }

    var totalProtein: Double {
        foodItems.reduce(0) { $0 + $1.protein }
    }

    var totalFat: Double {
        foodItems.reduce(0) { $0 + $1.fat }
    }

    var totalCarbohydrates: Double {
        foodItems.reduce(0) { $0 + $1.carbohydrates }
    }

    var isValid: Bool {
        !menuName.isEmpty && !foodItems.isEmpty
    }

    func addFoodItem() {
        foodItems.append(FoodEditItem())
    }

    func removeFoodItem(at index: Int) {
        foodItems.remove(at: index)
    }

    func save() async {
        guard isValid else { return }

        isSaving = true
        errorMessage = nil

        let foods = foodItems.map { item in
            NutritionInfo(
                name: item.name,
                estimatedAmount: item.quantity,
                caloriesKcal: item.calories,
                proteinG: item.protein,
                fatG: item.fat,
                carbohydratesG: item.carbohydrates
            )
        }

        do {
            if let item = existingMenuItem {
                _ = try await repository.updateMyMenu(id: item.id, name: menuName, foods: foods)
            } else {
                _ = try await repository.createMyMenu(name: menuName, foods: foods)
            }
            shouldDismiss = true // 成功フラグ（画面を閉じるため）
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "保存に失敗しました"
        }

        isSaving = false
    }

    func delete() async {
        guard let item = existingMenuItem else { return }

        isSaving = true
        errorMessage = nil

        do {
            try await repository.deleteMyMenu(id: item.id)
            shouldDismiss = true
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "削除に失敗しました"
        }

        isSaving = false
    }

    // MARK: - Manual Food Input

    func addManualFood() {
        manualFoods.append(FoodEditItem())
    }

    func removeManualFood(_ item: FoodEditItem) {
        manualFoods.removeAll { $0.id == item.id }
        if manualFoods.isEmpty {
            manualFoods.append(FoodEditItem())
        }
    }

    var hasValidManualInput: Bool {
        manualFoods.contains { !$0.name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty }
    }

    private func buildInputText(from foods: [FoodEditItem]) -> String {
        foods
            .filter { !$0.name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty }
            .map { food in
                let name = food.name.trimmingCharacters(in: .whitespacesAndNewlines)
                let quantity = food.quantity.trimmingCharacters(in: .whitespacesAndNewlines)
                return quantity.isEmpty ? name : "\(name) \(quantity)"
            }
            .joined(separator: ", ")
    }

    var canAnalyze: Bool {
        selectedImage != nil || hasValidManualInput
    }

    // MARK: - Analysis

    func analyze() async {
        if selectedImage != nil {
            await analyzeImage()
        } else if hasValidManualInput {
            await analyzeText()
        } else {
            errorMessage = "食事内容を入力するか、画像を選択してください"
        }
    }

    func analyzeImage() async {
        guard let image = selectedImage else {
            errorMessage = "画像を選択してください"
            return
        }

        // JPEG圧縮 - 失敗した場合は画像データが無効
        guard let imageData = image.jpegData(compressionQuality: 0.8) else {
            errorMessage = "画像の処理に失敗しました。別の画像を選択してください。"
            return
        }

        isAnalyzing = true
        errorMessage = nil
        defer { isAnalyzing = false }

        do {
            // NOTE: 既存の分析APIはmealTypeとmealDateを要求するため、
            // マイメニュー登録ではダミー値を使用しています。
            // 将来的にはマイメニュー専用の分析エンドポイントを作成することを推奨します。
            let id = try await mealRepository.uploadImage(
                data: imageData,
                filename: "mymenu.jpg",
                mealType: .lunch,  // ダミー値（マイメニューでは使用しない）
                mealDate: Date()    // ダミー値（マイメニューでは使用しない）
            )

            guard !Task.isCancelled else { return }

            try await pollForCompletion(id: id)

            guard !Task.isCancelled else { return }

            analysisResult = try await mealRepository.getAnalysisResult(id: id)

            // 分析結果をfoodItemsに反映
            applyAnalysisResult()
        } catch is CancellationError {
            return
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            #if DEBUG
            debugPrint("[MyMenuEditViewModel] analyzeImage unexpected error: \(error)")
            #endif
            errorMessage = "画像分析に失敗しました。ネットワークを確認してやり直してください。"
        }
    }

    func analyzeText() async {
        let inputText = buildInputText(from: manualFoods)
        let trimmedText = inputText.trimmingCharacters(in: .whitespacesAndNewlines)

        guard !trimmedText.isEmpty else {
            errorMessage = "食事内容を入力してください"
            return
        }

        guard trimmedText.count <= 1_000 else {
            errorMessage = "入力は1000文字以内にしてください"
            return
        }

        isAnalyzing = true
        errorMessage = nil
        defer { isAnalyzing = false }

        do {
            // NOTE: 既存の分析APIはmealTypeとmealDateを要求するため、
            // マイメニュー登録ではダミー値を使用しています。
            // 将来的にはマイメニュー専用の分析エンドポイントを作成することを推奨します。
            let id = try await mealRepository.analyzeText(
                inputText: trimmedText,
                mealType: .lunch,  // ダミー値（マイメニューでは使用しない）
                mealDate: Date()    // ダミー値（マイメニューでは使用しない）
            )

            guard !Task.isCancelled else { return }

            try await pollForCompletion(id: id)

            guard !Task.isCancelled else { return }

            analysisResult = try await mealRepository.getAnalysisResult(id: id)

            // 分析結果をfoodItemsに反映
            applyAnalysisResult()
        } catch is CancellationError {
            return
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            #if DEBUG
            debugPrint("[MyMenuEditViewModel] analyzeText unexpected error: \(error)")
            #endif
            errorMessage = "テキスト分析に失敗しました。ネットワークを確認してやり直してください。"
        }
    }

    private func pollForCompletion(id: String, maxAttempts: Int = Constants.maxPollingAttempts) async throws {
        for attempt in 0 ..< maxAttempts {
            let status = try await mealRepository.checkAnalysisStatus(id: id)

            switch status.status {
            case "completed":
                return
            case "failed":
                let errorMsg = status.error ?? "分析に失敗しました"
                throw APIError.serverError(errorMsg)
            case "pending", "processing":
                try await Task.sleep(nanoseconds: Constants.pollingIntervalNanoseconds)
            default:
                // 不明なステータスコード - バックエンドの仕様変更の可能性
                #if DEBUG
                debugPrint("[MyMenuEditViewModel] Unknown analysis status: \(status.status) at attempt \(attempt + 1)")
                #endif
                // 新しいステータスに対応できるよう、処理中として扱う
                try await Task.sleep(nanoseconds: Constants.pollingIntervalNanoseconds)
            }
        }

        throw APIError.serverError("分析がタイムアウトしました。時間をおいてやり直してください。")
    }

    private func applyAnalysisResult() {
        guard let result = analysisResult else {
            // 分析成功後に結果がないのはロジックエラー
            #if DEBUG
            assertionFailure("applyAnalysisResult called but analysisResult is nil")
            #endif
            return
        }

        // 分析結果をfoodItemsに反映
        // 分析後は、このfoodItemsをベースに保存される
        foodItems = result.result.foods.map { nutritionInfo in
            FoodEditItem(
                name: nutritionInfo.name,
                quantity: nutritionInfo.estimatedAmount,
                calories: nutritionInfo.caloriesKcal,
                protein: nutritionInfo.proteinG,
                fat: nutritionInfo.fatG,
                carbohydrates: nutritionInfo.carbohydratesG
            )
        }
    }
}
