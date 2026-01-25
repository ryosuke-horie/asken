-- 000012_create_training_tables.up.sql
-- トレーニング管理機能用テーブル

-- training_locations テーブル: トレーニング場所マスタ
CREATE TABLE training_locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);

-- ユーザーによる検索用インデックス
CREATE INDEX idx_training_locations_user ON training_locations(user_id, sort_order);

-- training_locations の updated_at 自動更新トリガー
CREATE TRIGGER update_training_locations_updated_at
    BEFORE UPDATE ON training_locations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- training_equipment テーブル: 器具設定
CREATE TABLE training_equipment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id UUID NOT NULL REFERENCES training_locations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    original_name VARCHAR(200),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(location_id, name)
);

-- 場所による検索用インデックス
CREATE INDEX idx_training_equipment_location ON training_equipment(location_id, sort_order);

-- training_equipment の updated_at 自動更新トリガー
CREATE TRIGGER update_training_equipment_updated_at
    BEFORE UPDATE ON training_equipment
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- training_records テーブル: 練習記録
-- 1ユーザー1日1記録の制約を持つ
CREATE TABLE training_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    location_id UUID REFERENCES training_locations(id) ON DELETE SET NULL,
    recorded_at DATE NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, recorded_at)
);

-- ユーザーと日付による検索用インデックス
CREATE INDEX idx_training_records_user_date ON training_records(user_id, recorded_at DESC);

-- training_records の updated_at 自動更新トリガー
CREATE TRIGGER update_training_records_updated_at
    BEFORE UPDATE ON training_records
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
