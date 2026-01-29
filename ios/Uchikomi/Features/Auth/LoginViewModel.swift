import Foundation

@Observable
final class LoginViewModel {
    var email = ""
    var password = ""
    var isLoading = false
    var errorMessage: String?

    var isValid: Bool {
        !email.isEmpty && !password.isEmpty && password.count >= 8
    }

    private let authManager: AuthManager

    init(authManager: AuthManager) {
        self.authManager = authManager
    }

    func login() async {
        guard isValid else {
            errorMessage = "メールアドレスとパスワード（8文字以上）を入力してください"
            return
        }

        isLoading = true
        errorMessage = nil

        do {
            try await authManager.login(email: email, password: password)
        } catch let error as APIError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "ログインに失敗しました"
        }

        isLoading = false
    }
}
