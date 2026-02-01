import SwiftUI

// MARK: - FoodItemEditRow

struct FoodItemEditRow: View {
    @Bindable var item: FoodEditItem
    let onDelete: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("食材")
                    .font(.caption)
                    .foregroundStyle(.secondary)
                Spacer()
                Button(action: onDelete) {
                    Image(systemName: "trash")
                        .foregroundStyle(.red)
                }
            }

            TextField("料理名（例：鶏むね肉、ご飯）", text: $item.name)
                .textFieldStyle(.roundedBorder)

            TextField("量（例：100g、1杯、大盛り）", text: $item.quantity)
                .textFieldStyle(.roundedBorder)

            // 現在の栄養素を参考情報として表示（読み取り専用）
            if item.calories > 0 {
                HStack(spacing: 12) {
                    NutrientLabel(label: "kcal", value: item.calories)
                    NutrientLabel(label: "P", value: item.protein)
                    NutrientLabel(label: "F", value: item.fat)
                    NutrientLabel(label: "C", value: item.carbohydrates)
                }
                .padding(.top, 4)
            }
        }
        .padding()
        .background(Color(.secondarySystemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 10))
    }
}

// MARK: - NutrientLabel

private struct NutrientLabel: View {
    let label: String
    let value: Double

    var body: some View {
        VStack(spacing: 2) {
            Text(label)
                .font(.caption2)
                .foregroundStyle(.secondary)
            Text(String(format: "%.0f", value))
                .font(.caption)
                .foregroundStyle(.secondary)
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
        onDelete: {}
    )
    .padding()
}
