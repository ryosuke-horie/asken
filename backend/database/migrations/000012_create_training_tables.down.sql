-- 000011_create_training_tables.down.sql
-- トレーニング管理機能用テーブルの削除

DROP TRIGGER IF EXISTS update_training_records_updated_at ON training_records;
DROP INDEX IF EXISTS idx_training_records_user_date;
DROP TABLE IF EXISTS training_records;

DROP TRIGGER IF EXISTS update_training_equipment_updated_at ON training_equipment;
DROP INDEX IF EXISTS idx_training_equipment_location;
DROP TABLE IF EXISTS training_equipment;

DROP TRIGGER IF EXISTS update_training_locations_updated_at ON training_locations;
DROP INDEX IF EXISTS idx_training_locations_user;
DROP TABLE IF EXISTS training_locations;
