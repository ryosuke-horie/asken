import Foundation

@Observable
final class NutritionEditorViewModel {
    var foods: [FoodEditItem] = []
    var isLoading = false
    var isSaving = false
    var errorMessage: String?
    var isSaved = false

    private let historyId: String?
    private let repository: MealRepositoryProtocol

    var totalCalories: Double {
        foods.reduce(0) { $0 + $1.calories }
    }

    var totalProtein: Double {
        foods.reduce(0) { $0 + $1.protein }
    }

    var totalFat: Double {
        foods.reduce(0) { $0 + $1.fat }
    }

    var totalCarbohydrates: Double {
        foods.reduce(0) { $0 + $1.carbohydrates }
    }

    var canSave: Bool {
        !foods.isEmpty && foods.allSatisfy { !$0.name.isEmpty }
    }

    init(
        historyId: String? = nil,
        foods: [NutritionInfo] = [],
        repository: MealRepositoryProtocol = MealRepository()
    ) {
        self.historyId = historyId
        self.repository = repository
        self.foods = foods.map { FoodEditItem(from: $0) }
    }

    func loadFromHistory() async {
        guard let historyId else { return }

        isLoading = true
        errorMessage = nil

        do {
            let detail = try await repository.getHistoryDetail(id: historyId)
            foods = detail.foods.map { FoodEditItem(from: $0) }
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "データの取得に失敗しました"
        }

        isLoading = false
    }

    func addFood() {
        foods.append(FoodEditItem())
    }

    func removeFood(_ item: FoodEditItem) {
        foods.removeAll { $0.id == item.id }
    }

    func save() async {
        guard let historyId, canSave else { return }

        isSaving = true
        errorMessage = nil

        do {
            let updateFoods = foods.map { $0.toUpdateFoodItem() }
            _ = try await repository.updateHistory(historyId: historyId, foods: updateFoods)
            isSaved = true
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "保存に失敗しました"
        }

        isSaving = false
    }
}
