import AppIntents
import WidgetKit

// MARK: - WidgetMealType

enum WidgetMealType: String, AppEnum {
    case breakfast
    case lunch
    case dinner
    case snack

    static var typeDisplayRepresentation = TypeDisplayRepresentation(name: "食事タイプ")
    static var caseDisplayRepresentations: [WidgetMealType: DisplayRepresentation] = [
        .breakfast: DisplayRepresentation(title: "朝食"),
        .lunch: DisplayRepresentation(title: "昼食"),
        .dinner: DisplayRepresentation(title: "夕食"),
        .snack: DisplayRepresentation(title: "間食"),
    ]

    var displayName: String {
        switch self {
        case .breakfast: "朝食"
        case .lunch: "昼食"
        case .dinner: "夕食"
        case .snack: "間食"
        }
    }

    var icon: String {
        switch self {
        case .breakfast: "sunrise"
        case .lunch: "sun.max"
        case .dinner: "moon"
        case .snack: "cup.and.saucer"
        }
    }
}

// MARK: - MealWidgetConfiguration

struct MealWidgetConfiguration: WidgetConfigurationIntent {
    static var title: LocalizedStringResource = "食事ウィジェット設定"
    static var description = IntentDescription("タップ一つで食事を Gemini で分析・記録できます")

    @Parameter(title: "食事タイプ", default: .breakfast)
    var mealType: WidgetMealType

    @Parameter(title: "食事内容", description: "よく食べる食事を入力してください（例: ご飯、味噌汁、焼き鮭）", default: "")
    var foodDescription: String
}

// MARK: - RecordMealIntent

/// 事前設定された食事内容を Gemini で分析・記録する App Intent
struct RecordMealIntent: AppIntent {
    static var title: LocalizedStringResource = "食事を記録"
    static var description = IntentDescription("設定した食事内容を Gemini で分析して記録します")
    static var openAppWhenRun: Bool = false

    @Parameter(title: "食事タイプ")
    var mealType: WidgetMealType

    @Parameter(title: "食事内容")
    var foodDescription: String

    init() {
        mealType = .breakfast
        foodDescription = ""
    }

    init(mealType: WidgetMealType, foodDescription: String) {
        self.mealType = mealType
        self.foodDescription = foodDescription
    }

    func perform() async throws -> some IntentResult {
        guard !foodDescription.isEmpty else {
            return .result()
        }

        let client = WidgetAPIClient()
        let analysisID = try await client.analyzeText(
            inputText: foodDescription,
            mealType: mealType.rawValue,
            mealDate: Date()
        )

        try await client.waitForAnalysisCompletion(id: analysisID)
        WidgetCenter.shared.reloadTimelines(ofKind: "MealWidget")
        return .result()
    }
}
