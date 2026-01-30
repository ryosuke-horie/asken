import Foundation

protocol WeightRepositoryProtocol {
    func getRecords(period: WeightPeriod) async throws -> WeightRecordsResponse
    func createRecord(weight: Double, recordedAt: Date) async throws -> WeightRecord
    func getGoal() async throws -> WeightGoal?
    func setGoal(targetWeight: Double, targetDate: Date) async throws
}

final class WeightRepository: WeightRepositoryProtocol {
    private let apiClient = APIClient.shared
    private let dateFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd"
        return formatter
    }()

    func getRecords(period: WeightPeriod) async throws -> WeightRecordsResponse {
        return try await apiClient.request(endpoint: .weightRecords(period: period.rawValue))
    }

    func createRecord(weight: Double, recordedAt: Date) async throws -> WeightRecord {
        let request = CreateWeightRecordRequest(
            weight: weight,
            recordedAt: dateFormatter.string(from: recordedAt)
        )
        return try await apiClient.request(endpoint: .createWeightRecord, body: request)
    }

    func getGoal() async throws -> WeightGoal? {
        do {
            return try await apiClient.request(endpoint: .weightGoal)
        } catch APIError.notFound {
            return nil
        }
    }

    func setGoal(targetWeight: Double, targetDate: Date) async throws {
        let request = SetWeightGoalRequest(
            targetWeight: targetWeight,
            targetDate: dateFormatter.string(from: targetDate)
        )
        let _: EmptyResponse = try await apiClient.request(endpoint: .setWeightGoal, body: request)
    }
}
