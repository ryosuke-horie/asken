import Foundation

// MARK: - WeightRecord

struct WeightRecord: Codable, Identifiable {
    let id: String
    let weightKg: Double
    let recordedAt: String
    let note: String?
    let createdAt: String
    let updatedAt: String
}

// MARK: - WeightGoal

struct WeightGoal: Codable {
    let targetWeightKg: Double
    let updatedAt: String
}

// MARK: - WeightGoalNullableResponse

struct WeightGoalNullableResponse: Codable {
    let goal: WeightGoal?
}

// MARK: - DailySummary

struct DailySummary: Codable {
    let latestWeight: Double
    let count: Int
}

// MARK: - WeightRecordsListResponse

struct WeightRecordsListResponse: Codable {
    let records: [WeightRecord]
    let dailySummary: [String: DailySummary]
    let goal: WeightGoal?
}

// MARK: - ChartPeriod

enum ChartPeriod: String, CaseIterable, Identifiable {
    case week
    case month
    case threeMonths

    var id: String {
        rawValue
    }

    var displayName: String {
        switch self {
        case .week: "1週間"
        case .month: "1ヶ月"
        case .threeMonths: "3ヶ月"
        }
    }

    var days: Int {
        switch self {
        case .week: 7
        case .month: 30
        case .threeMonths: 90
        }
    }
}

// MARK: - CreateWeightRecordRequest

struct CreateWeightRecordRequest: Encodable {
    let weightKg: Double
    let recordedAt: String
    let note: String

    enum CodingKeys: String, CodingKey {
        case weightKg = "weight_kg"
        case recordedAt = "recorded_at"
        case note
    }
}

// MARK: - UpdateWeightRecordRequest

struct UpdateWeightRecordRequest: Encodable {
    let weightKg: Double
    let note: String

    enum CodingKeys: String, CodingKey {
        case weightKg = "weight_kg"
        case note
    }
}

// MARK: - SetWeightGoalRequest

struct SetWeightGoalRequest: Encodable {
    let targetWeightKg: Double

    enum CodingKeys: String, CodingKey {
        case targetWeightKg = "target_weight_kg"
    }
}
