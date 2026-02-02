import AuthenticationServices
import Foundation

// MARK: - AppleSignInManagerProtocol

/// @mockable
public protocol AppleSignInManagerProtocol {
    func signIn() async throws -> (credential: ASAuthorizationAppleIDCredential, nonce: String)
}

// MARK: - AuthManager

@Observable
public final class AuthManager {
    public var currentUser: User?
    public var isAuthenticated: Bool {
        currentUser != nil
    }

    private let firebaseAuthService: FirebaseAuthServiceProtocol
    private let appleSignInManager: AppleSignInManagerProtocol
    private var authStateListenerHandle: AuthStateListenerHandle?

    public init(
        firebaseAuthService: FirebaseAuthServiceProtocol,
        appleSignInManager: AppleSignInManagerProtocol
    ) {
        self.firebaseAuthService = firebaseAuthService
        self.appleSignInManager = appleSignInManager
        setupAuthStateListener()
    }

    deinit {
        if let handle = authStateListenerHandle {
            firebaseAuthService.removeStateDidChangeListener(handle)
        }
    }

    @MainActor
    public func signInWithGoogle(credential: GoogleCredential) async throws {
        let firebaseUser = try await firebaseAuthService.signInWithGoogle(credential: credential)
        currentUser = User(
            id: firebaseUser.uid,
            email: firebaseUser.email ?? "",
            name: firebaseUser.displayName
        )
    }

    @MainActor
    public func signInWithApple() async throws {
        let (credential, nonce) = try await appleSignInManager.signIn()
        let firebaseUser = try await firebaseAuthService.signInWithApple(credential: credential, nonce: nonce)
        currentUser = User(
            id: firebaseUser.uid,
            email: firebaseUser.email ?? "",
            name: firebaseUser.displayName
        )
    }

    @MainActor
    public func logout() throws {
        try firebaseAuthService.signOut()
        currentUser = nil
    }

    private func setupAuthStateListener() {
        authStateListenerHandle = firebaseAuthService.addStateDidChangeListener { [weak self] firebaseUser in
            Task { @MainActor in
                if let firebaseUser {
                    self?.currentUser = User(
                        id: firebaseUser.uid,
                        email: firebaseUser.email ?? "",
                        name: firebaseUser.displayName
                    )
                } else {
                    self?.currentUser = nil
                }
            }
        }
    }
}
