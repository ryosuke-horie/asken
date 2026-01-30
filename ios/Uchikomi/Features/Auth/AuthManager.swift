import Foundation

@Observable
final class AuthManager {
    var currentUser: User?
    var isAuthenticated: Bool {
        currentUser != nil
    }

    private let repository: AuthRepositoryProtocol

    init(repository: AuthRepositoryProtocol = AuthRepository()) {
        self.repository = repository
        Task {
            await checkExistingToken()
        }
    }

    func login(email: String, password: String) async throws {
        let response = try await repository.login(email: email, password: password)
        try await TokenManager.shared.saveToken(response.token)
        await MainActor.run {
            self.currentUser = response.user
        }
    }

    func logout() async {
        await TokenManager.shared.deleteToken()
        await MainActor.run {
            self.currentUser = nil
        }
    }

    private func checkExistingToken() async {
        guard let token = await TokenManager.shared.getToken() else {
            return
        }

        // トークンが存在する場合、ユーザー情報をデコード
        // JWT のペイロードからユーザー情報を取得
        if let user = decodeUserFromToken(token) {
            await MainActor.run {
                self.currentUser = user
            }
        }
    }

    private func decodeUserFromToken(_ token: String) -> User? {
        let parts = token.split(separator: ".")
        guard parts.count == 3 else { return nil }

        let payload = String(parts[1])

        // Base64 URL デコード
        var base64 = payload
            .replacingOccurrences(of: "-", with: "+")
            .replacingOccurrences(of: "_", with: "/")

        // パディング追加
        let remainder = base64.count % 4
        if remainder > 0 {
            base64 += String(repeating: "=", count: 4 - remainder)
        }

        guard let data = Data(base64Encoded: base64) else { return nil }

        struct JWTPayload: Decodable {
            let userId: String
            let email: String

            enum CodingKeys: String, CodingKey {
                case userId = "user_id"
                case email
            }
        }

        guard let payload = try? JSONDecoder().decode(JWTPayload.self, from: data) else {
            return nil
        }

        return User(id: payload.userId, email: payload.email, name: nil)
    }
}
