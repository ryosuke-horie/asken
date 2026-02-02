import Foundation

// MARK: - User

public struct User: Codable, Equatable {
    public let id: String
    public let email: String
    public let name: String?

    public init(id: String, email: String, name: String?) {
        self.id = id
        self.email = email
        self.name = name
    }
}
