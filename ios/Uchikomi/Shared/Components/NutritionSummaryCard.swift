import SwiftUI

// MARK: - NutritionSummaryCard

struct NutritionSummaryCard: View {
    let calories: Double
    let protein: Double
    let fat: Double
    let carbohydrates: Double

    var goal: NutritionGoal?

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
                    .foregroundStyle(Theme.primary)
                Text("kcal")
                    .font(.subheadline)
                    .foregroundStyle(.primary)
                    .opacity(0.5)
            }

            // カロリープログレスバー（目標がある場合）
            if let goal {
                CalorieProgressView(current: calories, goal: goal.calories)
            }

            Divider()

            // PFC円グラフとプログレスバー
            if let goal {
                VStack(spacing: 16) {
                    // 円グラフ
                    PFCPieChart(
                        protein: protein,
                        fat: fat,
                        carbohydrates: carbohydrates
                    )

                    // PFCプログレスバー
                    PFCProgressRow(
                        currentProtein: protein,
                        currentFat: fat,
                        currentCarbs: carbohydrates,
                        goalProtein: goal.protein,
                        goalFat: goal.fat,
                        goalCarbs: goal.carbohydrates
                    )
                }
            } else {
                // 目標がない場合は従来の表示
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
        }
        .padding()
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .shadow(color: .black.opacity(0.1), radius: 4, x: 0, y: 2)
    }
}

// MARK: - MacroItem

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
                .foregroundStyle(.primary)
                .opacity(0.6)

            HStack(alignment: .lastTextBaseline, spacing: 2) {
                Text(String(format: "%.1f", value))
                    .font(.subheadline)
                    .fontWeight(.semibold)
                    .foregroundStyle(.primary)
                Text(unit)
                    .font(.caption2)
                    .foregroundStyle(.primary)
                    .opacity(0.5)
            }
        }
        .frame(maxWidth: .infinity)
    }
}

#Preview("目標なし") {
    NutritionSummaryCard(
        calories: 650,
        protein: 25.5,
        fat: 22.3,
        carbohydrates: 78.0
    )
    .padding()
    .background(Color(.systemGroupedBackground))
}

#Preview("目標あり") {
    NutritionSummaryCard(
        calories: 1_450,
        protein: 85.5,
        fat: 52.3,
        carbohydrates: 178.0,
        goal: NutritionGoal(
            calories: 2_000,
            protein: 100,
            fat: 60,
            carbohydrates: 250
        )
    )
    .padding()
    .background(Color(.systemGroupedBackground))
}
