import OSLog
import SwiftUI
import WidgetKit

private let logger = Logger(subsystem: "dev.exe.uchikomi.widget", category: "MealWidget")

// MARK: - NutritionSummary

struct NutritionSummary {
    let calories: Double
    let protein: Double
    let fat: Double
    let carbs: Double
}

// MARK: - MealEntry

struct MealEntry: TimelineEntry {
    enum State {
        case notLoggedIn
        case loaded(NutritionSummary?)
    }

    let date: Date
    let state: State
    let configuration: MealWidgetConfiguration
}

// MARK: - MealProvider

struct MealProvider: AppIntentTimelineProvider {
    typealias Entry = MealEntry
    typealias Intent = MealWidgetConfiguration

    func placeholder(in _: Context) -> MealEntry {
        let nutrition = NutritionSummary(calories: 1_800, protein: 80, fat: 60, carbs: 220)
        return MealEntry(date: Date(), state: .loaded(nutrition), configuration: MealWidgetConfiguration())
    }

    func snapshot(for configuration: MealWidgetConfiguration, in _: Context) async -> MealEntry {
        let nutrition = NutritionSummary(calories: 1_800, protein: 80, fat: 60, carbs: 220)
        return MealEntry(date: Date(), state: .loaded(nutrition), configuration: configuration)
    }

    func timeline(for configuration: MealWidgetConfiguration, in _: Context) async -> Timeline<MealEntry> {
        guard SharedDefaults.authToken != nil else {
            let entry = MealEntry(date: Date(), state: .notLoggedIn, configuration: configuration)
            return Timeline(entries: [entry], policy: .never)
        }

        let client = WidgetAPIClient()
        do {
            let dailyMeals = try await client.getDailyMeals(date: Date())
            let total = dailyMeals.dailyTotal
            let nutrition = NutritionSummary(
                calories: total.totalCalories,
                protein: total.totalProtein,
                fat: total.totalFat,
                carbs: total.totalCarbohydrates
            )
            let entry = MealEntry(date: Date(), state: .loaded(nutrition), configuration: configuration)
            // .never: 自動バックグラウンド更新を行わない。
            // Cloud Run のコールドスタートによるレイテンシとリクエスト費用を抑えるため、
            // 更新は記録後（RecordMealIntent）とアプリ起動時（UchikomiApp）に限定する。
            return Timeline(entries: [entry], policy: .never)
        } catch {
            logger.error("食事データ取得失敗: \(error.localizedDescription)")
            let entry = MealEntry(date: Date(), state: .loaded(nil), configuration: configuration)
            return Timeline(entries: [entry], policy: .never)
        }
    }
}

// MARK: - MealWidget

struct MealWidget: Widget {
    let kind = "MealWidget"

    var body: some WidgetConfiguration {
        AppIntentConfiguration(kind: kind, intent: MealWidgetConfiguration.self, provider: MealProvider()) { entry in
            MealWidgetView(entry: entry)
                .containerBackground(.fill.tertiary, for: .widget)
        }
        .configurationDisplayName("食事記録")
        .description("事前設定した食事をタップ一発で Gemini 分析・記録")
        .supportedFamilies([.systemSmall, .systemMedium])
    }
}
