import Testing
@testable import Uchikomi

@Suite struct FoodEditItemTests {
    // MARK: - 元の値の保持

    @Test func 初期化時に元の値が保持されるべき() {
        let item = FoodEditItem(
            name: "白米",
            quantity: "100g",
            calories: 168,
            protein: 2.5,
            fat: 0.3,
            carbohydrates: 37.1
        )

        #expect(item.originalQuantity == "100g")
        #expect(item.originalCalories == 168)
        #expect(item.originalProtein == 2.5)
        #expect(item.originalFat == 0.3)
        #expect(item.originalCarbohydrates == 37.1)
    }

    // MARK: - 栄養素の再計算（グラム表記）

    @Test func グラム数が増加した時に栄養素が比率で再計算されるべき() {
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

        // 1.5倍になることを確認
        #expect(item.calories == 252)
        #expect(item.protein == 3.8) // 2.5 * 1.5 = 3.75 → 3.8
        #expect(item.fat == 0.5) // 0.3 * 1.5 = 0.45 → 0.5
        #expect(item.carbohydrates == 55.7) // 37.1 * 1.5 = 55.65 → 55.7
    }

    @Test func グラム数が減少した時に栄養素が比率で再計算されるべき() {
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

        // 0.5倍になることを確認
        #expect(item.calories == 168)
        #expect(item.protein == 2.5)
        #expect(item.fat == 0.3)
        #expect(item.carbohydrates == 37.1)
    }

    // MARK: - 栄養素の再計算（数量表記）

    @Test func 杯数が増加した時に栄養素が比率で再計算されるべき() {
        let item = FoodEditItem(
            name: "味噌汁",
            quantity: "1杯",
            calories: 40,
            protein: 3.0,
            fat: 1.5,
            carbohydrates: 3.5
        )

        item.quantity = "2杯"
        item.recalculateNutrition()

        // 2倍になることを確認
        #expect(item.calories == 80)
        #expect(item.protein == 6.0)
        #expect(item.fat == 3.0)
        #expect(item.carbohydrates == 7.0)
    }

    @Test func 小数の杯数でも再計算されるべき() {
        let item = FoodEditItem(
            name: "味噌汁",
            quantity: "1杯",
            calories: 40,
            protein: 3.0,
            fat: 1.5,
            carbohydrates: 3.5
        )

        item.quantity = "1.5杯"
        item.recalculateNutrition()

        // 1.5倍になることを確認
        #expect(item.calories == 60)
        #expect(item.protein == 4.5)
        #expect(item.fat == 2.3) // 1.5 * 1.5 = 2.25 → 2.3
        #expect(item.carbohydrates == 5.3) // 3.5 * 1.5 = 5.25 → 5.3
    }

    // MARK: - 再計算がスキップされるケース

    @Test func 単位が異なる場合は再計算されないべき() {
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

        // 元の値が維持されることを確認
        #expect(item.calories == 168)
        #expect(item.protein == 2.5)
        #expect(item.fat == 0.3)
        #expect(item.carbohydrates == 37.1)
    }

    @Test func パースできない量の場合は再計算されないべき() {
        let item = FoodEditItem(
            name: "ラーメン",
            quantity: "1杯",
            calories: 500,
            protein: 20,
            fat: 15,
            carbohydrates: 60
        )

        item.quantity = "大盛り"
        item.recalculateNutrition()

        // 元の値が維持されることを確認
        #expect(item.calories == 500)
        #expect(item.protein == 20)
        #expect(item.fat == 15)
        #expect(item.carbohydrates == 60)
    }

    @Test func 元の量がパースできない場合は再計算されないべき() {
        let item = FoodEditItem(
            name: "ラーメン",
            quantity: "普通盛り",
            calories: 500,
            protein: 20,
            fat: 15,
            carbohydrates: 60
        )

        item.quantity = "150g"
        item.recalculateNutrition()

        // 元の値が維持されることを確認
        #expect(item.calories == 500)
        #expect(item.protein == 20)
        #expect(item.fat == 15)
        #expect(item.carbohydrates == 60)
    }

    // MARK: - 元の値をベースにした再計算

    @Test func 複数回変更しても常に元の値から計算されるべき() {
        let item = FoodEditItem(
            name: "白米",
            quantity: "100g",
            calories: 168,
            protein: 2.5,
            fat: 0.3,
            carbohydrates: 37.1
        )

        // 1回目: 150gに変更
        item.quantity = "150g"
        item.recalculateNutrition()
        #expect(item.calories == 252)

        // 2回目: 200gに変更（150gからではなく、元の100gから計算）
        item.quantity = "200g"
        item.recalculateNutrition()
        #expect(item.calories == 336) // 168 * 2 = 336
    }
}
