import Foundation

// MARK: - IngredientRepositoryProtocol

/// @mockable
protocol IngredientRepositoryProtocol {
    func fetchIngredients(category: IngredientCategory?) async throws -> [Ingredient]
    func createIngredient(_ request: CreateIngredientRequest) async throws -> Ingredient
    func updateIngredient(id: String, request: CreateIngredientRequest) async throws -> Ingredient
    func deleteIngredient(id: String) async throws
    func scanReceipt(imageData: Data) async throws -> [ScannedIngredient]
}

// MARK: - CreateIngredientRequest

struct CreateIngredientRequest: Encodable {
    let name: String
    let category: String
    let quantity: Double
    let unit: String
    let purchaseDate: String?
    let expiryDate: String?
    let source: String
}

// MARK: - IngredientsListResponse

private struct IngredientsListResponse: Decodable {
    let ingredients: [Ingredient]
}

// MARK: - ScannedIngredientResponse

private struct ScannedIngredientResponse: Decodable {
    let name: String
    let category: IngredientCategory
    let quantity: Double
    let unit: String
    let source: String
}

// MARK: - ScanReceiptAPIResponse

private struct ScanReceiptAPIResponse: Decodable {
    let ingredients: [ScannedIngredientResponse]
}

// MARK: - IngredientRepository

final class IngredientRepository: IngredientRepositoryProtocol {
    private let client: APIClient

    init(client: APIClient = .shared) {
        self.client = client
    }

    func fetchIngredients(category: IngredientCategory? = nil) async throws -> [Ingredient] {
        let endpoint = APIEndpoint.ingredientsList(category: category?.rawValue)
        let response: IngredientsListResponse = try await client.request(endpoint: endpoint)
        return response.ingredients
    }

    func createIngredient(_ request: CreateIngredientRequest) async throws -> Ingredient {
        try await client.request(endpoint: .createIngredient, body: request)
    }

    func updateIngredient(id: String, request: CreateIngredientRequest) async throws -> Ingredient {
        try await client.request(endpoint: .updateIngredient(id: id), body: request)
    }

    func deleteIngredient(id: String) async throws {
        try await client.requestWithoutResponse(endpoint: .deleteIngredient(id: id))
    }

    func scanReceipt(imageData: Data) async throws -> [ScannedIngredient] {
        let response: ScanReceiptAPIResponse = try await client.uploadImageDecoded(
            endpoint: .scanReceipt,
            imageData: imageData,
            filename: "receipt.jpg"
        )
        return response.ingredients.map { item in
            ScannedIngredient(
                name: item.name,
                category: item.category,
                quantity: item.quantity,
                unit: item.unit
            )
        }
    }
}
