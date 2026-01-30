import Foundation

actor APIClient {
    static let shared = APIClient()

    private let session: URLSession
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder

    private init() {
        let config = URLSessionConfiguration.default
        config.timeoutIntervalForRequest = 30
        config.timeoutIntervalForResource = 60
        self.session = URLSession(configuration: config)

        self.decoder = JSONDecoder()
        self.decoder.keyDecodingStrategy = .convertFromSnakeCase

        self.encoder = JSONEncoder()
        self.encoder.keyEncodingStrategy = .convertToSnakeCase
    }

    // MARK: - JSON Request

    func request<T: Decodable>(
        endpoint: APIEndpoint,
        body: (any Encodable)? = nil
    ) async throws -> T {
        var request = try await createRequest(endpoint: endpoint)

        if let body = body {
            let bodyData = try encoder.encode(body)
            request.httpBody = bodyData
            request.setValue("application/json", forHTTPHeaderField: "Content-Type")
            #if DEBUG
            if let jsonString = String(data: bodyData, encoding: .utf8) {
                print("[APIClient] Request URL: \(endpoint.url)")
                print("[APIClient] Request Body: \(jsonString)")
            }
            #endif
        }

        return try await performRequest(request)
    }

    // MARK: - Multipart Request (for image upload)

    func uploadImage(
        endpoint: APIEndpoint,
        imageData: Data,
        filename: String,
        additionalFields: [String: String] = [:]
    ) async throws -> AnalyzeResponse {
        var request = try await createRequest(endpoint: endpoint)

        let boundary = UUID().uuidString
        request.setValue("multipart/form-data; boundary=\(boundary)", forHTTPHeaderField: "Content-Type")

        var body = Data()

        // Add additional fields
        for (key, value) in additionalFields {
            body.append("--\(boundary)\r\n".data(using: .utf8)!)
            body.append("Content-Disposition: form-data; name=\"\(key)\"\r\n\r\n".data(using: .utf8)!)
            body.append("\(value)\r\n".data(using: .utf8)!)
        }

        // Add image data
        body.append("--\(boundary)\r\n".data(using: .utf8)!)
        body.append("Content-Disposition: form-data; name=\"image\"; filename=\"\(filename)\"\r\n".data(using: .utf8)!)
        body.append("Content-Type: image/jpeg\r\n\r\n".data(using: .utf8)!)
        body.append(imageData)
        body.append("\r\n".data(using: .utf8)!)
        body.append("--\(boundary)--\r\n".data(using: .utf8)!)

        request.httpBody = body

        return try await performRequest(request)
    }

    // MARK: - Private Helpers

    private func createRequest(endpoint: APIEndpoint) async throws -> URLRequest {
        var request = URLRequest(url: endpoint.url)
        request.httpMethod = endpoint.method.rawValue

        if endpoint.requiresAuth {
            guard let token = await TokenManager.shared.getToken() else {
                throw APIError.unauthorized
            }
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        return request
    }

    private func performRequest<T: Decodable>(_ request: URLRequest) async throws -> T {
        let (data, response) = try await session.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }

        #if DEBUG
        print("[APIClient] Response Status: \(httpResponse.statusCode)")
        if let responseString = String(data: data, encoding: .utf8) {
            print("[APIClient] Response Body: \(responseString)")
        }
        #endif

        switch httpResponse.statusCode {
        case 200...299:
            do {
                return try decoder.decode(T.self, from: data)
            } catch {
                #if DEBUG
                print("[APIClient] Decoding Error: \(error)")
                if let jsonString = String(data: data, encoding: .utf8) {
                    print("[APIClient] Raw Response: \(jsonString)")
                }
                #endif
                throw APIError.decodingError(error)
            }

        case 401:
            throw APIError.unauthorized

        case 404:
            throw APIError.notFound

        default:
            let message = String(data: data, encoding: .utf8)
            throw APIError.httpError(statusCode: httpResponse.statusCode, message: message)
        }
    }
}

// MARK: - Analyze Response (used for multipart upload)

struct AnalyzeResponse: Decodable {
    let id: String
}
