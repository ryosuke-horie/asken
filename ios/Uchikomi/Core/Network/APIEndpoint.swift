import Foundation

// MARK: - HTTPMethod

enum HTTPMethod: String {
    case get = "GET"
    case post = "POST"
    case put = "PUT"
    case delete = "DELETE"
}

// MARK: - APIEndpoint

struct APIEndpoint {
    let path: String
    let method: HTTPMethod
    let requiresAuth: Bool

    var url: URL {
        let baseURLString = AppEnvironment.current.apiBaseURL.absoluteString
        let urlString = baseURLString.hasSuffix("/") ? baseURLString + path : baseURLString + "/" + path
        guard let url = URL(string: urlString) else {
            fatalError("Invalid URL configuration: \(urlString)")
        }
        return url
    }

    // MARK: - Meals Endpoints

    static func dailyMeals(date: String, timezone: String) -> APIEndpoint {
        var allowedCharacters = CharacterSet.urlQueryAllowed
        allowedCharacters.remove(charactersIn: "/")
        let encodedTimezone = timezone.addingPercentEncoding(withAllowedCharacters: allowedCharacters) ?? timezone
        return APIEndpoint(
            path: "meals/daily?date=\(date)&tz=\(encodedTimezone)",
            method: .get,
            requiresAuth: true
        )
    }

    static let analyze = APIEndpoint(
        path: "analyze",
        method: .post,
        requiresAuth: true
    )

    static func analysisStatus(id: String) -> APIEndpoint {
        APIEndpoint(
            path: "analyze/\(id)/status",
            method: .get,
            requiresAuth: true
        )
    }

    static func analysisResult(id: String) -> APIEndpoint {
        APIEndpoint(
            path: "analyze/\(id)/result",
            method: .get,
            requiresAuth: true
        )
    }

    // MARK: - History Endpoints

    static func historyDetail(id: String) -> APIEndpoint {
        APIEndpoint(
            path: "history/\(id)",
            method: .get,
            requiresAuth: true
        )
    }

    static func updateHistory(id: String) -> APIEndpoint {
        APIEndpoint(
            path: "history/\(id)",
            method: .put,
            requiresAuth: true
        )
    }

    static func deleteHistory(id: String) -> APIEndpoint {
        APIEndpoint(
            path: "history/\(id)",
            method: .delete,
            requiresAuth: true
        )
    }
}
