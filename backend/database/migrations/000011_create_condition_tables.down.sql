-- 000010_create_condition_tables.down.sql
-- 体調・疲労度管理機能用テーブルの削除

-- テーブルを削除（インデックスは自動削除される）
DROP TABLE IF EXISTS condition_records;
