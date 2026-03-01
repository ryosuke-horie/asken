//go:build e2e

package e2e

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExerciseRecords_Create_Success(t *testing.T) {
	// 前のテストファイル群でユーザーレート制限バケットが枯渇している可能性があるため待機
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"exercise_name":    "柔術", // プリセット種目（MET計算）
		"duration_minutes": 60,
		"recorded_date":    time.Now().Format("2006-01-02"),
	}

	resp, err := client.Post(ctx, "/api/exercise/records", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]any
	err = resp.JSON(&body)
	require.NoError(t, err)

	assert.NotEmpty(t, body["id"], "Response should contain record ID")
	assert.Equal(t, "柔術", body["exercise_name"])
	assert.Equal(t, float64(60), body["duration_minutes"])
	assert.Equal(t, "met", body["estimation_method"], "プリセット種目はMET推定であるべき")

	burnedCalories, ok := body["burned_calories_kcal"].(float64)
	require.True(t, ok, "burned_calories_kcal should be a float64")
	assert.Greater(t, burnedCalories, 0.0, "burned_calories_kcal should be positive")
}

func TestExerciseRecords_Create_WithGemini_Success(t *testing.T) {
	skipIfGeminiDisabled(t)
	waitForGeminiRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"exercise_name":    "テニス", // プリセット外種目（Gemini推定）
		"duration_minutes": 45,
		"recorded_date":    time.Now().Format("2006-01-02"),
	}

	resp, err := client.Post(ctx, "/api/exercise/records", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]any
	err = resp.JSON(&body)
	require.NoError(t, err)

	assert.NotEmpty(t, body["id"], "Response should contain record ID")
	assert.Equal(t, "テニス", body["exercise_name"])
	assert.Equal(t, "gemini", body["estimation_method"], "プリセット外種目はGemini推定であるべき")

	burnedCalories, ok := body["burned_calories_kcal"].(float64)
	require.True(t, ok, "burned_calories_kcal should be a float64")
	assert.Greater(t, burnedCalories, 0.0, "burned_calories_kcal should be positive")
}

func TestExerciseRecords_Create_InvalidExerciseName_Empty(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"exercise_name":    "", // 空文字列は無効
		"duration_minutes": 60,
		"recorded_date":    time.Now().Format("2006-01-02"),
	}

	resp, err := client.Post(ctx, "/api/exercise/records", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestExerciseRecords_Create_InvalidExerciseName_TooLong(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"exercise_name":    strings.Repeat("あ", 101), // 101文字（100文字超過）
		"duration_minutes": 60,
		"recorded_date":    time.Now().Format("2006-01-02"),
	}

	resp, err := client.Post(ctx, "/api/exercise/records", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestExerciseRecords_Create_InvalidDurationMinutes_TooLow(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"exercise_name":    "柔術",
		"duration_minutes": 4, // 5未満は無効
		"recorded_date":    time.Now().Format("2006-01-02"),
	}

	resp, err := client.Post(ctx, "/api/exercise/records", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestExerciseRecords_Create_InvalidDurationMinutes_TooHigh(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"exercise_name":    "柔術",
		"duration_minutes": 601, // 600超過は無効
		"recorded_date":    time.Now().Format("2006-01-02"),
	}

	resp, err := client.Post(ctx, "/api/exercise/records", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestExerciseRecords_Create_MissingRecordedDate(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"exercise_name":    "柔術",
		"duration_minutes": 60,
		// recorded_date が欠落
	}

	resp, err := client.Post(ctx, "/api/exercise/records", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestExerciseRecords_Create_InvalidRecordedDate_Format(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"exercise_name":    "柔術",
		"duration_minutes": 60,
		"recorded_date":    "2024-02-13T00:00:00Z", // RFC3339形式はYYYY-MM-DD形式として無効
	}

	resp, err := client.Post(ctx, "/api/exercise/records", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestExerciseRecords_Create_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqBody := map[string]any{
		"exercise_name":    "柔術",
		"duration_minutes": 60,
		"recorded_date":    time.Now().Format("2006-01-02"),
	}

	resp, err := testClient.Post(ctx, "/api/exercise/records", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestExerciseRecords_Delete_Success(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	// まず記録を作成する
	createReq := map[string]any{
		"exercise_name":    "柔術",
		"duration_minutes": 60,
		"recorded_date":    time.Now().Format("2006-01-02"),
	}
	createResp, err := client.Post(ctx, "/api/exercise/records", createReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createBody map[string]any
	err = createResp.JSON(&createBody)
	require.NoError(t, err)

	recordID, ok := createBody["id"].(string)
	require.True(t, ok, "Response id field should be a string")
	require.NotEmpty(t, recordID)

	// 作成した記録を削除する
	deleteResp, err := client.Request(ctx, http.MethodDelete, "/api/exercise/records/"+recordID, nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	// 削除後は日次取得で対象レコードが存在しないことを確認（GETエンドポイント経由）
	date := time.Now().Format("2006-01-02")
	getResp, err := client.Get(ctx, "/api/exercise/daily?date="+date)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, getResp.StatusCode)

	var getBody map[string]any
	err = getResp.JSON(&getBody)
	require.NoError(t, err)

	recordsArray, ok := getBody["records"].([]any)
	require.True(t, ok, "Response should contain records array")

	// 削除したIDが一覧に含まれていないことを確認
	for _, record := range recordsArray {
		recordMap, ok := record.(map[string]any)
		require.True(t, ok)
		if id, ok := recordMap["id"].(string); ok {
			assert.NotEqual(t, recordID, id, "Deleted record %s should not be in the list", recordID)
		}
	}
}

func TestExerciseRecords_Delete_InvalidIDFormat(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// UUID形式でないIDを渡す（ハンドラーでUUID検証し400を返す）
	deleteResp, err := client.Request(ctx, http.MethodDelete, "/api/exercise/records/not-a-uuid", nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, deleteResp.StatusCode)
}

func TestExerciseRecords_Delete_NotFound(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	deleteResp, err := client.Request(ctx, http.MethodDelete, "/api/exercise/records/00000000-0000-0000-0000-000000000000", nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, deleteResp.StatusCode)
}

func TestExerciseRecords_Delete_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := testClient.Request(ctx, http.MethodDelete, "/api/exercise/records/00000000-0000-0000-0000-000000000000", nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestExerciseRecords_GetDaily_Success(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	// テスト用の日付を設定（既存データと衝突しないよう過去日付を使用）
	testDate := "2020-01-15"

	exercises := []struct {
		name     string
		duration int
	}{
		{"柔術", 60},
		{"ボクシング", 45},
		{"ランニング", 30},
	}

	var createdIDs []string
	for _, ex := range exercises {
		req := map[string]any{
			"exercise_name":    ex.name,
			"duration_minutes": ex.duration,
			"recorded_date":    testDate,
		}
		resp, err := client.Post(ctx, "/api/exercise/records", req)
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

	// 日次取得
	getResp, err := client.Get(ctx, "/api/exercise/daily?date="+testDate)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, getResp.StatusCode)
	assert.Contains(t, getResp.Headers.Get("Content-Type"), "application/json")

	var getBody map[string]any
	err = getResp.JSON(&getBody)
	require.NoError(t, err)

	recordsArray, ok := getBody["records"].([]any)
	require.True(t, ok, "Response should contain records array")
	// 過去のE2Eテスト実行でレコードが蓄積する可能性があるため、件数ではなく作成したIDの存在を検証
	assert.GreaterOrEqual(t, len(recordsArray), 3, "Should return at least 3 records")

	// 合計消費カロリーが正の値であることを確認
	totalCalories, ok := getBody["total_burned_calories_kcal"].(float64)
	require.True(t, ok, "total_burned_calories_kcal should be a float64")
	assert.Greater(t, totalCalories, 0.0, "total_burned_calories_kcal should be positive")

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

func TestExerciseRecords_GetDaily_EmptyResult(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// レコードが存在しない過去日付
	getResp, err := client.Get(ctx, "/api/exercise/daily?date=2000-01-01")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, getResp.StatusCode)
	assert.Contains(t, getResp.Headers.Get("Content-Type"), "application/json")

	var getBody map[string]any
	err = getResp.JSON(&getBody)
	require.NoError(t, err)

	recordsArray, ok := getBody["records"].([]any)
	require.True(t, ok, "Response should contain records array")
	assert.Empty(t, recordsArray, "Records should be empty for date with no records")

	totalCalories, ok := getBody["total_burned_calories_kcal"].(float64)
	require.True(t, ok, "total_burned_calories_kcal should be a float64")
	assert.Equal(t, 0.0, totalCalories, "total_burned_calories_kcal should be 0 when no records")
}

func TestExerciseRecords_GetDaily_MissingDate(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	resp, err := client.Get(ctx, "/api/exercise/daily")
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestExerciseRecords_GetDaily_InvalidDate_Format(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	resp, err := client.Get(ctx, "/api/exercise/daily?date=2024-02-13T00:00:00Z")
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestExerciseRecords_GetDaily_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := testClient.Get(ctx, "/api/exercise/daily?date=2024-01-01")
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
