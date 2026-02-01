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
            // ローカル開発時（シミュレータ: localhost、実機: Mac のIPアドレス）
            #if targetEnvironment(simulator)
            return URL(string: "http://localhost:8080")!
            #else
            // 実機テスト時は Mac の IP アドレスに変更
            return URL(string: "http://192.168.1.100:8080")!
            #endif
        case .production:
            return URL(string: "https://utikomi.exe.dev")!
        }
    }

    // swiftlint:enable force_unwrapping

    var apiBaseURL: URL {
        baseURL.appendingPathComponent("api")
    }
}
