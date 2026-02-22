import Foundation

// MARK: - MenuSuggestionRepositoryProtocol

/// @mockable
protocol MenuSuggestionRepositoryProtocol {
    func suggestMenu(mealType: MealType, count: Int) async throws -> [MenuSuggestion]
    func fetchSuggestions(status: String?, limit: Int) async throws -> [MenuSuggestion]
    func fetchSuggestionDetail(id: String) async throws -> MenuSuggestion
    func acceptSuggestion(id: String) async throws -> AcceptMenuSuggestionResult
    func dismissSuggestion(id: String) async throws
}

// MARK: - MenuSuggestionRepository

final class MenuSuggestionRepository: MenuSuggestionRepositoryProtocol {
    private let client: APIClient

    init(client: APIClient = .shared) {
        self.client = client
    }

    func suggestMenu(mealType: MealType, count: Int) async throws -> [MenuSuggestion] {
        let request = SuggestMenuRequest(mealType: mealType.rawValue, count: count)
        let response: MenuSuggestionListResponse = try await client.request(
            endpoint: .suggestMenu,
            body: request
        )
        return response.suggestions
    }

    func fetchSuggestions(status: String? = nil, limit: Int = 10) async throws -> [MenuSuggestion] {
        let response: MenuSuggestionListResponse = try await client.request(
            endpoint: .menuSuggestions(status: status, limit: limit)
        )
        return response.suggestions
    }

    func fetchSuggestionDetail(id: String) async throws -> MenuSuggestion {
        try await client.request(endpoint: .menuSuggestionDetail(id: id))
    }

    func acceptSuggestion(id: String) async throws -> AcceptMenuSuggestionResult {
        try await client.request(endpoint: .acceptMenuSuggestion(id: id))
    }

    func dismissSuggestion(id: String) async throws {
        try await client.requestWithoutResponse(endpoint: .dismissMenuSuggestion(id: id))
    }
}
