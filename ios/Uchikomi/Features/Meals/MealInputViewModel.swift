import Foundation
import SwiftUI
import PhotosUI

@Observable
final class MealInputViewModel {
    var selectedMealType: MealType = .lunch
    var mealDate = Date()
    var selectedImage: UIImage?
    var analysisResult: AnalysisResultResponse?
    var isAnalyzing = false
    var isSaving = false
    var errorMessage: String?
    var isCompleted = false

    private var analysisId: String?
    private let repository: MealRepositoryProtocol

    init(repository: MealRepositoryProtocol = MealRepository()) {
        self.repository = repository
    }

    var canSave: Bool {
        analysisResult != nil && !isSaving
    }

    func analyzeImage() async {
        guard let image = selectedImage,
              let imageData = image.jpegData(compressionQuality: 0.8) else {
            errorMessage = "画像を選択してください"
            return
        }

        isAnalyzing = true
        errorMessage = nil

        do {
            // 画像アップロード
            let id = try await repository.uploadImage(data: imageData, filename: "meal.jpg")
            analysisId = id

            // ポーリングで完了を待つ
            try await pollForCompletion(id: id)

            // 結果を取得
            analysisResult = try await repository.getAnalysisResult(id: id)
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "画像分析に失敗しました"
        }

        isAnalyzing = false
    }

    private func pollForCompletion(id: String, maxAttempts: Int = 30) async throws {
        for _ in 0..<maxAttempts {
            let status = try await repository.checkAnalysisStatus(id: id)

            switch status.status {
            case "completed":
                return
            case "failed":
                throw APIError.serverError("分析に失敗しました")
            default:
                // pending or processing
                try await Task.sleep(nanoseconds: 2_000_000_000) // 2秒待機
            }
        }

        throw APIError.serverError("分析がタイムアウトしました")
    }

    func saveMeal() async {
        guard let analysisId = analysisId else {
            errorMessage = "分析が完了していません"
            return
        }

        isSaving = true
        errorMessage = nil

        do {
            try await repository.saveMeal(
                analysisId: analysisId,
                mealType: selectedMealType,
                mealDate: mealDate
            )
            isCompleted = true
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "保存に失敗しました"
        }

        isSaving = false
    }

    func reset() {
        selectedImage = nil
        analysisResult = nil
        analysisId = nil
        errorMessage = nil
        isCompleted = false
    }
}
