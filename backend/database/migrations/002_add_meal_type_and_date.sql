-- 002_add_meal_type_and_date.sql
-- meal_type と meal_date カラムを追加

-- meal_type と meal_date カラムを追加
ALTER TABLE analysis_requests
ADD COLUMN meal_type VARCHAR(20) CHECK (meal_type IN ('breakfast', 'lunch', 'dinner', 'snack')),
ADD COLUMN meal_date DATE;

-- 複合インデックス作成（日付・食事タイプ・ステータスでの効率的な検索）
CREATE INDEX idx_analysis_requests_meal_date_type ON analysis_requests(meal_date, meal_type, status);

-- 日付単独のインデックス作成
CREATE INDEX idx_analysis_requests_meal_date ON analysis_requests(meal_date);
