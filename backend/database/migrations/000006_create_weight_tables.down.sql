-- 000006_create_weight_tables.down.sql
-- 体重管理機能用テーブルの削除

-- トリガーを削除
DROP TRIGGER IF EXISTS update_weight_goals_updated_at ON weight_goals;

-- テーブルを削除（インデックスは自動削除される）
DROP TABLE IF EXISTS weight_goals;
DROP TABLE IF EXISTS weight_records;
