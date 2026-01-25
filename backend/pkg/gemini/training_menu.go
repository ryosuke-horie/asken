package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// MenuItem はトレーニングメニューの1項目を表す
type MenuItem struct {
	Name        string `json:"name"`
	Duration    int    `json:"duration"`    // 分
	Sets        int    `json:"sets"`        // セット数
	Reps        string `json:"reps"`        // 回数（例: "10回", "30秒"）
	Equipment   string `json:"equipment"`   // 使用器具
	Description string `json:"description"` // 説明・ポイント
}

// MenuSuggester はトレーニングメニューを提案するクライアント
type MenuSuggester struct {
	client *Client
}

// NewMenuSuggester は新しいMenuSuggesterを作成
func NewMenuSuggester(timeout time.Duration) *MenuSuggester {
	return &MenuSuggester{
		client: NewClient(timeout),
	}
}

// SuggestMenuParams はメニュー提案のパラメータ
type SuggestMenuParams struct {
	Equipment       []string
	DurationMinutes int
	Goals           []string
	Fatigue         *int    // 疲労度 (1-3): 1=低い, 2=普通, 3=高い
	Condition       *int    // 体調 (1-3): 1=悪い, 2=普通, 3=良い
	SportType       *string // 競技種別（柔術、キックボクシング等）
}

// SuggestMenu は利用可能な器具と時間に基づいてトレーニングメニューを提案
func (ms *MenuSuggester) SuggestMenu(ctx context.Context, params SuggestMenuParams) ([]MenuItem, error) {
	if len(params.Equipment) == 0 {
		return nil, fmt.Errorf("利用可能な器具が指定されていません")
	}
	if params.DurationMinutes <= 0 {
		return nil, fmt.Errorf("トレーニング時間は0より大きい値が必要です")
	}
	if params.DurationMinutes > 240 {
		return nil, fmt.Errorf("トレーニング時間は240分（4時間）以下で指定してください")
	}

	equipmentList := strings.Join(params.Equipment, "、")

	// オプショナルなパラメータを構築
	var optionalParams []string

	if len(params.Goals) > 0 {
		optionalParams = append(optionalParams, fmt.Sprintf("目標・重点: %s", strings.Join(params.Goals, "、")))
	}

	if params.Fatigue != nil {
		fatigueLabels := map[int]string{1: "低い", 2: "普通", 3: "高い"}
		if label, ok := fatigueLabels[*params.Fatigue]; ok {
			optionalParams = append(optionalParams, fmt.Sprintf("本日の疲労度: %s", label))
		}
	}

	if params.Condition != nil {
		conditionLabels := map[int]string{1: "悪い", 2: "普通", 3: "良い"}
		if label, ok := conditionLabels[*params.Condition]; ok {
			optionalParams = append(optionalParams, fmt.Sprintf("本日の体調: %s", label))
		}
	}

	sportTypeText := "格闘技（柔術・キックボクシング）"
	if params.SportType != nil && *params.SportType != "" {
		sportTypeText = *params.SportType
	}

	optionalText := ""
	if len(optionalParams) > 0 {
		optionalText = "\n" + strings.Join(optionalParams, "\n")
	}

	prompt := fmt.Sprintf(`以下の条件でトレーニングメニューを提案してください。

利用可能な器具: %s
トレーニング時間: %d分%s

出力フォーマット（JSON配列）:
[
  {
    "name": "エクササイズ名",
    "duration": 5,
    "sets": 3,
    "reps": "10回",
    "equipment": "使用器具",
    "description": "実施のポイント"
  }
]

重要なルール:
- 指定された器具のみを使用してください
- 合計時間が指定時間に収まるようにしてください
- ウォームアップとクールダウンも含めてください
- %s向けのメニューを優先してください
- 疲労度が高い場合は回復系のメニューを多めに、低い場合は強度を上げてください
- 体調が悪い場合は軽めのメニューにしてください
- 各エクササイズには具体的なセット数と回数を指定してください
- JSON形式で出力してください`, equipmentList, params.DurationMinutes, optionalText, sportTypeText)

	response, err := ms.client.Execute(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("Gemini APIによるメニュー提案に失敗: %w", err)
	}

	menuJSON := removeCodeBlock(response.Response)

	var menu []MenuItem
	if err := json.Unmarshal([]byte(menuJSON), &menu); err != nil {
		return nil, fmt.Errorf("メニューのパースエラー: %w\nデータ: %s", err, menuJSON)
	}

	return menu, nil
}
