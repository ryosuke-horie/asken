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
        !foods.isEmpty && foods.allSatisfy { !$0.name.isEmpty && !$0.isEstimating }
    }

    /// ローディング中のアイテムがあるかどうか
    var hasEstimatingItems: Bool {
        foods.contains { $0.isEstimating }
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

    /// 杯数を変更（即時計算、API呼び出しなし）
    func updateServingCount(for item: FoodEditItem, newCount: Int) {
        item.updateServingCount(newCount)
    }

    /// メニュー名を変更してLLMで栄養素を再推定
    func updateFoodName(for item: FoodEditItem, newName: String) async {
        guard !newName.isEmpty, newName != item.originalName else { return }

        item.name = newName
        item.isEstimating = true
        errorMessage = nil

        do {
            let response = try await repository.estimateNutrition(foodName: newName, quantity: 1)
            item.updateBaseNutrition(
                name: response.name,
                estimatedAmount: response.estimatedAmount,
                calories: response.caloriesKcal,
                protein: response.proteinG,
                fat: response.fatG,
                carbohydrates: response.carbohydratesG
            )
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "栄養素の取得に失敗しました"
        }

        item.isEstimating = false
    }
}
