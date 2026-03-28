package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
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
func (s *stubAnalysisRepository) SaveResult(_ context.Context, _ uuid.UUID, _ *repository.AnalysisResult) error {
	return nil
}
func (s *stubAnalysisRepository) GetResult(_ context.Context, _ string, _ uuid.UUID) (*repository.AnalysisResult, error) {
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
func (s *stubAnalysisRepository) CreateSkippedMeal(_ context.Context, _ string, _ string, _ *string) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (s *stubAnalysisRepository) UpdateResult(_ context.Context, _ string, _ uuid.UUID, _ []gemini.NutritionInfo) error {
	return nil
}
func (s *stubAnalysisRepository) GetPendingAnalysesForDate(_ context.Context, _ string, _ string, _ string) ([]repository.PendingAnalysisEntry, error) {
	return []repository.PendingAnalysisEntry{}, nil
}

// stubFoodService はテスト用のスタブ
type stubFoodService struct{}

func (s *stubFoodService) AnalyzeFoodImage(_ context.Context, _ string) (*repository.AnalysisResult, error) {
	return nil, nil
}
func (s *stubFoodService) AnalyzeFoodText(_ context.Context, _ string) (*repository.AnalysisResult, error) {
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

// stubExerciseRepository はテスト用のスタブ
type stubExerciseRepository struct{}

func (s *stubExerciseRepository) Create(_ context.Context, _ string, _ repository.CreateExerciseInput) (*repository.ExerciseRecord, error) {
	return nil, nil
}
func (s *stubExerciseRepository) ListByDate(_ context.Context, _ string, _ string) ([]repository.ExerciseRecord, error) {
	return nil, nil
}
func (s *stubExerciseRepository) Delete(_ context.Context, _ string, _ string) error {
	return nil
}

// stubExerciseService はテスト用のスタブ
type stubExerciseService struct{}

func (s *stubExerciseService) CreateExerciseRecord(_ context.Context, _ string, _ service.CreateExerciseInput, _ string) (*repository.ExerciseRecord, error) {
	return nil, nil
}
func (s *stubExerciseService) GetDailyExercise(_ context.Context, _ string, _ string) (*repository.ExerciseDailyResult, error) {
	return nil, nil
}
func (s *stubExerciseService) DeleteExerciseRecord(_ context.Context, _ string, _ string) error {
	return nil
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
		dailyMeals:    handler.NewDailyMealsHandler(analysisRepo, &stubExerciseRepository{}),
		skipMeal:      handler.NewSkipMealHandler(analysisRepo),
		weightRecord:  handler.NewWeightRecordHandler(&stubWeightRecordRepository{}, &stubWeightGoalRepository{}),
		weightGoal:    handler.NewWeightGoalHandler(&stubWeightGoalRepository{}),
		exercise:      handler.NewExerciseHandler(&stubExerciseService{}),
	}

	authMiddleware := middleware.NewAuthMiddleware(&testutil.MockTokenVerifier{
		VerifyFunc: func(token string) (string, error) {
			return "test-user", nil
		},
	})

	rateLimitConfig := middleware.LoadRateLimitConfig()
	rl := middleware.NewRateLimitMiddleware(rateLimitConfig)
	defer rl.Stop()

	mux := http.NewServeMux()
	setupRoutes(mux, h, authMiddleware, rl)

	// 認証ヘッダーなしでリクエスト → 401
	req := httptest.NewRequest(http.MethodGet, "/api/images/test.jpg", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "認証が必要です")
}

func TestEnableCORS_VaryHeader(t *testing.T) {
	origins := map[string]struct{}{
		"https://example.com": {},
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := enableCORS(next, origins)

	t.Run("Varyヘッダーが設定されるべき", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Origin", "https://example.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, "Origin", w.Header().Get("Vary"))
		assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("Origin無しでもVaryヘッダーが設定されるべき", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, "Origin", w.Header().Get("Vary"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("許可されていないオリジンではAccess-Control-Allow-Originが設定されないべき", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("Origin", "https://evil.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, "Origin", w.Header().Get("Vary"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestParseAllowedOrigins(t *testing.T) {
	t.Run("開発環境ではlocalhostオリジンが含まれるべき", func(t *testing.T) {
		originalEnv := os.Getenv("APP_ENV")
		defer os.Setenv("APP_ENV", originalEnv)

		os.Setenv("APP_ENV", "development")
		origins := parseAllowedOrigins("")

		assert.Contains(t, origins, "http://localhost:3000")
		assert.Contains(t, origins, "http://localhost:3001")
		assert.Contains(t, origins, "http://localhost:3002")
	})

	t.Run("本番環境ではlocalhostオリジンが含まれないべき", func(t *testing.T) {
		originalEnv := os.Getenv("APP_ENV")
		defer os.Setenv("APP_ENV", originalEnv)

		os.Setenv("APP_ENV", "production")
		origins := parseAllowedOrigins("")

		_, hasLocalhost := origins["http://localhost:3000"]
		assert.False(t, hasLocalhost)
	})

	t.Run("環境変数のオリジンが追加されるべき", func(t *testing.T) {
		originalEnv := os.Getenv("APP_ENV")
		defer os.Setenv("APP_ENV", originalEnv)

		os.Setenv("APP_ENV", "production")
		origins := parseAllowedOrigins("https://app.example.com,https://api.example.com")

		assert.Contains(t, origins, "https://app.example.com")
		assert.Contains(t, origins, "https://api.example.com")
	})
}
