import AuthenticationServices
import Foundation
import Testing
@testable import UchikomiCore

class MockAuthStateListener: NSObject {}

@Suite
struct AuthManagerTests {
    @Test
    func 初期状態で認証されていないべき() {
        let mockFirebaseService = FirebaseAuthServiceProtocolMock()
        mockFirebaseService.addStateDidChangeListenerHandler = { _ in
            MockAuthStateListener() as AuthStateListenerHandle
        }
        let mockAppleSignIn = AppleSignInManagerProtocolMock()

        let authManager = AuthManager(
            firebaseAuthService: mockFirebaseService,
            appleSignInManager: mockAppleSignIn
        )
        #expect(authManager.isAuthenticated == false)
    }

    @Test
    @MainActor
    func ログアウト成功時に認証状態がnilになるべき() throws {
        let mockFirebaseService = FirebaseAuthServiceProtocolMock()
        mockFirebaseService.addStateDidChangeListenerHandler = { _ in
            MockAuthStateListener() as AuthStateListenerHandle
        }
        mockFirebaseService.signOutHandler = {}
        let mockAppleSignIn = AppleSignInManagerProtocolMock()

        let authManager = AuthManager(
            firebaseAuthService: mockFirebaseService,
            appleSignInManager: mockAppleSignIn
        )

        try authManager.logout()

        #expect(authManager.currentUser == nil)
        #expect(mockFirebaseService.signOutCallCount == 1)
    }

    @Test
    @MainActor
    func ログアウト失敗時にエラーがスローされるべき() {
        let mockFirebaseService = FirebaseAuthServiceProtocolMock()
        mockFirebaseService.addStateDidChangeListenerHandler = { _ in
            MockAuthStateListener() as AuthStateListenerHandle
        }
        mockFirebaseService.signOutHandler = {
            throw FirebaseAuthError.notSignedIn
        }
        let mockAppleSignIn = AppleSignInManagerProtocolMock()

        let authManager = AuthManager(
            firebaseAuthService: mockFirebaseService,
            appleSignInManager: mockAppleSignIn
        )

        #expect(throws: FirebaseAuthError.self) {
            try authManager.logout()
        }
    }

    @Test
    @MainActor
    func Googleサインイン成功時にユーザーが設定されるべき() async throws {
        let mockFirebaseService = FirebaseAuthServiceProtocolMock()
        mockFirebaseService.addStateDidChangeListenerHandler = { _ in
            MockAuthStateListener() as AuthStateListenerHandle
        }
        mockFirebaseService.signInWithGoogleHandler = { _ in
            FirebaseAuthUser(uid: "test-uid", email: "test@example.com", displayName: "Test User")
        }
        let mockAppleSignIn = AppleSignInManagerProtocolMock()

        let authManager = AuthManager(
            firebaseAuthService: mockFirebaseService,
            appleSignInManager: mockAppleSignIn
        )

        let credential = GoogleCredential(idToken: "test-id-token", accessToken: "test-access-token")
        try await authManager.signInWithGoogle(credential: credential)

        #expect(authManager.currentUser?.id == "test-uid")
        #expect(authManager.currentUser?.email == "test@example.com")
        #expect(authManager.currentUser?.name == "Test User")
        #expect(mockFirebaseService.signInWithGoogleCallCount == 1)
    }
}
