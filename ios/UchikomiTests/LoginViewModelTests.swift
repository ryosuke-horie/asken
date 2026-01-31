import Testing
@testable import Uchikomi

@Suite struct LoginViewModelTests {

    // MARK: - isValid Tests

    @Suite("isValid")
    struct IsValidTests {

        @Test func 有効な入力でtrueを返すべき() {
            let authManager = AuthManager(repository: AuthRepositoryProtocolMock())
            let viewModel = LoginViewModel(authManager: authManager)

            viewModel.email = "test@example.com"
            viewModel.password = "Pass0123"

            #expect(viewModel.isValid == true)
        }

        @Test func 空のemailでfalseを返すべき() {
            let authManager = AuthManager(repository: AuthRepositoryProtocolMock())
            let viewModel = LoginViewModel(authManager: authManager)

            viewModel.email = ""
            viewModel.password = "Pass0123"

            #expect(viewModel.isValid == false)
        }

        @Test func 空のpasswordでfalseを返すべき() {
            let authManager = AuthManager(repository: AuthRepositoryProtocolMock())
            let viewModel = LoginViewModel(authManager: authManager)

            viewModel.email = "test@example.com"
            viewModel.password = ""

            #expect(viewModel.isValid == false)
        }

        @Test func 8文字未満のpasswordでfalseを返すべき() {
            let authManager = AuthManager(repository: AuthRepositoryProtocolMock())
            let viewModel = LoginViewModel(authManager: authManager)

            viewModel.email = "test@example.com"
            viewModel.password = "Pass012"  // 7文字

            #expect(viewModel.isValid == false)
        }

        @Test func 8文字のpasswordでtrueを返すべき() {
            let authManager = AuthManager(repository: AuthRepositoryProtocolMock())
            let viewModel = LoginViewModel(authManager: authManager)

            viewModel.email = "test@example.com"
            viewModel.password = "Pass0123"  // 8文字

            #expect(viewModel.isValid == true)
        }
    }

    // MARK: - login Tests

    @Suite("login")
    struct LoginTests {

        @Test func 無効な入力でエラーメッセージが設定されるべき() async {
            let authManager = AuthManager(repository: AuthRepositoryProtocolMock())
            let viewModel = LoginViewModel(authManager: authManager)

            viewModel.email = ""
            viewModel.password = ""

            await viewModel.login()

            #expect(viewModel.errorMessage == "メールアドレスとパスワード（8文字以上）を入力してください")
        }

        @Test func ログイン成功時にisLoadingがfalseになるべき() async throws {
            let mockRepo = AuthRepositoryProtocolMock()
            mockRepo.loginHandler = { _, _ in
                AuthResponse(
                    token: "test-token",
                    user: User(id: "1", email: "test@example.com", name: "Test User")
                )
            }

            let authManager = AuthManager(repository: mockRepo)
            let viewModel = LoginViewModel(authManager: authManager)

            viewModel.email = "test@example.com"
            viewModel.password = "Pass0123"

            await viewModel.login()

            #expect(viewModel.isLoading == false)
        }

        @Test func ログイン成功時にエラーメッセージがnilであるべき() async throws {
            let mockRepo = AuthRepositoryProtocolMock()
            mockRepo.loginHandler = { _, _ in
                AuthResponse(
                    token: "test-token",
                    user: User(id: "1", email: "test@example.com", name: "Test User")
                )
            }

            let authManager = AuthManager(repository: mockRepo)
            let viewModel = LoginViewModel(authManager: authManager)

            viewModel.email = "test@example.com"
            viewModel.password = "Pass0123"

            await viewModel.login()

            #expect(viewModel.errorMessage == nil)
        }

        @Test func APIエラー時にエラーメッセージが設定されるべき() async {
            let mockRepo = AuthRepositoryProtocolMock()
            mockRepo.loginHandler = { _, _ in
                throw APIError.unauthorized
            }

            let authManager = AuthManager(repository: mockRepo)
            let viewModel = LoginViewModel(authManager: authManager)

            viewModel.email = "test@example.com"
            viewModel.password = "Pass0123"

            await viewModel.login()

            #expect(viewModel.errorMessage != nil)
            #expect(viewModel.isLoading == false)
        }

        @Test func 一般エラー時に汎用エラーメッセージが設定されるべき() async {
            struct TestError: Error {}

            let mockRepo = AuthRepositoryProtocolMock()
            mockRepo.loginHandler = { _, _ in
                throw TestError()
            }

            let authManager = AuthManager(repository: mockRepo)
            let viewModel = LoginViewModel(authManager: authManager)

            viewModel.email = "test@example.com"
            viewModel.password = "Pass0123"

            await viewModel.login()

            #expect(viewModel.errorMessage == "ログインに失敗しました")
            #expect(viewModel.isLoading == false)
        }
    }
}
