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
	Duration    int    `json:"duration"`     // 分
	Sets        int    `json:"sets"`         // セット数
	Reps        string `json:"reps"`         // 回数（例: "10回", "30秒"）
	Equipment   string `json:"equipment"`    // 使用器具
	Description string `json:"description"`  // 説明・ポイント
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

// SuggestMenu は利用可能な器具と時間に基づいてトレーニングメニューを提案
func (ms *MenuSuggester) SuggestMenu(ctx context.Context, equipment []string, durationMinutes int, goals []string) ([]MenuItem, error) {
	if len(equipment) == 0 {
		return nil, fmt.Errorf("利用可能な器具が指定されていません")
	}
	if durationMinutes <= 0 {
		return nil, fmt.Errorf("トレーニング時間は0より大きい値が必要です")
	}

	equipmentList := strings.Join(equipment, "、")
	goalsText := ""
	if len(goals) > 0 {
		goalsText = fmt.Sprintf("\n目標・重点: %s", strings.Join(goals, "、"))
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
- 格闘技（柔術・キックボクシング）向けのメニューを優先してください
- 各エクササイズには具体的なセット数と回数を指定してください
- JSON形式で出力してください`, equipmentList, durationMinutes, goalsText)

	response, err := ms.client.Execute(ctx, prompt)
	if err != nil {
		return nil, err
	}

	menuJSON := removeCodeBlock(response.Response)

	var menu []MenuItem
	if err := json.Unmarshal([]byte(menuJSON), &menu); err != nil {
		return nil, fmt.Errorf("メニューのパースエラー: %w\nデータ: %s", err, menuJSON)
	}

	return menu, nil
}
