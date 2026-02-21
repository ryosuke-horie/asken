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

func TestMenuSuggest_Generate_Success(t *testing.T) {
	// Gemini APIを使用するためレート制限待機
	waitForGeminiRateLimit()

	client, ctx := authenticatedClient(t, 60*time.Second)

	// サジェスト生成に必要な食材を準備
	createTestIngredient(t, client, ctx, "鶏もも肉", "meat", 400.0, "g")
	createTestIngredient(t, client, ctx, "じゃがいも", "vegetable", 3.0, "個")

	reqBody := map[string]any{
		"mealType": "dinner",
		"count":    2,
	}

	resp, err := client.Post(ctx, "/api/menu/suggest", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]any
	err = resp.JSON(&body)
	require.NoError(t, err)

	suggestions, ok := body["suggestions"].([]any)
	require.True(t, ok, "Response should contain suggestions array")
	assert.NotEmpty(t, suggestions, "Suggestions should not be empty")

	// 各サジェストの構造を確認
	for _, s := range suggestions {
		m, ok := s.(map[string]any)
		require.True(t, ok)
		assert.NotEmpty(t, m["id"])
		assert.NotEmpty(t, m["title"])
		assert.Equal(t, "dinner", m["mealType"])
		assert.Equal(t, "suggested", m["status"])
	}
}

func TestMenuSuggest_Generate_InvalidMealType(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"mealType": "invalid_type",
		"count":    3,
	}

	resp, err := client.Post(ctx, "/api/menu/suggest", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestMenuSuggest_Generate_InvalidCount(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"mealType": "lunch",
		"count":    10, // 最大5を超過
	}

	resp, err := client.Post(ctx, "/api/menu/suggest", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestMenuSuggest_Generate_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqBody := map[string]any{
		"mealType": "lunch",
		"count":    3,
	}

	resp, err := testClient.Post(ctx, "/api/menu/suggest", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestMenuSuggest_List_Success(t *testing.T) {
	// 前のテストからのレート制限リセットを待つ
	waitForGeminiRateLimit()

	client, ctx := authenticatedClient(t, 60*time.Second)

	// 一覧取得前にサジェストを生成
	reqBody := map[string]any{
		"mealType": "lunch",
		"count":    1,
	}
	createResp, err := client.Post(ctx, "/api/menu/suggest", reqBody)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createBody map[string]any
	err = createResp.JSON(&createBody)
	require.NoError(t, err)

	suggestions, ok := createBody["suggestions"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, suggestions)

	firstSuggestion, ok := suggestions[0].(map[string]any)
	require.True(t, ok)
	createdID, ok := firstSuggestion["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, createdID)

	// 一覧取得（statusフィルターなし）
	listResp, err := client.Get(ctx, "/api/menu/suggestions?status=all")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, listResp.StatusCode)
	assert.Contains(t, listResp.Headers.Get("Content-Type"), "application/json")

	var listBody map[string]any
	err = listResp.JSON(&listBody)
	require.NoError(t, err)

	listSuggestions, ok := listBody["suggestions"].([]any)
	require.True(t, ok, "Response should contain suggestions array")

	// 作成したサジェストが一覧に含まれていることを確認
	found := false
	for _, item := range listSuggestions {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if id, ok := m["id"].(string); ok && id == createdID {
			found = true
			break
		}
	}
	assert.True(t, found, "Created suggestion should be in the list")
}

func TestMenuSuggest_List_WithSuggestedStatus(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// suggestedステータスでフィルター
	resp, err := client.Get(ctx, "/api/menu/suggestions?status=suggested")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]any
	err = resp.JSON(&body)
	require.NoError(t, err)

	suggestions, ok := body["suggestions"].([]any)
	require.True(t, ok, "Response should contain suggestions array")

	// フィルター結果はすべて suggested ステータスであることを確認
	for _, item := range suggestions {
		m, ok := item.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "suggested", m["status"])
	}
}

func TestMenuSuggest_List_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := testClient.Get(ctx, "/api/menu/suggestions")
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestMenuSuggest_GetDetail_Success(t *testing.T) {
	// Gemini APIを使用するためレート制限待機
	waitForGeminiRateLimit()

	client, ctx := authenticatedClient(t, 90*time.Second)

	// サジェストを生成
	reqBody := map[string]any{
		"mealType": "breakfast",
		"count":    1,
	}
	createResp, err := client.Post(ctx, "/api/menu/suggest", reqBody)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createBody map[string]any
	err = createResp.JSON(&createBody)
	require.NoError(t, err)

	suggestions, ok := createBody["suggestions"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, suggestions)

	firstSuggestion, ok := suggestions[0].(map[string]any)
	require.True(t, ok)
	suggestionID, ok := firstSuggestion["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, suggestionID)

	// 詳細取得（レシピ遅延生成が発生する）
	detailResp, err := client.Get(ctx, "/api/menu/suggestions/"+suggestionID)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, detailResp.StatusCode)
	assert.Contains(t, detailResp.Headers.Get("Content-Type"), "application/json")

	var detailBody map[string]any
	err = detailResp.JSON(&detailBody)
	require.NoError(t, err)
	assert.Equal(t, suggestionID, detailBody["id"])
	assert.NotEmpty(t, detailBody["title"])
	assert.NotEmpty(t, detailBody["mealType"])
}

func TestMenuSuggest_GetDetail_NotFound(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	resp, err := client.Get(ctx, "/api/menu/suggestions/00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMenuSuggest_GetDetail_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := testClient.Get(ctx, "/api/menu/suggestions/test-id")
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestMenuSuggest_Accept_Success(t *testing.T) {
	// Gemini APIを使用するためレート制限待機
	waitForGeminiRateLimit()

	client, ctx := authenticatedClient(t, 90*time.Second)

	// 採用テスト用の食材を準備
	createTestIngredient(t, client, ctx, "豆腐", "other", 300.0, "g")

	// サジェストを生成
	reqBody := map[string]any{
		"mealType": "dinner",
		"count":    1,
	}
	createResp, err := client.Post(ctx, "/api/menu/suggest", reqBody)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createBody map[string]any
	err = createResp.JSON(&createBody)
	require.NoError(t, err)

	suggestions, ok := createBody["suggestions"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, suggestions)

	firstSuggestion, ok := suggestions[0].(map[string]any)
	require.True(t, ok)
	suggestionID, ok := firstSuggestion["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, suggestionID)

	// サジェストを採用
	acceptResp, err := client.Post(ctx, "/api/menu/suggestions/"+suggestionID+"/accept", nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, acceptResp.StatusCode)
	assert.Contains(t, acceptResp.Headers.Get("Content-Type"), "application/json")

	var acceptBody map[string]any
	err = acceptResp.JSON(&acceptBody)
	require.NoError(t, err)

	// analysisRequestId が返されることを確認
	assert.NotEmpty(t, acceptBody["analysisRequestId"],
		"Accept response should contain analysisRequestId for meal record linkage")

	// deductedIngredients フィールドの確認
	_, hasDeducted := acceptBody["deductedIngredients"]
	assert.True(t, hasDeducted, "Accept response should contain deductedIngredients")

	// 採用後にステータスが accepted に変更されていることを確認
	detailResp, err := client.Get(ctx, "/api/menu/suggestions/"+suggestionID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, detailResp.StatusCode)

	var detailBody map[string]any
	err = detailResp.JSON(&detailBody)
	require.NoError(t, err)
	assert.Equal(t, "accepted", detailBody["status"])
}

func TestMenuSuggest_Accept_NotFound(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	resp, err := client.Post(ctx, "/api/menu/suggestions/00000000-0000-0000-0000-000000000000/accept", nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMenuSuggest_Accept_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := testClient.Post(ctx, "/api/menu/suggestions/test-id/accept", nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestMenuSuggest_Dismiss_Success(t *testing.T) {
	// Gemini APIを使用するためレート制限待機
	waitForGeminiRateLimit()

	client, ctx := authenticatedClient(t, 60*time.Second)

	// 却下テスト用のサジェストを生成
	reqBody := map[string]any{
		"mealType": "snack",
		"count":    1,
	}
	createResp, err := client.Post(ctx, "/api/menu/suggest", reqBody)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createBody map[string]any
	err = createResp.JSON(&createBody)
	require.NoError(t, err)

	suggestions, ok := createBody["suggestions"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, suggestions)

	firstSuggestion, ok := suggestions[0].(map[string]any)
	require.True(t, ok)
	suggestionID, ok := firstSuggestion["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, suggestionID)

	// サジェストを却下
	dismissResp, err := client.Post(ctx, "/api/menu/suggestions/"+suggestionID+"/dismiss", nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, dismissResp.StatusCode)

	// 却下後にステータスが dismissed に変更されていることを確認
	detailResp, err := client.Get(ctx, "/api/menu/suggestions/"+suggestionID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, detailResp.StatusCode)

	var detailBody map[string]any
	err = detailResp.JSON(&detailBody)
	require.NoError(t, err)
	assert.Equal(t, "dismissed", detailBody["status"])
}

func TestMenuSuggest_Dismiss_NotFound(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	resp, err := client.Post(ctx, "/api/menu/suggestions/00000000-0000-0000-0000-000000000000/dismiss", nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMenuSuggest_Dismiss_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := testClient.Post(ctx, "/api/menu/suggestions/test-id/dismiss", nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
