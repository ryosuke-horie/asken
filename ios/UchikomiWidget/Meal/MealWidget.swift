import SwiftUI
import WidgetKit

// MARK: - MealEntry

struct MealEntry: TimelineEntry {
    let date: Date
    let totalCalories: Double?
    let totalProtein: Double?
    let totalFat: Double?
    let totalCarbs: Double?
    let configuration: MealWidgetConfiguration
    let isLoggedIn: Bool
}

// MARK: - MealProvider

struct MealProvider: AppIntentTimelineProvider {
    typealias Entry = MealEntry
    typealias Intent = MealWidgetConfiguration

    func placeholder(in _: Context) -> MealEntry {
        MealEntry(
            date: Date(),
            totalCalories: 1_800,
            totalProtein: 80,
            totalFat: 60,
            totalCarbs: 220,
            configuration: MealWidgetConfiguration(),
            isLoggedIn: true
        )
    }

    func snapshot(for configuration: MealWidgetConfiguration, in _: Context) async -> MealEntry {
        MealEntry(
            date: Date(),
            totalCalories: 1_800,
            totalProtein: 80,
            totalFat: 60,
            totalCarbs: 220,
            configuration: configuration,
            isLoggedIn: true
        )
    }

    func timeline(for configuration: MealWidgetConfiguration, in _: Context) async -> Timeline<MealEntry> {
        guard SharedDefaults.authToken != nil else {
            let entry = MealEntry(
                date: Date(),
                totalCalories: nil,
                totalProtein: nil,
                totalFat: nil,
                totalCarbs: nil,
                configuration: configuration,
                isLoggedIn: false
            )
            return Timeline(entries: [entry], policy: .never)
        }

        let client = WidgetAPIClient()
        do {
            let dailyMeals = try await client.getDailyMeals(date: Date())
            let total = dailyMeals.dailyTotal
            let entry = MealEntry(
                date: Date(),
                totalCalories: total.totalCalories,
                totalProtein: total.totalProtein,
                totalFat: total.totalFat,
                totalCarbs: total.totalCarbohydrates,
                configuration: configuration,
                isLoggedIn: true
            )
            return Timeline(entries: [entry], policy: .never)
        } catch {
            let entry = MealEntry(
                date: Date(),
                totalCalories: nil,
                totalProtein: nil,
                totalFat: nil,
                totalCarbs: nil,
                configuration: configuration,
                isLoggedIn: true
            )
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
