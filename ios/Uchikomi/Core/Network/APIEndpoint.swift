import Foundation

enum HTTPMethod: String {
    case get = "GET"
    case post = "POST"
    case put = "PUT"
    case delete = "DELETE"
}

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

    // MARK: - Auth Endpoints

    static let login = APIEndpoint(
        path: "auth/login",
        method: .post,
        requiresAuth: false
    )

    static let register = APIEndpoint(
        path: "auth/register",
        method: .post,
        requiresAuth: false
    )

    // MARK: - Meals Endpoints

    static func dailyMeals(date: String) -> APIEndpoint {
        APIEndpoint(
            path: "meals/daily?date=\(date)",
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

    // MARK: - Weight Endpoints

    static func weightRecords(period: String) -> APIEndpoint {
        APIEndpoint(
            path: "weight-records?period=\(period)",
            method: .get,
            requiresAuth: true
        )
    }

    static let createWeightRecord = APIEndpoint(
        path: "weight-records",
        method: .post,
        requiresAuth: true
    )

    static let weightGoal = APIEndpoint(
        path: "weight-goals/current",
        method: .get,
        requiresAuth: true
    )

    static let setWeightGoal = APIEndpoint(
        path: "weight-goals",
        method: .post,
        requiresAuth: true
    )
}
