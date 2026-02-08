import Charts
import SwiftUI

struct WeightChartView: View {
    let records: [WeightRecord]
    let goal: WeightGoal?
    @Binding var selectedPeriod: ChartPeriod

    private static let iso8601Formatter: ISO8601DateFormatter = {
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return f
    }()

    private static let iso8601FallbackFormatter: ISO8601DateFormatter = {
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime]
        return f
    }()

    private var chartData: [(date: Date, weight: Double)] {
        records.compactMap { record in
            guard let date = Self.iso8601Formatter.date(from: record.recordedAt)
                ?? Self.iso8601FallbackFormatter.date(from: record.recordedAt) else {
                return nil
            }
            return (date: date, weight: record.weightKg)
        }
    }

    private var yRange: ClosedRange<Double> {
        var allWeights = chartData.map(\.weight)
        if let goal {
            allWeights.append(goal.targetWeightKg)
        }
        guard let minW = allWeights.min(), let maxW = allWeights.max() else {
            return 50 ... 80
        }
        let padding = max((maxW - minW) * 0.2, 1.0)
        return (minW - padding) ... (maxW + padding)
    }

    var body: some View {
        VStack(spacing: 12) {
            // 期間セグメント
            Picker("期間", selection: $selectedPeriod) {
                ForEach(ChartPeriod.allCases) { period in
                    Text(period.displayName).tag(period)
                }
            }
            .pickerStyle(.segmented)

            if chartData.isEmpty {
                Text("この期間のデータがありません")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .frame(height: 200)
            } else {
                Chart {
                    ForEach(Array(chartData.enumerated()), id: \.offset) { _, point in
                        LineMark(
                            x: .value("日付", point.date),
                            y: .value("体重", point.weight)
                        )
                        .foregroundStyle(Theme.primary)
                        .interpolationMethod(.catmullRom)

                        PointMark(
                            x: .value("日付", point.date),
                            y: .value("体重", point.weight)
                        )
                        .foregroundStyle(Theme.primary)
                        .symbolSize(30)
                    }

                    if let goal {
                        RuleMark(y: .value("目標", goal.targetWeightKg))
                            .foregroundStyle(.red.opacity(0.6))
                            .lineStyle(StrokeStyle(lineWidth: 1, dash: [5, 3]))
                            .annotation(position: .top, alignment: .trailing) {
                                Text("目標 \(String(format: "%.1f", goal.targetWeightKg))kg")
                                    .font(.caption2)
                                    .foregroundStyle(.red)
                            }
                    }
                }
                .chartYScale(domain: yRange)
                .frame(height: 200)
            }
        }
        .padding()
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}
