-- 000007_create_mylist_tables.up.sql
-- マイリスト機能用テーブル

-- mylist_items テーブル: よく食べるメニューを管理
CREATE TABLE mylist_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    base_amount VARCHAR(50) NOT NULL,
    unit VARCHAR(50) NOT NULL,
    calories DECIMAL(10, 2) NOT NULL,
    protein DECIMAL(10, 2) NOT NULL,
    fat DECIMAL(10, 2) NOT NULL,
    carbohydrates DECIMAL(10, 2) NOT NULL,
    foods JSONB NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ユーザーと並び順による検索用インデックス
CREATE INDEX idx_mylist_items_user_sort ON mylist_items(user_id, sort_order ASC);

-- mylist_itemsのupdated_at自動更新トリガー
CREATE TRIGGER update_mylist_items_updated_at
    BEFORE UPDATE ON mylist_items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
