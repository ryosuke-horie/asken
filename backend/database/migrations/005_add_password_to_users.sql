-- 005_add_password_to_users.sql
-- usersテーブルにpassword_hashカラムを追加

-- 開発環境のため、既存のデータを削除して再構築
-- 1. 外部キー制約のあるanalysis_requestsのuser_idをNULLに
UPDATE analysis_requests SET user_id = NULL;

-- 2. 既存ユーザーを削除
DELETE FROM users;

-- 3. password_hashカラムを追加
ALTER TABLE users
ADD COLUMN password_hash VARCHAR(255) NOT NULL;
