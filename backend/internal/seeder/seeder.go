package seeder

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
)

// Config はSeederの設定
type Config struct {
	UserCount        int
	AnalysesPerUser  int
	WeightRecordDays int
	CleanFirst       bool
	Verbose          bool
}

// Seeder はデータベースにテストデータを投入する構造体
type Seeder struct {
	db          *sql.DB
	authService *service.AuthService
	config      Config
}

// NewSeeder は新しいSeederを作成する
func NewSeeder(db *sql.DB, authService *service.AuthService, config Config) *Seeder {
	return &Seeder{
		db:          db,
		authService: authService,
		config:      config,
	}
}

// Run はシードを実行する
func (s *Seeder) Run(ctx context.Context) error {
	if s.config.CleanFirst {
		if err := s.clean(ctx); err != nil {
			return fmt.Errorf("データクリーンに失敗: %w", err)
		}
		s.log("既存データを削除しました")
	}

	if s.config.UserCount == 0 {
		s.log("ユーザー数が0のため、データ生成をスキップします")
		return nil
	}

	// ユーザーを作成
	users, err := s.seedUsers(ctx)
	if err != nil {
		return fmt.Errorf("ユーザーシードに失敗: %w", err)
	}
	s.log(fmt.Sprintf("%d件のユーザーを作成しました", len(users)))

	// 各ユーザーに対して分析データを作成
	for _, user := range users {
		if err := s.seedAnalysesForUser(ctx, user.ID); err != nil {
			return fmt.Errorf("ユーザー %s の分析シードに失敗: %w", user.Email, err)
		}
		s.log(fmt.Sprintf("ユーザー %s に %d 件の分析データを作成しました", user.Email, s.config.AnalysesPerUser))
	}

	// 体重記録データを作成
	if s.config.WeightRecordDays > 0 {
		for _, user := range users {
			recordCount, err := s.seedWeightRecordsForUser(ctx, user.ID)
			if err != nil {
				return fmt.Errorf("ユーザー %s の体重記録シードに失敗: %w", user.Email, err)
			}
			s.log(fmt.Sprintf("ユーザー %s に %d 件の体重記録データを作成しました", user.Email, recordCount))

			if err := s.seedWeightGoalForUser(ctx, user.ID); err != nil {
				return fmt.Errorf("ユーザー %s の目標体重シードに失敗: %w", user.Email, err)
			}
			s.log(fmt.Sprintf("ユーザー %s に目標体重を設定しました", user.Email))
		}
	}

	// マイリストデータを作成
	for _, user := range users {
		count, err := s.seedMylistForUser(ctx, user.ID)
		if err != nil {
			return fmt.Errorf("ユーザー %s のマイリストシードに失敗: %w", user.Email, err)
		}
		s.log(fmt.Sprintf("ユーザー %s に %d 件のマイリストアイテムを作成しました", user.Email, count))
	}

	return nil
}

// clean は既存データを削除する
func (s *Seeder) clean(ctx context.Context) error {
	queries := []string{
		"DELETE FROM analysis_results",
		"DELETE FROM analysis_requests",
		"DELETE FROM weight_records",
		"DELETE FROM weight_goals",
		"DELETE FROM mylist_items",
		"DELETE FROM users",
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("クエリ実行に失敗 (%s): %w", query, err)
		}
	}

	return tx.Commit()
}

// seedUsers はテストユーザーを作成する
func (s *Seeder) seedUsers(ctx context.Context) ([]*repository.User, error) {
	var users []*repository.User

	for i := 0; i < s.config.UserCount; i++ {
		var testUser TestUser
		if i < len(DefaultTestUsers) {
			testUser = DefaultTestUsers[i]
		} else {
			testUser = TestUser{
				Email:    fmt.Sprintf("user%d@example.com", i+1),
				Password: "password123",
				Name:     fmt.Sprintf("テストユーザー%d", i+1),
			}
		}

		// パスワードをハッシュ化
		hashedPassword, err := s.authService.HashPassword(testUser.Password)
		if err != nil {
			return nil, fmt.Errorf("パスワードハッシュ化に失敗: %w", err)
		}

		// ユーザー作成
		user, err := s.createUser(ctx, testUser.Email, testUser.Name, hashedPassword)
		if err != nil {
			return nil, fmt.Errorf("ユーザー作成に失敗 (%s): %w", testUser.Email, err)
		}

		users = append(users, user)
	}

	return users, nil
}

// createUser はユーザーをDBに作成する
func (s *Seeder) createUser(ctx context.Context, email, name, passwordHash string) (*repository.User, error) {
	query := `
		INSERT INTO users (email, name, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, name, password_hash, created_at, updated_at
	`

	var user repository.User
	var returnedName sql.NullString

	err := s.db.QueryRowContext(ctx, query, email, name, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&returnedName,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if returnedName.Valid {
		user.Name = returnedName.String
	}

	return &user, nil
}

// seedAnalysesForUser はユーザーに対して分析データを作成する
func (s *Seeder) seedAnalysesForUser(ctx context.Context, userID uuid.UUID) error {
	dates := GeneratePastDates(7)

	for i := 0; i < s.config.AnalysesPerUser; i++ {
		// 日付とミールタイプを決定
		dateIndex := i % len(dates)
		mealTypeIndex := i % len(MealTypes)
		mealDate := FormatDateForDB(dates[dateIndex])
		mealType := MealTypes[mealTypeIndex]

		// ステータスを決定（5件の場合: 3件completed, 1件pending, 1件failed）
		var status repository.AnalysisStatus
		switch {
		case i < 3:
			status = repository.StatusCompleted
		case i == 3:
			status = repository.StatusPending
		default:
			status = repository.StatusFailed
		}

		// 入力タイプを決定（5件の場合: 3件image, 2件text）
		var inputType repository.InputType
		if i < 3 {
			inputType = repository.InputTypeImage
		} else {
			inputType = repository.InputTypeText
		}

		// 分析リクエストを作成
		requestID, err := s.createAnalysisRequest(ctx, userID, mealType, mealDate, inputType, status, i)
		if err != nil {
			return fmt.Errorf("分析リクエスト作成に失敗: %w", err)
		}

		// completed の場合は分析結果も作成
		if status == repository.StatusCompleted {
			if err := s.createAnalysisResult(ctx, requestID, mealType); err != nil {
				return fmt.Errorf("分析結果作成に失敗: %w", err)
			}
		}
	}

	return nil
}

// createAnalysisRequest は分析リクエストを作成する
func (s *Seeder) createAnalysisRequest(
	ctx context.Context,
	userID uuid.UUID,
	mealType string,
	mealDate string,
	inputType repository.InputType,
	status repository.AnalysisStatus,
	index int,
) (uuid.UUID, error) {
	var query string
	var args []interface{}

	if inputType == repository.InputTypeImage {
		query = `
			INSERT INTO analysis_requests (status, input_type, image_path, meal_type, meal_date, user_id)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`
		imagePath := fmt.Sprintf("/uploads/sample_%d.jpg", index)
		args = []interface{}{status, inputType, imagePath, mealType, mealDate, userID}
	} else {
		query = `
			INSERT INTO analysis_requests (status, input_type, input_text, meal_type, meal_date, user_id)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`
		textIndex := index % len(SampleTextInputs)
		args = []interface{}{status, inputType, SampleTextInputs[textIndex], mealType, mealDate, userID}
	}

	var id uuid.UUID
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

// createAnalysisResult は分析結果を作成する
func (s *Seeder) createAnalysisResult(ctx context.Context, requestID uuid.UUID, mealType string) error {
	foods := SampleNutritionData[mealType]
	if foods == nil {
		foods = SampleNutritionData["lunch"]
	}

	foodsJSON, err := json.Marshal(foods)
	if err != nil {
		return fmt.Errorf("foods のJSON変換に失敗: %w", err)
	}

	totalCal, totalPro, totalFat, totalCarbs := CalculateTotals(foods)

	query := `
		INSERT INTO analysis_results (
			analysis_request_id,
			foods,
			total_calories,
			total_protein,
			total_fat,
			total_carbohydrates
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = s.db.ExecContext(ctx, query, requestID, foodsJSON, totalCal, totalPro, totalFat, totalCarbs)
	return err
}

// log はverboseモードの場合にログを出力する
func (s *Seeder) log(message string) {
	if s.config.Verbose {
		log.Println(message)
	}
}

// logWarning は警告を常に出力する（Verboseに関係なく）
func (s *Seeder) logWarning(message string) {
	log.Printf("WARNING: %s", message)
}

// GetSampleNutritionInfo はサンプルのNutritionInfoを返す（テスト用）
func GetSampleNutritionInfo(mealType string) []gemini.NutritionInfo {
	if foods, ok := SampleNutritionData[mealType]; ok {
		return foods
	}
	return SampleNutritionData["lunch"]
}

// seedWeightRecordsForUser はユーザーに対して体重記録データを作成する
func (s *Seeder) seedWeightRecordsForUser(ctx context.Context, userID uuid.UUID) (int, error) {
	records := GenerateWeightRecords(s.config.WeightRecordDays, DefaultWeightSeedConfig)

	query := `
		INSERT INTO weight_records (user_id, weight, recorded_at)
		VALUES ($1, $2, $3)
	`

	for _, record := range records {
		if _, err := s.db.ExecContext(ctx, query, userID, record.Weight, record.RecordedAt); err != nil {
			return 0, fmt.Errorf("体重記録の挿入に失敗: %w", err)
		}
	}

	return len(records), nil
}

// seedWeightGoalForUser はユーザーに対して目標体重を設定する
func (s *Seeder) seedWeightGoalForUser(ctx context.Context, userID uuid.UUID) error {
	query := `
		INSERT INTO weight_goals (user_id, target_weight, target_date)
		VALUES ($1, $2, $3)
	`

	targetDate := GetDefaultTargetDate()
	_, err := s.db.ExecContext(ctx, query, userID, DefaultWeightSeedConfig.TargetWeight, targetDate)
	if err != nil {
		return fmt.Errorf("目標体重の挿入に失敗: %w", err)
	}

	return nil
}

// seedMylistForUser はユーザーに対してマイリストデータを作成する
func (s *Seeder) seedMylistForUser(ctx context.Context, userID uuid.UUID) (int, error) {
	for i, item := range DefaultMylistItems {
		if err := s.createMylistItem(ctx, userID, item, i); err != nil {
			return 0, fmt.Errorf("マイリストアイテム作成に失敗: %w", err)
		}
	}
	return len(DefaultMylistItems), nil
}

// createMylistItem はマイリストアイテムを作成する
func (s *Seeder) createMylistItem(ctx context.Context, userID uuid.UUID, item MylistSeedItem, sortOrder int) error {
	foodsJSON, err := json.Marshal(item.Foods)
	if err != nil {
		return fmt.Errorf("foods のJSON変換に失敗: %w", err)
	}

	calories, protein, fat, carbs := CalculateMylistTotals(item.Foods)

	// 画像をコピーしてimage_pathを設定
	var imagePath interface{}
	if item.SeedImageSource != "" {
		copiedPath, copyErr := s.copySeedImage(item.SeedImageSource)
		if copyErr != nil {
			// 画像コピー失敗は警告として常に出力し、画像なしで続行
			s.logWarning(fmt.Sprintf("画像コピーに失敗（画像なしで続行）: %s - %v", item.SeedImageSource, copyErr))
			imagePath = nil
		} else {
			imagePath = copiedPath
		}
	}

	query := `
		INSERT INTO mylist_items (user_id, name, base_amount, unit, calories, protein, fat, carbohydrates, foods, image_path, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = s.db.ExecContext(ctx, query,
		userID,
		item.Name,
		item.BaseAmount,
		item.Unit,
		calories,
		protein,
		fat,
		carbs,
		foodsJSON,
		imagePath,
		sortOrder,
	)
	if err != nil {
		return fmt.Errorf("マイリストアイテムの挿入に失敗: %w", err)
	}

	return nil
}

// copySeedImage はシード画像をuploadsディレクトリにコピーする
func (s *Seeder) copySeedImage(sourceFileName string) (string, error) {
	// シード画像のソースパス
	sourcePath := filepath.Join("seeds", "images", sourceFileName)

	// uploadsディレクトリを作成
	uploadsDir := "uploads"
	if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
		return "", fmt.Errorf("uploadsディレクトリの作成に失敗: %w", err)
	}

	// 新しいファイル名（UUIDを使用）
	ext := filepath.Ext(sourceFileName)
	newFileName := uuid.New().String() + ext
	destPath := filepath.Join(uploadsDir, newFileName)

	// ファイルをコピー
	if err := copyFile(sourcePath, destPath); err != nil {
		return "", fmt.Errorf("ファイルコピーに失敗: %w", err)
	}

	// DBに保存するパス（/uploads/xxx.jpg形式）
	// filepath.Joinはプラットフォーム依存のセパレータを使うため、明示的にスラッシュを使用
	return "/uploads/" + newFileName, nil
}

// copyFile はファイルをコピーする
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("ソースファイルのオープンに失敗 (%s): %w", src, err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("コピー先ファイルの作成に失敗 (%s): %w", dst, err)
	}

	_, copyErr := io.Copy(destFile, sourceFile)

	// データをディスクに同期
	if syncErr := destFile.Sync(); syncErr != nil && copyErr == nil {
		copyErr = fmt.Errorf("ファイルの同期に失敗 (%s): %w", dst, syncErr)
	}

	if closeErr := destFile.Close(); closeErr != nil && copyErr == nil {
		copyErr = fmt.Errorf("ファイルのクローズに失敗 (%s): %w", dst, closeErr)
	}

	// 失敗時はコピー先ファイルを削除
	if copyErr != nil {
		_ = os.Remove(dst)
		return fmt.Errorf("ファイルコピーに失敗 (%s → %s): %w", src, dst, copyErr)
	}

	return nil
}
