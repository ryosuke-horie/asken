import Foundation

// MARK: - User

struct User: Codable, Equatable {
    let id: String
    let email: String
    let name: String?
}

// MARK: - LoginRequest

struct LoginRequest: Encodable {
    let email: String
    let password: String
}

// MARK: - RegisterRequest

struct RegisterRequest: Encodable {
    let email: String
    let password: String
    let name: String
}

// MARK: - AuthResponse

struct AuthResponse: Decodable {
    let token: String
    let user: User
}
