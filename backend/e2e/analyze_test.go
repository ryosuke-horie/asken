//go:build e2e

package e2e

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyze_TextInput_Success(t *testing.T) {
	skipIfGeminiDisabled(t)
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]string{
		"input_text": "テスト用の食事データ：ご飯一杯と味噌汁",
		"meal_type":  "lunch",
		"meal_date":  time.Now().Format("2006-01-02"),
	}

	resp, err := client.Post(ctx, "/api/analyze", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]string
	err = resp.JSON(&body)
	require.NoError(t, err)
	assert.NotEmpty(t, body["id"], "Response should contain analysis ID")
}

func TestAnalyze_TextInput_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqBody := map[string]string{
		"input_text": "テスト用の食事データ",
		"meal_type":  "lunch",
	}

	// 認証トークンなしのtestClientを直接使用
	resp, err := testClient.Post(ctx, "/api/analyze", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAnalyze_TextInput_InvalidMealType(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]string{
		"input_text": "テスト用の食事データ",
		"meal_type":  "invalid_type",
	}

	resp, err := client.Post(ctx, "/api/analyze", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAnalyze_GetStatus_Success(t *testing.T) {
	skipIfGeminiDisabled(t)
	// 前のテスト(TestAnalyze_TextInput_Success)からのレート制限リセットを待つ
	waitForGeminiRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	// まず分析リクエストを作成
	reqBody := map[string]string{
		"input_text": "テスト用：サラダ",
		"meal_type":  "dinner",
		"meal_date":  time.Now().Format("2006-01-02"),
	}

	createResp, err := client.Post(ctx, "/api/analyze", reqBody)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, createResp.StatusCode)

	var createBody map[string]string
	err = createResp.JSON(&createBody)
	require.NoError(t, err)

	analysisID := createBody["id"]
	require.NotEmpty(t, analysisID)

	// ステータスを取得
	statusResp, err := client.Get(ctx, "/api/analyze/"+analysisID)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, statusResp.StatusCode)
	assert.Contains(t, statusResp.Headers.Get("Content-Type"), "application/json")

	var statusBody map[string]any
	err = statusResp.JSON(&statusBody)
	require.NoError(t, err)

	status, ok := statusBody["status"].(string)
	require.True(t, ok, "Response should contain status field")
	assert.Contains(t, []string{"pending", "processing", "completed", "failed"}, status)
}

func TestAnalyze_GetStatus_NotFound(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// 存在しないIDでステータスを取得
	resp, err := client.Get(ctx, "/api/analyze/00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
