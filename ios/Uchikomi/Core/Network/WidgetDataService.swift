import Foundation
import WidgetKit

actor WidgetDataService {
    static let shared = WidgetDataService()

    private let appGroupId = "group.dev.exe.utikomi"
    private let dataKey = "widgetData"

    private var userDefaults: UserDefaults? {
        UserDefaults(suiteName: appGroupId)
    }

    func updateWeight(current: Double?, target: Double?) {
        var data = getData()
        data = WidgetDataForApp(
            currentWeight: current ?? data.currentWeight,
            targetWeight: target ?? data.targetWeight,
            todayCalories: data.todayCalories,
            lastUpdated: Date()
        )
        saveData(data)
        reloadWidgets()
    }

    func updateCalories(_ calories: Double?) {
        var data = getData()
        data = WidgetDataForApp(
            currentWeight: data.currentWeight,
            targetWeight: data.targetWeight,
            todayCalories: calories,
            lastUpdated: Date()
        )
        saveData(data)
        reloadWidgets()
    }

    private func getData() -> WidgetDataForApp {
        guard let defaults = userDefaults,
              let data = defaults.data(forKey: dataKey),
              let decoded = try? JSONDecoder().decode(WidgetDataForApp.self, from: data) else {
            return WidgetDataForApp(
                currentWeight: nil,
                targetWeight: nil,
                todayCalories: nil,
                lastUpdated: Date()
            )
        }
        return decoded
    }

    private func saveData(_ data: WidgetDataForApp) {
        guard let encoded = try? JSONEncoder().encode(data) else { return }
        userDefaults?.set(encoded, forKey: dataKey)
    }

    private func reloadWidgets() {
        WidgetCenter.shared.reloadAllTimelines()
    }
}

private struct WidgetDataForApp: Codable {
    let currentWeight: Double?
    let targetWeight: Double?
    let todayCalories: Double?
    let lastUpdated: Date
}
