-- 000013_extend_training_records.up.sql
-- トレーニング記録機能の拡張

-- 1. training_records テーブルに新カラム追加
ALTER TABLE training_records
ADD COLUMN duration INTEGER,
ADD COLUMN intensity INTEGER,
ADD COLUMN satisfaction INTEGER,
ADD COLUMN notes TEXT;

-- 2. UNIQUE制約を削除（1日複数記録対応）
ALTER TABLE training_records
DROP CONSTRAINT training_records_user_id_recorded_at_key;

-- 3. インデックスを追加（既存の idx_training_records_user_date と重複しないよう確認）
-- 既存のインデックスで十分なので追加は不要

-- 4. training_menus テーブル新規作成
CREATE TABLE training_menus (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT false,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- メニュー検索用インデックス
CREATE INDEX idx_training_menus_user ON training_menus(user_id, sort_order);

-- training_menus の updated_at 自動更新トリガー
CREATE TRIGGER update_training_menus_updated_at
    BEFORE UPDATE ON training_menus
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 5. 固定メニューをINSERT（user_id = NULL は固定メニュー）
INSERT INTO training_menus (user_id, name, is_default, sort_order) VALUES
(NULL, 'スパーリング', true, 1),
(NULL, '打ち込み・ミット', true, 2),
(NULL, '技練習（ドリル）', true, 3),
(NULL, '対人練習', true, 4);

-- 6. training_record_menus 中間テーブル新規作成
CREATE TABLE training_record_menus (
    record_id UUID NOT NULL REFERENCES training_records(id) ON DELETE CASCADE,
    menu_id UUID NOT NULL REFERENCES training_menus(id) ON DELETE CASCADE,
    PRIMARY KEY (record_id, menu_id)
);

-- 中間テーブルのインデックス
CREATE INDEX idx_training_record_menus_record ON training_record_menus(record_id);
CREATE INDEX idx_training_record_menus_menu ON training_record_menus(menu_id);
