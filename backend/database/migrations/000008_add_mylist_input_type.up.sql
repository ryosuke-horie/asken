-- マイリストからの食事記録用にinput_type制約を更新

-- 既存の制約を削除
ALTER TABLE analysis_requests DROP CONSTRAINT IF EXISTS analysis_requests_input_type_check;
ALTER TABLE analysis_requests DROP CONSTRAINT IF EXISTS chk_input_required;

-- input_type制約を更新（mylistを追加）
ALTER TABLE analysis_requests ADD CONSTRAINT analysis_requests_input_type_check
CHECK (input_type IN ('image', 'text', 'mylist'));

-- 入力必須制約を更新（mylistの場合はinput_textを使用）
ALTER TABLE analysis_requests ADD CONSTRAINT chk_input_required CHECK (
  (input_type = 'image' AND image_path IS NOT NULL) OR
  (input_type = 'text' AND input_text IS NOT NULL) OR
  (input_type = 'mylist' AND input_text IS NOT NULL)
);
