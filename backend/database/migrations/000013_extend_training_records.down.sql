-- 000013_extend_training_records.down.sql
-- トレーニング記録機能拡張のロールバック

-- 1. training_record_menus 中間テーブルを削除
DROP INDEX IF EXISTS idx_training_record_menus_menu;
DROP INDEX IF EXISTS idx_training_record_menus_record;
DROP TABLE IF EXISTS training_record_menus;

-- 2. training_menus テーブルを削除
DROP TRIGGER IF EXISTS update_training_menus_updated_at ON training_menus;
DROP INDEX IF EXISTS idx_training_menus_user;
DROP TABLE IF EXISTS training_menus;

-- 3. training_records のUNIQUE制約を復元
ALTER TABLE training_records
ADD CONSTRAINT training_records_user_id_recorded_at_key UNIQUE (user_id, recorded_at);

-- 4. training_records の新カラムを削除
ALTER TABLE training_records
DROP COLUMN IF EXISTS notes,
DROP COLUMN IF EXISTS satisfaction,
DROP COLUMN IF EXISTS intensity,
DROP COLUMN IF EXISTS duration;
