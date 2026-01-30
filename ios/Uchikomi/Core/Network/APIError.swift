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
        case .networkError(let error):
            return "ネットワークエラー: \(error.localizedDescription)"
        case .invalidResponse:
            return "無効なレスポンスです"
        case .httpError(let statusCode, let message):
            if let message = message {
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
        case .serverError(let message):
            return "サーバーエラー: \(message)"
        }
    }
}
