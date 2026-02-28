import Foundation
import os

private let logger = Logger(subsystem: Bundle.main.bundleIdentifier ?? "Uchikomi", category: "ExerciseInputViewModel")

// MARK: - ExerciseInputViewModel

@Observable
final class ExerciseInputViewModel {
    var exerciseName: String = ""
    var durationText: String = "60"
    var isSaving = false
    var errorMessage: String?
    var didSave = false

    let recordedDate: String
    private let repository: ExerciseRepositoryProtocol

    static let minDuration = 5
    static let maxDuration = 600
    static let durationStep = 5

    static let quickPickNames: [String] = [
        "柔術", "柔道", "レスリング", "MMA", "ボクシング", "キックボクシング",
        "ランニング", "筋トレ",
    ]

    init(
        recordedDate: String,
        repository: ExerciseRepositoryProtocol = ExerciseRepository()
    ) {
        self.recordedDate = recordedDate
        self.repository = repository
    }

    var durationValue: Int? {
        guard let value = Int(durationText),
              value >= Self.minDuration,
              value <= Self.maxDuration else { return nil }
        return value
    }

    var isValid: Bool {
        !exerciseName.trimmingCharacters(in: .whitespaces).isEmpty && durationValue != nil
    }

    func incrementDuration() {
        let current = Int(durationText) ?? Self.minDuration
        let newValue = min(current + Self.durationStep, Self.maxDuration)
        durationText = String(newValue)
    }

    func decrementDuration() {
        let current = Int(durationText) ?? Self.minDuration
        let newValue = max(current - Self.durationStep, Self.minDuration)
        durationText = String(newValue)
    }

    func save() async {
        guard isValid, let duration = durationValue else {
            errorMessage = "種目名と時間（\(Self.minDuration)〜\(Self.maxDuration)分）を入力してください"
            return
        }

        isSaving = true
        errorMessage = nil

        do {
            _ = try await repository.createRecord(
                exerciseName: exerciseName.trimmingCharacters(in: .whitespaces),
                durationMinutes: duration,
                recordedDate: recordedDate
            )
            didSave = true
        } catch let error as APIError {
            logger.error("運動記録保存でAPIエラー: \(error.localizedDescription)")
            errorMessage = error.localizedDescription
        } catch {
            logger.error("運動記録保存で予期しないエラー: \(error.localizedDescription)")
            errorMessage = "保存に失敗しました"
        }

        isSaving = false
    }
}
