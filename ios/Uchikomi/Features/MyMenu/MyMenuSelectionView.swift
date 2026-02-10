import SwiftUI

// MARK: - MyMenuSelectionView

struct MyMenuSelectionView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var viewModel = MyMenuListViewModel()

    let selectedMealType: MealType
    let mealDate: Date
    let onRecorded: () -> Void

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isLoading {
                    ProgressView()
                } else if let error = viewModel.errorMessage {
                    ErrorView(message: error) {
                        Task {
                            await viewModel.loadMyMenuList()
                        }
                    }
                } else if viewModel.myMenuItems.isEmpty {
                    ContentUnavailableView {
                        Label("マイメニューがありません", systemImage: "star")
                    } description: {
                        Text("よく食べる食事を登録すると、ワンタップで記録できます")
                    }
                } else {
                    ScrollView {
                        VStack(spacing: 16) {
                            ForEach(viewModel.myMenuItems) { item in
                                MyMenuSelectionCard(
                                    item: item,
                                    mealType: selectedMealType,
                                    mealDate: mealDate,
                                    onRecorded: onRecorded
                                )
                            }
                        }
                        .padding()
                    }
                }
            }
            .navigationTitle("マイメニューを選択")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("キャンセル") {
                        dismiss()
                    }
                }
            }
        }
    }
}

// MARK: - MyMenuSelectionCard

struct MyMenuSelectionCard: View {
    let item: MyMenuItem
    let mealType: MealType
    let mealDate: Date
    let onRecorded: () -> Void

    @State private var isRecording = false

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    Text(item.name)
                        .font(.headline)

                    Text(item.foods.map { $0.name }.joined(separator: ", "))
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .lineLimit(2)
                }

                Spacer()

                Button {
                    Task {
                        await recordFromMyMenu()
                    }
                } label: {
                    if isRecording {
                        ProgressView()
                            .scaleEffect(0.8)
                    } else {
                        Text("記録")
                    }
                }
                .buttonStyle(.borderedProminent)
                .disabled(isRecording)
            }

            HStack(spacing: 16) {
                NutritionLabel(value: Int(item.totalCalories), unit: "kcal", color: .orange)
                NutritionLabel(value: Int(item.totalProtein), unit: "g", label: "P", color: .red)
                NutritionLabel(value: Int(item.totalFat), unit: "g", label: "F", color: .yellow)
                NutritionLabel(value: Int(item.totalCarbohydrates), unit: "g", label: "C", color: .blue)
            }
            .font(.caption)
        }
        .padding()
        .background(Color(.secondarySystemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    private func recordFromMyMenu() async {
        isRecording = true

        let repository = MyMenuRepository()
        do {
            _ = try await repository.recordFromMyMenu(
                id: item.id,
                mealType: mealType,
                mealDate: mealDate
            )
            onRecorded()
        } catch {
            // TODO: エラー処理
        }

        isRecording = false
    }
}

// MARK: - Preview

#Preview {
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
            )
        ],
        totalCalories: 350,
        totalProtein: 10,
        totalFat: 5,
        totalCarbohydrates: 50,
        createdAt: Date(),
        updatedAt: Date()
    )

    return NavigationStack {
        MyMenuSelectionView(
            selectedMealType: .breakfast,
            mealDate: Date(),
            onRecorded: {}
        )
    }
}
