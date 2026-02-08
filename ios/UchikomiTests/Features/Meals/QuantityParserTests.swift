import Testing
@testable import Uchikomi

@Suite struct QuantityParserTests {
    // MARK: - グラム表記パーステスト

    @Test func グラム表記_数値とgをパースできるべき() {
        let result = QuantityParser.parse("100g")
        #expect(result == ParsedQuantity(value: 100, unit: "g"))
    }

    @Test func グラム表記_大文字Gをパースできるべき() {
        let result = QuantityParser.parse("100G")
        #expect(result == ParsedQuantity(value: 100, unit: "g"))
    }

    @Test func グラム表記_グラムをパースできるべき() {
        let result = QuantityParser.parse("100グラム")
        #expect(result == ParsedQuantity(value: 100, unit: "g"))
    }

    @Test func グラム表記_スペース付きをパースできるべき() {
        let result = QuantityParser.parse("100 g")
        #expect(result == ParsedQuantity(value: 100, unit: "g"))
    }

    @Test func グラム表記_小数をパースできるべき() {
        let result = QuantityParser.parse("150.5g")
        #expect(result == ParsedQuantity(value: 150.5, unit: "g"))
    }

    // MARK: - 日本語単位パーステスト

    @Test func 日本語単位_杯をパースできるべき() {
        let result = QuantityParser.parse("1杯")
        #expect(result == ParsedQuantity(value: 1, unit: "杯"))
    }

    @Test func 日本語単位_人前をパースできるべき() {
        let result = QuantityParser.parse("2人前")
        #expect(result == ParsedQuantity(value: 2, unit: "人前"))
    }

    @Test func 日本語単位_個をパースできるべき() {
        let result = QuantityParser.parse("3個")
        #expect(result == ParsedQuantity(value: 3, unit: "個"))
    }

    @Test func 日本語単位_枚をパースできるべき() {
        let result = QuantityParser.parse("2枚")
        #expect(result == ParsedQuantity(value: 2, unit: "枚"))
    }

    @Test func 日本語単位_本をパースできるべき() {
        let result = QuantityParser.parse("1本")
        #expect(result == ParsedQuantity(value: 1, unit: "本"))
    }

    @Test func 日本語単位_切れをパースできるべき() {
        let result = QuantityParser.parse("3切れ")
        #expect(result == ParsedQuantity(value: 3, unit: "切れ"))
    }

    // MARK: - パース失敗ケース

    @Test func パース失敗_空文字列() {
        let result = QuantityParser.parse("")
        #expect(result == nil)
    }

    @Test func パース失敗_テキストのみ() {
        let result = QuantityParser.parse("大盛り")
        #expect(result == nil)
    }

    @Test func パース失敗_単位なし数字() {
        let result = QuantityParser.parse("100")
        #expect(result == nil)
    }

    @Test func パース失敗_不明な単位() {
        let result = QuantityParser.parse("1リットル")
        #expect(result == nil)
    }

    // MARK: - 比率計算テスト

    @Test func 比率計算_同一グラム単位で正しい比率を返すべき() {
        let from = ParsedQuantity(value: 100, unit: "g")
        let to = ParsedQuantity(value: 150, unit: "g")
        let ratio = QuantityParser.calculateRatio(from: from, to: to)
        #expect(ratio == 1.5)
    }

    @Test func 比率計算_同一日本語単位で正しい比率を返すべき() {
        let from = ParsedQuantity(value: 1, unit: "杯")
        let to = ParsedQuantity(value: 2, unit: "杯")
        let ratio = QuantityParser.calculateRatio(from: from, to: to)
        #expect(ratio == 2.0)
    }

    @Test func 比率計算_異なる単位ではnilを返すべき() {
        let from = ParsedQuantity(value: 100, unit: "g")
        let to = ParsedQuantity(value: 2, unit: "杯")
        let ratio = QuantityParser.calculateRatio(from: from, to: to)
        #expect(ratio == nil)
    }

    @Test func 比率計算_ゼロからの変更ではnilを返すべき() {
        let from = ParsedQuantity(value: 0, unit: "g")
        let to = ParsedQuantity(value: 100, unit: "g")
        let ratio = QuantityParser.calculateRatio(from: from, to: to)
        #expect(ratio == nil)
    }

    @Test func 比率計算_半分への変更() {
        let from = ParsedQuantity(value: 200, unit: "g")
        let to = ParsedQuantity(value: 100, unit: "g")
        let ratio = QuantityParser.calculateRatio(from: from, to: to)
        #expect(ratio == 0.5)
    }
}
