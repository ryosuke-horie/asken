-- 000014_create_user_profiles.up.sql
-- ユーザープロフィール管理機能用テーブル

-- user_profiles テーブル: ユーザーのトレーニングプロフィールを管理
-- sport_type: 競技種別（柔術、キックボクシング、MMA、ボクシング、レスリング等）
-- training_goals: トレーニング目標（複数選択可）
-- weight_class: 体重階級（kg）
CREATE TABLE user_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    sport_type VARCHAR(50),
    training_goals TEXT[],
    weight_class INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ユーザーIDによる検索用インデックス
CREATE INDEX idx_user_profiles_user_id ON user_profiles(user_id);
