import Foundation

// MARK: - MeasurementUnit

enum MeasurementUnit: String, CaseIterable, Identifiable, Codable {
    case gram = "g"
    case cup = "杯"
    case serving = "人前"
    case piece = "個"
    case sheet = "枚"
    case stick = "本"
    case slice = "切れ"
    case meal = "食"
    case plate = "皿"
    case set = "膳"
    case block = "丁"
    case bundle = "束"
    case bag = "袋"
    case can = "缶"
    case go = "合"
    case ball = "玉"

    var id: String {
        rawValue
    }

    var displayName: String {
        rawValue
    }

    var inputType: QuantityInputType {
        .decimal
    }
}

// MARK: - QuantityInputType

enum QuantityInputType {
    case decimal
}
