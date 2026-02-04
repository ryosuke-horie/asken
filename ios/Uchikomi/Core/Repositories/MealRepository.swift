import Foundation

// MARK: - MealRepositoryProtocol

/// @mockable
protocol MealRepositoryProtocol {
    func getDailyMeals(date: Date) async throws -> DailyMeals
    func uploadImage(data: Data, filename: String, mealType: MealType, mealDate: Date) async throws -> String
    func analyzeText(inputText: String, mealType: MealType, mealDate: Date) async throws -> String
    func checkAnalysisStatus(id: String) async throws -> AnalysisStatusResponse
    func getAnalysisResult(id: String) async throws -> AnalysisResultResponse
    func getHistoryDetail(id: String) async throws -> HistoryDetail
    func updateHistory(historyId: String, foods: [UpdateFoodItem]) async throws -> HistoryDetail
    func deleteHistory(historyId: String) async throws
}

// MARK: - UpdateFoodItem

struct UpdateFoodItem: Encodable {
    let name: String
    let estimatedAmount: String
    let caloriesKcal: Double
    let proteinG: Double
    let fatG: Double
    let carbohydratesG: Double
}

// MARK: - UpdateHistoryRequest

struct UpdateHistoryRequest: Encodable {
    let foods: [UpdateFoodItem]
}

// MARK: - TextAnalyzeRequest

struct TextAnalyzeRequest: Encodable {
    let inputText: String
    let mealType: String
    let mealDate: String
    let tz: String

    enum CodingKeys: String, CodingKey {
        case inputText = "input_text"
        case mealType = "meal_type"
        case mealDate = "meal_date"
        case tz
    }
}

// MARK: - MealRepository

final class MealRepository: MealRepositoryProtocol {
    private let apiClient = APIClient.shared
    private let dateFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd"
        formatter.timeZone = TimeZone.current
        return formatter
    }()

    private var currentTimezoneIdentifier: String {
        TimeZone.current.identifier
    }

    func getDailyMeals(date: Date) async throws -> DailyMeals {
        let dateString = dateFormatter.string(from: date)
        return try await apiClient.request(endpoint: .dailyMeals(date: dateString, timezone: currentTimezoneIdentifier))
    }

    func uploadImage(data: Data, filename: String, mealType: MealType, mealDate: Date) async throws -> String {
        let response: AnalyzeResponse = try await apiClient.uploadImage(
            endpoint: .analyze,
            imageData: data,
            filename: filename,
            additionalFields: [
                "meal_type": mealType.rawValue,
                "meal_date": dateFormatter.string(from: mealDate),
                "tz": currentTimezoneIdentifier,
            ]
        )
        return response.id
    }

    func analyzeText(inputText: String, mealType: MealType, mealDate: Date) async throws -> String {
        let request = TextAnalyzeRequest(
            inputText: inputText,
            mealType: mealType.rawValue,
            mealDate: dateFormatter.string(from: mealDate),
            tz: currentTimezoneIdentifier
        )
        let response: AnalyzeResponse = try await apiClient.request(endpoint: .analyze, body: request)
        return response.id
    }

    func checkAnalysisStatus(id: String) async throws -> AnalysisStatusResponse {
        try await apiClient.request(endpoint: .analysisStatus(id: id))
    }

    func getAnalysisResult(id: String) async throws -> AnalysisResultResponse {
        try await apiClient.request(endpoint: .analysisResult(id: id))
    }

    func getHistoryDetail(id: String) async throws -> HistoryDetail {
        try await apiClient.request(endpoint: .historyDetail(id: id))
    }

    func updateHistory(historyId: String, foods: [UpdateFoodItem]) async throws -> HistoryDetail {
        let request = UpdateHistoryRequest(foods: foods)
        return try await apiClient.request(endpoint: .updateHistory(id: historyId), body: request)
    }

    func deleteHistory(historyId: String) async throws {
        try await apiClient.requestWithoutResponse(endpoint: .deleteHistory(id: historyId))
    }
}

// MARK: - EmptyResponse

struct EmptyResponse: Decodable {}
