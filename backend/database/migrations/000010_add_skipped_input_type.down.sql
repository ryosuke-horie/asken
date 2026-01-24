-- skippedを削除して元に戻す
ALTER TABLE analysis_requests DROP CONSTRAINT IF EXISTS analysis_requests_input_type_check;
ALTER TABLE analysis_requests DROP CONSTRAINT IF EXISTS chk_input_required;

ALTER TABLE analysis_requests ADD CONSTRAINT analysis_requests_input_type_check
CHECK (input_type IN ('image', 'text', 'mylist'));

ALTER TABLE analysis_requests ADD CONSTRAINT chk_input_required CHECK (
  (input_type = 'image' AND image_path IS NOT NULL) OR
  (input_type = 'text' AND input_text IS NOT NULL) OR
  (input_type = 'mylist' AND input_text IS NOT NULL)
);
