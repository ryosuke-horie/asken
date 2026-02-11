import Foundation
import os
import UserNotifications

// MARK: - NotificationSchedulerProtocol

/// @mockable
protocol NotificationSchedulerProtocol {
    func requestAuthorization() async throws -> Bool
    func getAuthorizationStatus() async -> UNAuthorizationStatus
    func scheduleAllNotifications(settings: NotificationSettings) async
    func cancelAllNotifications() async
    func cancelDeliveredNotification(for mealType: MealType) async
    var lastSchedulingError: Error? { get }
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

    func requestAuthorization() async throws -> Bool {
        try await notificationCenter.requestAuthorization(options: [.alert, .sound, .badge])
    }

    func getAuthorizationStatus() async -> UNAuthorizationStatus {
        await notificationCenter.notificationSettings().authorizationStatus
    }

    func scheduleAllNotifications(settings: NotificationSettings) async {
        lastSchedulingError = nil
        await cancelAllNotifications()

        guard settings.isGlobalEnabled else { return }

        for mealSetting in settings.meals where mealSetting.isEnabled {
            await scheduleMealNotification(mealSetting)
        }

        if settings.weight.isEnabled {
            await scheduleWeightNotification(settings.weight)
        }
    }

    func cancelAllNotifications() async {
        notificationCenter.removeAllPendingNotificationRequests()
    }

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

    private func scheduleNotification<S: TimedNotificationSetting>(
        identifier: String,
        body: String,
        setting: S,
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
