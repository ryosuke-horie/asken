# プラン: PR #153 レビュー指摘事項の修正

## Linear Issue
- Issue: EDG-597
- URL: https://linear.app/ryosuke-horie/issue/EDG-597

## 背景
PR #153（食事記録の量/メニュー名変更時の栄養素再計算）に対する5エージェント並列レビューで、Critical 3件・Important 8件の指摘があった。このプランではこれらの修正を行う。

## 変更対象ファイル

| ファイル | 変更概要 |
|:---|:---|
| `backend/internal/handler/history_handler.go` | 鮮度チェック、バリデーション、スタックトレース、コメント修正 |
| `backend/internal/handler/history_handler_test.go` | テスト追加5件、既存テスト改善 |
| `ios/Uchikomi/Features/Meals/Models/FoodEditItem.swift` | docコメント修正のみ |

## 実装計画

### Step 1: docコメント修正（C2, I7, I8）

#### FoodEditItem.swift（C2/I7）
`recalculateNutrition()`のdocコメントを修正:
```swift
/// 量の変更に基づいて栄養素を再計算する。
/// 元の量と現在の量をパースし、同じ単位の場合のみ比率で再計算する。
/// パース失敗時やゼロ除算の場合は何も変更しない（現在の栄養素値がそのまま残る）。
/// キーストロークごとに呼ばれるため、入力途中のパース失敗は正常動作でありUIエラー表示は行わない。
```

#### history_handler.go（I8）
`recalculateAsync`のdocコメントを更新:
```go
// recalculateAsync はGemini APIで非同期に全食材の栄養素を一括再計算し、結果をFirestoreに保存する。
// 名前が変更された食材だけでなく、全食材をGemini APIに渡して再計算する。
// goroutineとして起動されるため、panicリカバリを含む。
// 書き込み前に鮮度チェックを行い、再計算中にユーザーが再保存した場合は書き込みをスキップする。
```

### Step 2: panicリカバリにスタックトレース追加（I4）

`history_handler.go`のimportに`"runtime/debug"`を追加し、recover内で`debug.Stack()`を出力:
```go
defer func() {
    if r := recover(); r != nil {
        log.Printf("Panic recovered in recalculateAsync for history %s: %v\nStack trace:\n%s", historyID, r, debug.Stack())
    }
}()
```

### Step 3: detectNameChanges の長さ不一致時のガード追加（I5）

食材数が異なる場合（追加/削除時）はインデックスベース比較が不正確になるため、早期リターンで再計算をスキップ:
```go
func detectNameChanges(oldFoods []gemini.NutritionInfo, newFoods []gemini.NutritionInfo) []int {
    if len(oldFoods) != len(newFoods) {
        return nil
    }
    var changed []int
    for i := 0; i < len(oldFoods); i++ {
        if oldFoods[i].Name != newFoods[i].Name {
            changed = append(changed, i)
        }
    }
    return changed
}
```
コメントも更新: 要素数が異なる場合は空を返し再計算をスキップする旨を記載。

既存テスト「新しい食材が追加された場合」「食材が削除された場合」はいずれも`expected: nil`なので変更不要。

### Step 4: リクエストバリデーション追加（I6）

`history_handler.go`のimportに`"fmt"`を追加。

`UpdateFoodItem`にValidateメソッドを追加:
```go
func (f UpdateFoodItem) Validate() error {
    if strings.TrimSpace(f.Name) == "" {
        return fmt.Errorf("food name is required")
    }
    if f.Calories < 0 || f.Protein < 0 || f.Fat < 0 || f.Carbohydrates < 0 {
        return fmt.Errorf("nutrition values must be non-negative")
    }
    return nil
}
```

`UpdateHistoryRequest`にValidateメソッドを追加:
```go
func (r UpdateHistoryRequest) Validate() error {
    for i, f := range r.Foods {
        if err := f.Validate(); err != nil {
            return fmt.Errorf("foods[%d]: %w", i, err)
        }
    }
    return nil
}
```

`HandleUpdate`のリクエストボディデコード直後にバリデーション呼び出しを追加:
```go
if err := req.Validate(); err != nil {
    log.Printf("Validation error: %v", err)
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}
```

### Step 5: エラーログ改善（C3）

`recalculateAsync`内のエラーログにuserID・食材数を追加:
```go
// Gemini APIエラー時
log.Printf("ERROR: Async nutrition recalculation failed for history %s, userID=%s, foodCount=%d: %v", historyID, userID, len(foodItems), err)

// Firestore保存エラー時
log.Printf("ERROR: Failed to save recalculated nutrition for history %s, userID=%s: %v", historyID, userID, err)
```

### Step 6: 鮮度チェック追加（C1）

`recalculateAsync`のUpdateResult呼び出し前に、現在のFirestoreデータを再読み込みして変更検知:

ヘルパー関数`foodsMatch`を追加:
```go
func foodsMatch(a, b []gemini.NutritionInfo) bool {
    if len(a) != len(b) {
        return false
    }
    for i := range a {
        if a[i].Name != b[i].Name || a[i].EstimatedAmount != b[i].EstimatedAmount {
            return false
        }
    }
    return true
}
```

`recalculateAsync`のGemini API成功後、UpdateResult前に鮮度チェック:
```go
currentDetail, err := h.repository.GetHistoryDetail(ctx, userID, historyID)
if err != nil {
    log.Printf("ERROR: Staleness check failed for history %s: %v", historyID, err)
    return
}
if !foodsMatch(currentFoods, currentDetail.Foods) {
    log.Printf("Skipping async recalculation for history %s: data modified during recalculation", historyID)
    return
}
```

### Step 7: テスト追加・改善

#### 7-1. `TestHistoryHandler_HandleUpdate_RecalculateAsyncGeminiError`（I1）
Gemini APIがエラーを返した場合、UpdateResultは同期保存の1回のみ呼ばれることを検証。
- mockRecalculatorの`CalculateNutritionFunc`がerrorを返す
- channelで非同期完了を待つ
- `updateCallCount == 1`をassert

#### 7-2. `TestHistoryHandler_HandleUpdate_GetHistoryDetailInternalError`（I2）
GetHistoryDetailが汎用エラー（非"見つかりません"）を返した場合、500を返すことを検証。
- mockRepoの`GetHistoryDetailFunc`が`errors.New("internal error")`を返す
- HTTPステータス500をassert

#### 7-3. 既存`WithNameChange`テストの改善（I3 + C1対応）
- `GetHistoryDetailFunc`に呼び出しカウンタ（`atomic.Int32`）を追加
  - 1回目（更新前比較）: 旧データ（Name: "白米"）を返す
  - 2回目以降（レスポンス用、鮮度チェック用）: 更新後データ（Name: "玄米"）を返す
- `UpdateResultFunc`にuserID/historyIDのassertを追加
- 2回目のUpdateResult（非同期保存）で再計算済み栄養素値をassert

#### 7-4. `TestHistoryHandler_HandleUpdate_StaleDataSkipped`（C1テスト）
鮮度チェックが失敗した場合、非同期保存がスキップされることを検証。
- recalculatorが正常に値を返す
- GetHistoryDetailの鮮度チェック時に異なるデータ（別ユーザーが保存した状態を模倣）を返す
- UpdateResultは同期保存の1回のみ呼ばれることをassert

#### 7-5. バリデーションテスト2件（I6テスト）
- `TestHistoryHandler_HandleUpdate_ValidationEmptyName`: 空名前で400
- `TestHistoryHandler_HandleUpdate_ValidationNegativeCalories`: 負のカロリーで400

#### 7-6. `TestFoodsMatch`（C1ヘルパーテスト）
- 一致、不一致、長さ違い、空スライスのケースをテーブル駆動テストで検証

## 実装順序
1. Step 1-2: コメント修正、スタックトレース（影響範囲小）
2. Step 3: detectNameChanges ガード（ロジック変更小、既存テスト変更不要）
3. Step 4: バリデーション追加（新規コード）
4. Step 5: ログ改善（文字列変更のみ）
5. Step 6: 鮮度チェック（最も複雑な変更）
6. Step 7: テスト追加・改善（全ステップの検証）

## 検証方法
```bash
# Goテスト
task test

# Goリント
task lint

# データレース検出
cd backend && go test -race ./internal/handler/

# iOSはdocコメント変更のみのためテスト不要
```
