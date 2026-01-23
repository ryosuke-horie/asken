-- 006_create_weight_tables.sql
-- 体重管理機能用テーブル

-- weight_records テーブル: 体重記録を管理
-- 1ユーザー1日1記録の制約を持つ
CREATE TABLE weight_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    weight DECIMAL(5,1) NOT NULL,
    recorded_at DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, recorded_at)
);

-- ユーザーと日付による検索用インデックス
CREATE INDEX idx_weight_records_user_date ON weight_records(user_id, recorded_at DESC);

-- weight_goals テーブル: 目標体重を管理
-- 1ユーザー1目標の制約を持つ（UPSERT方式）
CREATE TABLE weight_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    target_weight DECIMAL(5,1) NOT NULL,
    target_date DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- weight_goalsのupdated_at自動更新トリガー
CREATE TRIGGER update_weight_goals_updated_at
    BEFORE UPDATE ON weight_goals
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
