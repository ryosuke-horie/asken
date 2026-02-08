import Foundation
import os

private let logger = Logger(subsystem: Bundle.main.bundleIdentifier ?? "Uchikomi", category: "WeightInputViewModel")

@Observable
final class WeightInputViewModel {
    var weightText: String = ""
    var memo: String = ""
    var isSaving = false
    var errorMessage: String?
    var didSave = false

    let editingRecord: WeightRecord?
    private let repository: WeightRepositoryProtocol

    init(
        editingRecord: WeightRecord? = nil,
        repository: WeightRepositoryProtocol = WeightRepository()
    ) {
        self.editingRecord = editingRecord
        self.repository = repository

        if let record = editingRecord {
            weightText = formatWeight(record.weightKg)
            memo = record.note ?? ""
        }
    }

    var isEditing: Bool {
        editingRecord != nil
    }

    var weightValue: Double? {
        Double(weightText)
    }

    var isValid: Bool {
        guard let weight = weightValue else { return false }
        return weight >= 20.0 && weight <= 300.0
    }

    func incrementWeight() {
        guard let current = weightValue else { return }
        let newValue = (current * 10 + 1).rounded() / 10
        if newValue <= 300.0 {
            weightText = formatWeight(newValue)
        }
    }

    func decrementWeight() {
        guard let current = weightValue else { return }
        let newValue = (current * 10 - 1).rounded() / 10
        if newValue >= 20.0 {
            weightText = formatWeight(newValue)
        }
    }

    func save() async {
        guard let weight = weightValue, isValid else {
            errorMessage = "体重は20.0〜300.0kgの範囲で入力してください"
            return
        }

        isSaving = true
        errorMessage = nil

        do {
            if let record = editingRecord {
                _ = try await repository.updateRecord(id: record.id, weightKg: weight, note: memo)
            } else {
                _ = try await repository.createRecord(weightKg: weight, recordedAt: Date(), note: memo)
            }
            didSave = true
        } catch let error as APIError {
            logger.error("体重記録保存でAPIエラー: \(error.localizedDescription)")
            errorMessage = error.localizedDescription
        } catch {
            logger.error("体重記録保存で予期しないエラー: \(error.localizedDescription)")
            errorMessage = "保存に失敗しました"
        }

        isSaving = false
    }

    func delete() async {
        guard let record = editingRecord else { return }

        isSaving = true
        errorMessage = nil

        do {
            try await repository.deleteRecord(id: record.id)
            didSave = true
        } catch let error as APIError {
            logger.error("体重記録削除でAPIエラー: \(error.localizedDescription)")
            errorMessage = error.localizedDescription
        } catch {
            logger.error("体重記録削除で予期しないエラー: \(error.localizedDescription)")
            errorMessage = "削除に失敗しました"
        }

        isSaving = false
    }

    func setQuickNote(_ note: String) {
        memo = note
    }

    private func formatWeight(_ value: Double) -> String {
        String(format: "%.1f", value)
    }
}
