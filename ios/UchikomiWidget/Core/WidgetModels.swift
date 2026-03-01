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
    enum Status {
        case completed
        case failed(reason: String)
        case processing
        case unknown(String)
    }

    let status: Status

    private enum CodingKeys: String, CodingKey {
        case status
        case error
        case message
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        let raw = try container.decode(String.self, forKey: .status)
        let errorMsg = try container.decodeIfPresent(String.self, forKey: .error)
        let message = try container.decodeIfPresent(String.self, forKey: .message)

        switch raw {
        case "completed":
            status = .completed
        case "failed":
            status = .failed(reason: errorMsg ?? message ?? "分析に失敗しました")
        case "processing":
            status = .processing
        default:
            status = .unknown(raw)
        }
    }
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
