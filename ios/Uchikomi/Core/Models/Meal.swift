import Foundation

enum MealType: String, Codable, CaseIterable, Identifiable {
    var id: String { rawValue }
    case breakfast
    case lunch
    case dinner
    case snack

    var displayName: String {
        switch self {
        case .breakfast: return "朝食"
        case .lunch: return "昼食"
        case .dinner: return "夕食"
        case .snack: return "間食"
        }
    }

    var icon: String {
        switch self {
        case .breakfast: return "sun.rise"
        case .lunch: return "sun.max"
        case .dinner: return "moon"
        case .snack: return "cup.and.saucer"
        }
    }
}

enum InputType: String, Codable {
    case image
    case text
    case mylist
    case skipped
}

struct NutritionInfo: Codable, Identifiable, Equatable {
    var id: String { name }

    let name: String
    let estimatedAmount: String
    let caloriesKcal: Double
    let proteinG: Double
    let fatG: Double
    let carbohydratesG: Double
}

struct HistoryDetail: Codable, Identifiable {
    let id: String
    let inputType: InputType
    let imagePath: String?
    let inputText: String?
    let createdAt: String
    let mealType: MealType?
    let mealDate: String?
    let totalCalories: Double
    let totalProtein: Double
    let totalFat: Double
    let totalCarbohydrates: Double
    let foods: [NutritionInfo]
}

struct MealsByType: Codable {
    let breakfast: [HistoryDetail]
    let lunch: [HistoryDetail]
    let dinner: [HistoryDetail]
    let snack: [HistoryDetail]

    func meals(for type: MealType) -> [HistoryDetail] {
        switch type {
        case .breakfast: return breakfast
        case .lunch: return lunch
        case .dinner: return dinner
        case .snack: return snack
        }
    }
}

struct DailyTotal: Codable {
    let totalCalories: Double
    let totalProtein: Double
    let totalFat: Double
    let totalCarbohydrates: Double
}

struct DailyMeals: Codable {
    let date: String
    let meals: MealsByType
    let dailyTotal: DailyTotal
}

// MARK: - API Request/Response

struct AnalysisStatusResponse: Decodable {
    let status: String
    let error: String?
    let message: String?
}

struct AnalysisResultResponse: Decodable {
    let id: String
    let foods: [NutritionInfo]
    let totalCalories: Double
    let totalProtein: Double
    let totalFat: Double
    let totalCarbohydrates: Double
}

struct SaveMealRequest: Encodable {
    let analysisId: String
    let mealType: MealType
    let mealDate: String
}
