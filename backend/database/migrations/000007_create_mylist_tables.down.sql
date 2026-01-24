-- 000007_create_mylist_tables.down.sql
-- マイリスト機能用テーブルの削除

-- トリガーを削除
DROP TRIGGER IF EXISTS update_mylist_items_updated_at ON mylist_items;

-- テーブルを削除（インデックスは自動削除される）
DROP TABLE IF EXISTS mylist_items;
