---
paths:
  - "database/**/*"
  - "backend/internal/repository/**/*.go"
---

# データベース規約（PostgreSQL）

## テーブル設計

- **食品テーブル**（foods）: 食品名、カロリー、栄養素
- **食事記録テーブル**（meals）: 食事の履歴（MVP後に追加予定）
- **適切なインデックス**を作成（検索性能向上）

```sql
-- 食品テーブル
CREATE TABLE foods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    calories INT NOT NULL,
    protein DECIMAL(10, 2),    -- タンパク質（g）
    fat DECIMAL(10, 2),         -- 脂質（g）
    carbs DECIMAL(10, 2),       -- 炭水化物（g）
    source VARCHAR(50),         -- データソース（'db', 'gemini'）
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_foods_name ON foods(name);
```

## データベース操作

- **マイグレーション**は必ずバージョン管理する
- **トランザクション**は適切に使用
- **プレースホルダ**を使用（SQLインジェクション対策）

```go
// ✅ 良い例 - SQLインジェクション対策
query := "SELECT * FROM foods WHERE name LIKE $1"
rows, err := db.QueryContext(ctx, query, "%"+searchTerm+"%")

// ❌ 悪い例 - SQLインジェクション脆弱性
query := fmt.Sprintf("SELECT * FROM foods WHERE name LIKE '%%%s%%'", searchTerm)
rows, err := db.QueryContext(ctx, query)
```

## クエリ最適化

- 検索クエリには適切なインデックスを活用
- N+1問題を避ける（JOINまたはバッチクエリを使用）
- 必要なカラムのみをSELECT（SELECT *を避ける）

```go
// ✅ 良い例 - 必要なカラムのみ
query := "SELECT id, name, calories FROM foods WHERE name LIKE $1"

// ❌ 悪い例 - すべてのカラムを取得
query := "SELECT * FROM foods WHERE name LIKE $1"
```
