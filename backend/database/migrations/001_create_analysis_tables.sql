-- 001_create_analysis_tables.sql
-- 非同期分析リクエストと結果を管理するテーブル

-- analysis_requests テーブル: 分析リクエストのステータスを管理
CREATE TABLE analysis_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    image_path VARCHAR(500) NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ステータス検索用インデックス（ワーカーのポーリングで使用）
CREATE INDEX idx_analysis_requests_status ON analysis_requests(status);

-- 作成日時検索用インデックス（古いリクエスト優先のため）
CREATE INDEX idx_analysis_requests_created_at ON analysis_requests(created_at);

-- analysis_results テーブル: 分析結果を永続化
CREATE TABLE analysis_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    analysis_request_id UUID NOT NULL UNIQUE REFERENCES analysis_requests(id) ON DELETE CASCADE,
    foods JSONB NOT NULL,
    total_calories DECIMAL(10, 2) NOT NULL,
    total_protein DECIMAL(10, 2) NOT NULL,
    total_fat DECIMAL(10, 2) NOT NULL,
    total_carbohydrates DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- リクエストIDによる結果検索用インデックス
CREATE INDEX idx_analysis_results_request_id ON analysis_results(analysis_request_id);

-- updated_at自動更新用トリガー関数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- analysis_requestsのupdated_at自動更新トリガー
CREATE TRIGGER update_analysis_requests_updated_at
    BEFORE UPDATE ON analysis_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
