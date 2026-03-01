package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// ExerciseEstimator は運動種目の消費カロリーをGeminiで推定するクライアント
type ExerciseEstimator struct {
	client *Client
}

// NewExerciseEstimator は新しいExerciseEstimatorを作成
// 環境変数GEMINI_API_KEYからAPIキーを読み取る
func NewExerciseEstimator(timeout time.Duration) (*ExerciseEstimator, error) {
	client, err := NewClient(timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create exercise estimator: %w", err)
	}
	return &ExerciseEstimator{client: client}, nil
}

// NewExerciseEstimatorWithAPIKey はAPIキーを指定してExerciseEstimatorを作成
func NewExerciseEstimatorWithAPIKey(apiKey string, timeout time.Duration) (*ExerciseEstimator, error) {
	client, err := NewClientWithAPIKey(apiKey, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create exercise estimator: %w", err)
	}
	return &ExerciseEstimator{client: client}, nil
}

// NewExerciseEstimatorWithClient はClientを受け取るコンストラクタ（テスト用）
func NewExerciseEstimatorWithClient(client *Client) *ExerciseEstimator {
	return &ExerciseEstimator{client: client}
}

// exerciseEstimateResponse はGemini APIレスポンスのデシリアライズ用構造体
type exerciseEstimateResponse struct {
	BurnedCaloriesKcal float64 `json:"burned_calories_kcal"`
}

// exerciseEstimateSchema はGemini出力スキーマ（消費カロリーのみ）
func exerciseEstimateSchema() *Schema {
	return &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"burned_calories_kcal": {
				Type: SchemaTypeNumber,
			},
		},
		Required: []string{"burned_calories_kcal"},
	}
}

// EstimateCalories は種目名と時間から消費カロリーを推定する
func (e *ExerciseEstimator) EstimateCalories(ctx context.Context, exerciseName string, durationMinutes int) (float64, error) {
	if exerciseName == "" {
		return 0, fmt.Errorf("exerciseNameが必要です")
	}
	if durationMinutes <= 0 {
		return 0, fmt.Errorf("durationMinutesは1以上が必要です")
	}

	log.Printf("ExerciseEstimator: 消費カロリー推定を開始 (種目: %s, 時間: %d分)", exerciseName, durationMinutes)

	prompt := fmt.Sprintf(`あなたは運動科学の専門家です。
以下の運動について、体重70kgの成人男性が行った場合の消費カロリーをkcalで推定してください。

運動種目: %s
実施時間: %d分

推定ルール:
- MET値（代謝当量）に基づいて計算してください
- 計算式: 消費カロリー = MET × 体重(kg) × 時間(h) × 1.05
- 体重は70kgとして計算してください
- 小数点以下は切り捨てず、1kcal単位で返してください
- 種目名が不明確な場合は最も近い種目のMET値を使用してください`, exerciseName, durationMinutes)

	schema := exerciseEstimateSchema()

	response, err := e.client.Execute(ctx, prompt, schema)
	if err != nil {
		log.Printf("ExerciseEstimator: Gemini API呼び出しエラー: %v", err)
		return 0, fmt.Errorf("Gemini APIコールエラー: %w", err)
	}

	resultJSON := removeCodeBlock(response.Response)

	var result exerciseEstimateResponse
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		log.Printf("ExerciseEstimator: レスポンスのJSONパースエラー: %v, データ: %s", err, resultJSON)
		return 0, fmt.Errorf("レスポンスのパースエラー: %w", err)
	}

	if result.BurnedCaloriesKcal <= 0 {
		return 0, fmt.Errorf("消費カロリーの推定に失敗しました: 種目=%s", exerciseName)
	}

	log.Printf("ExerciseEstimator: 推定完了 (種目: %s, %d分 → %.1fkcal)", exerciseName, durationMinutes, result.BurnedCaloriesKcal)
	return result.BurnedCaloriesKcal, nil
}
