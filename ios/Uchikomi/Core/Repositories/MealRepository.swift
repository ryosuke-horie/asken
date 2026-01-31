import Foundation

/// @mockable
protocol MealRepositoryProtocol {
    func getDailyMeals(date: Date) async throws -> DailyMeals
    func uploadImage(data: Data, filename: String, mealType: MealType, mealDate: Date) async throws -> String
    func checkAnalysisStatus(id: String) async throws -> AnalysisStatusResponse
    func getAnalysisResult(id: String) async throws -> AnalysisResultResponse
    func getHistoryDetail(id: String) async throws -> HistoryDetail
    func updateHistory(historyId: String, foods: [UpdateFoodItem]) async throws -> HistoryDetail
    func deleteHistory(historyId: String) async throws
}

struct UpdateFoodItem: Encodable {
    let name: String
    let estimatedAmount: String
    let caloriesKcal: Double
    let proteinG: Double
    let fatG: Double
    let carbohydratesG: Double
}

struct UpdateHistoryRequest: Encodable {
    let foods: [UpdateFoodItem]
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

    func uploadImage(data: Data, filename: String, mealType: MealType, mealDate: Date) async throws -> String {
        let response: AnalyzeResponse = try await apiClient.uploadImage(
            endpoint: .analyze,
            imageData: data,
            filename: filename,
            additionalFields: [
                "meal_type": mealType.rawValue,
                "meal_date": dateFormatter.string(from: mealDate)
            ]
        )
        return response.id
    }

    func checkAnalysisStatus(id: String) async throws -> AnalysisStatusResponse {
        return try await apiClient.request(endpoint: .analysisStatus(id: id))
    }

    func getAnalysisResult(id: String) async throws -> AnalysisResultResponse {
        return try await apiClient.request(endpoint: .analysisResult(id: id))
    }

    func getHistoryDetail(id: String) async throws -> HistoryDetail {
        return try await apiClient.request(endpoint: .historyDetail(id: id))
    }

    func updateHistory(historyId: String, foods: [UpdateFoodItem]) async throws -> HistoryDetail {
        let request = UpdateHistoryRequest(foods: foods)
        return try await apiClient.request(endpoint: .updateHistory(id: historyId), body: request)
    }

    func deleteHistory(historyId: String) async throws {
        try await apiClient.requestWithoutResponse(endpoint: .deleteHistory(id: historyId))
    }
}

// MARK: - Shared Response Types

struct EmptyResponse: Decodable {}
