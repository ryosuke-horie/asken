# プラン: 食事記録リマインダー通知機能

## Linear Issue
- Issue: EDG-688
- URL: https://linear.app/ryosuke-horie/issue/EDG-688

## コンテキスト

ユーザーが食事を記録し忘れることを防ぐため、食事時間帯ごと（朝食・昼食・夕食）にiOSローカル通知でリマインドする機能を追加する。現在アプリに通知機能・設定画面は一切ない。iOSローカル通知（UNUserNotificationCenter）を使うため、サーバー側の変更は不要で追加コストもゼロ。

## アプローチ: ハイブリッド方式

「固定リマインダー + アプリ起動時の最適化」を採用する。

- 基本動作: UNCalendarNotificationTrigger（repeats: true）で毎日同じ時間に通知をスケジュール
- 最適化: アプリがフォアグラウンドに来た時に getDailyMeals API で当日の記録を確認し、記録済みの食事タイプの配信済み通知を削除
- 通知文面: 「朝食を記録しませんか？」のような柔らかい表現

Background App Refresh は不採用（iOS側の実行保証なし、Firebase IDトークン期限の問題、バッテリー消費）。

## 新規作成ファイル

| ファイル | 内容 |
|:---|:---|
| `ios/Uchikomi/Core/Notification/NotificationManager.swift` | 通知スケジュール管理（NotificationSchedulerProtocol + 実装） |
| `ios/Uchikomi/Shared/Models/NotificationSettings.swift` | 設定モデル + UserDefaults永続化（NotificationSettingsStoreProtocol + 実装） |
| `ios/Uchikomi/Features/Settings/SettingsView.swift` | 設定画面（アカウント情報、通知設定リンク、ログアウト） |
| `ios/Uchikomi/Features/Settings/SettingsViewModel.swift` | 設定画面ViewModel |
| `ios/Uchikomi/Features/Settings/NotificationSettingsView.swift` | 通知設定詳細画面 |
| `ios/Uchikomi/Features/Settings/NotificationSettingsViewModel.swift` | 通知設定ViewModel |
| `ios/UchikomiTests/Features/Settings/NotificationSettingsViewModelTests.swift` | 通知設定VMテスト |
| `ios/UchikomiTests/Features/Settings/NotificationSettingsModelTests.swift` | 設定モデル・Storeテスト |

## 修正する既存ファイル

| ファイル | 修正内容 |
|:---|:---|
| `ios/Uchikomi/App/UchikomiApp.swift` | MainTabViewに「設定」タブ追加、scenePhase監視 |
| `ios/Uchikomi/Core/Models/Meal.swift` | MealType に `reminderTargets` 静的プロパティ追加 |

## 主要な設計

### MealType拡張（Meal.swift）

```swift
extension MealType {
    /// リマインダー通知の対象となる食事タイプ（snack除外）
    static let reminderTargets: [MealType] = [.breakfast, .lunch, .dinner]
}
```

### NotificationSettings（設定モデル）

```
NotificationSettings
  - isGlobalEnabled: Bool
  - meals: [MealNotificationSetting]  // MealType.reminderTargets の3つ
  - static default: 全ON、朝食9:00 / 昼食13:00 / 夕食20:00
  - setting(for:) -> MealNotificationSetting?
  - updatingSetting(for:transform:) -> NotificationSettings  // イミュータブル更新

MealNotificationSetting
  - mealType: MealType
  - isEnabled: Bool
  - hour: Int（0-23）, minute: Int（0-59）
  - timeComponents: DateComponents（computed）

NotificationSettingsStoreProtocol（@mockable）
  - load() -> NotificationSettings
  - save(_ settings: NotificationSettings)
  - 実装: UserDefaults + JSONEncoder/Decoder
```

### NotificationManager（通知スケジュール管理）

```
NotificationSchedulerProtocol（@mockable）
  - requestAuthorization() async throws -> Bool
  - getAuthorizationStatus() async -> UNAuthorizationStatus
  - scheduleAllNotifications(settings:) async
  - cancelAllNotifications() async
  - cancelDeliveredNotification(for mealType:) async

NotificationManager: NotificationSchedulerProtocol
  - init(notificationCenter: UNUserNotificationCenter = .current())
  - デフォルト引数パターンでDI（シングルトンではない）
  - 通知識別子: "meal_reminder_{mealType.rawValue}"
  - 通知文面: MealType.displayName + "を記録しませんか？"
```

### NotificationSettingsViewModel

```
@Observable NotificationSettingsViewModel
  - settings: NotificationSettings
  - systemPermissionGranted: Bool
  - showPermissionAlert: Bool  // 権限拒否時に設定アプリへのリンクを表示
  - init(store: NotificationSettingsStoreProtocol, scheduler: NotificationSchedulerProtocol)
  - checkPermission() async
  - toggleGlobalEnabled() async
    - 権限未取得時 → requestAuthorization()
    - 権限拒否済み（.denied）→ showPermissionAlert = true
  - toggleMealEnabled(for:) async
  - updateTime(for:hour:minute:) async
```

### SettingsViewModel

```
@Observable SettingsViewModel
  - init(authManager: AuthManager)
  - userName: String?（computed）
  - userEmail: String（computed）
  - logout()
```

### MainTabView変更

```
MainTabView
  - 3つ目のタブ「設定」（gearshape アイコン）を追加
  - @Environment(\.scenePhase) で .active 時に refreshTodayNotifications() を実行
    - NotificationSettingsStore.load() で設定を取得
    - isGlobalEnabled == false なら何もしない
    - MealRepository.getDailyMeals(date:) で当日の記録を確認
    - 記録済みの食事タイプ → cancelDeliveredNotification(for:)
    - APIエラー時は静かに失敗（通知はそのまま維持）
```

### 再利用する既存コード

| コード | ファイル | 用途 |
|:---|:---|:---|
| `MealType.displayName`, `.icon` | `Core/Models/Meal.swift` | 通知文面、UI表示 |
| `MealsByType.meals(for:)` | 同上 L99-106 | 記録有無の判定 |
| `MealRepository.getDailyMeals(date:)` | `Core/Repositories/MealRepository.swift` | 当日記録の確認 |
| `Theme.primary` | `Shared/Theme.swift` | UIテーマ |
| `AuthManager` | `UchikomiCore/Auth/AuthManager.swift` | ログアウト |

## 実装ステップ

1. MealType に `reminderTargets` を追加（Meal.swift）
2. 通知設定モデルと永続化を作成（NotificationSettings.swift）
3. NotificationManagerを作成（NotificationManager.swift）
4. Mockoloでモック生成（`task ios:generate-mocks`）
5. NotificationSettingsViewModelをTDDで作成（テスト先行）
6. SettingsViewModelを作成
7. NotificationSettingsView UIを作成
8. SettingsView UIを作成
9. MainTabViewに設定タブ追加 + scenePhase監視
10. `task ios:format` + `task ios:lint`
11. シミュレータで動作確認

## テスト計画

NotificationSettingsViewModelTests（Swift Testing、日本語命名）:
- グローバル通知をオンにできるべき
- 権限拒否済みの状態でオンにするとアラートが表示されるべき
- 権限未リクエスト時にオンにすると許可を要求すべき
- 個別食事タイプの通知をトグルできるべき
- 通知時間を更新できるべき
- グローバルオフ時にスケジュールがキャンセルされるべき

NotificationSettingsModelTests:
- デフォルト設定値が正しいべき
- UserDefaultsへのsave/loadが往復するべき
- setting(for:)で正しい食事タイプを取得できるべき
- updatingSettingでイミュータブルに更新されるべき

テスト不可能な部分（手動確認）:
- UNUserNotificationCenterの実際の通知配信
- scenePhase変化時のrefreshTodayNotifications
- View層のレイアウト（将来スナップショットテストで対応）

## 検証方法

1. `task ios:generate-mocks` でモック生成
2. `task ios:test` でユニットテスト実行
3. `task ios:format` + `task ios:lint` でコード品質チェック
4. シミュレータで以下を手動確認:
   - 設定タブが表示される
   - 通知権限リクエストダイアログが表示される
   - 権限拒否時に設定アプリへのリンクが表示される
   - 各食事タイプのON/OFFと時間変更が保存・復元される
   - アプリをバックグラウンドにして通知が届く
