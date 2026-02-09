import Foundation
import Testing
import UserNotifications

@testable import Uchikomi

@Suite
struct NotificationSettingsViewModelTests {
    private func makeStoreMock(
        settings: NotificationSettings = .default
    ) -> NotificationSettingsStoreProtocolMock {
        let mock = NotificationSettingsStoreProtocolMock()
        mock.loadHandler = { settings }
        mock.saveHandler = { _ in }
        return mock
    }

    private func makeSchedulerMock(
        authorizationStatus: UNAuthorizationStatus = .authorized
    ) -> NotificationSchedulerProtocolMock {
        let mock = NotificationSchedulerProtocolMock()
        mock.getAuthorizationStatusHandler = { authorizationStatus }
        mock.requestAuthorizationHandler = { authorizationStatus == .authorized }
        mock.scheduleAllNotificationsHandler = { _ in }
        mock.cancelAllNotificationsHandler = { }
        mock.cancelDeliveredNotificationHandler = { _ in }
        return mock
    }

    // MARK: - checkPermission

    @Test
    @MainActor
    func 権限チェックで許可済みの場合trueになるべき() async {
        let store = makeStoreMock()
        let scheduler = makeSchedulerMock(authorizationStatus: .authorized)
        let viewModel = NotificationSettingsViewModel(store: store, scheduler: scheduler)

        await viewModel.checkPermission()

        #expect(viewModel.systemPermissionGranted == true)
    }

    @Test
    @MainActor
    func 権限チェックで未許可の場合falseになるべき() async {
        let store = makeStoreMock()
        let scheduler = makeSchedulerMock(authorizationStatus: .denied)
        let viewModel = NotificationSettingsViewModel(store: store, scheduler: scheduler)

        await viewModel.checkPermission()

        #expect(viewModel.systemPermissionGranted == false)
    }

    // MARK: - toggleGlobalEnabled

    @Test
    @MainActor
    func グローバル通知をオンにできるべき() async {
        let store = makeStoreMock()
        let scheduler = makeSchedulerMock(authorizationStatus: .authorized)
        let viewModel = NotificationSettingsViewModel(store: store, scheduler: scheduler)
        await viewModel.checkPermission()

        await viewModel.toggleGlobalEnabled()

        #expect(viewModel.settings.isGlobalEnabled == true)
        #expect(store.saveCallCount == 1)
        #expect(scheduler.scheduleAllNotificationsCallCount == 1)
    }

    @Test
    @MainActor
    func グローバル通知をオフにできるべき() async {
        var enabledSettings = NotificationSettings.default
        enabledSettings.isGlobalEnabled = true
        let store = makeStoreMock(settings: enabledSettings)
        let scheduler = makeSchedulerMock(authorizationStatus: .authorized)
        let viewModel = NotificationSettingsViewModel(store: store, scheduler: scheduler)
        await viewModel.checkPermission()

        await viewModel.toggleGlobalEnabled()

        #expect(viewModel.settings.isGlobalEnabled == false)
        #expect(store.saveCallCount == 1)
        #expect(scheduler.scheduleAllNotificationsCallCount == 1)
    }

    @Test
    @MainActor
    func 権限未リクエスト時にオンにすると許可を要求すべき() async {
        let store = makeStoreMock()
        let scheduler = makeSchedulerMock(authorizationStatus: .notDetermined)
        scheduler.requestAuthorizationHandler = { true }
        let viewModel = NotificationSettingsViewModel(store: store, scheduler: scheduler)
        await viewModel.checkPermission()

        await viewModel.toggleGlobalEnabled()

        #expect(scheduler.requestAuthorizationCallCount == 1)
        #expect(viewModel.settings.isGlobalEnabled == true)
        #expect(viewModel.systemPermissionGranted == true)
    }

    @Test
    @MainActor
    func 権限拒否済みの状態でオンにするとアラートが表示されるべき() async {
        let store = makeStoreMock()
        let scheduler = makeSchedulerMock(authorizationStatus: .denied)
        let viewModel = NotificationSettingsViewModel(store: store, scheduler: scheduler)
        await viewModel.checkPermission()

        await viewModel.toggleGlobalEnabled()

        #expect(viewModel.showPermissionAlert == true)
        #expect(viewModel.settings.isGlobalEnabled == false)
        #expect(store.saveCallCount == 0)
    }

    // MARK: - toggleMealEnabled

    @Test
    @MainActor
    func 個別食事タイプの通知をオフにできるべき() async {
        var enabledSettings = NotificationSettings.default
        enabledSettings.isGlobalEnabled = true
        let store = makeStoreMock(settings: enabledSettings)
        let scheduler = makeSchedulerMock()
        let viewModel = NotificationSettingsViewModel(store: store, scheduler: scheduler)

        await viewModel.toggleMealEnabled(for: .breakfast)

        let breakfastSetting = viewModel.settings.setting(for: .breakfast)
        #expect(breakfastSetting?.isEnabled == false)
        #expect(store.saveCallCount == 1)
        #expect(scheduler.scheduleAllNotificationsCallCount == 1)
    }

    @Test
    @MainActor
    func 個別食事タイプの通知をオンにできるべき() async {
        var settings = NotificationSettings.default
        settings.isGlobalEnabled = true
        settings = settings.updatingSetting(for: .lunch) { setting in
            var modified = setting
            modified.isEnabled = false
            return modified
        }
        let store = makeStoreMock(settings: settings)
        let scheduler = makeSchedulerMock()
        let viewModel = NotificationSettingsViewModel(store: store, scheduler: scheduler)

        await viewModel.toggleMealEnabled(for: .lunch)

        let lunchSetting = viewModel.settings.setting(for: .lunch)
        #expect(lunchSetting?.isEnabled == true)
    }

    // MARK: - updateTime

    @Test
    @MainActor
    func 通知時間を更新できるべき() async {
        var settings = NotificationSettings.default
        settings.isGlobalEnabled = true
        let store = makeStoreMock(settings: settings)
        let scheduler = makeSchedulerMock()
        let viewModel = NotificationSettingsViewModel(store: store, scheduler: scheduler)

        await viewModel.updateTime(for: .lunch, hour: 12, minute: 30)

        let lunchSetting = viewModel.settings.setting(for: .lunch)
        #expect(lunchSetting?.hour == 12)
        #expect(lunchSetting?.minute == 30)
        #expect(store.saveCallCount == 1)
        #expect(scheduler.scheduleAllNotificationsCallCount == 1)
    }

    // MARK: - toggleGlobalEnabled エッジケース

    @Test
    @MainActor
    func 権限リクエストが失敗した場合にアラートが表示されるべき() async {
        let store = makeStoreMock()
        let scheduler = makeSchedulerMock(authorizationStatus: .notDetermined)
        scheduler.requestAuthorizationHandler = { throw NSError(domain: "test", code: -1) }
        let viewModel = NotificationSettingsViewModel(store: store, scheduler: scheduler)

        await viewModel.toggleGlobalEnabled()

        #expect(viewModel.showPermissionAlert == true)
        #expect(viewModel.settings.isGlobalEnabled == false)
        #expect(store.saveCallCount == 0)
    }

    @Test
    @MainActor
    func 権限リクエストでfalseが返った場合にアラートが表示されるべき() async {
        let store = makeStoreMock()
        let scheduler = makeSchedulerMock(authorizationStatus: .notDetermined)
        scheduler.requestAuthorizationHandler = { false }
        let viewModel = NotificationSettingsViewModel(store: store, scheduler: scheduler)

        await viewModel.toggleGlobalEnabled()

        #expect(viewModel.showPermissionAlert == true)
        #expect(viewModel.settings.isGlobalEnabled == false)
        #expect(viewModel.systemPermissionGranted == false)
        #expect(store.saveCallCount == 0)
    }

    // MARK: - updateTime バリデーション

    @Test
    @MainActor
    func 範囲外の時間がクランプされるべき() async {
        var settings = NotificationSettings.default
        settings.isGlobalEnabled = true
        let store = makeStoreMock(settings: settings)
        let scheduler = makeSchedulerMock()
        let viewModel = NotificationSettingsViewModel(store: store, scheduler: scheduler)

        await viewModel.updateTime(for: .breakfast, hour: 25, minute: 70)

        let breakfastSetting = viewModel.settings.setting(for: .breakfast)
        #expect(breakfastSetting?.hour == 23)
        #expect(breakfastSetting?.minute == 59)
    }

    @Test
    @MainActor
    func 負の時間が0にクランプされるべき() async {
        var settings = NotificationSettings.default
        settings.isGlobalEnabled = true
        let store = makeStoreMock(settings: settings)
        let scheduler = makeSchedulerMock()
        let viewModel = NotificationSettingsViewModel(store: store, scheduler: scheduler)

        await viewModel.updateTime(for: .dinner, hour: -1, minute: -5)

        let dinnerSetting = viewModel.settings.setting(for: .dinner)
        #expect(dinnerSetting?.hour == 0)
        #expect(dinnerSetting?.minute == 0)
    }

    // MARK: - save値の検証

    @Test
    @MainActor
    func toggleGlobalEnabledで保存された設定値が正しいべき() async {
        let store = makeStoreMock()
        var savedSettings: NotificationSettings?
        store.saveHandler = { savedSettings = $0 }
        let scheduler = makeSchedulerMock(authorizationStatus: .authorized)
        let viewModel = NotificationSettingsViewModel(store: store, scheduler: scheduler)

        await viewModel.toggleGlobalEnabled()

        #expect(savedSettings?.isGlobalEnabled == true)
    }

    // MARK: - グローバルオフ時の挙動

    @Test
    @MainActor
    func グローバルオフ時にスケジュールがキャンセルされるべき() async {
        var enabledSettings = NotificationSettings.default
        enabledSettings.isGlobalEnabled = true
        let store = makeStoreMock(settings: enabledSettings)
        let scheduler = makeSchedulerMock()
        let viewModel = NotificationSettingsViewModel(store: store, scheduler: scheduler)
        await viewModel.checkPermission()

        await viewModel.toggleGlobalEnabled()

        #expect(viewModel.settings.isGlobalEnabled == false)
        // scheduleAllNotifications は isGlobalEnabled=false で cancelAllNotifications を呼ぶ
        #expect(scheduler.scheduleAllNotificationsCallCount == 1)
    }
}
