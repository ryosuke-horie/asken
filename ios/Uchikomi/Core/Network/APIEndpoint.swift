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
            preconditionFailure("Invalid URL configuration for path: \(path)")
        }
        return url
    }

    /// URLパスに埋め込むIDをサニタイズする（英数字とハイフンのみ許可）
    private static func sanitizedPathID(_ id: String) -> String {
        let allowed = CharacterSet.alphanumerics.union(CharacterSet(charactersIn: "-"))
        return id.unicodeScalars.filter { allowed.contains($0) }.map { String($0) }.joined()
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

    static let skipMeal = APIEndpoint(
        path: "meals/skip",
        method: .post,
        requiresAuth: true
    )

    static let analyze = APIEndpoint(
        path: "analyze",
        method: .post,
        requiresAuth: true
    )

    static func analysisStatus(id: String) -> APIEndpoint {
        APIEndpoint(
            path: "analyze/\(Self.sanitizedPathID(id))/status",
            method: .get,
            requiresAuth: true
        )
    }

    static func analysisResult(id: String) -> APIEndpoint {
        APIEndpoint(
            path: "analyze/\(Self.sanitizedPathID(id))/result",
            method: .get,
            requiresAuth: true
        )
    }

    // MARK: - History Endpoints

    static func historyDetail(id: String) -> APIEndpoint {
        APIEndpoint(
            path: "history/\(Self.sanitizedPathID(id))",
            method: .get,
            requiresAuth: true
        )
    }

    static func updateHistory(id: String) -> APIEndpoint {
        APIEndpoint(
            path: "history/\(Self.sanitizedPathID(id))",
            method: .put,
            requiresAuth: true
        )
    }

    static func deleteHistory(id: String) -> APIEndpoint {
        APIEndpoint(
            path: "history/\(Self.sanitizedPathID(id))",
            method: .delete,
            requiresAuth: true
        )
    }

    // MARK: - Weight Endpoints

    static func weightRecords(from: String, to: String, timezone: String) -> APIEndpoint {
        var allowedCharacters = CharacterSet.urlQueryAllowed
        allowedCharacters.remove(charactersIn: "/")
        let encodedTimezone = timezone.addingPercentEncoding(withAllowedCharacters: allowedCharacters) ?? timezone
        return APIEndpoint(
            path: "weight/records?from=\(from)&to=\(to)&tz=\(encodedTimezone)",
            method: .get,
            requiresAuth: true
        )
    }

    static let createWeightRecord = APIEndpoint(
        path: "weight/records",
        method: .post,
        requiresAuth: true
    )

    static func updateWeightRecord(id: String) -> APIEndpoint {
        APIEndpoint(
            path: "weight/records/\(Self.sanitizedPathID(id))",
            method: .put,
            requiresAuth: true
        )
    }

    static func deleteWeightRecord(id: String) -> APIEndpoint {
        APIEndpoint(
            path: "weight/records/\(Self.sanitizedPathID(id))",
            method: .delete,
            requiresAuth: true
        )
    }

    static let getWeightGoal = APIEndpoint(
        path: "weight/goal",
        method: .get,
        requiresAuth: true
    )

    static let setWeightGoal = APIEndpoint(
        path: "weight/goal",
        method: .put,
        requiresAuth: true
    )

    // MARK: - Nutrition Goal Endpoints

    static func getNutritionGoal(currentWeight: Double?, goalWeight: Double?) -> APIEndpoint {
        var components = URLComponents()
        components.path = "nutrition/goal"

        var queryItems: [URLQueryItem] = []
        if let current = currentWeight {
            queryItems.append(URLQueryItem(name: "current_weight", value: String(format: "%.1f", current)))
        }
        if let goal = goalWeight {
            queryItems.append(URLQueryItem(name: "target_weight", value: String(format: "%.1f", goal)))
        }

        if !queryItems.isEmpty {
            components.queryItems = queryItems
        }

        let path = if let queryString = components.percentEncodedQuery {
            "nutrition/goal?\(queryString)"
        } else {
            "nutrition/goal"
        }

        return APIEndpoint(
            path: path,
            method: .get,
            requiresAuth: true
        )
    }

    static let setNutritionGoal = APIEndpoint(
        path: "nutrition/goal",
        method: .put,
        requiresAuth: true
    )

    // MARK: - MyMenu Endpoints

    static let myMenuList = APIEndpoint(
        path: "my-menu",
        method: .get,
        requiresAuth: true
    )

    static let createMyMenu = APIEndpoint(
        path: "my-menu",
        method: .post,
        requiresAuth: true
    )

    static func updateMyMenu(id: String) -> APIEndpoint {
        APIEndpoint(
            path: "my-menu/\(Self.sanitizedPathID(id))",
            method: .put,
            requiresAuth: true
        )
    }

    static func deleteMyMenu(id: String) -> APIEndpoint {
        APIEndpoint(
            path: "my-menu/\(Self.sanitizedPathID(id))",
            method: .delete,
            requiresAuth: true
        )
    }

    static func recordMyMenu(id: String) -> APIEndpoint {
        APIEndpoint(
            path: "my-menu/\(Self.sanitizedPathID(id))/record",
            method: .post,
            requiresAuth: true
        )
    }
}
