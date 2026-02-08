import Foundation

@Observable
final class FoodEditItem: Identifiable {
    let id: UUID
    var name: String
    var quantity: String
    var calories: Double
    var protein: Double
    var fat: Double
    var carbohydrates: Double

    // 元の値を保持（累積誤差防止のため、常に元の値をベースに比率計算する）
    let originalName: String
    let originalQuantity: String
    let originalCalories: Double
    let originalProtein: Double
    let originalFat: Double
    let originalCarbohydrates: Double

    var hasNameChanged: Bool {
        !originalName.isEmpty && name != originalName
    }

    init(
        id: UUID = UUID(),
        name: String = "",
        quantity: String = "",
        calories: Double = 0,
        protein: Double = 0,
        fat: Double = 0,
        carbohydrates: Double = 0
    ) {
        self.id = id
        self.name = name
        self.quantity = quantity
        self.calories = calories
        self.protein = protein
        self.fat = fat
        self.carbohydrates = carbohydrates
        self.originalName = name
        self.originalQuantity = quantity
        self.originalCalories = calories
        self.originalProtein = protein
        self.originalFat = fat
        self.originalCarbohydrates = carbohydrates
    }

    convenience init(from nutritionInfo: NutritionInfo) {
        self.init(
            name: nutritionInfo.name,
            quantity: nutritionInfo.estimatedAmount,
            calories: nutritionInfo.caloriesKcal,
            protein: nutritionInfo.proteinG,
            fat: nutritionInfo.fatG,
            carbohydrates: nutritionInfo.carbohydratesG
        )
    }

    func toUpdateFoodItem() -> UpdateFoodItem {
        UpdateFoodItem(
            name: name,
            estimatedAmount: quantity,
            caloriesKcal: calories,
            proteinG: protein,
            fatG: fat,
            carbohydratesG: carbohydrates
        )
    }

    /// 量の変更に基づいて栄養素を再計算する。
    /// 元の量と現在の量をパースし、同じ単位の場合のみ比率で再計算する。
    /// パース失敗時やゼロ除算の場合は元の値を維持する。
    func recalculateNutrition() {
        guard let originalParsed = QuantityParser.parse(originalQuantity),
              let currentParsed = QuantityParser.parse(quantity),
              let ratio = QuantityParser.calculateRatio(from: originalParsed, to: currentParsed)
        else { return }

        calories = (originalCalories * ratio).rounded()
        protein = (originalProtein * ratio * 10).rounded() / 10
        fat = (originalFat * ratio * 10).rounded() / 10
        carbohydrates = (originalCarbohydrates * ratio * 10).rounded() / 10
    }
}
