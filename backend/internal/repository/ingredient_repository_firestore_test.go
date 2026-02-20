package repository

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/iterator"
)

// cleanupIngredientTestData はテスト用の食材データをクリーンアップします
func cleanupIngredientTestData(ctx context.Context, client *firestore.Client, userID string) error {
	iter := client.Collection("users").Doc(userID).Collection("ingredients").Documents(ctx)
	defer iter.Stop()

	bw := client.BulkWriter(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		bw.Delete(doc.Ref)
	}
	bw.Flush()
	bw.End()
	return nil
}

// newCreateInput はテスト用のCreateIngredientInputを生成するヘルパー
func newCreateInput(name, category, unit, source string, quantity float64) CreateIngredientInput {
	return CreateIngredientInput{
		Name:     name,
		Category: category,
		Quantity: quantity,
		Unit:     unit,
		Source:   source,
	}
}

func TestNewIngredientRepository(t *testing.T) {
	t.Run("nilクライアントでエラー", func(t *testing.T) {
		repo, err := NewIngredientRepository(nil)
		assert.Error(t, err)
		assert.Nil(t, repo)
	})
}

func TestIngredientRepository_Create(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-ingredient-user-create"

	t.Cleanup(func() {
		cleanupIngredientTestData(ctx, client, userID)
	})

	repo, err := NewIngredientRepository(client)
	require.NoError(t, err)

	t.Run("正常に食材を作成できる", func(t *testing.T) {
		input := newCreateInput("鶏むね肉", "meat", "g", "manual", 500)

		item, err := repo.Create(ctx, userID, input)

		require.NoError(t, err)
		assert.NotEmpty(t, item.ID)
		assert.Equal(t, "鶏むね肉", item.Name)
		assert.Equal(t, "meat", item.Category)
		assert.Equal(t, float64(500), item.Quantity)
		assert.Equal(t, "g", item.Unit)
		assert.Equal(t, "manual", item.Source)
		assert.False(t, item.CreatedAt.IsZero())
		assert.False(t, item.UpdatedAt.IsZero())
		assert.Nil(t, item.PurchaseDate)
		assert.Nil(t, item.ExpiryDate)
	})

	t.Run("購入日・消費期限付きで食材を作成できる", func(t *testing.T) {
		purchaseDate := time.Date(2026, 2, 18, 0, 0, 0, 0, time.UTC)
		expiryDate := time.Date(2026, 2, 22, 0, 0, 0, 0, time.UTC)
		input := CreateIngredientInput{
			Name:         "牛乳",
			Category:     "dairy",
			Quantity:     1000,
			Unit:         "ml",
			Source:       "receipt",
			PurchaseDate: &purchaseDate,
			ExpiryDate:   &expiryDate,
		}

		item, err := repo.Create(ctx, userID, input)

		require.NoError(t, err)
		assert.NotNil(t, item.PurchaseDate)
		assert.NotNil(t, item.ExpiryDate)
		assert.Equal(t, 2026, item.PurchaseDate.Year())
		assert.Equal(t, 2026, item.ExpiryDate.Year())
	})

	t.Run("空のuserIDでエラー", func(t *testing.T) {
		input := newCreateInput("テスト", "meat", "g", "manual", 100)
		_, err := repo.Create(ctx, "", input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "userIDが必要です")
	})
}

func TestIngredientRepository_List(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-ingredient-user-list"

	t.Cleanup(func() {
		cleanupIngredientTestData(ctx, client, userID)
	})

	repo, err := NewIngredientRepository(client)
	require.NoError(t, err)

	// テストデータ投入
	_, err = repo.Create(ctx, userID, newCreateInput("鶏むね肉", "meat", "g", "manual", 500))
	require.NoError(t, err)
	_, err = repo.Create(ctx, userID, newCreateInput("豚バラ", "meat", "g", "receipt", 300))
	require.NoError(t, err)
	_, err = repo.Create(ctx, userID, newCreateInput("牛乳", "dairy", "ml", "manual", 1000))
	require.NoError(t, err)

	t.Run("全件取得できる", func(t *testing.T) {
		items, err := repo.List(ctx, userID, "")
		require.NoError(t, err)
		assert.Len(t, items, 3)
	})

	t.Run("カテゴリでフィルタできる", func(t *testing.T) {
		items, err := repo.List(ctx, userID, "meat")
		require.NoError(t, err)
		assert.Len(t, items, 2)
		for _, item := range items {
			assert.Equal(t, "meat", item.Category)
		}
	})

	t.Run("カテゴリフィルタで名前昇順ソートされる", func(t *testing.T) {
		items, err := repo.List(ctx, userID, "meat")
		require.NoError(t, err)
		require.Len(t, items, 2)
		// 鶏むね肉 < 豚バラ（五十音順）
		assert.Equal(t, "鶏むね肉", items[0].Name)
		assert.Equal(t, "豚バラ", items[1].Name)
	})

	t.Run("存在しないカテゴリは空スライスを返す", func(t *testing.T) {
		items, err := repo.List(ctx, userID, "fish")
		require.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("空のuserIDでエラー", func(t *testing.T) {
		_, err := repo.List(ctx, "", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "userIDが必要です")
	})
}

func TestIngredientRepository_GetByID(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-ingredient-user-get"

	t.Cleanup(func() {
		cleanupIngredientTestData(ctx, client, userID)
	})

	repo, err := NewIngredientRepository(client)
	require.NoError(t, err)

	t.Run("作成した食材を取得できる", func(t *testing.T) {
		created, err := repo.Create(ctx, userID, newCreateInput("鶏むね肉", "meat", "g", "manual", 500))
		require.NoError(t, err)

		got, err := repo.GetByID(ctx, userID, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, "鶏むね肉", got.Name)
	})

	t.Run("存在しないIDでErrNotFound", func(t *testing.T) {
		_, err := repo.GetByID(ctx, userID, "nonexistent-id")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("空のuserIDでエラー", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "", "some-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "userIDが必要です")
	})
}

func TestIngredientRepository_Update(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-ingredient-user-update"

	t.Cleanup(func() {
		cleanupIngredientTestData(ctx, client, userID)
	})

	repo, err := NewIngredientRepository(client)
	require.NoError(t, err)

	t.Run("食材を正常更新できる", func(t *testing.T) {
		created, err := repo.Create(ctx, userID, newCreateInput("鶏むね肉", "meat", "g", "manual", 500))
		require.NoError(t, err)

		updateInput := UpdateIngredientInput{
			Name:     "鶏もも肉",
			Category: "meat",
			Quantity: 300,
			Unit:     "g",
		}
		updated, err := repo.Update(ctx, userID, created.ID, updateInput)
		require.NoError(t, err)
		assert.Equal(t, "鶏もも肉", updated.Name)
		assert.Equal(t, float64(300), updated.Quantity)

		// 再取得して確認
		got, err := repo.GetByID(ctx, userID, created.ID)
		require.NoError(t, err)
		assert.Equal(t, "鶏もも肉", got.Name)
		assert.Equal(t, float64(300), got.Quantity)
	})

	t.Run("消費期限を更新できる", func(t *testing.T) {
		created, err := repo.Create(ctx, userID, newCreateInput("牛乳", "dairy", "ml", "manual", 1000))
		require.NoError(t, err)

		expiryDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		updateInput := UpdateIngredientInput{
			Name:       "牛乳",
			Category:   "dairy",
			Quantity:   900,
			Unit:       "ml",
			ExpiryDate: &expiryDate,
		}
		updated, err := repo.Update(ctx, userID, created.ID, updateInput)
		require.NoError(t, err)
		assert.NotNil(t, updated.ExpiryDate)
		assert.Equal(t, 2026, updated.ExpiryDate.Year())
	})

	t.Run("存在しないIDでErrNotFound", func(t *testing.T) {
		updateInput := UpdateIngredientInput{Name: "テスト", Category: "meat", Quantity: 100, Unit: "g"}
		_, err := repo.Update(ctx, userID, "nonexistent-id", updateInput)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestIngredientRepository_Delete(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-ingredient-user-delete"

	t.Cleanup(func() {
		cleanupIngredientTestData(ctx, client, userID)
	})

	repo, err := NewIngredientRepository(client)
	require.NoError(t, err)

	t.Run("食材を正常削除できる", func(t *testing.T) {
		created, err := repo.Create(ctx, userID, newCreateInput("鶏むね肉", "meat", "g", "manual", 500))
		require.NoError(t, err)

		err = repo.Delete(ctx, userID, created.ID)
		require.NoError(t, err)

		// 削除後は取得できないことを確認
		_, err = repo.GetByID(ctx, userID, created.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("存在しないIDでErrNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, userID, "nonexistent-id")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("空のuserIDでエラー", func(t *testing.T) {
		err := repo.Delete(ctx, "", "some-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "userIDが必要です")
	})
}
