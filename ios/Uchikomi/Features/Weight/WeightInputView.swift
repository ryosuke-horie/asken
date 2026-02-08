import SwiftUI

struct WeightInputView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var viewModel: WeightInputViewModel
    @FocusState private var isWeightFocused: Bool

    let onSaved: () -> Void

    private let quickNotes = ["起床時", "練習前", "練習後", "就寝前"]

    init(
        editingRecord: WeightRecord? = nil,
        lastWeight _: Double? = nil,
        repository: WeightRepositoryProtocol = WeightRepository(),
        onSaved: @escaping () -> Void
    ) {
        _viewModel = State(
            initialValue: WeightInputViewModel(editingRecord: editingRecord, repository: repository)
        )
        self.onSaved = onSaved
    }

    var body: some View {
        NavigationStack {
            VStack(spacing: 20) {
                // 体重入力
                VStack(spacing: 12) {
                    Text("体重")
                        .font(.headline)
                        .frame(maxWidth: .infinity, alignment: .leading)

                    HStack(spacing: 16) {
                        Button {
                            viewModel.decrementWeight()
                        } label: {
                            Image(systemName: "minus.circle.fill")
                                .font(.title)
                                .foregroundStyle(Theme.primary)
                        }

                        HStack(alignment: .firstTextBaseline, spacing: 4) {
                            TextField("65.0", text: $viewModel.weightText)
                                .keyboardType(.decimalPad)
                                .font(.system(size: 48, weight: .bold))
                                .multilineTextAlignment(.center)
                                .frame(width: 160)
                                .focused($isWeightFocused)

                            Text("kg")
                                .font(.title2)
                                .foregroundStyle(.secondary)
                        }

                        Button {
                            viewModel.incrementWeight()
                        } label: {
                            Image(systemName: "plus.circle.fill")
                                .font(.title)
                                .foregroundStyle(Theme.primary)
                        }
                    }
                }

                // クイックノート
                VStack(spacing: 8) {
                    Text("メモ")
                        .font(.headline)
                        .frame(maxWidth: .infinity, alignment: .leading)

                    ScrollView(.horizontal, showsIndicators: false) {
                        HStack(spacing: 8) {
                            ForEach(quickNotes, id: \.self) { note in
                                Button(note) {
                                    viewModel.setQuickNote(note)
                                }
                                .buttonStyle(.bordered)
                                .tint(viewModel.memo == note ? Theme.primary : .secondary)
                            }
                        }
                    }

                    TextField("メモを入力", text: $viewModel.memo)
                        .textFieldStyle(.roundedBorder)
                }

                if let errorMessage = viewModel.errorMessage {
                    Text(errorMessage)
                        .foregroundStyle(.red)
                        .font(.caption)
                }

                Spacer()

                // 削除ボタン（編集モード時）
                if viewModel.isEditing {
                    Button(role: .destructive) {
                        Task {
                            await viewModel.delete()
                        }
                    } label: {
                        Label("この記録を削除", systemImage: "trash")
                            .frame(maxWidth: .infinity)
                    }
                    .buttonStyle(.bordered)
                }
            }
            .padding()
            .navigationTitle(viewModel.isEditing ? "体重を編集" : "体重を記録")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("キャンセル") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("保存") {
                        Task { await viewModel.save() }
                    }
                    .disabled(!viewModel.isValid || viewModel.isSaving)
                }
            }
            .onChange(of: viewModel.didSave) { _, didSave in
                if didSave {
                    onSaved()
                    dismiss()
                }
            }
            .onAppear {
                isWeightFocused = true
            }
        }
    }
}
