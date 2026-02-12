//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHistory_List_Success は認証済みで履歴一覧取得が成功することを確認する
func TestHistory_List_Success(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// テストデータを作成（analyze API経由）
	waitForGeminiRateLimit()
	historyID := createTestHistory(t, ctx, client)

	// 履歴一覧を取得
	resp, err := client.Get(ctx, "/api/history")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]any
	err = resp.JSON(&body)
	require.NoError(t, err)

	// レスポンス構造を検証
	items, ok := body["items"].([]any)
	require.True(t, ok, "Response should contain items array")
	assert.GreaterOrEqual(t, len(items), 1, "Items should contain at least the created history")

	total, ok := body["total"].(float64)
	require.True(t, ok, "Response should contain total")
	assert.GreaterOrEqual(t, total, float64(1), "Total should be at least 1")

	// 作成した履歴が一覧に含まれていることを確認
	found := false
	for _, item := range items {
		itemMap, ok := item.(map[string]any)
		require.True(t, ok)
		if itemMap["id"] == historyID.String() {
			found = true
			break
		}
	}
	assert.True(t, found, "Created history should be in the list")
}

// TestHistory_List_Unauthorized は認証なしで401が返ることを確認する
func TestHistory_List_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 認証トークンなしのtestClientを直接使用
	resp, err := testClient.Get(ctx, "/api/history")
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestHistory_Detail_Success は認証済みで履歴詳細取得が成功することを確認する
func TestHistory_Detail_Success(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// テストデータを作成（analyze API経由）
	waitForGeminiRateLimit()
	historyID := createTestHistory(t, ctx, client)

	// 履歴詳細を取得
	resp, err := client.Get(ctx, "/api/history/"+historyID.String())
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]any
	err = resp.JSON(&body)
	require.NoError(t, err)

	// 必須フィールドを検証
	assert.NotEmpty(t, body["id"], "Response should contain id")
	assert.NotEmpty(t, body["meal_type"], "Response should contain meal_type")
	assert.NotEmpty(t, body["created_at"], "Response should contain created_at")
	assert.NotEmpty(t, body["total_calories"], "Response should contain total_calories")

	// foods配列が含まれていることを確認
	foods, ok := body["foods"].([]any)
	require.True(t, ok, "Response should contain foods array")
	assert.GreaterOrEqual(t, len(foods), 1, "Foods should contain at least one item")
}

// TestHistory_Detail_NotFound は存在しないIDで404が返ることを確認する
func TestHistory_Detail_NotFound(t *testing.T) {
	waitForGeminiRateLimit()
	client, ctx := authenticatedClient(t, 30*time.Second)

	// 存在しないUUIDで詳細を取得
	resp, err := client.Get(ctx, "/api/history/00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestHistory_Update_Success は認証済みで履歴更新が成功することを確認する
func TestHistory_Update_Success(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// テストデータを作成（analyze API経由）
	waitForGeminiRateLimit()
	historyID := createTestHistory(t, ctx, client)

	// 更新リクエストボディ
	updateReq := map[string]any{
		"foods": []map[string]any{
			{
				"name":             "更新されたテスト食品",
				"estimated_amount": "100g",
				"calories_kcal":    150.0,
				"protein_g":        10.0,
				"fat_g":            5.0,
				"carbohydrates_g":  20.0,
			},
		},
	}

	// 履歴を更新
	resp, err := client.Request(ctx, http.MethodPut, "/api/history/"+historyID.String(), updateReq)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]any
	err = resp.JSON(&body)
	require.NoError(t, err)

	// 更新されたfoodsを確認
	foods, ok := body["foods"].([]any)
	require.True(t, ok, "Response should contain foods array")
	assert.GreaterOrEqual(t, len(foods), 1, "Foods should contain at least one item")

	// 最初の食品の名前が更新されていることを確認
	if len(foods) > 0 {
		firstFood, ok := foods[0].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "更新されたテスト食品", firstFood["name"], "Food name should be updated")
	}
}

// TestHistory_Update_NotFound は存在しないIDで404が返ることを確認する
func TestHistory_Update_NotFound(t *testing.T) {
	waitForGeminiRateLimit()
	client, ctx := authenticatedClient(t, 30*time.Second)

	// 更新リクエストボディ
	updateReq := map[string]any{
		"foods": []map[string]any{
			{
				"name":             "テスト食品",
				"estimated_amount": "100g",
				"calories_kcal":    150.0,
				"protein_g":        10.0,
				"fat_g":            5.0,
				"carbohydrates_g":  20.0,
			},
		},
	}

	// 存在しないUUIDで更新
	resp, err := client.Request(ctx, http.MethodPut, "/api/history/00000000-0000-0000-0000-000000000000", updateReq)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestHistory_Update_Unauthorized は認証なしで401が返ることを確認する
func TestHistory_Update_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 更新リクエストボディ
	updateReq := map[string]any{
		"foods": []map[string]any{
			{
				"name":             "テスト食品",
				"estimated_amount": "100g",
				"calories_kcal":    150.0,
				"protein_g":        10.0,
				"fat_g":            5.0,
				"carbohydrates_g":  20.0,
			},
		},
	}

	// 認証トークンなしで更新
	resp, err := testClient.Request(ctx, http.MethodPut, "/api/history/00000000-0000-0000-0000-000000000000", updateReq)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestHistory_Update_InvalidRequest_EmptyFoods は空のfoods配列で400が返ることを確認する
func TestHistory_Update_InvalidRequest_EmptyFoods(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// テストデータを作成（analyze API経由）
	waitForGeminiRateLimit()
	historyID := createTestHistory(t, ctx, client)

	// 空のfoods配列でリクエスト
	updateReq := map[string]any{
		"foods": []any{},
	}

	resp, err := client.Request(ctx, http.MethodPut, "/api/history/"+historyID.String(), updateReq)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestHistory_Update_InvalidRequest_NegativeCalories は負のカロリー値で400が返ることを確認する
func TestHistory_Update_InvalidRequest_NegativeCalories(t *testing.T) {
	waitForGeminiRateLimit()
	client, ctx := authenticatedClient(t, 30*time.Second)

	// テストデータを作成（analyze API経由）
	waitForGeminiRateLimit()
	historyID := createTestHistory(t, ctx, client)

	// 負のカロリー値でリクエスト
	updateReq := map[string]any{
		"foods": []map[string]any{
			{
				"name":             "テスト食品",
				"estimated_amount": "100g",
				"calories_kcal":    -10.0,
				"protein_g":        10.0,
				"fat_g":            5.0,
				"carbohydrates_g":  20.0,
			},
		},
	}

	resp, err := client.Request(ctx, http.MethodPut, "/api/history/"+historyID.String(), updateReq)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestHistory_Update_InvalidRequest_EmptyName は空の食品名で400が返ることを確認する
func TestHistory_Update_InvalidRequest_EmptyName(t *testing.T) {
	waitForGeminiRateLimit()
	client, ctx := authenticatedClient(t, 30*time.Second)

	// テストデータを作成（analyze API経由）
	waitForGeminiRateLimit()
	historyID := createTestHistory(t, ctx, client)

	// 空の食品名でリクエスト
	updateReq := map[string]any{
		"foods": []map[string]any{
			{
				"name":             "",
				"estimated_amount": "100g",
				"calories_kcal":    150.0,
				"protein_g":        10.0,
				"fat_g":            5.0,
				"carbohydrates_g":  20.0,
			},
		},
	}

	resp, err := client.Request(ctx, http.MethodPut, "/api/history/"+historyID.String(), updateReq)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestHistory_Delete_Success は認証済みで履歴削除が成功することを確認する
func TestHistory_Delete_Success(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// テストデータを作成（analyze API経由）
	waitForGeminiRateLimit()
	historyID := createTestHistory(t, ctx, client)

	// 履歴を削除
	resp, err := client.Request(ctx, http.MethodDelete, "/api/history/"+historyID.String(), nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// 削除されたことを確認（再度GETして404を確認）
	getResp, err := client.Get(ctx, "/api/history/"+historyID.String())
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode, "Deleted history should return 404")
}

// TestHistory_Delete_NotFound は存在しないIDで404が返ることを確認する
func TestHistory_Delete_NotFound(t *testing.T) {
	waitForGeminiRateLimit()
	client, ctx := authenticatedClient(t, 30*time.Second)

	// 存在しないUUIDで削除
	resp, err := client.Request(ctx, http.MethodDelete, "/api/history/00000000-0000-0000-0000-000000000000", nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestHistory_Delete_Unauthorized は認証なしで401が返ることを確認する
func TestHistory_Delete_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 認証トークンなしで削除
	resp, err := testClient.Request(ctx, http.MethodDelete, "/api/history/00000000-0000-0000-0000-000000000000", nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// createTestHistory はanalyze API経由でテスト用の履歴データを作成するヘルパー関数
//
// 解析が完了するまでポーリングし、完了したhistoryIDを返す
func createTestHistory(t *testing.T, ctx context.Context, client *Client) uuid.UUID {
	t.Helper()

	reqBody := map[string]string{
		"input_text": "テスト用の食事データ：ご飯一杯と味噌汁",
		"meal_type":  "lunch",
		"meal_date":  time.Now().Format("2006-01-02"),
	}

	// 1. analyze APIでリクエストを作成
	resp, err := client.Post(ctx, "/api/analyze", reqBody)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode, "analyze API should return 202 Accepted")

	var body map[string]string
	err = resp.JSON(&body)
	require.NoError(t, err)

	analysisID, err := uuid.Parse(body["id"])
	require.NoError(t, err)
	require.NotEmpty(t, analysisID, "Response should contain analysis ID")

	// 2. 解析完了までポーリング（最大30秒）
	// 現在のアーキテクチャでは analysis ID = history ID（統合コレクション）
	const (
		pollInterval = 2 * time.Second
		maxPollTime  = 30 * time.Second
	)
	start := time.Now()

	for time.Since(start) < maxPollTime {
		statusResp, err := client.Get(ctx, "/api/analyze/"+analysisID.String())
		require.NoError(t, err)

		if statusResp.StatusCode == http.StatusNotFound {
			time.Sleep(pollInterval)
			continue
		}
		if statusResp.StatusCode != http.StatusOK {
			require.Fail(t, fmt.Sprintf("Unexpected status code during polling: %d", statusResp.StatusCode))
		}

		var statusBody map[string]any
		err = statusResp.JSON(&statusBody)
		require.NoError(t, err)

		status, ok := statusBody["status"].(string)
		require.True(t, ok, "Response should contain status field")

		if status == "completed" {
			return analysisID
		}

		if status == "failed" {
			errorMsg := "unknown error"
			if errMsg, ok := statusBody["error"].(string); ok && errMsg != "" {
				errorMsg = errMsg
			}
			require.Fail(t, "Analysis failed: "+errorMsg)
		}

		time.Sleep(pollInterval)
	}

	require.Fail(t, fmt.Sprintf("Analysis did not complete within %v (analysisID: %s)", maxPollTime, analysisID))
	return uuid.UUID{}
}
