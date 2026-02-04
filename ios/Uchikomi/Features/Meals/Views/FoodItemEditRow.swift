import SwiftUI

// MARK: - FoodItemEditRow

struct FoodItemEditRow: View {
    @Bindable var item: FoodEditItem
    let onDelete: () -> Void
    let onServingCountChanged: (Int) -> Void
    let onNameCommitted: (String) -> Void

    @State private var editingName: String = ""
    @FocusState private var isNameFocused: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("食材")
                    .font(.caption)
                    .foregroundStyle(.secondary)
                Spacer()
                if item.isEstimating {
                    ProgressView()
                        .scaleEffect(0.8)
                } else {
                    Button(action: onDelete) {
                        Image(systemName: "trash")
                            .foregroundStyle(.red)
                    }
                }
            }

            TextField("料理名（例：鶏むね肉、ご飯）", text: $editingName)
                .textFieldStyle(.roundedBorder)
                .focused($isNameFocused)
                .disabled(item.isEstimating)
                .onAppear {
                    editingName = item.name
                }
                .onChange(of: isNameFocused) { _, focused in
                    if !focused && editingName != item.originalName && !editingName.isEmpty {
                        onNameCommitted(editingName)
                    }
                }
                .onSubmit {
                    if editingName != item.originalName && !editingName.isEmpty {
                        onNameCommitted(editingName)
                    }
                }

            // 杯数変更UI
            HStack {
                Text("数量")
                    .font(.caption)
                    .foregroundStyle(.secondary)
                Spacer()
                Stepper(value: Binding(
                    get: { item.servingCount },
                    set: { onServingCountChanged($0) }
                ), in: 1...10) {
                    Text("\(item.servingCount) 人前")
                        .font(.subheadline)
                }
                .disabled(item.isEstimating)
            }

            TextField("量の説明（例：100g、1杯、大盛り）", text: $item.quantity)
                .textFieldStyle(.roundedBorder)
                .disabled(item.isEstimating)

            // 現在の栄養素を参考情報として表示（読み取り専用）
            if item.calories > 0 || item.isEstimating {
                HStack(spacing: 12) {
                    NutrientLabel(label: "kcal", value: item.calories, isLoading: item.isEstimating)
                    NutrientLabel(label: "P", value: item.protein, isLoading: item.isEstimating)
                    NutrientLabel(label: "F", value: item.fat, isLoading: item.isEstimating)
                    NutrientLabel(label: "C", value: item.carbohydrates, isLoading: item.isEstimating)
                }
                .padding(.top, 4)
            }
        }
        .padding()
        .background(Color(.secondarySystemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 10))
        .opacity(item.isEstimating ? 0.7 : 1.0)
    }
}

// MARK: - NutrientLabel

private struct NutrientLabel: View {
    let label: String
    let value: Double
    var isLoading: Bool = false

    var body: some View {
        VStack(spacing: 2) {
            Text(label)
                .font(.caption2)
                .foregroundStyle(.secondary)
            if isLoading {
                Text("--")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            } else {
                Text(String(format: "%.0f", value))
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
        .frame(maxWidth: .infinity)
    }
}

#Preview {
    FoodItemEditRow(
        item: FoodEditItem(
            name: "鶏むね肉",
            quantity: "100g",
            calories: 165,
            protein: 31,
            fat: 3.6,
            carbohydrates: 0
        ),
        onDelete: {},
        onServingCountChanged: { _ in },
        onNameCommitted: { _ in }
    )
    .padding()
}
