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
                logger.warning("未知の通知権限ステータス: \(String(describing: status))")
                systemPermissionGranted = false
                return
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
        let clampedHour = min(max(hour, 0), 23)
        let clampedMinute = min(max(minute, 0), 59)
        settings = settings.updatingSetting(for: mealType) { setting in
            var updated = setting
            updated.hour = clampedHour
            updated.minute = clampedMinute
            return updated
        }
        store.save(settings)
        await scheduler.scheduleAllNotifications(settings: settings)
    }

    // MARK: - Weight Notification

    func toggleWeightEnabled() async {
        settings = settings.updatingWeightSetting { setting in
            var updated = setting
            updated.isEnabled.toggle()
            return updated
        }
        store.save(settings)
        await scheduler.scheduleAllNotifications(settings: settings)
    }

    func updateWeightTime(hour: Int, minute: Int) async {
        let clampedHour = min(max(hour, 0), 23)
        let clampedMinute = min(max(minute, 0), 59)
        settings = settings.updatingWeightSetting { setting in
            var updated = setting
            updated.hour = clampedHour
            updated.minute = clampedMinute
            return updated
        }
        store.save(settings)
        await scheduler.scheduleAllNotifications(settings: settings)
    }
}
