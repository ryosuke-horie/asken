import SwiftUI

// MARK: - PFCProgressBar

struct PFCProgressBar: View {
    let current: Double
    let goal: Double
    let name: String
    let color: Color

    private var progress: Double {
        guard goal > 0 else { return 0 }
        return min(current / goal, 1.0)
    }

    private var isOverGoal: Bool {
        current > goal
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            HStack {
                Text(name)
                    .font(.caption)
                    .foregroundStyle(.secondary)

                Spacer()

                Text(String(format: "%.1fg", current))
                    .font(.caption)
                    .foregroundStyle(isOverGoal ? .orange : .primary)

                Text("/")
                    .font(.caption)
                    .foregroundStyle(.secondary)

                Text(String(format: "%.1fg", goal))
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            GeometryReader { geometry in
                ZStack(alignment: .leading) {
                    // 背景バー
                    RoundedRectangle(cornerRadius: 3)
                        .fill(Color(.systemGray5))
                        .frame(height: 6)

                    // プログレスバー
                    RoundedRectangle(cornerRadius: 3)
                        .fill(isOverGoal ? .orange : color)
                        .frame(width: geometry.size.width * CGFloat(progress), height: 6)
                        .animation(.easeInOut, value: progress)
                }
            }
            .frame(height: 6)
        }
    }
}

// MARK: - PFCProgressRow

struct PFCProgressRow: View {
    let currentProtein: Double
    let currentFat: Double
    let currentCarbs: Double
    let goalProtein: Double
    let goalFat: Double
    let goalCarbs: Double

    var body: some View {
        VStack(spacing: 12) {
            PFCProgressBar(
                current: currentProtein,
                goal: goalProtein,
                name: "たんぱく質",
                color: .red
            )

            PFCProgressBar(
                current: currentFat,
                goal: goalFat,
                name: "脂質",
                color: .yellow
            )

            PFCProgressBar(
                current: currentCarbs,
                goal: goalCarbs,
                name: "炭水化物",
                color: .blue
            )
        }
    }
}

#Preview {
    VStack(spacing: 32) {
        PFCProgressRow(
            currentProtein: 65,
            currentFat: 45,
            currentCarbs: 180,
            goalProtein: 100,
            goalFat: 60,
            goalCarbs: 250
        )

        PFCProgressRow(
            currentProtein: 120,
            currentFat: 80,
            currentCarbs: 300,
            goalProtein: 100,
            goalFat: 60,
            goalCarbs: 250
        )
    }
    .padding()
    .background(Color(.systemGroupedBackground))
}
