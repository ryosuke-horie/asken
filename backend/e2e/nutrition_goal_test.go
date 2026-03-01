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

func TestNutritionGoal_Set_Success(t *testing.T) {
	// 前のテストファイル群でユーザーレート制限バケットが枯渇している可能性があるため待機
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"target_calories": 2000.0,
	}

	resp, err := client.Request(ctx, http.MethodPut, "/api/nutrition/goal", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]any
	err = resp.JSON(&body)
	require.NoError(t, err)

	// target_calories が設定値と一致すること
	targetCalories, ok := body["target_calories"].(float64)
	require.True(t, ok, "target_calories should be a float64")
	assertFloat64Equal(t, 2000.0, targetCalories, "target_calories should match")

	// PFC が自動計算されること
	assert.NotNil(t, body["target_protein"], "target_protein should be present")
	assert.NotNil(t, body["target_fat"], "target_fat should be present")
	assert.NotNil(t, body["target_carbohydrates"], "target_carbohydrates should be present")

	// phase が設定されること
	phase, ok := body["phase"].(string)
	require.True(t, ok, "phase should be a string")
	assert.NotEmpty(t, phase, "phase should not be empty")

	// updated_at が設定されること
	assert.NotEmpty(t, body["updated_at"], "updated_at should be present")
}

func TestNutritionGoal_Get_Success(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	// 事前に目標を設定
	setReq := map[string]any{
		"target_calories": 1800.0,
	}
	setResp, err := client.Request(ctx, http.MethodPut, "/api/nutrition/goal", setReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, setResp.StatusCode)

	// 目標を取得
	getResp, err := client.Get(ctx, "/api/nutrition/goal")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, getResp.StatusCode)
	assert.Contains(t, getResp.Headers.Get("Content-Type"), "application/json")

	var body map[string]any
	err = getResp.JSON(&body)
	require.NoError(t, err)

	// goal フィールドが存在し、設定値が反映されていること
	goal, ok := body["goal"].(map[string]any)
	require.True(t, ok, "goal should be an object")

	targetCalories, ok := goal["target_calories"].(float64)
	require.True(t, ok, "target_calories should be a float64")
	assertFloat64Equal(t, 1800.0, targetCalories, "target_calories should match")
	assert.NotEmpty(t, goal["updated_at"], "updated_at should be present")
}

func TestNutritionGoal_Get_DefaultResponse(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	resp, err := client.Get(ctx, "/api/nutrition/goal")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]any
	err = resp.JSON(&body)
	require.NoError(t, err)

	// goal フィールドが存在すること（値は null または設定済みオブジェクト）
	_, exists := body["goal"]
	assert.True(t, exists, "Response should contain 'goal' field")
}

func TestNutritionGoal_Get_WithWeightParams_Success(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	// 事前に目標カロリーを設定
	setReq := map[string]any{
		"target_calories": 2200.0,
	}
	setResp, err := client.Request(ctx, http.MethodPut, "/api/nutrition/goal", setReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, setResp.StatusCode)

	// 減量期（current_weight > target_weight + 1.0）で取得
	getResp, err := client.Get(ctx, "/api/nutrition/goal?current_weight=80.0&target_weight=70.0")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	var body map[string]any
	err = getResp.JSON(&body)
	require.NoError(t, err)

	goal, ok := body["goal"].(map[string]any)
	require.True(t, ok, "goal should be an object")

	phase, ok := goal["phase"].(string)
	require.True(t, ok, "phase should be a string")
	assert.Equal(t, "weight_loss", phase, "phase should be weight_loss when current_weight > target_weight + 1.0")
}

func TestNutritionGoal_Set_InvalidCalories_TooLow(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"target_calories": 799.9, // 800.0未満は無効
	}

	resp, err := client.Request(ctx, http.MethodPut, "/api/nutrition/goal", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestNutritionGoal_Set_InvalidCalories_TooHigh(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"target_calories": 5000.1, // 5000.0超過は無効
	}

	resp, err := client.Request(ctx, http.MethodPut, "/api/nutrition/goal", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestNutritionGoal_Get_InvalidCurrentWeight_TooLow(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// current_weight が 20.0 未満は無効
	resp, err := client.Get(ctx, "/api/nutrition/goal?current_weight=19.9")
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestNutritionGoal_Get_InvalidCurrentWeight_TooHigh(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// current_weight が 300.0 超過は無効
	resp, err := client.Get(ctx, "/api/nutrition/goal?current_weight=300.1")
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestNutritionGoal_Get_InvalidTargetWeight_TooLow(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// target_weight が 20.0 未満は無効
	resp, err := client.Get(ctx, "/api/nutrition/goal?target_weight=19.9")
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestNutritionGoal_Get_InvalidTargetWeight_TooHigh(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// target_weight が 300.0 超過は無効
	resp, err := client.Get(ctx, "/api/nutrition/goal?target_weight=300.1")
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestNutritionGoal_Get_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := testClient.Get(ctx, "/api/nutrition/goal")
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestNutritionGoal_Set_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqBody := map[string]any{
		"target_calories": 2000.0,
	}

	resp, err := testClient.Request(ctx, http.MethodPut, "/api/nutrition/goal", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
