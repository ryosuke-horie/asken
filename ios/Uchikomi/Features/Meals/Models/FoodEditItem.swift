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

    // 杯数管理用
    var servingCount: Int
    var originalName: String
    var baseCalories: Double
    var baseProtein: Double
    var baseFat: Double
    var baseCarbohydrates: Double

    // メニュー名変更時のローディング状態
    var isEstimating: Bool = false

    init(
        id: UUID = UUID(),
        name: String = "",
        quantity: String = "",
        calories: Double = 0,
        protein: Double = 0,
        fat: Double = 0,
        carbohydrates: Double = 0,
        servingCount: Int = 1
    ) {
        self.id = id
        self.name = name
        self.originalName = name
        self.quantity = quantity
        self.calories = calories
        self.protein = protein
        self.fat = fat
        self.carbohydrates = carbohydrates
        self.servingCount = servingCount
        // 1杯あたりの基準値として保存
        self.baseCalories = calories
        self.baseProtein = protein
        self.baseFat = fat
        self.baseCarbohydrates = carbohydrates
    }

    convenience init(from nutritionInfo: NutritionInfo) {
        self.init(
            name: nutritionInfo.name,
            quantity: nutritionInfo.estimatedAmount,
            calories: nutritionInfo.caloriesKcal,
            protein: nutritionInfo.proteinG,
            fat: nutritionInfo.fatG,
            carbohydrates: nutritionInfo.carbohydratesG,
            servingCount: 1
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

    /// 杯数を変更して栄養素を再計算
    func updateServingCount(_ newCount: Int) {
        guard newCount >= 1 else { return }
        servingCount = newCount
        recalculateNutrition()
    }

    /// 基準栄養素から現在の栄養素を再計算
    func recalculateNutrition() {
        let multiplier = Double(servingCount)
        calories = baseCalories * multiplier
        protein = baseProtein * multiplier
        fat = baseFat * multiplier
        carbohydrates = baseCarbohydrates * multiplier
    }

    /// 新しい基準栄養素を設定して再計算
    func updateBaseNutrition(
        name: String,
        estimatedAmount: String,
        calories: Double,
        protein: Double,
        fat: Double,
        carbohydrates: Double
    ) {
        self.name = name
        self.originalName = name
        self.quantity = estimatedAmount
        self.baseCalories = calories
        self.baseProtein = protein
        self.baseFat = fat
        self.baseCarbohydrates = carbohydrates
        recalculateNutrition()
    }

    /// メニュー名が変更されたかどうか
    var isNameChanged: Bool {
        name != originalName
    }
}
