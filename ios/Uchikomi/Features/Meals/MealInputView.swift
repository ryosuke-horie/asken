import PhotosUI
import SwiftUI

// MARK: - URL + @retroactive Identifiable

extension URL: @retroactive Identifiable {
    public var id: String {
        absoluteString
    }
}

// MARK: - MealInputView

struct MealInputView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var viewModel = MealInputViewModel()
    @State private var selectedItem: PhotosPickerItem?
    @State private var showingCamera = false
    @State private var editingMeal: HistoryDetail?
    @State private var deletingMeal: HistoryDetail?
    @State private var showingImagePreview: URL?
    @State private var showingSkipConfirmation = false

    let mealDate: Date
    let initialMealType: MealType
    let existingMeals: [HistoryDetail]
    let onSaved: () -> Void

    private var canAnalyze: Bool {
        viewModel.selectedImage != nil || viewModel.hasValidManualInput
    }

    init(
        mealDate: Date = Date(),
        initialMealType: MealType = .lunch,
        existingMeals: [HistoryDetail] = [],
        onSaved: @escaping () -> Void = {}
    ) {
        self.mealDate = mealDate
        self.initialMealType = initialMealType
        self.existingMeals = existingMeals
        self.onSaved = onSaved
    }

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 24) {
                    // Existing Meals Section
                    if !existingMeals.isEmpty {
                        ExistingMealsSection(
                            meals: existingMeals,
                            onEdit: { meal in editingMeal = meal },
                            onDelete: { meal in deletingMeal = meal },
                            onImageTap: { url in showingImagePreview = url }
                        )
                    }

                    Divider()
                        .padding(.horizontal)

                    // New Meal Section
                    Text("新しく記録する")
                        .font(.headline)
                        .frame(maxWidth: .infinity, alignment: .leading)

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

                    // Unified Analyze Button
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
                    .background(canAnalyze ? Theme.primary : Color.gray)
                    .foregroundStyle(.white)
                    .clipShape(RoundedRectangle(cornerRadius: 10))
                    .disabled(!canAnalyze || viewModel.isAnalyzing)

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
                    }

                    // Skip Meal Option
                    if existingMeals.isEmpty {
                        Divider()

                        Button {
                            showingSkipConfirmation = true
                        } label: {
                            Label("食べなかった", systemImage: "moon.zzz")
                                .frame(maxWidth: .infinity)
                                .padding()
                        }
                        .buttonStyle(.bordered)
                        .tint(.secondary)
                        .disabled(viewModel.isSkipping)
                    }
                }
                .padding()
            }
            .navigationTitle("\(initialMealType.displayName)を記録")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("閉じる") {
                        dismiss()
                    }
                }
            }
            .onChange(of: selectedItem) { _, newValue in
                Task {
                    guard let newValue else { return }
                    do {
                        guard let data = try await newValue.loadTransferable(type: Data.self) else {
                            viewModel.errorMessage = "画像を読み込めませんでした"
                            return
                        }
                        guard let image = UIImage(data: data) else {
                            viewModel.errorMessage = "サポートされていない画像形式です"
                            return
                        }
                        viewModel.selectedImage = image
                    } catch {
                        #if DEBUG
                        debugPrint("[MealInputView] Image load error: \(error)")
                        #endif
                        viewModel.errorMessage = "画像の読み込みに失敗しました"
                    }
                }
            }
            .onChange(of: viewModel.isCompleted) { _, isCompleted in
                if isCompleted {
                    onSaved()
                    dismiss()
                }
            }
            .fullScreenCover(isPresented: $showingCamera) {
                CameraView { image in
                    viewModel.selectedImage = image
                }
            }
            .sheet(isPresented: $viewModel.showEditor) {
                if let analysisId = viewModel.analysisId,
                   let result = viewModel.analysisResult {
                    NutritionEditorView(
                        historyId: analysisId,
                        foods: result.result.foods
                    ) {
                        viewModel.markCompleted()
                    }
                }
            }
            .sheet(item: $editingMeal) { meal in
                NutritionEditorView(
                    historyId: meal.id,
                    foods: meal.foods
                ) {
                    onSaved()
                }
            }
            .sheet(item: $showingImagePreview) { url in
                ImagePreviewView(imageURL: url)
            }
            .alert("削除の確認", isPresented: Binding(
                get: { deletingMeal != nil },
                set: { if !$0 { deletingMeal = nil } }
            )) {
                Button("キャンセル", role: .cancel) {
                    deletingMeal = nil
                }
                Button("削除", role: .destructive) {
                    if let meal = deletingMeal {
                        Task {
                            let success = await viewModel.deleteHistory(id: meal.id)
                            if success {
                                onSaved()
                            }
                        }
                    }
                    deletingMeal = nil
                }
            } message: {
                Text("この食事記録を削除しますか？")
            }
            .alert("確認", isPresented: $showingSkipConfirmation) {
                Button("キャンセル", role: .cancel) {}
                Button("食べなかった") {
                    Task {
                        let success = await viewModel.skipMeal()
                        if success {
                            onSaved()
                            dismiss()
                        }
                    }
                }
            } message: {
                Text("\(initialMealType.displayName)を「食べなかった」として記録しますか？")
            }
        }
        .onAppear {
            viewModel.mealDate = mealDate
            viewModel.selectedMealType = initialMealType
        }
    }
}

// MARK: - ExistingMealsSection

private struct ExistingMealsSection: View {
    let meals: [HistoryDetail]
    let onEdit: (HistoryDetail) -> Void
    let onDelete: (HistoryDetail) -> Void
    let onImageTap: (URL) -> Void

    private var skippedMeals: [HistoryDetail] {
        meals.filter { $0.inputType == .skipped }
    }

    private var normalMeals: [HistoryDetail] {
        meals.filter { $0.inputType != .skipped }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("登録済みの記録")
                .font(.headline)

            ForEach(skippedMeals) { meal in
                SkippedMealCard(onDelete: { onDelete(meal) })
            }

            ForEach(normalMeals) { meal in
                ExistingMealCard(
                    meal: meal,
                    onEdit: { onEdit(meal) },
                    onDelete: { onDelete(meal) },
                    onImageTap: onImageTap
                )
            }
        }
    }
}

// MARK: - SkippedMealCard

private struct SkippedMealCard: View {
    let onDelete: () -> Void

    var body: some View {
        HStack {
            Image(systemName: "moon.zzz")
                .foregroundStyle(.secondary)
            Text("食べませんでした")
                .font(.subheadline)
                .foregroundStyle(.secondary)
            Spacer()
            Button(action: onDelete) {
                Label("取り消す", systemImage: "arrow.uturn.backward")
                    .font(.caption)
            }
            .buttonStyle(.bordered)
            .tint(.secondary)
        }
        .padding()
        .background(Color(.secondarySystemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

// MARK: - ExistingMealCard

private struct ExistingMealCard: View {
    let meal: HistoryDetail
    let onEdit: () -> Void
    let onDelete: () -> Void
    let onImageTap: (URL) -> Void

    private var imageURL: URL? {
        guard let imagePath = meal.imagePath, !imagePath.isEmpty else { return nil }
        // imagePath is stored as "uploads/xxx.jpg", extract just the filename
        let filename = (imagePath as NSString).lastPathComponent
        guard ImageFilenameValidator.isValid(filename) else { return nil }
        let baseURL = AppEnvironment.current.baseURL
        return baseURL.appendingPathComponent("api/images/\(filename)")
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            // Image preview if available
            if let url = imageURL {
                AsyncImage(url: url) { phase in
                    switch phase {
                    case .empty:
                        ProgressView()
                            .frame(height: 120)
                    case let .success(image):
                        image
                            .resizable()
                            .scaledToFill()
                            .frame(height: 120)
                            .clipped()
                            .clipShape(RoundedRectangle(cornerRadius: 8))
                            .onTapGesture {
                                onImageTap(url)
                            }
                    case .failure:
                        Image(systemName: "photo")
                            .frame(height: 120)
                            .foregroundStyle(.secondary)
                    @unknown default:
                        EmptyView()
                    }
                }
            }

            // Food list
            ForEach(meal.foods) { food in
                HStack {
                    Text(food.name)
                        .font(.subheadline)
                    Text("(\(food.estimatedAmount))")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                    Spacer()
                    Text("\(Int(food.caloriesKcal)) kcal")
                        .font(.caption)
                        .foregroundStyle(Theme.primary)
                }
            }

            // Total and action buttons
            HStack {
                Text("合計: \(Int(meal.totalCalories)) kcal")
                    .font(.subheadline)
                    .fontWeight(.medium)
                    .foregroundStyle(Theme.primary)

                Spacer()

                Button(action: onEdit) {
                    Label("編集", systemImage: "pencil")
                        .font(.caption)
                }
                .buttonStyle(.bordered)
                .tint(Theme.primary)

                Button(action: onDelete) {
                    Label("削除", systemImage: "trash")
                        .font(.caption)
                }
                .buttonStyle(.bordered)
                .tint(.red)
            }
        }
        .padding()
        .background(Color(.secondarySystemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

// MARK: - ImagePreviewView

private struct ImagePreviewView: View {
    @Environment(\.dismiss) private var dismiss
    let imageURL: URL

    var body: some View {
        NavigationStack {
            AsyncImage(url: imageURL) { phase in
                switch phase {
                case .empty:
                    ProgressView()
                case let .success(image):
                    image
                        .resizable()
                        .scaledToFit()
                case .failure:
                    VStack {
                        Image(systemName: "photo")
                            .font(.largeTitle)
                        Text("画像を読み込めませんでした")
                    }
                    .foregroundStyle(.secondary)
                @unknown default:
                    EmptyView()
                }
            }
            .navigationTitle("画像プレビュー")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("閉じる") {
                        dismiss()
                    }
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
            Text("食事内容を入力")
                .font(.subheadline)
                .foregroundStyle(.secondary)

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
                carbohydrates: result.totalCarbohydrates
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
    }
}

#Preview {
    MealInputView()
}
