import Foundation

// MARK: - IngredientCategory

enum IngredientCategory: String, Codable, CaseIterable, Identifiable {
    var id: String {
        rawValue
    }

    case meat
    case fish
    case vegetable
    case fruit
    case dairy
    case grain
    case seasoning
    case beverage
    case other

    var displayName: String {
        switch self {
        case .meat: "肉類"
        case .fish: "魚介類"
        case .vegetable: "野菜"
        case .fruit: "果物"
        case .dairy: "乳製品"
        case .grain: "穀物"
        case .seasoning: "調味料"
        case .beverage: "飲料"
        case .other: "その他"
        }
    }
}

// MARK: - IngredientSource

enum IngredientSource: String, Codable {
    case receipt
    case manual
}

// MARK: - IngredientUnit

enum IngredientUnit: String, Codable, CaseIterable, Identifiable {
    var id: String {
        rawValue
    }

    case gram = "g"
    case milliliter = "ml"
    case cup = "杯"
    case serving = "人前"
    case piece = "個"
    case sheet = "枚"
    case stick = "本"
    case slice = "切れ"
    case meal = "食"
    case plate = "皿"
    case bowl = "膳"
    case block = "丁"
    case bunch = "束"
    case bag = "袋"
    case can = "缶"
    case go = "合"
    case ball = "玉"
    case grain = "粒"
    case pack = "パック"
    case tablespoon = "大さじ"
    case teaspoon = "小さじ"
}

// MARK: - Ingredient

struct Ingredient: Identifiable, Codable, Equatable {
    let id: String
    let name: String
    let category: IngredientCategory
    let quantity: Double
    let unit: String
    let purchaseDate: Date?
    let expiryDate: Date?
    let source: IngredientSource
    let createdAt: Date
    let updatedAt: Date

    private enum CodingKeys: String, CodingKey {
        case id, name, category, quantity, unit, purchaseDate, expiryDate, source, createdAt, updatedAt
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = try container.decode(String.self, forKey: .id)
        name = try container.decode(String.self, forKey: .name)
        category = try container.decode(IngredientCategory.self, forKey: .category)
        quantity = try container.decode(Double.self, forKey: .quantity)
        unit = try container.decode(String.self, forKey: .unit)
        source = try container.decode(IngredientSource.self, forKey: .source)
        createdAt = try container.decodeISO8601Date(forKey: .createdAt)
        updatedAt = try container.decodeISO8601Date(forKey: .updatedAt)
        purchaseDate = try Self.decodeOptionalDateOnly(from: container, forKey: .purchaseDate)
        expiryDate = try Self.decodeOptionalDateOnly(from: container, forKey: .expiryDate)
    }

    private static func decodeOptionalDateOnly(
        from container: KeyedDecodingContainer<CodingKeys>,
        forKey key: CodingKeys
    ) throws -> Date? {
        guard let str = try container.decodeIfPresent(String.self, forKey: key), !str.isEmpty else {
            return nil
        }
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd"
        formatter.locale = Locale(identifier: "en_US_POSIX")
        guard let date = formatter.date(from: str) else {
            throw DecodingError.dataCorruptedError(
                forKey: key,
                in: container,
                debugDescription: "Invalid date format: \(str)"
            )
        }
        return date
    }

    /// 消費期限が3日以内かどうか（期限切れは含まない）
    var isExpiringWithinThreeDays: Bool {
        guard let expiryDate else { return false }
        let daysUntilExpiry = Calendar.current.dateComponents([.day], from: Date(), to: expiryDate).day ?? 0
        return daysUntilExpiry >= 0 && daysUntilExpiry <= 3
    }

    /// 消費期限が切れているかどうか
    var isExpired: Bool {
        guard let expiryDate else { return false }
        return expiryDate < Date()
    }
}

// MARK: - ScannedIngredient

/// レシートスキャン結果（未保存状態）
struct ScannedIngredient: Identifiable, Equatable {
    let id: UUID
    var name: String
    var category: IngredientCategory
    var quantity: Double
    var unit: String

    init(
        id: UUID = UUID(),
        name: String,
        category: IngredientCategory,
        quantity: Double,
        unit: String
    ) {
        self.id = id
        self.name = name
        self.category = category
        self.quantity = quantity
        self.unit = unit
    }
}
