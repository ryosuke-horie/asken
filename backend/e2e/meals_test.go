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

func TestMeals_Daily_Success(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// 今日の日付を指定
	today := time.Now().Format("2006-01-02")
	resp, err := client.Get(ctx, "/api/meals/daily?date="+today+"&tz=Asia/Tokyo")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]any
	err = resp.JSON(&body)
	require.NoError(t, err)

	assert.NotEmpty(t, body["date"], "Response should contain date")
	assert.Contains(t, body, "meals", "Response should contain meals")
	assert.Contains(t, body, "daily_total", "Response should contain daily_total")
}

func TestMeals_Daily_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := testClient.Get(ctx, "/api/meals/daily")
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestMeals_Skip_Success(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]string{
		"meal_type": "lunch",
		"meal_date": time.Now().Format("2006-01-02"),
	}

	resp, err := client.Post(ctx, "/api/meals/skip", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]string
	err = resp.JSON(&body)
	require.NoError(t, err)

	assert.NotEmpty(t, body["id"], "Response should contain id")
}

func TestMeals_Skip_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqBody := map[string]string{
		"meal_type": "lunch",
		"meal_date": time.Now().Format("2006-01-02"),
	}

	resp, err := testClient.Post(ctx, "/api/meals/skip", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestMeals_Skip_InvalidMealType(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]string{
		"meal_type": "invalid_type",
		"meal_date": time.Now().Format("2006-01-02"),
	}

	resp, err := client.Post(ctx, "/api/meals/skip", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestMeals_Skip_MissingMealType(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]string{
		"meal_date": time.Now().Format("2006-01-02"),
	}

	resp, err := client.Post(ctx, "/api/meals/skip", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
