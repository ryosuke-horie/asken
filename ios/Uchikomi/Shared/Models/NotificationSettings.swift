import Foundation

// MARK: - MealNotificationSetting

struct MealNotificationSetting: Codable, Equatable {
    let mealType: MealType
    var isEnabled: Bool
    var hour: Int
    var minute: Int

    var timeComponents: DateComponents {
        var components = DateComponents()
        components.hour = hour
        components.minute = minute
        return components
    }
}

// MARK: - NotificationSettings

struct NotificationSettings: Codable, Equatable {
    var isGlobalEnabled: Bool
    var meals: [MealNotificationSetting]

    static let `default` = NotificationSettings(
        isGlobalEnabled: false,
        meals: MealType.reminderTargets.map { mealType in
            MealNotificationSetting(
                mealType: mealType,
                isEnabled: true,
                hour: defaultHour(for: mealType),
                minute: 0
            )
        }
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
        guard let data = userDefaults.data(forKey: key),
              let settings = try? JSONDecoder().decode(NotificationSettings.self, from: data) else {
            return .default
        }
        return settings
    }

    func save(_ settings: NotificationSettings) {
        guard let data = try? JSONEncoder().encode(settings) else { return }
        userDefaults.set(data, forKey: key)
    }
}
