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

    /// UchikomiApp.refreshTodayNotificationsから具象型で直接呼び出されるためプロトコルには含めない
    func cancelDeliveredNotification(for mealType: MealType) async {
        let identifier = notificationIdentifier(for: mealType)
        let delivered = await notificationCenter.deliveredNotifications()
        let toRemove = delivered
            .filter { $0.request.identifier == identifier }
            .map(\.request.identifier)
        if !toRemove.isEmpty {
            notificationCenter.removeDeliveredNotifications(withIdentifiers: toRemove)
        }
    }

    // MARK: - Private

    private func cancelAllNotifications() async {
        notificationCenter.removeAllPendingNotificationRequests()
    }

    private func scheduleMealNotification(_ setting: MealNotificationSetting) async {
        await scheduleNotification(
            identifier: notificationIdentifier(for: setting.mealType),
            body: "\(setting.mealType.displayName)を記録しませんか？",
            setting: setting,
            logLabel: setting.mealType.rawValue
        )
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
        let content = UNMutableNotificationContent()
        content.title = "ウチコミ"
        content.body = body
        content.sound = .default

        let trigger = UNCalendarNotificationTrigger(
            dateMatching: setting.timeComponents,
            repeats: true
        )

        let request = UNNotificationRequest(
            identifier: identifier,
            content: content,
            trigger: trigger
        )

        do {
            try await notificationCenter.add(request)
            logger.info("通知スケジュール成功: \(logLabel) \(setting.hour):\(setting.minute)")
        } catch {
            lastSchedulingError = error
            logger.error("通知スケジュール失敗: \(logLabel): \(error.localizedDescription)")
        }
    }

    private func notificationIdentifier(for mealType: MealType) -> String {
        "meal_reminder_\(mealType.rawValue)"
    }
}
