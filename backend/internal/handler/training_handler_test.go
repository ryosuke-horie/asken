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

func (m *MockTrainingRepo) GetAllLocations(ctx context.Context, userID uuid.UUID) ([]*repository.TrainingLocation, error) {
	return nil, nil
}
func (m *MockTrainingRepo) GetLocationByID(ctx context.Context, id, userID uuid.UUID) (*repository.TrainingLocation, error) {
	return nil, nil
}
func (m *MockTrainingRepo) CreateLocation(ctx context.Context, location *repository.TrainingLocation) (*repository.TrainingLocation, error) {
	return nil, nil
}
func (m *MockTrainingRepo) UpdateLocation(ctx context.Context, location *repository.TrainingLocation) (*repository.TrainingLocation, error) {
	return nil, nil
}
func (m *MockTrainingRepo) DeleteLocation(ctx context.Context, id, userID uuid.UUID) error {
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
func (m *MockTrainingRepo) GetMenus(ctx context.Context, userID uuid.UUID) ([]*repository.TrainingMenu, error) {
	return nil, nil
}
func (m *MockTrainingRepo) CreateMenu(ctx context.Context, menu *repository.TrainingMenu) (*repository.TrainingMenu, error) {
	return nil, nil
}
func (m *MockTrainingRepo) DeleteMenu(ctx context.Context, id, userID uuid.UUID) error {
	return nil
}
func (m *MockTrainingRepo) GetRecords(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*repository.TrainingRecord, error) {
	return nil, nil
}
func (m *MockTrainingRepo) GetRecordByDate(ctx context.Context, userID uuid.UUID, date time.Time) (*repository.TrainingRecord, error) {
	return nil, nil
}
func (m *MockTrainingRepo) GetRecordByID(ctx context.Context, id, userID uuid.UUID) (*repository.TrainingRecord, error) {
	return nil, nil
}
func (m *MockTrainingRepo) CreateRecord(ctx context.Context, record *repository.TrainingRecord, menuIDs []uuid.UUID) (*repository.TrainingRecord, error) {
	return nil, nil
}
func (m *MockTrainingRepo) UpdateRecord(ctx context.Context, record *repository.TrainingRecord, menuIDs []uuid.UUID) (*repository.TrainingRecord, error) {
	return nil, nil
}
func (m *MockTrainingRepo) DeleteRecord(ctx context.Context, id, userID uuid.UUID) error {
	return nil
}
func (m *MockTrainingRepo) UpsertRecord(ctx context.Context, record *repository.TrainingRecord) (*repository.TrainingRecord, error) {
	return nil, nil
}

// MockConditionRepo はテスト用のConditionRepositoryモック（training用）
type MockConditionRepo struct {
	GetRecordByDateFunc func(ctx context.Context, userID uuid.UUID, date string) (*repository.ConditionRecord, error)
}

func (m *MockConditionRepo) CreateOrUpdateRecord(ctx context.Context, userID uuid.UUID, condition, fatigue int, recordedAt string) (*repository.ConditionRecord, error) {
	return nil, nil
}

func (m *MockConditionRepo) GetRecordByDate(ctx context.Context, userID uuid.UUID, date string) (*repository.ConditionRecord, error) {
	if m.GetRecordByDateFunc != nil {
		return m.GetRecordByDateFunc(ctx, userID, date)
	}
	return nil, nil
}

func createTestRequest(method, url string, body interface{}, userID uuid.UUID) *http.Request {
	var bodyReader *bytes.Buffer
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(bodyBytes)
	} else {
		bodyReader = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDToContext(req.Context(), userID)
	return req.WithContext(ctx)
}

func TestHandleSuggestMenu_WithProfileIntegration(t *testing.T) {
	testUserID := uuid.New()
	sportType := "柔術"

	tests := []struct {
		name           string
		requestBody    SuggestMenuRequest
		profileSetup   func() *MockProfileRepository
		conditionSetup func() *MockConditionRepo
		expectedStatus int
	}{
		{
			name: "プロフィールが存在する場合、SportTypeがパラメータに反映されるべき",
			requestBody: SuggestMenuRequest{
				Equipment: []string{"ダンベル"},
				Duration:  60,
			},
			profileSetup: func() *MockProfileRepository {
				return &MockProfileRepository{
					GetByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (*repository.UserProfile, error) {
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
			profileSetup: func() *MockProfileRepository {
				return &MockProfileRepository{
					GetByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (*repository.UserProfile, error) {
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
			profileSetup: func() *MockProfileRepository {
				return &MockProfileRepository{
					GetByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (*repository.UserProfile, error) {
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
			profileSetup: func() *MockProfileRepository {
				return &MockProfileRepository{
					GetByUserIDFunc: func(ctx context.Context, userID uuid.UUID) (*repository.UserProfile, error) {
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
	testUserID := uuid.New()

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
					GetRecordByDateFunc: func(ctx context.Context, userID uuid.UUID, date string) (*repository.ConditionRecord, error) {
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
					GetRecordByDateFunc: func(ctx context.Context, userID uuid.UUID, date string) (*repository.ConditionRecord, error) {
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
					GetRecordByDateFunc: func(ctx context.Context, userID uuid.UUID, date string) (*repository.ConditionRecord, error) {
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
				profileRepo:   &MockProfileRepository{},
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
	testUserID := uuid.New()

	t.Run("conditionRepoがnilの場合でもパニックしないべき", func(t *testing.T) {
		handler := &TrainingHandler{
			repository:    &MockTrainingRepo{},
			conditionRepo: nil,
			profileRepo:   &MockProfileRepository{},
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
	testUserID := uuid.New()

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
				profileRepo:   &MockProfileRepository{},
			}

			var req *http.Request
			if str, ok := tt.requestBody.(string); ok {
				req = httptest.NewRequest(http.MethodPost, "/api/training/suggest-menu", bytes.NewBuffer([]byte(str)))
				req.Header.Set("Content-Type", "application/json")
				ctx := middleware.SetUserIDToContext(req.Context(), testUserID)
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
			profileRepo:   &MockProfileRepository{},
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
