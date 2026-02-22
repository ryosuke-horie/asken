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

func TestIngredients_Create_Success(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"name":          "鶏むね肉",
		"category":      "meat",
		"quantity":      300.0,
		"unit":          "g",
		"purchase_date": time.Now().Format("2006-01-02"),
		"source":        "manual",
	}

	resp, err := client.Post(ctx, "/api/ingredients", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]any
	err = resp.JSON(&body)
	require.NoError(t, err)
	assert.NotEmpty(t, body["id"])
	assert.Equal(t, "鶏むね肉", body["name"])
	assert.Equal(t, "meat", body["category"])
	assert.Equal(t, "manual", body["source"])
}

func TestIngredients_Create_InvalidCategory(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"name":     "テスト食材",
		"category": "invalid_category",
		"quantity": 100.0,
		"unit":     "g",
		"source":   "manual",
	}

	resp, err := client.Post(ctx, "/api/ingredients", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestIngredients_Create_MissingName(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"category": "meat",
		"quantity": 100.0,
		"unit":     "g",
		"source":   "manual",
	}

	resp, err := client.Post(ctx, "/api/ingredients", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestIngredients_Create_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqBody := map[string]any{
		"name":     "鶏むね肉",
		"category": "meat",
		"quantity": 300.0,
		"unit":     "g",
		"source":   "manual",
	}

	resp, err := testClient.Post(ctx, "/api/ingredients", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestIngredients_List_Success(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	// 一覧取得前に食材を作成
	reqBody := map[string]any{
		"name":     "玉ねぎ",
		"category": "vegetable",
		"quantity": 2.0,
		"unit":     "個",
		"source":   "manual",
	}
	createResp, err := client.Post(ctx, "/api/ingredients", reqBody)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createBody map[string]any
	err = createResp.JSON(&createBody)
	require.NoError(t, err)
	createdID, ok := createBody["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, createdID)

	// 一覧取得
	listResp, err := client.Get(ctx, "/api/ingredients")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, listResp.StatusCode)
	assert.Contains(t, listResp.Headers.Get("Content-Type"), "application/json")

	var listBody map[string]any
	err = listResp.JSON(&listBody)
	require.NoError(t, err)

	ingredients, ok := listBody["ingredients"].([]any)
	require.True(t, ok, "Response should contain ingredients array")

	// 作成した食材が一覧に含まれていることを確認
	found := false
	for _, item := range ingredients {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if id, ok := m["id"].(string); ok && id == createdID {
			found = true
			assert.Equal(t, "玉ねぎ", m["name"])
			assert.Equal(t, "vegetable", m["category"])
			break
		}
	}
	assert.True(t, found, "Created ingredient should be in the list")
}

func TestIngredients_List_WithCategoryFilter(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	// categoryフィルターで取得
	resp, err := client.Get(ctx, "/api/ingredients?category=vegetable")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]any
	err = resp.JSON(&body)
	require.NoError(t, err)

	ingredients, ok := body["ingredients"].([]any)
	require.True(t, ok, "Response should contain ingredients array")

	// フィルター結果はすべて vegetable であることを確認
	for _, item := range ingredients {
		m, ok := item.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "vegetable", m["category"])
	}
}

func TestIngredients_List_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := testClient.Get(ctx, "/api/ingredients")
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestIngredients_Update_Success(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	// 更新対象の食材を作成
	createReq := map[string]any{
		"name":     "にんじん",
		"category": "vegetable",
		"quantity": 3.0,
		"unit":     "本",
		"source":   "manual",
	}
	createResp, err := client.Post(ctx, "/api/ingredients", createReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createBody map[string]any
	err = createResp.JSON(&createBody)
	require.NoError(t, err)

	ingredientID, ok := createBody["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, ingredientID)

	// 数量を更新
	updateReq := map[string]any{
		"name":     "にんじん",
		"category": "vegetable",
		"quantity": 1.0,
		"unit":     "本",
		"source":   "manual",
	}
	updateResp, err := client.Request(ctx, http.MethodPut, "/api/ingredients/"+ingredientID, updateReq)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, updateResp.StatusCode)
	assert.Contains(t, updateResp.Headers.Get("Content-Type"), "application/json")

	var updateBody map[string]any
	err = updateResp.JSON(&updateBody)
	require.NoError(t, err)
	assert.Equal(t, ingredientID, updateBody["id"])

	quantity, ok := updateBody["quantity"].(float64)
	require.True(t, ok, "quantity should be float64")
	assert.InDelta(t, 1.0, quantity, 0.001)
}

func TestIngredients_Update_NotFound(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	updateReq := map[string]any{
		"name":     "テスト食材",
		"category": "vegetable",
		"quantity": 1.0,
		"unit":     "個",
		"source":   "manual",
	}
	resp, err := client.Request(ctx, http.MethodPut, "/api/ingredients/00000000-0000-0000-0000-000000000000", updateReq)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestIngredients_Update_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	updateReq := map[string]any{
		"name":     "テスト食材",
		"category": "vegetable",
		"quantity": 1.0,
		"unit":     "個",
		"source":   "manual",
	}
	resp, err := testClient.Request(ctx, http.MethodPut, "/api/ingredients/test-id", updateReq)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestIngredients_Delete_Success(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	// 削除対象の食材を作成
	createReq := map[string]any{
		"name":     "削除用食材",
		"category": "other",
		"quantity": 1.0,
		"unit":     "個",
		"source":   "manual",
	}
	createResp, err := client.Post(ctx, "/api/ingredients", createReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createBody map[string]any
	err = createResp.JSON(&createBody)
	require.NoError(t, err)

	ingredientID, ok := createBody["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, ingredientID)

	// 削除
	deleteResp, err := client.Request(ctx, http.MethodDelete, "/api/ingredients/"+ingredientID, nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	// 削除後に一覧から消えていることを確認
	listResp, err := client.Get(ctx, "/api/ingredients")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, listResp.StatusCode)

	var listBody map[string]any
	err = listResp.JSON(&listBody)
	require.NoError(t, err)

	ingredients, ok := listBody["ingredients"].([]any)
	require.True(t, ok)

	for _, item := range ingredients {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if id, ok := m["id"].(string); ok {
			assert.NotEqual(t, ingredientID, id, "Deleted ingredient should not be in the list")
		}
	}
}

func TestIngredients_Delete_NotFound(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	resp, err := client.Request(ctx, http.MethodDelete, "/api/ingredients/00000000-0000-0000-0000-000000000000", nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestIngredients_Delete_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := testClient.Request(ctx, http.MethodDelete, "/api/ingredients/test-id", nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestIngredients_ScanReceipt_Success(t *testing.T) {
	// Gemini APIを使用するためレート制限待機
	waitForGeminiRateLimit()

	client, ctx := authenticatedClient(t, 60*time.Second)

	resp, err := client.UploadImage(ctx, "/api/ingredients/scan-receipt", minimalJPEG, "receipt.jpg")
	require.NoError(t, err)

	// Gemini APIがレシートを解析できない場合でも、APIは正常なレスポンスを返す
	// （空の食材リストもしくは解析結果を返す）
	assert.Contains(t, []int{http.StatusOK, http.StatusUnprocessableEntity}, resp.StatusCode,
		"ScanReceipt should return 200 or 422 for minimal JPEG (Gemini may not extract ingredients)")

	if resp.StatusCode == http.StatusOK {
		var body map[string]any
		err = resp.JSON(&body)
		require.NoError(t, err)
		_, ok := body["ingredients"]
		assert.True(t, ok, "Response should contain ingredients field")
	}
}

func TestIngredients_ScanReceipt_InvalidFile(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// テキストファイルをJPEGとして送信（マジックバイト不一致）
	invalidData := []byte("This is not an image file")

	resp, err := client.UploadImage(ctx, "/api/ingredients/scan-receipt", invalidData, "receipt.jpg")
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestIngredients_ScanReceipt_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := testClient.UploadImage(ctx, "/api/ingredients/scan-receipt", minimalJPEG, "receipt.jpg")
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
