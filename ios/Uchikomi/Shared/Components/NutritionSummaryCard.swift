import SwiftUI

struct NutritionSummaryCard: View {
    let calories: Double
    let protein: Double
    let fat: Double
    let carbohydrates: Double

    var body: some View {
        VStack(spacing: 12) {
            // Calories (main)
            HStack {
                Text("カロリー")
                    .font(.headline)
                Spacer()
                Text("\(Int(calories))")
                    .font(.title)
                    .fontWeight(.bold)
                Text("kcal")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }

            Divider()

            // Macros
            HStack(spacing: 16) {
                MacroItem(
                    name: "たんぱく質",
                    value: protein,
                    unit: "g",
                    color: .red
                )

                MacroItem(
                    name: "脂質",
                    value: fat,
                    unit: "g",
                    color: .yellow
                )

                MacroItem(
                    name: "炭水化物",
                    value: carbohydrates,
                    unit: "g",
                    color: .blue
                )
            }
        }
        .padding()
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .shadow(color: .black.opacity(0.1), radius: 4, x: 0, y: 2)
    }
}

private struct MacroItem: View {
    let name: String
    let value: Double
    let unit: String
    let color: Color

    var body: some View {
        VStack(spacing: 4) {
            Circle()
                .fill(color)
                .frame(width: 8, height: 8)

            Text(name)
                .font(.caption)
                .foregroundStyle(.secondary)

            HStack(alignment: .lastTextBaseline, spacing: 2) {
                Text(String(format: "%.1f", value))
                    .font(.subheadline)
                    .fontWeight(.semibold)
                Text(unit)
                    .font(.caption2)
                    .foregroundStyle(.secondary)
            }
        }
        .frame(maxWidth: .infinity)
    }
}

#Preview {
    NutritionSummaryCard(
        calories: 650,
        protein: 25.5,
        fat: 22.3,
        carbohydrates: 78.0
    )
    .padding()
    .background(Color(.systemGroupedBackground))
}
