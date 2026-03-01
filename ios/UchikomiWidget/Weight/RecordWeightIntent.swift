import AppIntents
import OSLog
import WidgetKit

private let logger = Logger(subsystem: "dev.exe.uchikomi.widget", category: "RecordWeightIntent")

// MARK: - RecordWeightIntent

/// 体重を記録する App Intent
/// 最後に記録した体重に delta を加算して API に送信する
struct RecordWeightIntent: AppIntent {
    static var title: LocalizedStringResource = "体重を記録"
    static var description = IntentDescription("ウィジェットから体重を記録します")
    static var openAppWhenRun: Bool = false

    @Parameter(title: "調整量(kg)")
    var delta: Double

    @Parameter(title: "タイミング")
    var timingNote: String

    init() {
        delta = 0
        timingNote = ""
    }

    init(delta: Double, timingNote: String = "") {
        self.delta = delta
        self.timingNote = timingNote
    }

    func perform() async throws -> some IntentResult {
        guard let lastWeight = SharedDefaults.latestWeightKg, lastWeight > 0 else {
            logger
                .warning(
                    "RecordWeightIntent: キャッシュ体重が存在しない（アプリを開いて体重を記録してください）。latestWeightKg=\(String(describing: SharedDefaults.latestWeightKg))"
                )
            return .result()
        }

        let newWeight = ((lastWeight + delta) * 10).rounded() / 10

        let client = WidgetAPIClient()
        let record = try await client.createWeightRecord(
            weightKg: newWeight,
            recordedAt: Date(),
            note: timingNote
        )

        SharedDefaults.latestWeightKg = record.weightKg
        WidgetCenter.shared.reloadTimelines(ofKind: "WeightWidget")
        return .result()
    }
}
