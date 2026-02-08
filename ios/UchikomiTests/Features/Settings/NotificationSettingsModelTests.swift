import Foundation
import Testing

@testable import Uchikomi

@Suite
struct NotificationSettingsModelTests {
    // MARK: - NotificationSettings Default

    @Test
    func デフォルト設定値が正しいべき() {
        let settings = NotificationSettings.default

        #expect(settings.isGlobalEnabled == false)
        #expect(settings.meals.count == 3)

        let breakfast = settings.setting(for: .breakfast)
        #expect(breakfast?.isEnabled == true)
        #expect(breakfast?.hour == 9)
        #expect(breakfast?.minute == 0)

        let lunch = settings.setting(for: .lunch)
        #expect(lunch?.isEnabled == true)
        #expect(lunch?.hour == 13)
        #expect(lunch?.minute == 0)

        let dinner = settings.setting(for: .dinner)
        #expect(dinner?.isEnabled == true)
        #expect(dinner?.hour == 20)
        #expect(dinner?.minute == 0)

        let snack = settings.setting(for: .snack)
        #expect(snack == nil)
    }

    // MARK: - setting(for:)

    @Test
    func settingForで正しい食事タイプを取得できるべき() {
        let settings = NotificationSettings.default

        let breakfast = settings.setting(for: .breakfast)
        #expect(breakfast?.mealType == .breakfast)

        let lunch = settings.setting(for: .lunch)
        #expect(lunch?.mealType == .lunch)

        let dinner = settings.setting(for: .dinner)
        #expect(dinner?.mealType == .dinner)
    }

    // MARK: - updatingSetting

    @Test
    func updatingSettingでイミュータブルに更新されるべき() {
        let original = NotificationSettings.default
        let updated = original.updatingSetting(for: .breakfast) { setting in
            var modified = setting
            modified.isEnabled = false
            modified.hour = 8
            modified.minute = 30
            return modified
        }

        // 元のオブジェクトは変更されない
        #expect(original.setting(for: .breakfast)?.isEnabled == true)
        #expect(original.setting(for: .breakfast)?.hour == 9)

        // 更新されたオブジェクトは変更されている
        #expect(updated.setting(for: .breakfast)?.isEnabled == false)
        #expect(updated.setting(for: .breakfast)?.hour == 8)
        #expect(updated.setting(for: .breakfast)?.minute == 30)

        // 他の食事タイプは変更されない
        #expect(updated.setting(for: .lunch)?.hour == 13)
        #expect(updated.setting(for: .dinner)?.hour == 20)
    }

    // MARK: - MealNotificationSetting.timeComponents

    @Test
    func timeComponentsが正しいDateComponentsを返すべき() {
        let setting = MealNotificationSetting(
            mealType: .breakfast,
            isEnabled: true,
            hour: 9,
            minute: 30
        )

        let components = setting.timeComponents
        #expect(components.hour == 9)
        #expect(components.minute == 30)
    }

    // MARK: - NotificationSettingsStore

    @Test
    func UserDefaultsへのsaveとloadが往復するべき() {
        let testDefaults = UserDefaults(suiteName: "test_notification_settings")!
        testDefaults.removePersistentDomain(forName: "test_notification_settings")

        let store = NotificationSettingsStore(userDefaults: testDefaults)

        var settings = NotificationSettings.default
        settings.isGlobalEnabled = true
        settings = settings.updatingSetting(for: .lunch) { setting in
            var modified = setting
            modified.hour = 12
            modified.minute = 30
            modified.isEnabled = false
            return modified
        }

        store.save(settings)
        let loaded = store.load()

        #expect(loaded == settings)
        #expect(loaded.isGlobalEnabled == true)
        #expect(loaded.setting(for: .lunch)?.hour == 12)
        #expect(loaded.setting(for: .lunch)?.minute == 30)
        #expect(loaded.setting(for: .lunch)?.isEnabled == false)

        testDefaults.removePersistentDomain(forName: "test_notification_settings")
    }

    @Test
    func データがない場合にデフォルト設定を返すべき() {
        let testDefaults = UserDefaults(suiteName: "test_notification_empty")!
        testDefaults.removePersistentDomain(forName: "test_notification_empty")

        let store = NotificationSettingsStore(userDefaults: testDefaults)
        let loaded = store.load()

        #expect(loaded == .default)

        testDefaults.removePersistentDomain(forName: "test_notification_empty")
    }
}
