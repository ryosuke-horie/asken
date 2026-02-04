import Foundation
import PhotosUI
import SwiftUI

@Observable
final class MealInputViewModel {
    // MARK: - Constants

    private enum Constants {
        static let pollingIntervalNanoseconds: UInt64 = 2_000_000_000
        static let maxPollingAttempts = 60
        static let pollingTimeoutSeconds = 120
    }

    // MARK: - Properties

    var selectedMealType: MealType = .lunch
    var mealDate = Date()
    var selectedImage: UIImage?
    var inputText: String = ""
    var manualFoods: [FoodEditItem] = [FoodEditItem()]
    var analysisResult: AnalysisResultResponse?
    var isAnalyzing = false
    var errorMessage: String?
    var isCompleted = false
    var showEditor = false

    private(set) var analysisId: String?
    private let repository: MealRepositoryProtocol

    init(repository: MealRepositoryProtocol = MealRepository()) {
        self.repository = repository
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

    // MARK: - Unified Analysis

    func analyze() async {
        if selectedImage != nil {
            await analyzeImage()
        } else if hasValidManualInput {
            inputText = buildInputText(from: manualFoods)
            await analyzeText()
        } else {
            errorMessage = "食事内容を入力するか、画像を選択してください"
        }
    }

    func analyzeImage() async {
        guard let image = selectedImage,
              let imageData = image.jpegData(compressionQuality: 0.8) else {
            errorMessage = "画像を選択してください"
            return
        }

        isAnalyzing = true
        errorMessage = nil
        defer { isAnalyzing = false }

        do {
            let id = try await repository.uploadImage(
                data: imageData,
                filename: "meal.jpg",
                mealType: selectedMealType,
                mealDate: mealDate
            )

            guard !Task.isCancelled else { return }
            analysisId = id

            try await pollForCompletion(id: id)

            guard !Task.isCancelled else { return }

            analysisResult = try await repository.getAnalysisResult(id: id)
            showEditor = true
        } catch is CancellationError {
            return
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            #if DEBUG
            debugPrint("[MealInputViewModel] Unexpected error: \(error)")
            #endif
            errorMessage = "画像分析に失敗しました: \(error.localizedDescription)"
        }
    }

    func analyzeText() async {
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
            let id = try await repository.analyzeText(
                inputText: trimmedText,
                mealType: selectedMealType,
                mealDate: mealDate
            )

            guard !Task.isCancelled else { return }
            analysisId = id

            try await pollForCompletion(id: id)

            guard !Task.isCancelled else { return }

            analysisResult = try await repository.getAnalysisResult(id: id)
            showEditor = true
        } catch is CancellationError {
            return
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            #if DEBUG
            debugPrint("[MealInputViewModel] Unexpected error: \(error)")
            #endif
            errorMessage = "テキスト分析に失敗しました: \(error.localizedDescription)"
        }
    }

    private func pollForCompletion(id: String, maxAttempts: Int = Constants.maxPollingAttempts) async throws {
        for _ in 0 ..< maxAttempts {
            let status = try await repository.checkAnalysisStatus(id: id)

            switch status.status {
            case "completed":
                return
            case "failed":
                let errorMessage = status.error ?? "分析に失敗しました"
                throw APIError.serverError(errorMessage)
            case "pending", "processing":
                try await Task.sleep(nanoseconds: Constants.pollingIntervalNanoseconds)
            default:
                #if DEBUG
                debugPrint("[MealInputViewModel] Unknown analysis status: \(status.status)")
                #endif
                throw APIError.serverError("分析ステータスが不明です: \(status.status)")
            }
        }

        throw APIError.serverError("分析がタイムアウトしました（\(Constants.pollingTimeoutSeconds)秒経過）")
    }

    func reset() {
        selectedImage = nil
        inputText = ""
        manualFoods = [FoodEditItem()]
        analysisResult = nil
        analysisId = nil
        errorMessage = nil
        isCompleted = false
        showEditor = false
    }

    func markCompleted() {
        isCompleted = true
    }

    func deleteHistory(id: String) async -> Bool {
        do {
            try await repository.deleteHistory(historyId: id)
            return true
        } catch let error as APIError {
            #if DEBUG
            debugPrint("[MealInputViewModel] Delete error: \(error)")
            #endif
            errorMessage = "削除に失敗しました: \(error.localizedDescription)"
            return false
        } catch {
            #if DEBUG
            debugPrint("[MealInputViewModel] Delete unexpected error: \(error)")
            #endif
            errorMessage = "削除に失敗しました: \(error.localizedDescription)"
            return false
        }
    }
}
