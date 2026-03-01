import SwiftUI
import WidgetKit

// MARK: - WeightWidgetView

struct WeightWidgetView: View {
    let entry: WeightEntry

    @Environment(\.widgetFamily) private var family

    var body: some View {
        switch family {
        case .systemSmall:
            smallView
        case .systemMedium:
            mediumView
        default:
            smallView
        }
    }

    // MARK: - Small Widget

    private var smallView: some View {
        VStack(alignment: .leading, spacing: 6) {
            Label("体重", systemImage: "scalemass")
                .font(.caption2)
                .foregroundStyle(.secondary)

            if !entry.isLoggedIn {
                notLoggedInView
            } else if let weight = entry.latestWeightKg {
                weightDisplayView(weight: weight)
                recordButtonsSmall
            } else {
                noRecordView
            }
        }
        .padding(12)
    }

    // MARK: - Medium Widget

    private var mediumView: some View {
        HStack(spacing: 16) {
            VStack(alignment: .leading, spacing: 6) {
                Label("体重", systemImage: "scalemass")
                    .font(.caption2)
                    .foregroundStyle(.secondary)

                if !entry.isLoggedIn {
                    notLoggedInView
                } else if let weight = entry.latestWeightKg {
                    weightDisplayView(weight: weight)
                    if let target = entry.targetWeightKg {
                        Text("目標: \(target, format: .number.precision(.fractionLength(1)))kg")
                            .font(.caption2)
                            .foregroundStyle(.secondary)
                    }
                } else {
                    noRecordView
                }
            }

            Spacer()

            if entry.isLoggedIn, entry.latestWeightKg != nil {
                recordButtonsMedium
            }
        }
        .padding(14)
    }

    // MARK: - Weight Display

    private func weightDisplayView(weight: Double) -> some View {
        HStack(alignment: .lastTextBaseline, spacing: 2) {
            Text("\(weight, format: .number.precision(.fractionLength(1)))")
                .font(.system(.title2, design: .rounded, weight: .bold))
            Text("kg")
                .font(.caption)
                .foregroundStyle(.secondary)
        }
    }

    // MARK: - Record Buttons (Small)

    private var recordButtonsSmall: some View {
        HStack(spacing: 4) {
            Button(intent: RecordWeightIntent(delta: -0.1)) {
                Text("-0.1")
                    .font(.caption2)
                    .frame(maxWidth: .infinity)
            }
            .buttonStyle(.bordered)
            .tint(.blue)

            Button(intent: RecordWeightIntent(delta: 0)) {
                Text("記録")
                    .font(.caption2)
                    .frame(maxWidth: .infinity)
            }
            .buttonStyle(.borderedProminent)
            .tint(.blue)

            Button(intent: RecordWeightIntent(delta: 0.1)) {
                Text("+0.1")
                    .font(.caption2)
                    .frame(maxWidth: .infinity)
            }
            .buttonStyle(.bordered)
            .tint(.blue)
        }
    }

    // MARK: - Record Buttons (Medium)

    private var recordButtonsMedium: some View {
        VStack(spacing: 6) {
            HStack(spacing: 6) {
                Button(intent: RecordWeightIntent(delta: -0.5)) {
                    Text("-0.5")
                        .font(.caption)
                        .frame(width: 50)
                }
                .buttonStyle(.bordered)
                .tint(.orange)

                Button(intent: RecordWeightIntent(delta: 0.5)) {
                    Text("+0.5")
                        .font(.caption)
                        .frame(width: 50)
                }
                .buttonStyle(.bordered)
                .tint(.orange)
            }

            HStack(spacing: 6) {
                Button(intent: RecordWeightIntent(delta: -0.1)) {
                    Text("-0.1")
                        .font(.caption)
                        .frame(width: 50)
                }
                .buttonStyle(.bordered)
                .tint(.blue)

                Button(intent: RecordWeightIntent(delta: 0.1)) {
                    Text("+0.1")
                        .font(.caption)
                        .frame(width: 50)
                }
                .buttonStyle(.bordered)
                .tint(.blue)
            }

            Button(intent: RecordWeightIntent(delta: 0)) {
                Text("そのまま記録")
                    .font(.caption)
                    .frame(width: 106)
            }
            .buttonStyle(.borderedProminent)
            .tint(.blue)
        }
    }

    // MARK: - States

    private var notLoggedInView: some View {
        Text("ログインが必要です")
            .font(.caption2)
            .foregroundStyle(.secondary)
    }

    private var noRecordView: some View {
        Text("記録なし")
            .font(.caption)
            .foregroundStyle(.secondary)
    }
}
