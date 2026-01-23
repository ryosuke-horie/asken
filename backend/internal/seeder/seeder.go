package seeder

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/asken/backend/internal/repository"
	"github.com/ryosuke-horie/asken/backend/internal/service"
	"github.com/ryosuke-horie/asken/backend/pkg/gemini"
)

// Config はSeederの設定
type Config struct {
	UserCount      int
	AnalysesPerUser int
	CleanFirst     bool
	Verbose        bool
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

	return nil
}

// clean は既存データを削除する
func (s *Seeder) clean(ctx context.Context) error {
	queries := []string{
		"DELETE FROM analysis_results",
		"DELETE FROM analysis_requests",
		"DELETE FROM users",
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer tx.Rollback()

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("クエリ実行に失敗 (%s): %w", query, err)
		}
	}

	return tx.Commit()
}

// seedUsers はテストユーザーを作成する
func (s *Seeder) seedUsers(ctx context.Context) ([]*repository.User, error) {
	testUsers := DefaultTestUsers
	if s.config.UserCount < len(testUsers) {
		testUsers = testUsers[:s.config.UserCount]
	}

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

// GetSampleNutritionInfo はサンプルのNutritionInfoを返す（テスト用）
func GetSampleNutritionInfo(mealType string) []gemini.NutritionInfo {
	if foods, ok := SampleNutritionData[mealType]; ok {
		return foods
	}
	return SampleNutritionData["lunch"]
}
