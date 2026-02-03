package repository

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
	"github.com/ryosuke-horie/uchikomi/backend/internal/testutil"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStorageRepositoryForAnalysis はテスト用のモックStorageRepository
// DeleteFuncを設定可能にして、削除失敗テストに対応
type mockStorageRepositoryForAnalysis struct {
	DeleteFunc func(ctx context.Context, objectName string) error
}

func (m *mockStorageRepositoryForAnalysis) Upload(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
	return "uploads/test-uuid.jpg", nil
}

func (m *mockStorageRepositoryForAnalysis) Download(ctx context.Context, objectName string) ([]byte, error) {
	return []byte("test image data"), nil
}

func (m *mockStorageRepositoryForAnalysis) GetSignedURL(ctx context.Context, objectName string, expiration time.Duration) (string, error) {
	return "https://storage.googleapis.com/bucket/" + objectName + "?signature=xxx", nil
}

func (m *mockStorageRepositoryForAnalysis) Delete(ctx context.Context, objectName string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, objectName)
	}
	return nil
}

// testutilMock はtestutilのモックインターフェースを確認するための型チェック
var _ = testutil.MockStorageRepository{}

// getTestFirestoreClient はテスト用のFirestoreクライアントを取得します。
// Firestoreエミュレータが起動していない場合はテストをスキップします。
func getTestFirestoreClient(t *testing.T) *firestore.Client {
	t.Helper()

	emulatorHost := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if emulatorHost == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST が設定されていないためスキップします。firebase emulators:start --only firestore を実行してください。")
	}

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "test-project")
	require.NoError(t, err, "Firestoreクライアントの作成に失敗")

	t.Cleanup(func() {
		client.Close()
	})

	return client
}

// cleanupTestData はテストデータをクリーンアップします
func cleanupTestData(ctx context.Context, client *firestore.Client, userID string) error {
	iter := client.Collection("users").Doc(userID).Collection("analysisRequests").Documents(ctx)
	defer iter.Stop()

	bw := client.BulkWriter(ctx)
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		bw.Delete(doc.Ref)
	}
	bw.Flush()
	bw.End()

	return nil
}

func TestCreateRequest(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	repo, err := NewAnalysisRepositoryFirestore(client, &mockStorageRepositoryForAnalysis{})
	require.NoError(t, err)
	userID := "test-user-" + uuid.New().String()

	t.Cleanup(func() {
		cleanupTestData(ctx, client, userID)
	})

	t.Run("正常系: 画像分析リクエストを作成できる", func(t *testing.T) {
		imagePath := "/uploads/test-image.jpg"
		mealType := "breakfast"
		mealDate := "2024-01-15"

		id, err := repo.CreateRequest(ctx, imagePath, mealType, mealDate, &userID)

		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)

		// 作成されたリクエストを確認（userIDでスコープ）
		request, err := repo.GetRequest(ctx, userID, id)
		require.NoError(t, err)
		assert.Equal(t, StatusPending, request.Status)
		assert.Equal(t, InputTypeImage, request.InputType)
		assert.Equal(t, imagePath, request.ImagePath)
	})

	t.Run("異常系: userIDがnilの場合エラー", func(t *testing.T) {
		_, err := repo.CreateRequest(ctx, "/test.jpg", "lunch", "2024-01-15", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "userIDが必要です")
	})

	t.Run("異常系: 不正な日付形式の場合エラー", func(t *testing.T) {
		_, err := repo.CreateRequest(ctx, "/test.jpg", "lunch", "invalid-date", &userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "日付のパースに失敗")
	})
}

func TestCreateRequestWithText(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	repo, err := NewAnalysisRepositoryFirestore(client, &mockStorageRepositoryForAnalysis{})
	require.NoError(t, err)
	userID := "test-user-" + uuid.New().String()

	t.Cleanup(func() {
		cleanupTestData(ctx, client, userID)
	})

	t.Run("正常系: テキスト分析リクエストを作成できる", func(t *testing.T) {
		inputText := "鶏むね肉 200g, 白米 150g"
		mealType := "lunch"
		mealDate := "2024-01-15"

		id, err := repo.CreateRequestWithText(ctx, inputText, mealType, mealDate, &userID)

		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)

		// 作成されたリクエストを確認（userIDでスコープ）
		request, err := repo.GetRequest(ctx, userID, id)
		require.NoError(t, err)
		assert.Equal(t, StatusPending, request.Status)
		assert.Equal(t, InputTypeText, request.InputType)
		assert.Equal(t, inputText, request.InputText)
	})
}

func TestUpdateStatus(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	repo, err := NewAnalysisRepositoryFirestore(client, &mockStorageRepositoryForAnalysis{})
	require.NoError(t, err)
	userID := "test-user-" + uuid.New().String()

	t.Cleanup(func() {
		cleanupTestData(ctx, client, userID)
	})

	t.Run("正常系: ステータスを更新できる", func(t *testing.T) {
		// リクエストを作成
		id, err := repo.CreateRequest(ctx, "/test.jpg", "breakfast", "2024-01-15", &userID)
		require.NoError(t, err)

		// ステータスを更新
		err = repo.UpdateStatus(ctx, id, StatusProcessing, "")
		require.NoError(t, err)

		// 更新されたことを確認（userIDでスコープ）
		request, err := repo.GetRequest(ctx, userID, id)
		require.NoError(t, err)
		assert.Equal(t, StatusProcessing, request.Status)
	})

	t.Run("正常系: エラーメッセージ付きでステータスを更新できる", func(t *testing.T) {
		id, err := repo.CreateRequest(ctx, "/test.jpg", "lunch", "2024-01-15", &userID)
		require.NoError(t, err)

		errorMsg := "分析に失敗しました"
		err = repo.UpdateStatus(ctx, id, StatusFailed, errorMsg)
		require.NoError(t, err)

		// 更新されたことを確認（userIDでスコープ）
		request, err := repo.GetRequest(ctx, userID, id)
		require.NoError(t, err)
		assert.Equal(t, StatusFailed, request.Status)
		assert.Equal(t, errorMsg, request.ErrorMessage)
	})
}

func TestSaveResultAndGetResult(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	repo, err := NewAnalysisRepositoryFirestore(client, &mockStorageRepositoryForAnalysis{})
	require.NoError(t, err)
	userID := "test-user-" + uuid.New().String()

	t.Cleanup(func() {
		cleanupTestData(ctx, client, userID)
	})

	t.Run("正常系: 分析結果を保存して取得できる", func(t *testing.T) {
		// リクエストを作成
		id, err := repo.CreateRequest(ctx, "/test.jpg", "breakfast", "2024-01-15", &userID)
		require.NoError(t, err)

		// 結果を保存
		result := &service.AnalysisResult{
			Foods: []gemini.NutritionInfo{
				{Name: "鶏むね肉", Calories: 200, Protein: 40, Fat: 5, Carbohydrates: 0},
				{Name: "白米", Calories: 250, Protein: 5, Fat: 1, Carbohydrates: 55},
			},
			TotalCalories:      450,
			TotalProtein:       45,
			TotalFat:           6,
			TotalCarbohydrates: 55,
		}

		err = repo.SaveResult(ctx, id, result)
		require.NoError(t, err)

		// 結果を取得（userIDでスコープ）
		savedResult, err := repo.GetResult(ctx, userID, id)
		require.NoError(t, err)
		assert.Equal(t, result.TotalCalories, savedResult.TotalCalories)
		assert.Equal(t, result.TotalProtein, savedResult.TotalProtein)
		assert.Len(t, savedResult.Foods, 2)
	})
}

func TestGetPendingRequests(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	repo, err := NewAnalysisRepositoryFirestore(client, &mockStorageRepositoryForAnalysis{})
	require.NoError(t, err)
	userID := "test-user-" + uuid.New().String()

	t.Cleanup(func() {
		cleanupTestData(ctx, client, userID)
	})

	t.Run("正常系: pendingステータスのリクエストを取得できる", func(t *testing.T) {
		// 複数のリクエストを作成
		_, err := repo.CreateRequest(ctx, "/test1.jpg", "breakfast", "2024-01-15", &userID)
		require.NoError(t, err)
		_, err = repo.CreateRequest(ctx, "/test2.jpg", "lunch", "2024-01-15", &userID)
		require.NoError(t, err)

		// 少し待機（Firestoreの書き込み反映）
		time.Sleep(100 * time.Millisecond)

		// pendingリクエストを取得
		requests, err := repo.GetPendingRequests(ctx, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(requests), 2)

		for _, req := range requests {
			assert.Equal(t, StatusPending, req.Status)
		}
	})
}

func TestGetHistoryListAndDetail(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	repo, err := NewAnalysisRepositoryFirestore(client, &mockStorageRepositoryForAnalysis{})
	require.NoError(t, err)
	userID := "test-user-" + uuid.New().String()

	t.Cleanup(func() {
		cleanupTestData(ctx, client, userID)
	})

	t.Run("正常系: 履歴一覧と詳細を取得できる", func(t *testing.T) {
		// リクエストを作成して結果を保存
		id, err := repo.CreateRequest(ctx, "/test.jpg", "breakfast", "2024-01-15", &userID)
		require.NoError(t, err)

		result := &service.AnalysisResult{
			Foods: []gemini.NutritionInfo{
				{Name: "テスト食品", Calories: 100, Protein: 10, Fat: 5, Carbohydrates: 10},
			},
			TotalCalories:      100,
			TotalProtein:       10,
			TotalFat:           5,
			TotalCarbohydrates: 10,
		}
		err = repo.SaveResult(ctx, id, result)
		require.NoError(t, err)

		// 少し待機
		time.Sleep(100 * time.Millisecond)

		// 履歴一覧を取得（userIDでスコープ）
		items, total, err := repo.GetHistoryList(ctx, userID, 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 1)
		assert.GreaterOrEqual(t, len(items), 1)

		// 履歴詳細を取得（userIDでスコープ）
		detail, err := repo.GetHistoryDetail(ctx, userID, id)
		require.NoError(t, err)
		assert.Equal(t, id, detail.ID)
		assert.Len(t, detail.Foods, 1)
	})
}

func TestDeleteHistory(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	repo, err := NewAnalysisRepositoryFirestore(client, &mockStorageRepositoryForAnalysis{})
	require.NoError(t, err)
	userID := "test-user-" + uuid.New().String()

	t.Cleanup(func() {
		cleanupTestData(ctx, client, userID)
	})

	t.Run("正常系: 履歴を削除できる", func(t *testing.T) {
		// リクエストを作成して結果を保存
		id, err := repo.CreateRequest(ctx, "/test.jpg", "breakfast", "2024-01-15", &userID)
		require.NoError(t, err)

		result := &service.AnalysisResult{
			Foods:         []gemini.NutritionInfo{},
			TotalCalories: 0,
		}
		err = repo.SaveResult(ctx, id, result)
		require.NoError(t, err)

		// 削除（userIDでスコープ）
		err = repo.DeleteHistory(ctx, userID, id)
		require.NoError(t, err)

		// 削除されたことを確認（GetRequestでエラー、userIDでスコープ）
		_, err = repo.GetRequest(ctx, userID, id)
		assert.Error(t, err)
	})
}

func TestDeleteHistory_StorageDeleteFailure(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()

	storageErr := errors.New("storage delete failed")
	mockStorage := &mockStorageRepositoryForAnalysis{
		DeleteFunc: func(ctx context.Context, objectName string) error {
			return storageErr
		},
	}

	repo, err := NewAnalysisRepositoryFirestore(client, mockStorage)
	require.NoError(t, err)
	userID := "test-user-" + uuid.New().String()

	t.Cleanup(func() {
		cleanupTestData(ctx, client, userID)
	})

	t.Run("異常系: Storage削除失敗でエラーを返す", func(t *testing.T) {
		// 画像付きリクエストを作成して結果を保存
		id, err := repo.CreateRequest(ctx, "uploads/test-image.jpg", "breakfast", "2024-01-15", &userID)
		require.NoError(t, err)

		result := &service.AnalysisResult{
			Foods:         []gemini.NutritionInfo{},
			TotalCalories: 0,
		}
		err = repo.SaveResult(ctx, id, result)
		require.NoError(t, err)

		// 削除実行（Storage削除失敗でエラーになるべき）
		err = repo.DeleteHistory(ctx, userID, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "画像の削除に失敗")

		// Firestoreドキュメントはまだ存在する（Storage削除を先に実行するため）
		req, err := repo.GetRequest(ctx, userID, id)
		assert.NoError(t, err)
		assert.NotNil(t, req)
	})
}

func TestGetDailyMeals(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	repo, err := NewAnalysisRepositoryFirestore(client, &mockStorageRepositoryForAnalysis{})
	require.NoError(t, err)
	userID := "test-user-" + uuid.New().String()

	t.Cleanup(func() {
		cleanupTestData(ctx, client, userID)
	})

	t.Run("正常系: 日次食事データを取得できる", func(t *testing.T) {
		// 朝食を作成
		id, err := repo.CreateRequest(ctx, "/breakfast.jpg", "breakfast", "2024-01-15", &userID)
		require.NoError(t, err)

		result := &service.AnalysisResult{
			Foods: []gemini.NutritionInfo{
				{Name: "朝食", Calories: 300, Protein: 15, Fat: 10, Carbohydrates: 40},
			},
			TotalCalories:      300,
			TotalProtein:       15,
			TotalFat:           10,
			TotalCarbohydrates: 40,
		}
		err = repo.SaveResult(ctx, id, result)
		require.NoError(t, err)

		// 少し待機
		time.Sleep(100 * time.Millisecond)

		// 日次データを取得（userIDでスコープ）
		meals, total, err := repo.GetDailyMeals(ctx, userID, "2024-01-15")
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(meals["breakfast"]), 1)
		assert.Equal(t, 300.0, total.TotalCalories)
	})
}

func TestCreateSkippedMeal(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	repo, err := NewAnalysisRepositoryFirestore(client, &mockStorageRepositoryForAnalysis{})
	require.NoError(t, err)
	userID := "test-user-" + uuid.New().String()

	t.Cleanup(func() {
		cleanupTestData(ctx, client, userID)
	})

	t.Run("正常系: スキップ記録を作成できる", func(t *testing.T) {
		id, err := repo.CreateSkippedMeal(ctx, "breakfast", "2024-01-15", &userID)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)

		// 少し待機
		time.Sleep(100 * time.Millisecond)

		// 日次データで確認（userIDでスコープ）
		meals, total, err := repo.GetDailyMeals(ctx, userID, "2024-01-15")
		require.NoError(t, err)
		assert.Equal(t, 0.0, total.TotalCalories)
		assert.GreaterOrEqual(t, len(meals["breakfast"]), 1)
	})

	t.Run("正常系: スキップ記録は既存記録を置き換える", func(t *testing.T) {
		// まず通常の記録を作成
		id1, err := repo.CreateRequest(ctx, "/test.jpg", "lunch", "2024-01-16", &userID)
		require.NoError(t, err)

		result := &service.AnalysisResult{
			Foods:         []gemini.NutritionInfo{{Name: "テスト", Calories: 500}},
			TotalCalories: 500,
		}
		err = repo.SaveResult(ctx, id1, result)
		require.NoError(t, err)

		// スキップ記録を作成（既存記録を置き換え）
		id2, err := repo.CreateSkippedMeal(ctx, "lunch", "2024-01-16", &userID)
		require.NoError(t, err)
		assert.NotEqual(t, id1, id2)

		// 少し待機
		time.Sleep(100 * time.Millisecond)

		// 元の記録は削除されている（userIDでスコープ）
		_, err = repo.GetRequest(ctx, userID, id1)
		assert.Error(t, err)
	})
}

func TestUpdateResult(t *testing.T) {
	client := getTestFirestoreClient(t)
	ctx := context.Background()
	repo, err := NewAnalysisRepositoryFirestore(client, &mockStorageRepositoryForAnalysis{})
	require.NoError(t, err)
	userID := "test-user-" + uuid.New().String()

	t.Cleanup(func() {
		cleanupTestData(ctx, client, userID)
	})

	t.Run("正常系: 分析結果を更新できる", func(t *testing.T) {
		// リクエストを作成して結果を保存
		id, err := repo.CreateRequest(ctx, "/test.jpg", "breakfast", "2024-01-15", &userID)
		require.NoError(t, err)

		initialResult := &service.AnalysisResult{
			Foods: []gemini.NutritionInfo{
				{Name: "食品1", Calories: 100, Protein: 10, Fat: 5, Carbohydrates: 10},
			},
			TotalCalories:      100,
			TotalProtein:       10,
			TotalFat:           5,
			TotalCarbohydrates: 10,
		}
		err = repo.SaveResult(ctx, id, initialResult)
		require.NoError(t, err)

		// 結果を更新（userIDでスコープ）
		newFoods := []gemini.NutritionInfo{
			{Name: "食品1", Calories: 150, Protein: 15, Fat: 7, Carbohydrates: 15},
			{Name: "食品2", Calories: 200, Protein: 20, Fat: 10, Carbohydrates: 20},
		}
		err = repo.UpdateResult(ctx, userID, id, newFoods)
		require.NoError(t, err)

		// 更新された結果を確認（userIDでスコープ）
		updatedResult, err := repo.GetResult(ctx, userID, id)
		require.NoError(t, err)
		assert.Len(t, updatedResult.Foods, 2)
		assert.Equal(t, 350.0, updatedResult.TotalCalories) // 150 + 200
	})
}
