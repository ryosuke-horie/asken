import SwiftUI

// MARK: - CalorieBalanceBarView

struct CalorieBalanceBarView: View {
    let intake: Double
    let burned: Double
    let goal: Double

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                Text("カロリーバランス")
                    .font(.caption)
                    .foregroundStyle(.secondary)
                Spacer()
                Text("目標 \(Int(goal)) kcal")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            CalorieBarRow(
                label: "摂取",
                value: intake,
                goal: goal,
                color: Theme.primary
            )

            CalorieBarRow(
                label: "消費",
                value: burned,
                goal: goal,
                color: .blue
            )
        }
    }
}

// MARK: - CalorieBarRow

private struct CalorieBarRow: View {
    let label: String
    let value: Double
    let goal: Double
    let color: Color

    private var progress: Double {
        guard goal > 0 else { return 0 }
        return min(value / goal, 1.0)
    }

    var body: some View {
        HStack(spacing: 10) {
            HStack(spacing: 4) {
                Circle()
                    .fill(color)
                    .frame(width: 8, height: 8)
                Text(label)
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .frame(width: 24, alignment: .leading)
            }

            GeometryReader { geometry in
                ZStack(alignment: .leading) {
                    RoundedRectangle(cornerRadius: 4)
                        .fill(Color(.systemGray5))
                        .frame(height: 12)

                    RoundedRectangle(cornerRadius: 4)
                        .fill(color.opacity(0.85))
                        .frame(width: geometry.size.width * CGFloat(progress), height: 12)
                        .animation(.easeInOut(duration: 0.3), value: progress)
                }
            }
            .frame(height: 12)

            Text("\(Int(value))")
                .font(.caption)
                .fontWeight(.semibold)
                .foregroundStyle(color)
                .frame(width: 40, alignment: .trailing)
        }
    }
}

#Preview {
    VStack(spacing: 24) {
        CalorieBalanceBarView(intake: 1_450, burned: 514, goal: 2_000)
        CalorieBalanceBarView(intake: 2_100, burned: 800, goal: 2_000)
        CalorieBalanceBarView(intake: 0, burned: 300, goal: 2_000)
    }
    .padding()
    .background(Color(.systemGroupedBackground))
}
