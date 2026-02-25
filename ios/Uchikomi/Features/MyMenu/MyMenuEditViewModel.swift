import Foundation
import UIKit

// MARK: - MyMenuEditViewModel

@Observable
final class MyMenuEditViewModel {
    // MARK: - Constants

    private enum Constants {
        static let pollingIntervalNanoseconds: UInt64 = 2_000_000_000
        static let maxPollingAttempts = 60
    }

    // MARK: - Properties

    var menuName: String = ""
    var foodItems: [FoodEditItem] = []
    var isLoading = false
    var isSaving = false
    var errorMessage: String?
    /// 保存は成功したがmicronutrientsの自動分析に失敗した場合のノンブロッキング警告
    var analysisWarning: String?
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

    var totalMicronutrients: [String: Double] {
        // 編集モードかつAPIから取得したマイクロニュートリエント合計がある場合は、それを優先して使用する
        // （保存時にバックエンドで集計・永続化された値がFirestoreの状態と一致しているため）
        if isEditMode, let micros = existingMenuItem?.totalMicronutrients, !micros.isEmpty {
            return micros
        }
        var result: [String: Double] = [:]
        for item in foodItems {
            for (key, value) in item.micronutrients {
                result[key, default: 0] += value
            }
        }
        return result
    }

    var isValid: Bool {
        !menuName.isEmpty && !foodItems.isEmpty
    }

    func save() async {
        guard isValid else {
            errorMessage = "メニュー名と少なくとも1つの食品を入力してください"
            return
        }

        isSaving = true
        errorMessage = nil
        defer { isSaving = false }

        // 新規作成時のみ: micronutrientsがない食品があれば自動的にGemini分析を実行してから保存する
        // 分析に失敗した場合はmicronutrientsなしで保存を続行し、analysisWarningがセットされる
        if !isEditMode {
            let needsAnalysis = foodItems.contains { $0.micronutrients.isEmpty }
            if needsAnalysis {
                await autoAnalyzeForMicronutrients()
            }
        }

        let foods = foodItems.map { item in
            NutritionInfo(
                name: item.name,
                estimatedAmount: item.quantity,
                caloriesKcal: item.calories,
                proteinG: item.protein,
                fatG: item.fat,
                carbohydratesG: item.carbohydrates,
                micronutrients: item.micronutrients.isEmpty ? nil : item.micronutrients
            )
        }

        do {
            if let item = existingMenuItem {
                _ = try await repository.updateMyMenu(id: item.id, name: menuName, foods: foods)
            } else {
                _ = try await repository.createMyMenu(name: menuName, foods: foods)
            }
            shouldDismiss = true
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            #if DEBUG
            debugPrint("[MyMenuEditViewModel] save unexpected error: \(error)")
            #endif
            errorMessage = "保存に失敗しました"
        }
    }

    func delete() async {
        guard let item = existingMenuItem else { return }

        isSaving = true
        errorMessage = nil
        defer { isSaving = false }

        do {
            try await repository.deleteMyMenu(id: item.id)
            shouldDismiss = true
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "削除に失敗しました"
        }
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

    var canAnalyze: Bool {
        selectedImage != nil || hasValidManualInput
    }
}

// MARK: - Analysis

extension MyMenuEditViewModel {
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
                mealType: .lunch, // ダミー値（マイメニューでは使用しない）
                mealDate: Date() // ダミー値（マイメニューでは使用しない）
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
                mealType: .lunch, // ダミー値（マイメニューでは使用しない）
                mealDate: Date() // ダミー値（マイメニューでは使用しない）
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

    func pollForCompletion(id: String, maxAttempts: Int = Constants.maxPollingAttempts) async throws {
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

    func buildInputText(from foods: [FoodEditItem]) -> String {
        foods
            .filter { !$0.name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty }
            .map { food in
                let name = food.name.trimmingCharacters(in: .whitespacesAndNewlines)
                let quantity = food.quantity.trimmingCharacters(in: .whitespacesAndNewlines)
                return quantity.isEmpty ? name : "\(name) \(quantity)"
            }
            .joined(separator: ", ")
    }

    func applyAnalysisResult() {
        guard let result = analysisResult else {
            // 分析成功後に結果がないのはロジックエラー
            #if DEBUG
            assertionFailure("applyAnalysisResult called but analysisResult is nil")
            #endif
            errorMessage = "分析結果の読み込みに失敗しました。再度分析を行ってください。"
            return
        }

        // 分析結果をfoodItemsに反映
        // 分析後は、このfoodItemsをベースに保存される
        // micronutrientsがnilの場合は空辞書にフォールバックし、UIでセクションが非表示になる（正常動作）
        foodItems = result.result.foods.map { nutritionInfo in
            FoodEditItem(
                name: nutritionInfo.name,
                quantity: nutritionInfo.estimatedAmount,
                calories: nutritionInfo.caloriesKcal,
                protein: nutritionInfo.proteinG,
                fat: nutritionInfo.fatG,
                carbohydrates: nutritionInfo.carbohydratesG,
                micronutrients: nutritionInfo.micronutrients ?? [:]
            )
        }
    }

    /// 全食品をGemini分析してmicronutrientsをfoodItemsに反映する。
    /// 実際には全食品を一括送信し、APIのレスポンス順序がfoodItemsと一致することを前提としてインデックスで突き合わせる。
    /// 分析に失敗した場合はmicronutrientsなしで保存を続行し、analysisWarningにメッセージをセットする。
    func autoAnalyzeForMicronutrients() async {
        let inputText = buildInputText(from: foodItems)
        guard !inputText.isEmpty else { return }

        isAnalyzing = true
        defer { isAnalyzing = false }

        do {
            let id = try await mealRepository.analyzeText(
                inputText: inputText,
                mealType: .lunch, // ダミー値（マイメニューでは使用しない）
                mealDate: Date() // ダミー値（マイメニューでは使用しない）
            )

            guard !Task.isCancelled else { return }

            try await pollForCompletion(id: id)

            guard !Task.isCancelled else { return }

            let result = try await mealRepository.getAnalysisResult(id: id)

            // インデックスでマッチングしてmicronutrientsのみを反映する
            // （カロリー等の手動入力値は上書きしない）
            // APIがfoodsを追加・統合した場合でもfoodItemsの個数を超えるレスポンスは無視する（意図的）
            for (index, analysisFood) in result.result.foods.enumerated() {
                guard index < foodItems.count else { break }
                if let micros = analysisFood.micronutrients, !micros.isEmpty {
                    foodItems[index].micronutrients = micros
                }
            }
        } catch is CancellationError {
            return
        } catch let error as APIError {
            // 既知のAPIエラー - micronutrientsなしで保存を続行
            analysisWarning = "ビタミン・ミネラルの自動分析に失敗しました。保存後に手動で再分析できます。"
            #if DEBUG
            debugPrint("[MyMenuEditViewModel] autoAnalyzeForMicronutrients APIError: \(error)")
            #endif
        } catch {
            // 予期しないエラー - micronutrientsなしで保存を続行
            analysisWarning = "ビタミン・ミネラルの自動分析に失敗しました。保存後に手動で再分析できます。"
            #if DEBUG
            debugPrint("[MyMenuEditViewModel] autoAnalyzeForMicronutrients unexpected error: \(error)")
            #endif
        }
    }
}
