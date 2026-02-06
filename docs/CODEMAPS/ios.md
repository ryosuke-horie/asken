# iOSアプリアーキテクチャ

最終更新: 2026-02-04
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
│   │   │   └── Meal.swift
│   │   ├── Network/            # API通信
│   │   │   ├── APIClient.swift       # AuthServiceProvider含む
│   │   │   ├── APIEndpoint.swift
│   │   │   └── APIError.swift
│   │   └── Repositories/       # データアクセス
│   │       └── MealRepository.swift
│   ├── Features/               # 機能モジュール
│   │   ├── Auth/               # 認証UI
│   │   │   ├── LoginView.swift
│   │   │   └── LoginViewModel.swift
│   │   └── Meals/              # 食事
│   │       ├── MealsView.swift
│   │       ├── MealsViewModel.swift
│   │       ├── MealInputView.swift
│   │       ├── MealInputViewModel.swift
│   │       ├── Models/
│   │       │   └── FoodEditItem.swift
│   │       ├── ViewModels/
│   │       │   └── NutritionEditorViewModel.swift
│   │       └── Views/
│   │           ├── NutritionEditorView.swift
│   │           └── FoodItemEditRow.swift
│   ├── Shared/
│   │   ├── Components/
│   │   │   └── NutritionSummaryCard.swift
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
    ├── Generated/
    │   └── MockGenerated.swift  # Mockolo生成
    ├── AuthManagerTests.swift
    ├── Features/
    │   └── Meals/
    │       ├── MealInputViewModelTests.swift
    │       └── MealInputManualFoodTests.swift
    └── Disabled/                # 一時無効化テスト
        ├── MealsViewModelTests.swift
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
    static var shared: FirebaseAuthServiceProtocol {
        #if DEBUG && targetEnvironment(simulator)
        return MockFirebaseAuthService()
        #else
        return FirebaseAuthService.shared
        #endif
    }
}
```

### 画面構成

認証状態に応じて画面を切り替え:
- 未認証: LoginView（シミュレータでは「開発用ログイン」ボタン表示）
- 認証済み: MealsView（食事記録画面）

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
| MealsView.swift | 食事一覧UI |
| MealInputView.swift | 食事入力UI |
| NutritionEditorView.swift | 栄養素編集UI |
| FoodItemEditRow.swift | 食品アイテム行 |
| FoodEditItem.swift | 編集用食品モデル |

## Core層

### Models (UchikomiCore/Models/ + Uchikomi/Core/Models/)

| ファイル | 場所 | 内容 |
|:---|:---|:---|
| Auth.swift | UchikomiCore | User, FirebaseAuthUser, GoogleCredential, FirebaseAuthError |
| Meal.swift | Uchikomi | 食事・栄養素モデル (MealType, NutritionInfo, DailyMeals) |

### Network (Core/Network/)

| ファイル | 責務 |
|:---|:---|
| APIClient.swift | HTTP通信 (actor)、AuthServiceProvider |
| APIEndpoint.swift | エンドポイント定義 |
| APIError.swift | エラー定義 |

### Repositories (Core/Repositories/)

| ファイル | 責務 |
|:---|:---|
| MealRepository.swift | 食事データアクセス |

## 共通コンポーネント (Shared/)

| ファイル | 用途 |
|:---|:---|
| NutritionSummaryCard.swift | 栄養素サマリーカード |
| Theme.swift | アプリテーマ定義 |

## データモデル

### 食事関連 (Meal.swift)

```swift
enum MealType: String, Codable, CaseIterable {
    case breakfast, lunch, dinner, snack
}

struct NutritionInfo: Codable, Identifiable {
    let name: String
    let estimatedAmount: String
    let caloriesKcal: Double
    let proteinG: Double
    let fatG: Double
    let carbohydratesG: Double
}

struct DailyMeals: Codable {
    let date: String
    let meals: MealsByType
    let dailyTotal: DailyTotal
}
```

### 認証関連 (UchikomiCore/Models/Auth.swift)

```swift
struct User: Codable, Sendable {
    let id: String
    let email: String
    let name: String?
}

struct FirebaseAuthUser: Sendable {
    let uid: String
    let email: String?
    let displayName: String?
}

struct GoogleCredential: Sendable {
    let idToken: String
    let accessToken: String
}

enum FirebaseAuthError: Error {
    case notSignedIn
    case tokenRetrievalFailed
    case signInFailed(Error)
    case signOutFailed(Error)
}
```

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
    └── MealsView (認証済み)
        ├── MealsViewModel
        │   └── MealRepository
        │       └── APIClient
        │           └── AuthServiceProvider.shared.getIDToken()
        └── MealInputView
            └── MealInputViewModel
                └── MealRepository
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

### エンドポイント

```swift
enum APIEndpoint {
    case dailyMeals(date: String)
    case analyze
    case analysisStatus(id: String)
    case analysisResult(id: String)
    // ...
}
```

## テスト構成

UchikomiTestsはUchikomiCoreフレームワークのみに依存し、Firebase SDKを初期化せずにテスト実行可能。

### ユニットテスト

| ファイル | テスト対象 | 状態 |
|:---|:---|:---|
| AuthManagerTests.swift | 認証状態管理 | 有効 |
| MealInputViewModelTests.swift | 食事入力ロジック | 有効 |
| MealInputManualFoodTests.swift | 食事手入力ロジック | 有効 |
| MealsViewModelTests.swift | 食事一覧ロジック | 一時無効化 (Disabled/) |

### スナップショットテスト

| ファイル | テスト対象 | 状態 |
|:---|:---|:---|
| NutritionSummaryCardSnapshotTests.swift | 栄養素カードUI | 一時無効化 (Disabled/) |

### モック生成

Mockoloを使用してプロトコルからモックを自動生成:

```bash
task ios:generate-mocks
```

生成先: `UchikomiTests/Generated/MockGenerated.swift`

### テスト実行

```bash
# Firebase SDK初期化なしでテスト実行可能
task ios:test
```

## 関連コードマップ

- [全体アーキテクチャ](./architecture.md)
- [データモデル](./data.md)
