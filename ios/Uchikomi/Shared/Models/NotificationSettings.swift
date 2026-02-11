import Foundation
import os

private let logger = Logger(
    subsystem: Bundle.main.bundleIdentifier ?? "Uchikomi",
    category: "NotificationSettingsStore"
)

// MARK: - TimedNotificationSetting

/// 時刻指定のある通知設定の共通プロトコル
protocol TimedNotificationSetting {
    var isEnabled: Bool { get }
    var hour: Int { get }
    var minute: Int { get }
}

extension TimedNotificationSetting {
    var timeComponents: DateComponents {
        var components = DateComponents()
        components.hour = hour
        components.minute = minute
        return components
    }
}

// MARK: - MealNotificationSetting

struct MealNotificationSetting: Codable, Equatable, TimedNotificationSetting {
    let mealType: MealType
    var isEnabled: Bool
    var hour: Int
    var minute: Int
}

// MARK: - WeightNotificationSetting

struct WeightNotificationSetting: Codable, Equatable, TimedNotificationSetting {
    var isEnabled: Bool
    var hour: Int
    var minute: Int

    static let `default` = WeightNotificationSetting(
        isEnabled: true,
        hour: 7,
        minute: 0
    )
}

// MARK: - NotificationSettings

struct NotificationSettings: Codable, Equatable {
    var isGlobalEnabled: Bool
    var meals: [MealNotificationSetting]
    var weight: WeightNotificationSetting

    static let `default` = NotificationSettings(
        isGlobalEnabled: false,
        meals: MealType.reminderTargets.map { mealType in
            MealNotificationSetting(
                mealType: mealType,
                isEnabled: true,
                hour: defaultHour(for: mealType),
                minute: 0
            )
        },
        weight: .default
    )

    func setting(for mealType: MealType) -> MealNotificationSetting? {
        meals.first { $0.mealType == mealType }
    }

    func updatingSetting(
        for mealType: MealType,
        transform: (MealNotificationSetting) -> MealNotificationSetting
    ) -> NotificationSettings {
        var updated = self
        updated.meals = meals.map { setting in
            setting.mealType == mealType ? transform(setting) : setting
        }
        return updated
    }

    func updatingWeightSetting(
        transform: (WeightNotificationSetting) -> WeightNotificationSetting
    ) -> NotificationSettings {
        var updated = self
        updated.weight = transform(weight)
        return updated
    }

    private static func defaultHour(for mealType: MealType) -> Int {
        switch mealType {
        case .breakfast: 9
        case .lunch: 13
        case .dinner: 20
        case .snack: 15
        }
    }
}

// MARK: - NotificationSettingsStoreProtocol

/// @mockable
protocol NotificationSettingsStoreProtocol {
    func load() -> NotificationSettings
    func save(_ settings: NotificationSettings)
}

// MARK: - NotificationSettingsStore

final class NotificationSettingsStore: NotificationSettingsStoreProtocol {
    private let userDefaults: UserDefaults
    private let key = "notification_settings"

    init(userDefaults: UserDefaults = .standard) {
        self.userDefaults = userDefaults
    }

    func load() -> NotificationSettings {
        guard let data = userDefaults.data(forKey: key) else {
            return .default
        }
        do {
            return try JSONDecoder().decode(NotificationSettings.self, from: data)
        } catch {
            logger.error("通知設定のデコード失敗（デフォルト値を使用）: \(error.localizedDescription)")
            return .default
        }
    }

    func save(_ settings: NotificationSettings) {
        do {
            let data = try JSONEncoder().encode(settings)
            userDefaults.set(data, forKey: key)
        } catch {
            logger.error("通知設定のエンコード失敗: \(error.localizedDescription)")
        }
    }
}
