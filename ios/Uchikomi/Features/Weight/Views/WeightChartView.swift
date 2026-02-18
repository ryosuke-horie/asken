import Charts
import SwiftUI

// MARK: - WeightTiming + Color

extension WeightTiming {
    var color: Color {
        switch self {
        case .morning: Color.blue
        case .beforePractice: Color.orange
        case .afterPractice: Color.green
        case .beforeSleep: Color.purple
        }
    }
}

// MARK: - WeightChartView

struct WeightChartView: View {
    let records: [WeightRecord]
    let goal: WeightGoal?
    @Binding var selectedPeriod: ChartPeriod

    private struct ChartPoint: Identifiable {
        let id: String
        let date: Date
        let weight: Double
        let timing: WeightTiming
    }

    private var chartPoints: [ChartPoint] {
        records.compactMap { record in
            guard let date = WeightRecord.parseISO8601(record.recordedAt),
                  let timing = WeightTiming.from(note: record.note) else {
                return nil
            }
            return ChartPoint(id: record.id, date: date, weight: record.weightKg, timing: timing)
        }
    }

    private var yRange: ClosedRange<Double> {
        var allWeights = chartPoints.map(\.weight)
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
            Picker("期間", selection: $selectedPeriod) {
                ForEach(ChartPeriod.allCases) { period in
                    Text(period.displayName).tag(period)
                }
            }
            .pickerStyle(.segmented)

            if chartPoints.isEmpty {
                Text("この期間のデータがありません")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .frame(height: 200)
            } else {
                Chart {
                    ForEach(chartPoints) { point in
                        LineMark(
                            x: .value("日付", point.date),
                            y: .value("体重", point.weight),
                            series: .value("タイミング", point.timing.displayName)
                        )
                        .foregroundStyle(by: .value("タイミング", point.timing.displayName))
                        .interpolationMethod(.catmullRom)

                        PointMark(
                            x: .value("日付", point.date),
                            y: .value("体重", point.weight)
                        )
                        .foregroundStyle(by: .value("タイミング", point.timing.displayName))
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
                .chartForegroundStyleScale(
                    domain: WeightTiming.allCases.map(\.displayName),
                    range: WeightTiming.allCases.map(\.color)
                )
                .chartYScale(domain: yRange)
                .chartLegend(position: .bottom, spacing: 8)
                .frame(height: 200)
            }
        }
        .padding()
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}
