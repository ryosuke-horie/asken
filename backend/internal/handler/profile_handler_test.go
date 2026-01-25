package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
)

type MockProfileRepository struct {
	GetByUserIDFunc    func(ctx context.Context, userID uuid.UUID) (*repository.UserProfile, error)
	CreateOrUpdateFunc func(ctx context.Context, profile *repository.UserProfile) (*repository.UserProfile, error)
}

func (m *MockProfileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*repository.UserProfile, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockProfileRepository) CreateOrUpdate(ctx context.Context, profile *repository.UserProfile) (*repository.UserProfile, error) {
	if m.CreateOrUpdateFunc != nil {
		return m.CreateOrUpdateFunc(ctx, profile)
	}
	return nil, nil
}

func createProfileRequestWithUserID(method, url string, body []byte, userID uuid.UUID) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDToContext(req.Context(), userID)
	return req.WithContext(ctx)
}

func TestProfileHandler_HandleGetProfile(t *testing.T) {
	userID := uuid.New()
	profileID := uuid.New()
	sportType := "柔術"
	weightClass := 65

	tests := []struct {
		name           string
		userID         uuid.UUID
		mockProfile    *repository.UserProfile
		mockErr        error
		expectedStatus int
	}{
		{
			name:   "正常にプロフィールを取得できるべき",
			userID: userID,
			mockProfile: &repository.UserProfile{
				ID:            profileID,
				UserID:        userID,
				SportType:     &sportType,
				TrainingGoals: []string{"減量", "スタミナ強化"},
				WeightClass:   &weightClass,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "プロフィールが存在しない場合はnullを返すべき",
			userID:         userID,
			mockProfile:    nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "未認証の場合は401を返すべき",
			userID:         uuid.Nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "リポジトリエラーの場合は500を返すべき",
			userID:         userID,
			mockErr:        errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockProfileRepository{
				GetByUserIDFunc: func(ctx context.Context, uid uuid.UUID) (*repository.UserProfile, error) {
					if tt.mockErr != nil {
						return nil, tt.mockErr
					}
					return tt.mockProfile, nil
				},
			}
			handler := NewProfileHandler(mockRepo)

			req := createProfileRequestWithUserID(http.MethodGet, "/api/profile", nil, tt.userID)
			rec := httptest.NewRecorder()

			handler.HandleGetProfile(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("期待されるステータス %d, 実際のステータス %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestProfileHandler_HandleUpdateProfile(t *testing.T) {
	userID := uuid.New()
	profileID := uuid.New()
	sportType := "柔術"
	weightClass := 65

	tests := []struct {
		name           string
		requestBody    UpdateProfileRequest
		userID         uuid.UUID
		mockFunc       func(ctx context.Context, profile *repository.UserProfile) (*repository.UserProfile, error)
		expectedStatus int
	}{
		{
			name: "正常にプロフィールを更新できるべき",
			requestBody: UpdateProfileRequest{
				SportType:     &sportType,
				TrainingGoals: []string{"減量", "スタミナ強化"},
				WeightClass:   &weightClass,
			},
			userID: userID,
			mockFunc: func(ctx context.Context, profile *repository.UserProfile) (*repository.UserProfile, error) {
				return &repository.UserProfile{
					ID:            profileID,
					UserID:        profile.UserID,
					SportType:     profile.SportType,
					TrainingGoals: profile.TrainingGoals,
					WeightClass:   profile.WeightClass,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "未認証の場合は401を返すべき",
			requestBody: UpdateProfileRequest{
				SportType: &sportType,
			},
			userID:         uuid.Nil,
			mockFunc:       nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "体重階級が範囲外（0）の場合は400を返すべき",
			requestBody: UpdateProfileRequest{
				WeightClass: intPtr(0),
			},
			userID:         userID,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "体重階級が範囲外（201）の場合は400を返すべき",
			requestBody: UpdateProfileRequest{
				WeightClass: intPtr(201),
			},
			userID:         userID,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "リポジトリエラーの場合は500を返すべき",
			requestBody: UpdateProfileRequest{
				SportType: &sportType,
			},
			userID: userID,
			mockFunc: func(ctx context.Context, profile *repository.UserProfile) (*repository.UserProfile, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "無効なSportTypeの場合は400を返すべき",
			requestBody: UpdateProfileRequest{
				SportType: strPtr("無効な競技"),
			},
			userID:         userID,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "無効なTrainingGoalsの場合は400を返すべき",
			requestBody: UpdateProfileRequest{
				TrainingGoals: []string{"減量", "無効な目標"},
			},
			userID:         userID,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "有効なSportTypeとTrainingGoalsの場合は正常に更新できるべき",
			requestBody: UpdateProfileRequest{
				SportType:     strPtr("MMA"),
				TrainingGoals: []string{"減量", "スタミナ強化", "技術向上"},
			},
			userID: userID,
			mockFunc: func(ctx context.Context, profile *repository.UserProfile) (*repository.UserProfile, error) {
				return &repository.UserProfile{
					ID:            profileID,
					UserID:        profile.UserID,
					SportType:     profile.SportType,
					TrainingGoals: profile.TrainingGoals,
					WeightClass:   profile.WeightClass,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "空のSportTypeは許可されるべき",
			requestBody: UpdateProfileRequest{
				SportType: strPtr(""),
			},
			userID: userID,
			mockFunc: func(ctx context.Context, profile *repository.UserProfile) (*repository.UserProfile, error) {
				return &repository.UserProfile{
					ID:            profileID,
					UserID:        profile.UserID,
					SportType:     profile.SportType,
					TrainingGoals: profile.TrainingGoals,
					WeightClass:   profile.WeightClass,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockProfileRepository{
				CreateOrUpdateFunc: tt.mockFunc,
			}
			handler := NewProfileHandler(mockRepo)

			body, _ := json.Marshal(tt.requestBody)
			req := createProfileRequestWithUserID(http.MethodPut, "/api/profile", body, tt.userID)
			rec := httptest.NewRecorder()

			handler.HandleUpdateProfile(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("期待されるステータス %d, 実際のステータス %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestProfileHandler_HandleUpdateProfile_InvalidJSON(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		body           []byte
		userID         uuid.UUID
		expectedStatus int
	}{
		{
			name:           "不正なJSON形式の場合は400を返すべき",
			body:           []byte(`{invalid json`),
			userID:         userID,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "空のボディの場合は400を返すべき",
			body:           []byte(``),
			userID:         userID,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockProfileRepository{}
			handler := NewProfileHandler(mockRepo)

			req := createProfileRequestWithUserID(http.MethodPut, "/api/profile", tt.body, tt.userID)
			rec := httptest.NewRecorder()

			handler.HandleUpdateProfile(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("期待されるステータス %d, 実際のステータス %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestProfileHandler_HandleUpdateProfile_WithNilValues(t *testing.T) {
	userID := uuid.New()
	profileID := uuid.New()

	mockRepo := &MockProfileRepository{
		CreateOrUpdateFunc: func(ctx context.Context, profile *repository.UserProfile) (*repository.UserProfile, error) {
			return &repository.UserProfile{
				ID:            profileID,
				UserID:        profile.UserID,
				SportType:     nil,
				TrainingGoals: nil,
				WeightClass:   nil,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}, nil
		},
	}
	handler := NewProfileHandler(mockRepo)

	requestBody := UpdateProfileRequest{}
	body, _ := json.Marshal(requestBody)
	req := createProfileRequestWithUserID(http.MethodPut, "/api/profile", body, userID)
	rec := httptest.NewRecorder()

	handler.HandleUpdateProfile(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("期待されるステータス %d, 実際のステータス %d", http.StatusOK, rec.Code)
	}
}

func intPtr(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
}
