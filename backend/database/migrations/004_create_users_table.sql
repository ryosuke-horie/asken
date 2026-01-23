-- 004_create_users_table.sql
-- ユーザーテーブルとanalysis_requestsへのuser_id追加

-- usersテーブル: メールアドレスでユーザーを識別
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- email検索用インデックス
CREATE INDEX idx_users_email ON users(email);

-- usersのupdated_at自動更新トリガー
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- analysis_requestsテーブルにuser_idカラムを追加
ALTER TABLE analysis_requests
ADD COLUMN user_id UUID REFERENCES users(id);

-- user_id検索用インデックス
CREATE INDEX idx_analysis_requests_user_id ON analysis_requests(user_id);
