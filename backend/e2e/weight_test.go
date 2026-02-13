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

func TestWeightRecords_Create_Success(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	recordedAt := time.Now().Format(time.RFC3339)
	reqBody := map[string]any{
		"weight_kg":   70.5,
		"recorded_at": recordedAt,
		"note":        "テスト用記録",
	}

	resp, err := client.Post(ctx, "/api/weight/records", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]any
	err = resp.JSON(&body)
	require.NoError(t, err)
	assert.NotEmpty(t, body["id"], "Response should contain record ID")
	assert.Equal(t, 70.5, body["weight_kg"])
	assert.Equal(t, "テスト用記録", body["note"])
}

func TestWeightRecords_Get_Success(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// まず体重記録を作成
	recordedAt := time.Now().Format(time.RFC3339)
	createReq := map[string]any{
		"weight_kg":   68.0,
		"recorded_at": recordedAt,
	}

	createResp, err := client.Post(ctx, "/api/weight/records", createReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createBody map[string]any
	err = createResp.JSON(&createBody)
	require.NoError(t, err)

	recordID := createBody["id"].(string)
	require.NotEmpty(t, recordID)

	// 記録を取得
	getResp, err := client.Get(ctx, "/api/weight/records/"+recordID)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, getResp.StatusCode)
	assert.Contains(t, getResp.Headers.Get("Content-Type"), "application/json")

	var getBody map[string]any
	err = getResp.JSON(&getBody)
	require.NoError(t, err)
	assert.Equal(t, recordID, getBody["id"])
	assert.Equal(t, 68.0, getBody["weight_kg"])
}

func TestWeightRecords_Update_Success(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// まず体重記録を作成
	recordedAt := time.Now().Format(time.RFC3339)
	createReq := map[string]any{
		"weight_kg":   70.0,
		"recorded_at": recordedAt,
	}

	createResp, err := client.Post(ctx, "/api/weight/records", createReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createBody map[string]any
	err = createResp.JSON(&createBody)
	require.NoError(t, err)

	recordID := createBody["id"].(string)
	require.NotEmpty(t, recordID)

	// 記録を更新
	updateReq := map[string]any{
		"weight_kg": 69.5,
		"note":      "更新後のメモ",
	}

	updateResp, err := client.Request(ctx, http.MethodPut, "/api/weight/records/"+recordID, updateReq)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, updateResp.StatusCode)
	assert.Contains(t, updateResp.Headers.Get("Content-Type"), "application/json")

	var updateBody map[string]any
	err = updateResp.JSON(&updateBody)
	require.NoError(t, err)
	assert.Equal(t, recordID, updateBody["id"])
	assert.Equal(t, 69.5, updateBody["weight_kg"])
	assert.Equal(t, "更新後のメモ", updateBody["note"])
}

func TestWeightRecords_Delete_Success(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// まず体重記録を作成
	recordedAt := time.Now().Format(time.RFC3339)
	createReq := map[string]any{
		"weight_kg":   70.0,
		"recorded_at": recordedAt,
	}

	createResp, err := client.Post(ctx, "/api/weight/records", createReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createBody map[string]any
	err = createResp.JSON(&createBody)
	require.NoError(t, err)

	recordID := createBody["id"].(string)
	require.NotEmpty(t, recordID)

	// 記録を削除
	deleteResp, err := client.Request(ctx, http.MethodDelete, "/api/weight/records/"+recordID, nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	// 削除確認（404になるはず）
	getResp, err := client.Get(ctx, "/api/weight/records/"+recordID)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
}

func TestWeightRecords_List_Success(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// 複数の記録を作成
	now := time.Now()
	records := []struct {
		weightKg   float64
		recordedAt string
		note       string
	}{
		{70.0, now.Add(-48 * time.Hour).Format(time.RFC3339), "2日前"},
		{69.5, now.Add(-24 * time.Hour).Format(time.RFC3339), "昨日"},
		{69.0, now.Format(time.RFC3339), "今日"},
	}

	var createdIDs []string
	for _, r := range records {
		req := map[string]any{
			"weight_kg":   r.weightKg,
			"recorded_at": r.recordedAt,
			"note":        r.note,
		}
		resp, err := client.Post(ctx, "/api/weight/records", req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var body map[string]any
		err = resp.JSON(&body)
		require.NoError(t, err)
		createdIDs = append(createdIDs, body["id"].(string))
	}
	require.Len(t, createdIDs, 3)

	// 一覧を取得
	from := now.Add(-72 * time.Hour).Format("2006-01-02")
	to := now.Add(24 * time.Hour).Format("2006-01-02")
	listResp, err := client.Get(ctx, "/api/weight/records?from="+from+"&to="+to)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, listResp.StatusCode)
	assert.Contains(t, listResp.Headers.Get("Content-Type"), "application/json")

	var listBody map[string]any
	err = listResp.JSON(&listBody)
	require.NoError(t, err)

	recordsArray, ok := listBody["records"].([]any)
	require.True(t, ok, "Response should contain records array")
	assert.Len(t, recordsArray, 3, "Should return 3 records")
}

func TestWeightRecords_Get_NotFound(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// 存在しないIDで取得
	resp, err := client.Get(ctx, "/api/weight/records/00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestWeightGoal_Set_Success(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"target_weight_kg": 65.0,
	}

	resp, err := client.Request(ctx, http.MethodPut, "/api/weight/goal", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]any
	err = resp.JSON(&body)
	require.NoError(t, err)
	assert.Equal(t, 65.0, body["target_weight_kg"])
	assert.NotEmpty(t, body["updated_at"])
}

func TestWeightGoal_Get_Success(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// まず目標を設定
	setReq := map[string]any{
		"target_weight_kg": 66.0,
	}
	setResp, err := client.Request(ctx, http.MethodPut, "/api/weight/goal", setReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, setResp.StatusCode)

	// 目標を取得
	getResp, err := client.Get(ctx, "/api/weight/goal")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, getResp.StatusCode)
	assert.Contains(t, getResp.Headers.Get("Content-Type"), "application/json")

	var getBody map[string]any
	err = getResp.JSON(&getBody)
	require.NoError(t, err)

	goal, ok := getBody["goal"].(map[string]any)
	require.True(t, ok, "Response should contain goal object")
	assert.Equal(t, 66.0, goal["target_weight_kg"])
	assert.NotEmpty(t, goal["updated_at"])
}

func TestWeightGoal_Get_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 認証なしのtestClientを直接使用
	resp, err := testClient.Get(ctx, "/api/weight/goal")
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
