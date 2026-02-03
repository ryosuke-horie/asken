# データモデルとスキーマ

最終更新: 2026-02-03
データベース: Firestore
認証: Firebase Authentication（ユーザーIDはFirebase UID）

## コレクション構造

```
users/{userId}/analysisRequests/{requestId}
```

**注意**: 体重、体調、トレーニング、マイリスト、プロフィールのコレクションは未実装です。ADR-002に設計仕様があります。

## ドキュメント定義

### analysisRequests/{requestId}

食事分析リクエストと結果（1ドキュメントに統合）

| フィールド | 型 | 説明 |
|:---|:---|:---|
| status | string | ステータス (pending/processing/completed/failed) |
| inputType | string | 入力タイプ (image/text/mylist/skipped) |
| imagePath | string | 画像パス |
| inputText | string | テキスト入力 |
| mealType | string | 食事タイプ (breakfast/lunch/dinner/snack) |
| mealDate | timestamp | 食事日 |
| errorMessage | string | エラーメッセージ |
| createdAt | timestamp | 作成日時 |
| updatedAt | timestamp | 更新日時 |
| result | map | 分析結果（下記参照） |

### result フィールド構造

| フィールド | 型 | 説明 |
|:---|:---|:---|
| foods | array | 食品リスト |
| totalCalories | number | 総カロリー |
| totalProtein | number | 総タンパク質 (g) |
| totalFat | number | 総脂質 (g) |
| totalCarbohydrates | number | 総炭水化物 (g) |

### foods 配列要素

```json
{
  "name": "食品名",
  "estimated_amount": "推定量",
  "calories": 100.0,
  "protein": 10.0,
  "fat": 5.0,
  "carbohydrates": 15.0
}
```

## 将来のコレクション設計（ADR-002より）

```
users/
  └── {userId}/
        ├── weightGoal/  (サブコレクション、1ドキュメント)
        │     └── goal/
        │           ├── targetWeight: number
        │           └── targetDate: timestamp
        │
        ├── weightRecords/  (サブコレクション)
        │     └── {recordId}/
        │           ├── weight: number
        │           └── recordedAt: timestamp
        │
        ├── conditionRecords/  (サブコレクション)
        │     └── {recordId}/
        │           ├── condition: number (1-3)
        │           ├── fatigue: number (1-3)
        │           └── recordedAt: timestamp
        │
        ├── trainingLocations/  (サブコレクション)
        │     └── {locationId}/
        │           ├── name: string
        │           ├── sortOrder: number
        │           │
        │           └── equipment/  (サブコレクション)
        │                 └── {equipmentId}/
        │                       ├── name: string
        │                       └── sortOrder: number
        │
        ├── trainingRecords/  (サブコレクション)
        │     └── {recordId}/
        │           ├── locationId: string
        │           ├── recordedAt: timestamp
        │           └── completed: boolean
        │
        ├── analysisRequests/  (サブコレクション) ← 実装済み
        │     └── {requestId}/
        │           ├── status: string
        │           ├── imagePath: string
        │           ├── mealType: string
        │           ├── mealDate: timestamp
        │           ├── result: map (foods, totalCalories, etc.)
        │           └── createdAt: timestamp
        │
        └── mylist/  (サブコレクション)
              └── {itemId}/
                    ├── name: string
                    ├── foods: array
                    └── totalCalories: number
```

## インデックス

Firestoreの複合インデックスは自動作成されますが、以下のクエリパターンで使用されます:

| コレクション | クエリパターン | 用途 |
|:---|:---|:---|
| analysisRequests | status == "pending", orderBy createdAt | ワーカーのポーリング |
| analysisRequests | status == "completed", orderBy createdAt DESC | 履歴一覧 |
| analysisRequests | mealDate == date, status == "completed" | 日次食事取得 |

## 関連コードマップ

- [全体アーキテクチャ](./architecture.md)
- [バックエンド構造](./backend.md)
