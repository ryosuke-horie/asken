import Foundation

private let iso8601Formatter = ISO8601DateFormatter()

extension KeyedDecodingContainer {
    func decodeISO8601Date(forKey key: Key) throws -> Date {
        let str = try decode(String.self, forKey: key)
        guard let date = iso8601Formatter.date(from: str) else {
            throw DecodingError.dataCorruptedError(
                forKey: key,
                in: self,
                debugDescription: "Invalid ISO8601 date: \(str)"
            )
        }
        return date
    }
}
