import Foundation

@Observable
final class NutritionEditorViewModel {
    var foods: [FoodEditItem] = []
    var isLoading = false
    var isSaving = false
    var errorMessage: String?
    var recalculatingMessage: String?
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
        !foods.isEmpty && foods.allSatisfy { food in
            !food.name.isEmpty &&
                !food.quantityValue.isEmpty
        }
    }

    var hasAnyNameChanged: Bool {
        foods.contains { $0.hasNameChanged }
    }

    var hasAnyUnitChanged: Bool {
        foods.contains { $0.hasUnitChanged }
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

        // 単位変更がある場合はメッセージを表示
        if hasAnyUnitChanged {
            recalculatingMessage = "栄養素を再計算中です..."
        }

        do {
            let updateFoods = foods.map { $0.toUpdateFoodItem() }
            let detail = try await repository.updateHistory(historyId: historyId, foods: updateFoods)
            // サーバー側で再計算された栄養素値を反映
            foods = detail.foods.map { FoodEditItem(from: $0) }
            isSaved = true
            recalculatingMessage = nil
        } catch let error as APIError {
            errorMessage = error.localizedDescription
            recalculatingMessage = nil
        } catch {
            errorMessage = "保存に失敗しました"
            recalculatingMessage = nil
        }

        isSaving = false
    }
}
