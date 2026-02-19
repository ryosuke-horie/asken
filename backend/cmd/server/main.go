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

	"cloud.google.com/go/firestore"

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
	nutritionGoal *handler.NutritionGoalHandler
	myMenu        *handler.MyMenuHandler
}

func setupRoutes(mux *http.ServeMux, h handlers, authMiddleware middleware.Authenticator, rl *middleware.RateLimitMiddleware) {
	// ヘルスチェックエンドポイント（認証不要 - Cloud Runのヘルスチェック用）
	mux.HandleFunc("/api/health", h.health.Handle)

	// 画像配信エンドポイント（認証必須）
	mux.Handle("/api/images/", authMiddleware.Authenticate(rl.LimitByUser(http.HandlerFunc(h.image.Handle))))

	// 認証が必要なエンドポイント
	setupAnalyzeRoutes(mux, h, authMiddleware, rl)
	setupHistoryRoutes(mux, h, authMiddleware, rl)
	setupMealsRoutes(mux, h, authMiddleware, rl)
	setupWeightRoutes(mux, h, authMiddleware, rl)
	setupNutritionRoutes(mux, h, authMiddleware, rl)
	setupMyMenuRoutes(mux, h, authMiddleware, rl)
}

func setupAnalyzeRoutes(mux *http.ServeMux, h handlers, authMiddleware middleware.Authenticator, rl *middleware.RateLimitMiddleware) {
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
	mux.Handle("/api/analyze", authMiddleware.Authenticate(rl.LimitByUser(http.HandlerFunc(analyzeRouteHandler))))
	mux.Handle("/api/analyze/", authMiddleware.Authenticate(rl.LimitByUser(http.HandlerFunc(analyzeRouteHandler))))

	uploadImageRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.analyze.HandleUploadImage(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/upload-image", authMiddleware.Authenticate(rl.LimitByUser(http.HandlerFunc(uploadImageRouteHandler))))
}

func setupHistoryRoutes(mux *http.ServeMux, h handlers, authMiddleware middleware.Authenticator, rl *middleware.RateLimitMiddleware) {
	mux.Handle("/api/history", authMiddleware.Authenticate(rl.LimitByUser(http.HandlerFunc(h.history.HandleList))))

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
	mux.Handle("/api/history/", authMiddleware.Authenticate(rl.LimitByUser(http.HandlerFunc(historyDetailRouteHandler))))
}

func setupMealsRoutes(mux *http.ServeMux, h handlers, authMiddleware middleware.Authenticator, rl *middleware.RateLimitMiddleware) {
	mux.Handle("/api/meals/daily", authMiddleware.Authenticate(rl.LimitByUser(http.HandlerFunc(h.dailyMeals.Handle))))

	mealsSkipRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.skipMeal.Handle(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/meals/skip", authMiddleware.Authenticate(rl.LimitByUser(http.HandlerFunc(mealsSkipRouteHandler))))
}

func setupWeightRoutes(mux *http.ServeMux, h handlers, authMiddleware middleware.Authenticator, rl *middleware.RateLimitMiddleware) {
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
	mux.Handle("/api/weight/records", authMiddleware.Authenticate(rl.LimitByUser(http.HandlerFunc(weightRecordsRouteHandler))))

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
	mux.Handle("/api/weight/records/", authMiddleware.Authenticate(rl.LimitByUser(http.HandlerFunc(weightRecordDetailRouteHandler))))

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
	mux.Handle("/api/weight/goal", authMiddleware.Authenticate(rl.LimitByUser(http.HandlerFunc(weightGoalRouteHandler))))
}

func setupNutritionRoutes(mux *http.ServeMux, h handlers, authMiddleware middleware.Authenticator, rl *middleware.RateLimitMiddleware) {
	nutritionGoalRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.nutritionGoal.HandleGet(w, r)
		case http.MethodPut:
			h.nutritionGoal.HandleSet(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/nutrition/goal", authMiddleware.Authenticate(rl.LimitByUser(http.HandlerFunc(nutritionGoalRouteHandler))))
}

func setupMyMenuRoutes(mux *http.ServeMux, h handlers, authMiddleware middleware.Authenticator, rl *middleware.RateLimitMiddleware) {
	myMenuListRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.myMenu.HandleList(w, r)
		case http.MethodPost:
			h.myMenu.HandleCreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/my-menu", authMiddleware.Authenticate(rl.LimitByUser(http.HandlerFunc(myMenuListRouteHandler))))

	myMenuDetailRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		normalizedPath := strings.TrimSuffix(r.URL.Path, "/")

		// /api/my-menu/:id/record エンドポイントの処理
		if strings.HasSuffix(normalizedPath, "/record") {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.myMenu.HandleRecord(w, r)
			return
		}

		// 通常の詳細エンドポイント
		switch r.Method {
		case http.MethodGet:
			h.myMenu.HandleGet(w, r)
		case http.MethodPut:
			h.myMenu.HandleUpdate(w, r)
		case http.MethodDelete:
			h.myMenu.HandleDelete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/my-menu/", authMiddleware.Authenticate(rl.LimitByUser(http.HandlerFunc(myMenuDetailRouteHandler))))
}

// repositories はリポジトリ群をまとめた構造体
type repositories struct {
	analysis      repository.AnalysisRepository
	storage       repository.StorageRepository
	weightRecord  repository.WeightRecordRepository
	weightGoal    repository.WeightGoalRepository
	nutritionGoal repository.NutritionGoalRepository
	myMenu        repository.MyMenuRepository
}

func initRepositories(firestoreClient *firestore.Client, storageRepo repository.StorageRepository) repositories {
	analysisRepo, err := repository.NewAnalysisRepositoryFirestore(firestoreClient, storageRepo)
	if err != nil {
		log.Fatalf("Failed to initialize AnalysisRepository: %v", err)
	}

	weightRecordRepo, weightGoalRepo, err := repository.NewWeightRepositories(firestoreClient)
	if err != nil {
		log.Fatalf("Failed to initialize WeightRepositories: %v", err)
	}

	nutritionGoalRepo, err := repository.NewNutritionGoalRepository(firestoreClient)
	if err != nil {
		log.Fatalf("Failed to initialize NutritionGoalRepository: %v", err)
	}

	myMenuRepo, err := repository.NewMyMenuRepository(firestoreClient)
	if err != nil {
		log.Fatalf("Failed to initialize MyMenuRepository: %v", err)
	}

	return repositories{
		analysis:      analysisRepo,
		storage:       storageRepo,
		weightRecord:  weightRecordRepo,
		weightGoal:    weightGoalRepo,
		nutritionGoal: nutritionGoalRepo,
		myMenu:        myMenuRepo,
	}
}

func initGeminiClient() *RealGeminiClient {
	classifier, err := gemini.NewClassifier(120 * time.Second)
	if err != nil {
		log.Fatalf("Failed to initialize Gemini Classifier: %v", err)
	}
	textParser, err := gemini.NewTextParser(120 * time.Second)
	if err != nil {
		log.Fatalf("Failed to initialize Gemini TextParser: %v", err)
	}
	calculator, err := gemini.NewNutritionCalculator(120 * time.Second)
	if err != nil {
		log.Fatalf("Failed to initialize Gemini NutritionCalculator: %v", err)
	}
	log.Println("Gemini API clients initialized")
	return &RealGeminiClient{
		classifier: classifier,
		textParser: textParser,
		calculator: calculator,
	}
}

func initAuthMiddleware(ctx context.Context, firebaseCredentials string) middleware.Authenticator {
	if middleware.IsDevMode() {
		log.Println("WARNING: Running in development mode with mock authentication")
		return middleware.NewDevAuthMiddleware()
	}
	firebaseAuthService, err := service.NewFirebaseAuthService(ctx, firebaseCredentials)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase Auth Service: %v", err)
	}
	log.Println("Firebase Auth Service initialized")
	return middleware.NewAuthMiddleware(firebaseAuthService)
}

func run() error {
	ctx := context.Background()

	// Firestoreクライアントの初期化
	firebaseCredentials := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	firestoreClient, err := database.NewFirestoreClient(ctx, firebaseCredentials)
	if err != nil {
		log.Fatalf("Failed to connect to Firestore: %v", err)
	}
	defer func() {
		if err := firestoreClient.Close(); err != nil {
			log.Printf("Error closing Firestore client: %v", err)
		}
	}()
	log.Println("Firestore connection established")

	// Cloud Storageクライアントの初期化
	storageClient, err := storage.NewStorageClient(ctx, firebaseCredentials)
	if err != nil {
		log.Fatalf("Failed to connect to Cloud Storage: %v", err)
	}
	defer func() {
		if err := storageClient.Close(); err != nil {
			log.Printf("Error closing Storage client: %v", err)
		}
	}()
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

	repos := initRepositories(firestoreClient, storageRepo)
	geminiClient := initGeminiClient()
	foodService := service.NewFoodService(geminiClient, storageRepo)
	authMiddleware := initAuthMiddleware(ctx, firebaseCredentials)

	// ハンドラーの初期化
	h := handlers{
		health:        handler.NewHealthHandler(),
		analyze:       handler.NewAnalyzeHandler(foodService, repos.analysis, storageRepo),
		status:        handler.NewStatusHandler(repos.analysis),
		history:       handler.NewHistoryHandler(repos.analysis, geminiClient),
		historyDelete: handler.NewHistoryDeleteHandler(repos.analysis),
		image:         handler.NewImageHandler(storageRepo),
		dailyMeals:    handler.NewDailyMealsHandler(repos.analysis),
		skipMeal:      handler.NewSkipMealHandler(repos.analysis),
		weightRecord:  handler.NewWeightRecordHandler(repos.weightRecord, repos.weightGoal),
		weightGoal:    handler.NewWeightGoalHandler(repos.weightGoal),
		nutritionGoal: handler.NewNutritionGoalHandler(repos.nutritionGoal, repos.weightGoal),
		myMenu:        handler.NewMyMenuHandler(repos.myMenu, repos.analysis),
	}

	// ワーカーの初期化
	analysisWorker := worker.NewAnalysisWorker(foodService, repos.analysis, 5*time.Second)

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	go analysisWorker.Start(workerCtx)

	// レート制限ミドルウェアの初期化
	rateLimitConfig := middleware.LoadRateLimitConfig()
	log.Printf("Rate limit config: IP=%v rps (burst %d), User=%v rps (burst %d), Gemini=%v rps (burst %d)",
		rateLimitConfig.IPRateLimit, rateLimitConfig.IPBurstSize,
		rateLimitConfig.UserRateLimit, rateLimitConfig.UserBurstSize,
		rateLimitConfig.GeminiRateLimit, rateLimitConfig.GeminiBurstSize)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(rateLimitConfig)
	defer rateLimitMiddleware.Stop()

	// ルーティング
	mux := http.NewServeMux()
	setupRoutes(mux, h, authMiddleware, rateLimitMiddleware)

	// ミドルウェアチェーンを構築（リクエスト処理順: セキュリティヘッダー → IPレート制限 → CORS → mux）
	allowedOrigins := parseAllowedOrigins(os.Getenv("ALLOWED_ORIGINS"))
	corsHandler := enableCORS(mux, allowedOrigins)
	ipLimited := rateLimitMiddleware.LimitByIP(corsHandler)
	secureHandler := middleware.SecurityHeaders(ipLimited)

	// HTTPサーバー設定
	server := &http.Server{
		Addr:              ":8080",
		Handler:           secureHandler,
		ReadTimeout:       60 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      150 * time.Second, // Gemini API処理（最大120秒）を考慮
		IdleTimeout:       120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")

		workerCancel()
		log.Println("Worker stopped")

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

		w.Header().Set("Vary", "Origin")
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
	origins := make(map[string]struct{})

	// 開発環境の場合のみlocalhostオリジンを許可
	if middleware.IsDevMode() {
		origins["http://localhost:3000"] = struct{}{}
		origins["http://localhost:3001"] = struct{}{}
		origins["http://localhost:3002"] = struct{}{}
		origins["https://utikomi.exe.xyz:3000"] = struct{}{}
	}

	for _, origin := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(origin); trimmed != "" {
			origins[trimmed] = struct{}{}
		}
	}

	return origins
}
