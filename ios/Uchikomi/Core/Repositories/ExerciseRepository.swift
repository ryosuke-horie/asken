import Foundation

// MARK: - ExerciseRepositoryProtocol

/// @mockable
protocol ExerciseRepositoryProtocol {
    func createRecord(exerciseName: String, durationMinutes: Int, recordedDate: String) async throws -> ExerciseRecord
    func getDailyExercise(date: String) async throws -> DailyExerciseResponse
    func deleteRecord(id: String) async throws
}

// MARK: - ExerciseRepository

final class ExerciseRepository: ExerciseRepositoryProtocol {
    private let apiClient = APIClient.shared

    func createRecord(exerciseName: String, durationMinutes: Int, recordedDate: String) async throws -> ExerciseRecord {
        let request = CreateExerciseRecordRequest(
            exerciseName: exerciseName,
            durationMinutes: durationMinutes,
            recordedDate: recordedDate
        )
        return try await apiClient.request(endpoint: .createExerciseRecord, body: request)
    }

    func getDailyExercise(date: String) async throws -> DailyExerciseResponse {
        try await apiClient.request(endpoint: .dailyExercise(date: date))
    }

    func deleteRecord(id: String) async throws {
        try await apiClient.requestWithoutResponse(endpoint: .deleteExerciseRecord(id: id))
    }
}
