import Foundation

// MARK: - MealType

enum MealType: String, Codable, CaseIterable, Identifiable {
    var id: String {
        rawValue
    }

    case breakfast
    case lunch
    case dinner
    case snack

    var displayName: String {
        switch self {
        case .breakfast: "朝食"
        case .lunch: "昼食"
        case .dinner: "夕食"
        case .snack: "間食"
        }
    }

    var icon: String {
        switch self {
        case .breakfast: "sunrise"
        case .lunch: "sun.max"
        case .dinner: "moon"
        case .snack: "cup.and.saucer"
        }
    }
}

// MARK: - InputType

enum InputType: String, Codable {
    case image
    case text
    case mylist
    case skipped
}

// MARK: - NutritionInfo

struct NutritionInfo: Codable, Identifiable, Equatable {
    var id: String {
        name
    }

    let name: String
    let estimatedAmount: String
    let caloriesKcal: Double
    let proteinG: Double
    let fatG: Double
    let carbohydratesG: Double
}

// MARK: - HistoryDetail

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

// MARK: - MealsByType

struct MealsByType: Codable {
    let breakfast: [HistoryDetail]
    let lunch: [HistoryDetail]
    let dinner: [HistoryDetail]
    let snack: [HistoryDetail]

    func meals(for type: MealType) -> [HistoryDetail] {
        switch type {
        case .breakfast: breakfast
        case .lunch: lunch
        case .dinner: dinner
        case .snack: snack
        }
    }
}

// MARK: - DailyTotal

struct DailyTotal: Codable {
    let totalCalories: Double
    let totalProtein: Double
    let totalFat: Double
    let totalCarbohydrates: Double
}

// MARK: - DailyMeals

struct DailyMeals: Codable {
    let date: String
    let meals: MealsByType
    let dailyTotal: DailyTotal
}

// MARK: - AnalysisStatusResponse

struct AnalysisStatusResponse: Decodable {
    let status: String
    let error: String?
    let message: String?
}

// MARK: - AnalysisResultResponse

struct AnalysisResultResponse: Decodable {
    let status: String
    let result: AnalysisResult
}

// MARK: - AnalysisResult

struct AnalysisResult: Decodable {
    let foods: [NutritionInfo]
    let totalCalories: Double
    let totalProtein: Double
    let totalFat: Double
    let totalCarbohydrates: Double
}
