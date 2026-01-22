-- 003_add_text_input.sql
-- テキスト入力対応のためのスキーマ変更

-- input_type カラムを追加（'image' または 'text'）
ALTER TABLE analysis_requests
ADD COLUMN input_type VARCHAR(10) NOT NULL DEFAULT 'image'
CHECK (input_type IN ('image', 'text'));

-- input_text カラムを追加（テキスト入力の場合に使用）
ALTER TABLE analysis_requests
ADD COLUMN input_text VARCHAR(1000);

-- image_path の NOT NULL 制約を解除
ALTER TABLE analysis_requests
ALTER COLUMN image_path DROP NOT NULL;

-- CHECK制約追加: input_typeに応じて必要なカラムが入力されていることを検証
ALTER TABLE analysis_requests
ADD CONSTRAINT chk_input_required
CHECK (
    (input_type = 'image' AND image_path IS NOT NULL) OR
    (input_type = 'text' AND input_text IS NOT NULL)
);
