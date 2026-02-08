---
paths:
  - "backend/internal/repository/**/*.go"
  - "firestore.indexes.json"
  - "firebase.json"
---

# Firestore規約

## 複合インデックスの管理

Firestoreでは複数フィールドを使用するクエリには複合インデックスが必要。インデックスがないとクエリ実行時にエラーとなる。

### インデックスが必要になるケース

以下のようなクエリを追加・変更する場合、インデックスの更新が必要:

```go
// 複数フィールドでのWhere + OrderBy
collection.
    Where("status", "==", "completed").
    Where("confirmed", "==", true).
    OrderBy("createdAt", firestore.Desc)

// 範囲クエリと等価クエリの組み合わせ
collection.
    Where("mealDate", ">=", startOfDay).
    Where("mealDate", "<", endOfDay).
    Where("status", "==", "completed")
```

### インデックス更新手順

リポジトリ層でFirestoreクエリを追加・変更した場合、必ず以下を実行:

1. ローカルで動作確認を実行
2. インデックスエラーが発生した場合、エラーメッセージのURLを確認
3. `firestore.indexes.json`にインデックス定義を追加
4. インデックスをデプロイ

```bash
# インデックスのデプロイ
firebase deploy --only firestore:indexes --project utikomi-dev

# 現在のインデックス一覧を取得（firestore.indexes.jsonの更新用）
firebase firestore:indexes --project utikomi-dev
```

### インデックス定義のルール

Firestoreの複合インデックスでは、フィールドの順序が重要:

| フィルタ種別 | 順序 |
|:---|:---|
| 等価フィルタ（`==`） | 先に配置 |
| 範囲フィルタ（`>=`, `<`, `>`, `<=`） | 後に配置 |
| OrderByフィールド | 最後に配置 |

```json
// 正しい順序の例
{
  "collectionGroup": "analysisRequests",
  "queryScope": "COLLECTION",
  "fields": [
    { "fieldPath": "confirmed", "order": "ASCENDING" },
    { "fieldPath": "status", "order": "ASCENDING" },
    { "fieldPath": "mealDate", "order": "ASCENDING" }
  ]
}
```

### チェックリスト

リポジトリ層のFirestoreクエリを変更した場合:

- [ ] ローカル環境で動作確認を実施
- [ ] インデックスエラーが出た場合は`firestore.indexes.json`を更新
- [ ] `firebase deploy --only firestore:indexes`でデプロイ
- [ ] インデックス構築完了後に再度動作確認

## ドキュメント構造

### コレクションパス

```
users/{userID}/analysisRequests/{requestID}
```

### フィールド命名規則

| 命名規則 | 例 |
|:---|:---|
| キャメルケース | `mealType`, `createdAt`, `totalCalories` |
| bool型は肯定形 | `confirmed`, `isActive`（`notConfirmed`は避ける） |
