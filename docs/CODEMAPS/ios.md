# iOSアプリアーキテクチャ

最終更新: 2026-02-22
フレームワーク: Swift, SwiftUI
エントリーポイント: ios/Uchikomi/App/UchikomiApp.swift

## ディレクトリ構造

```
ios/
├── Uchikomi/                    # メインアプリ (Firebase依存)
│   ├── App/
│   │   ├── UchikomiApp.swift   # アプリエントリーポイント
│   │   ├── AppDelegate.swift   # Firebase初期化
│   │   └── AppEnvironment.swift # 環境設定
│   ├── Core/
│   │   ├── Auth/               # 認証実装
│   │   │   ├── FirebaseAuthService.swift    # Firebase Auth実装
│   │   │   ├── AppleSignInManager.swift     # Apple Sign-In
│   │   │   └── MockFirebaseAuthService.swift # 開発用モック (#if DEBUG)
│   │   ├── Models/             # データモデル
│   │   │   ├── Ingredient.swift
│   │   │   ├── Meal.swift
│   │   │   ├── MenuSuggestion.swift
│   │   │   ├── Micronutrient.swift
│   │   │   ├── MyMenu.swift
│   │   │   ├── NutritionGoal.swift
│   │   │   └── UserProfile.swift
│   │   ├── Extensions/         # Swift拡張
│   │   │   └── KeyedDecodingContainer+ISO8601.swift
│   │   ├── Network/            # API通信
│   │   │   ├── APIClient.swift       # AuthServiceProvider含む
│   │   │   ├── APIEndpoint.swift
│   │   │   └── APIError.swift
│   │   ├── Notification/       # 通知機能
│   │   │   └── NotificationManager.swift
│   │   ├── Views/              # 共通ビュー
│   │   │   └── CameraView.swift      # カメラ撮影UI
│   │   └── Repositories/       # データアクセス
│   │       ├── IngredientRepository.swift
│   │       ├── MealRepository.swift
│   │       ├── MenuSuggestionRepository.swift
│   │       ├── MyMenuRepository.swift
│   │       ├── NutritionGoalRepository.swift
│   │       └── WeightRepository.swift
│   ├── Features/               # 機能モジュール
│   │   ├── Auth/               # 認証UI
│   │   ├── CookingSuggestion/  # メニューサジェストUI
│   │   ├── Meals/              # 食事
│   │   ├── MyMenu/             # マイメニュー
│   │   ├── Pantry/             # 食材管理（パントリー）
│   │   ├── Settings/           # 設定
│   │   └── Weight/             # 体重
│   ├── Shared/
│   │   ├── Components/
│   │   │   ├── CalorieProgressView.swift
│   │   │   ├── MicronutrientProgressSection.swift
│   │   │   ├── NutritionSummaryCard.swift
│   │   │   ├── PFCPieChart.swift
│   │   │   └── PFCProgressBar.swift
│   │   ├── Models/
│   │   │   └── NotificationSettings.swift
│   │   └── Theme.swift
│   └── Resources/
│       └── Info.plist          # URL Schemes (Google Sign-In)
├── UchikomiCore/                # コアフレームワーク (Firebase非依存)
│   ├── Auth/
│   │   ├── AuthManager.swift              # 認証状態管理 (@Observable)
│   │   └── FirebaseAuthServiceProtocol.swift # 認証プロトコル
│   └── Models/
│       └── Auth.swift                     # User, FirebaseAuthUser, GoogleCredential
└── UchikomiTests/               # テスト (UchikomiCoreのみ依存)
    └── Disabled/                # 一時無効化テスト
        ├── AuthManagerTests.swift
        ├── MealsViewModelTests.swift
        ├── Generated/
        │   └── MockGenerated.swift  # Mockolo生成
        ├── Features/
        │   ├── Meals/
        │   │   ├── MealInputViewModelTests.swift
        │   │   ├── MealInputViewModelSkipTests.swift
        │   │   ├── MealInputManualFoodTests.swift
        │   │   ├── FoodEditItemTests.swift
        │   │   ├── ImageFilenameValidatorTests.swift
        │   │   └── QuantityParserTests.swift
        │   ├── Settings/
        │   │   ├── NotificationSettingsModelTests.swift
        │   │   └── NotificationSettingsViewModelTests.swift
        │   └── Weight/
        │       ├── WeightViewModelTests.swift
        │       └── WeightInputViewModelTests.swift
        └── Snapshots/
            └── NutritionSummaryCardSnapshotTests.swift
```

## アーキテクチャパターン

MVVM + Repository パターンを採用。UchikomiCoreフレームワークでFirebase非依存のテストを実現。

```
View (SwiftUI)
    ↓ @State / @Environment
 ViewModel (@Observable)
    ↓
Repository (Protocol)
    ↓
APIClient (actor) ←── AuthServiceProvider
    ↓                        ↓
バックエンドAPI        FirebaseAuthService / MockFirebaseAuthService
```

### フレームワーク構成

```
UchikomiCore (Firebase非依存)
├── Protocol (FirebaseAuthServiceProtocol, AppleSignInManagerProtocol)
├── Model (User, FirebaseAuthUser, GoogleCredential)
└── AuthManager (@Observable)

Uchikomi App (Firebase依存)
├── FirebaseAuthService (実装)
├── MockFirebaseAuthService (開発用モック)
├── AppleSignInManager (実装)
└── LoginViewModel (UIフロー)

UchikomiTests (Firebase非依存)
└── UchikomiCoreのみに依存 → テスト実行可能
```

## 認証アーキテクチャ

Firebase Authenticationを使用。Google Sign-In / Apple Sign-Inに対応。

### 本番環境

```
LoginView → LoginViewModel → FirebaseAuthService → Firebase Auth
                                    ↓
                              IDトークン取得
                                    ↓
APIClient ←── AuthServiceProvider.shared.getIDToken()
    ↓
Go Backend (Authorization: Bearer {token})
```

### 開発環境（シミュレータ）

`#if DEBUG && targetEnvironment(simulator)`で自動切り替え:

```
LoginView (「開発用ログイン」ボタン)
    ↓
MockFirebaseAuthService
    ↓
固定トークン "dev-mock-token" を返す
    ↓
Go Backend (DevAuthMiddleware で検証)
```

## アプリ構成

### エントリーポイント (App/)

```swift
@main
struct UchikomiApp: App {
    @State private var authManager: AuthManager

    init() {
        // AuthServiceProvider.shared を使用してAPIClientと共有
        _authManager = State(initialValue: AuthManager(
            firebaseAuthService: AuthServiceProvider.shared,
            appleSignInManager: AppleSignInManager()
        ))
    }

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environment(authManager)
        }
    }
}
```

### AuthServiceProvider

APIClientとAuthManagerで同じ認証サービスを共有:

```swift
enum AuthServiceProvider {
    private static var _shared: FirebaseAuthServiceProtocol?
    static var shared: FirebaseAuthServiceProtocol {
        if let service = _shared { return service }
        #if DEBUG && targetEnvironment(simulator)
        let service = MockFirebaseAuthService()
        #else
        let service = FirebaseAuthService.shared
        #endif
        _shared = service
        return service
    }
}
```

### 画面構成

認証状態に応じて画面を切り替え:
- 未認証: LoginView（シミュレータでは「開発用ログイン」ボタン表示）
- 認証済み: MainTabView
  - タブ1: MealsView（食事記録画面）
  - タブ2: WeightView（体重記録画面）
  - タブ3: PantryListView（食材管理画面）
  - タブ4: MyMenuListView（マイメニュー画面）
  - タブ5: SettingsView（設定画面）

## 機能モジュール

### 認証 (UchikomiCore/Auth/ + Uchikomi/Core/Auth/)

| ファイル | 場所 | 責務 |
|:---|:---|:---|
| AuthManager.swift | UchikomiCore | 認証状態管理 (@Observable) |
| FirebaseAuthServiceProtocol.swift | UchikomiCore | 認証サービスプロトコル |
| Auth.swift | UchikomiCore | User, FirebaseAuthUser, GoogleCredential |
| FirebaseAuthService.swift | Uchikomi | Firebase Auth実装 |
| MockFirebaseAuthService.swift | Uchikomi | 開発用モック (#if DEBUG) |
| AppleSignInManager.swift | Uchikomi | Apple Sign-In実装 |

### 認証UI (Features/Auth/)

| ファイル | 責務 |
|:---|:---|
| LoginViewModel.swift | ログイン画面ロジック、Google Sign-In UIフロー |
| LoginView.swift | ログイン画面UI（シミュレータでは開発用ボタン表示） |

### 食事 (Features/Meals/)

| ファイル | 責務 |
|:---|:---|
| MealsViewModel.swift | 食事一覧ロジック |
| MealInputViewModel.swift | 食事入力ロジック |
| NutritionEditorViewModel.swift | 栄養素編集ロジック |
| MealsView.swift | 食事一覧UI（メニューサジェストへの導線含む） |
| MealInputView.swift | 食事入力UI |
| NutritionEditorView.swift | 栄養素編集UI |
| NutritionGoalSettingView.swift | 栄養目標設定UI（推奨カロリー計算機能付き） |
| FoodItemEditRow.swift | 食品アイテム行 |
| FoodEditItem.swift | 編集用食品モデル（量変更時の栄養素再計算ロジック含む） |
| ImageFilenameValidator.swift | 画像ファイル名バリデーション |
| MeasurementUnit.swift | 計量単位定義（g, ml, 杯, 人前, 個, 枚, 本, 切, 匹, 尾, パック, 袋, 束, 丁, 缶, 合, 玉, 粒） |
| QuantityParser.swift | 量の文字列パーサー（数値と単位の抽出） |

### メニューサジェスト (Features/CookingSuggestion/)

| ファイル | 責務 |
|:---|:---|
| CookingSuggestionViewModel.swift | サジェスト一覧・リクエストロジック |
| RecipeDetailViewModel.swift | レシピ詳細・採用・却下ロジック |
| SuggestionListView.swift | サジェスト一覧UI |
| SuggestionRequestView.swift | サジェストリクエストUI |
| RecipeDetailView.swift | レシピ詳細UI |

### 食材管理 (Features/Pantry/)

| ファイル | 責務 |
|:---|:---|
| PantryViewModel.swift | 食材一覧・管理ロジック |
| PantryListView.swift | 食材一覧UI |
| IngredientEditView.swift | 食材編集UI |
| ReceiptScanView.swift | レシート読取UI |

### マイメニュー (Features/MyMenu/)

| ファイル | 責務 |
|:---|:---|
| MyMenuListView.swift | マイメニュー一覧UI |
| MyMenuListViewModel.swift | マイメニュー一覧ロジック |
| MyMenuEditView.swift | マイメニュー編集UI |
| MyMenuEditViewModel.swift | マイメニュー編集ロジック |
| MyMenuSelectionView.swift | マイメニュー選択UI |

### 体重 (Features/Weight/)

| ファイル | 責務 |
|:---|:---|
| WeightViewModel.swift | 体重一覧・チャート・目標ロジック |
| WeightInputViewModel.swift | 体重入力・更新・削除ロジック |
| WeightView.swift | 体重メイン画面UI（チャート + 記録一覧） |
| WeightInputView.swift | 体重入力UI（±0.1kg増減ボタン） |
| WeightChartView.swift | Swift Chartsによる体重推移グラフ |
| WeightGoalCard.swift | 現在体重・目標・差分表示 |
| WeightGoalSheet.swift | 目標体重設定シート |
| WeightRecordRow.swift | 体重記録行コンポーネント |

### 設定 (Features/Settings/)

| ファイル | 責務 |
|:---|:---|
| SettingsView.swift | 設定画面UI |
| SettingsViewModel.swift | 設定画面ロジック |
| NotificationSettingsView.swift | 通知設定UI（食事・体重リマインダー） |
| NotificationSettingsViewModel.swift | 通知設定ロジック |

## Core層

### Models (UchikomiCore/Models/ + Uchikomi/Core/Models/)

| ファイル | 場所 | 内容 |
|:---|:---|:---|
| Auth.swift | UchikomiCore | User, FirebaseAuthUser, GoogleCredential, FirebaseAuthError |
| Ingredient.swift | Uchikomi | 食材モデル (Ingredient, IngredientCategory) |
| Meal.swift | Uchikomi | 食事・栄養素モデル (MealType, NutritionInfo, DailyMeals) |
| MenuSuggestion.swift | Uchikomi | メニューサジェストモデル (MenuSuggestion, EstimatedNutrition) |
| Micronutrient.swift | Uchikomi | 微量栄養素モデル (MicronutrientType, MicronutrientInfo) |
| MyMenu.swift | Uchikomi | マイメニューモデル |
| NutritionGoal.swift | Uchikomi | 栄養目標モデル (NutritionGoal, NutritionPhase, PFCRatios, NutritionGoalCalculator) |
| UserProfile.swift | Uchikomi | ユーザー属性モデル (Gender, ActivityLevel, RecommendedCaloriesCalculator) |
| NotificationSettings.swift | Shared | 通知設定モデル (MealNotificationSetting, WeightNotificationSetting) |

### Extensions (Core/Extensions/)

| ファイル | 責務 |
|:---|:---|
| KeyedDecodingContainer+ISO8601.swift | ISO8601日付デコード拡張 |

### Views (Core/Views/)

| ファイル | 責務 |
|:---|:---|
| CameraView.swift | カメラ撮影UI（UIViewControllerRepresentable） |

### Network (Core/Network/)

| ファイル | 責務 |
|:---|:---|
| APIClient.swift | HTTP通信 (actor)、AuthServiceProvider |
| APIEndpoint.swift | エンドポイント定義 |
| APIError.swift | エラー定義 |

### Notification (Core/Notification/)

| ファイル | 責務 |
|:---|:---|
| NotificationManager.swift | 通知スケジュール管理（食事・体重リマインダー、食事記録時の通知キャンセル・翌日再スケジュール） |

NotificationSchedulerProtocolの主要メソッド:
- `scheduleAllNotifications(settings:)` - 全通知のスケジュール
- `cancelDeliveredNotification(for:)` - 配信済み通知の削除
- `handleMealRecorded(mealType:settings:)` - 食事記録時の通知キャンセルと翌日再スケジュール
- `refreshMealNotifications(settings:recordedMealTypes:)` - 記録済み食事を考慮した通知再スケジュール

### Repositories (Core/Repositories/)

| ファイル | 責務 |
|:---|:---|
| IngredientRepository.swift | 食材データアクセス |
| MealRepository.swift | 食事データアクセス |
| MenuSuggestionRepository.swift | メニューサジェストデータアクセス |
| MyMenuRepository.swift | マイメニューデータアクセス |
| NutritionGoalRepository.swift | 栄養目標取得・設定 |
| WeightRepository.swift | 体重記録・目標データアクセス |

## 共通コンポーネント (Shared/)

| ファイル | 用途 |
|:---|:---|
| CalorieProgressView.swift | カロリー進捗バー |
| MicronutrientProgressSection.swift | 微量栄養素進捗表示セクション |
| NutritionSummaryCard.swift | 栄養素サマリーカード |
| PFCPieChart.swift | PFCバランス円グラフ |
| PFCProgressBar.swift | PFC進捗バー |
| Theme.swift | アプリテーマ定義 |

## 依存関係図

```
UchikomiApp.swift
├── AuthServiceProvider.shared
│   ├── MockFirebaseAuthService (DEBUG + simulator)
│   └── FirebaseAuthService.shared (本番)
├── AuthManager (Environment)
│   └── FirebaseAuthServiceProtocol (AuthServiceProvider.shared)
└── ContentView
    ├── LoginView (未認証時)
    │   └── LoginViewModel
    │       └── AuthManager.signInWithGoogle()
    └── MainTabView (認証済み)
        ├── MealsView (タブ1)
        │   ├── MealsViewModel
        │   │   ├── MealRepository
        │   │   ├── NutritionGoalRepository
        │   │   └── WeightRepository
        │   │       └── APIClient
        │   │           └── AuthServiceProvider.shared.getIDToken()
        │   ├── NutritionGoalSettingView
        │   │   └── NutritionGoalRepository
        │   ├── MealInputView
        │   │   └── MealInputViewModel
        │   │       └── MealRepository
        │   └── SuggestionRequestView (メニューサジェスト導線)
        │       └── CookingSuggestionViewModel
        │           └── MenuSuggestionRepository
        ├── WeightView (タブ2)
        │   ├── WeightViewModel
        │   │   └── WeightRepository
        │   │       └── APIClient
        │   └── WeightInputView
        │       └── WeightInputViewModel
        │           └── WeightRepository
        ├── PantryListView (タブ3)
        │   └── PantryViewModel
        │       └── IngredientRepository
        │           └── APIClient
        ├── MyMenuListView (タブ4)
        │   └── MyMenuListViewModel
        │       └── MyMenuRepository
        │           └── APIClient
        └── SettingsView (タブ5)
            └── SettingsViewModel
                └── NotificationSettingsStore
```

## API通信 (APIClient)

```swift
// 認証サービスプロバイダー（APIClientとAuthManagerで共有）
enum AuthServiceProvider {
    static var shared: FirebaseAuthServiceProtocol {
        #if DEBUG && targetEnvironment(simulator)
        return MockFirebaseAuthService()
        #else
        return FirebaseAuthService.shared
        #endif
    }
}

actor APIClient {
    static let shared = APIClient()

    func request<T: Decodable>(endpoint: APIEndpoint, body: Encodable?) async throws -> T
    func uploadImage(endpoint: APIEndpoint, imageData: Data, ...) async throws -> AnalyzeResponse

    // 認証が必要なリクエストは AuthServiceProvider.shared.getIDToken() を使用
}
```

## テスト構成

UchikomiTestsはUchikomiCoreフレームワークのみに依存し、Firebase SDKを初期化せずにテスト実行可能。

現在全テストは `Disabled/` ディレクトリに移動され一時無効化中（macOS/Xcodeバージョン問題による）。

### テスト一覧

| ファイル | テスト対象 | 状態 |
|:---|:---|:---|
| AuthManagerTests.swift | 認証状態管理 | 一時無効化 (Disabled/) |
| MealInputViewModelTests.swift | 食事入力ロジック | 一時無効化 (Disabled/) |
| MealInputViewModelSkipTests.swift | 食事スキップロジック | 一時無効化 (Disabled/) |
| MealInputManualFoodTests.swift | 食事手入力ロジック | 一時無効化 (Disabled/) |
| FoodEditItemTests.swift | 食品編集モデル（栄養素再計算） | 一時無効化 (Disabled/) |
| ImageFilenameValidatorTests.swift | 画像ファイル名バリデーション | 一時無効化 (Disabled/) |
| QuantityParserTests.swift | 量パーサー | 一時無効化 (Disabled/) |
| NotificationSettingsModelTests.swift | 通知設定モデル | 一時無効化 (Disabled/) |
| NotificationSettingsViewModelTests.swift | 通知設定ロジック | 一時無効化 (Disabled/) |
| WeightViewModelTests.swift | 体重一覧ロジック | 一時無効化 (Disabled/) |
| WeightInputViewModelTests.swift | 体重入力ロジック | 一時無効化 (Disabled/) |
| MealsViewModelTests.swift | 食事一覧ロジック | 一時無効化 (Disabled/) |
| NutritionSummaryCardSnapshotTests.swift | 栄養素カードUI | 一時無効化 (Disabled/) |

### モック生成

Mockoloを使用してプロトコルからモックを自動生成:

```bash
task ios:generate-mocks
```

生成先: `UchikomiTests/Disabled/Generated/MockGenerated.swift`

## 関連コードマップ

- [全体アーキテクチャ](./architecture.md)
- [データモデル](./data.md)
