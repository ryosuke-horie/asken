import Foundation

enum AppEnvironment {
    case development
    case production

    static var current: AppEnvironment {
        #if DEBUG
        return .development
        #else
        return .production
        #endif
    }

    // 静的な文字列リテラルからのURL生成は必ず成功するため、強制アンラップを許可
    // swiftlint:disable force_unwrapping
    var baseURL: URL {
        switch self {
        case .development:
            // ローカル開発時（シミュレータ: localhost、実機: Cloud Run検証環境）
            #if targetEnvironment(simulator)
            return URL(string: "http://localhost:8080")!
            #else
            // 実機テスト時は Cloud Run の検証環境を使用
            return URL(string: "https://uchikomi-api-dev-425786510917.asia-northeast1.run.app")!
            #endif
        case .production:
            return URL(string: "https://utikomi.exe.dev")!
        }
    }

    // swiftlint:enable force_unwrapping

    var apiBaseURL: URL {
        baseURL.appendingPathComponent("api")
    }

    /// シミュレータ + DEBUG ビルドでモック認証を使用するかどうか
    var useMockAuth: Bool {
        #if DEBUG && targetEnvironment(simulator)
        return true
        #else
        return false
        #endif
    }
}
