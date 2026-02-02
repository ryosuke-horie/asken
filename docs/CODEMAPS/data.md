# データモデルとスキーマ

最終更新: 2026-02-03
データベース: PostgreSQL
認証: Firebase Authentication（ユーザーIDはFirebase UID）

## ER図

```
users
  │
  ├──< analysis_requests ──< analysis_results
  │
  ├──< weight_records
  │
  ├──< weight_goals
  │
  ├──< mylist_items
  │
  ├──< condition_records
  │
  ├──< training_locations ──< training_equipment
  │         │
  │         └──< training_records ──< training_menus
  │
  └──< user_profiles
```

## テーブル定義

### users

ユーザーマスタ（Firebase Authentication連携）

| カラム | 型 | 説明 |
|:---|:---|:---|
| id | VARCHAR(128) | 主キー（Firebase UID） |
| email | VARCHAR(255) | メールアドレス |
| name | VARCHAR(255) | 表示名 |
| created_at | TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | 更新日時 |

**注意**: 認証はFirebase Authenticationで管理。パスワードはFirebase側で管理されるため、DBには保存しない。

### analysis_requests

食事分析リクエスト

| カラム | 型 | 説明 |
|:---|:---|:---|
| id | UUID | 主キー |
| user_id | UUID | ユーザーID (FK) |
| status | VARCHAR(20) | ステータス (pending/processing/completed/failed) |
| input_type | VARCHAR(20) | 入力タイプ (image/text/mylist/skipped) |
| image_path | VARCHAR(500) | 画像パス |
| input_text | TEXT | テキスト入力 |
| meal_type | VARCHAR(20) | 食事タイプ (breakfast/lunch/dinner/snack) |
| meal_date | DATE | 食事日 |
| error_message | TEXT | エラーメッセージ |
| created_at | TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | 更新日時 |

### analysis_results

食事分析結果

| カラム | 型 | 説明 |
|:---|:---|:---|
| id | UUID | 主キー |
| analysis_request_id | UUID | 分析リクエストID (FK, UNIQUE) |
| foods | JSONB | 食品リスト |
| total_calories | DECIMAL(10,2) | 総カロリー |
| total_protein | DECIMAL(10,2) | 総タンパク質 (g) |
| total_fat | DECIMAL(10,2) | 総脂質 (g) |
| total_carbohydrates | DECIMAL(10,2) | 総炭水化物 (g) |
| created_at | TIMESTAMP | 作成日時 |

### weight_records

体重記録

| カラム | 型 | 説明 |
|:---|:---|:---|
| id | UUID | 主キー |
| user_id | UUID | ユーザーID (FK) |
| weight | DECIMAL(5,1) | 体重 (kg) |
| recorded_at | DATE | 記録日 |
| created_at | TIMESTAMP | 作成日時 |

制約: UNIQUE(user_id, recorded_at) - 1ユーザー1日1記録

### weight_goals

目標体重

| カラム | 型 | 説明 |
|:---|:---|:---|
| id | UUID | 主キー |
| user_id | UUID | ユーザーID (FK, UNIQUE) |
| target_weight | DECIMAL(5,1) | 目標体重 (kg) |
| target_date | DATE | 目標日 |
| created_at | TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | 更新日時 |

### mylist_items

マイリスト（よく食べるもの）

| カラム | 型 | 説明 |
|:---|:---|:---|
| id | UUID | 主キー |
| user_id | UUID | ユーザーID (FK) |
| name | VARCHAR(100) | アイテム名 |
| base_amount | VARCHAR(50) | 基本量 |
| unit | VARCHAR(20) | 単位 |
| image_path | VARCHAR(500) | 画像パス |
| foods | JSONB | 食品リスト |
| total_calories | DECIMAL(10,2) | 総カロリー |
| total_protein | DECIMAL(10,2) | 総タンパク質 |
| total_fat | DECIMAL(10,2) | 総脂質 |
| total_carbohydrates | DECIMAL(10,2) | 総炭水化物 |
| sort_order | INTEGER | 表示順 |
| created_at | TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | 更新日時 |

### condition_records

体調記録

| カラム | 型 | 説明 |
|:---|:---|:---|
| id | UUID | 主キー |
| user_id | UUID | ユーザーID (FK) |
| recorded_at | DATE | 記録日 |
| fatigue_level | INTEGER | 疲労度 (1-5) |
| stress_level | INTEGER | ストレス度 (1-5) |
| sleep_quality | INTEGER | 睡眠の質 (1-5) |
| motivation | INTEGER | モチベーション (1-5) |
| note | TEXT | メモ |
| created_at | TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | 更新日時 |

制約: UNIQUE(user_id, recorded_at)

### training_locations

トレーニング場所

| カラム | 型 | 説明 |
|:---|:---|:---|
| id | UUID | 主キー |
| user_id | UUID | ユーザーID (FK) |
| name | VARCHAR(100) | 場所名 |
| sort_order | INTEGER | 表示順 |
| created_at | TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | 更新日時 |

制約: UNIQUE(user_id, name)

### training_equipment

トレーニング器具

| カラム | 型 | 説明 |
|:---|:---|:---|
| id | UUID | 主キー |
| location_id | UUID | 場所ID (FK) |
| name | VARCHAR(100) | 器具名（正規化後） |
| original_name | VARCHAR(200) | 元の器具名 |
| sort_order | INTEGER | 表示順 |
| created_at | TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | 更新日時 |

制約: UNIQUE(location_id, name)

### training_records

トレーニング記録

| カラム | 型 | 説明 |
|:---|:---|:---|
| id | UUID | 主キー |
| user_id | UUID | ユーザーID (FK) |
| location_id | UUID | 場所ID (FK) |
| recorded_at | DATE | 記録日 |
| completed | BOOLEAN | 完了フラグ |
| created_at | TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | 更新日時 |

制約: UNIQUE(user_id, recorded_at) - 1ユーザー1日1記録

### training_menus

トレーニングメニュー

| カラム | 型 | 説明 |
|:---|:---|:---|
| id | UUID | 主キー |
| training_record_id | UUID | 記録ID (FK) |
| equipment_id | UUID | 器具ID (FK) |
| sets | INTEGER | セット数 |
| reps | INTEGER | 回数 |
| weight_kg | DECIMAL(5,1) | 重量 (kg) |
| duration_minutes | INTEGER | 時間（分） |
| note | TEXT | メモ |
| sort_order | INTEGER | 表示順 |
| created_at | TIMESTAMP | 作成日時 |

### user_profiles

ユーザープロフィール

| カラム | 型 | 説明 |
|:---|:---|:---|
| id | UUID | 主キー |
| user_id | UUID | ユーザーID (FK, UNIQUE) |
| height_cm | DECIMAL(5,1) | 身長 (cm) |
| activity_level | VARCHAR(20) | 活動レベル |
| training_frequency | VARCHAR(50) | トレーニング頻度 |
| created_at | TIMESTAMP | 作成日時 |
| updated_at | TIMESTAMP | 更新日時 |

## JSONB構造

### foods (analysis_results.foods, mylist_items.foods)

```json
[
  {
    "name": "食品名",
    "estimated_amount": "推定量",
    "calories": 100.0,
    "protein": 10.0,
    "fat": 5.0,
    "carbohydrates": 15.0
  }
]
```

## インデックス

| テーブル | インデックス | 用途 |
|:---|:---|:---|
| analysis_requests | status | ワーカーのポーリング |
| analysis_requests | created_at | 古いリクエスト優先 |
| analysis_requests | user_id | ユーザー別検索 |
| users | email | ログイン時検索 |
| weight_records | user_id, recorded_at | ユーザー別日付検索 |
| training_locations | user_id, sort_order | ユーザー別並び順 |
| training_equipment | location_id, sort_order | 場所別並び順 |
| training_records | user_id, recorded_at | ユーザー別日付検索 |

## 共通パターン

### 自動更新トリガー

updated_atカラムを持つテーブルには自動更新トリガーが設定:

```sql
CREATE TRIGGER update_{table}_updated_at
    BEFORE UPDATE ON {table}
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

### UUID主キー

すべてのテーブルでUUIDを主キーとして使用:

```sql
id UUID PRIMARY KEY DEFAULT gen_random_uuid()
```

## 関連コードマップ

- [全体アーキテクチャ](./architecture.md)
- [バックエンド構造](./backend.md)
