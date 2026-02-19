import PhotosUI
import SwiftUI

// MARK: - MyMenuEditView

struct MyMenuEditView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var viewModel: MyMenuEditViewModel
    @State private var selectedItem: PhotosPickerItem?
    @State private var showingCamera = false

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

                            // 既存の栄養素がある場合は表示
                            if viewModel.totalCalories > 0, viewModel.analysisResult == nil {
                                NutritionSummaryCard(
                                    calories: viewModel.totalCalories,
                                    protein: viewModel.totalProtein,
                                    fat: viewModel.totalFat,
                                    carbohydrates: viewModel.totalCarbohydrates
                                )

                                Divider()

                                Text("現在の登録内容")
                                    .font(.subheadline)
                                    .foregroundStyle(.secondary)

                                // 現在の食品リスト（読み取り専用）
                                ForEach(viewModel.foodItems) { food in
                                    HStack {
                                        Text(food.name)
                                            .font(.subheadline)
                                        Text("(\(food.quantity))")
                                            .font(.caption)
                                            .foregroundStyle(.secondary)
                                        Spacer()
                                        Text("\(Int(food.calories)) kcal")
                                            .font(.caption)
                                            .foregroundStyle(Theme.primary)
                                    }
                                    .padding()
                                    .background(Color(.secondarySystemBackground))
                                    .clipShape(RoundedRectangle(cornerRadius: 10))
                                }
                            }

                            // 分析入力セクション（新規作成時のみ）
                            if !viewModel.isEditMode {
                                Divider()

                                Text("食事内容を入力")
                                    .font(.headline)
                                    .frame(maxWidth: .infinity, alignment: .leading)
                                    .padding(.horizontal)

                                // Manual Food Input Section
                                ManualFoodInputSection(
                                    foods: viewModel.manualFoods,
                                    onAdd: { viewModel.addManualFood() },
                                    onRemove: { viewModel.removeManualFood($0) }
                                )

                                // Image Selection Section
                                VStack(alignment: .leading, spacing: 8) {
                                    Text("または画像から入力")
                                        .font(.subheadline)
                                        .foregroundStyle(.secondary)

                                    ImageSelectionSection(
                                        selectedImage: viewModel.selectedImage,
                                        selectedItem: $selectedItem,
                                        showingCamera: $showingCamera
                                    )
                                }
                            }

                            // Analyze Button
                            if !viewModel.isEditMode {
                                Button {
                                    Task {
                                        await viewModel.analyze()
                                    }
                                } label: {
                                    if viewModel.isAnalyzing {
                                        HStack {
                                            ProgressView()
                                                .progressViewStyle(.circular)
                                                .tint(.white)
                                            Text("分析中...")
                                        }
                                    } else {
                                        Label("分析する", systemImage: "sparkles")
                                    }
                                }
                                .frame(maxWidth: .infinity)
                                .padding()
                                .background(viewModel.canAnalyze ? Theme.primary : Color.gray)
                                .foregroundStyle(.white)
                                .clipShape(RoundedRectangle(cornerRadius: 10))
                                .disabled(!viewModel.canAnalyze || viewModel.isAnalyzing)
                                .padding(.horizontal)
                            }

                            // Analysis Result
                            if let result = viewModel.analysisResult {
                                AnalysisResultSection(response: result)
                            }

                            // Error Message
                            if let error = viewModel.errorMessage {
                                Text(error)
                                    .font(.caption)
                                    .foregroundStyle(.red)
                                    .multilineTextAlignment(.center)
                                    .padding(.horizontal)
                            }
                        }
                        .padding(.vertical)
                    }
                    .scrollDismissesKeyboard(.immediately)
                    .safeAreaInset(edge: .bottom) {
                        Color.clear.frame(height: 0)
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
            .onChange(of: selectedItem) { _, newValue in
                Task {
                    guard let newValue else { return }
                    do {
                        guard let data = try await newValue.loadTransferable(type: Data.self) else {
                            viewModel.errorMessage = "画像データを取得できませんでした。別の画像を選択してください。"
                            return
                        }
                        guard let image = UIImage(data: data) else {
                            viewModel.errorMessage = "サポートされていない画像形式です。JPEGまたはPNGを使用してください。"
                            return
                        }
                        viewModel.selectedImage = image
                    } catch is CancellationError {
                        return
                    } catch {
                        #if DEBUG
                        debugPrint("[MyMenuEditView] Image load error: \(error)")
                        #endif
                        viewModel.errorMessage = "画像の読み込みに失敗しました。再度お試しください。"
                    }
                }
            }
            .fullScreenCover(isPresented: $showingCamera) {
                CameraView { image in
                    viewModel.selectedImage = image
                }
            }
        }
    }
}

// MARK: - ManualFoodInputSection

private struct ManualFoodInputSection: View {
    let foods: [FoodEditItem]
    let onAdd: () -> Void
    let onRemove: (FoodEditItem) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            ForEach(foods) { food in
                FoodItemEditRow(item: food) {
                    onRemove(food)
                }
            }

            Button(action: onAdd) {
                Label("メニューを追加", systemImage: "plus.circle")
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(Color(.secondarySystemBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 10))
            }
            .foregroundStyle(Theme.primary)
        }
        .padding(.horizontal)
    }
}

// MARK: - ImageSelectionSection

private struct ImageSelectionSection: View {
    let selectedImage: UIImage?
    @Binding var selectedItem: PhotosPickerItem?
    @Binding var showingCamera: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            if let image = selectedImage {
                Image(uiImage: image)
                    .resizable()
                    .scaledToFit()
                    .frame(maxHeight: 200)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
            }

            HStack(spacing: 12) {
                PhotosPicker(selection: $selectedItem, matching: .images) {
                    Label("写真を選択", systemImage: "photo.on.rectangle")
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color(.secondarySystemBackground))
                        .clipShape(RoundedRectangle(cornerRadius: 10))
                }

                Button {
                    showingCamera = true
                } label: {
                    Label("カメラ", systemImage: "camera")
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color(.secondarySystemBackground))
                        .clipShape(RoundedRectangle(cornerRadius: 10))
                }
            }
        }
        .padding(.horizontal)
    }
}

// MARK: - AnalysisResultSection

private struct AnalysisResultSection: View {
    let response: AnalysisResultResponse

    private var result: AnalysisResult {
        response.result
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("分析結果")
                .font(.headline)

            NutritionSummaryCard(
                calories: result.totalCalories,
                protein: result.totalProtein,
                fat: result.totalFat,
                carbohydrates: result.totalCarbohydrates,
                micronutrients: result.totalMicronutrients
            )

            VStack(alignment: .leading, spacing: 8) {
                Text("検出されたメニュー")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)

                ForEach(result.foods) { food in
                    HStack {
                        Text(food.name)
                        Text("(\(food.estimatedAmount))")
                            .foregroundStyle(.secondary)
                        Spacer()
                        Text("\(Int(food.caloriesKcal)) kcal")
                            .foregroundStyle(Theme.primary)
                    }
                    .font(.subheadline)
                }
            }
            .padding()
            .background(Color(.secondarySystemBackground))
            .clipShape(RoundedRectangle(cornerRadius: 10))
        }
        .padding(.horizontal)
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
            ),
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
