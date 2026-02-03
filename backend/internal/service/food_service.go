package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
)

// GeminiClient はGemini APIクライアントのインターフェース（モック可能）
type GeminiClient interface {
	ClassifyFoods(ctx context.Context, imagePath string) ([]gemini.FoodItem, error)
	ClassifyFoodsFromData(ctx context.Context, imageData []byte, mimeType string) ([]gemini.FoodItem, error)
	ParseTextToFoods(ctx context.Context, inputText string) ([]gemini.FoodItem, error)
	CalculateNutrition(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error)
}

// ImageDownloader は画像のダウンロードを担当するインターフェース
type ImageDownloader interface {
	Download(ctx context.Context, objectName string) ([]byte, error)
}

// AnalysisResult は分析結果を表す構造体
type AnalysisResult struct {
	Foods              []gemini.NutritionInfo `json:"foods"`
	TotalCalories      float64                `json:"total_calories"`
	TotalProtein       float64                `json:"total_protein"`
	TotalFat           float64                `json:"total_fat"`
	TotalCarbohydrates float64                `json:"total_carbohydrates"`
}

// FoodServiceInterface は食品分析サービスのインターフェース
type FoodServiceInterface interface {
	AnalyzeFoodImage(ctx context.Context, imagePath string) (*AnalysisResult, error)
	AnalyzeFoodText(ctx context.Context, inputText string) (*AnalysisResult, error)
}

// FoodService は食品分析サービス
type FoodService struct {
	geminiClient    GeminiClient
	imageDownloader ImageDownloader
}

// NewFoodService は新しいFoodServiceを作成
func NewFoodService(geminiClient GeminiClient, imageDownloader ImageDownloader) *FoodService {
	return &FoodService{
		geminiClient:    geminiClient,
		imageDownloader: imageDownloader,
	}
}

// AnalyzeFoodImage は画像から食材を分析し、栄養素を計算する
func (s *FoodService) AnalyzeFoodImage(ctx context.Context, imagePath string) (*AnalysisResult, error) {
	var foods []gemini.FoodItem
	var err error

	// Cloud Storageのパス（uploads/で始まる）かローカルファイルかを判定
	if strings.HasPrefix(imagePath, "uploads/") && s.imageDownloader != nil {
		// Cloud Storageから画像をダウンロード
		imageData, downloadErr := s.imageDownloader.Download(ctx, imagePath)
		if downloadErr != nil {
			return nil, fmt.Errorf("画像ダウンロードエラー: %w", downloadErr)
		}

		// MIMEタイプを判定（拡張子から）
		mimeType := detectMimeTypeFromPath(imagePath)

		// バイトデータから食材分類
		foods, err = s.geminiClient.ClassifyFoodsFromData(ctx, imageData, mimeType)
	} else {
		// ローカルファイルから食材分類
		foods, err = s.geminiClient.ClassifyFoods(ctx, imagePath)
	}

	if err != nil {
		return nil, fmt.Errorf("食材分類エラー: %w", err)
	}

	// Step 2: 栄養素計算
	nutritionList, err := s.geminiClient.CalculateNutrition(ctx, foods)
	if err != nil {
		return nil, fmt.Errorf("栄養素計算エラー: %w", err)
	}

	// 合計カロリー・栄養素を計算
	totalCal, totalPro, totalFat, totalCarbs := calculateTotals(nutritionList)

	// AnalysisResultを返却
	return &AnalysisResult{
		Foods:              nutritionList,
		TotalCalories:      totalCal,
		TotalProtein:       totalPro,
		TotalFat:           totalFat,
		TotalCarbohydrates: totalCarbs,
	}, nil
}

// detectMimeTypeFromPath はファイルパスの拡張子からMIMEタイプを判定する
func detectMimeTypeFromPath(filePath string) string {
	lowerPath := strings.ToLower(filePath)
	switch {
	case strings.HasSuffix(lowerPath, ".jpg"), strings.HasSuffix(lowerPath, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(lowerPath, ".png"):
		return "image/png"
	case strings.HasSuffix(lowerPath, ".gif"):
		return "image/gif"
	case strings.HasSuffix(lowerPath, ".webp"):
		return "image/webp"
	default:
		return "image/jpeg" // デフォルト
	}
}

// AnalyzeFoodText はテキストから食材を分析し、栄養素を計算する
func (s *FoodService) AnalyzeFoodText(ctx context.Context, inputText string) (*AnalysisResult, error) {
	// Step 1: テキストから食材リストを生成
	foods, err := s.geminiClient.ParseTextToFoods(ctx, inputText)
	if err != nil {
		return nil, fmt.Errorf("テキスト解析エラー: %w", err)
	}

	// Step 2: 栄養素計算（画像分析と共通）
	nutritionList, err := s.geminiClient.CalculateNutrition(ctx, foods)
	if err != nil {
		return nil, fmt.Errorf("栄養素計算エラー: %w", err)
	}

	// 合計カロリー・栄養素を計算
	totalCal, totalPro, totalFat, totalCarbs := calculateTotals(nutritionList)

	return &AnalysisResult{
		Foods:              nutritionList,
		TotalCalories:      totalCal,
		TotalProtein:       totalPro,
		TotalFat:           totalFat,
		TotalCarbohydrates: totalCarbs,
	}, nil
}

// calculateTotals は栄養素の合計を計算する
func calculateTotals(nutritionList []gemini.NutritionInfo) (totalCal, totalPro, totalFat, totalCarbs float64) {
	for _, nutrition := range nutritionList {
		totalCal += nutrition.Calories
		totalPro += nutrition.Protein
		totalFat += nutrition.Fat
		totalCarbs += nutrition.Carbohydrates
	}
	return
}
