package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockGeminiExecutor はテスト用のexecutor実装
type mockGeminiExecutor struct {
	executeFunc func(ctx context.Context, prompt string) (*Response, error)
}

func (m *mockGeminiExecutor) Execute(ctx context.Context, prompt string) (*Response, error) {
	return m.executeFunc(ctx, prompt)
}

func TestMenuSuggester_SuggestMenu_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  SuggestMenuParams
		wantErr string
	}{
		{
			name: "Equipment空でエラー",
			params: SuggestMenuParams{
				Equipment:       []string{},
				DurationMinutes: 60,
			},
			wantErr: "利用可能な器具が指定されていません",
		},
		{
			name: "DurationMinutes <= 0 でエラー",
			params: SuggestMenuParams{
				Equipment:       []string{"ダンベル"},
				DurationMinutes: 0,
			},
			wantErr: "トレーニング時間は0より大きい値が必要です",
		},
		{
			name: "DurationMinutes負値でエラー",
			params: SuggestMenuParams{
				Equipment:       []string{"ダンベル"},
				DurationMinutes: -10,
			},
			wantErr: "トレーニング時間は0より大きい値が必要です",
		},
		{
			name: "DurationMinutes > 240 でエラー",
			params: SuggestMenuParams{
				Equipment:       []string{"ダンベル"},
				DurationMinutes: 241,
			},
			wantErr: "トレーニング時間は240分（4時間）以下で指定してください",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &mockGeminiExecutor{}
			ms := newMenuSuggesterWithExecutor(executor)

			result, err := ms.SuggestMenu(context.Background(), tt.params)
			assert.Nil(t, result)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestMenuSuggester_SuggestMenu_Success(t *testing.T) {
	menu := []MenuItem{
		{
			Name:        "ウォームアップ",
			Duration:    5,
			Sets:        1,
			Reps:        "5分",
			Equipment:   "なし",
			Description: "軽いストレッチ",
		},
		{
			Name:        "ダンベルカール",
			Duration:    10,
			Sets:        3,
			Reps:        "10回",
			Equipment:   "ダンベル",
			Description: "上腕二頭筋を鍛える",
		},
	}
	menuJSON, err := json.Marshal(menu)
	require.NoError(t, err)

	executor := &mockGeminiExecutor{
		executeFunc: func(ctx context.Context, prompt string) (*Response, error) {
			return &Response{Response: string(menuJSON)}, nil
		},
	}
	ms := newMenuSuggesterWithExecutor(executor)

	result, err := ms.SuggestMenu(context.Background(), SuggestMenuParams{
		Equipment:       []string{"ダンベル", "ベンチ"},
		DurationMinutes: 60,
	})

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "ウォームアップ", result[0].Name)
	assert.Equal(t, "ダンベルカール", result[1].Name)
}

func TestMenuSuggester_SuggestMenu_OptionalParams(t *testing.T) {
	fatigue := 3
	condition := 1
	sportType := "柔術"

	var capturedPrompt string
	executor := &mockGeminiExecutor{
		executeFunc: func(ctx context.Context, prompt string) (*Response, error) {
			capturedPrompt = prompt
			return &Response{Response: "[]"}, nil
		},
	}
	ms := newMenuSuggesterWithExecutor(executor)

	_, err := ms.SuggestMenu(context.Background(), SuggestMenuParams{
		Equipment:       []string{"マット"},
		DurationMinutes: 30,
		Goals:           []string{"柔軟性向上", "体幹強化"},
		Fatigue:         &fatigue,
		Condition:       &condition,
		SportType:       &sportType,
	})

	require.NoError(t, err)
	assert.Contains(t, capturedPrompt, "柔軟性向上、体幹強化")
	assert.Contains(t, capturedPrompt, "疲労度: 高い")
	assert.Contains(t, capturedPrompt, "体調: 悪い")
	assert.Contains(t, capturedPrompt, "柔術")
}

func TestMenuSuggester_SuggestMenu_DefaultSportType(t *testing.T) {
	var capturedPrompt string
	executor := &mockGeminiExecutor{
		executeFunc: func(ctx context.Context, prompt string) (*Response, error) {
			capturedPrompt = prompt
			return &Response{Response: "[]"}, nil
		},
	}
	ms := newMenuSuggesterWithExecutor(executor)

	_, err := ms.SuggestMenu(context.Background(), SuggestMenuParams{
		Equipment:       []string{"ダンベル"},
		DurationMinutes: 60,
	})

	require.NoError(t, err)
	assert.Contains(t, capturedPrompt, "格闘技（柔術・キックボクシング）")
}

func TestMenuSuggester_SuggestMenu_GeminiAPIError(t *testing.T) {
	executor := &mockGeminiExecutor{
		executeFunc: func(ctx context.Context, prompt string) (*Response, error) {
			return nil, errors.New("API error")
		},
	}
	ms := newMenuSuggesterWithExecutor(executor)

	result, err := ms.SuggestMenu(context.Background(), SuggestMenuParams{
		Equipment:       []string{"ダンベル"},
		DurationMinutes: 60,
	})

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Gemini APIによるメニュー提案に失敗")
}

func TestMenuSuggester_SuggestMenu_JSONParseError(t *testing.T) {
	executor := &mockGeminiExecutor{
		executeFunc: func(ctx context.Context, prompt string) (*Response, error) {
			return &Response{Response: "invalid json"}, nil
		},
	}
	ms := newMenuSuggesterWithExecutor(executor)

	result, err := ms.SuggestMenu(context.Background(), SuggestMenuParams{
		Equipment:       []string{"ダンベル"},
		DurationMinutes: 60,
	})

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "メニューのパースエラー")
}

func TestMenuSuggester_SuggestMenu_BoundaryDuration(t *testing.T) {
	executor := &mockGeminiExecutor{
		executeFunc: func(ctx context.Context, prompt string) (*Response, error) {
			return &Response{Response: "[]"}, nil
		},
	}
	ms := newMenuSuggesterWithExecutor(executor)

	t.Run("DurationMinutes=1は有効", func(t *testing.T) {
		_, err := ms.SuggestMenu(context.Background(), SuggestMenuParams{
			Equipment:       []string{"ダンベル"},
			DurationMinutes: 1,
		})
		assert.NoError(t, err)
	})

	t.Run("DurationMinutes=240は有効", func(t *testing.T) {
		_, err := ms.SuggestMenu(context.Background(), SuggestMenuParams{
			Equipment:       []string{"ダンベル"},
			DurationMinutes: 240,
		})
		assert.NoError(t, err)
	})
}

func TestMenuSuggester_SuggestMenu_CodeBlockResponse(t *testing.T) {
	menu := []MenuItem{
		{Name: "スクワット", Duration: 10, Sets: 3, Reps: "10回", Equipment: "なし", Description: "下半身強化"},
	}
	menuJSON, err := json.Marshal(menu)
	require.NoError(t, err)

	executor := &mockGeminiExecutor{
		executeFunc: func(ctx context.Context, prompt string) (*Response, error) {
			response := "```json\n" + string(menuJSON) + "\n```"
			return &Response{Response: response}, nil
		},
	}
	ms := newMenuSuggesterWithExecutor(executor)

	result, err := ms.SuggestMenu(context.Background(), SuggestMenuParams{
		Equipment:       []string{"なし"},
		DurationMinutes: 30,
	})

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "スクワット", result[0].Name)
}

func TestMenuSuggester_SuggestMenu_PromptContainsEquipment(t *testing.T) {
	var capturedPrompt string
	executor := &mockGeminiExecutor{
		executeFunc: func(ctx context.Context, prompt string) (*Response, error) {
			capturedPrompt = prompt
			return &Response{Response: "[]"}, nil
		},
	}
	ms := newMenuSuggesterWithExecutor(executor)

	_, err := ms.SuggestMenu(context.Background(), SuggestMenuParams{
		Equipment:       []string{"ダンベル", "バーベル", "チンニングバー"},
		DurationMinutes: 90,
	})

	require.NoError(t, err)
	assert.True(t, strings.Contains(capturedPrompt, "ダンベル、バーベル、チンニングバー"))
	assert.True(t, strings.Contains(capturedPrompt, "90分"))
}
