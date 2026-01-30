import Foundation

struct WeightRecord: Codable, Identifiable {
    let id: String
    let userId: String
    let weight: Double
    let recordedAt: String
    let createdAt: String

    var recordedDate: Date? {
        ISO8601DateFormatter().date(from: recordedAt)
    }
}

struct WeightStats: Codable {
    let min: Double
    let max: Double
    let average: Double
}

struct WeightRecordsResponse: Codable {
    let records: [WeightRecord]
    let latest: WeightRecord?
    let stats: WeightStats
}

struct WeightGoal: Codable {
    let targetWeight: Double
    let targetDate: String
    let daysRemaining: Int
    let weightToLose: Double

    var targetDateFormatted: Date? {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd"
        return formatter.date(from: targetDate)
    }
}

enum WeightPeriod: String, CaseIterable {
    case week
    case month
    case threeMonths = "3months"

    var displayName: String {
        switch self {
        case .week: return "1週間"
        case .month: return "1ヶ月"
        case .threeMonths: return "3ヶ月"
        }
    }
}

// MARK: - API Request

struct CreateWeightRecordRequest: Encodable {
    let weight: Double
    let recordedAt: String
}

struct SetWeightGoalRequest: Encodable {
    let targetWeight: Double
    let targetDate: String
}
