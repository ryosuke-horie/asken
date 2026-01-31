import SwiftUI

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

            TextField("食材名", text: $item.name)
                .textFieldStyle(.roundedBorder)

            TextField("量（例：100g、1個）", text: $item.quantity)
                .textFieldStyle(.roundedBorder)

            HStack(spacing: 8) {
                NutrientTextField(label: "kcal", value: $item.calories)
                NutrientTextField(label: "P", value: $item.protein)
                NutrientTextField(label: "F", value: $item.fat)
                NutrientTextField(label: "C", value: $item.carbohydrates)
            }
        }
        .padding()
        .background(Color(.secondarySystemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 10))
    }
}

private struct NutrientTextField: View {
    let label: String
    @Binding var value: Double

    @State private var text: String = ""

    var body: some View {
        VStack(spacing: 4) {
            Text(label)
                .font(.caption2)
                .foregroundStyle(.secondary)

            TextField("0", text: $text)
                .textFieldStyle(.roundedBorder)
                .keyboardType(.decimalPad)
                .multilineTextAlignment(.center)
                .frame(maxWidth: .infinity)
                .onChange(of: text) { _, newValue in
                    if let parsed = Double(newValue) {
                        value = parsed
                    }
                }
        }
        .onAppear {
            text = value > 0 ? String(format: "%.1f", value) : ""
        }
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
