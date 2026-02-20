package repository

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/iterator"
)

// cleanupUserCollection はテスト用の指定コレクションデータをクリーンアップする
func cleanupUserCollection(t *testing.T, ctx context.Context, client *firestore.Client, userID, collection string) {
	t.Helper()
	iter := client.Collection("users").Doc(userID).Collection(collection).Documents(ctx)
	defer iter.Stop()

	bw := client.BulkWriter(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			t.Logf("クリーンアップ中にドキュメント取得エラー (%s): %v", collection, err)
			break
		}
		bw.Delete(doc.Ref)
	}
	bw.Flush()
	bw.End()
}

// newSampleCreateInput はテスト用のCreateMenuSuggestionInputを生成するヘルパー
func newSampleCreateInput(mealType string, ingredientID string) CreateMenuSuggestionInput {
	ings := []MenuSuggestionIngredient{}
	if ingredientID != "" {
		ings = append(ings, MenuSuggestionIngredient{
			IngredientID: ingredientID,
			Name:         "鶏むね肉",
			Quantity:     200,
			Unit:         "g",
		})
	}
	return CreateMenuSuggestionInput{
		Title:           "テストメニュー_" + mealType,
		Description:     "テスト用の説明",
		Reason:          "テスト用の提案理由",
		IngredientsUsed: ings,
		EstimatedNutrition: EstimatedNutrition{
			Calories:      350,
			Protein:       40,
			Fat:           8,
			Carbohydrates: 15,
		},
		MealType: mealType,
	}
}

// createIngredientDoc はテスト用の食材ドキュメントをFirestoreに直接作成する
func createIngredientDoc(ctx context.Context, client *firestore.Client, userID, ingredientID, name string, quantity float64) error {
	now := time.Now()
	doc := firestoreIngredientDocument{
		ID:        ingredientID,
		Name:      name,
		Category:  "meat",
		Quantity:  quantity,
		Unit:      "g",
		Source:    "manual",
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err := client.Collection("users").Doc(userID).Collection("ingredients").Doc(ingredientID).Set(ctx, doc)
	return err
}

// --- NewMenuSuggestionRepository ---

func TestNewMenuSuggestionRepository(t *testing.T) {
	t.Run("nilクライアントでエラー", func(t *testing.T) {
		repo, err := NewMenuSuggestionRepository(nil)
		require.Error(t, err)
		assert.Nil(t, repo)
	})
}

// --- Create ---

func TestMenuSuggestionRepository_Create(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-menu-user-" + uuid.New().String()

	t.Cleanup(func() {
		cleanupUserCollection(t, ctx, client, userID, "menuSuggestions")
	})

	repo, err := NewMenuSuggestionRepository(client)
	require.NoError(t, err)

	t.Run("正常に作成できる", func(t *testing.T) {
		input := newSampleCreateInput("lunch", "")

		result, err := repo.Create(ctx, userID, input)

		require.NoError(t, err)
		assert.NotEmpty(t, result.ID)
		assert.Equal(t, "テストメニュー_lunch", result.Title)
		assert.Equal(t, "テスト用の説明", result.Description)
		assert.Equal(t, "テスト用の提案理由", result.Reason)
		assert.Equal(t, "lunch", result.MealType)
		assert.Equal(t, string(MenuStatusSuggested), result.Status)
		assert.Equal(t, float64(350), result.EstimatedNutrition.Calories)
		assert.Equal(t, float64(40), result.EstimatedNutrition.Protein)
		assert.False(t, result.CreatedAt.IsZero())
	})

	t.Run("食材ありで作成できる", func(t *testing.T) {
		input := newSampleCreateInput("dinner", "ing-test-1")

		result, err := repo.Create(ctx, userID, input)

		require.NoError(t, err)
		require.Len(t, result.IngredientsUsed, 1)
		assert.Equal(t, "ing-test-1", result.IngredientsUsed[0].IngredientID)
		assert.Equal(t, "鶏むね肉", result.IngredientsUsed[0].Name)
		assert.Equal(t, float64(200), result.IngredientsUsed[0].Quantity)
	})

	t.Run("userIDが空の場合エラー", func(t *testing.T) {
		input := newSampleCreateInput("lunch", "")
		_, err := repo.Create(ctx, "", input)
		require.Error(t, err)
	})
}

// --- List ---

func TestMenuSuggestionRepository_List(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-menu-list-user-" + uuid.New().String()

	t.Cleanup(func() {
		cleanupUserCollection(t, ctx, client, userID, "menuSuggestions")
	})

	repo, err := NewMenuSuggestionRepository(client)
	require.NoError(t, err)

	// テストデータを3件作成
	for _, mealType := range []string{"breakfast", "lunch", "dinner"} {
		_, err := repo.Create(ctx, userID, newSampleCreateInput(mealType, ""))
		require.NoError(t, err)
	}

	t.Run("ステータスなしで全件取得できる", func(t *testing.T) {
		items, err := repo.List(ctx, userID, "", 10)
		require.NoError(t, err)
		assert.Len(t, items, 3)
	})

	t.Run("suggestedステータスでフィルタできる", func(t *testing.T) {
		items, err := repo.List(ctx, userID, string(MenuStatusSuggested), 10)
		require.NoError(t, err)
		assert.Len(t, items, 3)
		for _, item := range items {
			assert.Equal(t, string(MenuStatusSuggested), item.Status)
		}
	})

	t.Run("acceptedステータスでフィルタすると0件", func(t *testing.T) {
		items, err := repo.List(ctx, userID, string(MenuStatusAccepted), 10)
		require.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("dismissedステータスでフィルタできる", func(t *testing.T) {
		dismissedInput := newSampleCreateInput("snack", "")
		created, err := repo.Create(ctx, userID, dismissedInput)
		require.NoError(t, err)
		err = repo.Dismiss(ctx, userID, created.ID)
		require.NoError(t, err)

		items, err := repo.List(ctx, userID, string(MenuStatusDismissed), 10)
		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, string(MenuStatusDismissed), items[0].Status)
	})

	t.Run("limit=1で1件のみ取得できる", func(t *testing.T) {
		items, err := repo.List(ctx, userID, "", 1)
		require.NoError(t, err)
		assert.Len(t, items, 1)
	})

	t.Run("userIDが空の場合エラー", func(t *testing.T) {
		_, err := repo.List(ctx, "", "", 10)
		require.Error(t, err)
	})
}

// --- GetByID ---

func TestMenuSuggestionRepository_GetByID(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-menu-get-user-" + uuid.New().String()

	t.Cleanup(func() {
		cleanupUserCollection(t, ctx, client, userID, "menuSuggestions")
	})

	repo, err := NewMenuSuggestionRepository(client)
	require.NoError(t, err)

	created, err := repo.Create(ctx, userID, newSampleCreateInput("lunch", ""))
	require.NoError(t, err)

	t.Run("正常に取得できる", func(t *testing.T) {
		result, err := repo.GetByID(ctx, userID, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, result.ID)
		assert.Equal(t, created.Title, result.Title)
		assert.Equal(t, string(MenuStatusSuggested), result.Status)
	})

	t.Run("存在しないIDでErrNotFound", func(t *testing.T) {
		_, err := repo.GetByID(ctx, userID, "nonexistent-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("userIDが空の場合エラー", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "", created.ID)
		require.Error(t, err)
	})

	t.Run("IDが空の場合エラー", func(t *testing.T) {
		_, err := repo.GetByID(ctx, userID, "")
		require.Error(t, err)
	})
}

// --- UpdateRecipe ---

func TestMenuSuggestionRepository_UpdateRecipe(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-menu-recipe-user-" + uuid.New().String()

	t.Cleanup(func() {
		cleanupUserCollection(t, ctx, client, userID, "menuSuggestions")
	})

	repo, err := NewMenuSuggestionRepository(client)
	require.NoError(t, err)

	created, err := repo.Create(ctx, userID, newSampleCreateInput("dinner", ""))
	require.NoError(t, err)

	t.Run("正常にレシピを更新できる", func(t *testing.T) {
		recipe := "1. 材料を切る\n2. 炒める\n3. 味付けする"
		err := repo.UpdateRecipe(ctx, userID, created.ID, recipe)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, userID, created.ID)
		require.NoError(t, err)
		assert.Equal(t, recipe, updated.Recipe)
	})

	t.Run("存在しないIDでErrNotFound", func(t *testing.T) {
		err := repo.UpdateRecipe(ctx, userID, "nonexistent-id", "レシピ")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("userIDが空の場合エラー", func(t *testing.T) {
		err := repo.UpdateRecipe(ctx, "", created.ID, "レシピ")
		require.Error(t, err)
	})
}

// --- Dismiss ---

func TestMenuSuggestionRepository_Dismiss(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-menu-dismiss-user-" + uuid.New().String()

	t.Cleanup(func() {
		cleanupUserCollection(t, ctx, client, userID, "menuSuggestions")
	})

	repo, err := NewMenuSuggestionRepository(client)
	require.NoError(t, err)

	t.Run("正常にdismissできる", func(t *testing.T) {
		created, err := repo.Create(ctx, userID, newSampleCreateInput("snack", ""))
		require.NoError(t, err)

		err = repo.Dismiss(ctx, userID, created.ID)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, userID, created.ID)
		require.NoError(t, err)
		assert.Equal(t, string(MenuStatusDismissed), updated.Status)
	})

	t.Run("存在しないIDでNotFound", func(t *testing.T) {
		err := repo.Dismiss(ctx, userID, "nonexistent-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("既にdismissされている場合ErrAlreadyProcessed", func(t *testing.T) {
		created, err := repo.Create(ctx, userID, newSampleCreateInput("breakfast", ""))
		require.NoError(t, err)

		// 1回目は成功
		err = repo.Dismiss(ctx, userID, created.ID)
		require.NoError(t, err)

		// 2回目はErrAlreadyProcessed
		err = repo.Dismiss(ctx, userID, created.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAlreadyProcessed)
	})

	t.Run("acceptedは却下できない（ErrAlreadyProcessed）", func(t *testing.T) {
		created, err := repo.Create(ctx, userID, newSampleCreateInput("lunch", ""))
		require.NoError(t, err)

		// ステータスを直接acceptedに変更
		_, err = client.Collection("users").Doc(userID).Collection("menuSuggestions").Doc(created.ID).Update(ctx, []firestore.Update{
			{Path: "status", Value: string(MenuStatusAccepted)},
		})
		require.NoError(t, err)

		err = repo.Dismiss(ctx, userID, created.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAlreadyProcessed)
	})
}

// --- Accept ---

func TestMenuSuggestionRepository_Accept(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-menu-accept-user-" + uuid.New().String()

	t.Cleanup(func() {
		cleanupUserCollection(t, ctx, client, userID, "menuSuggestions")
		cleanupUserCollection(t, ctx, client, userID, "ingredients")
		cleanupUserCollection(t, ctx, client, userID, "analysisRequests")
	})

	repo, err := NewMenuSuggestionRepository(client)
	require.NoError(t, err)

	t.Run("食材なしで正常にacceptできる", func(t *testing.T) {
		input := newSampleCreateInput("lunch", "")
		created, err := repo.Create(ctx, userID, input)
		require.NoError(t, err)

		result, err := repo.Accept(ctx, userID, created.ID)
		require.NoError(t, err)
		assert.NotEmpty(t, result.AnalysisRequestID)
		assert.Empty(t, result.DeductedIngredients)

		// ステータスがacceptedになっている
		updated, err := repo.GetByID(ctx, userID, created.ID)
		require.NoError(t, err)
		assert.Equal(t, string(MenuStatusAccepted), updated.Status)

		// analysisRequestsにドキュメントが作成されている
		analysisDoc, err := client.Collection("users").Doc(userID).Collection("analysisRequests").Doc(result.AnalysisRequestID).Get(ctx)
		require.NoError(t, err)
		assert.True(t, analysisDoc.Exists())
		data := analysisDoc.Data()
		assert.Equal(t, "completed", data["status"])
		assert.Equal(t, "suggestion", data["inputType"])
		assert.Equal(t, created.Title, data["inputText"])
		assert.Equal(t, true, data["confirmed"])
	})

	t.Run("食材ありで部分控除できる", func(t *testing.T) {
		ingID := uuid.New().String()
		// 食材を500g用意
		require.NoError(t, createIngredientDoc(ctx, client, userID, ingID, "鶏むね肉", 500))

		input := newSampleCreateInput("dinner", ingID) // 200g使用
		created, err := repo.Create(ctx, userID, input)
		require.NoError(t, err)

		result, err := repo.Accept(ctx, userID, created.ID)
		require.NoError(t, err)
		require.Len(t, result.DeductedIngredients, 1)
		assert.Equal(t, ingID, result.DeductedIngredients[0].IngredientID)
		assert.Equal(t, float64(200), result.DeductedIngredients[0].Deducted)
		assert.Equal(t, float64(300), result.DeductedIngredients[0].Remaining) // 500 - 200 = 300

		// 食材の在庫が減っている
		ingDoc, err := client.Collection("users").Doc(userID).Collection("ingredients").Doc(ingID).Get(ctx)
		require.NoError(t, err)
		data := ingDoc.Data()
		assert.Equal(t, float64(300), data["quantity"])
	})

	t.Run("食材を使い切った場合（境界値: 数量がちょうど0）ドキュメント削除", func(t *testing.T) {
		ingID := uuid.New().String()
		// 食材を200g用意（サジェストと同量）
		require.NoError(t, createIngredientDoc(ctx, client, userID, ingID, "鶏むね肉", 200))

		input := newSampleCreateInput("breakfast", ingID) // 200g使用
		created, err := repo.Create(ctx, userID, input)
		require.NoError(t, err)

		result, err := repo.Accept(ctx, userID, created.ID)
		require.NoError(t, err)
		require.Len(t, result.DeductedIngredients, 1)
		assert.Equal(t, float64(200), result.DeductedIngredients[0].Deducted)
		assert.Equal(t, float64(0), result.DeductedIngredients[0].Remaining)

		// 食材ドキュメントが削除されている
		ingDoc, err := client.Collection("users").Doc(userID).Collection("ingredients").Doc(ingID).Get(ctx)
		require.NoError(t, err)
		assert.False(t, ingDoc.Exists())
	})

	t.Run("食材を超過した場合（境界値: 数量が負になる）ドキュメント削除", func(t *testing.T) {
		ingID := uuid.New().String()
		// 食材を100gのみ用意（サジェストは200g使用）
		require.NoError(t, createIngredientDoc(ctx, client, userID, ingID, "鶏むね肉", 100))

		input := newSampleCreateInput("snack", ingID) // 200g使用（100g不足）
		created, err := repo.Create(ctx, userID, input)
		require.NoError(t, err)

		result, err := repo.Accept(ctx, userID, created.ID)
		require.NoError(t, err)
		require.Len(t, result.DeductedIngredients, 1)
		// 実際にあった分だけ控除（100g）されてRemainingは0
		assert.Equal(t, float64(100), result.DeductedIngredients[0].Deducted)
		assert.Equal(t, float64(0), result.DeductedIngredients[0].Remaining)

		// 食材ドキュメントが削除されている
		ingDoc, err := client.Collection("users").Doc(userID).Collection("ingredients").Doc(ingID).Get(ctx)
		require.NoError(t, err)
		assert.False(t, ingDoc.Exists())
	})

	t.Run("食材が存在しない場合スキップして控除なし", func(t *testing.T) {
		nonExistentIngID := uuid.New().String()
		input := CreateMenuSuggestionInput{
			Title:    "食材なしメニュー",
			MealType: "lunch",
			IngredientsUsed: []MenuSuggestionIngredient{
				{IngredientID: nonExistentIngID, Name: "存在しない食材", Quantity: 100, Unit: "g"},
			},
			EstimatedNutrition: EstimatedNutrition{Calories: 200},
		}
		created, err := repo.Create(ctx, userID, input)
		require.NoError(t, err)

		result, err := repo.Accept(ctx, userID, created.ID)
		require.NoError(t, err)
		require.Len(t, result.DeductedIngredients, 1)
		// 食材が存在しないためDeductedは0、Remainingは0
		assert.Equal(t, float64(0), result.DeductedIngredients[0].Deducted)
		assert.Equal(t, float64(0), result.DeductedIngredients[0].Remaining)
	})

	t.Run("IngredientIDが空の場合スキップ", func(t *testing.T) {
		input := CreateMenuSuggestionInput{
			Title:    "ID空メニュー",
			MealType: "dinner",
			IngredientsUsed: []MenuSuggestionIngredient{
				{IngredientID: "", Name: "ID未設定食材", Quantity: 100, Unit: "g"},
			},
			EstimatedNutrition: EstimatedNutrition{Calories: 300},
		}
		created, err := repo.Create(ctx, userID, input)
		require.NoError(t, err)

		result, err := repo.Accept(ctx, userID, created.ID)
		require.NoError(t, err)
		// IngredientIDが空の場合は控除リストに含まれない
		assert.Empty(t, result.DeductedIngredients)
	})

	t.Run("既にacceptedの場合ErrAlreadyProcessed", func(t *testing.T) {
		created, err := repo.Create(ctx, userID, newSampleCreateInput("breakfast", ""))
		require.NoError(t, err)

		// 1回目は成功
		_, err = repo.Accept(ctx, userID, created.ID)
		require.NoError(t, err)

		// 2回目はErrAlreadyProcessed
		_, err = repo.Accept(ctx, userID, created.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAlreadyProcessed)
	})

	t.Run("dismissedはacceptできない（ErrAlreadyProcessed）", func(t *testing.T) {
		created, err := repo.Create(ctx, userID, newSampleCreateInput("lunch", ""))
		require.NoError(t, err)

		// ステータスを直接dismissedに変更
		_, err = client.Collection("users").Doc(userID).Collection("menuSuggestions").Doc(created.ID).Update(ctx, []firestore.Update{
			{Path: "status", Value: string(MenuStatusDismissed)},
		})
		require.NoError(t, err)

		_, err = repo.Accept(ctx, userID, created.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrAlreadyProcessed)
	})

	t.Run("存在しないIDでNotFound", func(t *testing.T) {
		_, err := repo.Accept(ctx, userID, "nonexistent-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("userIDが空の場合エラー", func(t *testing.T) {
		_, err := repo.Accept(ctx, "", "some-id")
		require.Error(t, err)
	})
}
