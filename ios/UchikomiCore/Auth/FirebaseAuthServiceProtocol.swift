import AuthenticationServices
import Foundation

// MARK: - AuthStateListenerHandle

public typealias AuthStateListenerHandle = NSObjectProtocol

// MARK: - FirebaseAuthUser

public struct FirebaseAuthUser: Equatable {
    public let uid: String
    public let email: String?
    public let displayName: String?

    public init(uid: String, email: String?, displayName: String?) {
        self.uid = uid
        self.email = email
        self.displayName = displayName
    }
}

// MARK: - GoogleCredential

public struct GoogleCredential {
    public let idToken: String
    public let accessToken: String

    public init(idToken: String, accessToken: String) {
        self.idToken = idToken
        self.accessToken = accessToken
    }
}

// MARK: - FirebaseAuthServiceProtocol

/// @mockable
public protocol FirebaseAuthServiceProtocol {
    var currentUser: FirebaseAuthUser? { get } // periphery:ignore - プロトコルのAPI完全性のため保持
    var isSignedIn: Bool { get } // periphery:ignore - プロトコルのAPI完全性のため保持

    func signInWithGoogle(credential: GoogleCredential) async throws -> FirebaseAuthUser
    // periphery:ignore - Apple Sign-In用（Apple Developer Program登録後に有効化）
    func signInWithApple(credential: ASAuthorizationAppleIDCredential, nonce: String) async throws -> FirebaseAuthUser
    func signOut() throws
    func getIDToken() async throws -> String
    func addStateDidChangeListener(_ listener: @escaping (FirebaseAuthUser?) -> Void) -> AuthStateListenerHandle
    func removeStateDidChangeListener(_ handle: AuthStateListenerHandle)
}

// MARK: - FirebaseAuthError

public enum FirebaseAuthError: LocalizedError {
    case notSignedIn
    case tokenRetrievalFailed
    case configurationError

    public var errorDescription: String? {
        switch self {
        case .notSignedIn:
            "サインインしていません"
        case .tokenRetrievalFailed:
            "トークンの取得に失敗しました"
        case .configurationError:
            "Firebase の設定に問題があります"
        }
    }
}
