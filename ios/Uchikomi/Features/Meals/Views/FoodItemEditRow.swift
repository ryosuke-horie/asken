import SwiftUI

// MARK: - FoodItemEditRow

struct FoodItemEditRow: View {
    @Bindable var item: FoodEditItem
    let onDelete: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("メニュー")
                    .font(.caption)
                    .foregroundStyle(.primary)
                    .opacity(0.6)
                Spacer()
                Button(action: onDelete) {
                    Image(systemName: "trash")
                        .foregroundStyle(.red)
                }
            }

            TextField("メニュー名（例：醤油ラーメン、カレーライス）", text: $item.name)
                .textFieldStyle(.roundedBorder)

            if item.hasNameChanged {
                Text("保存後に栄養素が再計算されます")
                    .font(.caption2)
                    .foregroundStyle(.orange)
            }

            HStack(spacing: 8) {
                TextField("数値", text: $item.quantityValue)
                    .textFieldStyle(.roundedBorder)
                    .keyboardType(.decimalPad)
                    .onChange(of: item.quantityValue) {
                        item.updateQuantityString()
                        item.recalculateNutrition()
                    }

                Picker("単位", selection: $item.quantityUnit) {
                    Text("選択").tag(nil as MeasurementUnit?)
                    ForEach(MeasurementUnit.allCases) { unit in
                        Text(unit.displayName).tag(unit as MeasurementUnit?)
                    }
                }
                .pickerStyle(.menu)
                .onChange(of: item.quantityUnit) { _, _ in
                    item.updateQuantityString()
                }
            }

            if item.hasUnitChanged {
                Text("保存後に栄養素が再計算されます")
                    .font(.caption2)
                    .foregroundStyle(.orange)
            }

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
                .fontWeight(.medium)
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
