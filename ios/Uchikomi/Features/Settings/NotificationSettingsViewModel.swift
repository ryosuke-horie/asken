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
    var schedulingErrorMessage: String?

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
        let previousSettings = settings
        store.save(settings)
        await scheduler.scheduleAllNotifications(settings: settings)

        // スケジュール失敗時にロールバック
        if let error = scheduler.lastSchedulingError {
            logger.error("通知スケジュール失敗、設定をロールバック: \(error.localizedDescription)")
            settings = previousSettings
            settings.isGlobalEnabled.toggle()
            store.save(settings)
            schedulingErrorMessage = "通知の登録に失敗しました。もう一度お試しください。"
        }
    }

    // MARK: - Meal Notifications

    func toggleMealEnabled(for mealType: MealType) async {
        await updateMealSetting(mealType) { setting in
            var updated = setting
            updated.isEnabled.toggle()
            return updated
        }
    }

    func updateTime(for mealType: MealType, hour: Int, minute: Int) async {
        await updateMealSetting(mealType) { [self] setting in
            var updated = setting
            updated.hour = clampedHour(hour)
            updated.minute = clampedMinute(minute)
            return updated
        }
    }

    private func updateMealSetting(
        _ mealType: MealType,
        transform: @escaping (MealNotificationSetting) -> MealNotificationSetting
    ) async {
        let previousSettings = settings
        settings = settings.updatingSetting(for: mealType, transform: transform)
        store.save(settings)
        await scheduler.scheduleAllNotifications(settings: settings)

        // スケジュール失敗時にロールバック
        if let error = scheduler.lastSchedulingError {
            logger.error("通知スケジュール失敗、設定をロールバック: \(error.localizedDescription)")
            settings = previousSettings
            store.save(settings)
            schedulingErrorMessage = "通知の登録に失敗しました。もう一度お試しください。"
        }
    }

    // MARK: - Weight Notifications

    func toggleWeightEnabled() async {
        await updateWeightSetting { setting in
            var updated = setting
            updated.isEnabled.toggle()
            return updated
        }
    }

    func updateWeightTime(hour: Int, minute: Int) async {
        await updateWeightSetting { [self] setting in
            var updated = setting
            updated.hour = clampedHour(hour)
            updated.minute = clampedMinute(minute)
            return updated
        }
    }

    private func updateWeightSetting(
        transform: @escaping (WeightNotificationSetting) -> WeightNotificationSetting
    ) async {
        let previousSettings = settings
        settings = settings.updatingWeightSetting(transform: transform)
        store.save(settings)
        await scheduler.scheduleAllNotifications(settings: settings)

        // スケジュール失敗時にロールバック
        if let error = scheduler.lastSchedulingError {
            logger.error("通知スケジュール失敗、設定をロールバック: \(error.localizedDescription)")
            settings = previousSettings
            store.save(settings)
            schedulingErrorMessage = "通知の登録に失敗しました。もう一度お試しください。"
        }
    }

    // MARK: - Private Helpers

    private func clampedHour(_ value: Int) -> Int {
        min(max(value, 0), 23)
    }

    private func clampedMinute(_ value: Int) -> Int {
        min(max(value, 0), 59)
    }
}
