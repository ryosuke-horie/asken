import Foundation
import os
import UserNotifications

// MARK: - NotificationSchedulerProtocol

/// @mockable
protocol NotificationSchedulerProtocol {
    func requestAuthorization() async throws -> Bool
    func getAuthorizationStatus() async -> UNAuthorizationStatus
    func scheduleAllNotifications(settings: NotificationSettings) async
    var lastSchedulingError: Error? { get }
    func resetLastError()
}

// MARK: - NotificationManager

final class NotificationManager: NotificationSchedulerProtocol {
    private let notificationCenter: UNUserNotificationCenter
    private let logger = Logger(
        subsystem: Bundle.main.bundleIdentifier ?? "Uchikomi",
        category: "NotificationManager"
    )
    private(set) var lastSchedulingError: Error?

    init(notificationCenter: UNUserNotificationCenter = .current()) {
        self.notificationCenter = notificationCenter
    }

    func resetLastError() {
        lastSchedulingError = nil
    }

    func requestAuthorization() async throws -> Bool {
        try await notificationCenter.requestAuthorization(options: [.alert, .sound, .badge])
    }

    func getAuthorizationStatus() async -> UNAuthorizationStatus {
        await notificationCenter.notificationSettings().authorizationStatus
    }

    func scheduleAllNotifications(settings: NotificationSettings) async {
        await cancelAllNotifications()

        guard settings.isGlobalEnabled else { return }

        for mealSetting in settings.meals where mealSetting.isEnabled {
            await scheduleMealNotification(mealSetting)
            // 最初のエラーで中断（エラーがあれば即座に検知）
            if lastSchedulingError != nil {
                return
            }
        }

        if settings.weight.isEnabled {
            await scheduleWeightNotification(settings.weight)
        }
    }

    /// 配信済み通知を削除する（通知センターに表示済みの通知を除去）
    func cancelDeliveredNotification(for mealType: MealType) async {
        let identifier = notificationIdentifier(for: mealType)
        notificationCenter.removeDeliveredNotifications(withIdentifiers: [identifier])
    }

    /// 食事が記録された時に呼び出し、該当の通知をキャンセルして翌日に再スケジュールする
    func handleMealRecorded(mealType: MealType, settings: NotificationSettings) async {
        cancelPendingNotification(for: mealType)
        await cancelDeliveredNotification(for: mealType)
        if let setting = settings.meals.first(where: { $0.mealType == mealType && $0.isEnabled }) {
            await scheduleMealNotificationFromTomorrow(setting)
        }
    }

    /// 記録済みの食事タイプを考慮して食事通知を再スケジュールする
    ///
    /// まず全リマインダー対象の pending 通知をキャンセルし、その後：
    /// - 記録済みの食事: 翌日に非リピートで再スケジュール
    /// - 未記録の食事: repeating 通知として再スケジュール（通常通り）
    func refreshMealNotifications(
        settings: NotificationSettings,
        recordedMealTypes: Set<MealType>
    ) async {
        for mealType in MealType.reminderTargets {
            cancelPendingNotification(for: mealType)
        }

        for setting in settings.meals where setting.isEnabled {
            if recordedMealTypes.contains(setting.mealType) {
                await scheduleMealNotificationFromTomorrow(setting)
            } else {
                await scheduleMealNotification(setting)
            }
        }
    }

    // MARK: - Private

    private func cancelPendingNotification(for mealType: MealType) {
        let identifier = notificationIdentifier(for: mealType)
        notificationCenter.removePendingNotificationRequests(withIdentifiers: [identifier])
    }

    private func scheduleMealNotificationFromTomorrow(_ setting: MealNotificationSetting) async {
        guard let tomorrow = Calendar.current.date(
            byAdding: .day,
            value: 1,
            to: Calendar.current.startOfDay(for: Date())
        ) else {
            logger.warning("翌日の日付計算に失敗: \(setting.mealType.rawValue)")
            return
        }

        var components = Calendar.current.dateComponents([.year, .month, .day], from: tomorrow)
        components.hour = setting.hour
        components.minute = setting.minute

        let trigger = UNCalendarNotificationTrigger(dateMatching: components, repeats: false)
        await scheduleMealNotificationRequest(setting, trigger: trigger, logLabel: "翌日再スケジュール")
    }

    private func cancelAllNotifications() async {
        notificationCenter.removeAllPendingNotificationRequests()
    }

    private func scheduleMealNotification(_ setting: MealNotificationSetting) async {
        let trigger = UNCalendarNotificationTrigger(
            dateMatching: setting.timeComponents,
            repeats: true
        )
        await scheduleMealNotificationRequest(setting, trigger: trigger, logLabel: setting.mealType.rawValue)
    }

    private func scheduleMealNotificationRequest(
        _ setting: MealNotificationSetting,
        trigger: UNCalendarNotificationTrigger,
        logLabel: String
    ) async {
        let identifier = notificationIdentifier(for: setting.mealType)
        let content = createNotificationContent(body: "\(setting.mealType.displayName)を記録しませんか？")
        let request = UNNotificationRequest(identifier: identifier, content: content, trigger: trigger)

        do {
            try await notificationCenter.add(request)
            logger.info("通知スケジュール成功(\(logLabel)): \(setting.mealType.rawValue) \(setting.hour):\(setting.minute)")
        } catch {
            lastSchedulingError = error
            logger.error("通知スケジュール失敗(\(logLabel)): \(setting.mealType.rawValue): \(error.localizedDescription)")
        }
    }

    private func scheduleWeightNotification(_ setting: WeightNotificationSetting) async {
        await scheduleNotification(
            identifier: "weight_reminder",
            body: "今日の体重を記録しませんか？",
            setting: setting,
            logLabel: "体重"
        )
    }

    private func scheduleNotification(
        identifier: String,
        body: String,
        setting: some TimedNotificationSetting,
        logLabel: String
    ) async {
        let content = createNotificationContent(body: body)
        let trigger = UNCalendarNotificationTrigger(
            dateMatching: setting.timeComponents,
            repeats: true
        )
        let request = UNNotificationRequest(identifier: identifier, content: content, trigger: trigger)

        do {
            try await notificationCenter.add(request)
            logger.info("通知スケジュール成功: \(logLabel) \(setting.hour):\(setting.minute)")
        } catch {
            lastSchedulingError = error
            logger.error("通知スケジュール失敗: \(logLabel): \(error.localizedDescription)")
        }
    }

    private func createNotificationContent(body: String) -> UNMutableNotificationContent {
        let content = UNMutableNotificationContent()
        content.title = "ウチコミ"
        content.body = body
        content.sound = .default
        return content
    }

    private func notificationIdentifier(for mealType: MealType) -> String {
        "meal_reminder_\(mealType.rawValue)"
    }
}
