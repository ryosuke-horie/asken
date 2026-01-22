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

	"github.com/ryosuke-horie/asken/backend/internal/handler"
	"github.com/ryosuke-horie/asken/backend/internal/repository"
	"github.com/ryosuke-horie/asken/backend/internal/service"
	"github.com/ryosuke-horie/asken/backend/internal/worker"
	"github.com/ryosuke-horie/asken/backend/pkg/database"
	"github.com/ryosuke-horie/asken/backend/pkg/gemini"
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

	// ハンドラーの初期化（リポジトリを渡す）
	analyzeHandler := handler.NewAnalyzeHandler(foodService, analysisRepo, userRepo)
	statusHandler := handler.NewStatusHandler(analysisRepo)
	historyHandler := handler.NewHistoryHandler(analysisRepo)
	historyDeleteHandler := handler.NewHistoryDeleteHandler(analysisRepo)
	imageHandler := handler.NewImageHandler("uploads")
	dailyMealsHandler := handler.NewDailyMealsHandler(analysisRepo)

	// ワーカーの初期化
	analysisWorker := worker.NewAnalysisWorker(foodService, analysisRepo, 5*time.Second)

	// ワーカー用のコンテキスト
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	// ワーカーを別ゴルーチンで起動
	go analysisWorker.Start(workerCtx)

	// ルーティング
	mux := http.NewServeMux()

	// POST /api/analyze - 画像アップロード
	// GET /api/analyze/:id - ステータス取得
	analyzeRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			analyzeHandler.Handle(w, r)
		} else if r.Method == http.MethodGet {
			statusHandler.Handle(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}

	// 両方のパスパターンを登録
	mux.HandleFunc("/api/analyze", analyzeRouteHandler)   // POST用
	mux.HandleFunc("/api/analyze/", analyzeRouteHandler)  // GET /api/analyze/:id 用

	// 履歴エンドポイント
	// GET /api/history - 履歴一覧
	mux.HandleFunc("/api/history", historyHandler.HandleList)

	// GET /api/history/:id - 履歴詳細
	// DELETE /api/history/:id - 履歴削除
	historyDetailRouteHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			historyHandler.HandleDetail(w, r)
		} else if r.Method == http.MethodDelete {
			historyDeleteHandler.Handle(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
	mux.HandleFunc("/api/history/", historyDetailRouteHandler)

	// 画像配信エンドポイント
	// GET /api/images/:filename
	mux.HandleFunc("/api/images/", imageHandler.Handle)

	// 日次食事データエンドポイント
	// GET /api/meals/daily
	mux.HandleFunc("/api/meals/daily", dailyMealsHandler.Handle)

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
			log.Fatalf("Server shutdown error: %v", err)
		}
	}()

	log.Println("Server starting on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

// enableCORS はCORSヘッダーを追加するミドルウェア
func enableCORS(next http.Handler, allowedOrigins map[string]struct{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			if _, ok := allowedOrigins[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

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
		"http://localhost:3000":    {},
		"http://localhost:3001":    {},
		"https://asken.exe.xyz:3000": {},
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
