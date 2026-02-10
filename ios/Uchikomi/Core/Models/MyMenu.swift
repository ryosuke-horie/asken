import Foundation

// MARK: - MyMenuItem

struct MyMenuItem: Identifiable, Codable, Equatable {
    let id: String
    let name: String
    let foods: [NutritionInfo]
    let totalCalories: Double
    let totalProtein: Double
    let totalFat: Double
    let totalCarbohydrates: Double
    let createdAt: Date
    let updatedAt: Date

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case foods
        case totalCalories = "totalCalories"
        case totalProtein = "totalProtein"
        case totalFat = "totalFat"
        case totalCarbohydrates = "totalCarbohydrates"
        case createdAt = "createdAt"
        case updatedAt = "updatedAt"
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = try container.decode(String.self, forKey: .id)
        name = try container.decode(String.self, forKey: .name)
        foods = try container.decode([NutritionInfo].self, forKey: .foods)
        totalCalories = try container.decode(Double.self, forKey: .totalCalories)
        totalProtein = try container.decode(Double.self, forKey: .totalProtein)
        totalFat = try container.decode(Double.self, forKey: .totalFat)
        totalCarbohydrates = try container.decode(Double.self, forKey: .totalCarbohydrates)

        let dateFormatter = ISO8601DateFormatter()
        let createdAtString = try container.decode(String.self, forKey: .createdAt)
        let updatedAtString = try container.decode(String.self, forKey: .updatedAt)

        createdAt = dateFormatter.date(from: createdAtString) ?? Date()
        updatedAt = dateFormatter.date(from: updatedAtString) ?? Date()
    }

    init(
        id: String,
        name: String,
        foods: [NutritionInfo],
        totalCalories: Double,
        totalProtein: Double,
        totalFat: Double,
        totalCarbohydrates: Double,
        createdAt: Date,
        updatedAt: Date
    ) {
        self.id = id
        self.name = name
        self.foods = foods
        self.totalCalories = totalCalories
        self.totalProtein = totalProtein
        self.totalFat = totalFat
        self.totalCarbohydrates = totalCarbohydrates
        self.createdAt = createdAt
        self.updatedAt = updatedAt
    }
}
