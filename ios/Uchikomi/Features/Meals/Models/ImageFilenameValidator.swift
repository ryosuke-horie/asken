import Foundation

enum ImageFilenameValidator {
    private static let allowedExtensions: Set<String> = ["jpg", "jpeg", "png", "heic"]

    private static let filenamePattern = #/^[a-zA-Z0-9._-]+$/#

    static func isValid(_ filename: String) -> Bool {
        guard !filename.isEmpty else { return false }
        guard !filename.hasPrefix(".") else { return false }
        guard !filename.contains("..") else { return false }
        guard filename.wholeMatch(of: filenamePattern) != nil else { return false }

        let ext = (filename as NSString).pathExtension.lowercased()
        guard allowedExtensions.contains(ext) else { return false }

        return true
    }
}
