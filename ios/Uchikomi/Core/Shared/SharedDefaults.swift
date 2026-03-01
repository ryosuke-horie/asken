import Foundation

// MARK: - SharedDefaults

/// App Groups を使用してメインアプリとウィジェット間でデータを共有する
///
/// セキュリティ注記: Firebase ID Token は UserDefaults（App Groups）に保存される。
/// UserDefaults はプロパティリスト形式で保存されるため、脱獄デバイスでは平文参照が可能。
/// Firebase ID Token の有効期間は1時間と短命であり、クリティカルな操作には
/// Firebase 認証による二重検証が必須のため、現時点でのリスクは許容範囲と判断する。
/// セキュリティ要件が高まった場合は Keychain（kSecAttrAccessGroup）への移行を検討すること。
enum SharedDefaults {
    static let appGroupID = "group.dev.exe.uchikomi"

    private static var defaults: UserDefaults {
        UserDefaults(suiteName: appGroupID) ?? .standard
    }

    // MARK: - Keys

    enum Keys {
        static let authToken = "authToken"
        static let apiBaseURL = "apiBaseURL"
        static let latestWeightKg = "latestWeightKg"
    }

    // MARK: - Auth Token

    static var authToken: String? {
        get { defaults.string(forKey: Keys.authToken) }
        set { defaults.set(newValue, forKey: Keys.authToken) }
    }

    // MARK: - API Base URL

    static var apiBaseURL: String? {
        get { defaults.string(forKey: Keys.apiBaseURL) }
        set { defaults.set(newValue, forKey: Keys.apiBaseURL) }
    }

    // MARK: - Latest Weight (ウィジェット用キャッシュ)

    static var latestWeightKg: Double? {
        get { defaults.object(forKey: Keys.latestWeightKg) as? Double }
        set { defaults.set(newValue, forKey: Keys.latestWeightKg) }
    }

    // MARK: - Helpers

    static func clearAuthToken() {
        defaults.removeObject(forKey: Keys.authToken)
    }
}
