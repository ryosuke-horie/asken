package service

import (
	"context"
	"testing"

	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockGeminiClient はテスト用のモックGeminiクライアント
type MockGeminiClient struct {
	ClassifyFoodsFunc         func(ctx context.Context, imagePath string) ([]gemini.FoodItem, error)
	ClassifyFoodsFromDataFunc func(ctx context.Context, imageData []byte, mimeType string) ([]gemini.FoodItem, error)
	ParseTextToFoodsFunc      func(ctx context.Context, inputText string) ([]gemini.FoodItem, error)
	CalculateNutritionFunc    func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error)
}

func (m *MockGeminiClient) ClassifyFoods(ctx context.Context, imagePath string) ([]gemini.FoodItem, error) {
	if m.ClassifyFoodsFunc != nil {
		return m.ClassifyFoodsFunc(ctx, imagePath)
	}
	return nil, nil
}

func (m *MockGeminiClient) ClassifyFoodsFromData(ctx context.Context, imageData []byte, mimeType string) ([]gemini.FoodItem, error) {
	if m.ClassifyFoodsFromDataFunc != nil {
		return m.ClassifyFoodsFromDataFunc(ctx, imageData, mimeType)
	}
	return nil, nil
}

func (m *MockGeminiClient) ParseTextToFoods(ctx context.Context, inputText string) ([]gemini.FoodItem, error) {
	if m.ParseTextToFoodsFunc != nil {
		return m.ParseTextToFoodsFunc(ctx, inputText)
	}
	return nil, nil
}

func (m *MockGeminiClient) CalculateNutrition(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
	if m.CalculateNutritionFunc != nil {
		return m.CalculateNutritionFunc(ctx, foods)
	}
	return nil, nil
}

// MockImageDownloader はテスト用のImageDownloader実装
type MockImageDownloader struct {
	DownloadFunc func(ctx context.Context, objectName string) ([]byte, error)
}

func (m *MockImageDownloader) Download(ctx context.Context, objectName string) ([]byte, error) {
	if m.DownloadFunc != nil {
		return m.DownloadFunc(ctx, objectName)
	}
	return nil, nil
}

func TestAnalyzeFoodImage_Success(t *testing.T) {
	mockClient := &MockGeminiClient{
		ClassifyFoodsFunc: func(ctx context.Context, imagePath string) ([]gemini.FoodItem, error) {
			return []gemini.FoodItem{
				{Name: "刺身盛り合わせ", EstimatedAmount: "8切れ"},
				{Name: "サラダ", EstimatedAmount: "1皿"},
			}, nil
		},
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			return []gemini.NutritionInfo{
				{
					Name:            "刺身盛り合わせ",
					EstimatedAmount: "8切れ",
					Calories:        360.0,
					Protein:         30.0,
					Fat:             24.6,
					Carbohydrates:   0.4,
				},
				{
					Name:            "サラダ",
					EstimatedAmount: "1皿",
					Calories:        50.0,
					Protein:         2.0,
					Fat:             1.5,
					Carbohydrates:   8.0,
				},
			}, nil
		},
	}

	service := NewFoodService(mockClient, nil)
	ctx := context.Background()

	result, err := service.AnalyzeFoodImage(ctx, "/path/to/image.jpg")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Foods, 2)
	assert.Equal(t, 410.0, result.TotalCalories)    // 360 + 50
	assert.Equal(t, 32.0, result.TotalProtein)      // 30 + 2
	assert.Equal(t, 26.1, result.TotalFat)          // 24.6 + 1.5
	assert.Equal(t, 8.4, result.TotalCarbohydrates) // 0.4 + 8
}

func TestAnalyzeFoodImage_Step1Error(t *testing.T) {
	mockClient := &MockGeminiClient{
		ClassifyFoodsFunc: func(ctx context.Context, imagePath string) ([]gemini.FoodItem, error) {
			return nil, assert.AnError
		},
	}

	service := NewFoodService(mockClient, nil)
	ctx := context.Background()

	_, err := service.AnalyzeFoodImage(ctx, "/path/to/image.jpg")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "食材分類エラー")
}

func TestAnalyzeFoodImage_Step2Error(t *testing.T) {
	mockClient := &MockGeminiClient{
		ClassifyFoodsFunc: func(ctx context.Context, imagePath string) ([]gemini.FoodItem, error) {
			return []gemini.FoodItem{
				{Name: "刺身盛り合わせ", EstimatedAmount: "8切れ"},
			}, nil
		},
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			return nil, assert.AnError
		},
	}

	service := NewFoodService(mockClient, nil)
	ctx := context.Background()

	_, err := service.AnalyzeFoodImage(ctx, "/path/to/image.jpg")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "栄養素計算エラー")
}

func TestCalculateTotals(t *testing.T) {
	nutritionList := []gemini.NutritionInfo{
		{
			Name:          "刺身盛り合わせ",
			Calories:      360.0,
			Protein:       30.0,
			Fat:           24.6,
			Carbohydrates: 0.4,
		},
		{
			Name:          "サラダ",
			Calories:      50.0,
			Protein:       2.0,
			Fat:           1.5,
			Carbohydrates: 8.0,
		},
	}

	totalCal, totalPro, totalFat, totalCarbs, totalMicro := calculateTotals(nutritionList)

	assert.Equal(t, 410.0, totalCal)
	assert.Equal(t, 32.0, totalPro)
	assert.Equal(t, 26.1, totalFat)
	assert.Equal(t, 8.4, totalCarbs)
	assert.NotNil(t, totalMicro)
}

func TestAnalyzeFoodText_Success(t *testing.T) {
	mockClient := &MockGeminiClient{
		ParseTextToFoodsFunc: func(ctx context.Context, inputText string) ([]gemini.FoodItem, error) {
			return []gemini.FoodItem{
				{Name: "白米", EstimatedAmount: "300g"},
				{Name: "焼肉", EstimatedAmount: "100g"},
			}, nil
		},
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			return []gemini.NutritionInfo{
				{
					Name:            "白米",
					EstimatedAmount: "300g",
					Calories:        504.0,
					Protein:         7.6,
					Fat:             1.0,
					Carbohydrates:   111.4,
				},
				{
					Name:            "焼肉",
					EstimatedAmount: "100g",
					Calories:        371.0,
					Protein:         17.1,
					Fat:             32.9,
					Carbohydrates:   0.1,
				},
			}, nil
		},
	}

	service := NewFoodService(mockClient, nil)
	ctx := context.Background()

	result, err := service.AnalyzeFoodText(ctx, "ご飯二杯, 焼肉")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Foods, 2)
	assert.InDelta(t, 875.0, result.TotalCalories, 0.01)      // 504 + 371
	assert.InDelta(t, 24.7, result.TotalProtein, 0.01)        // 7.6 + 17.1
	assert.InDelta(t, 33.9, result.TotalFat, 0.01)            // 1.0 + 32.9
	assert.InDelta(t, 111.5, result.TotalCarbohydrates, 0.01) // 111.4 + 0.1
}

func TestAnalyzeFoodText_ParseError(t *testing.T) {
	mockClient := &MockGeminiClient{
		ParseTextToFoodsFunc: func(ctx context.Context, inputText string) ([]gemini.FoodItem, error) {
			return nil, assert.AnError
		},
	}

	service := NewFoodService(mockClient, nil)
	ctx := context.Background()

	_, err := service.AnalyzeFoodText(ctx, "ご飯二杯")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "テキスト解析エラー")
}

func TestAnalyzeFoodText_NutritionError(t *testing.T) {
	mockClient := &MockGeminiClient{
		ParseTextToFoodsFunc: func(ctx context.Context, inputText string) ([]gemini.FoodItem, error) {
			return []gemini.FoodItem{
				{Name: "白米", EstimatedAmount: "300g"},
			}, nil
		},
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			return nil, assert.AnError
		},
	}

	service := NewFoodService(mockClient, nil)
	ctx := context.Background()

	_, err := service.AnalyzeFoodText(ctx, "ご飯二杯")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "栄養素計算エラー")
}

func TestDetectMimeTypeFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"jpg拡張子", "uploads/image.jpg", "image/jpeg"},
		{"jpeg拡張子", "uploads/image.jpeg", "image/jpeg"},
		{"JPG大文字", "uploads/image.JPG", "image/jpeg"},
		{"JPEG大文字", "uploads/image.JPEG", "image/jpeg"},
		{"png拡張子", "uploads/image.png", "image/png"},
		{"PNG大文字", "uploads/image.PNG", "image/png"},
		{"gif拡張子", "uploads/image.gif", "image/gif"},
		{"webp拡張子", "uploads/image.webp", "image/webp"},
		{"bmpはデフォルト", "uploads/image.bmp", "image/jpeg"},
		{"拡張子なしはデフォルト", "uploads/image", "image/jpeg"},
		{"空文字はデフォルト", "", "image/jpeg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectMimeTypeFromPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalyzeFoodImage_CloudStoragePath_Success(t *testing.T) {
	imageData := []byte("fake image data")

	mockDownloader := &MockImageDownloader{
		DownloadFunc: func(ctx context.Context, objectName string) ([]byte, error) {
			assert.Equal(t, "uploads/test-uuid.jpg", objectName)
			return imageData, nil
		},
	}

	var capturedImageData []byte
	var capturedMimeType string
	mockClient := &MockGeminiClient{
		ClassifyFoodsFromDataFunc: func(ctx context.Context, data []byte, mimeType string) ([]gemini.FoodItem, error) {
			capturedImageData = data
			capturedMimeType = mimeType
			return []gemini.FoodItem{
				{Name: "カレーライス", EstimatedAmount: "1皿"},
			}, nil
		},
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			return []gemini.NutritionInfo{
				{
					Name:            "カレーライス",
					EstimatedAmount: "1皿",
					Calories:        600.0,
					Protein:         15.0,
					Fat:             20.0,
					Carbohydrates:   80.0,
				},
			}, nil
		},
	}

	svc := NewFoodService(mockClient, mockDownloader)
	result, err := svc.AnalyzeFoodImage(context.Background(), "uploads/test-uuid.jpg")

	require.NoError(t, err)
	assert.Equal(t, imageData, capturedImageData)
	assert.Equal(t, "image/jpeg", capturedMimeType)
	assert.Len(t, result.Foods, 1)
	assert.Equal(t, 600.0, result.TotalCalories)
}

func TestAnalyzeFoodImage_CloudStoragePath_DownloadError(t *testing.T) {
	mockDownloader := &MockImageDownloader{
		DownloadFunc: func(ctx context.Context, objectName string) ([]byte, error) {
			return nil, assert.AnError
		},
	}
	mockClient := &MockGeminiClient{}

	svc := NewFoodService(mockClient, mockDownloader)
	_, err := svc.AnalyzeFoodImage(context.Background(), "uploads/test-uuid.jpg")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "画像ダウンロードエラー")
}

func TestAnalyzeFoodImage_CloudStoragePath_ClassifyError(t *testing.T) {
	mockDownloader := &MockImageDownloader{
		DownloadFunc: func(ctx context.Context, objectName string) ([]byte, error) {
			return []byte("image"), nil
		},
	}
	mockClient := &MockGeminiClient{
		ClassifyFoodsFromDataFunc: func(ctx context.Context, data []byte, mimeType string) ([]gemini.FoodItem, error) {
			return nil, assert.AnError
		},
	}

	svc := NewFoodService(mockClient, mockDownloader)
	_, err := svc.AnalyzeFoodImage(context.Background(), "uploads/test-uuid.png")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "食材分類エラー")
}

func TestAnalyzeFoodImage_CloudStoragePath_NilDownloader(t *testing.T) {
	mockClient := &MockGeminiClient{
		ClassifyFoodsFunc: func(ctx context.Context, imagePath string) ([]gemini.FoodItem, error) {
			assert.Equal(t, "uploads/test-uuid.jpg", imagePath)
			return []gemini.FoodItem{
				{Name: "白米", EstimatedAmount: "1杯"},
			}, nil
		},
		CalculateNutritionFunc: func(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error) {
			return []gemini.NutritionInfo{
				{Name: "白米", Calories: 252.0, Protein: 3.8, Fat: 0.5, Carbohydrates: 55.7},
			}, nil
		},
	}

	svc := NewFoodService(mockClient, nil)
	result, err := svc.AnalyzeFoodImage(context.Background(), "uploads/test-uuid.jpg")

	require.NoError(t, err)
	assert.Len(t, result.Foods, 1)
}
