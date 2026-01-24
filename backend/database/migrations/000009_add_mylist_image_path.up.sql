-- マイリストアイテムに画像パスを追加
ALTER TABLE mylist_items ADD COLUMN IF NOT EXISTS image_path VARCHAR(255);
