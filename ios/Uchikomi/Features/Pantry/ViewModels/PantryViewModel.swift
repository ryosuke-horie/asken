import Foundation

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

        do {
            ingredients = try await repository.fetchIngredients(category: nil)
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "食材の取得に失敗しました"
        }

        isLoading = false
    }

    func deleteIngredient(id: String) async {
        errorMessage = nil

        do {
            try await repository.deleteIngredient(id: id)
            ingredients.removeAll { $0.id == id }
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "削除に失敗しました"
        }
    }

    func addIngredient(_ ingredient: Ingredient) {
        ingredients.append(ingredient)
    }

    func updateIngredient(_ updated: Ingredient) {
        if let index = ingredients.firstIndex(where: { $0.id == updated.id }) {
            ingredients[index] = updated
        }
    }
}
