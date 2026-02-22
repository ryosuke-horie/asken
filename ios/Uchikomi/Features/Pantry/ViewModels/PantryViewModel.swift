import Foundation
import os

private let logger = Logger(subsystem: Bundle.main.bundleIdentifier ?? "Uchikomi", category: "PantryViewModel")

// MARK: - PantryViewModel

@MainActor
@Observable
final class PantryViewModel {
    var ingredients: [Ingredient] = []
    var isLoading = false
    var errorMessage: String?

    let repository: IngredientRepositoryProtocol

    init(repository: IngredientRepositoryProtocol = IngredientRepository()) {
        self.repository = repository
    }

    /// カテゴリ別にグループ化した食材一覧（表示順: カテゴリのdisplayName順）
    var groupedIngredients: [(category: IngredientCategory, items: [Ingredient])] {
        let grouped = Dictionary(grouping: ingredients, by: \.category)
        return IngredientCategory.allCases.compactMap { category in
            guard let items = grouped[category], !items.isEmpty else { return nil }
            return (category: category, items: items)
        }
    }

    func loadIngredients() async {
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            ingredients = try await repository.fetchIngredients(category: nil)
        } catch let error as APIError {
            logger.error("食材取得でAPIエラー: \(error.localizedDescription)")
            errorMessage = error.localizedDescription
        } catch {
            logger.error("食材取得で予期しないエラー: \(error.localizedDescription)")
            errorMessage = "食材の取得に失敗しました"
        }
    }

    func deleteIngredient(id: String) async {
        errorMessage = nil

        do {
            try await repository.deleteIngredient(id: id)
            ingredients.removeAll { $0.id == id }
        } catch let error as APIError {
            logger.error("食材削除でAPIエラー: id=\(id), error=\(error.localizedDescription)")
            errorMessage = error.localizedDescription
        } catch {
            logger.error("食材削除で予期しないエラー: id=\(id), error=\(error.localizedDescription)")
            errorMessage = "削除に失敗しました"
        }
    }

    func addIngredient(_ ingredient: Ingredient) {
        ingredients.append(ingredient)
    }

    func updateIngredient(_ updated: Ingredient) {
        guard let index = ingredients.firstIndex(where: { $0.id == updated.id }) else {
            logger.error("updateIngredient: 食材ID \(updated.id) がローカルリストに存在しない")
            return
        }
        ingredients[index] = updated
    }
}
