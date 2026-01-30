import SwiftUI
import Charts

struct WeightView: View {
    @State private var viewModel = WeightViewModel()
    @State private var showingWeightInput = false
    @State private var showingGoalInput = false

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 20) {
                    // Current Weight Card
                    CurrentWeightCard(
                        weight: viewModel.latestWeightText,
                        goalDifference: viewModel.goalDifferenceText
                    )

                    // Period Picker
                    Picker("期間", selection: $viewModel.selectedPeriod) {
                        ForEach(WeightPeriod.allCases, id: \.self) { period in
                            Text(period.displayName).tag(period)
                        }
                    }
                    .pickerStyle(.segmented)
                    .onChange(of: viewModel.selectedPeriod) { _, newValue in
                        viewModel.changePeriod(to: newValue)
                    }

                    // Chart
                    if !viewModel.records.isEmpty {
                        WeightChart(
                            records: viewModel.records,
                            goal: viewModel.goal?.targetWeight
                        )
                        .frame(height: 200)
                        .padding()
                        .background(Color(.systemBackground))
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                    }

                    // Stats
                    if let stats = viewModel.stats {
                        StatsCard(stats: stats)
                    }

                    // Goal Section
                    GoalSection(
                        goal: viewModel.goal,
                        onSetGoal: { showingGoalInput = true }
                    )

                    // Error Message
                    if let error = viewModel.errorMessage {
                        Text(error)
                            .font(.caption)
                            .foregroundStyle(.red)
                            .multilineTextAlignment(.center)
                    }
                }
                .padding()
            }
            .navigationTitle("体重管理")
            .toolbar {
                ToolbarItem(placement: .primaryAction) {
                    Button {
                        showingWeightInput = true
                    } label: {
                        Image(systemName: "plus.circle.fill")
                            .font(.title2)
                            .foregroundStyle(Theme.primary)
                    }
                }
            }
            .refreshable {
                await viewModel.loadData()
            }
            .sheet(isPresented: $showingWeightInput) {
                WeightInputSheet(viewModel: viewModel)
            }
            .sheet(isPresented: $showingGoalInput) {
                GoalInputSheet(viewModel: viewModel)
            }
        }
        .task {
            await viewModel.loadData()
        }
    }
}

// MARK: - Subviews

private struct CurrentWeightCard: View {
    let weight: String
    let goalDifference: String?

    var body: some View {
        VStack(spacing: 8) {
            Text("現在の体重")
                .font(.subheadline)
                .foregroundStyle(.secondary)

            HStack(alignment: .lastTextBaseline, spacing: 4) {
                Text(weight)
                    .font(.system(size: 48, weight: .bold))
                    .foregroundStyle(Theme.primary)
                Text("kg")
                    .font(.title2)
                    .foregroundStyle(.secondary)
            }

            if let diff = goalDifference {
                Text(diff)
                    .font(.subheadline)
                    .foregroundStyle(Theme.primary)
            }
        }
        .frame(maxWidth: .infinity)
        .padding()
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .shadow(color: .black.opacity(0.1), radius: 4, x: 0, y: 2)
    }
}

private struct WeightChart: View {
    let records: [WeightRecord]
    let goal: Double?

    var chartData: [(date: Date, weight: Double)] {
        records.compactMap { record in
            guard let date = record.recordedDate else { return nil }
            return (date: date, weight: record.weight)
        }
        .sorted { $0.date < $1.date }
    }

    var body: some View {
        Chart {
            ForEach(chartData, id: \.date) { item in
                LineMark(
                    x: .value("日付", item.date),
                    y: .value("体重", item.weight)
                )
                .foregroundStyle(Theme.primary)

                PointMark(
                    x: .value("日付", item.date),
                    y: .value("体重", item.weight)
                )
                .foregroundStyle(Theme.primary)
            }

            if let goal = goal {
                RuleMark(y: .value("目標", goal))
                    .foregroundStyle(.red.opacity(0.5))
                    .lineStyle(StrokeStyle(lineWidth: 1, dash: [5]))
            }
        }
        .chartYScale(domain: yAxisDomain)
    }

    var yAxisDomain: ClosedRange<Double> {
        let weights = chartData.map(\.weight)
        let minWeight = (weights.min() ?? 50) - 2
        let maxWeight = (weights.max() ?? 80) + 2

        if let goal = goal {
            return min(minWeight, goal - 2)...max(maxWeight, goal + 2)
        }
        return minWeight...maxWeight
    }
}

private struct StatsCard: View {
    let stats: WeightStats

    var body: some View {
        HStack(spacing: 0) {
            StatItem(title: "最小", value: stats.min)
            Divider()
            StatItem(title: "平均", value: stats.average)
            Divider()
            StatItem(title: "最大", value: stats.max)
        }
        .padding()
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

private struct StatItem: View {
    let title: String
    let value: Double

    var body: some View {
        VStack(spacing: 4) {
            Text(title)
                .font(.caption)
                .foregroundStyle(.secondary)
            Text(String(format: "%.1f", value))
                .font(.headline)
            Text("kg")
                .font(.caption2)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity)
    }
}

private struct GoalSection: View {
    let goal: WeightGoal?
    let onSetGoal: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("目標")
                    .font(.headline)
                Spacer()
                Button(goal == nil ? "設定" : "変更", action: onSetGoal)
                    .font(.subheadline)
            }

            if let goal = goal {
                HStack {
                    VStack(alignment: .leading, spacing: 4) {
                        Text("目標体重")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                        Text(String(format: "%.1f kg", goal.targetWeight))
                            .font(.title3)
                            .fontWeight(.semibold)
                    }

                    Spacer()

                    VStack(alignment: .trailing, spacing: 4) {
                        Text("残り")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                        Text("\(goal.daysRemaining) 日")
                            .font(.title3)
                            .fontWeight(.semibold)
                    }
                }
            } else {
                Text("目標を設定して減量をトラッキングしましょう")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }
        }
        .padding()
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

// MARK: - Input Sheets

private struct WeightInputSheet: View {
    @Environment(\.dismiss) private var dismiss
    @Bindable var viewModel: WeightViewModel

    var body: some View {
        NavigationStack {
            Form {
                Section("体重") {
                    HStack {
                        TextField("体重", text: $viewModel.inputWeight)
                            .keyboardType(.decimalPad)
                        Text("kg")
                            .foregroundStyle(.secondary)
                    }
                }

                Section("日付") {
                    DatePicker(
                        "記録日",
                        selection: $viewModel.inputDate,
                        in: ...Date(),
                        displayedComponents: .date
                    )
                }

                if let error = viewModel.errorMessage {
                    Section {
                        Text(error)
                            .foregroundStyle(.red)
                    }
                }
            }
            .navigationTitle("体重を記録")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("キャンセル") {
                        dismiss()
                    }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("保存") {
                        Task {
                            await viewModel.saveWeight()
                            if viewModel.errorMessage == nil {
                                dismiss()
                            }
                        }
                    }
                    .disabled(viewModel.inputWeight.isEmpty || viewModel.isSaving)
                }
            }
        }
    }
}

private struct GoalInputSheet: View {
    @Environment(\.dismiss) private var dismiss
    @Bindable var viewModel: WeightViewModel

    var body: some View {
        NavigationStack {
            Form {
                Section("目標体重") {
                    HStack {
                        TextField("目標体重", text: $viewModel.goalTargetWeight)
                            .keyboardType(.decimalPad)
                        Text("kg")
                            .foregroundStyle(.secondary)
                    }
                }

                Section("目標日") {
                    DatePicker(
                        "目標日",
                        selection: $viewModel.goalTargetDate,
                        in: Date()...,
                        displayedComponents: .date
                    )
                }

                if let error = viewModel.errorMessage {
                    Section {
                        Text(error)
                            .foregroundStyle(.red)
                    }
                }
            }
            .navigationTitle("目標を設定")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("キャンセル") {
                        dismiss()
                    }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("設定") {
                        Task {
                            await viewModel.setGoal()
                            if viewModel.errorMessage == nil {
                                dismiss()
                            }
                        }
                    }
                    .disabled(viewModel.goalTargetWeight.isEmpty || viewModel.isSettingGoal)
                }
            }
        }
    }
}

#Preview {
    WeightView()
}
