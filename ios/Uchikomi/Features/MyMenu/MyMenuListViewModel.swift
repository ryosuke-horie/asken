import Foundation

@Observable
final class MyMenuListViewModel {
    var myMenuItems: [MyMenuItem] = []
    var isLoading = false
    var errorMessage: String?
    var isDeleting = false

    let repository: MyMenuRepositoryProtocol

    init(repository: MyMenuRepositoryProtocol = MyMenuRepository()) {
        self.repository = repository
    }

    func loadMyMenuList() async {
        isLoading = true
        errorMessage = nil

        do {
            myMenuItems = try await repository.fetchMyMenuList()
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "マイメニューの取得に失敗しました"
        }

        isLoading = false
    }

    func deleteMyMenu(id: String) async throws {
        isDeleting = true
        errorMessage = nil
        defer { isDeleting = false }

        do {
            try await repository.deleteMyMenu(id: id)
            myMenuItems.removeAll { $0.id == id }
        } catch let error as APIError {
            errorMessage = error.localizedDescription
            throw error
        } catch {
            errorMessage = "削除に失敗しました"
            throw error
        }
    }
}
