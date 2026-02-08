package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/handler"
	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
	"github.com/ryosuke-horie/uchikomi/backend/internal/testutil"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
	"github.com/stretchr/testify/assert"
)

// stubAnalysisRepository はテスト用のスタブ（ルート登録のみ使用、実際には呼ばれない）
type stubAnalysisRepository struct{}

func (s *stubAnalysisRepository) CreateRequest(_ context.Context, _ string, _ string, _ string, _ *string) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (s *stubAnalysisRepository) CreateRequestWithText(_ context.Context, _ string, _ string, _ string, _ *string) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (s *stubAnalysisRepository) GetRequest(_ context.Context, _ string, _ uuid.UUID) (*repository.AnalysisRequest, error) {
	return nil, nil
}
func (s *stubAnalysisRepository) UpdateStatus(_ context.Context, _ uuid.UUID, _ repository.AnalysisStatus, _ string) error {
	return nil
}
func (s *stubAnalysisRepository) SaveResult(_ context.Context, _ uuid.UUID, _ *service.AnalysisResult) error {
	return nil
}
func (s *stubAnalysisRepository) GetResult(_ context.Context, _ string, _ uuid.UUID) (*service.AnalysisResult, error) {
	return nil, nil
}
func (s *stubAnalysisRepository) GetPendingRequests(_ context.Context, _ int) ([]repository.AnalysisRequest, error) {
	return nil, nil
}
func (s *stubAnalysisRepository) GetHistoryList(_ context.Context, _ string, _, _ int) ([]repository.HistoryItem, int, error) {
	return nil, 0, nil
}
func (s *stubAnalysisRepository) GetHistoryDetail(_ context.Context, _ string, _ uuid.UUID) (*repository.HistoryDetail, error) {
	return nil, nil
}
func (s *stubAnalysisRepository) DeleteHistory(_ context.Context, _ string, _ uuid.UUID) error {
	return nil
}
func (s *stubAnalysisRepository) GetDailyMeals(_ context.Context, _ string, _ string, _ string) (map[string][]repository.HistoryDetail, repository.DailyTotal, error) {
	return nil, repository.DailyTotal{}, nil
}
func (s *stubAnalysisRepository) CreateRequestFromMylist(_ context.Context, _ string, _ string, _ string, _ *string, _ *service.AnalysisResult) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (s *stubAnalysisRepository) CreateSkippedMeal(_ context.Context, _ string, _ string, _ *string) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (s *stubAnalysisRepository) UpdateResult(_ context.Context, _ string, _ uuid.UUID, _ []gemini.NutritionInfo) error {
	return nil
}

// stubFoodService はテスト用のスタブ
type stubFoodService struct{}

func (s *stubFoodService) AnalyzeFoodImage(_ context.Context, _ string) (*service.AnalysisResult, error) {
	return nil, nil
}
func (s *stubFoodService) AnalyzeFoodText(_ context.Context, _ string) (*service.AnalysisResult, error) {
	return nil, nil
}

// stubWeightRecordRepository はテスト用のスタブ
type stubWeightRecordRepository struct{}

func (s *stubWeightRecordRepository) CreateRecord(_ context.Context, _ string, _ float64, _ time.Time, _ string) (*repository.WeightRecord, error) {
	return nil, nil
}
func (s *stubWeightRecordRepository) GetRecord(_ context.Context, _ string, _ string) (*repository.WeightRecord, error) {
	return nil, nil
}
func (s *stubWeightRecordRepository) UpdateRecord(_ context.Context, _ string, _ string, _ float64, _ string) (*repository.WeightRecord, error) {
	return nil, nil
}
func (s *stubWeightRecordRepository) DeleteRecord(_ context.Context, _ string, _ string) error {
	return nil
}
func (s *stubWeightRecordRepository) ListRecords(_ context.Context, _ string, _ int, _ string) ([]repository.WeightRecord, string, error) {
	return nil, "", nil
}
func (s *stubWeightRecordRepository) ListRecordsByDateRange(_ context.Context, _ string, _, _ time.Time) ([]repository.WeightRecord, error) {
	return nil, nil
}

// stubWeightGoalRepository はテスト用のスタブ
type stubWeightGoalRepository struct{}

func (s *stubWeightGoalRepository) GetGoal(_ context.Context, _ string) (*repository.WeightGoal, error) {
	return nil, nil
}
func (s *stubWeightGoalRepository) SetGoal(_ context.Context, _ string, _ float64) (*repository.WeightGoal, error) {
	return nil, nil
}

func TestSetupRoutes_ImageEndpointRequiresAuth(t *testing.T) {
	// 認証なしで画像エンドポイントにアクセスすると401が返ることを検証
	analysisRepo := &stubAnalysisRepository{}
	storageRepo := &testutil.MockStorageRepository{}

	h := handlers{
		health:        handler.NewHealthHandler(),
		analyze:       handler.NewAnalyzeHandler(&stubFoodService{}, analysisRepo, storageRepo),
		status:        handler.NewStatusHandler(analysisRepo),
		history:       handler.NewHistoryHandler(analysisRepo, nil),
		historyDelete: handler.NewHistoryDeleteHandler(analysisRepo),
		image:         handler.NewImageHandler(storageRepo),
		dailyMeals:    handler.NewDailyMealsHandler(analysisRepo),
		skipMeal:      handler.NewSkipMealHandler(analysisRepo),
		weightRecord:  handler.NewWeightRecordHandler(&stubWeightRecordRepository{}, &stubWeightGoalRepository{}),
		weightGoal:    handler.NewWeightGoalHandler(&stubWeightGoalRepository{}),
	}

	authMiddleware := middleware.NewAuthMiddleware(&testutil.MockTokenVerifier{
		VerifyFunc: func(token string) (string, error) {
			return "test-user", nil
		},
	})

	mux := http.NewServeMux()
	setupRoutes(mux, h, authMiddleware)

	// 認証ヘッダーなしでリクエスト → 401
	req := httptest.NewRequest(http.MethodGet, "/api/images/test.jpg", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "認証が必要です")
}
