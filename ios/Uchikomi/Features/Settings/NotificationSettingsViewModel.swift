import Foundation
import os
import UserNotifications

private let logger = Logger(
    subsystem: Bundle.main.bundleIdentifier ?? "Uchikomi",
    category: "NotificationSettingsViewModel"
)

// MARK: - NotificationSettingsViewModel

@Observable
final class NotificationSettingsViewModel {
    var settings: NotificationSettings
    var systemPermissionGranted = false
    var showPermissionAlert = false

    private let store: NotificationSettingsStoreProtocol
    private let scheduler: NotificationSchedulerProtocol

    init(
        store: NotificationSettingsStoreProtocol = NotificationSettingsStore(),
        scheduler: NotificationSchedulerProtocol = NotificationManager()
    ) {
        self.store = store
        self.scheduler = scheduler
        self.settings = store.load()
    }

    // MARK: - Permission

    func checkPermission() async {
        let status = await scheduler.getAuthorizationStatus()
        systemPermissionGranted = (status == .authorized)
    }

    // MARK: - Global Toggle

    func toggleGlobalEnabled() async {
        if !settings.isGlobalEnabled {
            let status = await scheduler.getAuthorizationStatus()

            switch status {
            case .notDetermined:
                do {
                    let granted = try await scheduler.requestAuthorization()
                    systemPermissionGranted = granted
                    if !granted {
                        showPermissionAlert = true
                        return
                    }
                } catch {
                    logger.error("通知許可リクエスト失敗: \(error.localizedDescription)")
                    showPermissionAlert = true
                    return
                }
            case .denied:
                showPermissionAlert = true
                return
            case .authorized, .provisional, .ephemeral:
                systemPermissionGranted = true
            @unknown default:
                break
            }
        }

        settings.isGlobalEnabled.toggle()
        store.save(settings)
        await scheduler.scheduleAllNotifications(settings: settings)
    }

    // MARK: - Per-Meal Toggle

    func toggleMealEnabled(for mealType: MealType) async {
        settings = settings.updatingSetting(for: mealType) { setting in
            var updated = setting
            updated.isEnabled.toggle()
            return updated
        }
        store.save(settings)
        await scheduler.scheduleAllNotifications(settings: settings)
    }

    // MARK: - Time Update

    func updateTime(for mealType: MealType, hour: Int, minute: Int) async {
        settings = settings.updatingSetting(for: mealType) { setting in
            var updated = setting
            updated.hour = hour
            updated.minute = minute
            return updated
        }
        store.save(settings)
        await scheduler.scheduleAllNotifications(settings: settings)
    }
}
