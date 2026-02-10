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
                                    onRecorded: onRecorded,
                                    repository: viewModel.repository
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
            .task {
                await viewModel.loadMyMenuList()
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
    let repository: MyMenuRepositoryProtocol

    @State private var isRecording = false
    @State private var errorMessage: String?

    init(item: MyMenuItem, mealType: MealType, mealDate: Date, onRecorded: @escaping () -> Void, repository: MyMenuRepositoryProtocol = MyMenuRepository()) {
        self.item = item
        self.mealType = mealType
        self.mealDate = mealDate
        self.onRecorded = onRecorded
        self.repository = repository
    }

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
        .alert("エラー", isPresented: Binding(
            get: { errorMessage != nil },
            set: { if !$0 { errorMessage = nil } }
        )) {
            Button("OK") {
                errorMessage = nil
            }
        } message: {
            Text(errorMessage ?? "")
        }
    }

    private func recordFromMyMenu() async {
        isRecording = true
        defer { isRecording = false }

        do {
            _ = try await repository.recordFromMyMenu(
                id: item.id,
                mealType: mealType,
                mealDate: mealDate
            )
            onRecorded()
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            #if DEBUG
            debugPrint("[MyMenuSelectionCard] Unexpected record error: \(error)")
            #endif
            errorMessage = "記録に失敗しました。ネットワークを確認してやり直してください。"
        }
    }
}

// MARK: - ErrorView

private struct ErrorView: View {
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

// MARK: - Preview

#Preview {
    NavigationStack {
        MyMenuSelectionView(
            selectedMealType: .breakfast,
            mealDate: Date(),
            onRecorded: {}
        )
    }
}
