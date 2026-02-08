import Foundation
import os
import UchikomiCore

private let logger = Logger(
    subsystem: Bundle.main.bundleIdentifier ?? "Uchikomi",
    category: "SettingsViewModel"
)

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
            logger.error("ログアウト失敗: \(error.localizedDescription)")
            logoutErrorMessage = "ログアウトに失敗しました"
            showLogoutError = true
        }
    }
}
