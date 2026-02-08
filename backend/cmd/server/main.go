package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/internal/handler"
	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
	"github.com/ryosuke-horie/uchikomi/backend/internal/worker"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/database"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/storage"
)

// RealGeminiClient はGeminiClientインターフェースを実装する
type RealGeminiClient struct {
	classifier *gemini.Classifier
	textParser *gemini.TextParser
	calculator *gemini.NutritionCalculator
}

func (r *RealGeminiClient) ClassifyFoods(ctx context.Context, imagePath string) ([]gemini.FoodItem, error) {
	return r.classifier.ClassifyFoods(ctx, imagePath)
}

func (r *RealGeminiClient) ClassifyFoodsFromData(ctx context.Context, imageData []byte, mimeType string) ([]gemini.FoodItem, error) {
	return r.classifier.ClassifyFoodsFromData(ctx, imageData, mimeType)
}

func (r *RealGeminiClient) ParseTextToFoods(ctx context.Context, inputText string) ([]gemini.FoodItem, error) {
	return r.textParser.ParseTextToFoods(ctx, inputText)
}

func (r *RealGeminiClient) CalculateNutrition(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
	return r.calculator.CalculateNutrition(ctx, foods)
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

type handlers struct {
	health        *handler.HealthHandler
	analyze       *handler.AnalyzeHandler
	status        *handler.StatusHandler
	history       *handler.HistoryHandler
	historyDelete *handler.HistoryDeleteHandler
	image         *handler.ImageHandler
	dailyMeals    *handler.DailyMealsHandler
	skipMeal      *handler.SkipMealHandler
	weightRecord  *handler.WeightRecordHandler
	weightGoal    *handler.WeightGoalHandler
}

func setupRoutes(mux *http.ServeMux, h handlers, authMiddleware middleware.Authenticator) {
	// ヘルスチェックエンドポイント（認証不要 - Cloud Runのヘルスチェック用）
	mux.HandleFunc("/api/health", h.health.Handle)

	// 画像配信エンドポイント（認証必須）
	mux.Handle("/api/images/", authMiddleware.Authenticate(http.HandlerFunc(h.image.Handle)))

	// 認証が必要なエンドポイント
	setupAnalyzeRoutes(mux, h, authMiddleware)
	setupHistoryRoutes(mux, h, authMiddleware)
	setupMealsRoutes(mux, h, authMiddleware)
	setupWeightRoutes(mux, h, authMiddleware)
}

func setupAnalyzeRoutes(mux *http.ServeMux, h handlers, authMiddleware middleware.Authenticator) {
	analyzeRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.analyze.Handle(w, r)
		case http.MethodGet:
			h.status.Handle(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/analyze", authMiddleware.Authenticate(http.HandlerFunc(analyzeRouteHandler)))
	mux.Handle("/api/analyze/", authMiddleware.Authenticate(http.HandlerFunc(analyzeRouteHandler)))

	uploadImageRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.analyze.HandleUploadImage(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/upload-image", authMiddleware.Authenticate(http.HandlerFunc(uploadImageRouteHandler)))
}

func setupHistoryRoutes(mux *http.ServeMux, h handlers, authMiddleware middleware.Authenticator) {
	mux.Handle("/api/history", authMiddleware.Authenticate(http.HandlerFunc(h.history.HandleList)))

	historyDetailRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.history.HandleDetail(w, r)
		case http.MethodPut:
			h.history.HandleUpdate(w, r)
		case http.MethodDelete:
			h.historyDelete.Handle(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/history/", authMiddleware.Authenticate(http.HandlerFunc(historyDetailRouteHandler)))
}

func setupMealsRoutes(mux *http.ServeMux, h handlers, authMiddleware middleware.Authenticator) {
	mux.Handle("/api/meals/daily", authMiddleware.Authenticate(http.HandlerFunc(h.dailyMeals.Handle)))

	mealsSkipRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.skipMeal.Handle(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/meals/skip", authMiddleware.Authenticate(http.HandlerFunc(mealsSkipRouteHandler)))
}

func setupWeightRoutes(mux *http.ServeMux, h handlers, authMiddleware middleware.Authenticator) {
	weightRecordsRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.weightRecord.HandleList(w, r)
		case http.MethodPost:
			h.weightRecord.HandleCreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/weight/records", authMiddleware.Authenticate(http.HandlerFunc(weightRecordsRouteHandler)))

	weightRecordDetailRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.weightRecord.HandleGet(w, r)
		case http.MethodPut:
			h.weightRecord.HandleUpdate(w, r)
		case http.MethodDelete:
			h.weightRecord.HandleDelete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/weight/records/", authMiddleware.Authenticate(http.HandlerFunc(weightRecordDetailRouteHandler)))

	weightGoalRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.weightGoal.HandleGet(w, r)
		case http.MethodPut:
			h.weightGoal.HandleSet(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/weight/goal", authMiddleware.Authenticate(http.HandlerFunc(weightGoalRouteHandler)))
}

func run() error {
	ctx := context.Background()

	// Firestoreクライアントの初期化
	firebaseCredentials := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	firestoreClient, err := database.NewFirestoreClient(ctx, firebaseCredentials)
	if err != nil {
		log.Fatalf("Failed to connect to Firestore: %v", err)
	}
	defer firestoreClient.Close()
	log.Println("Firestore connection established")

	// Cloud Storageクライアントの初期化
	storageClient, err := storage.NewStorageClient(ctx, firebaseCredentials)
	if err != nil {
		log.Fatalf("Failed to connect to Cloud Storage: %v", err)
	}
	defer storageClient.Close()
	log.Println("Cloud Storage connection established")

	// StorageRepositoryの初期化
	gcsBucketName := os.Getenv("GCS_BUCKET_NAME")
	if gcsBucketName == "" {
		log.Fatalf("GCS_BUCKET_NAME environment variable is required")
	}
	storageRepo, err := repository.NewStorageRepositoryCloudStorage(storageClient, gcsBucketName)
	if err != nil {
		log.Fatalf("Failed to initialize StorageRepository: %v", err)
	}

	// リポジトリの初期化
	analysisRepo, err := repository.NewAnalysisRepositoryFirestore(firestoreClient, storageRepo)
	if err != nil {
		log.Fatalf("Failed to initialize AnalysisRepository: %v", err)
	}

	weightRecordRepo, weightGoalRepo, err := repository.NewWeightRepositories(firestoreClient)
	if err != nil {
		log.Fatalf("Failed to initialize WeightRepositories: %v", err)
	}

	// Gemini API Keyの確認
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		log.Println("WARNING: GEMINI_API_KEY not set. Gemini API calls will fail.")
	} else {
		log.Println("Gemini API Key configured")
	}

	// 依存関係の初期化
	classifier := gemini.NewClassifier(120 * time.Second)
	textParser := gemini.NewTextParser(120 * time.Second)
	calculator := gemini.NewNutritionCalculator(120 * time.Second)
	geminiClient := &RealGeminiClient{
		classifier: classifier,
		textParser: textParser,
		calculator: calculator,
	}

	foodService := service.NewFoodService(geminiClient, storageRepo)

	// 認証ミドルウェアの初期化
	var authMiddleware middleware.Authenticator
	if middleware.IsDevMode() {
		log.Println("WARNING: Running in development mode with mock authentication")
		authMiddleware = middleware.NewDevAuthMiddleware()
	} else {
		firebaseAuthService, err := service.NewFirebaseAuthService(ctx, firebaseCredentials)
		if err != nil {
			log.Fatalf("Failed to initialize Firebase Auth Service: %v", err)
		}
		log.Println("Firebase Auth Service initialized")
		authMiddleware = middleware.NewAuthMiddleware(firebaseAuthService)
	}

	// ハンドラーの初期化
	h := handlers{
		health:        handler.NewHealthHandler(),
		analyze:       handler.NewAnalyzeHandler(foodService, analysisRepo, storageRepo),
		status:        handler.NewStatusHandler(analysisRepo),
		history:       handler.NewHistoryHandler(analysisRepo, geminiClient),
		historyDelete: handler.NewHistoryDeleteHandler(analysisRepo),
		image:         handler.NewImageHandler(storageRepo),
		dailyMeals:    handler.NewDailyMealsHandler(analysisRepo),
		skipMeal:      handler.NewSkipMealHandler(analysisRepo),
		weightRecord:  handler.NewWeightRecordHandler(weightRecordRepo, weightGoalRepo),
		weightGoal:    handler.NewWeightGoalHandler(weightGoalRepo),
	}

	// ワーカーの初期化
	analysisWorker := worker.NewAnalysisWorker(foodService, analysisRepo, 5*time.Second)

	// ワーカー用のコンテキスト
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	// ワーカーを別ゴルーチンで起動
	go analysisWorker.Start(workerCtx)

	// ルーティング
	mux := http.NewServeMux()
	setupRoutes(mux, h, authMiddleware)

	// ミドルウェアチェーンを構築（リクエスト処理順: セキュリティヘッダー → CORS → mux）
	allowedOrigins := parseAllowedOrigins(os.Getenv("ALLOWED_ORIGINS"))
	corsHandler := enableCORS(mux, allowedOrigins)
	secureHandler := middleware.SecurityHeaders(corsHandler)

	// HTTPサーバー設定
	server := &http.Server{
		Addr:         ":8080",
		Handler:      secureHandler,
		ReadTimeout:  150 * time.Second,
		WriteTimeout: 150 * time.Second,
		IdleTimeout:  150 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")

		// ワーカーを停止
		workerCancel()
		log.Println("Worker stopped")

		// HTTPサーバーを停止
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Println("Server starting on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// enableCORS はCORSヘッダーを追加するミドルウェア
func enableCORS(next http.Handler, allowedOrigins map[string]struct{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			if _, ok := allowedOrigins[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// プリフライトリクエスト（OPTIONS）への対応
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func parseAllowedOrigins(raw string) map[string]struct{} {
	origins := map[string]struct{}{
		"http://localhost:3000":        {},
		"http://localhost:3001":        {},
		"http://localhost:3002":        {},
		"https://utikomi.exe.xyz:3000": {},
	}

	if raw == "" {
		return origins
	}

	for _, origin := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(origin); trimmed != "" {
			origins[trimmed] = struct{}{}
		}
	}

	return origins
}
