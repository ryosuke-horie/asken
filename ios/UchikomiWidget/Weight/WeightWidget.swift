import OSLog
import SwiftUI
import WidgetKit

private let logger = Logger(subsystem: "dev.exe.uchikomi.widget", category: "WeightWidget")

// MARK: - WeightEntry

struct WeightEntry: TimelineEntry {
    enum State {
        case notLoggedIn
        case noRecord(targetWeightKg: Double?)
        case loaded(weightKg: Double, targetWeightKg: Double?)
    }

    let date: Date
    let state: State
}

// MARK: - WeightProvider

struct WeightProvider: TimelineProvider {
    func placeholder(in _: Context) -> WeightEntry {
        WeightEntry(date: Date(), state: .loaded(weightKg: 65.0, targetWeightKg: 63.0))
    }

    func getSnapshot(in _: Context, completion: @escaping (WeightEntry) -> Void) {
        completion(WeightEntry(date: Date(), state: .loaded(weightKg: 65.0, targetWeightKg: 63.0)))
    }

    func getTimeline(in _: Context, completion: @escaping (Timeline<WeightEntry>) -> Void) {
        guard SharedDefaults.authToken != nil else {
            let entry = WeightEntry(date: Date(), state: .notLoggedIn)
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

                let state: WeightEntry.State = if let weight = latestWeight {
                    .loaded(weightKg: weight, targetWeightKg: targetWeight)
                } else {
                    .noRecord(targetWeightKg: targetWeight)
                }
                let entry = WeightEntry(date: Date(), state: state)
                // .never: 自動バックグラウンド更新を行わない。
                // Cloud Run のコールドスタートによるレイテンシとリクエスト費用を抑えるため、
                // 更新は記録後（RecordWeightIntent）とアプリ起動時（UchikomiApp）に限定する。
                completion(Timeline(entries: [entry], policy: .never))
            } catch {
                logger.error("体重データ取得失敗: \(error.localizedDescription)")
                let state: WeightEntry.State = if let cached = SharedDefaults.latestWeightKg {
                    .loaded(weightKg: cached, targetWeightKg: nil)
                } else {
                    .noRecord(targetWeightKg: nil)
                }
                let entry = WeightEntry(date: Date(), state: state)
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
