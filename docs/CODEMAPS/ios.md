# iOSアプリアーキテクチャ

最終更新: 2026-01-30
フレームワーク: Swift, SwiftUI
エントリーポイント: ios/Uchikomi/App/UchikomiApp.swift

## ディレクトリ構造

```
ios/
├── Uchikomi/                    # メインアプリ
│   ├── App/
│   │   ├── UchikomiApp.swift   # アプリエントリーポイント
│   │   └── AppEnvironment.swift # 環境設定
│   ├── Core/
│   │   ├── Models/             # データモデル
│   │   ├── Network/            # API通信
│   │   └── Repositories/       # データアクセス
│   ├── Features/               # 機能モジュール
│   │   ├── Auth/               # 認証
│   │   ├── Meals/              # 食事
│   │   └── Weight/             # 体重
│   └── Shared/
│       └── Components/         # 共通コンポーネント
├── UchikomiWidget/              # ウィジェット
│   ├── UchikomiWidget.swift    # ウィジェットエントリーポイント
│   ├── WidgetDataProvider.swift # データプロバイダ
│   ├── WeightWidget.swift      # 体重ウィジェット
│   └── CaloriesWidget.swift    # カロリーウィジェット
└── UchikomiTests/               # テスト
    ├── AuthManagerTests.swift
    └── MealsViewModelTests.swift
```

## アーキテクチャパターン

MVVM + Repository パターンを採用

```
View (SwiftUI)
    ↓ @State / @Environment
ViewModel (@Observable)
    ↓
Repository
    ↓
APIClient (actor)
    ↓
バックエンドAPI
```

## アプリ構成

### エントリーポイント (App/)

```swift
@main
struct UchikomiApp: App {
    @State private var authManager = AuthManager()

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environment(authManager)
        }
    }
}
```

### タブ構成

| タブ | View | 機能 |
|:---|:---|:---|
| 食事 | MealsView | 日次食事記録 |
| 体重 | WeightView | 体重記録・グラフ |

## 機能モジュール

### 認証 (Features/Auth/)

| ファイル | 責務 |
|:---|:---|
| AuthManager.swift | 認証状態管理 (@Observable) |
| LoginViewModel.swift | ログイン画面ロジック |
| LoginView.swift | ログイン画面UI |

### 食事 (Features/Meals/)

| ファイル | 責務 |
|:---|:---|
| MealsViewModel.swift | 食事一覧ロジック |
| MealInputViewModel.swift | 食事入力ロジック |
| MealsView.swift | 食事一覧UI |
| MealInputView.swift | 食事入力UI |

### 体重 (Features/Weight/)

| ファイル | 責務 |
|:---|:---|
| WeightViewModel.swift | 体重記録ロジック |
| WeightView.swift | 体重記録・グラフUI |

## Core層

### Models (Core/Models/)

| ファイル | 内容 |
|:---|:---|
| Auth.swift | 認証関連モデル |
| Meal.swift | 食事・栄養素モデル |
| Weight.swift | 体重モデル |

### Network (Core/Network/)

| ファイル | 責務 |
|:---|:---|
| APIClient.swift | HTTP通信 (actor) |
| APIEndpoint.swift | エンドポイント定義 |
| APIError.swift | エラー定義 |
| TokenManager.swift | トークン管理 |
| WidgetDataService.swift | ウィジェット用データ取得 |

### Repositories (Core/Repositories/)

| ファイル | 責務 |
|:---|:---|
| AuthRepository.swift | 認証データアクセス |
| MealRepository.swift | 食事データアクセス |
| WeightRepository.swift | 体重データアクセス |

## 共通コンポーネント (Shared/Components/)

| ファイル | 用途 |
|:---|:---|
| NutritionSummaryCard.swift | 栄養素サマリーカード |

## ウィジェット (UchikomiWidget/)

### ウィジェット構成

```swift
@main
struct UchikomiWidgetBundle: WidgetBundle {
    var body: some Widget {
        WeightWidget()
        CaloriesWidget()
    }
}
```

| ウィジェット | 用途 |
|:---|:---|
| WeightWidget | 現在の体重と目標表示 |
| CaloriesWidget | 今日のカロリー摂取量表示 |

### データ共有

App Groupsでメインアプリとウィジェットでデータ共有

```
Main App → UserDefaults (App Groups) → Widget
           ↓
        Keychain (トークン)
```

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

### 認証関連 (Auth.swift)

```swift
struct AuthResponse: Codable {
    let token: String
    let user: User
}

struct User: Codable {
    let id: String
    let email: String
    let name: String?
}
```

## 依存関係図

```
UchikomiApp.swift
├── AuthManager (Environment)
└── ContentView
    ├── LoginView (未認証時)
    │   └── LoginViewModel
    │       └── AuthRepository
    │           └── APIClient
    └── MainTabView (認証済み)
        ├── MealsView
        │   └── MealsViewModel
        │       └── MealRepository
        │           └── APIClient
        └── WeightView
            └── WeightViewModel
                └── WeightRepository
                    └── APIClient

UchikomiWidgetBundle
├── WeightWidget
│   └── WidgetDataProvider
│       └── WidgetDataService
└── CaloriesWidget
    └── WidgetDataProvider
```

## API通信 (APIClient)

```swift
actor APIClient {
    static let shared = APIClient()

    func request<T: Decodable>(endpoint: APIEndpoint, body: Encodable?) async throws -> T
    func uploadImage(endpoint: APIEndpoint, imageData: Data, ...) async throws -> AnalyzeResponse
}
```

### エンドポイント例

```swift
enum APIEndpoint {
    case login
    case dailyMeals(date: String)
    case weightRecords
    case analyze
    // ...
}
```

## 関連コードマップ

- [全体アーキテクチャ](./architecture.md)
- [データモデル](./data.md)
