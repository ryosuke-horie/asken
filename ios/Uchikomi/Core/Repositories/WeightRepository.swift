import Foundation

// MARK: - WeightRepositoryProtocol

/// @mockable
protocol WeightRepositoryProtocol {
    func getRecords(from: Date, to: Date) async throws -> WeightRecordsListResponse
    func createRecord(weightKg: Double, recordedAt: Date, note: String) async throws -> WeightRecord
    func updateRecord(id: String, weightKg: Double, note: String) async throws -> WeightRecord
    func deleteRecord(id: String) async throws
    func getGoal() async throws -> WeightGoal?
    func setGoal(targetWeightKg: Double) async throws -> WeightGoal
}

// MARK: - WeightRepository

final class WeightRepository: WeightRepositoryProtocol {
    private let apiClient = APIClient.shared
    private let dateFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd"
        formatter.timeZone = TimeZone.current
        return formatter
    }()

    private let iso8601Formatter: ISO8601DateFormatter = {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime]
        return formatter
    }()

    private var currentTimezoneIdentifier: String {
        TimeZone.current.identifier
    }

    func getRecords(from: Date, to: Date) async throws -> WeightRecordsListResponse {
        let fromString = dateFormatter.string(from: from)
        let toString = dateFormatter.string(from: to)
        return try await apiClient.request(
            endpoint: .weightRecords(from: fromString, to: toString, timezone: currentTimezoneIdentifier)
        )
    }

    func createRecord(weightKg: Double, recordedAt: Date, note: String) async throws -> WeightRecord {
        let request = CreateWeightRecordRequest(
            weightKg: weightKg,
            recordedAt: iso8601Formatter.string(from: recordedAt),
            note: note
        )
        return try await apiClient.request(endpoint: .createWeightRecord, body: request)
    }

    func updateRecord(id: String, weightKg: Double, note: String) async throws -> WeightRecord {
        let request = UpdateWeightRecordRequest(weightKg: weightKg, note: note)
        return try await apiClient.request(endpoint: .updateWeightRecord(id: id), body: request)
    }

    func deleteRecord(id: String) async throws {
        try await apiClient.requestWithoutResponse(endpoint: .deleteWeightRecord(id: id))
    }

    func getGoal() async throws -> WeightGoal? {
        let response: WeightGoalNullableResponse = try await apiClient.request(endpoint: .getWeightGoal)
        return response.goal
    }

    func setGoal(targetWeightKg: Double) async throws -> WeightGoal {
        let request = SetWeightGoalRequest(targetWeightKg: targetWeightKg)
        return try await apiClient.request(endpoint: .setWeightGoal, body: request)
    }
}
