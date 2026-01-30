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
    var errorMessage: String?
    var isCompleted = false

    private var analysisId: String?
    private let repository: MealRepositoryProtocol

    init(repository: MealRepositoryProtocol = MealRepository()) {
        self.repository = repository
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
            // 画像アップロード（meal_typeとmeal_dateも一緒に送信）
            // バックエンドは分析完了時に自動的に食事データを保存する
            let id = try await repository.uploadImage(
                data: imageData,
                filename: "meal.jpg",
                mealType: selectedMealType,
                mealDate: mealDate
            )
            analysisId = id

            // ポーリングで完了を待つ
            try await pollForCompletion(id: id)

            // 結果を取得
            analysisResult = try await repository.getAnalysisResult(id: id)

            // 分析完了 = 保存完了（バックエンドで自動保存される）
            isCompleted = true
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "画像分析に失敗しました"
        }

        isAnalyzing = false
    }

    private func pollForCompletion(id: String, maxAttempts: Int = 60) async throws {
        for _ in 0..<maxAttempts {
            let status = try await repository.checkAnalysisStatus(id: id)

            switch status.status {
            case "completed":
                return
            case "failed":
                // バックエンドのエラーメッセージがあれば表示
                let errorMessage = status.error ?? "分析に失敗しました"
                throw APIError.serverError(errorMessage)
            default:
                // pending or processing
                try await Task.sleep(nanoseconds: 2_000_000_000) // 2秒待機
            }
        }

        throw APIError.serverError("分析がタイムアウトしました（2分経過）")
    }


    func reset() {
        selectedImage = nil
        analysisResult = nil
        analysisId = nil
        errorMessage = nil
        isCompleted = false
    }
}
