package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
)

func TestValidateRecordDate(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "有効な日付形式",
			input:   "2024-01-15",
			wantErr: false,
		},
		{
			name:        "空の日付",
			input:       "",
			wantErr:     true,
			errContains: "recorded_atは必須です",
		},
		{
			name:        "無効な日付形式",
			input:       "2024/01/15",
			wantErr:     true,
			errContains: "無効な日付形式です（YYYY-MM-DD）",
		},
		{
			name:        "不正な日付",
			input:       "invalid-date",
			wantErr:     true,
			errContains: "無効な日付形式です（YYYY-MM-DD）",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validateRecordDate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("エラーが期待されましたが、nilが返されました")
				} else if tt.errContains != "" && err.Error() != tt.errContains {
					t.Errorf("エラーメッセージが一致しません: got %q, want %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
				if result.IsZero() {
					t.Error("結果がゼロ値です")
				}
			}
		})
	}
}

func TestValidateRecordFields(t *testing.T) {
	intPtr := func(v int) *int { return &v }

	tests := []struct {
		name         string
		intensity    *int
		satisfaction *int
		duration     *int
		wantErr      bool
		errContains  string
	}{
		{
			name:         "全てnil",
			intensity:    nil,
			satisfaction: nil,
			duration:     nil,
			wantErr:      false,
		},
		{
			name:         "有効な値",
			intensity:    intPtr(3),
			satisfaction: intPtr(4),
			duration:     intPtr(60),
			wantErr:      false,
		},
		{
			name:         "強度が範囲外（下限）",
			intensity:    intPtr(0),
			satisfaction: nil,
			duration:     nil,
			wantErr:      true,
			errContains:  "強度は1-5の範囲で指定してください",
		},
		{
			name:         "強度が範囲外（上限）",
			intensity:    intPtr(6),
			satisfaction: nil,
			duration:     nil,
			wantErr:      true,
			errContains:  "強度は1-5の範囲で指定してください",
		},
		{
			name:         "満足度が範囲外（下限）",
			intensity:    nil,
			satisfaction: intPtr(0),
			duration:     nil,
			wantErr:      true,
			errContains:  "満足度は1-5の範囲で指定してください",
		},
		{
			name:         "満足度が範囲外（上限）",
			intensity:    nil,
			satisfaction: intPtr(6),
			duration:     nil,
			wantErr:      true,
			errContains:  "満足度は1-5の範囲で指定してください",
		},
		{
			name:         "練習時間が負の値",
			intensity:    nil,
			satisfaction: nil,
			duration:     intPtr(-1),
			wantErr:      true,
			errContains:  "練習時間は0以上の値を指定してください",
		},
		{
			name:         "練習時間が0は有効",
			intensity:    nil,
			satisfaction: nil,
			duration:     intPtr(0),
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRecordFields(tt.intensity, tt.satisfaction, tt.duration)
			if tt.wantErr {
				if err == nil {
					t.Error("エラーが期待されましたが、nilが返されました")
				} else if tt.errContains != "" && err.Error() != tt.errContains {
					t.Errorf("エラーメッセージが一致しません: got %q, want %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
			}
		})
	}
}

func TestParseLocationID(t *testing.T) {
	strPtr := func(s string) *string { return &s }
	validUUID := uuid.New().String()

	tests := []struct {
		name        string
		input       *string
		wantNil     bool
		wantErr     bool
		errContains string
	}{
		{
			name:    "nil入力",
			input:   nil,
			wantNil: true,
			wantErr: false,
		},
		{
			name:    "有効なUUID",
			input:   strPtr(validUUID),
			wantNil: false,
			wantErr: false,
		},
		{
			name:        "無効なUUID形式",
			input:       strPtr("invalid-uuid"),
			wantNil:     true,
			wantErr:     true,
			errContains: "無効なlocation_id形式です",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseLocationID(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("エラーが期待されましたが、nilが返されました")
				} else if tt.errContains != "" && err.Error() != tt.errContains {
					t.Errorf("エラーメッセージが一致しません: got %q, want %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
				if tt.wantNil && result != nil {
					t.Error("nilが期待されましたが、値が返されました")
				}
				if !tt.wantNil && result == nil {
					t.Error("値が期待されましたが、nilが返されました")
				}
			}
		})
	}
}

func TestParseMenuIDs(t *testing.T) {
	validUUID1 := uuid.New().String()
	validUUID2 := uuid.New().String()

	tests := []struct {
		name        string
		input       []string
		wantLen     int
		wantErr     bool
		errContains string
	}{
		{
			name:    "空のスライス",
			input:   []string{},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "有効なUUID複数",
			input:   []string{validUUID1, validUUID2},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:        "無効なUUID含む",
			input:       []string{validUUID1, "invalid-uuid"},
			wantLen:     0,
			wantErr:     true,
			errContains: "無効なmenu_id形式です",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseMenuIDs(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("エラーが期待されましたが、nilが返されました")
				} else if tt.errContains != "" && err.Error() != tt.errContains {
					t.Errorf("エラーメッセージが一致しません: got %q, want %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
				if len(result) != tt.wantLen {
					t.Errorf("結果の長さが一致しません: got %d, want %d", len(result), tt.wantLen)
				}
			}
		})
	}
}

// ConfigurableMockTrainingRepo はテスト用の設定可能なモック
type ConfigurableMockTrainingRepo struct {
	MockTrainingRepo
	GetLocationByIDFunc func(ctx context.Context, id, userID uuid.UUID) (*repository.TrainingLocation, error)
}

func (m *ConfigurableMockTrainingRepo) GetLocationByID(ctx context.Context, id, userID uuid.UUID) (*repository.TrainingLocation, error) {
	if m.GetLocationByIDFunc != nil {
		return m.GetLocationByIDFunc(ctx, id, userID)
	}
	return nil, nil
}

func TestValidateAndParseLocationID(t *testing.T) {
	strPtr := func(s string) *string { return &s }
	validUUID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name        string
		locationID  *string
		mockSetup   func() *ConfigurableMockTrainingRepo
		wantNil     bool
		wantErr     bool
		errContains string
	}{
		{
			name:       "nil入力",
			locationID: nil,
			mockSetup: func() *ConfigurableMockTrainingRepo {
				return &ConfigurableMockTrainingRepo{}
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name:       "有効なlocation_id、場所が存在",
			locationID: strPtr(validUUID.String()),
			mockSetup: func() *ConfigurableMockTrainingRepo {
				return &ConfigurableMockTrainingRepo{
					GetLocationByIDFunc: func(ctx context.Context, id, userID uuid.UUID) (*repository.TrainingLocation, error) {
						return &repository.TrainingLocation{
							ID:     id,
							UserID: userID,
							Name:   "テスト場所",
						}, nil
					},
				}
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name:       "有効なlocation_id、場所が存在しない",
			locationID: strPtr(validUUID.String()),
			mockSetup: func() *ConfigurableMockTrainingRepo {
				return &ConfigurableMockTrainingRepo{
					GetLocationByIDFunc: func(ctx context.Context, id, userID uuid.UUID) (*repository.TrainingLocation, error) {
						return nil, nil
					},
				}
			},
			wantNil:     true,
			wantErr:     true,
			errContains: "指定された場所が見つかりません",
		},
		{
			name:       "有効なlocation_id、DB取得エラー",
			locationID: strPtr(validUUID.String()),
			mockSetup: func() *ConfigurableMockTrainingRepo {
				return &ConfigurableMockTrainingRepo{
					GetLocationByIDFunc: func(ctx context.Context, id, userID uuid.UUID) (*repository.TrainingLocation, error) {
						return nil, errors.New("database connection error")
					},
				}
			},
			wantNil:     true,
			wantErr:     true,
			errContains: "場所の取得に失敗しました",
		},
		{
			name:       "無効なlocation_id形式",
			locationID: strPtr("invalid-uuid"),
			mockSetup: func() *ConfigurableMockTrainingRepo {
				return &ConfigurableMockTrainingRepo{}
			},
			wantNil:     true,
			wantErr:     true,
			errContains: "無効なlocation_id形式です",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := tt.mockSetup()
			handler := &TrainingHandler{
				repository: mockRepo,
			}

			result, err := handler.validateAndParseLocationID(context.Background(), userID, tt.locationID)
			if tt.wantErr {
				if err == nil {
					t.Error("エラーが期待されましたが、nilが返されました")
				} else if tt.errContains != "" && err.Error() != tt.errContains {
					t.Errorf("エラーメッセージが一致しません: got %q, want %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
				if tt.wantNil && result != nil {
					t.Error("nilが期待されましたが、値が返されました")
				}
				if !tt.wantNil && result == nil {
					t.Error("値が期待されましたが、nilが返されました")
				}
			}
		})
	}
}

func TestExtractAndValidateRecordID(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name        string
		urlPath     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "有効なURL",
			urlPath: "/api/training/records/" + validUUID.String(),
			wantErr: false,
		},
		{
			name:        "IDがない",
			urlPath:     "/api/training/records/",
			wantErr:     true,
			errContains: "IDが指定されていません",
		},
		{
			name:        "無効なID形式",
			urlPath:     "/api/training/records/invalid-uuid",
			wantErr:     true,
			errContains: "無効なID形式です",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractAndValidateRecordID(tt.urlPath)
			if tt.wantErr {
				if err == nil {
					t.Error("エラーが期待されましたが、nilが返されました")
				} else if tt.errContains != "" && err.Error() != tt.errContains {
					t.Errorf("エラーメッセージが一致しません: got %q, want %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
				if result == uuid.Nil {
					t.Error("有効なUUIDが期待されましたが、Nilが返されました")
				}
			}
		})
	}
}
