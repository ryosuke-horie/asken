# 体重管理機能 設計書

## 概要

格闘技の減量・体重コントロールを支援するための体重管理機能。
日々の体重を記録し、目標体重との差分をグラフで可視化する。

## 確定要件

### 記録仕様

| 項目 | 内容 |
|:---|:---|
| 記録頻度 | 1日1回（朝のみ） |
| 記録項目 | 体重のみ |
| 単位 | kg（小数点第1位まで） |
| 目標 | 1つのみ設定可能（目標体重 + 期限） |

### UI/UX

| 項目 | 内容 |
|:---|:---|
| 配置 | ダッシュボード（ホーム） |
| レイアウト | グラフ上 + 記録ボタン下 |
| グラフ期間 | 1週間 / 1ヶ月 / 3ヶ月の切り替え |
| グラフライブラリ | shadcn/ui（Recharts） |

### MVP対象外

- 記録の編集・削除
- 複数目標の管理
- 目標履歴の保存

---

## データベース設計

### weight_records テーブル

体重の記録を保存するテーブル。1ユーザー1日1記録の制約を持つ。

```sql
CREATE TABLE weight_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    weight DECIMAL(5,1) NOT NULL,
    recorded_at DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, recorded_at)
);

CREATE INDEX idx_weight_records_user_date ON weight_records(user_id, recorded_at DESC);
```

### weight_goals テーブル

目標体重を保存するテーブル。1ユーザー1目標の制約を持つ（UPSERT方式）。

```sql
CREATE TABLE weight_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    target_weight DECIMAL(5,1) NOT NULL,
    target_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

---

## API設計

### 体重記録

#### POST /api/weight-records

体重を記録する。同日の記録がある場合は上書き（UPSERT）。

**リクエスト:**
```json
{
  "weight": 70.5,
  "recorded_at": "2026-01-24"
}
```

**レスポンス:**
```json
{
  "id": "uuid",
  "weight": 70.5,
  "recorded_at": "2026-01-24",
  "created_at": "2026-01-24T07:00:00Z"
}
```

#### GET /api/weight-records?period=week|month|3months

指定期間の体重記録を取得する。

**レスポンス:**
```json
{
  "records": [
    {
      "id": "uuid",
      "weight": 70.5,
      "recorded_at": "2026-01-24"
    }
  ],
  "latest": {
    "weight": 70.5,
    "recorded_at": "2026-01-24"
  },
  "stats": {
    "min": 69.0,
    "max": 72.0,
    "average": 70.3
  }
}
```

### 目標体重

#### GET /api/weight-goal

現在の目標を取得する。

**レスポンス:**
```json
{
  "target_weight": 66.0,
  "target_date": "2026-03-15",
  "days_remaining": 50,
  "weight_to_lose": 4.5
}
```

#### PUT /api/weight-goal

目標を設定・更新する（UPSERT）。

**リクエスト:**
```json
{
  "target_weight": 66.0,
  "target_date": "2026-03-15"
}
```

---

## フロントエンド設計

### コンポーネント構成

```
components/client/
├── WeightSection.tsx        # ダッシュボード用ラッパー
├── WeightChart.tsx          # 体重推移グラフ
├── WeightRecordForm.tsx     # 体重記録フォーム
└── WeightGoalSetting.tsx    # 目標設定UI
```

### WeightSection（ダッシュボード統合）

```
┌─────────────────────────────────────┐
│ 体重管理                            │
├─────────────────────────────────────┤
│ 現在: 70.5kg  目標: 66.0kg (-4.5kg) │
│ 目標日: 2026/03/15 (残り50日)       │
├─────────────────────────────────────┤
│ [グラフ]                            │
│ [1週間] [1ヶ月] [3ヶ月]             │
├─────────────────────────────────────┤
│ [体重を記録する]                    │
└─────────────────────────────────────┘
```

### 使用ライブラリ

- **グラフ**: shadcn/ui chart (Recharts)
- **日付ピッカー**: shadcn/ui date-picker
- **フォーム**: shadcn/ui input, button

---

## 実装順序

1. DBマイグレーション
2. Backend: Repository層
3. Backend: Handler層 + ルーティング
4. Frontend: API連携（types, fetcher）
5. Frontend: WeightChart
6. Frontend: WeightRecordForm
7. Frontend: WeightGoalSetting
8. Frontend: WeightSection（統合）
9. ダッシュボードへの組み込み
