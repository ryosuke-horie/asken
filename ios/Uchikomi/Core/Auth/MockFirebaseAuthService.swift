#if DEBUG
import AuthenticationServices
import Foundation
import UchikomiCore

// MARK: - Development Constants

enum DevAuthConstants {
    static let mockToken = "dev-mock-token"
    static let mockUserID = "dev-mock-user"
    static let mockEmail = "dev@example.com"
    static let mockDisplayName = "Development User"
}

// MARK: - MockFirebaseAuthService

/// DEBUG + Simulator 環境でのみ使用するモック認証サービス
/// パスキー認証をバイパスして開発テストを可能にする
final class MockFirebaseAuthService: FirebaseAuthServiceProtocol {
    private var mockUser: FirebaseAuthUser?
    private var stateChangeListeners: [(AuthStateListenerHandle, (FirebaseAuthUser?) -> Void)] = []

    var currentUser: FirebaseAuthUser? { mockUser }
    var isSignedIn: Bool { mockUser != nil }

    /// 開発用のモックサインイン（Google/Apple共通）
    func signInWithMock() async throws -> FirebaseAuthUser {
        let user = FirebaseAuthUser(
            uid: DevAuthConstants.mockUserID,
            email: DevAuthConstants.mockEmail,
            displayName: DevAuthConstants.mockDisplayName
        )
        mockUser = user
        notifyStateChange(user)
        return user
    }

    func signInWithGoogle(credential _: GoogleCredential) async throws -> FirebaseAuthUser {
        try await signInWithMock()
    }

    func signInWithApple(credential _: ASAuthorizationAppleIDCredential, nonce _: String) async throws -> FirebaseAuthUser {
        try await signInWithMock()
    }

    func signOut() throws {
        mockUser = nil
        notifyStateChange(nil)
    }

    func getIDToken() async throws -> String {
        guard mockUser != nil else {
            throw FirebaseAuthError.notSignedIn
        }
        return DevAuthConstants.mockToken
    }

    func addStateDidChangeListener(_ listener: @escaping (FirebaseAuthUser?) -> Void) -> AuthStateListenerHandle {
        let handle = NSObject()
        stateChangeListeners.append((handle, listener))
        // 即座に現在の状態を通知
        listener(mockUser)
        return handle
    }

    func removeStateDidChangeListener(_ handle: AuthStateListenerHandle) {
        stateChangeListeners.removeAll { $0.0 === handle }
    }

    private func notifyStateChange(_ user: FirebaseAuthUser?) {
        for (_, listener) in stateChangeListeners {
            listener(user)
        }
    }
}
#endif
