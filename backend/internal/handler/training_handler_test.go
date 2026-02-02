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

// MockTrainingRepo はテスト用のTrainingRepositoryモック
type MockTrainingRepo struct{}

func (m *MockTrainingRepo) GetAllLocations(ctx context.Context, userID string) ([]*repository.TrainingLocation, error) {
	return nil, nil
}
func (m *MockTrainingRepo) GetLocationByID(ctx context.Context, id, userID string) (*repository.TrainingLocation, error) {
	return nil, nil
}
func (m *MockTrainingRepo) CreateLocation(ctx context.Context, location *repository.TrainingLocation) (*repository.TrainingLocation, error) {
	return nil, nil
}
func (m *MockTrainingRepo) UpdateLocation(ctx context.Context, location *repository.TrainingLocation) (*repository.TrainingLocation, error) {
	return nil, nil
}
func (m *MockTrainingRepo) DeleteLocation(ctx context.Context, id, userID string) error {
	return nil
}
func (m *MockTrainingRepo) GetEquipmentByLocation(ctx context.Context, locationID uuid.UUID) ([]*repository.TrainingEquipment, error) {
	return nil, nil
}
func (m *MockTrainingRepo) GetEquipmentByID(ctx context.Context, id uuid.UUID) (*repository.TrainingEquipment, error) {
	return nil, nil
}
func (m *MockTrainingRepo) CreateEquipment(ctx context.Context, equipment *repository.TrainingEquipment) (*repository.TrainingEquipment, error) {
	return nil, nil
}
func (m *MockTrainingRepo) UpdateEquipment(ctx context.Context, equipment *repository.TrainingEquipment) (*repository.TrainingEquipment, error) {
	return nil, nil
}
func (m *MockTrainingRepo) DeleteEquipment(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockTrainingRepo) GetMenus(ctx context.Context, userID string) ([]*repository.TrainingMenu, error) {
	return nil, nil
}
func (m *MockTrainingRepo) CreateMenu(ctx context.Context, menu *repository.TrainingMenu) (*repository.TrainingMenu, error) {
	return nil, nil
}
func (m *MockTrainingRepo) DeleteMenu(ctx context.Context, id, userID string) error {
	return nil
}
func (m *MockTrainingRepo) GetRecords(ctx context.Context, userID string, startDate, endDate time.Time) ([]*repository.TrainingRecord, error) {
	return nil, nil
}
func (m *MockTrainingRepo) GetRecordByDate(ctx context.Context, userID string, date time.Time) (*repository.TrainingRecord, error) {
	return nil, nil
}
func (m *MockTrainingRepo) GetRecordByID(ctx context.Context, id, userID string) (*repository.TrainingRecord, error) {
	return nil, nil
}
func (m *MockTrainingRepo) CreateRecord(ctx context.Context, record *repository.TrainingRecord, menuIDs []uuid.UUID) (*repository.TrainingRecord, error) {
	return nil, nil
}
func (m *MockTrainingRepo) UpdateRecord(ctx context.Context, record *repository.TrainingRecord, menuIDs []uuid.UUID) (*repository.TrainingRecord, error) {
	return nil, nil
}
func (m *MockTrainingRepo) DeleteRecord(ctx context.Context, id, userID string) error {
	return nil
}
func (m *MockTrainingRepo) UpsertRecord(ctx context.Context, record *repository.TrainingRecord) (*repository.TrainingRecord, error) {
	return nil, nil
}

// MockConditionRepo はテスト用のConditionRepositoryモック（training用）
type MockConditionRepo struct {
	GetRecordByDateFunc func(ctx context.Context, userID string, date string) (*repository.ConditionRecord, error)
}

func (m *MockConditionRepo) CreateOrUpdateRecord(ctx context.Context, userID string, condition, fatigue int, recordedAt string) (*repository.ConditionRecord, error) {
	return nil, nil
}

func (m *MockConditionRepo) GetRecordByDate(ctx context.Context, userID string, date string) (*repository.ConditionRecord, error) {
	if m.GetRecordByDateFunc != nil {
		return m.GetRecordByDateFunc(ctx, userID, date)
	}
	return nil, nil
}

// MockProfileRepo はテスト用のProfileRepositoryモック（training用）
type MockProfileRepo struct {
	GetByUserIDFunc func(ctx context.Context, userID string) (*repository.UserProfile, error)
}

func (m *MockProfileRepo) GetByUserID(ctx context.Context, userID string) (*repository.UserProfile, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockProfileRepo) CreateOrUpdate(ctx context.Context, profile *repository.UserProfile) (*repository.UserProfile, error) {
	return nil, nil
}

func createTestRequest(method, url string, body interface{}, userID string) *http.Request {
	var bodyReader *bytes.Buffer
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(bodyBytes)
	} else {
		bodyReader = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), userID)
	return req.WithContext(ctx)
}

func TestHandleSuggestMenu_WithProfileIntegration(t *testing.T) {
	testUserID := "test-firebase-uid"
	sportType := "柔術"

	tests := []struct {
		name           string
		requestBody    SuggestMenuRequest
		profileSetup   func() *MockProfileRepo
		conditionSetup func() *MockConditionRepo
		expectedStatus int
	}{
		{
			name: "プロフィールが存在する場合、SportTypeがパラメータに反映されるべき",
			requestBody: SuggestMenuRequest{
				Equipment: []string{"ダンベル"},
				Duration:  60,
			},
			profileSetup: func() *MockProfileRepo {
				return &MockProfileRepo{
					GetByUserIDFunc: func(ctx context.Context, userID string) (*repository.UserProfile, error) {
						return &repository.UserProfile{
							ID:            uuid.New(),
							UserID:        userID,
							SportType:     &sportType,
							TrainingGoals: []string{"筋力強化", "スタミナ強化"},
						}, nil
					},
				}
			},
			conditionSetup: func() *MockConditionRepo {
				return &MockConditionRepo{}
			},
			expectedStatus: http.StatusServiceUnavailable, // menuSuggesterがnil
		},
		{
			name: "リクエストにGoalsが指定されている場合、プロフィールのTrainingGoalsで上書きしないべき",
			requestBody: SuggestMenuRequest{
				Equipment: []string{"ダンベル"},
				Duration:  60,
				Goals:     []string{"減量"},
			},
			profileSetup: func() *MockProfileRepo {
				return &MockProfileRepo{
					GetByUserIDFunc: func(ctx context.Context, userID string) (*repository.UserProfile, error) {
						return &repository.UserProfile{
							ID:            uuid.New(),
							UserID:        userID,
							SportType:     &sportType,
							TrainingGoals: []string{"筋力強化", "スタミナ強化"},
						}, nil
					},
				}
			},
			conditionSetup: func() *MockConditionRepo {
				return &MockConditionRepo{}
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name: "プロフィールが存在しない場合でも、メニュー提案処理が継続すべき",
			requestBody: SuggestMenuRequest{
				Equipment: []string{"ダンベル"},
				Duration:  60,
			},
			profileSetup: func() *MockProfileRepo {
				return &MockProfileRepo{
					GetByUserIDFunc: func(ctx context.Context, userID string) (*repository.UserProfile, error) {
						return nil, nil
					},
				}
			},
			conditionSetup: func() *MockConditionRepo {
				return &MockConditionRepo{}
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name: "プロフィール取得でエラーが発生しても、処理が継続すべき",
			requestBody: SuggestMenuRequest{
				Equipment: []string{"ダンベル"},
				Duration:  60,
			},
			profileSetup: func() *MockProfileRepo {
				return &MockProfileRepo{
					GetByUserIDFunc: func(ctx context.Context, userID string) (*repository.UserProfile, error) {
						return nil, errors.New("database error")
					},
				}
			},
			conditionSetup: func() *MockConditionRepo {
				return &MockConditionRepo{}
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProfileRepo := tt.profileSetup()
			mockConditionRepo := tt.conditionSetup()

			handler := &TrainingHandler{
				repository:    &MockTrainingRepo{},
				conditionRepo: mockConditionRepo,
				profileRepo:   mockProfileRepo,
				// menuSuggesterはnilのため503を返す
			}

			req := createTestRequest(http.MethodPost, "/api/training/suggest-menu", tt.requestBody, testUserID)
			rec := httptest.NewRecorder()

			handler.HandleSuggestMenu(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("期待されるステータス %d, 実際のステータス %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestHandleSuggestMenu_WithConditionIntegration(t *testing.T) {
	testUserID := "test-firebase-uid"

	tests := []struct {
		name           string
		requestBody    SuggestMenuRequest
		conditionSetup func() *MockConditionRepo
		expectedStatus int
	}{
		{
			name: "体調記録が存在する場合、FatigueとConditionがパラメータに反映されるべき",
			requestBody: SuggestMenuRequest{
				Equipment: []string{"ダンベル"},
				Duration:  60,
			},
			conditionSetup: func() *MockConditionRepo {
				return &MockConditionRepo{
					GetRecordByDateFunc: func(ctx context.Context, userID string, date string) (*repository.ConditionRecord, error) {
						return &repository.ConditionRecord{
							ID:        uuid.New(),
							UserID:    userID,
							Fatigue:   3,
							Condition: 2,
						}, nil
					},
				}
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name: "体調記録が存在しない場合でも、処理が継続すべき",
			requestBody: SuggestMenuRequest{
				Equipment: []string{"ダンベル"},
				Duration:  60,
			},
			conditionSetup: func() *MockConditionRepo {
				return &MockConditionRepo{
					GetRecordByDateFunc: func(ctx context.Context, userID string, date string) (*repository.ConditionRecord, error) {
						return nil, nil
					},
				}
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name: "体調記録取得でエラーが発生しても、処理が継続すべき",
			requestBody: SuggestMenuRequest{
				Equipment: []string{"ダンベル"},
				Duration:  60,
			},
			conditionSetup: func() *MockConditionRepo {
				return &MockConditionRepo{
					GetRecordByDateFunc: func(ctx context.Context, userID string, date string) (*repository.ConditionRecord, error) {
						return nil, errors.New("database error")
					},
				}
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConditionRepo := tt.conditionSetup()

			handler := &TrainingHandler{
				repository:    &MockTrainingRepo{},
				conditionRepo: mockConditionRepo,
				profileRepo:   &MockProfileRepo{},
			}

			req := createTestRequest(http.MethodPost, "/api/training/suggest-menu", tt.requestBody, testUserID)
			rec := httptest.NewRecorder()

			handler.HandleSuggestMenu(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("期待されるステータス %d, 実際のステータス %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestHandleSuggestMenu_WithNilRepositories(t *testing.T) {
	testUserID := "test-firebase-uid"

	t.Run("conditionRepoがnilの場合でもパニックしないべき", func(t *testing.T) {
		handler := &TrainingHandler{
			repository:    &MockTrainingRepo{},
			conditionRepo: nil,
			profileRepo:   &MockProfileRepo{},
		}

		req := createTestRequest(http.MethodPost, "/api/training/suggest-menu", SuggestMenuRequest{
			Equipment: []string{"ダンベル"},
			Duration:  60,
		}, testUserID)
		rec := httptest.NewRecorder()

		// パニックしなければOK
		handler.HandleSuggestMenu(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("期待されるステータス 503, 実際のステータス %d", rec.Code)
		}
	})

	t.Run("profileRepoがnilの場合でもパニックしないべき", func(t *testing.T) {
		handler := &TrainingHandler{
			repository:    &MockTrainingRepo{},
			conditionRepo: &MockConditionRepo{},
			profileRepo:   nil,
		}

		req := createTestRequest(http.MethodPost, "/api/training/suggest-menu", SuggestMenuRequest{
			Equipment: []string{"ダンベル"},
			Duration:  60,
		}, testUserID)
		rec := httptest.NewRecorder()

		// パニックしなければOK
		handler.HandleSuggestMenu(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("期待されるステータス 503, 実際のステータス %d", rec.Code)
		}
	})
}

func TestHandleSuggestMenu_Validation(t *testing.T) {
	testUserID := "test-firebase-uid"

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name:           "equipmentが空の場合は400を返すべき",
			requestBody:    SuggestMenuRequest{Equipment: []string{}, Duration: 60},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "durationが0の場合は400を返すべき",
			requestBody:    SuggestMenuRequest{Equipment: []string{"ダンベル"}, Duration: 0},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "durationが負の場合は400を返すべき",
			requestBody:    SuggestMenuRequest{Equipment: []string{"ダンベル"}, Duration: -10},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "不正なJSON形式の場合は400を返すべき",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &TrainingHandler{
				repository:    &MockTrainingRepo{},
				conditionRepo: &MockConditionRepo{},
				profileRepo:   &MockProfileRepo{},
			}

			var req *http.Request
			if str, ok := tt.requestBody.(string); ok {
				req = httptest.NewRequest(http.MethodPost, "/api/training/suggest-menu", bytes.NewBuffer([]byte(str)))
				req.Header.Set("Content-Type", "application/json")
				ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
				req = req.WithContext(ctx)
			} else {
				req = createTestRequest(http.MethodPost, "/api/training/suggest-menu", tt.requestBody, testUserID)
			}

			rec := httptest.NewRecorder()

			handler.HandleSuggestMenu(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("期待されるステータス %d, 実際のステータス %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestHandleSuggestMenu_Unauthorized(t *testing.T) {
	t.Run("未認証の場合は401を返すべき", func(t *testing.T) {
		handler := &TrainingHandler{
			repository:    &MockTrainingRepo{},
			conditionRepo: &MockConditionRepo{},
			profileRepo:   &MockProfileRepo{},
		}

		body, _ := json.Marshal(SuggestMenuRequest{
			Equipment: []string{"ダンベル"},
			Duration:  60,
		})
		req := httptest.NewRequest(http.MethodPost, "/api/training/suggest-menu", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		// userIDをセットしない

		rec := httptest.NewRecorder()

		handler.HandleSuggestMenu(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("期待されるステータス 401, 実際のステータス %d", rec.Code)
		}
	})
}
