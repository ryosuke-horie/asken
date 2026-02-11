import Charts
import SwiftUI

// MARK: - PFCPieChart

struct PFCPieChart: View {
    let protein: Double
    let fat: Double
    let carbohydrates: Double

    private var total: Double {
        protein + fat + carbohydrates
    }

    private var proteinRatio: Double {
        guard total > 0 else { return 0 }
        return protein / total
    }

    private var fatRatio: Double {
        guard total > 0 else { return 0 }
        return fat / total
    }

    private var carbsRatio: Double {
        guard total > 0 else { return 0 }
        return carbohydrates / total
    }

    var body: some View {
        VStack(spacing: 8) {
            ZStack {
                // 円グラフ
                Chart {
                    SectorMark(
                        angle: .value("たんぱく質", proteinRatio * 360),
                        innerRadius: .ratio(0.6),
                        angularInset: 2
                    )
                    .foregroundStyle(.red)
                    .opacity(proteinRatio > 0 ? 1 : 0)

                    SectorMark(
                        angle: .value("脂質", fatRatio * 360),
                        innerRadius: .ratio(0.6),
                        angularInset: 2
                    )
                    .foregroundStyle(.yellow)
                    .opacity(fatRatio > 0 ? 1 : 0)

                    SectorMark(
                        angle: .value("炭水化物", carbsRatio * 360),
                        innerRadius: .ratio(0.6),
                        angularInset: 2
                    )
                    .foregroundStyle(.blue)
                    .opacity(carbsRatio > 0 ? 1 : 0)
                }
                .frame(width: 120, height: 120)

                // 中心のテキスト（全カロリーまたは合計g）
                VStack(spacing: 2) {
                    Text("\(Int(total))")
                        .font(.title2)
                        .fontWeight(.bold)
                        .foregroundStyle(.primary)
                    Text("g")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }

            // 凡例
            HStack(spacing: 12) {
                LegendItem(color: .red, name: "P", value: protein, ratio: proteinRatio)
                LegendItem(color: .yellow, name: "F", value: fat, ratio: fatRatio)
                LegendItem(color: .blue, name: "C", value: carbohydrates, ratio: carbsRatio)
            }
            .font(.caption)
        }
    }
}

// MARK: - LegendItem

private struct LegendItem: View {
    let color: Color
    let name: String
    let value: Double
    let ratio: Double

    var body: some View {
        VStack(spacing: 2) {
            Circle()
                .fill(color)
                .frame(width: 8, height: 8)

            Text(name)
                .foregroundStyle(.secondary)

            Text(String(format: "%.1fg", value))
                .foregroundStyle(.primary)

            Text(String(format: "%.0f%%", ratio * 100))
                .foregroundStyle(.secondary)
        }
    }
}

#Preview {
    HStack(spacing: 32) {
        PFCPieChart(
            protein: 25.5,
            fat: 22.3,
            carbohydrates: 78.0
        )

        PFCPieChart(
            protein: 0,
            fat: 0,
            carbohydrates: 0
        )
    }
    .padding()
    .background(Color(.systemGroupedBackground))
}
