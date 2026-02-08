import SwiftUI

struct WeightGoalSheet: View {
    @Environment(\.dismiss) private var dismiss
    @State private var targetWeightText: String
    @State private var isSaving = false
    @State private var errorMessage: String?

    let currentGoal: WeightGoal?
    let repository: WeightRepositoryProtocol
    let onSaved: () -> Void

    init(
        currentGoal: WeightGoal?,
        repository: WeightRepositoryProtocol = WeightRepository(),
        onSaved: @escaping () -> Void
    ) {
        self.currentGoal = currentGoal
        self.repository = repository
        self.onSaved = onSaved
        _targetWeightText = State(
            initialValue: currentGoal.map { String(format: "%.1f", $0.targetWeightKg) } ?? ""
        )
    }

    private var isValid: Bool {
        guard let weight = Double(targetWeightText) else { return false }
        return weight >= 20.0 && weight <= 300.0
    }

    var body: some View {
        NavigationStack {
            Form {
                Section("目標体重") {
                    HStack {
                        TextField("63.0", text: $targetWeightText)
                            .keyboardType(.decimalPad)
                            .font(.title2)

                        Text("kg")
                            .foregroundStyle(.secondary)
                    }
                }

                if let errorMessage {
                    Section {
                        Text(errorMessage)
                            .foregroundStyle(.red)
                            .font(.caption)
                    }
                }
            }
            .navigationTitle("目標体重の設定")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("キャンセル") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("保存") {
                        Task { await save() }
                    }
                    .disabled(!isValid || isSaving)
                }
            }
        }
    }

    private func save() async {
        guard let weight = Double(targetWeightText) else { return }

        isSaving = true
        errorMessage = nil

        do {
            _ = try await repository.setGoal(targetWeightKg: weight)
            isSaving = false
            onSaved()
            dismiss()
            return
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "保存に失敗しました"
        }

        isSaving = false
    }
}
