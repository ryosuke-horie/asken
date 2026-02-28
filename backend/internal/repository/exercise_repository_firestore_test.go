package repository

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/iterator"
)

// cleanupExerciseTestData はテスト用の運動記録データをクリーンアップします
func cleanupExerciseTestData(ctx context.Context, client *firestore.Client, userID string) error {
	iter := client.Collection("users").Doc(userID).Collection("exerciseRecords").Documents(ctx)
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

func TestNewExerciseRepository(t *testing.T) {
	t.Run("nilクライアントでエラー", func(t *testing.T) {
		repo, err := NewExerciseRepository(nil)
		assert.Error(t, err)
		assert.Nil(t, repo)
	})
}

func TestExerciseRepository_Create(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-exercise-user-create"

	t.Cleanup(func() {
		cleanupExerciseTestData(ctx, client, userID)
	})

	repo, err := NewExerciseRepository(client)
	require.NoError(t, err)

	t.Run("正常に運動記録を作成できる", func(t *testing.T) {
		input := CreateExerciseInput{
			ExerciseName:       "柔術",
			DurationMinutes:    90,
			BurnedCaloriesKcal: 486.0,
			EstimationMethod:   EstimationMethodMET,
			RecordedDate:       "2026-02-28",
		}
		record, err := repo.Create(ctx, userID, input)

		require.NoError(t, err)
		assert.NotEmpty(t, record.ID)
		assert.Equal(t, "柔術", record.ExerciseName)
		assert.Equal(t, 90, record.DurationMinutes)
		assert.Equal(t, 486.0, record.BurnedCaloriesKcal)
		assert.Equal(t, EstimationMethodMET, record.EstimationMethod)
		assert.Equal(t, "2026-02-28", record.RecordedDate)
		assert.False(t, record.CreatedAt.IsZero())
	})

	t.Run("userIDが空のときエラー", func(t *testing.T) {
		input := CreateExerciseInput{
			ExerciseName:       "柔術",
			DurationMinutes:    60,
			BurnedCaloriesKcal: 324.0,
			EstimationMethod:   EstimationMethodMET,
			RecordedDate:       "2026-02-28",
		}
		record, err := repo.Create(ctx, "", input)
		assert.Error(t, err)
		assert.Nil(t, record)
	})
}

func TestExerciseRepository_ListByDate(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-exercise-user-list"

	t.Cleanup(func() {
		cleanupExerciseTestData(ctx, client, userID)
	})

	repo, err := NewExerciseRepository(client)
	require.NoError(t, err)

	t.Run("指定日の運動記録を取得できる", func(t *testing.T) {
		// 同日に2件作成
		inputs := []CreateExerciseInput{
			{ExerciseName: "柔術", DurationMinutes: 90, BurnedCaloriesKcal: 486.0, EstimationMethod: EstimationMethodMET, RecordedDate: "2026-02-28"},
			{ExerciseName: "ランニング", DurationMinutes: 30, BurnedCaloriesKcal: 240.0, EstimationMethod: EstimationMethodMET, RecordedDate: "2026-02-28"},
		}
		for _, input := range inputs {
			_, err := repo.Create(ctx, userID, input)
			require.NoError(t, err)
		}

		// 別日に1件作成（取得対象外）
		_, err = repo.Create(ctx, userID, CreateExerciseInput{
			ExerciseName: "水泳", DurationMinutes: 60, BurnedCaloriesKcal: 378.0, EstimationMethod: EstimationMethodMET, RecordedDate: "2026-02-27",
		})
		require.NoError(t, err)

		records, err := repo.ListByDate(ctx, userID, "2026-02-28")
		require.NoError(t, err)
		assert.Len(t, records, 2)
		for _, r := range records {
			assert.Equal(t, "2026-02-28", r.RecordedDate)
		}
	})

	t.Run("記録がない日は空スライスを返す", func(t *testing.T) {
		records, err := repo.ListByDate(ctx, userID, "2099-01-01")
		require.NoError(t, err)
		assert.Empty(t, records)
	})

	t.Run("userIDが空のときエラー", func(t *testing.T) {
		records, err := repo.ListByDate(ctx, "", "2026-02-28")
		assert.Error(t, err)
		assert.Nil(t, records)
	})
}

func TestExerciseRepository_Delete(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-exercise-user-delete"

	t.Cleanup(func() {
		cleanupExerciseTestData(ctx, client, userID)
	})

	repo, err := NewExerciseRepository(client)
	require.NoError(t, err)

	t.Run("正常に削除できる", func(t *testing.T) {
		record, err := repo.Create(ctx, userID, CreateExerciseInput{
			ExerciseName: "柔術", DurationMinutes: 90, BurnedCaloriesKcal: 486.0, EstimationMethod: EstimationMethodMET, RecordedDate: "2026-02-28",
		})
		require.NoError(t, err)

		err = repo.Delete(ctx, userID, record.ID)
		require.NoError(t, err)

		// 削除後は一覧に出ないことを確認
		records, err := repo.ListByDate(ctx, userID, "2026-02-28")
		require.NoError(t, err)
		for _, r := range records {
			assert.NotEqual(t, record.ID, r.ID)
		}
	})

	t.Run("存在しないIDはErrNotFoundを返す", func(t *testing.T) {
		err := repo.Delete(ctx, userID, "non-existent-id")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
	})

	t.Run("userIDが空のときエラー", func(t *testing.T) {
		err := repo.Delete(ctx, "", "some-id")
		assert.Error(t, err)
	})
}
