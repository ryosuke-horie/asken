import Foundation

enum APIError: LocalizedError {
    case invalidURL
    case networkError(Error)
    case invalidResponse
    case httpError(statusCode: Int, message: String?)
    case decodingError(Error)
    case encodingError(Error)
    case unauthorized
    case notFound
    case serverError(String)

    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "無効なURLです"
        case let .networkError(error):
            return "ネットワークエラー: \(error.localizedDescription)"
        case .invalidResponse:
            return "無効なレスポンスです"
        case let .httpError(statusCode, message):
            if let message {
                return message
            }
            return "HTTPエラー: \(statusCode)"
        case .decodingError:
            return "レスポンスの解析に失敗しました"
        case .encodingError:
            return "リクエストの作成に失敗しました"
        case .unauthorized:
            return "認証が必要です"
        case .notFound:
            return "データが見つかりません"
        case let .serverError(message):
            return "サーバーエラー: \(message)"
        }
    }
}
