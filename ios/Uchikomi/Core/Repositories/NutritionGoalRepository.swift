import Foundation

// MARK: - NutritionGoalRepositoryProtocol

/// @mockable
protocol NutritionGoalRepositoryProtocol {
    func getGoal(currentWeight: Double?, goalWeight: Double?) async throws -> NutritionGoal?
    func setGoal(targetCalories: Double) async throws -> NutritionGoal
}

// MARK: - NutritionGoalRepository

final class NutritionGoalRepository: NutritionGoalRepositoryProtocol {
    private let apiClient = APIClient.shared

    func getGoal(currentWeight: Double?, goalWeight: Double?) async throws -> NutritionGoal? {
        let response: NutritionGoalNullableResponse = try await apiClient.request(
            endpoint: .getNutritionGoal(currentWeight: currentWeight, goalWeight: goalWeight)
        )
        return response.goal
    }

    func setGoal(targetCalories: Double) async throws -> NutritionGoal {
        let request = SetNutritionGoalRequest(targetCalories: targetCalories)
        return try await apiClient.request(endpoint: .setNutritionGoal, body: request)
    }
}
