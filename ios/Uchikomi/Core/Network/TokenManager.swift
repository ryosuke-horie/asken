import Foundation
import Security

// MARK: - TokenManager

actor TokenManager {
    static let shared = TokenManager()

    private let serviceName = "dev.exe.utikomi"
    private let accountName = "authToken"

    private init() {}

    func saveToken(_ token: String) throws {
        let tokenData = Data(token.utf8)

        // 既存のトークンを削除
        let deleteQuery: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: serviceName,
            kSecAttrAccount as String: accountName,
        ]
        SecItemDelete(deleteQuery as CFDictionary)

        // 新しいトークンを保存
        let addQuery: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: serviceName,
            kSecAttrAccount as String: accountName,
            kSecValueData as String: tokenData,
            kSecAttrAccessible as String: kSecAttrAccessibleAfterFirstUnlock,
        ]

        let status = SecItemAdd(addQuery as CFDictionary, nil)
        guard status == errSecSuccess else {
            throw TokenError.saveFailed(status)
        }
    }

    func getToken() -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: serviceName,
            kSecAttrAccount as String: accountName,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne,
        ]

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)

        guard status == errSecSuccess,
              let data = result as? Data,
              let token = String(data: data, encoding: .utf8) else {
            return nil
        }

        return token
    }

    func deleteToken() {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: serviceName,
            kSecAttrAccount as String: accountName,
        ]
        SecItemDelete(query as CFDictionary)
    }
}

// MARK: - TokenError

enum TokenError: LocalizedError {
    case saveFailed(OSStatus)
    case notFound

    var errorDescription: String? {
        switch self {
        case let .saveFailed(status):
            "トークンの保存に失敗しました (status: \(status))"
        case .notFound:
            "トークンが見つかりません"
        }
    }
}
