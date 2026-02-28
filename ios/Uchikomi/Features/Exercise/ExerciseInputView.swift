import SwiftUI

// MARK: - ExerciseInputView

struct ExerciseInputView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var viewModel: ExerciseInputViewModel
    @FocusState private var isNameFocused: Bool

    let onSaved: () -> Void

    init(
        recordedDate: String,
        repository: ExerciseRepositoryProtocol = ExerciseRepository(),
        onSaved: @escaping () -> Void
    ) {
        _viewModel = State(
            initialValue: ExerciseInputViewModel(recordedDate: recordedDate, repository: repository)
        )
        self.onSaved = onSaved
    }

    var body: some View {
        NavigationStack {
            VStack(spacing: 24) {
                // 種目名
                VStack(spacing: 8) {
                    Text("種目名")
                        .font(.headline)
                        .frame(maxWidth: .infinity, alignment: .leading)

                    TextField("例: 柔術", text: $viewModel.exerciseName)
                        .textFieldStyle(.roundedBorder)
                        .focused($isNameFocused)

                    ScrollView(.horizontal, showsIndicators: false) {
                        HStack(spacing: 8) {
                            ForEach(ExerciseInputViewModel.quickPickNames, id: \.self) { name in
                                Button(name) {
                                    viewModel.exerciseName = name
                                    isNameFocused = false
                                }
                                .buttonStyle(.bordered)
                                .tint(viewModel.exerciseName == name ? Theme.primary : .secondary)
                            }
                        }
                    }
                }

                // 時間（分）
                VStack(spacing: 12) {
                    Text("時間")
                        .font(.headline)
                        .frame(maxWidth: .infinity, alignment: .leading)

                    HStack(spacing: 16) {
                        Button {
                            viewModel.decrementDuration()
                        } label: {
                            Image(systemName: "minus.circle.fill")
                                .font(.title)
                                .foregroundStyle(Theme.primary)
                        }

                        HStack(alignment: .firstTextBaseline, spacing: 4) {
                            TextField("60", text: $viewModel.durationText)
                                .keyboardType(.numberPad)
                                .font(.system(size: 48, weight: .bold))
                                .multilineTextAlignment(.center)
                                .frame(width: 120)

                            Text("分")
                                .font(.title2)
                                .foregroundStyle(.secondary)
                        }

                        Button {
                            viewModel.incrementDuration()
                        } label: {
                            Image(systemName: "plus.circle.fill")
                                .font(.title)
                                .foregroundStyle(Theme.primary)
                        }
                    }
                }

                if let errorMessage = viewModel.errorMessage {
                    Text(errorMessage)
                        .foregroundStyle(.red)
                        .font(.caption)
                        .frame(maxWidth: .infinity, alignment: .leading)
                }

                Spacer()
            }
            .padding()
            .navigationTitle("運動を記録")
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
                isNameFocused = true
            }
        }
    }
}
