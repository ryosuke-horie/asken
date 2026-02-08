import SwiftUI

struct WeightRecordRow: View {
    let record: WeightRecord

    private static let timeFormatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.dateFormat = "HH:mm"
        formatter.timeZone = TimeZone.current
        return formatter
    }()

    private var formattedTime: String {
        guard let date = WeightRecord.parseISO8601(record.recordedAt) else {
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
