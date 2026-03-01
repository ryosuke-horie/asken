import SwiftUI
import WidgetKit

// MARK: - WeightEntry

struct WeightEntry: TimelineEntry {
    let date: Date
    let latestWeightKg: Double?
    let targetWeightKg: Double?
    let isLoggedIn: Bool
}

// MARK: - WeightProvider

struct WeightProvider: TimelineProvider {
    func placeholder(in _: Context) -> WeightEntry {
        WeightEntry(date: Date(), latestWeightKg: 65.0, targetWeightKg: 63.0, isLoggedIn: true)
    }

    func getSnapshot(in _: Context, completion: @escaping (WeightEntry) -> Void) {
        completion(WeightEntry(date: Date(), latestWeightKg: 65.0, targetWeightKg: 63.0, isLoggedIn: true))
    }

    func getTimeline(in _: Context, completion: @escaping (Timeline<WeightEntry>) -> Void) {
        guard SharedDefaults.authToken != nil else {
            let entry = WeightEntry(date: Date(), latestWeightKg: nil, targetWeightKg: nil, isLoggedIn: false)
            completion(Timeline(entries: [entry], policy: .never))
            return
        }

        // Task キャンセル時は WidgetKit がタイムライン更新を再スケジュールするため completion 未呼び出しは許容される
        Task {
            let client = WidgetAPIClient()
            do {
                let response = try await client.getLatestWeightRecord()
                let latestWeight = response.records.first?.weightKg
                let targetWeight = response.goal?.targetWeightKg

                if let latestWeight {
                    SharedDefaults.latestWeightKg = latestWeight
                }

                let entry = WeightEntry(
                    date: Date(),
                    latestWeightKg: latestWeight,
                    targetWeightKg: targetWeight,
                    isLoggedIn: true
                )
                completion(Timeline(entries: [entry], policy: .never))
            } catch {
                let entry = WeightEntry(
                    date: Date(),
                    latestWeightKg: SharedDefaults.latestWeightKg,
                    targetWeightKg: nil,
                    isLoggedIn: true
                )
                completion(Timeline(entries: [entry], policy: .never))
            }
        }
    }
}

// MARK: - WeightWidget

struct WeightWidget: Widget {
    let kind = "WeightWidget"

    var body: some WidgetConfiguration {
        StaticConfiguration(kind: kind, provider: WeightProvider()) { entry in
            WeightWidgetView(entry: entry)
                .containerBackground(.fill.tertiary, for: .widget)
        }
        .configurationDisplayName("体重")
        .description("最新の体重を確認し、ワンタップで記録")
        .supportedFamilies([.systemSmall, .systemMedium])
    }
}
