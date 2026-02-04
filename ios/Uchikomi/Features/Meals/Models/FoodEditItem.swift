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

    // 元の値（再計算のベース）
    private(set) var originalQuantity: String
    private(set) var originalCalories: Double
    private(set) var originalProtein: Double
    private(set) var originalFat: Double
    private(set) var originalCarbohydrates: Double

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

        // 元の値を保持
        self.originalQuantity = quantity
        self.originalCalories = calories
        self.originalProtein = protein
        self.originalFat = fat
        self.originalCarbohydrates = carbohydrates
    }

    func recalculateNutrition() {
        // 元の量と現在の量をパース
        guard let originalParsed = QuantityParser.parse(originalQuantity),
              let currentParsed = QuantityParser.parse(quantity) else {
            return
        }

        // 比率を計算
        guard let ratio = QuantityParser.calculateRatio(from: originalParsed, to: currentParsed) else {
            return
        }

        // 栄養素を再計算（元の値をベースに、小数点以下1桁で丸める）
        calories = round(originalCalories * ratio * 10) / 10
        protein = round(originalProtein * ratio * 10) / 10
        fat = round(originalFat * ratio * 10) / 10
        carbohydrates = round(originalCarbohydrates * ratio * 10) / 10
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
}
