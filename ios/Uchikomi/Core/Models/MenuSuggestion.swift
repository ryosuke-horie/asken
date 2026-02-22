import Foundation

// MARK: - MenuSuggestionStatus

enum MenuSuggestionStatus: String, Codable {
    case suggested
    case accepted
    case dismissed
}

// MARK: - MenuSuggestionIngredient

struct MenuSuggestionIngredient: Decodable, Equatable, Identifiable {
    var id: String {
        ingredientId
    }

    let ingredientId: String
    let name: String
    let quantity: Double
    let unit: String
}

// MARK: - EstimatedNutrition

struct EstimatedNutrition: Decodable, Equatable {
    let calories: Double
    let protein: Double
    let fat: Double
    let carbohydrates: Double
}

// MARK: - MenuSuggestion

struct MenuSuggestion: Identifiable, Decodable, Equatable, Hashable {
    let id: String
    let title: String
    let description: String
    let reason: String
    let ingredientsUsed: [MenuSuggestionIngredient]
    let recipe: String?
    let estimatedNutrition: EstimatedNutrition
    let mealType: MealType
    let status: MenuSuggestionStatus
    let createdAt: Date

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = try container.decode(String.self, forKey: .id)
        title = try container.decode(String.self, forKey: .title)
        description = try container.decode(String.self, forKey: .description)
        reason = try container.decode(String.self, forKey: .reason)
        ingredientsUsed = try container.decode([MenuSuggestionIngredient].self, forKey: .ingredientsUsed)
        recipe = try container.decodeIfPresent(String.self, forKey: .recipe)
        estimatedNutrition = try container.decode(EstimatedNutrition.self, forKey: .estimatedNutrition)
        mealType = try container.decode(MealType.self, forKey: .mealType)
        status = try container.decode(MenuSuggestionStatus.self, forKey: .status)
        createdAt = try container.decodeISO8601Date(forKey: .createdAt)
    }

    static func == (lhs: MenuSuggestion, rhs: MenuSuggestion) -> Bool {
        lhs.id == rhs.id
    }

    func hash(into hasher: inout Hasher) {
        hasher.combine(id)
    }
}

// MARK: - SuggestMenuRequest

struct SuggestMenuRequest: Encodable {
    let mealType: MealType
    let count: Int
}

// MARK: - MenuSuggestionListResponse

struct MenuSuggestionListResponse: Decodable {
    let suggestions: [MenuSuggestion]
}

// MARK: - AcceptMenuSuggestionResult

struct AcceptMenuSuggestionResult: Decodable {
    let analysisRequestId: String
    let deductedIngredients: [DeductedIngredient]
}

// MARK: - DeductedIngredient

struct DeductedIngredient: Decodable, Identifiable {
    var id: String {
        ingredientId
    }

    let ingredientId: String
    let name: String
    let deducted: Double
    let remaining: Double
}
