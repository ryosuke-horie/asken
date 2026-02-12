# プラン: E2Eテスト History API（履歴一覧・詳細・更新・削除）

## Linear Issue
- Issue: EDG-682
- URL: https://linear.app/ryosuke-horie/issue/EDG-682/e2eテスト-history-api履歴一覧詳細更新削除

## 概要
History API（`/api/history`）のE2Eテストを追加する。対象エンドポイントは以下の4つ：
1. `GET /api/history` - 履歴一覧取得
2. `GET /api/history/{id}` - 履歴詳細取得
3. `PUT /api/history/{id}` - 履歴更新
4. `DELETE /api/history/{id}` - 履歴削除

## 既存実装の確認結果

### API ハンドラー
- `history_handler.go`: HandleList, HandleDetail, HandleUpdate
- `history_delete_handler.go`: Handle (DELETE)
- 認証必須、userIDでスコープ
- レスポンス構造: `HistoryItem`, `HistoryDetail`

### E2E テスト構造
- `backend/e2e/analyze_test.go`: 参考実装
- `authenticatedClient(t, timeout)`: 認証済みクライアント取得
- `testClient`: 認証なしクライアント
- `CleanupTestData(ctx)`: テストデータ削除

### データモデル
```go
// HistoryItem（一覧）
type HistoryItem struct {
    ID                 uuid.UUID `json:"id"`
    InputType          InputType `json:"input_type"`
    ImagePath          string    `json:"image_path"`
    InputText          string    `json:"input_text"`
    CreatedAt          time.Time `json:"created_at"`
    MealType           string    `json:"meal_type"`
    MealDate           time.Time `json:"meal_date"`
    TotalCalories      float64   `json:"total_calories"`
    TotalProtein       float64   `json:"total_protein"`
    TotalFat           float64   `json:"total_fat"`
    TotalCarbohydrates float64   `json:"total_carbohydrates"`
}

// HistoryDetail（詳細）Foodsを含む
type HistoryDetail struct {
    HistoryItem
    Foods []gemini.NutritionInfo `json:"foods"`
}
```

## 実装計画

### 1. 新規ファイル作成
- `backend/e2e/history_test.go` を作成

### 2. テストケース実装

#### GET /api/history（一覧）
- `TestHistory_List_Success`: 認証済みで一覧取得成功
  - まず analyze API でテストデータを作成
  - `GET /api/history` で一覧取得
  - items配列、total、page、limitを検証
- `TestHistory_List_Unauthorized`: 認証なしで401
  - 認証トークンなしで `GET /api/history`
  - 401を確認

#### GET /api/history/{id}（詳細）
- `TestHistory_Detail_Success`: 認証済みで詳細取得成功
  - analyze API でテストデータ作成
  - `GET /api/history/{id}` で詳細取得
  - foods配列を含むことを確認
- `TestHistory_Detail_NotFound`: 存在しないIDで404
  - 存在しないUUIDで `GET /api/history/{id}`
  - 404を確認

#### PUT /api/history/{id}（更新）
- `TestHistory_Update_Success`: 認証済みで更新成功
  - analyze API でテストデータ作成
  - `PUT /api/history/{id}` で更新
  - レスポンスのfoodsを検証
- `TestHistory_Update_NotFound`: 存在しないIDで404
  - 存在しないUUIDで更新リクエスト
  - 404を確認
- `TestHistory_Update_InvalidRequest`: バリデーションエラーで400
  - 空のfoods配列や不正な値で送信
  - 400を確認

#### DELETE /api/history/{id}（削除）
- `TestHistory_Delete_Success`: 認証済みで削除成功
  - analyze API でテストデータ作成
  - `DELETE /api/history/{id}` で削除
  - 204 No Contentを確認
  - 再度GETで404を確認（削除されたことを検証）
- `TestHistory_Delete_NotFound`: 存在しないIDで404
  - 存在しないUUIDで削除リクエスト
  - 404を確認

### 3. ヘルパー関数追加（必要に応じて）
- `createTestHistory`: analyze API経由でテストデータを作成するヘルパー
- `waitForGeminiRateLimit`: 既存の関数を利用（5秒待機）

## 技術的な考慮事項

### レート制限対応
- Gemini APIのレート制限（5秒に1回）に対応
- analyze APIを呼ぶテストケースの先頭で `waitForGeminiRateLimit()` を実行

### クリーンアップ
- `TestMain` の `CleanupTestData` が自動的にテストデータを削除
- `users/{testUID}/analysisRequests` コレクションを削除

### 認証
- `authenticatedClient(t, timeout)` で認証済みクライアント取得
- `testClient` で認証なしクライアント

### テストの命名規則
- 日本語「〜すべき」表現を使用（プロジェクトルール）
- 例: `TestHistory_List_Success`

## テスト計画
- 全てのテストで認証の有無による挙動の違いを検証
- データの作成→操作→検証の流れを一貫して実装
- エラーケース（401, 404, 400）を網羅
