import SwiftUI

// MARK: - ReceiptScanView

struct ReceiptScanView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var viewModel = ReceiptScanViewModel()

    let onSaved: ([Ingredient]) -> Void

    var body: some View {
        NavigationStack {
            Group {
                switch viewModel.phase {
                case .camera:
                    cameraPhase
                case .analyzing:
                    analyzingPhase
                case .review:
                    reviewPhase
                case .saving:
                    savingPhase
                }
            }
            .navigationTitle("レシートを撮影")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("キャンセル") { dismiss() }
                }
            }
        }
    }

    // MARK: - Phases

    private var cameraPhase: some View {
        CameraView { image in
            Task { await viewModel.analyzeReceipt(image: image) }
        }
    }

    private var analyzingPhase: some View {
        VStack(spacing: 24) {
            ProgressView()
                .scaleEffect(1.5)
            Text("レシートを解析中...")
                .font(.headline)
            Text("食材の情報を読み取っています")
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private var reviewPhase: some View {
        VStack(spacing: 0) {
            if let error = viewModel.errorMessage {
                errorBanner(error)
            }

            if viewModel.scannedIngredients.isEmpty {
                emptyResultView
            } else {
                ingredientReviewList
            }
        }
    }

    private var savingPhase: some View {
        VStack(spacing: 24) {
            ProgressView()
                .scaleEffect(1.5)
            Text("保存中...")
                .font(.headline)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    // MARK: - Review Subviews

    private var emptyResultView: some View {
        VStack(spacing: 16) {
            Image(systemName: "doc.text.magnifyingglass")
                .font(.system(size: 60))
                .foregroundStyle(.secondary)
            Text("食品が検出されませんでした")
                .font(.headline)
            Text("別のレシートを試すか、手動で食材を追加してください")
                .font(.subheadline)
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)
            Button("再撮影") {
                viewModel.retakePhoto()
            }
            .buttonStyle(.bordered)
        }
        .padding()
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private var ingredientReviewList: some View {
        List {
            Section {
                ForEach($viewModel.scannedIngredients) { $item in
                    ScannedIngredientRow(item: $item)
                }
                .onDelete { offsets in
                    viewModel.scannedIngredients.remove(atOffsets: offsets)
                }
            } header: {
                Text("検出された食材 (\(viewModel.scannedIngredients.count)件)")
            } footer: {
                Text("各項目をタップして編集できます。不要な食材はスワイプで削除してください。")
            }
        }
        .listStyle(.insetGrouped)
        .safeAreaInset(edge: .bottom) {
            saveButton
        }
    }

    private var saveButton: some View {
        Button {
            Task {
                let saved = await viewModel.saveAll()
                if !saved.isEmpty {
                    onSaved(saved)
                    dismiss()
                }
            }
        } label: {
            Text("一括保存 (\(viewModel.scannedIngredients.count)件)")
                .frame(maxWidth: .infinity)
        }
        .buttonStyle(.borderedProminent)
        .disabled(viewModel.scannedIngredients.isEmpty)
        .padding()
        .background(.regularMaterial)
    }

    private func errorBanner(_ message: String) -> some View {
        HStack {
            Image(systemName: "exclamationmark.triangle.fill")
                .foregroundStyle(.orange)
            Text(message)
                .font(.subheadline)
        }
        .padding()
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color.orange.opacity(0.1))
    }
}

// MARK: - ScannedIngredientRow

private struct ScannedIngredientRow: View {
    @Binding var item: ScannedIngredient

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            TextField("食材名", text: $item.name)
                .font(.body)

            HStack {
                Picker("カテゴリ", selection: $item.category) {
                    ForEach(IngredientCategory.allCases) { category in
                        Text(category.displayName).tag(category)
                    }
                }
                .pickerStyle(.menu)
                .font(.caption)

                Spacer()

                TextField("数量", value: $item.quantity, format: .number)
                    .keyboardType(.decimalPad)
                    .frame(width: 60)
                    .multilineTextAlignment(.trailing)
                    .font(.subheadline)

                Picker("単位", selection: $item.unit) {
                    ForEach(IngredientUnit.allCases) { unit in
                        Text(unit.rawValue).tag(unit.rawValue)
                    }
                }
                .pickerStyle(.menu)
                .font(.caption)
            }
        }
        .padding(.vertical, 4)
    }
}

// MARK: - ReceiptScanViewModel

@Observable
final class ReceiptScanViewModel {
    enum Phase {
        case camera
        case analyzing
        case review
        case saving
    }

    var phase: Phase = .camera
    var scannedIngredients: [ScannedIngredient] = []
    var errorMessage: String?

    private let repository: IngredientRepositoryProtocol

    init(repository: IngredientRepositoryProtocol = IngredientRepository()) {
        self.repository = repository
    }

    func analyzeReceipt(image: UIImage) async {
        phase = .analyzing
        errorMessage = nil

        guard let imageData = image.jpegData(compressionQuality: 0.8) else {
            errorMessage = "画像の変換に失敗しました"
            phase = .review
            return
        }

        do {
            scannedIngredients = try await repository.scanReceipt(imageData: imageData)
            phase = .review
        } catch let error as APIError {
            errorMessage = error.localizedDescription
            phase = .review
        } catch {
            errorMessage = "解析に失敗しました"
            phase = .review
        }
    }

    func retakePhoto() {
        scannedIngredients = []
        errorMessage = nil
        phase = .camera
    }

    func saveAll() async -> [Ingredient] {
        phase = .saving
        errorMessage = nil

        var saved: [Ingredient] = []

        for item in scannedIngredients {
            let request = CreateIngredientRequest(
                name: item.name,
                category: item.category.rawValue,
                quantity: item.quantity,
                unit: item.unit,
                purchaseDate: nil,
                expiryDate: nil,
                source: IngredientSource.receipt.rawValue
            )

            do {
                let ingredient = try await repository.createIngredient(request)
                saved.append(ingredient)
            } catch {
                // 個別の失敗はスキップして続行
            }
        }

        if saved.isEmpty, !scannedIngredients.isEmpty {
            errorMessage = "保存に失敗しました"
            phase = .review
        }

        return saved
    }
}
