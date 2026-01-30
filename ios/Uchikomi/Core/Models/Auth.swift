import Foundation

struct User: Codable, Equatable {
    let id: String
    let email: String
    let name: String?
}

struct LoginRequest: Encodable {
    let email: String
    let password: String
}

struct RegisterRequest: Encodable {
    let email: String
    let password: String
    let name: String
}

struct AuthResponse: Decodable {
    let token: String
    let user: User
}
