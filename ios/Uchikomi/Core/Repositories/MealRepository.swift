import Foundation

protocol MealRepositoryProtocol {
    func getDailyMeals(date: Date) async throws -> DailyMeals
    func uploadImage(data: Data, filename: String) async throws -> String
    func checkAnalysisStatus(id: String) async throws -> AnalysisStatusResponse
    func getAnalysisResult(id: String) async throws -> AnalysisResultResponse
    func saveMeal(analysisId: String, mealType: MealType, mealDate: Date) async throws
}

final class MealRepository: MealRepositoryProtocol {
    private let apiClient = APIClient.shared
    private let dateFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd"
        return formatter
    }()

    func getDailyMeals(date: Date) async throws -> DailyMeals {
        let dateString = dateFormatter.string(from: date)
        return try await apiClient.request(endpoint: .dailyMeals(date: dateString))
    }

    func uploadImage(data: Data, filename: String) async throws -> String {
        let response: AnalyzeResponse = try await apiClient.uploadImage(
            endpoint: .analyze,
            imageData: data,
            filename: filename
        )
        return response.id
    }

    func checkAnalysisStatus(id: String) async throws -> AnalysisStatusResponse {
        return try await apiClient.request(endpoint: .analysisStatus(id: id))
    }

    func getAnalysisResult(id: String) async throws -> AnalysisResultResponse {
        return try await apiClient.request(endpoint: .analysisResult(id: id))
    }

    func saveMeal(analysisId: String, mealType: MealType, mealDate: Date) async throws {
        let request = SaveMealRequest(
            analysisId: analysisId,
            mealType: mealType,
            mealDate: dateFormatter.string(from: mealDate)
        )
        let _: EmptyResponse = try await apiClient.request(endpoint: .saveMeal, body: request)
    }
}

struct EmptyResponse: Decodable {}
