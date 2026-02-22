import Foundation
import os

private let logger = Logger(
    subsystem: Bundle.main.bundleIdentifier ?? "Uchikomi",
    category: "CookingSuggestionViewModel"
)

// MARK: - CookingSuggestionViewModel

@MainActor
@Observable
final class CookingSuggestionViewModel {
    var selectedMealType: MealType = .dinner
    var suggestionCount = 3
    var suggestions: [MenuSuggestion] = []
    var isGenerating = false
    var errorMessage: String?

    private let repository: MenuSuggestionRepositoryProtocol

    init(repository: MenuSuggestionRepositoryProtocol = MenuSuggestionRepository()) {
        self.repository = repository
    }

    func generateSuggestions() async {
        isGenerating = true
        errorMessage = nil
        defer { isGenerating = false }

        do {
            suggestions = try await repository.suggestMenu(
                mealType: selectedMealType,
                count: suggestionCount
            )
        } catch is CancellationError {
            return
        } catch let error as APIError {
            logger.error("サジェスト生成でAPIエラー: \(error.localizedDescription)")
            errorMessage = error.localizedDescription
        } catch {
            logger.error("サジェスト生成で予期しないエラー: \(error.localizedDescription)")
            errorMessage = "メニューサジェストの生成に失敗しました"
        }
    }

    func dismissSuggestion(id: String) async {
        errorMessage = nil

        do {
            try await repository.dismissSuggestion(id: id)
            suggestions.removeAll { $0.id == id }
        } catch is CancellationError {
            return
        } catch let error as APIError {
            logger.error("サジェスト却下でAPIエラー: id=\(id), error=\(error.localizedDescription)")
            errorMessage = error.localizedDescription
        } catch {
            logger.error("サジェスト却下で予期しないエラー: id=\(id), error=\(error.localizedDescription)")
            errorMessage = "サジェストの却下に失敗しました"
        }
    }
}
