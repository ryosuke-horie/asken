import Testing
@testable import Uchikomi

@Suite struct QuantityParserTests {
    // MARK: - グラム表記のパース

    @Test func グラム表記をパースできるべき() {
        let result = QuantityParser.parse("100g")
        #expect(result?.value == 100)
        #expect(result?.unit == "g")
    }

    @Test func 大文字Gでもパースできるべき() {
        let result = QuantityParser.parse("150G")
        #expect(result?.value == 150)
        #expect(result?.unit == "g")
    }

    @Test func グラム表記でもパースできるべき() {
        let result = QuantityParser.parse("200グラム")
        #expect(result?.value == 200)
        #expect(result?.unit == "g")
    }

    @Test func スペース付きグラム表記をパースできるべき() {
        let result = QuantityParser.parse("120 g")
        #expect(result?.value == 120)
        #expect(result?.unit == "g")
    }

    @Test func 小数点付きグラム表記をパースできるべき() {
        let result = QuantityParser.parse("150.5g")
        #expect(result?.value == 150.5)
        #expect(result?.unit == "g")
    }

    // MARK: - 数量表記のパース（杯）

    @Test func 杯表記をパースできるべき() {
        let result = QuantityParser.parse("1杯")
        #expect(result?.value == 1)
        #expect(result?.unit == "杯")
    }

    @Test func 複数杯をパースできるべき() {
        let result = QuantityParser.parse("2杯")
        #expect(result?.value == 2)
        #expect(result?.unit == "杯")
    }

    @Test func 小数杯をパースできるべき() {
        let result = QuantityParser.parse("1.5杯")
        #expect(result?.value == 1.5)
        #expect(result?.unit == "杯")
    }

    // MARK: - 数量表記のパース（人前）

    @Test func 人前表記をパースできるべき() {
        let result = QuantityParser.parse("1人前")
        #expect(result?.value == 1)
        #expect(result?.unit == "人前")
    }

    @Test func 複数人前をパースできるべき() {
        let result = QuantityParser.parse("2人前")
        #expect(result?.value == 2)
        #expect(result?.unit == "人前")
    }

    // MARK: - 数量表記のパース（個・枚・本・切れ）

    @Test func 個表記をパースできるべき() {
        let result = QuantityParser.parse("3個")
        #expect(result?.value == 3)
        #expect(result?.unit == "個")
    }

    @Test func 枚表記をパースできるべき() {
        let result = QuantityParser.parse("2枚")
        #expect(result?.value == 2)
        #expect(result?.unit == "枚")
    }

    @Test func 本表記をパースできるべき() {
        let result = QuantityParser.parse("1本")
        #expect(result?.value == 1)
        #expect(result?.unit == "本")
    }

    @Test func 切れ表記をパースできるべき() {
        let result = QuantityParser.parse("3切れ")
        #expect(result?.value == 3)
        #expect(result?.unit == "切れ")
    }

    // MARK: - パース失敗ケース

    @Test func 曖昧な表現はパースできないべき() {
        let result = QuantityParser.parse("大盛り")
        #expect(result == nil)
    }

    @Test func 空文字はパースできないべき() {
        let result = QuantityParser.parse("")
        #expect(result == nil)
    }

    @Test func 数値のない文字列はパースできないべき() {
        let result = QuantityParser.parse("たくさん")
        #expect(result == nil)
    }

    // MARK: - 比率計算

    @Test func 同じ単位での比率を計算できるべき() {
        let original = QuantityParser.ParsedQuantity(value: 100, unit: "g")
        let updated = QuantityParser.ParsedQuantity(value: 150, unit: "g")
        let ratio = QuantityParser.calculateRatio(from: original, to: updated)
        #expect(ratio == 1.5)
    }

    @Test func 杯の比率を計算できるべき() {
        let original = QuantityParser.ParsedQuantity(value: 1, unit: "杯")
        let updated = QuantityParser.ParsedQuantity(value: 2, unit: "杯")
        let ratio = QuantityParser.calculateRatio(from: original, to: updated)
        #expect(ratio == 2.0)
    }

    @Test func 異なる単位では比率計算できないべき() {
        let original = QuantityParser.ParsedQuantity(value: 100, unit: "g")
        let updated = QuantityParser.ParsedQuantity(value: 2, unit: "杯")
        let ratio = QuantityParser.calculateRatio(from: original, to: updated)
        #expect(ratio == nil)
    }

    @Test func 元の値がゼロの場合は比率計算できないべき() {
        let original = QuantityParser.ParsedQuantity(value: 0, unit: "g")
        let updated = QuantityParser.ParsedQuantity(value: 100, unit: "g")
        let ratio = QuantityParser.calculateRatio(from: original, to: updated)
        #expect(ratio == nil)
    }
}
