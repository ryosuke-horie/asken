import Foundation
import os

private let logger = Logger(subsystem: Bundle.main.bundleIdentifier ?? "Uchikomi", category: "WeightRecord")

// MARK: - WeightTiming

enum WeightTiming: String, CaseIterable, Identifiable {
    case morning
    case beforePractice
    case afterPractice
    case beforeSleep

    var id: String {
        rawValue
    }

    var displayName: String {
        switch self {
        case .morning: "起床時"
        case .beforePractice: "練習前"
        case .afterPractice: "練習後"
        case .beforeSleep: "就寝前"
        }
    }

    static func from(note: String?) -> WeightTiming? {
        guard let note else { return nil }
        let result = allCases.first { $0.displayName == note }
        if result == nil, !note.isEmpty {
            logger.debug("WeightTiming変換失敗: 未知のnote値 '\(note)'")
        }
        return result
    }
}

// MARK: - WeightRecord

struct WeightRecord: Codable, Identifiable {
    let id: String
    let weightKg: Double
    let recordedAt: String
    let note: String?
    let createdAt: String
    let updatedAt: String

    static let minWeightKg = 20.0
    static let maxWeightKg = 300.0

    private static let iso8601Formatter: ISO8601DateFormatter = {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return formatter
    }()

    private static let iso8601FallbackFormatter: ISO8601DateFormatter = {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime]
        return formatter
    }()

    static func parseISO8601(_ string: String) -> Date? {
        let result = iso8601Formatter.date(from: string) ?? iso8601FallbackFormatter.date(from: string)
        if result == nil {
            logger.warning("ISO8601日付パースに失敗: \(string)")
        }
        return result
    }
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
