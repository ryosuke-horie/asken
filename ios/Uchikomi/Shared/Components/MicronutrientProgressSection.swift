import SwiftUI

// MARK: - MicronutrientProgressSection

struct MicronutrientProgressSection: View {
    let current: [String: Double]
    let targets: [String: Double]

    @State private var isExpanded = false

    var body: some View {
        VStack(spacing: 8) {
            Button {
                withAnimation(.easeInOut(duration: 0.2)) {
                    isExpanded.toggle()
                }
            } label: {
                HStack {
                    Text("微量栄養素")
                        .font(.subheadline)
                        .foregroundStyle(.primary)

                    Spacer()

                    // サマリー（折りたたみ時）
                    if !isExpanded {
                        micronutrientSummary
                    }

                    Image(systemName: "chevron.right")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .rotationEffect(.degrees(isExpanded ? 90 : 0))
                }
            }
            .buttonStyle(.plain)

            if isExpanded {
                VStack(spacing: 8) {
                    // ミネラル・食物繊維
                    MicronutrientGroupView(
                        title: "ミネラル・食物繊維",
                        keys: MicronutrientKey.minerals,
                        current: current,
                        targets: targets
                    )

                    // ビタミン
                    MicronutrientGroupView(
                        title: "ビタミン",
                        keys: MicronutrientKey.vitamins,
                        current: current,
                        targets: targets
                    )
                }
            }
        }
    }

    private var micronutrientSummary: some View {
        let achievedCount = MicronutrientKey.allCases.filter { key in
            let currentVal = current[key.rawValue] ?? 0
            let targetVal = targets[key.rawValue] ?? key.defaultTarget
            return targetVal > 0 && currentVal >= targetVal
        }.count
        let totalCount = MicronutrientKey.allCases.count

        return Text("\(achievedCount)/\(totalCount) 達成")
            .font(.caption)
            .foregroundStyle(.secondary)
    }
}

// MARK: - MicronutrientGroupView

private struct MicronutrientGroupView: View {
    let title: String
    let keys: [MicronutrientKey]
    let current: [String: Double]
    let targets: [String: Double]

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(title)
                .font(.caption)
                .foregroundStyle(.secondary)
                .padding(.top, 4)

            ForEach(keys) { key in
                MicronutrientProgressBar(
                    key: key,
                    current: current[key.rawValue] ?? 0,
                    goal: targets[key.rawValue] ?? key.defaultTarget
                )
            }
        }
    }
}

// MARK: - MicronutrientProgressBar

private struct MicronutrientProgressBar: View {
    let key: MicronutrientKey
    let current: Double
    let goal: Double

    private var progress: Double {
        guard goal > 0 else { return 0 }
        return min(current / goal, 1.0)
    }

    private var isOverGoal: Bool {
        current > goal
    }

    private var formattedCurrent: String {
        formatValue(current)
    }

    private var formattedGoal: String {
        formatValue(goal)
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 2) {
            HStack {
                Circle()
                    .fill(key.color)
                    .frame(width: 6, height: 6)

                Text(key.displayName)
                    .font(.caption2)
                    .foregroundStyle(.secondary)

                Spacer()

                Text("\(formattedCurrent)\(key.unit)")
                    .font(.caption2)
                    .foregroundStyle(isOverGoal ? .orange : .primary)

                Text("/")
                    .font(.caption2)
                    .foregroundStyle(.secondary)

                Text("\(formattedGoal)\(key.unit)")
                    .font(.caption2)
                    .foregroundStyle(.secondary)
            }

            GeometryReader { geometry in
                ZStack(alignment: .leading) {
                    RoundedRectangle(cornerRadius: 2)
                        .fill(Color(.systemGray5))
                        .frame(height: 4)

                    RoundedRectangle(cornerRadius: 2)
                        .fill(isOverGoal ? .orange : key.color)
                        .frame(width: geometry.size.width * CGFloat(progress), height: 4)
                        .animation(.easeInOut, value: progress)
                }
            }
            .frame(height: 4)
        }
    }

    private func formatValue(_ value: Double) -> String {
        if value == value.rounded(), value < 1_000 {
            return String(format: "%.0f", value)
        }
        return String(format: "%.1f", value)
    }
}

#Preview {
    MicronutrientProgressSection(
        current: [
            "iron_mg": 5.2,
            "calcium_mg": 450,
            "zinc_mg": 6.5,
            "fiber_g": 12,
            "vitamin_a_ug": 500,
            "vitamin_c_mg": 120,
        ],
        targets: MicronutrientKey.defaultTargets
    )
    .padding()
    .background(Color(.systemGroupedBackground))
}
