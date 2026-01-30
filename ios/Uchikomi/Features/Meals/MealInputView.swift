import SwiftUI
import PhotosUI

struct MealInputView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var viewModel = MealInputViewModel()
    @State private var selectedItem: PhotosPickerItem?
    @State private var showingCamera = false

    let mealDate: Date
    let initialMealType: MealType
    let onSaved: () -> Void

    init(mealDate: Date = Date(), initialMealType: MealType = .lunch, onSaved: @escaping () -> Void = {}) {
        self.mealDate = mealDate
        self.initialMealType = initialMealType
        self.onSaved = onSaved
    }

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 24) {
                    // Meal Type Picker
                    MealTypePicker(selectedType: $viewModel.selectedMealType)

                    // Image Selection
                    ImageSelectionSection(
                        selectedImage: viewModel.selectedImage,
                        selectedItem: $selectedItem,
                        showingCamera: $showingCamera
                    )

                    // Analysis Button
                    if viewModel.selectedImage != nil && viewModel.analysisResult == nil {
                        Button {
                            Task {
                                await viewModel.analyzeImage()
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
                                Label("画像を分析", systemImage: "sparkles")
                            }
                        }
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Theme.primary)
                        .foregroundStyle(.white)
                        .clipShape(RoundedRectangle(cornerRadius: 10))
                        .disabled(viewModel.isAnalyzing)
                    }

                    // Analysis Result
                    if let result = viewModel.analysisResult {
                        AnalysisResultSection(result: result)
                    }

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
            .navigationTitle("食事を記録")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("キャンセル") {
                        dismiss()
                    }
                }
            }
            .onChange(of: selectedItem) { _, newValue in
                Task {
                    if let data = try? await newValue?.loadTransferable(type: Data.self),
                       let image = UIImage(data: data) {
                        viewModel.selectedImage = image
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
        }
        .onAppear {
            viewModel.mealDate = mealDate
            viewModel.selectedMealType = initialMealType
        }
    }
}

// MARK: - Subviews

private struct MealTypePicker: View {
    @Binding var selectedType: MealType

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("食事タイプ")
                .font(.headline)

            HStack(spacing: 12) {
                ForEach(MealType.allCases, id: \.self) { type in
                    MealTypeButton(
                        type: type,
                        isSelected: selectedType == type
                    ) {
                        selectedType = type
                    }
                }
            }
        }
    }
}

private struct MealTypeButton: View {
    let type: MealType
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            VStack(spacing: 4) {
                Image(systemName: type.icon)
                    .font(.title3)
                Text(type.displayName)
                    .font(.caption)
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, 12)
            .background(isSelected ? Theme.primary : Color(.secondarySystemBackground))
            .foregroundStyle(isSelected ? .white : .primary)
            .clipShape(RoundedRectangle(cornerRadius: 10))
        }
    }
}

private struct ImageSelectionSection: View {
    let selectedImage: UIImage?
    @Binding var selectedItem: PhotosPickerItem?
    @Binding var showingCamera: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("食事の画像")
                .font(.headline)

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

private struct AnalysisResultSection: View {
    let result: AnalysisResultResponse

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
                Text("検出された食品")
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

// MARK: - Camera View

struct CameraView: UIViewControllerRepresentable {
    @Environment(\.dismiss) private var dismiss
    let onImageCaptured: (UIImage) -> Void

    func makeUIViewController(context: Context) -> UIImagePickerController {
        let picker = UIImagePickerController()
        picker.sourceType = .camera
        picker.delegate = context.coordinator
        return picker
    }

    func updateUIViewController(_ uiViewController: UIImagePickerController, context: Context) {}

    func makeCoordinator() -> Coordinator {
        Coordinator(self)
    }

    class Coordinator: NSObject, UIImagePickerControllerDelegate, UINavigationControllerDelegate {
        let parent: CameraView

        init(_ parent: CameraView) {
            self.parent = parent
        }

        func imagePickerController(
            _ picker: UIImagePickerController,
            didFinishPickingMediaWithInfo info: [UIImagePickerController.InfoKey: Any]
        ) {
            if let image = info[.originalImage] as? UIImage {
                parent.onImageCaptured(image)
            }
            parent.dismiss()
        }

        func imagePickerControllerDidCancel(_ picker: UIImagePickerController) {
            parent.dismiss()
        }
    }
}

#Preview {
    MealInputView()
}
