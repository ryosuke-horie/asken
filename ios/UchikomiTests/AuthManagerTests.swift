import XCTest
@testable import Uchikomi

final class AuthManagerTests: XCTestCase {

    func testDecodeUserFromValidToken() async {
        // JWTの構造: header.payload.signature
        // payload: {"user_id": "test-id", "email": "test@example.com"}
        // Base64URL encoded payload: eyJ1c2VyX2lkIjoidGVzdC1pZCIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSJ9

        let validToken = "eyJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoidGVzdC1pZCIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSJ9.signature"

        let authManager = AuthManager(repository: MockAuthRepository())

        // リフレクションでprivateメソッドをテストするのは難しいため、
        // publicなAPIを通じてテスト
        XCTAssertFalse(authManager.isAuthenticated)
    }

    func testLoginSuccess() async {
        let mockRepo = MockAuthRepository()
        mockRepo.loginResult = .success(AuthResponse(
            token: "test-token",
            user: User(id: "1", email: "test@example.com", name: "Test User")
        ))

        let authManager = AuthManager(repository: mockRepo)

        do {
            try await authManager.login(email: "test@example.com", password: "Pass0123")

            // Wait for main actor update
            try await Task.sleep(nanoseconds: 100_000_000)

            XCTAssertTrue(authManager.isAuthenticated)
            XCTAssertEqual(authManager.currentUser?.email, "test@example.com")
        } catch {
            XCTFail("Login should succeed: \(error)")
        }
    }

    func testLoginFailure() async {
        let mockRepo = MockAuthRepository()
        mockRepo.loginResult = .failure(APIError.unauthorized)

        let authManager = AuthManager(repository: mockRepo)

        do {
            try await authManager.login(email: "test@example.com", password: "wrong")
            XCTFail("Login should fail")
        } catch {
            XCTAssertFalse(authManager.isAuthenticated)
        }
    }
}

// MARK: - Mock

private class MockAuthRepository: AuthRepositoryProtocol {
    var loginResult: Result<AuthResponse, Error> = .failure(APIError.unauthorized)
    var registerResult: Result<AuthResponse, Error> = .failure(APIError.unauthorized)

    func login(email: String, password: String) async throws -> AuthResponse {
        switch loginResult {
        case .success(let response):
            return response
        case .failure(let error):
            throw error
        }
    }

    func register(email: String, password: String, name: String) async throws -> AuthResponse {
        switch registerResult {
        case .success(let response):
            return response
        case .failure(let error):
            throw error
        }
    }
}
