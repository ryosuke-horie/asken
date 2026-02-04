import Foundation
import Testing
@testable import Uchikomi

@Suite
struct FoodEditItemTests {
    @Test
    func 初期化時に基準値が設定されるべき() {
        let item = FoodEditItem(
            name: "ラーメン",
            quantity: "1杯",
            calories: 500,
            protein: 20,
            fat: 15,
            carbohydrates: 60
        )

        #expect(item.baseCalories == 500)
        #expect(item.baseProtein == 20)
        #expect(item.baseFat == 15)
        #expect(item.baseCarbohydrates == 60)
        #expect(item.servingCount == 1)
        #expect(item.originalName == "ラーメン")
    }

    @Test
    func 杯数を2倍にすると栄養素も2倍になるべき() {
        let item = FoodEditItem(
            name: "ラーメン",
            quantity: "1杯",
            calories: 500,
            protein: 20,
            fat: 15,
            carbohydrates: 60
        )

        item.updateServingCount(2)

        #expect(item.servingCount == 2)
        #expect(item.calories == 1_000)
        #expect(item.protein == 40)
        #expect(item.fat == 30)
        #expect(item.carbohydrates == 120)
    }

    @Test
    func 杯数を3倍にすると栄養素も3倍になるべき() {
        let item = FoodEditItem(
            name: "ご飯",
            quantity: "1杯",
            calories: 234,
            protein: 3.5,
            fat: 0.5,
            carbohydrates: 51.3
        )

        item.updateServingCount(3)

        #expect(item.servingCount == 3)
        #expect(item.calories == 702)
        #expect(item.protein == 10.5)
        #expect(item.fat == 1.5)
        #expect(item.carbohydrates == 153.9)
    }

    @Test
    func 杯数を1未満にしても1のままであるべき() {
        let item = FoodEditItem(
            name: "ラーメン",
            quantity: "1杯",
            calories: 500,
            protein: 20,
            fat: 15,
            carbohydrates: 60
        )

        item.updateServingCount(0)

        #expect(item.servingCount == 1)
        #expect(item.calories == 500)
    }

    @Test
    func 基準栄養素を更新すると杯数に応じて再計算されるべき() {
        let item = FoodEditItem(
            name: "ラーメン",
            quantity: "1杯",
            calories: 500,
            protein: 20,
            fat: 15,
            carbohydrates: 60
        )

        // 先に杯数を2に変更
        item.updateServingCount(2)

        // 新しい基準栄養素に更新
        item.updateBaseNutrition(
            name: "味噌ラーメン",
            estimatedAmount: "1杯",
            calories: 600,
            protein: 25,
            fat: 20,
            carbohydrates: 70
        )

        // 杯数2なので2倍になるべき
        #expect(item.name == "味噌ラーメン")
        #expect(item.originalName == "味噌ラーメン")
        #expect(item.baseCalories == 600)
        #expect(item.calories == 1_200)
        #expect(item.protein == 50)
        #expect(item.fat == 40)
        #expect(item.carbohydrates == 140)
    }

    @Test
    func メニュー名変更検出が正しく動作するべき() {
        let item = FoodEditItem(
            name: "ラーメン",
            quantity: "1杯",
            calories: 500,
            protein: 20,
            fat: 15,
            carbohydrates: 60
        )

        #expect(item.isNameChanged == false)

        item.name = "味噌ラーメン"
        #expect(item.isNameChanged == true)
    }

    @Test
    func NutritionInfoから初期化できるべき() {
        let nutritionInfo = NutritionInfo(
            name: "鶏むね肉",
            estimatedAmount: "100g",
            caloriesKcal: 165,
            proteinG: 31,
            fatG: 3.6,
            carbohydratesG: 0,
            servingCount: 1
        )

        let item = FoodEditItem(from: nutritionInfo)

        #expect(item.name == "鶏むね肉")
        #expect(item.quantity == "100g")
        #expect(item.calories == 165)
        #expect(item.protein == 31)
        #expect(item.fat == 3.6)
        #expect(item.carbohydrates == 0)
        #expect(item.servingCount == 1)
    }

    @Test
    func NutritionInfoから杯数付きで初期化できるべき() {
        // サーバーから取得した2人前のデータ（栄養素は計算後の値）
        let nutritionInfo = NutritionInfo(
            name: "ラーメン",
            estimatedAmount: "1杯",
            caloriesKcal: 1_000, // 500 × 2
            proteinG: 40, // 20 × 2
            fatG: 30, // 15 × 2
            carbohydratesG: 120, // 60 × 2
            servingCount: 2
        )

        let item = FoodEditItem(from: nutritionInfo)

        // 杯数と計算後の栄養素が正しく設定されること
        #expect(item.servingCount == 2)
        #expect(item.calories == 1_000)
        #expect(item.protein == 40)
        // 基準値（1人前あたり）が逆算されること
        #expect(item.baseCalories == 500)
        #expect(item.baseProtein == 20)
        #expect(item.baseFat == 15)
        #expect(item.baseCarbohydrates == 60)
    }

    @Test
    func UpdateFoodItemに変換できるべき() {
        let item = FoodEditItem(
            name: "ラーメン",
            quantity: "1杯",
            calories: 500,
            protein: 20,
            fat: 15,
            carbohydrates: 60
        )

        item.updateServingCount(2)

        let updateItem = item.toUpdateFoodItem()

        #expect(updateItem.name == "ラーメン")
        #expect(updateItem.estimatedAmount == "1杯")
        #expect(updateItem.caloriesKcal == 1_000) // 杯数2なので2倍
        #expect(updateItem.proteinG == 40)
        #expect(updateItem.fatG == 30)
        #expect(updateItem.carbohydratesG == 120)
        #expect(updateItem.servingCount == 2) // 杯数も保存される
    }
}
