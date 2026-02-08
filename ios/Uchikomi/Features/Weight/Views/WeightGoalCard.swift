import SwiftUI

struct WeightGoalCard: View {
    let latestWeight: Double?
    let goal: WeightGoal?
    let difference: Double?
    let onGoalTap: () -> Void

    var body: some View {
        VStack(spacing: 12) {
            HStack {
                Text("体重サマリー")
                    .font(.headline)
                Spacer()
                Button(action: onGoalTap) {
                    Image(systemName: "target")
                        .font(.title3)
                        .foregroundStyle(Theme.primary)
                }
            }

            HStack(spacing: 24) {
                // 現在体重
                VStack(spacing: 4) {
                    Text("現在")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                    if let weight = latestWeight {
                        Text(String(format: "%.1f", weight))
                            .font(.title)
                            .fontWeight(.bold)
                            + Text(" kg")
                            .font(.subheadline)
                            .foregroundColor(.secondary)
                    } else {
                        Text("-- kg")
                            .font(.title)
                            .foregroundStyle(.secondary)
                    }
                }

                if let goal {
                    Divider()
                        .frame(height: 40)

                    // 目標体重
                    VStack(spacing: 4) {
                        Text("目標")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                        Text(String(format: "%.1f", goal.targetWeightKg))
                            .font(.title2)
                            .fontWeight(.semibold)
                            + Text(" kg")
                            .font(.caption)
                            .foregroundColor(.secondary)
                    }

                    Divider()
                        .frame(height: 40)

                    // 差分
                    VStack(spacing: 4) {
                        Text("あと")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                        if let diff = difference {
                            Text(String(format: "%+.1f", diff))
                                .font(.title2)
                                .fontWeight(.semibold)
                                .foregroundStyle(diff <= 0 ? .green : Theme.primary)
                                + Text(" kg")
                                .font(.caption)
                                .foregroundColor(.secondary)
                        } else {
                            Text("-- kg")
                                .font(.title2)
                                .foregroundStyle(.secondary)
                        }
                    }
                }

                Spacer()
            }
        }
        .padding()
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}
