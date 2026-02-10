import Foundation

@Observable
final class MyMenuEditViewModel {
    var menuName: String = ""
    var foodItems: [FoodEditItem] = []
    var isLoading = false
    var isSaving = false
    var errorMessage: String?
    var shouldDismiss = false

    private let repository: MyMenuRepositoryProtocol
    private let existingMenuItem: MyMenuItem?

    init(repository: MyMenuRepositoryProtocol = MyMenuRepository(), menuItem: MyMenuItem? = nil) {
        self.repository = repository
        self.existingMenuItem = menuItem

        if let item = menuItem {
            menuName = item.name
            foodItems = item.foods.map { FoodEditItem(from: $0) }
        }
    }

    var isEditMode: Bool {
        existingMenuItem != nil
    }

    var totalCalories: Double {
        foodItems.reduce(0) { $0 + $1.calories }
    }

    var totalProtein: Double {
        foodItems.reduce(0) { $0 + $1.protein }
    }

    var totalFat: Double {
        foodItems.reduce(0) { $0 + $1.fat }
    }

    var totalCarbohydrates: Double {
        foodItems.reduce(0) { $0 + $1.carbohydrates }
    }

    var isValid: Bool {
        !menuName.isEmpty && !foodItems.isEmpty
    }

    func addFoodItem() {
        foodItems.append(FoodEditItem())
    }

    func removeFoodItem(at index: Int) {
        foodItems.remove(at: index)
    }

    func save() async {
        guard isValid else { return }

        isSaving = true
        errorMessage = nil

        let foods = foodItems.map { item in
            NutritionInfo(
                name: item.name,
                estimatedAmount: item.quantity,
                caloriesKcal: item.calories,
                proteinG: item.protein,
                fatG: item.fat,
                carbohydratesG: item.carbohydrates
            )
        }

        do {
            if let item = existingMenuItem {
                _ = try await repository.updateMyMenu(id: item.id, name: menuName, foods: foods)
            } else {
                _ = try await repository.createMyMenu(name: menuName, foods: foods)
            }
            shouldDismiss = true // 成功フラグ（画面を閉じるため）
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "保存に失敗しました"
        }

        isSaving = false
    }

    func delete() async {
        guard let item = existingMenuItem else { return }

        isSaving = true
        errorMessage = nil

        do {
            try await repository.deleteMyMenu(id: item.id)
            shouldDismiss = true
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "削除に失敗しました"
        }

        isSaving = false
    }
}
