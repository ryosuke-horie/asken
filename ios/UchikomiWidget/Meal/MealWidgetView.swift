import SwiftUI
import WidgetKit

// MARK: - MealWidgetView

struct MealWidgetView: View {
    let entry: MealEntry

    @Environment(\.widgetFamily) private var family

    var body: some View {
        switch family {
        case .systemSmall:
            smallView
        case .systemMedium:
            mediumView
        default:
            smallView
        }
    }

    // MARK: - Small Widget

    private var smallView: some View {
        VStack(alignment: .leading, spacing: 6) {
            mealTypeHeader

            switch entry.state {
            case .notLoggedIn:
                notLoggedInView
            case let .loaded(nutrition):
                caloriesSummaryView(calories: nutrition?.calories)
                recordButton
            }
        }
        .padding(12)
    }

    // MARK: - Medium Widget

    private var mediumView: some View {
        HStack(spacing: 16) {
            VStack(alignment: .leading, spacing: 6) {
                mealTypeHeader

                switch entry.state {
                case .notLoggedIn:
                    notLoggedInView
                case let .loaded(nutrition):
                    nutritionSummaryView(nutrition: nutrition)
                }
            }

            if case .loaded = entry.state {
                Spacer()
                VStack(alignment: .trailing, spacing: 8) {
                    foodDescriptionPreview
                    recordButton
                }
            }
        }
        .padding(14)
    }

    // MARK: - Components

    private var mealTypeHeader: some View {
        Label(entry.configuration.mealType.displayName, systemImage: entry.configuration.mealType.icon)
            .font(.caption2)
            .foregroundStyle(.secondary)
    }

    private func caloriesSummaryView(calories: Double?) -> some View {
        Group {
            if let calories {
                HStack(alignment: .lastTextBaseline, spacing: 2) {
                    Text("\(Int(calories))")
                        .font(.system(.title2, design: .rounded, weight: .bold))
                    Text("kcal")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            } else {
                Text("本日の記録なし")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
    }

    private func nutritionSummaryView(nutrition: NutritionSummary?) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            caloriesSummaryView(calories: nutrition?.calories)

            if let nutrition {
                HStack(spacing: 6) {
                    nutrientLabel(value: nutrition.protein, unit: "P", color: .blue)
                    nutrientLabel(value: nutrition.fat, unit: "F", color: .orange)
                    nutrientLabel(value: nutrition.carbs, unit: "C", color: .green)
                }
            }
        }
    }

    private func nutrientLabel(value: Double, unit: String, color: Color) -> some View {
        HStack(spacing: 1) {
            Text(unit)
                .font(.system(size: 9, weight: .semibold))
                .foregroundStyle(color)
            Text("\(Int(value))g")
                .font(.system(size: 10))
                .foregroundStyle(.secondary)
        }
    }

    private var foodDescriptionPreview: some View {
        Group {
            if !entry.configuration.foodDescription.isEmpty {
                Text(entry.configuration.foodDescription)
                    .font(.caption2)
                    .foregroundStyle(.secondary)
                    .lineLimit(2)
                    .multilineTextAlignment(.trailing)
                    .frame(maxWidth: 120, alignment: .trailing)
            }
        }
    }

    private var recordButton: some View {
        Button(
            intent: RecordMealIntent(
                mealType: entry.configuration.mealType,
                foodDescription: entry.configuration.foodDescription
            )
        ) {
            Label("記録する", systemImage: "plus.circle.fill")
                .font(.caption)
        }
        .buttonStyle(.borderedProminent)
        .tint(.green)
        .disabled(entry.configuration.foodDescription.isEmpty)
    }

    // MARK: - States

    private var notLoggedInView: some View {
        Text("ログインが必要です")
            .font(.caption2)
            .foregroundStyle(.secondary)
    }
}
