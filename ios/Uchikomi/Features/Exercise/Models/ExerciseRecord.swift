import Foundation

// MARK: - ExerciseRecord

struct ExerciseRecord: Codable, Identifiable {
    let id: String
    let exerciseName: String
    let durationMinutes: Int
    let burnedCaloriesKcal: Double
    let estimationMethod: String
    let recordedDate: String
    let createdAt: String
}

// MARK: - DailyExerciseResponse

struct DailyExerciseResponse: Codable {
    let records: [ExerciseRecord]
    let totalBurnedCaloriesKcal: Double
}

// MARK: - CreateExerciseRecordRequest

struct CreateExerciseRecordRequest: Encodable {
    let exerciseName: String
    let durationMinutes: Int
    let recordedDate: String
}
