import Foundation
import PhotosUI
import SwiftUI

@Observable
final class MealInputViewModel {
    // MARK: - Constants

    private enum Constants {
        static let maxImageSizeBytes = 10 * 1_024 * 1_024 // 10MB
    }

    // MARK: - Properties

    var selectedMealType: MealType = .lunch
    var mealDate = Date()
    var selectedImage: UIImage?
    var inputText: String = ""
    var manualFoods: [FoodEditItem] = [FoodEditItem()]
    var isAnalyzing = false
    var errorMessage: String?
    var isCompleted = false
    var isSkipping = false

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

        guard imageData.count <= Constants.maxImageSizeBytes else {
            errorMessage = "画像サイズが大きすぎます。10MB以下の画像を選択してください"
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
            // 分析リクエストを送信したら即座に完了とする（バックグラウンドで分析が続く）
            isCompleted = true
        } catch is CancellationError {
            return
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            #if DEBUG
            debugPrint("[MealInputViewModel] Unexpected error: \(error)")
            #endif
            errorMessage = "画像分析リクエストに失敗しました: \(error.localizedDescription)"
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
            // 分析リクエストを送信したら即座に完了とする（バックグラウンドで分析が続く）
            isCompleted = true
        } catch is CancellationError {
            return
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            #if DEBUG
            debugPrint("[MealInputViewModel] Unexpected error: \(error)")
            #endif
            errorMessage = "テキスト分析リクエストに失敗しました: \(error.localizedDescription)"
        }
    }

    func markCompleted() {
        isCompleted = true
    }

    func skipMeal() async -> Bool {
        guard !isSkipping else { return false }
        isSkipping = true
        defer { isSkipping = false }

        do {
            try await repository.skipMeal(mealType: selectedMealType, mealDate: mealDate)
            return true
        } catch is CancellationError {
            return false
        } catch let error as APIError {
            #if DEBUG
            debugPrint("[MealInputViewModel] Skip meal error: \(error)")
            #endif
            errorMessage = "スキップに失敗しました: \(error.localizedDescription)"
            return false
        } catch {
            #if DEBUG
            debugPrint("[MealInputViewModel] Skip meal unexpected error: \(error)")
            #endif
            errorMessage = "スキップに失敗しました: \(error.localizedDescription)"
            return false
        }
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
