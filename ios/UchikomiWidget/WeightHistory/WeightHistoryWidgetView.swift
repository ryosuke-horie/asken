import Charts
import SwiftUI
import WidgetKit

// MARK: - WeightHistoryWidgetView

struct WeightHistoryWidgetView: View {
    let entry: WeightHistoryEntry

    var body: some View {
        switch entry.state {
        case .notLoggedIn:
            notLoggedInView
        case .noData:
            noDataView
        case let .loaded(weeklyData):
            loadedView(weeklyData: weeklyData)
        }
    }

    // MARK: - Loaded View

    private func loadedView(weeklyData: [DailyWeightData]) -> some View {
        VStack(alignment: .leading, spacing: 6) {
            headerRow(weeklyData: weeklyData)
            WeightMiniChart(weeklyData: weeklyData)
            Spacer(minLength: 0)
            grassRow(weeklyData: weeklyData)
        }
        .padding(12)
    }

    // MARK: - Header Row

    private func headerRow(weeklyData: [DailyWeightData]) -> some View {
        HStack(alignment: .firstTextBaseline) {
            Label("体重推移", systemImage: "scalemass")
                .font(.caption2)
                .foregroundStyle(.secondary)

            Spacer()

            if let latest = weeklyData.compactMap(\.latestWeightKg).last {
                HStack(alignment: .lastTextBaseline, spacing: 1) {
                    Text("\(latest, format: .number.precision(.fractionLength(1)))")
                        .font(.system(.caption, design: .rounded, weight: .semibold))
                    Text("kg")
                        .font(.system(size: 8))
                        .foregroundStyle(.secondary)
                }
            }
        }
    }

    // MARK: - Grass Row

    private func grassRow(weeklyData: [DailyWeightData]) -> some View {
        VStack(spacing: 3) {
            HStack(spacing: 4) {
                ForEach(weeklyData) { day in
                    RoundedRectangle(cornerRadius: 3)
                        .fill(day.hasRecord ? Color.blue : Color.secondary.opacity(0.2))
                        .frame(maxWidth: .infinity)
                        .aspectRatio(1, contentMode: .fit)
                }
            }

            HStack(spacing: 4) {
                ForEach(weeklyData) { day in
                    Text(Self.dayLabel(for: day.date))
                        .font(.system(size: 7))
                        .foregroundStyle(.secondary)
                        .frame(maxWidth: .infinity)
                }
            }
        }
    }

    // MARK: - State Views

    private var notLoggedInView: some View {
        VStack(spacing: 4) {
            Label("体重履歴", systemImage: "scalemass")
                .font(.caption2)
                .foregroundStyle(.secondary)
            Spacer()
            Text("ログインが必要です")
                .font(.caption2)
                .foregroundStyle(.secondary)
            Spacer()
        }
        .padding(12)
    }

    private var noDataView: some View {
        VStack(spacing: 4) {
            Label("体重履歴", systemImage: "scalemass")
                .font(.caption2)
                .foregroundStyle(.secondary)
            Spacer()
            Text("データを取得できませんでした")
                .font(.caption2)
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)
            Spacer()
        }
        .padding(12)
    }

    // MARK: - Helpers

    private static let dayLabelFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "ja_JP")
        formatter.dateFormat = "E"
        return formatter
    }()

    private static func dayLabel(for date: Date) -> String {
        dayLabelFormatter.string(from: date)
    }
}

// MARK: - WeightMiniChart

private struct WeightMiniChart: View {
    let weeklyData: [DailyWeightData]

    private struct ChartPoint: Identifiable {
        let id: Date
        let date: Date
        let weight: Double
    }

    private var chartPoints: [ChartPoint] {
        weeklyData.compactMap { day in
            guard let weight = day.latestWeightKg else { return nil }
            return ChartPoint(id: day.date, date: day.date, weight: weight)
        }
    }

    private var yDomain: ClosedRange<Double> {
        let weights = chartPoints.map(\.weight)
        let minW = (weights.min() ?? 60) - 0.5
        let maxW = (weights.max() ?? 70) + 0.5
        return minW ... maxW
    }

    var body: some View {
        if chartPoints.isEmpty {
            Text("データなし")
                .font(.caption2)
                .foregroundStyle(.secondary)
                .frame(maxWidth: .infinity, maxHeight: .infinity)
        } else {
            Chart {
                ForEach(chartPoints) { point in
                    LineMark(
                        x: .value("日付", point.date),
                        y: .value("体重", point.weight)
                    )
                    .foregroundStyle(Color.blue)
                    .interpolationMethod(.catmullRom)

                    PointMark(
                        x: .value("日付", point.date),
                        y: .value("体重", point.weight)
                    )
                    .foregroundStyle(Color.blue)
                    .symbolSize(16)
                }
            }
            .chartXAxis(.hidden)
            .chartYAxis(.hidden)
            .chartYScale(domain: yDomain)
            .frame(height: 52)
        }
    }
}
