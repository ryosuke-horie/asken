import Foundation

enum QuantityParser {
    struct ParsedQuantity: Equatable {
        let value: Double
        let unit: String
    }

    static func parse(_ text: String) -> ParsedQuantity? {
        let trimmed = text.trimmingCharacters(in: .whitespaces)
        guard !trimmed.isEmpty else { return nil }

        // グラム表記: "100g", "100G", "100グラム", "100 g"
        if let result = parseGrams(trimmed) {
            return result
        }

        // 日本語単位: "1杯", "2人前", "3個", "2枚", "1本", "3切れ"
        if let result = parseJapaneseUnit(trimmed) {
            return result
        }

        return nil
    }

    static func calculateRatio(from original: ParsedQuantity, to updated: ParsedQuantity) -> Double? {
        // 単位が異なる場合は計算不可
        guard original.unit == updated.unit else { return nil }

        // 元の値がゼロの場合は計算不可
        guard original.value > 0 else { return nil }

        return updated.value / original.value
    }

    private static func parseGrams(_ text: String) -> ParsedQuantity? {
        // パターン: "100g", "100G", "100 g", "100グラム"
        let patterns = [
            #"^(\d+\.?\d*)\s*[gG]$"#,
            #"^(\d+\.?\d*)\s*グラム$"#
        ]

        for pattern in patterns {
            if let regex = try? NSRegularExpression(pattern: pattern),
               let match = regex.firstMatch(
                   in: text,
                   range: NSRange(text.startIndex..., in: text)
               ) {
                if let valueRange = Range(match.range(at: 1), in: text),
                   let value = Double(text[valueRange]) {
                    return ParsedQuantity(value: value, unit: "g")
                }
            }
        }
        return nil
    }

    private static func parseJapaneseUnit(_ text: String) -> ParsedQuantity? {
        // 対応単位: 杯, 人前, 個, 枚, 本, 切れ
        let units = ["杯", "人前", "個", "枚", "本", "切れ"]
        let pattern = #"^(\d+\.?\d*)("# + units.joined(separator: "|") + #")$"#

        if let regex = try? NSRegularExpression(pattern: pattern),
           let match = regex.firstMatch(
               in: text,
               range: NSRange(text.startIndex..., in: text)
           ) {
            if let valueRange = Range(match.range(at: 1), in: text),
               let unitRange = Range(match.range(at: 2), in: text),
               let value = Double(text[valueRange]) {
                let unit = String(text[unitRange])
                return ParsedQuantity(value: value, unit: unit)
            }
        }
        return nil
    }
}
