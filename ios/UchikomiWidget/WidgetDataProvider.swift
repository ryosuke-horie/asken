import Foundation

struct WidgetData {
    let currentWeight: Double?
    let targetWeight: Double?
    let todayCalories: Double?
    let lastUpdated: Date

    static let placeholder = WidgetData(
        currentWeight: 70.0,
        targetWeight: 65.0,
        todayCalories: 1500,
        lastUpdated: Date()
    )

    static let empty = WidgetData(
        currentWeight: nil,
        targetWeight: nil,
        todayCalories: nil,
        lastUpdated: Date()
    )

    var weightDifference: Double? {
        guard let current = currentWeight, let target = targetWeight else { return nil }
        return current - target
    }
}

actor WidgetDataProvider {
    static let shared = WidgetDataProvider()

    private let appGroupId = "group.dev.exe.utikomi"
    private let dataKey = "widgetData"

    private var userDefaults: UserDefaults? {
        UserDefaults(suiteName: appGroupId)
    }

    func getData() -> WidgetData {
        guard let defaults = userDefaults,
              let data = defaults.data(forKey: dataKey),
              let decoded = try? JSONDecoder().decode(StoredWidgetData.self, from: data) else {
            return .empty
        }

        return WidgetData(
            currentWeight: decoded.currentWeight,
            targetWeight: decoded.targetWeight,
            todayCalories: decoded.todayCalories,
            lastUpdated: decoded.lastUpdated
        )
    }

    func saveData(_ data: WidgetData) {
        let stored = StoredWidgetData(
            currentWeight: data.currentWeight,
            targetWeight: data.targetWeight,
            todayCalories: data.todayCalories,
            lastUpdated: data.lastUpdated
        )

        guard let encoded = try? JSONEncoder().encode(stored) else { return }
        userDefaults?.set(encoded, forKey: dataKey)
    }
}

private struct StoredWidgetData: Codable {
    let currentWeight: Double?
    let targetWeight: Double?
    let todayCalories: Double?
    let lastUpdated: Date
}
