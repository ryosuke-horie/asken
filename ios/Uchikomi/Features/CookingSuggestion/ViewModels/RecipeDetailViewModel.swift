import Foundation
import os

private let logger = Logger(
    subsystem: Bundle.main.bundleIdentifier ?? "Uchikomi",
    category: "RecipeDetailViewModel"
)

// MARK: - RecipeDetailViewModel

@MainActor
@Observable
final class RecipeDetailViewModel {
    var suggestion: MenuSuggestion
    var isLoadingRecipe = false
    var isAccepting = false
    var acceptResult: AcceptMenuSuggestionResult?
    var errorMessage: String?

    let repository: MenuSuggestionRepositoryProtocol

    init(
        suggestion: MenuSuggestion,
        repository: MenuSuggestionRepositoryProtocol = MenuSuggestionRepository()
    ) {
        self.suggestion = suggestion
        self.repository = repository
    }

    func loadRecipe() async {
        guard suggestion.recipe == nil else { return }

        isLoadingRecipe = true
        errorMessage = nil
        defer { isLoadingRecipe = false }

        do {
            suggestion = try await repository.fetchSuggestionDetail(id: suggestion.id)
        } catch let error as APIError {
            logger.error("レシピ取得でAPIエラー: id=\(self.suggestion.id), error=\(error.localizedDescription)")
            errorMessage = "レシピの取得に失敗しました"
        } catch {
            logger.error("レシピ取得で予期しないエラー: id=\(self.suggestion.id), error=\(error.localizedDescription)")
            errorMessage = "レシピの取得に失敗しました"
        }
    }

    func accept() async -> Bool {
        isAccepting = true
        errorMessage = nil
        defer { isAccepting = false }

        do {
            acceptResult = try await repository.acceptSuggestion(id: suggestion.id)
            return true
        } catch let error as APIError {
            logger.error("サジェスト採用でAPIエラー: id=\(self.suggestion.id), error=\(error.localizedDescription)")
            errorMessage = error.localizedDescription
            return false
        } catch {
            logger.error("サジェスト採用で予期しないエラー: id=\(self.suggestion.id), error=\(error.localizedDescription)")
            errorMessage = "サジェストの採用に失敗しました"
            return false
        }
    }
}
