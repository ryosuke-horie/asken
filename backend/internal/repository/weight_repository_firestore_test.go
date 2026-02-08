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

// cleanupWeightTestData はテスト用の体重データをクリーンアップします
func cleanupWeightTestData(ctx context.Context, client *firestore.Client, userID string) error {
	// weightRecords のクリーンアップ
	recordsIter := client.Collection("users").Doc(userID).Collection("weightRecords").Documents(ctx)
	defer recordsIter.Stop()

	bw := client.BulkWriter(ctx)
	for {
		doc, err := recordsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		bw.Delete(doc.Ref)
	}

	// weightGoal のクリーンアップ
	goalIter := client.Collection("users").Doc(userID).Collection("weightGoal").Documents(ctx)
	defer goalIter.Stop()

	for {
		doc, err := goalIter.Next()
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

func TestNewWeightRepositories(t *testing.T) {
	t.Run("nilクライアントでエラー", func(t *testing.T) {
		recordRepo, goalRepo, err := NewWeightRepositories(nil)
		assert.Error(t, err)
		assert.Nil(t, recordRepo)
		assert.Nil(t, goalRepo)
	})
}

func TestWeightRecordRepository_CreateRecord(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-weight-user-create"

	t.Cleanup(func() {
		cleanupWeightTestData(ctx, client, userID)
	})

	repo, _, err := NewWeightRepositories(client)
	require.NoError(t, err)

	t.Run("正常に体重記録を作成できる", func(t *testing.T) {
		recordedAt := time.Date(2026, 2, 8, 7, 30, 0, 0, time.UTC)
		record, err := repo.CreateRecord(ctx, userID, 65.3, recordedAt, "朝食前")

		require.NoError(t, err)
		assert.NotEmpty(t, record.ID)
		assert.Equal(t, 65.3, record.WeightKg)
		assert.Equal(t, "朝食前", record.Note)
		assert.False(t, record.CreatedAt.IsZero())
	})

	t.Run("小数点1桁に丸められる", func(t *testing.T) {
		recordedAt := time.Date(2026, 2, 8, 8, 0, 0, 0, time.UTC)
		record, err := repo.CreateRecord(ctx, userID, 65.35, recordedAt, "")

		require.NoError(t, err)
		assert.Equal(t, 65.4, record.WeightKg)
	})

	t.Run("空のuserIDでエラー", func(t *testing.T) {
		recordedAt := time.Date(2026, 2, 8, 7, 30, 0, 0, time.UTC)
		_, err := repo.CreateRecord(ctx, "", 65.3, recordedAt, "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "userIDが必要です")
	})
}

func TestWeightRecordRepository_GetRecord(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-weight-user-get"

	t.Cleanup(func() {
		cleanupWeightTestData(ctx, client, userID)
	})

	repo, _, err := NewWeightRepositories(client)
	require.NoError(t, err)

	t.Run("作成した記録を取得できる", func(t *testing.T) {
		recordedAt := time.Date(2026, 2, 8, 7, 30, 0, 0, time.UTC)
		created, err := repo.CreateRecord(ctx, userID, 70.0, recordedAt, "テスト")
		require.NoError(t, err)

		got, err := repo.GetRecord(ctx, userID, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, 70.0, got.WeightKg)
		assert.Equal(t, "テスト", got.Note)
	})

	t.Run("存在しないIDでErrNotFound", func(t *testing.T) {
		_, err := repo.GetRecord(ctx, userID, "nonexistent-id")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("空のuserIDでエラー", func(t *testing.T) {
		_, err := repo.GetRecord(ctx, "", "some-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "userIDが必要です")
	})
}

func TestWeightRecordRepository_UpdateRecord(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-weight-user-update"

	t.Cleanup(func() {
		cleanupWeightTestData(ctx, client, userID)
	})

	repo, _, err := NewWeightRepositories(client)
	require.NoError(t, err)

	t.Run("記録を更新できる", func(t *testing.T) {
		recordedAt := time.Date(2026, 2, 8, 7, 30, 0, 0, time.UTC)
		created, err := repo.CreateRecord(ctx, userID, 70.0, recordedAt, "更新前")
		require.NoError(t, err)

		updated, err := repo.UpdateRecord(ctx, userID, created.ID, 69.5, "更新後")
		require.NoError(t, err)
		assert.Equal(t, 69.5, updated.WeightKg)
		assert.Equal(t, "更新後", updated.Note)

		// 再取得して確認
		got, err := repo.GetRecord(ctx, userID, created.ID)
		require.NoError(t, err)
		assert.Equal(t, 69.5, got.WeightKg)
		assert.Equal(t, "更新後", got.Note)
	})

	t.Run("存在しないIDでErrNotFound", func(t *testing.T) {
		_, err := repo.UpdateRecord(ctx, userID, "nonexistent-id", 65.0, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestWeightRecordRepository_DeleteRecord(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-weight-user-delete"

	t.Cleanup(func() {
		cleanupWeightTestData(ctx, client, userID)
	})

	repo, _, err := NewWeightRepositories(client)
	require.NoError(t, err)

	t.Run("記録を削除できる", func(t *testing.T) {
		recordedAt := time.Date(2026, 2, 8, 7, 30, 0, 0, time.UTC)
		created, err := repo.CreateRecord(ctx, userID, 70.0, recordedAt, "削除対象")
		require.NoError(t, err)

		err = repo.DeleteRecord(ctx, userID, created.ID)
		require.NoError(t, err)

		// 削除後は取得できないことを確認
		_, err = repo.GetRecord(ctx, userID, created.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("存在しないIDでErrNotFound", func(t *testing.T) {
		err := repo.DeleteRecord(ctx, userID, "nonexistent-id")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestWeightRecordRepository_ListRecordsByDateRange(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-weight-user-list"

	t.Cleanup(func() {
		cleanupWeightTestData(ctx, client, userID)
	})

	repo, _, err := NewWeightRepositories(client)
	require.NoError(t, err)

	t.Run("期間内の記録を取得できる", func(t *testing.T) {
		// テストデータ作成
		day1 := time.Date(2026, 2, 1, 7, 0, 0, 0, time.UTC)
		day2 := time.Date(2026, 2, 5, 8, 0, 0, 0, time.UTC)
		day3 := time.Date(2026, 2, 10, 9, 0, 0, 0, time.UTC)

		_, err := repo.CreateRecord(ctx, userID, 70.0, day1, "1日")
		require.NoError(t, err)
		_, err = repo.CreateRecord(ctx, userID, 69.5, day2, "5日")
		require.NoError(t, err)
		_, err = repo.CreateRecord(ctx, userID, 69.0, day3, "10日")
		require.NoError(t, err)

		// 2/1 ~ 2/7 の範囲で取得
		from := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 2, 7, 23, 59, 59, 0, time.UTC)

		records, err := repo.ListRecordsByDateRange(ctx, userID, from, to)
		require.NoError(t, err)
		assert.Len(t, records, 2)
		assert.Equal(t, 70.0, records[0].WeightKg)
		assert.Equal(t, 69.5, records[1].WeightKg)
	})

	t.Run("データがない期間では空スライスを返す", func(t *testing.T) {
		from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)

		records, err := repo.ListRecordsByDateRange(ctx, userID, from, to)
		require.NoError(t, err)
		assert.Empty(t, records)
	})
}

func TestWeightGoalRepository(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	userID := "test-weight-user-goal"

	t.Cleanup(func() {
		cleanupWeightTestData(ctx, client, userID)
	})

	_, repo, err := NewWeightRepositories(client)
	require.NoError(t, err)

	t.Run("目標体重が未設定の場合nilを返す", func(t *testing.T) {
		goal, err := repo.GetGoal(ctx, userID)
		require.NoError(t, err)
		assert.Nil(t, goal)
	})

	t.Run("目標体重を設定・取得できる", func(t *testing.T) {
		goal, err := repo.SetGoal(ctx, userID, 63.0)
		require.NoError(t, err)
		assert.Equal(t, 63.0, goal.TargetWeightKg)

		got, err := repo.GetGoal(ctx, userID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, 63.0, got.TargetWeightKg)
	})

	t.Run("目標体重を更新できる", func(t *testing.T) {
		_, err := repo.SetGoal(ctx, userID, 62.5)
		require.NoError(t, err)

		got, err := repo.GetGoal(ctx, userID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, 62.5, got.TargetWeightKg)
	})

	t.Run("小数点1桁に丸められる", func(t *testing.T) {
		goal, err := repo.SetGoal(ctx, userID, 62.55)
		require.NoError(t, err)
		assert.Equal(t, 62.6, goal.TargetWeightKg)
	})
}
