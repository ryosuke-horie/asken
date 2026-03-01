import Foundation

// MARK: - WidgetAPIError

enum WidgetAPIError: Error {
    case noToken
    case noBaseURL
    case invalidURL
    case invalidResponse
    case unauthorized
    case httpError(statusCode: Int)
    case decodingError(Error)
    case analysisFailed(String)
    case analysisTimeout
}

// MARK: - WidgetAPIClient

/// ウィジェット拡張用の軽量 HTTP クライアント（Firebase SDK 非依存）
/// App Groups UserDefaults (SharedDefaults) から認証トークンと API ベース URL を読み取る
struct WidgetAPIClient {
    private static let maxPollingAttempts = 10
    private static let pollingIntervalNanoseconds: UInt64 = 2_500_000_000 // 2.5秒

    private let session: URLSession
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder

    init() {
        let config = URLSessionConfiguration.default
        config.timeoutIntervalForRequest = 30
        config.timeoutIntervalForResource = 60
        session = URLSession(configuration: config)

        decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase

        // エンコーダーは CodingKeys で明示的にキー名を管理するため convertToSnakeCase は不使用
        encoder = JSONEncoder()
    }

    // MARK: - Weight API

    func createWeightRecord(weightKg: Double, recordedAt: Date, note: String) async throws -> WidgetWeightRecord {
        struct Request: Encodable {
            let weightKg: Double
            let recordedAt: String
            let note: String

            enum CodingKeys: String, CodingKey {
                case weightKg = "weight_kg"
                case recordedAt = "recorded_at"
                case note
            }
        }

        let iso8601 = ISO8601DateFormatter()
        iso8601.formatOptions = [.withInternetDateTime]

        let body = Request(
            weightKg: weightKg,
            recordedAt: iso8601.string(from: recordedAt),
            note: note
        )
        return try await post(path: "weight/records", body: body)
    }

    func getLatestWeightRecord() async throws -> WidgetWeightRecordsResponse {
        let today = localDateString(from: Date())
        let thirtyDaysAgo = localDateString(from: Date().addingTimeInterval(-30 * 24 * 60 * 60))
        return try await get(path: "weight/records?from=\(thirtyDaysAgo)&to=\(today)&tz=\(percentEncodedTimezone)")
    }

    // MARK: - Meal API

    func getDailyMeals(date: Date) async throws -> WidgetDailyMeals {
        try await get(path: "meals/daily?date=\(localDateString(from: date))&tz=\(percentEncodedTimezone)")
    }

    func analyzeText(inputText: String, mealType: String, mealDate: Date) async throws -> String {
        struct Request: Encodable {
            let inputText: String
            let mealType: String
            let mealDate: String
            let tz: String

            enum CodingKeys: String, CodingKey {
                case inputText = "input_text"
                case mealType = "meal_type"
                case mealDate = "meal_date"
                case tz
            }
        }

        let body = Request(
            inputText: inputText,
            mealType: mealType,
            mealDate: localDateString(from: mealDate),
            tz: TimeZone.current.identifier
        )
        let response: WidgetAnalyzeResponse = try await post(path: "analyze", body: body)
        return response.id
    }

    func checkAnalysisStatus(id: String) async throws -> WidgetAnalysisStatus {
        let safeID = sanitizePathID(id)
        return try await get(path: "analyze/\(safeID)/status")
    }

    /// 分析が完了するまでポーリングする（最大 maxPollingAttempts 回 x 2.5秒 = 25秒）
    func waitForAnalysisCompletion(id: String) async throws {
        for attempt in 0 ..< Self.maxPollingAttempts {
            try await Task.sleep(nanoseconds: Self.pollingIntervalNanoseconds)
            let status = try await checkAnalysisStatus(id: id)
            switch status.status {
            case .completed:
                return
            case let .failed(reason):
                throw WidgetAPIError.analysisFailed(reason)
            case .processing, .unknown:
                if attempt == Self.maxPollingAttempts - 1 {
                    throw WidgetAPIError.analysisTimeout
                }
            }
        }
    }

    // MARK: - Private HTTP Helpers

    private func get<T: Decodable>(path: String) async throws -> T {
        guard let token = SharedDefaults.authToken else { throw WidgetAPIError.noToken }
        guard let baseURL = SharedDefaults.apiBaseURL else { throw WidgetAPIError.noBaseURL }

        let urlString = baseURL.hasSuffix("/") ? baseURL + path : baseURL + "/" + path
        guard let url = URL(string: urlString) else { throw WidgetAPIError.invalidURL }

        var request = URLRequest(url: url)
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        return try await perform(request)
    }

    private func post<Response: Decodable>(path: String, body: some Encodable) async throws -> Response {
        guard let token = SharedDefaults.authToken else { throw WidgetAPIError.noToken }
        guard let baseURL = SharedDefaults.apiBaseURL else { throw WidgetAPIError.noBaseURL }

        let urlString = baseURL.hasSuffix("/") ? baseURL + path : baseURL + "/" + path
        guard let url = URL(string: urlString) else { throw WidgetAPIError.invalidURL }

        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try encoder.encode(body)
        return try await perform(request)
    }

    private func perform<T: Decodable>(_ request: URLRequest) async throws -> T {
        let (data, response) = try await session.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw WidgetAPIError.invalidResponse
        }

        switch httpResponse.statusCode {
        case 200 ..< 300:
            do {
                return try decoder.decode(T.self, from: data)
            } catch {
                throw WidgetAPIError.decodingError(error)
            }
        case 401:
            throw WidgetAPIError.unauthorized
        default:
            throw WidgetAPIError.httpError(statusCode: httpResponse.statusCode)
        }
    }

    private func localDateString(from date: Date) -> String {
        let f = DateFormatter()
        f.dateFormat = "yyyy-MM-dd"
        f.timeZone = .current
        return f.string(from: date)
    }

    private var percentEncodedTimezone: String {
        var allowed = CharacterSet.urlQueryAllowed
        allowed.remove(charactersIn: "/")
        return TimeZone.current.identifier
            .addingPercentEncoding(withAllowedCharacters: allowed) ?? TimeZone.current.identifier
    }

    private func sanitizePathID(_ id: String) -> String {
        let allowed = CharacterSet.alphanumerics.union(CharacterSet(charactersIn: "-"))
        return id.unicodeScalars.filter { allowed.contains($0) }.map { String($0) }.joined()
    }
}
