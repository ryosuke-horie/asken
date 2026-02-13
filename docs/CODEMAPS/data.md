# データモデルとスキーマ

最終更新: 2026-02-13
データベース: Firestore
認証: Firebase Authentication（ユーザーIDはFirebase UID）

## コレクション構造

```
users/{userId}/analysisRequests/{requestId}
users/{userId}/weightRecords/{recordId}
users/{userId}/weightGoal/current
users/{userId}/myMenu/{menuId}
```

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
| confirmed | boolean | ユーザーが保存確定したか（trueのみ一覧に表示） |
| createdAt | timestamp | 作成日時 |
| updatedAt | timestamp | 更新日時 |
| result | map | 分析結果（下記参照） |

confirmedフィールドの動作:
- 分析開始時: `confirmed: false`（一覧に表示されない）
- ユーザーが「保存」: `confirmed: true`（一覧に表示される）
- マイリスト/スキップ記録: 即座に`confirmed: true`

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

### weightRecords/{recordId}

体重記録

| フィールド | 型 | 説明 |
|:---|:---|:---|
| id | string | レコードID |
| weightKg | number | 体重 (kg, 20.0-300.0, 小数点1桁) |
| recordedAt | timestamp | 記録日時 |
| note | string | メモ |
| createdAt | timestamp | 作成日時 |
| updatedAt | timestamp | 更新日時 |

バリデーション:
- 体重範囲: 20.0 - 300.0 kg
- 小数点1桁に丸め

### weightGoal/current

目標体重（ユーザーごとに1ドキュメント）

| フィールド | 型 | 説明 |
|:---|:---|:---|
| targetWeightKg | number | 目標体重 (kg) |
| updatedAt | timestamp | 更新日時 |

### myMenu/{menuId}

マイメニュー（よく食べるメニューの登録）

| フィールド | 型 | 説明 |
|:---|:---|:---|
| id | string | メニューID |
| name | string | メニュー名 |
| foods | array | 食品リスト（NutritionInfo配列） |
| totalCalories | number | 総カロリー |
| totalProtein | number | 総タンパク質 (g) |
| totalFat | number | 総脂質 (g) |
| totalCarbohydrates | number | 総炭水化物 (g) |
| createdAt | timestamp | 作成日時 |
| updatedAt | timestamp | 更新日時 |

## インデックス

Firestoreの複合インデックスは`firestore.indexes.json`で管理されています。

| コレクション | フィールド | 用途 |
|:---|:---|:---|
| analysisRequests | status, createdAt | ワーカーのポーリング |
| analysisRequests | status, confirmed, createdAt DESC | 履歴一覧（confirmed=true のみ） |
| analysisRequests | confirmed, status, mealDate | 日次食事取得（confirmed=true のみ） |
| analysisRequests | mealType, confirmed, mealDate | 未確定レコード削除用 |
| analysisRequests | inputType, mealType, mealDate | 既存マイリスト/スキップ検索 |
| analysisRequests | mealType, mealDate, inputType | スキップ記録削除用 |
| analysisRequests | status, mealDate | ステータス別日次検索 |
| weightRecords | recordedAt | 期間別体重記録取得 |

インデックス更新手順は[docs/CONTRIB.md](../CONTRIB.md#firestoreインデックス管理)を参照。

## 関連コードマップ

- [全体アーキテクチャ](./architecture.md)
- [バックエンド構造](./backend.md)
