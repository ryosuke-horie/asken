import Foundation
import UchikomiCore

// MARK: - SettingsViewModel

@Observable
final class SettingsViewModel {
    var showLogoutError = false
    var logoutErrorMessage: String?

    private let authManager: AuthManager

    init(authManager: AuthManager) {
        self.authManager = authManager
    }

    var userName: String? {
        authManager.currentUser?.name
    }

    var userEmail: String {
        authManager.currentUser?.email ?? ""
    }

    @MainActor
    func logout() {
        do {
            try authManager.logout()
        } catch {
            logoutErrorMessage = "ログアウトに失敗しました"
            showLogoutError = true
        }
    }
}
