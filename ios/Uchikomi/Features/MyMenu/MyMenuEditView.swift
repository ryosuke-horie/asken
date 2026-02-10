import SwiftUI

// MARK: - MyMenuEditView

struct MyMenuEditView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var viewModel: MyMenuEditViewModel

    init(menuItem: MyMenuItem? = nil) {
        _viewModel = State(initialValue: MyMenuEditViewModel(menuItem: menuItem))
    }

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isSaving {
                    ProgressView()
                } else {
                    ScrollView {
                        VStack(spacing: 16) {
                            // メニュー名入力
                            VStack(alignment: .leading, spacing: 8) {
                                Text("メニュー名")
                                    .font(.headline)

                                TextField("例: お気に入り朝食", text: $viewModel.menuName)
                                    .textFieldStyle(.roundedBorder)
                            }
                            .padding(.horizontal)

                            // 説明テキスト
                            Text("よく食べる食事を登録すると、ワンタップで記録できます。")
                                .font(.caption)
                                .foregroundStyle(.primary)
                                .opacity(0.6)
                                .multilineTextAlignment(.center)
                                .padding(.horizontal)

                            // 栄養素サマリー
                            if viewModel.totalCalories > 0 {
                                NutritionSummaryCard(
                                    calories: viewModel.totalCalories,
                                    protein: viewModel.totalProtein,
                                    fat: viewModel.totalFat,
                                    carbohydrates: viewModel.totalCarbohydrates
                                )
                            }

                            // 食品リスト
                            ForEach(viewModel.foodItems) { food in
                                FoodItemEditRow(item: food) {
                                    if let index = viewModel.foodItems.firstIndex(where: { $0.id == food.id }) {
                                        viewModel.removeFoodItem(at: index)
                                    }
                                }
                            }

                            Button {
                                viewModel.addFoodItem()
                            } label: {
                                Label("食品を追加", systemImage: "plus.circle")
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
                        }
                        .padding()
                    }
                }
            }
            .navigationTitle(viewModel.isEditMode ? "メニューを編集" : "メニューを登録")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("キャンセル") {
                        dismiss()
                    }
                }

                ToolbarItem(placement: .confirmationAction) {
                    if viewModel.isEditMode {
                        Menu {
                            Button {
                                Task {
                                    await viewModel.save()
                                }
                            } label: {
                                Text("保存")
                            }

                            Divider()

                            Button(role: .destructive) {
                                Task {
                                    await viewModel.delete()
                                }
                            } label: {
                                Text("削除")
                            }
                        } label: {
                            Text("完了")
                        }
                    } else {
                        Button("保存") {
                            Task {
                                await viewModel.save()
                            }
                        }
                        .disabled(!viewModel.isValid || viewModel.isSaving)
                    }
                }
            }
            .onChange(of: viewModel.shouldDismiss) { _, shouldDismiss in
                if shouldDismiss {
                    dismiss()
                }
            }
        }
    }
}

// MARK: - Preview

#Preview("新規作成") {
    NavigationStack {
        MyMenuEditView()
    }
}

#Preview("編集") {
    let sampleItem = MyMenuItem(
        id: UUID().uuidString,
        name: "お気に入り朝食",
        foods: [
            NutritionInfo(
                name: "グラノーラ",
                estimatedAmount: "100g",
                caloriesKcal: 350,
                proteinG: 10,
                fatG: 5,
                carbohydratesG: 50
            ),
            NutritionInfo(
                name: "牛乳",
                estimatedAmount: "200ml",
                caloriesKcal: 100,
                proteinG: 6,
                fatG: 3,
                carbohydratesG: 10
            )
        ],
        totalCalories: 450,
        totalProtein: 16,
        totalFat: 8,
        totalCarbohydrates: 60,
        createdAt: Date(),
        updatedAt: Date()
    )

    return NavigationStack {
        MyMenuEditView(menuItem: sampleItem)
    }
}
