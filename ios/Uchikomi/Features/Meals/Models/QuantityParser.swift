import Foundation

// MARK: - ParsedQuantity

struct ParsedQuantity: Equatable {
    let value: Double
    let unit: String
}

// MARK: - QuantityParser

enum QuantityParser {
    // グラム表記のパターン: "100g", "100G", "100グラム", "100 g"
    private static let gramPattern = #/(\d+(?:\.\d+)?)\s*(?:g|G|グラム)/#

    // ミリリットル表記のパターン: "200ml", "200ML", "200mL", "200ミリリットル", "200 ml"
    private static let mlPattern = #/(\d+(?:\.\d+)?)\s*(?:ml|ML|mL|ミリリットル)/#

    /// 日本語単位（MeasurementUnitからg/mlを除いて動的生成）
    private static let japaneseUnits: [String] = MeasurementUnit.allCases
        .filter { $0 != .gram && $0 != .milliliter }
        .map(\.rawValue)

    private static let japanesePattern: NSRegularExpression = {
        let units = japaneseUnits.map { NSRegularExpression.escapedPattern(for: $0) }.joined(separator: "|")
        let pattern = #"^(\d+(?:\.\d+)?)\s*("# + units + ")$"
        do {
            return try NSRegularExpression(pattern: pattern)
        } catch {
            preconditionFailure(
                "japanesePattern のコンパイルに失敗しました。MeasurementUnit に無効な rawValue が追加された可能性があります。エラー: \(error)"
            )
        }
    }()

    static func parse(_ text: String) -> ParsedQuantity? {
        let trimmed = text.trimmingCharacters(in: .whitespaces)
        guard !trimmed.isEmpty else { return nil }

        // グラム表記を試行
        if let match = trimmed.wholeMatch(of: gramPattern) {
            guard let value = Double(match.1) else { return nil }
            return ParsedQuantity(value: value, unit: "g")
        }

        // ミリリットル表記を試行
        if let match = trimmed.wholeMatch(of: mlPattern) {
            guard let value = Double(match.1) else { return nil }
            return ParsedQuantity(value: value, unit: "ml")
        }

        // 日本語単位を試行（キャッシュ済み正規表現を使用）
        let range = NSRange(trimmed.startIndex..., in: trimmed)
        if let match = japanesePattern.firstMatch(in: trimmed, range: range) {
            guard let valueRange = Range(match.range(at: 1), in: trimmed),
                  let unitRange = Range(match.range(at: 2), in: trimmed),
                  let value = Double(trimmed[valueRange]) else {
                return nil
            }
            let unit = String(trimmed[unitRange])
            return ParsedQuantity(value: value, unit: unit)
        }

        return nil
    }

    static func calculateRatio(from: ParsedQuantity, to: ParsedQuantity) -> Double? {
        guard from.unit == to.unit else { return nil }
        guard from.value > 0 else { return nil }
        return to.value / from.value
    }

    /// 既存の文字列からMeasurementUnitを取得
    static func parseUnit(_ text: String) -> MeasurementUnit? {
        guard let parsed = parse(text) else { return nil }
        return MeasurementUnit(rawValue: parsed.unit)
    }

    /// 数値部分のみを文字列として抽出
    static func parseValue(_ text: String) -> String? {
        guard let parsed = parse(text) else { return nil }
        return parsed.value == floor(parsed.value)
            ? String(Int(parsed.value))
            : String(parsed.value)
    }
}
