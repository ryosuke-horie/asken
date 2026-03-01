import OSLog
import SwiftUI
import WidgetKit

private let logger = Logger(subsystem: "dev.exe.uchikomi.widget", category: "WeightHistoryWidget")

// MARK: - DailyWeightData

struct DailyWeightData: Identifiable {
    let id: String
    let date: Date
    let hasRecord: Bool
    let latestWeightKg: Double?
}

// MARK: - WeightHistoryEntry

struct WeightHistoryEntry: TimelineEntry {
    enum State {
        case notLoggedIn
        case noData
        case loaded(weeklyData: [DailyWeightData])
    }

    let date: Date
    let state: State
}

// MARK: - WeightHistoryProvider

struct WeightHistoryProvider: TimelineProvider {
    func placeholder(in _: Context) -> WeightHistoryEntry {
        WeightHistoryEntry(date: Date(), state: .loaded(weeklyData: Self.sampleWeeklyData()))
    }

    func getSnapshot(in _: Context, completion: @escaping (WeightHistoryEntry) -> Void) {
        completion(WeightHistoryEntry(date: Date(), state: .loaded(weeklyData: Self.sampleWeeklyData())))
    }

    func getTimeline(in _: Context, completion: @escaping (Timeline<WeightHistoryEntry>) -> Void) {
        guard SharedDefaults.authToken != nil else {
            let entry = WeightHistoryEntry(date: Date(), state: .notLoggedIn)
            completion(Timeline(entries: [entry], policy: .never))
            return
        }

        Task {
            let client = WidgetAPIClient()
            do {
                let response = try await client.getWeeklyWeightRecords()
                let weeklyData = Self.buildWeeklyData(from: response.records)
                let entry = WeightHistoryEntry(date: Date(), state: .loaded(weeklyData: weeklyData))
                completion(Timeline(entries: [entry], policy: .never))
            } catch WidgetAPIError.unauthorized {
                SharedDefaults.clearAuthToken()
                let entry = WeightHistoryEntry(date: Date(), state: .notLoggedIn)
                completion(Timeline(entries: [entry], policy: .never))
            } catch {
                logger.error("体重履歴データ取得失敗: \(error.localizedDescription)")
                let entry = WeightHistoryEntry(date: Date(), state: .noData)
                completion(Timeline(entries: [entry], policy: .never))
            }
        }
    }

    // MARK: - Private Helpers

    private static let iso8601Formatter: ISO8601DateFormatter = {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return formatter
    }()

    private static let fallbackFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd'T'HH:mm:ssZ"
        formatter.locale = Locale(identifier: "en_US_POSIX")
        return formatter
    }()

    private static func parseRecordedAt(_ string: String) -> Date? {
        iso8601Formatter.date(from: string) ?? fallbackFormatter.date(from: string)
    }

    private static func buildWeeklyData(from records: [WidgetWeightRecord]) -> [DailyWeightData] {
        let calendar = Calendar.current
        let today = calendar.startOfDay(for: Date())

        // 日付ごとに記録をグループ化
        var recordsByDay: [Date: [WidgetWeightRecord]] = [:]
        for record in records {
            guard let recordDate = parseRecordedAt(record.recordedAt) else { continue }
            let dayStart = calendar.startOfDay(for: recordDate)
            recordsByDay[dayStart, default: []].append(record)
        }

        // 過去7日間（今日を含む）のDailyWeightDataを生成
        return (0 ..< 7).reversed().compactMap { offset -> DailyWeightData? in
            guard let dayStart = calendar.date(byAdding: .day, value: -offset, to: today) else {
                return nil
            }
            let dayRecords = recordsByDay[dayStart] ?? []
            let latestWeight = dayRecords
                .sorted { $0.recordedAt < $1.recordedAt }
                .last?.weightKg
            return DailyWeightData(
                id: "\(offset)",
                date: dayStart,
                hasRecord: !dayRecords.isEmpty,
                latestWeightKg: latestWeight
            )
        }
    }

    private static func sampleWeeklyData() -> [DailyWeightData] {
        let calendar = Calendar.current
        let today = calendar.startOfDay(for: Date())
        let weights: [Double?] = [65.2, 65.0, nil, 64.8, 64.9, nil, 64.7]
        return (0 ..< 7).reversed().enumerated().compactMap { index, offset -> DailyWeightData? in
            guard let dayStart = calendar.date(byAdding: .day, value: -offset, to: today) else {
                return nil
            }
            return DailyWeightData(
                id: "\(offset)",
                date: dayStart,
                hasRecord: weights[index] != nil,
                latestWeightKg: weights[index]
            )
        }
    }
}

// MARK: - WeightHistoryWidget

struct WeightHistoryWidget: Widget {
    let kind = "WeightHistoryWidget"

    var body: some WidgetConfiguration {
        StaticConfiguration(kind: kind, provider: WeightHistoryProvider()) { entry in
            WeightHistoryWidgetView(entry: entry)
                .containerBackground(.fill.tertiary, for: .widget)
        }
        .configurationDisplayName("体重履歴")
        .description("直近1週間の体重推移と記録状況を確認")
        .supportedFamilies([.systemSmall])
    }
}
