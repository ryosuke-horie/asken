import Testing
@testable import Uchikomi

@Suite struct FoodEditItemTests {
    // MARK: - 元の値保持テスト

    @Test func 初期化時にoriginal値が正しく設定されるべき() {
        let item = FoodEditItem(
            name: "白米",
            quantity: "100g",
            calories: 168,
            protein: 2.5,
            fat: 0.3,
            carbohydrates: 37.1
        )

        #expect(item.originalName == "白米")
        #expect(item.originalQuantity == "100g")
        #expect(item.originalCalories == 168)
        #expect(item.originalProtein == 2.5)
        #expect(item.originalFat == 0.3)
        #expect(item.originalCarbohydrates == 37.1)
    }

    @Test func NutritionInfoからの初期化でoriginal値が保持されるべき() {
        let info = NutritionInfo(
            name: "鶏むね肉",
            estimatedAmount: "100g",
            caloriesKcal: 165,
            proteinG: 31,
            fatG: 3.6,
            carbohydratesG: 0
        )
        let item = FoodEditItem(from: info)

        #expect(item.originalName == "鶏むね肉")
        #expect(item.originalQuantity == "100g")
        #expect(item.originalCalories == 165)
    }

    // MARK: - グラム変更再計算テスト

    @Test func グラム変更時に栄養素が正しく再計算されるべき() {
        let item = FoodEditItem(
            name: "白米",
            quantity: "100g",
            calories: 168,
            protein: 2.5,
            fat: 0.3,
            carbohydrates: 37.1
        )

        item.quantity = "150g"
        item.recalculateNutrition()

        // 1.5倍
        #expect(item.calories == 252) // 168 * 1.5 = 252
        #expect(item.protein == 3.8) // 2.5 * 1.5 = 3.75 -> 3.8
        #expect(item.fat == 0.5) // 0.3 * 1.5 = 0.45 -> 0.5
        #expect(item.carbohydrates == 55.7) // 37.1 * 1.5 = 55.65 -> 55.7
    }

    @Test func グラム半減時に栄養素が半分になるべき() {
        let item = FoodEditItem(
            name: "白米",
            quantity: "200g",
            calories: 336,
            protein: 5.0,
            fat: 0.6,
            carbohydrates: 74.2
        )

        item.quantity = "100g"
        item.recalculateNutrition()

        #expect(item.calories == 168) // 336 * 0.5
        #expect(item.protein == 2.5) // 5.0 * 0.5
        #expect(item.fat == 0.3) // 0.6 * 0.5
        #expect(item.carbohydrates == 37.1) // 74.2 * 0.5
    }

    // MARK: - 杯数変更再計算テスト

    @Test func 杯数変更時に栄養素が正しく再計算されるべき() {
        let item = FoodEditItem(
            name: "味噌汁",
            quantity: "1杯",
            calories: 40,
            protein: 3.0,
            fat: 1.0,
            carbohydrates: 4.0
        )

        item.quantity = "2杯"
        item.recalculateNutrition()

        #expect(item.calories == 80) // 40 * 2
        #expect(item.protein == 6.0) // 3.0 * 2
        #expect(item.fat == 2.0) // 1.0 * 2
        #expect(item.carbohydrates == 8.0) // 4.0 * 2
    }

    // MARK: - パース失敗時のスキップテスト

    @Test func パースできない量では栄養素が変更されないべき() {
        let item = FoodEditItem(
            name: "ラーメン",
            quantity: "大盛り",
            calories: 500,
            protein: 20,
            fat: 15,
            carbohydrates: 60
        )

        item.quantity = "特盛り"
        item.recalculateNutrition()

        // 元の値が維持される
        #expect(item.calories == 500)
        #expect(item.protein == 20)
        #expect(item.fat == 15)
        #expect(item.carbohydrates == 60)
    }

    @Test func 異なる単位への変更では栄養素が変更されないべき() {
        let item = FoodEditItem(
            name: "白米",
            quantity: "100g",
            calories: 168,
            protein: 2.5,
            fat: 0.3,
            carbohydrates: 37.1
        )

        item.quantity = "1杯"
        item.recalculateNutrition()

        // 異なる単位のため変更なし
        #expect(item.calories == 168)
        #expect(item.protein == 2.5)
    }

    // MARK: - 累積誤差防止テスト

    @Test func 複数回の変更でも累積誤差が発生しないべき() {
        let item = FoodEditItem(
            name: "白米",
            quantity: "100g",
            calories: 168,
            protein: 2.5,
            fat: 0.3,
            carbohydrates: 37.1
        )

        // 100g -> 150g
        item.quantity = "150g"
        item.recalculateNutrition()
        #expect(item.calories == 252)

        // 150g -> 200g（元の100gの2倍として計算される）
        item.quantity = "200g"
        item.recalculateNutrition()
        #expect(item.calories == 336) // 168 * 2.0 = 336

        // 200g -> 100g（元の値に戻る）
        item.quantity = "100g"
        item.recalculateNutrition()
        #expect(item.calories == 168) // 168 * 1.0 = 168（完全に元通り）
    }

    // MARK: - hasNameChanged テスト

    @Test func 名前が変更されていない場合はfalseを返すべき() {
        let item = FoodEditItem(name: "白米", quantity: "100g", calories: 168)
        #expect(item.hasNameChanged == false)
    }

    @Test func 名前が変更された場合はtrueを返すべき() {
        let item = FoodEditItem(name: "白米", quantity: "100g", calories: 168)
        item.name = "玄米"
        #expect(item.hasNameChanged == true)
    }

    @Test func 新規アイテムで元の名前が空の場合はfalseを返すべき() {
        let item = FoodEditItem()
        item.name = "白米"
        #expect(item.hasNameChanged == false)
    }
}
