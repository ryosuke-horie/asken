-- 000014_create_user_profiles.down.sql
-- ユーザープロフィール管理機能用テーブルの削除

-- テーブルを削除（インデックスは自動削除される）
DROP TABLE IF EXISTS user_profiles;
