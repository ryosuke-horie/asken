import AuthenticationServices
import FirebaseAuth
import FirebaseCore
import Foundation
import GoogleSignIn
import UchikomiCore

// MARK: - FirebaseAuthService

final class FirebaseAuthService: FirebaseAuthServiceProtocol {
    static let shared = FirebaseAuthService()

    private init() {
        // AuthManager が AppDelegate より先に初期化される場合があるため、
        // ここでも Firebase を初期化する
        if FirebaseApp.app() == nil {
            FirebaseApp.configure()
        }
    }

    // periphery:ignore - プロトコル準拠（AuthManagerがlistener経由で状態管理するため未使用）
    var currentUser: FirebaseAuthUser? {
        guard let user = Auth.auth().currentUser else { return nil }
        return FirebaseAuthUser(
            uid: user.uid,
            email: user.email,
            displayName: user.displayName
        )
    }

    // periphery:ignore - プロトコル準拠（AuthManagerがlistener経由で状態管理するため未使用）
    var isSignedIn: Bool {
        Auth.auth().currentUser != nil
    }

    func signInWithGoogle(credential: GoogleCredential) async throws -> FirebaseAuthUser {
        let firebaseCredential = GoogleAuthProvider.credential(
            withIDToken: credential.idToken,
            accessToken: credential.accessToken
        )

        let authResult = try await Auth.auth().signIn(with: firebaseCredential)

        return FirebaseAuthUser(
            uid: authResult.user.uid,
            email: authResult.user.email,
            displayName: authResult.user.displayName
        )
    }

    // periphery:ignore - Apple Sign-In用（Apple Developer Program登録後に有効化）
    func signInWithApple(credential: ASAuthorizationAppleIDCredential, nonce: String) async throws -> FirebaseAuthUser {
        guard let appleIDToken = credential.identityToken,
              let idTokenString = String(data: appleIDToken, encoding: .utf8) else {
            throw FirebaseAuthError.tokenRetrievalFailed
        }

        let oAuthCredential = OAuthProvider.appleCredential(
            withIDToken: idTokenString,
            rawNonce: nonce,
            fullName: credential.fullName
        )

        let authResult = try await Auth.auth().signIn(with: oAuthCredential)

        return FirebaseAuthUser(
            uid: authResult.user.uid,
            email: authResult.user.email,
            displayName: authResult.user.displayName
        )
    }

    func signOut() throws {
        try Auth.auth().signOut()
        GIDSignIn.sharedInstance.signOut()
        #if DEBUG
        debugPrint("[FirebaseAuthService] User signed out from Firebase and Google")
        #endif
    }

    func getIDToken() async throws -> String {
        guard let user = Auth.auth().currentUser else {
            throw FirebaseAuthError.notSignedIn
        }
        return try await user.getIDToken()
    }

    func addStateDidChangeListener(_ listener: @escaping (FirebaseAuthUser?) -> Void) -> AuthStateListenerHandle {
        Auth.auth().addStateDidChangeListener { _, user in
            if let user {
                listener(FirebaseAuthUser(
                    uid: user.uid,
                    email: user.email,
                    displayName: user.displayName
                ))
            } else {
                listener(nil)
            }
        }
    }

    func removeStateDidChangeListener(_ handle: AuthStateListenerHandle) {
        Auth.auth().removeStateDidChangeListener(handle)
    }
}
