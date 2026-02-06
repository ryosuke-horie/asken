# コーディングスタイル

## 基本原則

| 原則 | 内容 |
|:---|:---|
| KISS | 動作する最もシンプルな解決策を選ぶ |
| DRY | 共通ロジックを関数に抽出、コピペ禁止 |
| YAGNI | 必要になる前に機能を構築しない |

## イミュータビリティ（重要）

可能な限り値を変更せず、新しい値を作成すること。

### Go

```go
// NG: スライスの直接変更
func updateUser(users []User, idx int, name string) {
    users[idx].Name = name  // 直接変更
}

// OK: 新しいスライスを返す
func updateUser(users []User, idx int, name string) []User {
    updated := make([]User, len(users))
    copy(updated, users)
    updated[idx].Name = name
    return updated
}
```

### Swift

```swift
// OK: 構造体はValue Typeなのでイミュータブル
struct User {
    var name: String
}

func updateUser(_ user: User, name: String) -> User {
    var updated = user
    updated.name = name
    return updated
}
```

## ファイル構成

小さなファイルを多数作成することを推奨:

| 指標 | 基準 |
| :--- | :--- |
| ファイル行数（理想） | 200-400行 |
| ファイル行数（最大） | 800行 |
| 関数行数（最大） | 50行 |
| ネスト深度（最大） | 4レベル |

- 高凝集・低結合を意識する
- 大きなファイルからユーティリティを抽出する

## エラーハンドリング

### Go

エラーは必ず処理し、文脈情報を追加してラップすること:

```go
food, err := repo.GetFoodByID(ctx, id)
if err != nil {
    return nil, fmt.Errorf("failed to get food %s: %w", id, err)
}
```

### Swift

async/awaitとdo-catchで適切にエラーを処理すること:

```swift
do {
    let result = try await service.analyzeFoodImage(imageData)
    return result
} catch {
    throw AppError.analysisFailure(underlying: error)
}
```

## 入力検証

ユーザー入力は必ず検証すること:

```go
func validateSortOrder(order string) error {
    allowedOrders := map[string]bool{
        "asc":  true,
        "desc": true,
    }
    if !allowedOrders[order] {
        return errors.New("invalid sort order")
    }
    return nil
}
```

## コード品質チェックリスト

作業完了前に確認:

- [ ] コードが読みやすく、適切な命名がされている
- [ ] 関数が小さい（50行未満）
- [ ] ファイルが集中している（800行未満）
- [ ] 深いネストがない（4レベル以下）
- [ ] 適切なエラーハンドリングがされている
- [ ] デバッグ用のprint/log文が残っていない
- [ ] ハードコードされた値がない

## 非同期処理

### Go - 独立した処理はgoroutineで並列実行

```go
var wg sync.WaitGroup
wg.Add(2)

go func() {
    defer wg.Done()
    users, _ = fetchUsers(ctx)
}()
go func() {
    defer wg.Done()
    meals, _ = fetchMeals(ctx)
}()

wg.Wait()
```

### Swift - async letで並列実行

```swift
async let users = fetchUsers()
async let meals = fetchMeals()

let (fetchedUsers, fetchedMeals) = try await (users, meals)
```

## 早期リターン

深いネストは早期リターンで解消すること:

```go
// NG: 深いネスト
if user != nil {
    if user.IsValid {
        if meal != nil {
            // 処理
        }
    }
}

// OK: 早期リターン
if user == nil {
    return nil, ErrUserNotFound
}
if !user.IsValid {
    return nil, ErrInvalidUser
}
if meal == nil {
    return nil, ErrMealNotFound
}
// 処理
```

## 禁止事項

- Goでのエラー無視（`_ = err`）
- Swiftでのforce unwrap（`!`の濫用）
- 未使用のimport/変数の放置
- コメントアウトされたコードの放置
- マジックナンバー/マジックストリングの直接使用
- 環境変数の直接参照（設定ファイル経由で取得）
