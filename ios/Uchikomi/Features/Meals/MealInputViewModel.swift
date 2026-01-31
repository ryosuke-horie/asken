import Foundation
import SwiftUI
import PhotosUI

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

            // 編集画面を表示
            showEditor = true
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            #if DEBUG
            print("[MealInputViewModel] Unexpected error: \(error)")
            #endif
            errorMessage = "画像分析に失敗しました: \(error.localizedDescription)"
        }

        isAnalyzing = false
    }

    private func pollForCompletion(id: String, maxAttempts: Int = Constants.maxPollingAttempts) async throws {
        for _ in 0..<maxAttempts {
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
                print("[MealInputViewModel] Unknown analysis status: \(status.status)")
                #endif
                throw APIError.serverError("分析ステータスが不明です: \(status.status)")
            }
        }

        throw APIError.serverError("分析がタイムアウトしました（\(Constants.pollingTimeoutSeconds)秒経過）")
    }


    func reset() {
        selectedImage = nil
        analysisResult = nil
        analysisId = nil
        errorMessage = nil
        isCompleted = false
        showEditor = false
    }

    func markCompleted() {
        isCompleted = true
    }
}
