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

// MARK: - MealType Reminder

extension MealType {
    static let reminderTargets: [MealType] = [.breakfast, .lunch, .dinner]
}

// MARK: - InputType

enum InputType: String, Codable {
    case image
    case text
    case mylist
    case skipped
}

// MARK: - AnalysisStatus

enum AnalysisStatus: String, Codable {
    case pending
    case processing
    case completed
    case failed
    case unknown

    init(from decoder: Decoder) throws {
        let rawValue = try decoder.singleValueContainer().decode(String.self)
        self = AnalysisStatus(rawValue: rawValue) ?? .unknown
    }
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
    let micronutrients: [String: Double]?

    init(
        name: String,
        estimatedAmount: String,
        caloriesKcal: Double,
        proteinG: Double,
        fatG: Double,
        carbohydratesG: Double,
        micronutrients: [String: Double]? = nil
    ) {
        self.name = name
        self.estimatedAmount = estimatedAmount
        self.caloriesKcal = caloriesKcal
        self.proteinG = proteinG
        self.fatG = fatG
        self.carbohydratesG = carbohydratesG
        self.micronutrients = micronutrients
    }
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
    let totalMicronutrients: [String: Double]?
    let foods: [NutritionInfo]

    private enum CodingKeys: String, CodingKey {
        case id, inputType, imagePath, inputText, createdAt, mealType, mealDate
        case totalCalories, totalProtein, totalFat, totalCarbohydrates, totalMicronutrients, foods
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = try container.decode(String.self, forKey: .id)
        inputType = try container.decode(InputType.self, forKey: .inputType)
        imagePath = try container.decodeIfPresent(String.self, forKey: .imagePath)
        inputText = try container.decodeIfPresent(String.self, forKey: .inputText)
        createdAt = try container.decode(String.self, forKey: .createdAt)
        mealType = try container.decodeIfPresent(MealType.self, forKey: .mealType)
        mealDate = try container.decodeIfPresent(String.self, forKey: .mealDate)
        totalCalories = try container.decode(Double.self, forKey: .totalCalories)
        totalProtein = try container.decode(Double.self, forKey: .totalProtein)
        totalFat = try container.decode(Double.self, forKey: .totalFat)
        totalCarbohydrates = try container.decode(Double.self, forKey: .totalCarbohydrates)
        totalMicronutrients = try container.decodeIfPresent([String: Double].self, forKey: .totalMicronutrients)
        foods = try container.decodeIfPresent([NutritionInfo].self, forKey: .foods) ?? []
    }
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

    func isSkipped(for type: MealType) -> Bool {
        meals(for: type).contains { $0.inputType == .skipped }
    }
}

// MARK: - DailyTotal

struct DailyTotal: Codable {
    let totalCalories: Double
    let totalProtein: Double
    let totalFat: Double
    let totalCarbohydrates: Double
    let totalMicronutrients: [String: Double]?
}

// MARK: - DailyMeals

struct DailyMeals: Codable {
    let date: String
    let meals: MealsByType
    let dailyTotal: DailyTotal
    let totalBurnedCaloriesKcal: Double
    let pendingAnalyses: [PendingAnalysisEntry]

    private enum CodingKeys: String, CodingKey {
        case date, meals, dailyTotal, totalBurnedCaloriesKcal, pendingAnalyses
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        date = try container.decode(String.self, forKey: .date)
        meals = try container.decode(MealsByType.self, forKey: .meals)
        dailyTotal = try container.decode(DailyTotal.self, forKey: .dailyTotal)
        totalBurnedCaloriesKcal = try container.decodeIfPresent(Double.self, forKey: .totalBurnedCaloriesKcal) ?? 0.0
        pendingAnalyses = try container.decodeIfPresent([PendingAnalysisEntry].self, forKey: .pendingAnalyses) ?? []
    }

    func pendingEntries(for mealType: MealType) -> [PendingAnalysisEntry] {
        pendingAnalyses.filter { $0.mealType == mealType }
    }
}

// MARK: - PendingAnalysisEntry

struct PendingAnalysisEntry: Codable, Identifiable {
    let id: String
    let mealType: MealType?
    let status: AnalysisStatus
    let inputType: String
    let errorMessage: String?
    let createdAt: String

    var isAnalyzing: Bool {
        status == .pending || status == .processing
    }

    var isReadyToConfirm: Bool {
        status == .completed
    }

    var isFailed: Bool {
        status == .failed
    }
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
    let totalMicronutrients: [String: Double]?
}
