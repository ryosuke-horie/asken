import AuthenticationServices
import CryptoKit
import Foundation
import UchikomiCore

// MARK: - AppleSignInManager

public final class AppleSignInManager: NSObject, AppleSignInManagerProtocol {
    private var currentNonce: String?
    private var continuation: CheckedContinuation<ASAuthorizationAppleIDCredential, Error>?

    public func signIn() async throws -> (credential: ASAuthorizationAppleIDCredential, nonce: String) {
        let nonce = try randomNonceString()
        currentNonce = nonce

        let appleIDProvider = ASAuthorizationAppleIDProvider()
        let request = appleIDProvider.createRequest()
        request.requestedScopes = [.fullName, .email]
        request.nonce = sha256(nonce)

        let credential = try await withCheckedThrowingContinuation { continuation in
            self.continuation = continuation

            let authorizationController = ASAuthorizationController(authorizationRequests: [request])
            authorizationController.delegate = self
            authorizationController.performRequests()
        }

        return (credential, nonce)
    }

    private func randomNonceString(length: Int = 32) throws -> String {
        precondition(length > 0)
        var randomBytes = [UInt8](repeating: 0, count: length)
        let errorCode = SecRandomCopyBytes(kSecRandomDefault, randomBytes.count, &randomBytes)
        if errorCode != errSecSuccess {
            throw AppleSignInError.nonceGenerationFailed(osStatus: errorCode)
        }

        let charset: [Character] = Array("0123456789ABCDEFGHIJKLMNOPQRSTUVXYZabcdefghijklmnopqrstuvwxyz-._")
        let nonce = randomBytes.map { byte in
            charset[Int(byte) % charset.count]
        }

        return String(nonce)
    }

    private func sha256(_ input: String) -> String {
        let inputData = Data(input.utf8)
        let hashedData = SHA256.hash(data: inputData)
        return hashedData.compactMap {
            String(format: "%02x", $0)
        }.joined()
    }
}

// MARK: ASAuthorizationControllerDelegate

extension AppleSignInManager: ASAuthorizationControllerDelegate {
    public func authorizationController(
        controller _: ASAuthorizationController,
        didCompleteWithAuthorization authorization: ASAuthorization
    ) {
        guard let appleIDCredential = authorization.credential as? ASAuthorizationAppleIDCredential else {
            continuation?.resume(throwing: AppleSignInError.invalidCredential)
            continuation = nil
            return
        }
        continuation?.resume(returning: appleIDCredential)
        continuation = nil
    }

    public func authorizationController(
        controller _: ASAuthorizationController,
        didCompleteWithError error: Error
    ) {
        continuation?.resume(throwing: error)
        continuation = nil
    }
}

// MARK: - AppleSignInError

public enum AppleSignInError: LocalizedError {
    case invalidCredential
    case nonceGenerationFailed(osStatus: OSStatus)

    public var errorDescription: String? {
        switch self {
        case .invalidCredential:
            "Apple IDの認証情報が無効です"
        case let .nonceGenerationFailed(osStatus):
            "ノンス生成に失敗しました (OSStatus: \(osStatus))"
        }
    }
}
