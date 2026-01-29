import WidgetKit
import SwiftUI

struct WeightWidgetEntry: TimelineEntry {
    let date: Date
    let data: WidgetData
}

struct WeightWidgetProvider: TimelineProvider {
    func placeholder(in context: Context) -> WeightWidgetEntry {
        WeightWidgetEntry(date: Date(), data: .placeholder)
    }

    func getSnapshot(in context: Context, completion: @escaping (WeightWidgetEntry) -> Void) {
        Task {
            let data = await WidgetDataProvider.shared.getData()
            let entry = WeightWidgetEntry(date: Date(), data: data)
            completion(entry)
        }
    }

    func getTimeline(in context: Context, completion: @escaping (Timeline<WeightWidgetEntry>) -> Void) {
        Task {
            let data = await WidgetDataProvider.shared.getData()
            let entry = WeightWidgetEntry(date: Date(), data: data)

            // 1時間後に更新
            let nextUpdate = Calendar.current.date(byAdding: .hour, value: 1, to: Date()) ?? Date()
            let timeline = Timeline(entries: [entry], policy: .after(nextUpdate))
            completion(timeline)
        }
    }
}

struct WeightWidgetEntryView: View {
    var entry: WeightWidgetEntry
    @Environment(\.widgetFamily) var family

    var body: some View {
        switch family {
        case .systemSmall:
            SmallWeightView(data: entry.data)
        case .systemMedium:
            MediumWeightView(data: entry.data)
        default:
            SmallWeightView(data: entry.data)
        }
    }
}

private struct SmallWeightView: View {
    let data: WidgetData

    var body: some View {
        VStack(spacing: 4) {
            Image(systemName: "scalemass")
                .font(.title2)
                .foregroundStyle(.blue)

            if let weight = data.currentWeight {
                Text(String(format: "%.1f", weight))
                    .font(.system(size: 32, weight: .bold))
                Text("kg")
                    .font(.caption)
                    .foregroundStyle(.secondary)

                if let diff = data.weightDifference {
                    Text(diff > 0 ? String(format: "+%.1f", diff) : String(format: "%.1f", diff))
                        .font(.caption)
                        .foregroundStyle(diff > 0 ? .red : .green)
                }
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

private struct MediumWeightView: View {
    let data: WidgetData

    var body: some View {
        HStack {
            VStack(alignment: .leading, spacing: 4) {
                HStack {
                    Image(systemName: "scalemass")
                        .foregroundStyle(.blue)
                    Text("体重")
                        .font(.headline)
                }

                if let weight = data.currentWeight {
                    HStack(alignment: .lastTextBaseline, spacing: 2) {
                        Text(String(format: "%.1f", weight))
                            .font(.system(size: 36, weight: .bold))
                        Text("kg")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                    }
                } else {
                    Text("-")
                        .font(.system(size: 36, weight: .bold))
                }
            }

            Spacer()

            if let target = data.targetWeight {
                VStack(alignment: .trailing, spacing: 4) {
                    Text("目標")
                        .font(.caption)
                        .foregroundStyle(.secondary)

                    Text(String(format: "%.1f kg", target))
                        .font(.headline)

                    if let diff = data.weightDifference {
                        Text(diff > 0 ? String(format: "あと %.1f kg", diff) : "達成!")
                            .font(.caption)
                            .foregroundStyle(diff > 0 ? .orange : .green)
                    }
                }
            }
        }
        .padding()
        .containerBackground(.fill.tertiary, for: .widget)
    }
}

struct WeightWidget: Widget {
    let kind: String = "WeightWidget"

    var body: some WidgetConfiguration {
        StaticConfiguration(kind: kind, provider: WeightWidgetProvider()) { entry in
            WeightWidgetEntryView(entry: entry)
        }
        .configurationDisplayName("体重")
        .description("現在の体重と目標を表示します")
        .supportedFamilies([.systemSmall, .systemMedium])
    }
}

#Preview(as: .systemSmall) {
    WeightWidget()
} timeline: {
    WeightWidgetEntry(date: .now, data: .placeholder)
}

#Preview(as: .systemMedium) {
    WeightWidget()
} timeline: {
    WeightWidgetEntry(date: .now, data: .placeholder)
}
