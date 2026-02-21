import SwiftUI

// MARK: - IngredientEditView

struct IngredientEditView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var viewModel: IngredientEditViewModel

    let onSaved: (Ingredient) -> Void

    init(ingredient: Ingredient?, onSaved: @escaping (Ingredient) -> Void) {
        _viewModel = State(initialValue: IngredientEditViewModel(ingredient: ingredient))
        self.onSaved = onSaved
    }

    var body: some View {
        NavigationStack {
            Form {
                Section("基本情報") {
                    TextField("食材名（必須）", text: $viewModel.name)

                    Picker("カテゴリ", selection: $viewModel.category) {
                        ForEach(IngredientCategory.allCases) { category in
                            Text(category.displayName).tag(category)
                        }
                    }
                }

                Section("数量") {
                    HStack {
                        TextField("数量", value: $viewModel.quantity, format: .number)
                            .keyboardType(.decimalPad)

                        Picker("単位", selection: $viewModel.unit) {
                            ForEach(IngredientUnit.allCases) { unit in
                                Text(unit.rawValue).tag(unit.rawValue)
                            }
                        }
                        .pickerStyle(.menu)
                    }
                }

                Section("日付") {
                    Toggle("購入日を設定", isOn: $viewModel.hasPurchaseDate)
                    if viewModel.hasPurchaseDate {
                        DatePicker(
                            "購入日",
                            selection: $viewModel.purchaseDate,
                            displayedComponents: .date
                        )
                    }

                    Toggle("消費期限を設定", isOn: $viewModel.hasExpiryDate)
                    if viewModel.hasExpiryDate {
                        DatePicker(
                            "消費期限",
                            selection: $viewModel.expiryDate,
                            displayedComponents: .date
                        )
                    }
                }

                if let error = viewModel.errorMessage {
                    Section {
                        Text(error)
                            .foregroundStyle(.red)
                            .font(.caption)
                    }
                }
            }
            .navigationTitle(viewModel.isEditing ? "食材を編集" : "食材を追加")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("キャンセル") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("保存") {
                        Task {
                            if let saved = await viewModel.save() {
                                onSaved(saved)
                                dismiss()
                            }
                        }
                    }
                    .disabled(viewModel.isSaveDisabled || viewModel.isSaving)
                }
            }
            .overlay {
                if viewModel.isSaving {
                    ProgressView()
                }
            }
        }
    }
}

// MARK: - IngredientEditViewModel

@Observable
final class IngredientEditViewModel {
    var name: String
    var category: IngredientCategory
    var quantity: Double
    var unit: String
    var hasPurchaseDate: Bool
    var purchaseDate: Date
    var hasExpiryDate: Bool
    var expiryDate: Date
    var isSaving = false
    var errorMessage: String?

    let ingredient: Ingredient?
    let repository: IngredientRepositoryProtocol

    var isEditing: Bool {
        ingredient != nil
    }

    var isSaveDisabled: Bool {
        name.trimmingCharacters(in: .whitespaces).isEmpty || quantity <= 0
    }

    init(ingredient: Ingredient?, repository: IngredientRepositoryProtocol = IngredientRepository()) {
        self.ingredient = ingredient
        self.repository = repository

        if let ingredient {
            name = ingredient.name
            category = ingredient.category
            quantity = ingredient.quantity
            unit = ingredient.unit
            hasPurchaseDate = ingredient.purchaseDate != nil
            purchaseDate = ingredient.purchaseDate ?? Date()
            hasExpiryDate = ingredient.expiryDate != nil
            expiryDate = ingredient.expiryDate ?? Date()
        } else {
            name = ""
            category = .other
            quantity = 1
            unit = IngredientUnit.piece.rawValue
            hasPurchaseDate = false
            purchaseDate = Date()
            hasExpiryDate = false
            expiryDate = Date()
        }
    }

    func save() async -> Ingredient? {
        isSaving = true
        errorMessage = nil
        defer { isSaving = false }

        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"

        let request = CreateIngredientRequest(
            name: name.trimmingCharacters(in: .whitespaces),
            category: category.rawValue,
            quantity: quantity,
            unit: unit,
            purchaseDate: hasPurchaseDate ? dateFormatter.string(from: purchaseDate) : nil,
            expiryDate: hasExpiryDate ? dateFormatter.string(from: expiryDate) : nil,
            source: IngredientSource.manual.rawValue
        )

        do {
            if let existing = ingredient {
                return try await repository.updateIngredient(id: existing.id, request: request)
            } else {
                return try await repository.createIngredient(request)
            }
        } catch let error as APIError {
            errorMessage = error.localizedDescription
            return nil
        } catch {
            errorMessage = "保存に失敗しました"
            return nil
        }
    }
}
