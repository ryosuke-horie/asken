import SwiftUI

// MARK: - SuggestionRequestView

struct SuggestionRequestView: View {
    @State private var viewModel = CookingSuggestionViewModel()
    @State private var showingSuggestionList = false

    var body: some View {
        Form {
            Section("食事タイプ") {
                Picker("食事タイプ", selection: $viewModel.selectedMealType) {
                    ForEach(MealType.allCases) { mealType in
                        Label(mealType.displayName, systemImage: mealType.icon)
                            .tag(mealType)
                    }
                }
                .pickerStyle(.inline)
                .labelsHidden()
            }

            Section("提案数") {
                Picker("提案数", selection: $viewModel.suggestionCount) {
                    ForEach(1 ... 5, id: \.self) { num in
                        Text("\(num)件").tag(num)
                    }
                }
                .pickerStyle(.segmented)
            }

            Section {
                Button {
                    Task {
                        await viewModel.generateSuggestions()
                        if viewModel.errorMessage == nil {
                            showingSuggestionList = true
                        }
                    }
                } label: {
                    HStack {
                        Spacer()
                        if viewModel.isGenerating {
                            ProgressView()
                                .padding(.trailing, 8)
                            Text("生成中...")
                        } else {
                            Image(systemName: "sparkles")
                            Text("メニューを提案してもらう")
                        }
                        Spacer()
                    }
                }
                .disabled(viewModel.isGenerating)
            }

            if let error = viewModel.errorMessage {
                Section {
                    Text(error)
                        .foregroundStyle(.red)
                        .font(.caption)
                }
            }
        }
        .navigationTitle("メニューサジェスト")
        .navigationDestination(isPresented: $showingSuggestionList) {
            SuggestionListView(viewModel: viewModel)
        }
    }
}
