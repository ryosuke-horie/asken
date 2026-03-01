//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testMyMenuFoods はテスト用の食品リストを返す
func testMyMenuFoods() []map[string]any {
	return []map[string]any{
		{
			"name":             "白米",
			"estimated_amount": "150g",
			"calories_kcal":    252.0,
			"protein_g":        3.8,
			"fat_g":            0.5,
			"carbohydrates_g":  55.7,
		},
	}
}

// createTestMyMenu はテスト用のマイメニューを作成し、IDを返す
func createTestMyMenu(t *testing.T, client *Client, ctx context.Context) string {
	t.Helper()

	reqBody := map[string]any{
		"name":  "テスト定食",
		"foods": testMyMenuFoods(),
	}

	resp, err := client.Post(ctx, "/api/my-menu", reqBody)
	if err != nil {
		t.Fatalf("createTestMyMenu: failed to create my menu: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("createTestMyMenu: expected 201, got %d (body: %s)", resp.StatusCode, string(resp.Body))
	}

	var body map[string]any
	if err := resp.JSON(&body); err != nil {
		t.Fatalf("createTestMyMenu: failed to parse response: %v", err)
	}

	id, ok := body["id"].(string)
	if !ok || id == "" {
		t.Fatalf("createTestMyMenu: id is empty or not a string")
	}
	return id
}

// === List ===

func TestMyMenu_List_Success(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	// 事前に1件作成してIDを保存
	menuID := createTestMyMenu(t, client, ctx)

	resp, err := client.Get(ctx, "/api/my-menu")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body []any
	err = resp.JSON(&body)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(body), 1, "Response should contain at least 1 item")

	// 作成したメニューのIDが一覧に含まれることを確認
	foundID := false
	for _, item := range body {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if id, ok := itemMap["id"].(string); ok && id == menuID {
			foundID = true
			break
		}
	}
	assert.True(t, foundID, "Created menu %s should be in the list", menuID)
}

func TestMyMenu_List_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := testClient.Get(ctx, "/api/my-menu")
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// === Create ===

func TestMyMenu_Create_Success(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"name":  "テスト定食",
		"foods": testMyMenuFoods(),
	}

	resp, err := client.Post(ctx, "/api/my-menu", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]any
	err = resp.JSON(&body)
	require.NoError(t, err)

	assert.NotEmpty(t, body["id"], "Response should contain id")
	assert.Equal(t, "テスト定食", body["name"])
	assert.NotEmpty(t, body["createdAt"])
	assert.NotEmpty(t, body["updatedAt"])
}

func TestMyMenu_Create_EmptyName(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"name":  "",
		"foods": testMyMenuFoods(),
	}

	resp, err := client.Post(ctx, "/api/my-menu", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestMyMenu_Create_EmptyFoods(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"name":  "テスト定食",
		"foods": []any{},
	}

	resp, err := client.Post(ctx, "/api/my-menu", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestMyMenu_Create_NameTooLong(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// 51文字のメニュー名（上限50文字を超える）
	longName := strings.Repeat("あ", 51)
	reqBody := map[string]any{
		"name":  longName,
		"foods": testMyMenuFoods(),
	}

	resp, err := client.Post(ctx, "/api/my-menu", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestMyMenu_Create_TooManyFoods(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// 101件の食品リスト（上限100件を超える）
	foods := make([]map[string]any, 101)
	for i := range foods {
		foods[i] = map[string]any{
			"name":            fmt.Sprintf("食品%d", i),
			"calories_kcal":   100.0,
			"protein_g":       5.0,
			"fat_g":           3.0,
			"carbohydrates_g": 10.0,
		}
	}
	reqBody := map[string]any{
		"name":  "テスト定食",
		"foods": foods,
	}

	resp, err := client.Post(ctx, "/api/my-menu", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestMyMenu_Create_NameExactlyMaxLength(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// 50文字（上限ちょうど）のメニュー名は受け入れられるべき
	name := strings.Repeat("あ", 50)
	reqBody := map[string]any{
		"name":  name,
		"foods": testMyMenuFoods(),
	}

	resp, err := client.Post(ctx, "/api/my-menu", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestMyMenu_Create_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqBody := map[string]any{
		"name":  "テスト定食",
		"foods": testMyMenuFoods(),
	}

	resp, err := testClient.Post(ctx, "/api/my-menu", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// === Get ===

func TestMyMenu_Get_Success(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	menuID := createTestMyMenu(t, client, ctx)

	resp, err := client.Get(ctx, "/api/my-menu/"+menuID)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]any
	err = resp.JSON(&body)
	require.NoError(t, err)

	assert.Equal(t, menuID, body["id"])
	assert.Equal(t, "テスト定食", body["name"])
	assert.NotEmpty(t, body["foods"])
}

func TestMyMenu_Get_NotFound(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	resp, err := client.Get(ctx, "/api/my-menu/00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMyMenu_Get_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := testClient.Get(ctx, "/api/my-menu/00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// === Update ===

func TestMyMenu_Update_Success(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	menuID := createTestMyMenu(t, client, ctx)

	reqBody := map[string]any{
		"name": "更新後の定食",
		"foods": []map[string]any{
			{
				"name":             "焼き鮭",
				"estimated_amount": "80g",
				"calories_kcal":    149.0,
				"protein_g":        22.3,
				"fat_g":            5.0,
				"carbohydrates_g":  0.1,
			},
		},
	}

	resp, err := client.Request(ctx, http.MethodPut, "/api/my-menu/"+menuID, reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]any
	err = resp.JSON(&body)
	require.NoError(t, err)

	assert.Equal(t, menuID, body["id"])
	assert.Equal(t, "更新後の定食", body["name"])
}

func TestMyMenu_Update_NotFound(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"name":  "更新後の定食",
		"foods": testMyMenuFoods(),
	}

	resp, err := client.Request(ctx, http.MethodPut, "/api/my-menu/00000000-0000-0000-0000-000000000000", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMyMenu_Update_EmptyName(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// バリデーションは存在確認より先に行われるため、実在しないIDでも400が返る
	reqBody := map[string]any{
		"name":  "",
		"foods": testMyMenuFoods(),
	}

	resp, err := client.Request(ctx, http.MethodPut, "/api/my-menu/00000000-0000-0000-0000-000000000000", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestMyMenu_Update_EmptyFoods(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// バリデーションは存在確認より先に行われるため、実在しないIDでも400が返る
	reqBody := map[string]any{
		"name":  "テスト定食",
		"foods": []any{},
	}

	resp, err := client.Request(ctx, http.MethodPut, "/api/my-menu/00000000-0000-0000-0000-000000000000", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestMyMenu_Update_NameTooLong(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	// バリデーションは存在確認より先に行われるため、実在しないIDでも400が返る
	longName := strings.Repeat("あ", 51)
	reqBody := map[string]any{
		"name":  longName,
		"foods": testMyMenuFoods(),
	}

	resp, err := client.Request(ctx, http.MethodPut, "/api/my-menu/00000000-0000-0000-0000-000000000000", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestMyMenu_Update_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqBody := map[string]any{
		"name":  "更新後の定食",
		"foods": testMyMenuFoods(),
	}

	resp, err := testClient.Request(ctx, http.MethodPut, "/api/my-menu/00000000-0000-0000-0000-000000000000", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// === Delete ===

func TestMyMenu_Delete_Success(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	menuID := createTestMyMenu(t, client, ctx)

	deleteResp, err := client.Request(ctx, http.MethodDelete, "/api/my-menu/"+menuID, nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

	// 削除後にGetして404になることを確認
	getResp, err := client.Get(ctx, "/api/my-menu/"+menuID)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
}

func TestMyMenu_Delete_NotFound(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	resp, err := client.Request(ctx, http.MethodDelete, "/api/my-menu/00000000-0000-0000-0000-000000000000", nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMyMenu_Delete_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := testClient.Request(ctx, http.MethodDelete, "/api/my-menu/00000000-0000-0000-0000-000000000000", nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// === Record ===

func TestMyMenu_Record_Success(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	menuID := createTestMyMenu(t, client, ctx)

	reqBody := map[string]any{
		"meal_type": "lunch",
		"meal_date": time.Now().Format("2006-01-02"),
	}

	resp, err := client.Post(ctx, "/api/my-menu/"+menuID+"/record", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]any
	err = resp.JSON(&body)
	require.NoError(t, err)

	assert.NotEmpty(t, body["id"], "Response should contain analysis id")
}

func TestMyMenu_Record_NotFound(t *testing.T) {
	client, ctx := authenticatedClient(t, 30*time.Second)

	reqBody := map[string]any{
		"meal_type": "lunch",
		"meal_date": time.Now().Format("2006-01-02"),
	}

	resp, err := client.Post(ctx, "/api/my-menu/00000000-0000-0000-0000-000000000000/record", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMyMenu_Record_InvalidMealType(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	menuID := createTestMyMenu(t, client, ctx)

	reqBody := map[string]any{
		"meal_type": "invalid_type",
		"meal_date": time.Now().Format("2006-01-02"),
	}

	resp, err := client.Post(ctx, "/api/my-menu/"+menuID+"/record", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestMyMenu_Record_MissingMealType(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	menuID := createTestMyMenu(t, client, ctx)

	reqBody := map[string]any{
		"meal_date": time.Now().Format("2006-01-02"),
	}

	resp, err := client.Post(ctx, "/api/my-menu/"+menuID+"/record", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestMyMenu_Record_InvalidMealDate(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	menuID := createTestMyMenu(t, client, ctx)

	// YYYY-MM-DD 形式でない日付
	reqBody := map[string]any{
		"meal_type": "lunch",
		"meal_date": "2024/01/15",
	}

	resp, err := client.Post(ctx, "/api/my-menu/"+menuID+"/record", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestMyMenu_Record_DefaultMealDate(t *testing.T) {
	waitForUserRateLimit()

	client, ctx := authenticatedClient(t, 30*time.Second)

	menuID := createTestMyMenu(t, client, ctx)

	// meal_date を省略（今日の日付がデフォルトとして使用される）
	reqBody := map[string]any{
		"meal_type": "dinner",
	}

	resp, err := client.Post(ctx, "/api/my-menu/"+menuID+"/record", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Contains(t, resp.Headers.Get("Content-Type"), "application/json")

	var body map[string]any
	err = resp.JSON(&body)
	require.NoError(t, err)

	assert.NotEmpty(t, body["id"], "Response should contain analysis id")
}

func TestMyMenu_Record_Unauthorized(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqBody := map[string]any{
		"meal_type": "lunch",
		"meal_date": time.Now().Format("2006-01-02"),
	}

	resp, err := testClient.Post(ctx, "/api/my-menu/00000000-0000-0000-0000-000000000000/record", reqBody)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
