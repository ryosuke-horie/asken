-- 000010_create_condition_tables.up.sql
-- 体調・疲労度管理機能用テーブル

-- condition_records テーブル: 体調・疲労度記録を管理
-- 1ユーザー1日1記録の制約を持つ
-- condition: 1=悪い, 2=普通, 3=良い
-- fatigue: 1=低い, 2=普通, 3=高い
CREATE TABLE condition_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    condition INTEGER NOT NULL CHECK (condition BETWEEN 1 AND 3),
    fatigue INTEGER NOT NULL CHECK (fatigue BETWEEN 1 AND 3),
    recorded_at DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, recorded_at)
);

-- ユーザーと日付による検索用インデックス
CREATE INDEX idx_condition_records_user_date ON condition_records(user_id, recorded_at DESC);
