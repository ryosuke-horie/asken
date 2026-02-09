import Testing
@testable import Uchikomi

@Suite struct QuantityValidatorTests {
    // MARK: - 整数バリデーションテスト

    @Test func 整数バリデーション_正しい整数を許容するべき() {
        #expect(QuantityValidator.isValidInteger("123") == true)
        #expect(QuantityValidator.isValidInteger("0") == true)
        #expect(QuantityValidator.isValidInteger("1") == true)
    }

    @Test func 整数バリデーション_空文字を許容するべき() {
        #expect(QuantityValidator.isValidInteger("") == true)
    }

    @Test func 整数バリデーション_小数を拒否するべき() {
        #expect(QuantityValidator.isValidInteger("12.3") == false)
        #expect(QuantityValidator.isValidInteger(".5") == false)
    }

    @Test func 整数バリデーション_全角数字を拒否するべき() {
        #expect(QuantityValidator.isValidInteger("１２３") == false)
        #expect(QuantityValidator.isValidInteger("１") == false)
    }

    @Test func 整数バリデーション_文字列を拒否するべき() {
        #expect(QuantityValidator.isValidInteger("abc") == false)
        #expect(QuantityValidator.isValidInteger("12a") == false)
    }

    // MARK: - 小数バリデーションテスト

    @Test func 小数バリデーション_正しい小数を許容するべき() {
        #expect(QuantityValidator.isValidDecimal("123.45") == true)
        #expect(QuantityValidator.isValidDecimal(".5") == true)
        #expect(QuantityValidator.isValidDecimal("0.5") == true)
        #expect(QuantityValidator.isValidDecimal("123") == true)
    }

    @Test func 小数バリデーション_空文字を許容するべき() {
        #expect(QuantityValidator.isValidDecimal("") == true)
    }

    @Test func 小数バリデーション_複数の小数点を拒否するべき() {
        #expect(QuantityValidator.isValidDecimal("12.3.4") == false)
    }

    @Test func 小数バリデーション_全角数字を拒否するべき() {
        #expect(QuantityValidator.isValidDecimal("１２３．４５") == false)
    }

    // MARK: - 全角半角変換テスト

    @Test func 全角数字が半角に変換されるべき() {
        #expect(QuantityValidator.normalizeFullWidth("１２３") == "123")
        #expect(QuantityValidator.normalizeFullWidth("０") == "0")
    }

    @Test func 全角小数点が半角に変換されるべき() {
        #expect(QuantityValidator.normalizeFullWidth("１．５") == "1.5")
        #expect(QuantityValidator.normalizeFullWidth("．５") == ".5")
    }

    @Test func 混在する全角半角が正しく変換されるべき() {
        #expect(QuantityValidator.normalizeFullWidth("１２3．45") == "123.45")
    }

    @Test func 変換対象外の文字はそのまま保持されるべき() {
        #expect(QuantityValidator.normalizeFullWidth("abc") == "abc")
        #expect(QuantityValidator.normalizeFullWidth("123abc") == "123abc")
    }
}
