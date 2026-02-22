import Foundation

extension KeyedDecodingContainer {
    func decodeISO8601Date(forKey key: Key) throws -> Date {
        let str = try decode(String.self, forKey: key)
        guard let date = ISO8601DateFormatter().date(from: str) else {
            throw DecodingError.dataCorruptedError(
                forKey: key,
                in: self,
                debugDescription: "Invalid ISO8601 date: \(str)"
            )
        }
        return date
    }
}
