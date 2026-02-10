import SwiftUI

// MARK: - MyMenuListView

struct MyMenuListView: View {
    @State private var viewModel = MyMenuListViewModel()
    @State private var selectedMenuItem: MyMenuItem?
    @State private var showingCreateSheet = false

    var body: some View {
        Group {
            if viewModel.isLoading {
                ProgressView()
                    .frame(maxWidth: .infinity, alignment: .center)
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
                } actions: {
                    Button("マイメニューを追加") {
                        showingCreateSheet = true
                    }
                }
            } else {
                List {
                    ForEach(viewModel.myMenuItems) { item in
                        NavigationLink {
                            MyMenuEditView(menuItem: item)
                        } label: {
                            MyMenuRow(item: item)
                        }
                    }
                    .onDelete(perform: deleteMenus)
                }
                .listStyle(.plain)
            }
        }
        .navigationTitle("マイメニュー")
        .navigationBarTitleDisplayMode(.large)
        .toolbar {
            ToolbarItem(placement: .navigationBarTrailing) {
                Button {
                    showingCreateSheet = true
                } label: {
                    Image(systemName: "plus")
                }
            }
        }
        .sheet(isPresented: $showingCreateSheet) {
            MyMenuEditView()
        }
        .task {
            await viewModel.loadMyMenuList()
        }
    }

    private func deleteMenus(at offsets: IndexSet) {
        Task {
            for index in offsets {
                let item = viewModel.myMenuItems[index]
                await viewModel.deleteMyMenu(id: item.id)
            }
        }
    }
}

// MARK: - MyMenuRow

struct MyMenuRow: View {
    let item: MyMenuItem

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(item.name)
                .font(.headline)

            Text(item.foods.map { $0.name }.joined(separator: ", "))
                .font(.caption)
                .foregroundStyle(.secondary)
                .lineLimit(2)

            HStack(spacing: 16) {
                NutritionLabel(value: Int(item.totalCalories), unit: "kcal", color: .orange)
                NutritionLabel(value: Int(item.totalProtein), unit: "g", label: "P", color: .red)
                NutritionLabel(value: Int(item.totalFat), unit: "g", label: "F", color: .yellow)
                NutritionLabel(value: Int(item.totalCarbohydrates), unit: "g", label: "C", color: .blue)
            }
            .font(.caption)
        }
        .padding(.vertical, 4)
    }
}

// MARK: - NutritionLabel

struct NutritionLabel: View {
    let value: Int
    let unit: String
    let label: String?
    let color: Color

    init(value: Int, unit: String, label: String? = nil, color: Color) {
        self.value = value
        self.unit = unit
        self.label = label
        self.color = color
    }

    var body: some View {
        HStack(spacing: 2) {
            if let label {
                Text(label)
            }
            Text("\(value)\(unit)")
        }
        .foregroundStyle(color)
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
        MyMenuListView()
    }
}
