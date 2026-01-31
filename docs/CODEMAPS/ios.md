# iOSアプリアーキテクチャ

最終更新: 2026-01-31
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
│   │   │   ├── Auth.swift
│   │   │   └── Meal.swift
│   │   ├── Network/            # API通信
│   │   │   ├── APIClient.swift
│   │   │   ├── APIEndpoint.swift
│   │   │   ├── APIError.swift
│   │   │   └── TokenManager.swift
│   │   └── Repositories/       # データアクセス
│   │       ├── AuthRepository.swift
│   │       └── MealRepository.swift
│   ├── Features/               # 機能モジュール
│   │   ├── Auth/               # 認証
│   │   │   ├── AuthManager.swift
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
│   └── Shared/
│       ├── Components/
│       │   └── NutritionSummaryCard.swift
│       └── Theme.swift
└── UchikomiTests/               # テスト
    ├── Generated/
    │   └── MockGenerated.swift  # Mockolo生成
    ├── AuthManagerTests.swift
    ├── MealsViewModelTests.swift
    └── Snapshots/
        └── NutritionSummaryCardSnapshotTests.swift
```

## アーキテクチャパターン

MVVM + Repository パターンを採用

```
View (SwiftUI)
    ↓ @State / @Environment
 ViewModel (@Observable)
    ↓
Repository (Protocol)
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

### 画面構成

認証状態に応じて画面を切り替え:
- 未認証: LoginView
- 認証済み: MealsView（食事記録画面）

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
| NutritionEditorViewModel.swift | 栄養素編集ロジック |
| MealsView.swift | 食事一覧UI |
| MealInputView.swift | 食事入力UI |
| NutritionEditorView.swift | 栄養素編集UI |
| FoodItemEditRow.swift | 食品アイテム行 |
| FoodEditItem.swift | 編集用食品モデル |

## Core層

### Models (Core/Models/)

| ファイル | 内容 |
|:---|:---|
| Auth.swift | 認証関連モデル (AuthResponse, User) |
| Meal.swift | 食事・栄養素モデル (MealType, NutritionInfo, DailyMeals) |

### Network (Core/Network/)

| ファイル | 責務 |
|:---|:---|
| APIClient.swift | HTTP通信 (actor) |
| APIEndpoint.swift | エンドポイント定義 |
| APIError.swift | エラー定義 |
| TokenManager.swift | トークン管理 (Keychain) |

### Repositories (Core/Repositories/)

| ファイル | 責務 |
|:---|:---|
| AuthRepository.swift | 認証データアクセス |
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
    └── MealsView (認証済み)
        ├── MealsViewModel
        │   └── MealRepository
        │       └── APIClient
        └── MealInputView
            └── MealInputViewModel
                └── MealRepository
```

## API通信 (APIClient)

```swift
actor APIClient {
    static let shared = APIClient()

    func request<T: Decodable>(endpoint: APIEndpoint, body: Encodable?) async throws -> T
    func uploadImage(endpoint: APIEndpoint, imageData: Data, ...) async throws -> AnalyzeResponse
}
```

### エンドポイント

```swift
enum APIEndpoint {
    case login
    case dailyMeals(date: String)
    case analyze
    case analysisStatus(id: String)
    case analysisResult(id: String)
}
```

## テスト構成

### ユニットテスト

| ファイル | テスト対象 |
|:---|:---|
| AuthManagerTests.swift | 認証状態管理 |
| MealsViewModelTests.swift | 食事一覧ロジック |

### スナップショットテスト

| ファイル | テスト対象 |
|:---|:---|
| NutritionSummaryCardSnapshotTests.swift | 栄養素カードUI |

### モック生成

Mockoloを使用してプロトコルからモックを自動生成:

```bash
task ios:generate-mocks
```

生成先: `UchikomiTests/Generated/MockGenerated.swift`

## 関連コードマップ

- [全体アーキテクチャ](./architecture.md)
- [データモデル](./data.md)
