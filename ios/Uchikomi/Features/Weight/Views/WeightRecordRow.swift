import SwiftUI

struct WeightRecordRow: View {
    let record: WeightRecord

    private static let iso8601Formatter: ISO8601DateFormatter = {
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return f
    }()

    private static let iso8601FallbackFormatter: ISO8601DateFormatter = {
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime]
        return f
    }()

    private static let timeFormatter: DateFormatter = {
        let f = DateFormatter()
        f.dateFormat = "HH:mm"
        f.timeZone = TimeZone.current
        return f
    }()

    private var formattedTime: String {
        guard let date = Self.iso8601Formatter.date(from: record.recordedAt)
            ?? Self.iso8601FallbackFormatter.date(from: record.recordedAt) else {
            return "--:--"
        }
        return Self.timeFormatter.string(from: date)
    }

    var body: some View {
        HStack {
            VStack(alignment: .leading, spacing: 4) {
                Text(formattedTime)
                    .font(.subheadline)
                    .foregroundStyle(.secondary)

                if let note = record.note, !note.isEmpty {
                    Text(note)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }

            Spacer()

            Text(String(format: "%.1f", record.weightKg))
                .font(.title2)
                .fontWeight(.semibold)
                + Text(" kg")
                .font(.subheadline)
                .foregroundColor(.secondary)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 10)
        .background(Color(.secondarySystemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 8))
    }
}
