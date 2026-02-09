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

    /// 全角数字を半角に変換
    static func normalizeFullWidth(_ text: String) -> String {
        let fullWidthToHalfWidthMap: [Character: Character] = [
            "０": "0", "１": "1", "２": "2", "３": "3", "４": "4",
            "５": "5", "６": "6", "７": "7", "８": "8", "９": "9",
            "．": "."
        ]
        return String(text.map { fullWidthToHalfWidthMap[$0] ?? $0 })
    }
}
