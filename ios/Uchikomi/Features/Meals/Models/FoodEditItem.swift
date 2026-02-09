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

    // 数値と単位を分離
    var quantityValue: String = ""
    var quantityUnit: MeasurementUnit?

    var hasNameChanged: Bool {
        !originalName.isEmpty && name != originalName
    }

    var hasUnitChanged: Bool {
        guard let originalUnit = QuantityParser.parseUnit(originalQuantity) else {
            // 元の値をパースできない場合、現在単位が選択されていれば変更とみなす
            return quantityUnit != nil
        }
        guard let current = quantityUnit else { return false }
        return originalUnit != current
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

        // quantityからquantityValueとquantityUnitをパースして設定
        // パース失敗時は元の文字列をそのまま使用（UIで表示・編集可能）
        quantityValue = QuantityParser.parseValue(quantity) ?? quantity
        quantityUnit = QuantityParser.parseUnit(quantity)
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
    /// カロリーは整数に、たんぱく質・脂質・炭水化物は小数点第1位に丸める。
    /// パース失敗時やゼロ除算の場合は何も変更しない（現在の栄養素値がそのまま残る）。
    /// キーストロークごとに呼ばれるため、入力途中のパース失敗は正常動作でありUIエラー表示は行わない。
    func recalculateNutrition() {
        guard let originalParsed = QuantityParser.parse(originalQuantity),
              let currentParsed = QuantityParser.parse(quantity),
              let ratio = QuantityParser.calculateRatio(from: originalParsed, to: currentParsed) else {
            return
        }

        calories = (originalCalories * ratio).rounded()
        protein = (originalProtein * ratio * 10).rounded() / 10
        fat = (originalFat * ratio * 10).rounded() / 10
        carbohydrates = (originalCarbohydrates * ratio * 10).rounded() / 10
    }

    /// quantityValueとquantityUnitからquantity文字列を生成
    func updateQuantityString() {
        guard let unit = quantityUnit, !quantityValue.isEmpty else {
            quantity = ""
            return
        }
        quantity = "\(quantityValue)\(unit.rawValue)"
    }
}
