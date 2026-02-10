import SwiftUI

// MARK: - WeightView

struct WeightView: View {
    @State private var viewModel = WeightViewModel()
    @State private var showingInput = false
    @State private var editingRecord: WeightRecord?
    @State private var showingGoalSheet = false

    var body: some View {
        NavigationStack {
            VStack(spacing: 0) {
                if viewModel.isLoading {
                    Spacer()
                    ProgressView()
                    Spacer()
                } else if let error = viewModel.errorMessage {
                    Spacer()
                    WeightErrorView(message: error) {
                        Task { await viewModel.loadData() }
                    }
                    Spacer()
                } else {
                    ScrollView {
                        VStack(spacing: 16) {
                            // 目標カード
                            WeightGoalCard(
                                latestWeight: viewModel.latestWeight,
                                goal: viewModel.goal,
                                difference: viewModel.weightDifferenceFromGoal,
                                onGoalTap: { showingGoalSheet = true }
                            )

                            // グラフ
                            WeightChartView(
                                records: viewModel.chartRecords,
                                goal: viewModel.goal,
                                selectedPeriod: Binding(
                                    get: { viewModel.selectedPeriod },
                                    set: { newPeriod in
                                        viewModel.selectedPeriod = newPeriod
                                        Task { await viewModel.loadChartData() }
                                    }
                                ),
                                isLoading: viewModel.isLoading,
                                isInitialLoading: viewModel.isInitialLoading
                            )

                            // 今日の記録一覧
                            VStack(alignment: .leading, spacing: 8) {
                                HStack {
                                    Text("今日の記録")
                                        .font(.headline)
                                    Spacer()
                                    Button {
                                        showingInput = true
                                    } label: {
                                        Image(systemName: "plus.circle.fill")
                                            .font(.title2)
                                            .foregroundStyle(Theme.primary)
                                    }
                                }

                                if viewModel.todayRecords.isEmpty {
                                    Text("まだ記録がありません")
                                        .font(.subheadline)
                                        .foregroundStyle(.secondary)
                                        .padding(.vertical, 12)
                                } else {
                                    ForEach(viewModel.todayRecords) { record in
                                        Button {
                                            editingRecord = record
                                        } label: {
                                            WeightRecordRow(record: record)
                                        }
                                        .buttonStyle(.plain)
                                    }
                                }
                            }
                            .padding()
                            .background(Color(.systemBackground))
                            .clipShape(RoundedRectangle(cornerRadius: 12))
                        }
                        .padding()
                    }
                }
            }
            .background(Color(.systemGroupedBackground))
            .navigationTitle("体重記録")
            .sheet(isPresented: $showingInput) {
                WeightInputView {
                    Task { await viewModel.loadData() }
                }
            }
            .sheet(item: $editingRecord) { record in
                WeightInputView(editingRecord: record) {
                    Task { await viewModel.loadData() }
                }
            }
            .sheet(isPresented: $showingGoalSheet) {
                WeightGoalSheet(currentGoal: viewModel.goal) {
                    Task { await viewModel.loadData() }
                }
            }
        }
        .task {
            await viewModel.loadData()
        }
    }
}

// MARK: - WeightErrorView

private struct WeightErrorView: View {
    let message: String
    let onRetry: () -> Void

    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "exclamationmark.triangle")
                .font(.largeTitle)
                .foregroundStyle(.orange)

            Text(message)
                .multilineTextAlignment(.center)

            Button("再試行", action: onRetry)
                .buttonStyle(.bordered)
        }
        .padding()
    }
}

#Preview {
    WeightView()
}
