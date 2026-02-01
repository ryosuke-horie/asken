import Testing
@testable import Uchikomi

@Suite
struct AuthManagerTests {
    @Test
    func 初期状態で認証されていないべき() {
        let authManager = AuthManager(repository: AuthRepositoryProtocolMock())
        #expect(authManager.isAuthenticated == false)
    }

    @Test
    func ログイン成功時に認証状態がtrueになるべき() async throws {
        let mockRepo = AuthRepositoryProtocolMock()
        mockRepo.loginHandler = { _, _ in
            AuthResponse(
                token: "test-token",
                user: User(id: "1", email: "test@example.com", name: "Test User")
            )
        }

        let authManager = AuthManager(repository: mockRepo)

        try await authManager.login(email: "test@example.com", password: "Pass0123")

        try await Task.sleep(nanoseconds: 100_000_000)

        #expect(authManager.isAuthenticated == true)
        #expect(authManager.currentUser?.email == "test@example.com")
    }

    @Test
    func ログイン失敗時に認証状態がfalseのままであるべき() async {
        let mockRepo = AuthRepositoryProtocolMock()
        mockRepo.loginHandler = { _, _ in
            throw APIError.unauthorized
        }

        let authManager = AuthManager(repository: mockRepo)

        do {
            try await authManager.login(email: "test@example.com", password: "wrong")
            Issue.record("Login should fail")
        } catch {
            #expect(authManager.isAuthenticated == false)
        }
    }
}
