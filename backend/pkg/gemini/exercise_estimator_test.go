package gemini

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEstimateCalories_EmptyExerciseName(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{}
	client := NewClientWithHTTPClient(mockHTTPClient)
	estimator := NewExerciseEstimatorWithClient(client)
	ctx := context.Background()

	_, err := estimator.EstimateCalories(ctx, "", 60)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exerciseNameが必要です")
}

func TestEstimateCalories_ZeroDuration(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{}
	client := NewClientWithHTTPClient(mockHTTPClient)
	estimator := NewExerciseEstimatorWithClient(client)
	ctx := context.Background()

	_, err := estimator.EstimateCalories(ctx, "柔術", 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "durationMinutesは1以上が必要です")
}

func TestEstimateCalories_MockSuccess(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return &Response{
				Response: `{"burned_calories_kcal": 486.0}`,
			}, nil
		},
	}

	client := NewClientWithHTTPClient(mockHTTPClient)
	estimator := NewExerciseEstimatorWithClient(client)
	ctx := context.Background()

	kcal, err := estimator.EstimateCalories(ctx, "柔術", 90)

	require.NoError(t, err)
	assert.InDelta(t, 486.0, kcal, 0.001)
}

func TestEstimateCalories_MockAPIError(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return nil, assert.AnError
		},
	}

	client := NewClientWithHTTPClient(mockHTTPClient)
	estimator := NewExerciseEstimatorWithClient(client)
	ctx := context.Background()

	_, err := estimator.EstimateCalories(ctx, "柔術", 60)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Gemini APIコールエラー")
}

func TestEstimateCalories_MockInvalidJSON(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return &Response{Response: `{invalid json}`}, nil
		},
	}

	client := NewClientWithHTTPClient(mockHTTPClient)
	estimator := NewExerciseEstimatorWithClient(client)
	ctx := context.Background()

	_, err := estimator.EstimateCalories(ctx, "柔術", 60)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "パースエラー")
}

func TestEstimateCalories_MockZeroCalories(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return &Response{Response: `{"burned_calories_kcal": 0}`}, nil
		},
	}

	client := NewClientWithHTTPClient(mockHTTPClient)
	estimator := NewExerciseEstimatorWithClient(client)
	ctx := context.Background()

	_, err := estimator.EstimateCalories(ctx, "柔術", 60)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "消費カロリーの推定に失敗")
}

func TestEstimateCalories_MockCodeBlock(t *testing.T) {
	mockHTTPClient := &MockGeminiHTTPClient{
		ExecuteFunc: func(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
			return &Response{
				Response: "```json\n{\"burned_calories_kcal\": 324.0}\n```",
			}, nil
		},
	}

	client := NewClientWithHTTPClient(mockHTTPClient)
	estimator := NewExerciseEstimatorWithClient(client)
	ctx := context.Background()

	kcal, err := estimator.EstimateCalories(ctx, "ランニング", 60)

	require.NoError(t, err)
	assert.InDelta(t, 324.0, kcal, 0.001)
}

func TestNewExerciseEstimatorWithClient(t *testing.T) {
	mockClient := &MockGeminiHTTPClient{}
	client := NewClientWithHTTPClient(mockClient)
	estimator := NewExerciseEstimatorWithClient(client)

	assert.NotNil(t, estimator)
	assert.Equal(t, client, estimator.client)
}

func TestNewExerciseEstimator_EmptyAPIKey(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")

	estimator, err := NewExerciseEstimator(60 * time.Second)

	require.Error(t, err)
	assert.Nil(t, estimator)
	assert.ErrorIs(t, err, ErrEmptyAPIKey)
}

func TestNewExerciseEstimatorWithAPIKey_EmptyAPIKey(t *testing.T) {
	estimator, err := NewExerciseEstimatorWithAPIKey("", 60*time.Second)

	require.Error(t, err)
	assert.Nil(t, estimator)
	assert.ErrorIs(t, err, ErrEmptyAPIKey)
}

func TestEstimateCalories_RealAPI(t *testing.T) {
	skipIfNoGeminiAPIKey(t)

	estimator, err := NewExerciseEstimator(60 * time.Second)
	require.NoError(t, err)
	ctx := context.Background()

	kcal, err := estimator.EstimateCalories(ctx, "柔術", 90)

	require.NoError(t, err)
	assert.Greater(t, kcal, 0.0)
	assert.Less(t, kcal, 2000.0)
}
