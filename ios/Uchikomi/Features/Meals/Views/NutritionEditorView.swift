import SwiftUI

struct NutritionEditorView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var viewModel: NutritionEditorViewModel

    let onSaved: () -> Void

    init(
        historyId: String,
        foods: [NutritionInfo] = [],
        onSaved: @escaping () -> Void = {}
    ) {
        _viewModel = State(initialValue: NutritionEditorViewModel(
            historyId: historyId,
            foods: foods
        ))
        self.onSaved = onSaved
    }

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isLoading {
                    ProgressView()
                } else {
                    ScrollView {
                        VStack(spacing: 16) {
                            // 説明テキスト
                            Text("料理名と量を編集してください。\n量を変更すると栄養素が自動で再計算されます。")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                                .multilineTextAlignment(.center)
                                .padding(.horizontal)

                            // 現在の栄養素サマリー（参考値）
                            if viewModel.totalCalories > 0 {
                                NutritionSummaryCard(
                                    calories: viewModel.totalCalories,
                                    protein: viewModel.totalProtein,
                                    fat: viewModel.totalFat,
                                    carbohydrates: viewModel.totalCarbohydrates,
                                    micronutrients: viewModel.totalMicronutrients
                                )
                            }

                            ForEach(viewModel.foods) { food in
                                FoodItemEditRow(item: food) {
                                    viewModel.removeFood(food)
                                }
                            }

                            Button {
                                viewModel.addFood()
                            } label: {
                                Label("メニューを追加", systemImage: "plus.circle")
                                    .frame(maxWidth: .infinity)
                                    .padding()
                                    .background(Color(.secondarySystemBackground))
                                    .clipShape(RoundedRectangle(cornerRadius: 10))
                            }

                            if let error = viewModel.errorMessage {
                                Text(error)
                                    .font(.caption)
                                    .foregroundStyle(.red)
                                    .multilineTextAlignment(.center)
                            }

                            if let message = viewModel.recalculatingMessage {
                                Text(message)
                                    .font(.caption)
                                    .foregroundStyle(.orange)
                                    .multilineTextAlignment(.center)
                            }
                        }
                        .padding()
                    }
                }
            }
            .navigationTitle("食事を編集")
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
                            await viewModel.save()
                        }
                    }
                    .disabled(!viewModel.canSave || viewModel.isSaving)
                }
            }
            .onChange(of: viewModel.isSaved) { _, isSaved in
                if isSaved {
                    if viewModel.hasAnyNameChanged {
                        viewModel.recalculatingMessage = "栄養素を再計算中です。次回表示で反映されます。"
                        Task {
                            do {
                                try await Task.sleep(for: .seconds(2))
                            } catch {
                                // Taskキャンセル時は即座にdismiss
                            }
                            onSaved()
                            dismiss()
                        }
                    } else {
                        onSaved()
                        dismiss()
                    }
                }
            }
        }
        .task {
            if viewModel.foods.isEmpty {
                await viewModel.loadFromHistory()
            }
        }
        .interactiveDismissDisabled(viewModel.isSaving)
    }
}

#Preview {
    NutritionEditorView(
        historyId: "preview-id",
        foods: [
            NutritionInfo(
                name: "鶏むね肉",
                estimatedAmount: "100g",
                caloriesKcal: 165,
                proteinG: 31,
                fatG: 3.6,
                carbohydratesG: 0
            ),
        ]
    )
}
