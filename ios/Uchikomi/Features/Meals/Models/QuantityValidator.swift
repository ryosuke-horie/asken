import Foundation

// MARK: - QuantityValidator

enum QuantityValidator {
    /// 半角数字（整数）のバリデーション
    static func isValidInteger(_ text: String) -> Bool {
        guard !text.isEmpty else { return true }
        return text.allSatisfy { $0.isNumber && $0.isASCII }
    }

    /// 半角数字（小数許容）のバリデーション
    static func isValidDecimal(_ text: String) -> Bool {
        guard !text.isEmpty else { return true }
        let decimalPattern = /^(\d+\.?\d*|\.\d+)$/
        return text.wholeMatch(of: decimalPattern) != nil
    }
}
