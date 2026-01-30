import Foundation

protocol AuthRepositoryProtocol {
    func login(email: String, password: String) async throws -> AuthResponse
    func register(email: String, password: String, name: String) async throws -> AuthResponse
}

final class AuthRepository: AuthRepositoryProtocol {
    private let apiClient = APIClient.shared

    func login(email: String, password: String) async throws -> AuthResponse {
        let request = LoginRequest(email: email, password: password)
        return try await apiClient.request(endpoint: .login, body: request)
    }

    func register(email: String, password: String, name: String) async throws -> AuthResponse {
        let request = RegisterRequest(email: email, password: password, name: name)
        return try await apiClient.request(endpoint: .register, body: request)
    }
}
