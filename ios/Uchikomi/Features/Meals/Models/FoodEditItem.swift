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
