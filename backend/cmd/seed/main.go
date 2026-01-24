package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/internal/seeder"
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/database"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("シード実行に失敗: %v", err)
	}
}

func run() error {
	// コマンドラインフラグを定義
	users := flag.Int("users", 3, "生成するユーザー数")
	analyses := flag.Int("analyses", 5, "ユーザーあたりの分析数")
	clean := flag.Bool("clean", true, "既存データを削除")
	verbose := flag.Bool("verbose", false, "詳細ログ出力")
	flag.Parse()

	// 環境変数からDATABASE_URLを取得
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is required")
	}

	// データベース接続
	db, err := database.NewPostgresDB(database.Config{
		DatabaseURL: databaseURL,
	})
	if err != nil {
		return fmt.Errorf("データベース接続に失敗: %w", err)
	}
	defer db.Close()

	if *verbose {
		log.Println("データベースに接続しました")
	}

	// 認証サービスの初期化（パスワードハッシュ化に使用）
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "seed-dummy-secret"
	}
	authService := service.NewAuthService(jwtSecret, 24*time.Hour)

	// Seederの設定
	config := seeder.Config{
		UserCount:       *users,
		AnalysesPerUser: *analyses,
		CleanFirst:      *clean,
		Verbose:         *verbose,
	}

	// Seeder実行
	s := seeder.NewSeeder(db, authService, config)

	ctx := context.Background()
	if err := s.Run(ctx); err != nil {
		return fmt.Errorf("シード実行に失敗: %w", err)
	}

	fmt.Println("シードが正常に完了しました")
	fmt.Printf("  ユーザー数: %d\n", *users)
	fmt.Printf("  ユーザーあたりの分析数: %d\n", *analyses)
	fmt.Printf("  合計分析数: %d\n", *users*(*analyses))

	if *users > 0 {
		fmt.Println("\nテストユーザー:")
		for i := 0; i < *users && i < len(seeder.DefaultTestUsers); i++ {
			user := seeder.DefaultTestUsers[i]
			fmt.Printf("  - %s (パスワード: %s)\n", user.Email, user.Password)
		}
	}

	return nil
}
