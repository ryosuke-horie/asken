import Foundation

// MARK: - WidgetWeightRecord

struct WidgetWeightRecord: Decodable {
    let id: String
    let weightKg: Double
    let recordedAt: String
    let note: String?
    let createdAt: String
    let updatedAt: String
}

// MARK: - WidgetWeightGoal

struct WidgetWeightGoal: Decodable {
    let targetWeightKg: Double
}

// MARK: - WidgetWeightRecordsResponse

struct WidgetWeightRecordsResponse: Decodable {
    let records: [WidgetWeightRecord]
    let goal: WidgetWeightGoal?
}

// MARK: - WidgetAnalysisStatus

struct WidgetAnalysisStatus: Decodable {
    let status: String
    let error: String?
    let message: String?
}

// MARK: - WidgetAnalyzeResponse

struct WidgetAnalyzeResponse: Decodable {
    let id: String
}

// MARK: - WidgetDailyTotal

struct WidgetDailyTotal: Decodable {
    let totalCalories: Double
    let totalProtein: Double
    let totalFat: Double
    let totalCarbohydrates: Double
}

// MARK: - WidgetDailyMeals

struct WidgetDailyMeals: Decodable {
    let date: String
    let dailyTotal: WidgetDailyTotal
}
