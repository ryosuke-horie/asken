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

	// 依存関係の初期化
	classifier := gemini.NewClassifier(120 * time.Second)
	textParser := gemini.NewTextParser(120 * time.Second)
	calculator := gemini.NewNutritionCalculator(120 * time.Second)
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

	// ハンドラーの初期化（リポジトリを渡す）
	authHandler := handler.NewAuthHandler(authService, userRepo)
	analyzeHandler := handler.NewAnalyzeHandler(foodService, analysisRepo)
	statusHandler := handler.NewStatusHandler(analysisRepo)
	historyHandler := handler.NewHistoryHandler(analysisRepo)
	historyDeleteHandler := handler.NewHistoryDeleteHandler(analysisRepo)
	imageHandler := handler.NewImageHandler("uploads")
	dailyMealsHandler := handler.NewDailyMealsHandler(analysisRepo)
	weightHandler := handler.NewWeightHandler(weightRepo)
	mylistHandler := handler.NewMylistHandler(mylistRepo, analysisRepo, foodService)

	// ワーカーの初期化
	analysisWorker := worker.NewAnalysisWorker(foodService, analysisRepo, 5*time.Second)

	// ワーカー用のコンテキスト
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	// ワーカーを別ゴルーチンで起動
	go analysisWorker.Start(workerCtx)

	// ルーティング
	mux := http.NewServeMux()

	// 認証エンドポイント（認証不要）
	mux.HandleFunc("/api/auth/register", authHandler.HandleRegister)
	mux.HandleFunc("/api/auth/login", authHandler.HandleLogin)

	// POST /api/analyze - 画像アップロード（認証必須）
	// GET /api/analyze/:id - ステータス取得（認証必須）
	analyzeRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			analyzeHandler.Handle(w, r)
		case http.MethodGet:
			statusHandler.Handle(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}

	// 認証が必要なエンドポイントにミドルウェアを適用
	mux.Handle("/api/analyze", authMiddleware.Authenticate(http.HandlerFunc(analyzeRouteHandler)))
	mux.Handle("/api/analyze/", authMiddleware.Authenticate(http.HandlerFunc(analyzeRouteHandler)))

	// POST /api/upload-image - 画像アップロード（分析なし、認証必須）
	uploadImageRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			analyzeHandler.HandleUploadImage(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/upload-image", authMiddleware.Authenticate(http.HandlerFunc(uploadImageRouteHandler)))

	// 履歴エンドポイント（認証必須）
	// GET /api/history - 履歴一覧
	mux.Handle("/api/history", authMiddleware.Authenticate(http.HandlerFunc(historyHandler.HandleList)))

	// GET /api/history/:id - 履歴詳細
	// DELETE /api/history/:id - 履歴削除
	historyDetailRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			historyHandler.HandleDetail(w, r)
		case http.MethodDelete:
			historyDeleteHandler.Handle(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/history/", authMiddleware.Authenticate(http.HandlerFunc(historyDetailRouteHandler)))

	// 画像配信エンドポイント（認証必須）
	// GET /api/images/:filename
	mux.Handle("/api/images/", authMiddleware.Authenticate(http.HandlerFunc(imageHandler.Handle)))

	// 日次食事データエンドポイント（認証必須）
	// GET /api/meals/daily
	mux.Handle("/api/meals/daily", authMiddleware.Authenticate(http.HandlerFunc(dailyMealsHandler.Handle)))

	// マイリストからの食事記録エンドポイント（認証必須）
	// POST /api/meals/from-mylist
	mealsFromMylistRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			mylistHandler.HandleRecordFromMylist(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/meals/from-mylist", authMiddleware.Authenticate(http.HandlerFunc(mealsFromMylistRouteHandler)))

	// 体重記録エンドポイント（認証必須）
	// POST /api/weight-records - 体重記録作成
	// GET /api/weight-records - 体重記録一覧取得
	weightRecordsRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			weightHandler.HandleCreateRecord(w, r)
		case http.MethodGet:
			weightHandler.HandleGetRecords(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/weight-records", authMiddleware.Authenticate(http.HandlerFunc(weightRecordsRouteHandler)))

	// 目標体重エンドポイント（認証必須）
	// GET /api/weight-goal - 目標取得
	// PUT /api/weight-goal - 目標更新
	weightGoalRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			weightHandler.HandleGetGoal(w, r)
		case http.MethodPut:
			weightHandler.HandleUpdateGoal(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/weight-goal", authMiddleware.Authenticate(http.HandlerFunc(weightGoalRouteHandler)))

	// マイリストエンドポイント（認証必須）
	// GET /api/mylist - 一覧取得
	// POST /api/mylist - 新規作成
	mylistRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			mylistHandler.HandleList(w, r)
		} else if r.Method == http.MethodPost {
			mylistHandler.HandleCreate(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/mylist", authMiddleware.Authenticate(http.HandlerFunc(mylistRouteHandler)))

	// PUT /api/mylist/reorder - 並び替え
	mylistReorderRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			mylistHandler.HandleReorder(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/mylist/reorder", authMiddleware.Authenticate(http.HandlerFunc(mylistReorderRouteHandler)))

	// POST /api/mylist/analyze - AI分析
	mylistAnalyzeRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			mylistHandler.HandleAnalyze(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/mylist/analyze", authMiddleware.Authenticate(http.HandlerFunc(mylistAnalyzeRouteHandler)))

	// GET /api/mylist/:id - 単体取得
	// PUT /api/mylist/:id - 更新
	// DELETE /api/mylist/:id - 削除
	mylistDetailRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			mylistHandler.HandleGetByID(w, r)
		} else if r.Method == http.MethodPut {
			mylistHandler.HandleUpdate(w, r)
		} else if r.Method == http.MethodDelete {
			mylistHandler.HandleDelete(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.Handle("/api/mylist/", authMiddleware.Authenticate(http.HandlerFunc(mylistDetailRouteHandler)))

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
