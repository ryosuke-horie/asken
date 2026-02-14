import SwiftUI

// MARK: - CalorieProgressView

struct CalorieProgressView: View {
    let current: Double
    let goal: Double

    private var progress: Double {
        guard goal > 0 else { return 0 }
        return min(current / goal, 1.0)
    }

    private var remaining: Double {
        max(goal - current, 0)
    }

    private var isOverGoal: Bool {
        current > goal
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text("カロリー目標")
                    .font(.caption)
                    .foregroundStyle(.secondary)

                Spacer()

                Text("\(Int(current)) / \(Int(goal)) kcal")
                    .font(.caption)
                    .foregroundStyle(isOverGoal ? .orange : .primary)
            }

            GeometryReader { geometry in
                ZStack(alignment: .leading) {
                    // 背景バー
                    RoundedRectangle(cornerRadius: 4)
                        .fill(Color(.systemGray5))
                        .frame(height: 8)

                    // プログレスバー
                    RoundedRectangle(cornerRadius: 4)
                        .fill(isOverGoal ? .orange : Theme.primary)
                        .frame(width: geometry.size.width * CGFloat(progress), height: 8)
                        .animation(.easeInOut, value: progress)
                }
            }
            .frame(height: 8)

            if !isOverGoal, remaining > 0 {
                HStack {
                    Image(systemName: "flag.fill")
                        .font(.caption)
                        .foregroundStyle(.secondary)

                    Text("あと \(Int(remaining)) kcal")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }
        }
    }
}

#Preview {
    VStack(spacing: 24) {
        CalorieProgressView(current: 650, goal: 2_000)
        CalorieProgressView(current: 1_800, goal: 2_000)
        CalorieProgressView(current: 2_200, goal: 2_000)
        CalorieProgressView(current: 0, goal: 2_000)
    }
    .padding()
    .background(Color(.systemGroupedBackground))
}
