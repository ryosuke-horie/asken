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
	auth          *handler.AuthHandler
	analyze       *handler.AnalyzeHandler
	status        *handler.StatusHandler
	history       *handler.HistoryHandler
	historyDelete *handler.HistoryDeleteHandler
	image         *handler.ImageHandler
	dailyMeals    *handler.DailyMealsHandler
	weight        *handler.WeightHandler
	mylist        *handler.MylistHandler
	skipMeal      *handler.SkipMealHandler
	condition     *handler.ConditionHandler
	training      *handler.TrainingHandler
}

func setupRoutes(mux *http.ServeMux, h handlers, authMiddleware *middleware.AuthMiddleware) {
	// 認証エンドポイント（認証不要）
	mux.HandleFunc("/api/auth/register", h.auth.HandleRegister)
	mux.HandleFunc("/api/auth/login", h.auth.HandleLogin)

	// 画像配信エンドポイント（認証不要 - UUIDファイル名で保護）
	mux.HandleFunc("/api/images/", h.image.Handle)

	// 認証が必要なエンドポイント
	setupAnalyzeRoutes(mux, h, authMiddleware)
	setupHistoryRoutes(mux, h, authMiddleware)
	setupMealsRoutes(mux, h, authMiddleware)
	setupWeightRoutes(mux, h, authMiddleware)
	setupMylistRoutes(mux, h, authMiddleware)
	setupConditionRoutes(mux, h, authMiddleware)
	setupTrainingRoutes(mux, h, authMiddleware)
}

func setupAnalyzeRoutes(mux *http.ServeMux, h handlers, authMiddleware *middleware.AuthMiddleware) {
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

func setupHistoryRoutes(mux *http.ServeMux, h handlers, authMiddleware *middleware.AuthMiddleware) {
	mux.Handle("/api/history", authMiddleware.Authenticate(http.HandlerFunc(h.history.HandleList)))

	historyDetailRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.history.HandleDetail(w, r)
		case http.MethodDelete:
			h.historyDelete.Handle(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/history/", authMiddleware.Authenticate(http.HandlerFunc(historyDetailRouteHandler)))
}

func setupMealsRoutes(mux *http.ServeMux, h handlers, authMiddleware *middleware.AuthMiddleware) {
	mux.Handle("/api/meals/daily", authMiddleware.Authenticate(http.HandlerFunc(h.dailyMeals.Handle)))

	mealsFromMylistRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.mylist.HandleRecordFromMylist(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/meals/from-mylist", authMiddleware.Authenticate(http.HandlerFunc(mealsFromMylistRouteHandler)))

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

func setupWeightRoutes(mux *http.ServeMux, h handlers, authMiddleware *middleware.AuthMiddleware) {
	weightRecordsRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.weight.HandleCreateRecord(w, r)
		case http.MethodGet:
			h.weight.HandleGetRecords(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/weight-records", authMiddleware.Authenticate(http.HandlerFunc(weightRecordsRouteHandler)))

	weightGoalRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.weight.HandleGetGoal(w, r)
		case http.MethodPut:
			h.weight.HandleUpdateGoal(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/weight-goal", authMiddleware.Authenticate(http.HandlerFunc(weightGoalRouteHandler)))
}

func setupMylistRoutes(mux *http.ServeMux, h handlers, authMiddleware *middleware.AuthMiddleware) {
	mylistRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.mylist.HandleList(w, r)
		case http.MethodPost:
			h.mylist.HandleCreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/mylist", authMiddleware.Authenticate(http.HandlerFunc(mylistRouteHandler)))

	mylistReorderRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			h.mylist.HandleReorder(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/mylist/reorder", authMiddleware.Authenticate(http.HandlerFunc(mylistReorderRouteHandler)))

	mylistAnalyzeRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.mylist.HandleAnalyze(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/mylist/analyze", authMiddleware.Authenticate(http.HandlerFunc(mylistAnalyzeRouteHandler)))

	mylistDetailRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.mylist.HandleGetByID(w, r)
		case http.MethodPut:
			h.mylist.HandleUpdate(w, r)
		case http.MethodDelete:
			h.mylist.HandleDelete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/mylist/", authMiddleware.Authenticate(http.HandlerFunc(mylistDetailRouteHandler)))
}

func setupConditionRoutes(mux *http.ServeMux, h handlers, authMiddleware *middleware.AuthMiddleware) {
	conditionRecordsRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.condition.HandleCreateRecord(w, r)
		case http.MethodGet:
			h.condition.HandleGetRecord(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/condition-records", authMiddleware.Authenticate(http.HandlerFunc(conditionRecordsRouteHandler)))
}

//nolint:gocyclo // TODO: リファクタリングで複雑度を下げる
func setupTrainingRoutes(mux *http.ServeMux, h handlers, authMiddleware *middleware.AuthMiddleware) {
	// 場所一覧・作成
	locationsRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.training.HandleListLocations(w, r)
		case http.MethodPost:
			h.training.HandleCreateLocation(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/training/locations", authMiddleware.Authenticate(http.HandlerFunc(locationsRouteHandler)))

	// メニュー一覧・作成
	menusRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.training.HandleListMenus(w, r)
		case http.MethodPost:
			h.training.HandleCreateMenu(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/training/menus", authMiddleware.Authenticate(http.HandlerFunc(menusRouteHandler)))

	// メニュー詳細・削除
	menusDetailRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			h.training.HandleDeleteMenu(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/training/menus/", authMiddleware.Authenticate(http.HandlerFunc(menusDetailRouteHandler)))

	// 練習記録一覧・作成
	recordsRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.training.HandleListRecords(w, r)
		case http.MethodPost:
			h.training.HandleCreateRecord(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/training/records", authMiddleware.Authenticate(http.HandlerFunc(recordsRouteHandler)))

	// 練習記録詳細・更新・削除
	recordsDetailRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			h.training.HandleUpdateRecord(w, r)
		case http.MethodDelete:
			h.training.HandleDeleteRecord(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/training/records/", authMiddleware.Authenticate(http.HandlerFunc(recordsDetailRouteHandler)))

	// メニュー提案
	suggestMenuRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.training.HandleSuggestMenu(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/training/suggest-menu", authMiddleware.Authenticate(http.HandlerFunc(suggestMenuRouteHandler)))

	// 器具名正規化
	normalizeEquipmentRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.training.HandleNormalizeEquipment(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/training/normalize-equipment", authMiddleware.Authenticate(http.HandlerFunc(normalizeEquipmentRouteHandler)))

	// 器具の更新・削除
	equipmentDetailRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			h.training.HandleUpdateEquipment(w, r)
		case http.MethodDelete:
			h.training.HandleDeleteEquipment(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/training/equipment/", authMiddleware.Authenticate(http.HandlerFunc(equipmentDetailRouteHandler)))

	// 場所詳細・更新・削除、器具一覧・作成
	locationsDetailRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// /api/training/locations/{id}/equipment のパターンをチェック
		if strings.Contains(path, "/equipment") {
			switch r.Method {
			case http.MethodGet:
				h.training.HandleListEquipment(w, r)
			case http.MethodPost:
				h.training.HandleCreateEquipment(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}
		// /api/training/locations/{id} のパターン
		switch r.Method {
		case http.MethodPut:
			h.training.HandleUpdateLocation(w, r)
		case http.MethodDelete:
			h.training.HandleDeleteLocation(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/training/locations/", authMiddleware.Authenticate(http.HandlerFunc(locationsDetailRouteHandler)))
}

func run() error {
	// 環境変数から設定を読み込み
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// データベース接続
	db, err := database.NewPostgresDB(database.Config{
		DatabaseURL: databaseURL,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connection established")

	// リポジトリの初期化
	analysisRepo := repository.NewAnalysisRepository(db)
	userRepo := repository.NewUserRepository(db)
	weightRepo := repository.NewWeightRepository(db)
	mylistRepo := repository.NewMylistRepository(db)
	conditionRepo := repository.NewConditionRepository(db)
	trainingRepo := repository.NewTrainingRepository(db)

	// 依存関係の初期化
	classifier := gemini.NewClassifier(120 * time.Second)
	textParser := gemini.NewTextParser(120 * time.Second)
	calculator := gemini.NewNutritionCalculator(120 * time.Second)
	menuSuggester := gemini.NewMenuSuggester(120 * time.Second)
	equipmentNormalizer := gemini.NewEquipmentNormalizer(120 * time.Second)
	geminiClient := &RealGeminiClient{
		classifier: classifier,
		textParser: textParser,
		calculator: calculator,
	}

	foodService := service.NewFoodService(geminiClient)

	// 認証サービスの初期化
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production"
		log.Println("WARNING: JWT_SECRET not set, using default secret. Set JWT_SECRET in production!")
	}
	authService := service.NewAuthService(jwtSecret, 24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// ハンドラーの初期化
	h := handlers{
		auth:          handler.NewAuthHandler(authService, userRepo),
		analyze:       handler.NewAnalyzeHandler(foodService, analysisRepo),
		status:        handler.NewStatusHandler(analysisRepo),
		history:       handler.NewHistoryHandler(analysisRepo),
		historyDelete: handler.NewHistoryDeleteHandler(analysisRepo),
		image:         handler.NewImageHandler("uploads"),
		dailyMeals:    handler.NewDailyMealsHandler(analysisRepo),
		weight:        handler.NewWeightHandler(weightRepo),
		mylist:        handler.NewMylistHandler(mylistRepo, analysisRepo, foodService),
		skipMeal:      handler.NewSkipMealHandler(analysisRepo),
		condition:     handler.NewConditionHandler(conditionRepo),
		training:      handler.NewTrainingHandler(trainingRepo, menuSuggester, equipmentNormalizer),
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

	// CORSミドルウェアを適用
	allowedOrigins := parseAllowedOrigins(os.Getenv("ALLOWED_ORIGINS"))
	corsHandler := enableCORS(mux, allowedOrigins)

	// HTTPサーバー設定
	server := &http.Server{
		Addr:         ":8080",
		Handler:      corsHandler,
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
