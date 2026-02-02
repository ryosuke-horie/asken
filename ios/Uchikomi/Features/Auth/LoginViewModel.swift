import FirebaseCore
import Foundation
import GoogleSignIn
import UchikomiCore
import UIKit

@Observable
final class LoginViewModel {
    var isLoading = false
    var errorMessage: String?

    private let authManager: AuthManager

    init(authManager: AuthManager) {
        self.authManager = authManager
    }

    @MainActor
    func signInWithGoogle() async {
        isLoading = true
        errorMessage = nil

        do {
            // Google Sign-In の UI フローはここで処理
            guard let windowScene = UIApplication.shared.connectedScenes.first as? UIWindowScene,
                  let rootViewController = windowScene.windows.first?.rootViewController else {
                throw LoginError.noRootViewController
            }

            guard let clientID = FirebaseApp.app()?.options.clientID else {
                throw LoginError.firebaseNotConfigured
            }

            let config = GIDConfiguration(clientID: clientID)
            GIDSignIn.sharedInstance.configuration = config

            let result = try await GIDSignIn.sharedInstance.signIn(withPresenting: rootViewController)

            guard let idToken = result.user.idToken?.tokenString else {
                throw LoginError.tokenRetrievalFailed
            }

            let credential = GoogleCredential(
                idToken: idToken,
                accessToken: result.user.accessToken.tokenString
            )

            try await authManager.signInWithGoogle(credential: credential)
        } catch {
            errorMessage = error.localizedDescription
        }

        isLoading = false
    }

    @MainActor
    func signInWithApple() async {
        isLoading = true
        errorMessage = nil

        do {
            try await authManager.signInWithApple()
        } catch {
            errorMessage = error.localizedDescription
        }

        isLoading = false
    }
}

// MARK: - LoginError

enum LoginError: LocalizedError {
    case noRootViewController
    case firebaseNotConfigured
    case tokenRetrievalFailed

    var errorDescription: String? {
        switch self {
        case .noRootViewController:
            return "画面の取得に失敗しました"
        case .firebaseNotConfigured:
            return "Firebase の設定に問題があります"
        case .tokenRetrievalFailed:
            return "トークンの取得に失敗しました"
        }
    }
}
