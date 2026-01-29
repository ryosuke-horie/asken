import WidgetKit
import SwiftUI

struct CaloriesWidgetEntry: TimelineEntry {
    let date: Date
    let data: WidgetData
}

struct CaloriesWidgetProvider: TimelineProvider {
    func placeholder(in context: Context) -> CaloriesWidgetEntry {
        CaloriesWidgetEntry(date: Date(), data: .placeholder)
    }

    func getSnapshot(in context: Context, completion: @escaping (CaloriesWidgetEntry) -> Void) {
        Task {
            let data = await WidgetDataProvider.shared.getData()
            let entry = CaloriesWidgetEntry(date: Date(), data: data)
            completion(entry)
        }
    }

    func getTimeline(in context: Context, completion: @escaping (Timeline<CaloriesWidgetEntry>) -> Void) {
        Task {
            let data = await WidgetDataProvider.shared.getData()
            let entry = CaloriesWidgetEntry(date: Date(), data: data)

            // 1時間後に更新
            let nextUpdate = Calendar.current.date(byAdding: .hour, value: 1, to: Date()) ?? Date()
            let timeline = Timeline(entries: [entry], policy: .after(nextUpdate))
            completion(timeline)
        }
    }
}

struct CaloriesWidgetEntryView: View {
    var entry: CaloriesWidgetEntry
    @Environment(\.widgetFamily) var family

    var body: some View {
        switch family {
        case .systemSmall:
            SmallCaloriesView(data: entry.data)
        case .systemMedium:
            MediumCaloriesView(data: entry.data)
        default:
            SmallCaloriesView(data: entry.data)
        }
    }
}

private struct SmallCaloriesView: View {
    let data: WidgetData

    var body: some View {
        VStack(spacing: 4) {
            Image(systemName: "flame")
                .font(.title2)
                .foregroundStyle(.orange)

            if let calories = data.todayCalories {
                Text("\(Int(calories))")
                    .font(.system(size: 32, weight: .bold))
                Text("kcal")
                    .font(.caption)
                    .foregroundStyle(.secondary)
                Text("今日の摂取")
                    .font(.caption2)
                    .foregroundStyle(.secondary)
            } else {
                Text("-")
                    .font(.system(size: 32, weight: .bold))
                Text("未記録")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
        .containerBackground(.fill.tertiary, for: .widget)
    }
}

private struct MediumCaloriesView: View {
    let data: WidgetData

    var body: some View {
        HStack {
            VStack(alignment: .leading, spacing: 4) {
                HStack {
                    Image(systemName: "flame")
                        .foregroundStyle(.orange)
                    Text("今日のカロリー")
                        .font(.headline)
                }

                if let calories = data.todayCalories {
                    HStack(alignment: .lastTextBaseline, spacing: 2) {
                        Text("\(Int(calories))")
                            .font(.system(size: 36, weight: .bold))
                        Text("kcal")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                    }
                } else {
                    Text("-")
                        .font(.system(size: 36, weight: .bold))
                }
            }

            Spacer()

            // 簡易的なプログレスリング（目標2000kcalと仮定）
            if let calories = data.todayCalories {
                let target = 2000.0
                let progress = min(calories / target, 1.0)

                ZStack {
                    Circle()
                        .stroke(Color.gray.opacity(0.3), lineWidth: 8)

                    Circle()
                        .trim(from: 0, to: progress)
                        .stroke(
                            progress > 1.0 ? Color.red : Color.orange,
                            style: StrokeStyle(lineWidth: 8, lineCap: .round)
                        )
                        .rotationEffect(.degrees(-90))

                    VStack(spacing: 0) {
                        Text("\(Int(progress * 100))%")
                            .font(.caption)
                            .fontWeight(.semibold)
                    }
                }
                .frame(width: 60, height: 60)
            }
        }
        .padding()
        .containerBackground(.fill.tertiary, for: .widget)
    }
}

struct CaloriesWidget: Widget {
    let kind: String = "CaloriesWidget"

    var body: some WidgetConfiguration {
        StaticConfiguration(kind: kind, provider: CaloriesWidgetProvider()) { entry in
            CaloriesWidgetEntryView(entry: entry)
        }
        .configurationDisplayName("カロリー")
        .description("今日の摂取カロリーを表示します")
        .supportedFamilies([.systemSmall, .systemMedium])
    }
}

#Preview(as: .systemSmall) {
    CaloriesWidget()
} timeline: {
    CaloriesWidgetEntry(date: .now, data: .placeholder)
}

#Preview(as: .systemMedium) {
    CaloriesWidget()
} timeline: {
    CaloriesWidgetEntry(date: .now, data: .placeholder)
}
