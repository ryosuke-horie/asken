# プラン: マイメニュー機能追加

## Linear Issue
- Issue: EDG-703
- URL: https://linear.app/ryosuke-horie/issue/EDG-703/マイメニュー機能追加

## 概要

ユーザーがよく食べる食事を「マイメニュー」として予め登録し、ワンタップで食事記録できる機能を追加します。

### 課題
- 毎回同じ食事を記録する手間がかかる
- よく食べる食事を素早く記録したい

### 解決策
- マイメニューを登録・管理する機能
- マイメニューから選択してワンタップで記録

### 既存コードの状況
- `InputTypeMylist` は定義済み（`backend/internal/repository/analysis_models.go:32`）
- `CreateRequestFromMylist` メソッドは実装済み（`backend/internal/repository/analysis_repository_firestore.go:518-562`）
- しかし、マイメニューの保存場所（Firestoreコレクション）と管理UIが未実装

## 実装計画

### フェーズ1: バックエンド実装（Go）

#### Step 1-1: マイメニューリポジトリ作成
**新規ファイル**: `backend/internal/repository/my_menu_repository.go`
**新規ファイル**: `backend/internal/repository/my_menu_repository_firestore.go`

マイメニューのCRUD操作を実装します。

```go
type MyMenuRepository interface {
    Create(ctx context.Context, userID string, name string, foods []gemini.NutritionInfo) (*MyMenuItem, error)
    List(ctx context.Context, userID string) ([]MyMenuItem, error)
    Get(ctx context.Context, userID string, menuID string) (*MyMenuItem, error)
    Update(ctx context.Context, userID string, menuID string, name string, foods []gemini.NutritionInfo) (*MyMenuItem, error)
    Delete(ctx context.Context, userID string, menuID string) error
}
```

Firestoreコレクション構造: `users/{user_id}/myMenu/{menu_id}`

#### Step 1-2: マイメニューハンドラー作成
**新規ファイル**: `backend/internal/handler/my_menu_handler.go`

HTTPエンドポイントを実装します。

| メソッド | パス | 説明 |
|:---|:---|:---|
| POST | `/api/my-menu` | マイメニュー作成 |
| GET | `/api/my-menu` | マイメニュー一覧取得 |
| GET | `/api/my-menu/:id` | マイメニュー詳細取得 |
| PUT | `/api/my-menu/:id` | マイメニュー更新 |
| DELETE | `/api/my-menu/:id` | マイメニュー削除 |
| POST | `/api/my-menu/:id/record` | マイメニューから食事記録 |

`/record` エンドポイントは既存の `CreateRequestFromMylist` を呼び出します。

#### Step 1-3: ルーティング設定
**変更ファイル**: `backend/cmd/server/main.go`

マイメニュールートを追加します。

### フェーズ2: iOS実装（Swift）

#### Step 2-1: ModelとRepository
**新規ファイル**: `ios/Uchikomi/Core/Models/MyMenu.swift`

```swift
struct MyMenuItem: Identifiable, Codable, Equatable {
    let id: String
    let name: String
    let foods: [NutritionInfo]
    let totalCalories: Double
    let totalProtein: Double
    let totalFat: Double
    let totalCarbohydrates: Double
    let createdAt: Date
    let updatedAt: Date
}
```

**新規ファイル**: `ios/Uchikomi/Core/Repositories/MyMenuRepository.swift`

**変更ファイル**: `ios/Uchikomi/Core/Network/APIEndpoint.swift`

マイメニュー用のエンドポイント定義を追加します。

#### Step 2-2: ViewModelとView
**新規ファイル**: `ios/Uchikomi/Features/MyMenu/MyMenuListViewModel.swift`
**新規ファイル**: `ios/Uchikomi/Features/MyMenu/MyMenuListView.swift`

マイメニュー一覧画面を実装します。

**新規ファイル**: `ios/Uchikomi/Features/MyMenu/MyMenuEditViewModel.swift`
**新規ファイル**: `ios/Uchikomi/Features/MyMenu/MyMenuEditView.swift`

マイメニュー登録/編集画面を実装します。

**新規ファイル**: `ios/Uchikomi/Features/MyMenu/MyMenuSelectionView.swift`

食事記録画面からマイメニューを選択する画面を実装します。

#### Step 2-3: 食事記録画面への統合
**変更ファイル**: `ios/Uchikomi/Features/Meals/MealInputView.swift`

「マイメニューから選択」セクションを追加します。

### 実装順序
1. バックエンド（Go）の実装とテスト
2. iOSのModelとRepository
3. iOSの一覧画面
4. iOSの登録/編集画面
5. iOSの食事記録画面への統合

## 技術的な考慮事項

### Firestoreコレクション構造
```
users/{user_id}/myMenu/{menu_id}
```

既存の `analysisRequests` とは別コレクションとし、マイメニュー削除時も既存の履歴は残すスナップショット方式を採用します。

### 既存コードの再利用
- バックエンド: `CreateRequestFromMylist` メソッドを活用
- iOS: `FoodEditItem` コンポーネントを食品入力に再利用
- iOS: 体重記録のクイックノートUIパターンを参考

### Firestoreインデックス
複合クエリなしで実装可能（ユーザー単位でスコープ済み）

### 制限事項
- マイメニュー名: 1-50文字
- 食品数: 1-100個（Firestoreドキュメントサイズ制限考慮）

## テスト計画

### バックエンド（Go）
**ターゲットファイル**:
- `my_menu_repository_firestore_test.go`
- `my_menu_handler_test.go`

**テストケース**:
- 作成: 正常、バリデーションエラー（名前空、食品空）
- 一覧取得: 空の場合、複数件の場合
- 詳細取得: 存在する場合、存在しない場合
- 更新: 正常、存在しないID
- 削除: 正常、存在しないID、既存履歴は残ること
- 食事記録: `InputType.mylist` で記録されること

### iOS（Swift）
- iOSテストは一時停止中のため、手動テストで代替

### 結合テストシナリオ
1. マイメニュー作成 → 一覧に表示される
2. マイメニューから食事記録 → 履歴に `InputType.mylist` で表示される
3. マイメニュー削除 → 一覧から消える、既存記録は残る

## Critical Files

### 新規作成ファイル
- `backend/internal/repository/my_menu_repository.go`
- `backend/internal/repository/my_menu_repository_firestore.go`
- `backend/internal/handler/my_menu_handler.go`
- `ios/Uchikomi/Core/Models/MyMenu.swift`
- `ios/Uchikomi/Core/Repositories/MyMenuRepository.swift`
- `ios/Uchikomi/Features/MyMenu/MyMenuListViewModel.swift`
- `ios/Uchikomi/Features/MyMenu/MyMenuListView.swift`
- `ios/Uchikomi/Features/MyMenu/MyMenuEditViewModel.swift`
- `ios/Uchikomi/Features/MyMenu/MyMenuEditView.swift`
- `ios/Uchikomi/Features/MyMenu/MyMenuSelectionView.swift`

### 変更ファイル
- `backend/cmd/server/main.go` - ルーティング追加
- `ios/Uchikomi/Core/Network/APIEndpoint.swift` - エンドポイント定義追加
- `ios/Uchikomi/Features/Meals/MealInputView.swift` - マイメニュー選択セクション追加

### 参照ファイル
- `backend/internal/repository/analysis_repository_firestore.go:518-562` - `CreateRequestFromMylist` 実装
- `ios/Uchikomi/Features/Meals/Models/FoodEditItem.swift` - 食品編集UI再利用
- `ios/Uchikomi/Features/Weight/WeightInputView.swift` - クイックノートUIパターン

## 検証方法

1. バックエンド: `task test` と `task lint` でテストとリント実行
2. バックエンド: `task run` でローカルサーバー起動
3. iOS: Xcodeでビルドして動作確認
4. マイメニュー作成 → 一覧表示 → 食事記録 → 履歴確認
5. マイメニュー削除 → 履歴が残っていることを確認
