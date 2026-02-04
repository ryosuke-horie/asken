# プラン: 食事記録の保存タイミング修正

## Linear Issue
- Issue: EDG-592
- URL: https://linear.app/ryosuke-horie/issue/EDG-592

## 概要

現在の問題:
- 分析時点でDBに保存されてしまう
- ユーザーが削除しても、別の分析をすると前の分析結果もDBに残る
- 一覧画面に削除したはずのデータが表示される

期待する動作:
- ユーザーが「保存」を押した時点で初めて一覧に表示される
- 保存前の分析結果は一覧に表示されない

---

## 解決策: `confirmed`フラグの追加

### 仕組み

1. 分析レコードに `confirmed: bool` フィールドを追加
2. 分析開始時: `confirmed: false` で作成
3. 保存時（PUT /api/history）: `confirmed: true` に更新
4. 一覧取得: `confirmed: true` のみを返す
5. 新規分析開始時: 同じ日付/食事タイプの未確定レコードを自動削除

### フロー図

```
現在:
分析開始 → DB保存(completed) → 一覧に表示される ← 問題

修正後:
分析開始 → DB保存(confirmed:false) → 一覧に表示されない
        ↓
    NutritionEditorで「保存」
        ↓
    confirmed:true に更新 → 一覧に表示される ← 正常
```

---

## 実装計画

### 1. データモデルの変更

**ファイル:** `backend/internal/repository/analysis_models.go`

```go
type firestoreAnalysisDocument struct {
    // ... 既存フィールド
    Confirmed bool `firestore:"confirmed"` // 追加
}
```

### 2. リポジトリ層の変更

**ファイル:** `backend/internal/repository/analysis_repository_firestore.go`

| メソッド | 変更内容 |
|:---|:---|
| `CreateRequest` | `confirmed: false` で作成 |
| `CreateRequestWithText` | `confirmed: false` で作成、同日/同食事タイプの未確定レコード削除 |
| `GetHistoryList` | `confirmed == true` フィルタを追加 |
| `GetDailyMeals` | `confirmed == true` フィルタを追加 |
| `UpdateResult` | `confirmed: true` に更新（保存確定） |

### 3. 未確定レコードのクリーンアップ

新規分析開始時に、同じ日付/食事タイプの `confirmed: false` レコードを削除:

```go
// CreateRequestWithText 内で実行
func (r *firestoreAnalysisRepository) deleteUnconfirmedRecords(ctx context.Context, userID string, mealDate time.Time, mealType string) error {
    // 同日/同食事タイプの未確定レコードを削除
}
```

### 4. iOS側の変更

変更なし。既存の `PUT /api/history` を呼ぶと自動的に `confirmed: true` になる。

---

## 修正対象ファイル

| ファイル | 変更内容 |
|:---|:---|
| `backend/internal/repository/analysis_models.go` | `Confirmed`フィールド追加 |
| `backend/internal/repository/analysis_repository_firestore.go` | 上記の各メソッド変更 |
| `backend/internal/repository/analysis_repository_firestore_test.go` | テスト更新 |

---

## 検証方法

1. バックエンドを再起動
2. iOSアプリをビルド
3. 以下のシナリオで動作確認:
   - 分析する → 一覧に表示されない ✓
   - 分析する → 保存する → 一覧に表示される ✓
   - 分析する → 削除 → 別の分析 → 保存 → 前のデータが表示されない ✓
