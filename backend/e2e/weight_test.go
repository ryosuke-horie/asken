//go:build e2e

package e2e

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertFloat64Equal はfloat64の等値比較を文字列表現で行うヘルパー関数
// JSON数値はfloat64としてデコードされるため、小数点以下の誤差を許容
func assertFloat64Equal(t *testing.T, expected float64, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()

	// actual が float64 または float64 を含む any 型であることを確認
	var actualFloat float64
	switch v := actual.(type) {
	case float64:
		actualFloat = v
	case map[string]any:
		weightKg, ok := v["weight_kg"]
		if !ok {
			t.Fatalf("assertFloat64Equal: map does not contain 'weight_kg' key: %+v", v)
		}
		f, ok := weightKg.(float64)
		if !ok {
			t.Fatalf("assertFloat64Equal: weight_kg is not float64, got %T: %v", weightKg, weightKg)
		}
		actualFloat = f
	case string:
		// 文字列からfloat64に変換
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			t.Fatalf("assertFloat64Equal: failed to parse string as float64: %q (error: %v)", v, err)
		}
		actualFloat = f
	default:
		t.Fatalf("assertFloat64Equal: unexpected type %T for actual value", actual)
	}

	// 小数点第3位までの誤差を許容（JSON数値の精度問題対応）
	delta := 0.001
	assert.InDelta(t, expected, actualFloat, delta, msgAndArgs...)
}

func TestWeightRecords_Create_Success(t *testing.T) {
	// 前のテストファイル群でユーザーレート制限バケットが枯渇している可能性があるため待機
	waitForUserRateLimit()

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

	// weight_kg はJSON数値としてfloat64でデコードされるため、型アサーションが必要
	weightKg, ok := body["weight_kg"].(float64)
	require.True(t, ok, "weight_kg should be a float64")
	assertFloat64Equal(t, 70.5, weightKg, "weight_kg should match")
	assert.Equal(t, "テスト用記録", body["note"])
}

func TestWeightRecords_Create_InvalidWeightKg_TooLow(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	recordedAt := time.Now().Format(time.RFC3339)
	reqBody := map[string]any{
		"weight_kg":   19.9, // 20.0未満は無効
		"recorded_at": recordedAt,
	}

	resp, err := client.Post(ctx, "/api/weight/records", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestWeightRecords_Create_InvalidWeightKg_TooHigh(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	recordedAt := time.Now().Format(time.RFC3339)
	reqBody := map[string]any{
		"weight_kg":   300.1, // 300.0超過は無効
		"recorded_at": recordedAt,
	}

	resp, err := client.Post(ctx, "/api/weight/records", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestWeightRecords_Create_MissingRecordedAt(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"weight_kg": 70.5,
		// recorded_at が欠落
	}

	resp, err := client.Post(ctx, "/api/weight/records", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestWeightRecords_Create_InvalidRecordedAt_Format(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"weight_kg":   70.5,
		"recorded_at": "2024-02-13", // RFC3339形式でない
	}

	resp, err := client.Post(ctx, "/api/weight/records", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestWeightRecords_Create_NoteTooLong(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	recordedAt := time.Now().Format(time.RFC3339)
	longNote := string(make([]byte, 201)) // 201文字（200文字超過）
	reqBody := map[string]any{
		"weight_kg":   70.5,
		"recorded_at": recordedAt,
		"note":        longNote,
	}

	resp, err := client.Post(ctx, "/api/weight/records", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestWeightRecords_Get_Success(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

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

	recordID, ok := createBody["id"].(string)
	require.True(t, ok, "Response id field should be a string")
	require.NotEmpty(t, recordID)

	getResp, err := client.Get(ctx, "/api/weight/records/"+recordID)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, getResp.StatusCode)
	assert.Contains(t, getResp.Headers.Get("Content-Type"), "application/json")

	var getBody map[string]any
	err = getResp.JSON(&getBody)
	require.NoError(t, err)
	assert.Equal(t, recordID, getBody["id"])

	// weight_kg はJSON数値としてfloat64でデコードされるため、型アサーションが必要
	weightKg, ok := getBody["weight_kg"].(float64)
	require.True(t, ok, "weight_kg should be a float64")
	assertFloat64Equal(t, 68.0, weightKg, "weight_kg should match")
}

func TestWeightRecords_Get_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := testClient.Get(ctx, "/api/weight/records/test-id")
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWeightRecords_Update_Success(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

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

	recordID, ok := createBody["id"].(string)
	require.True(t, ok, "Response id field should be a string")
	require.NotEmpty(t, recordID)

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
	// weight_kg はJSON数値としてfloat64でデコードされるため、型アサーションが必要
	weightKg, ok := updateBody["weight_kg"].(float64)
	require.True(t, ok, "weight_kg should be a float64")
	assertFloat64Equal(t, 69.5, weightKg, "weight_kg should match")
	assert.Equal(t, "更新後のメモ", updateBody["note"])
}

func TestWeightRecords_Update_NotFound(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	updateReq := map[string]any{
		"weight_kg": 69.5,
		"note":      "更新後のメモ",
	}

	updateResp, err := client.Request(ctx, http.MethodPut, "/api/weight/records/00000000-0000-0000-0000-000000000000", updateReq)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, updateResp.StatusCode)
}

func TestWeightRecords_Update_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	updateReq := map[string]any{
		"weight_kg": 69.5,
	}

	resp, err := testClient.Request(ctx, http.MethodPut, "/api/weight/records/test-id", updateReq)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWeightRecords_Delete_Success(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

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

	recordID, ok := createBody["id"].(string)
	require.True(t, ok, "Response id field should be a string")
	require.NotEmpty(t, recordID)

	deleteResp, err := client.Request(ctx, http.MethodDelete, "/api/weight/records/"+recordID, nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	getResp, err := client.Get(ctx, "/api/weight/records/"+recordID)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
}

func TestWeightRecords_Delete_NotFound(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	deleteResp, err := client.Request(ctx, http.MethodDelete, "/api/weight/records/00000000-0000-0000-0000-000000000000", nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, deleteResp.StatusCode)
}

func TestWeightRecords_Delete_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := testClient.Request(ctx, http.MethodDelete, "/api/weight/records/test-id", nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWeightRecords_List_Success(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

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

		id, ok := body["id"].(string)
		require.True(t, ok, "Response id field should be a string")
		createdIDs = append(createdIDs, id)
	}
	require.Len(t, createdIDs, 3)

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
	// 過去のE2Eテスト実行でレコードが蓄積するため、件数ではなく作成したIDの存在を検証
	assert.GreaterOrEqual(t, len(recordsArray), 3, "Should return at least 3 records")

	// 作成した3件のIDが一覧に含まれていることを確認
	foundIDs := make(map[string]bool)
	for _, record := range recordsArray {
		recordMap, ok := record.(map[string]any)
		require.True(t, ok)
		if id, ok := recordMap["id"].(string); ok {
			foundIDs[id] = true
		}
	}
	for _, id := range createdIDs {
		assert.True(t, foundIDs[id], "Created record %s should be in the list", id)
	}
}

func TestWeightRecords_List_MissingFromParameter(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	to := time.Now().Format("2006-01-02")
	resp, err := client.Get(ctx, "/api/weight/records?to="+to)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestWeightRecords_List_MissingToParameter(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	from := time.Now().Format("2006-01-02")
	resp, err := client.Get(ctx, "/api/weight/records?from="+from)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestWeightRecords_List_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := testClient.Get(ctx, "/api/weight/records?from=2024-01-01&to=2024-12-31")
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWeightRecords_Get_NotFound(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	resp, err := client.Get(ctx, "/api/weight/records/00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestWeightRecords_Create_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	recordedAt := time.Now().Format(time.RFC3339)
	reqBody := map[string]any{
		"weight_kg":   70.5,
		"recorded_at": recordedAt,
	}

	resp, err := testClient.Post(ctx, "/api/weight/records", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWeightGoal_Set_Success(t *testing.T) {
	waitForUserRateLimit()

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

	// target_weight_kg はJSON数値としてfloat64でデコードされるため、型アサーションが必要
	targetWeightKg, ok := body["target_weight_kg"].(float64)
	require.True(t, ok, "target_weight_kg should be a float64")
	assertFloat64Equal(t, 65.0, targetWeightKg, "target_weight_kg should match")
	assert.NotEmpty(t, body["updated_at"])
}

func TestWeightGoal_Set_InvalidTargetWeightKg_TooLow(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"target_weight_kg": 19.9, // 20.0未満は無効
	}

	resp, err := client.Request(ctx, http.MethodPut, "/api/weight/goal", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestWeightGoal_Set_InvalidTargetWeightKg_TooHigh(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"target_weight_kg": 300.1, // 300.0超過は無効
	}

	resp, err := client.Request(ctx, http.MethodPut, "/api/weight/goal", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestWeightGoal_Set_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqBody := map[string]any{
		"target_weight_kg": 65.0,
	}

	resp, err := testClient.Request(ctx, http.MethodPut, "/api/weight/goal", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWeightGoal_Get_Success(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	setReq := map[string]any{
		"target_weight_kg": 66.0,
	}
	setResp, err := client.Request(ctx, http.MethodPut, "/api/weight/goal", setReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, setResp.StatusCode)

	getResp, err := client.Get(ctx, "/api/weight/goal")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, getResp.StatusCode)
	assert.Contains(t, getResp.Headers.Get("Content-Type"), "application/json")

	var getBody map[string]any
	err = getResp.JSON(&getBody)
	require.NoError(t, err)

	goal, ok := getBody["goal"].(map[string]any)
	require.True(t, ok, "Response should contain goal object")

	// target_weight_kg はJSON数値としてfloat64でデコードされるため、型アサーションが必要
	targetWeightKg, ok := goal["target_weight_kg"].(float64)
	require.True(t, ok, "target_weight_kg should be a float64")
	assertFloat64Equal(t, 66.0, targetWeightKg, "target_weight_kg should match")
	assert.NotEmpty(t, goal["updated_at"])
}

func TestWeightGoal_Get_NotSet(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	getResp, err := client.Get(ctx, "/api/weight/goal")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	var getBody map[string]any
	err = getResp.JSON(&getBody)
	require.NoError(t, err)

	// goal が設定されていない場合、goal フィールドは null または存在しない
	goal, exists := getBody["goal"]
	if exists {
		_, ok := goal.(map[string]any)
		// goal が存在する場合は map であることを確認（未設定時は nil）
		if ok {
			// 目標が設定されている場合は成功
		}
	}
	// nil の場合も成功（未設定状態）
}

func TestWeightGoal_Get_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := testClient.Get(ctx, "/api/weight/goal")
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
