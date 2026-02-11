import Foundation

// MARK: - MyMenuRepositoryProtocol

/// @mockable
protocol MyMenuRepositoryProtocol {
    func fetchMyMenuList() async throws -> [MyMenuItem]
    func createMyMenu(name: String, foods: [NutritionInfo]) async throws -> MyMenuItem
    func updateMyMenu(id: String, name: String, foods: [NutritionInfo]) async throws -> MyMenuItem
    func deleteMyMenu(id: String) async throws
    func recordFromMyMenu(id: String, mealType: MealType, mealDate: Date) async throws -> String
}

// MARK: - CreateMyMenuRequest

struct CreateMyMenuRequest: Encodable {
    let name: String
    let foods: [NutritionInfo]
}

// MARK: - UpdateMyMenuRequest

struct UpdateMyMenuRequest: Encodable {
    let name: String
    let foods: [NutritionInfo]
}

// MARK: - RecordMyMenuRequest

struct RecordMyMenuRequest: Encodable {
    let mealType: String
    let mealDate: String

    enum CodingKeys: String, CodingKey {
        case mealType = "meal_type"
        case mealDate = "meal_date"
    }
}

// MARK: - AnalysisIDResponse

struct AnalysisIDResponse: Decodable {
    let id: String
}

// MARK: - MyMenuRepository

final class MyMenuRepository: MyMenuRepositoryProtocol {
    private let apiClient = APIClient.shared
    private let dateFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd"
        formatter.timeZone = TimeZone.current
        return formatter
    }()

    func fetchMyMenuList() async throws -> [MyMenuItem] {
        try await apiClient.request(endpoint: .myMenuList)
    }

    func createMyMenu(name: String, foods: [NutritionInfo]) async throws -> MyMenuItem {
        let request = CreateMyMenuRequest(name: name, foods: foods)
        return try await apiClient.request(endpoint: .createMyMenu, body: request)
    }

    func updateMyMenu(id: String, name: String, foods: [NutritionInfo]) async throws -> MyMenuItem {
        let request = UpdateMyMenuRequest(name: name, foods: foods)
        return try await apiClient.request(endpoint: .updateMyMenu(id: id), body: request)
    }

    func deleteMyMenu(id: String) async throws {
        try await apiClient.requestWithoutResponse(endpoint: .deleteMyMenu(id: id))
    }

    func recordFromMyMenu(id: String, mealType: MealType, mealDate: Date) async throws -> String {
        let request = RecordMyMenuRequest(
            mealType: mealType.rawValue,
            mealDate: dateFormatter.string(from: mealDate)
        )
        let response: AnalysisIDResponse = try await apiClient.request(
            endpoint: .recordMyMenu(id: id),
            body: request
        )
        return response.id
    }
}
