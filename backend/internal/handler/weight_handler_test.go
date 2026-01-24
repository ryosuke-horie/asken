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

// MockWeightRepository はテスト用のモックリポジトリ
type MockWeightRepository struct {
	CreateOrUpdateRecordFunc func(ctx context.Context, userID uuid.UUID, weight float64, recordedAt string) (*repository.WeightRecord, error)
	GetRecordsByPeriodFunc   func(ctx context.Context, userID uuid.UUID, startDate, endDate string) ([]*repository.WeightRecord, error)
	GetLatestRecordFunc      func(ctx context.Context, userID uuid.UUID) (*repository.WeightRecord, error)
	GetStatsByPeriodFunc     func(ctx context.Context, userID uuid.UUID, startDate, endDate string) (*repository.WeightStats, error)
	GetGoalFunc              func(ctx context.Context, userID uuid.UUID) (*repository.WeightGoal, error)
	CreateOrUpdateGoalFunc   func(ctx context.Context, userID uuid.UUID, targetWeight float64, targetDate string) (*repository.WeightGoal, error)
}

func (m *MockWeightRepository) CreateOrUpdateRecord(ctx context.Context, userID uuid.UUID, weight float64, recordedAt string) (*repository.WeightRecord, error) {
	if m.CreateOrUpdateRecordFunc != nil {
		return m.CreateOrUpdateRecordFunc(ctx, userID, weight, recordedAt)
	}
	return nil, nil
}

func (m *MockWeightRepository) GetRecordsByPeriod(ctx context.Context, userID uuid.UUID, startDate, endDate string) ([]*repository.WeightRecord, error) {
	if m.GetRecordsByPeriodFunc != nil {
		return m.GetRecordsByPeriodFunc(ctx, userID, startDate, endDate)
	}
	return nil, nil
}

func (m *MockWeightRepository) GetLatestRecord(ctx context.Context, userID uuid.UUID) (*repository.WeightRecord, error) {
	if m.GetLatestRecordFunc != nil {
		return m.GetLatestRecordFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockWeightRepository) GetStatsByPeriod(ctx context.Context, userID uuid.UUID, startDate, endDate string) (*repository.WeightStats, error) {
	if m.GetStatsByPeriodFunc != nil {
		return m.GetStatsByPeriodFunc(ctx, userID, startDate, endDate)
	}
	return nil, nil
}

func (m *MockWeightRepository) GetGoal(ctx context.Context, userID uuid.UUID) (*repository.WeightGoal, error) {
	if m.GetGoalFunc != nil {
		return m.GetGoalFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockWeightRepository) CreateOrUpdateGoal(ctx context.Context, userID uuid.UUID, targetWeight float64, targetDate string) (*repository.WeightGoal, error) {
	if m.CreateOrUpdateGoalFunc != nil {
		return m.CreateOrUpdateGoalFunc(ctx, userID, targetWeight, targetDate)
	}
	return nil, nil
}

func createRequestWithUserID(method, url string, body []byte, userID uuid.UUID) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDToContext(req.Context(), userID)
	return req.WithContext(ctx)
}

func TestHandleCreateRecord(t *testing.T) {
	userID := uuid.New()
	recordID := uuid.New()

	tests := []struct {
		name           string
		requestBody    CreateWeightRecordRequest
		userID         uuid.UUID
		mockFunc       func(ctx context.Context, userID uuid.UUID, weight float64, recordedAt string) (*repository.WeightRecord, error)
		expectedStatus int
	}{
		{
			name: "正常に体重を記録できるべき",
			requestBody: CreateWeightRecordRequest{
				Weight:     65.5,
				RecordedAt: "2024-01-15",
			},
			userID: userID,
			mockFunc: func(ctx context.Context, uid uuid.UUID, weight float64, recordedAt string) (*repository.WeightRecord, error) {
				return &repository.WeightRecord{
					ID:         recordID,
					UserID:     uid,
					Weight:     weight,
					RecordedAt: recordedAt,
					CreatedAt:  time.Now(),
				}, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "未認証の場合は401を返すべき",
			requestBody: CreateWeightRecordRequest{
				Weight:     65.5,
				RecordedAt: "2024-01-15",
			},
			userID:         uuid.Nil,
			mockFunc:       nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "体重が範囲外（小さすぎる）の場合は400を返すべき",
			requestBody: CreateWeightRecordRequest{
				Weight:     0.05,
				RecordedAt: "2024-01-15",
			},
			userID:         userID,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "体重が範囲外（大きすぎる）の場合は400を返すべき",
			requestBody: CreateWeightRecordRequest{
				Weight:     1000.0,
				RecordedAt: "2024-01-15",
			},
			userID:         userID,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "日付形式が不正の場合は400を返すべき",
			requestBody: CreateWeightRecordRequest{
				Weight:     65.5,
				RecordedAt: "invalid-date",
			},
			userID:         userID,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "リポジトリエラーの場合は500を返すべき",
			requestBody: CreateWeightRecordRequest{
				Weight:     65.5,
				RecordedAt: "2024-01-15",
			},
			userID: userID,
			mockFunc: func(ctx context.Context, uid uuid.UUID, weight float64, recordedAt string) (*repository.WeightRecord, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockWeightRepository{
				CreateOrUpdateRecordFunc: tt.mockFunc,
			}
			handler := NewWeightHandler(mockRepo)

			body, _ := json.Marshal(tt.requestBody)
			req := createRequestWithUserID(http.MethodPost, "/api/weight-records", body, tt.userID)
			rec := httptest.NewRecorder()

			handler.HandleCreateRecord(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("期待されるステータス %d, 実際のステータス %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestHandleCreateRecord_InvalidJSON(t *testing.T) {
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
			mockRepo := &MockWeightRepository{}
			handler := NewWeightHandler(mockRepo)

			req := createRequestWithUserID(http.MethodPost, "/api/weight-records", tt.body, tt.userID)
			rec := httptest.NewRecorder()

			handler.HandleCreateRecord(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("期待されるステータス %d, 実際のステータス %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestHandleUpdateGoal_InvalidJSON(t *testing.T) {
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
			mockRepo := &MockWeightRepository{}
			handler := NewWeightHandler(mockRepo)

			req := createRequestWithUserID(http.MethodPut, "/api/weight-goal", tt.body, tt.userID)
			rec := httptest.NewRecorder()

			handler.HandleUpdateGoal(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("期待されるステータス %d, 実際のステータス %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestHandleGetRecords(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name                 string
		period               string
		userID               uuid.UUID
		mockRecords          []*repository.WeightRecord
		mockLatest           *repository.WeightRecord
		mockStats            *repository.WeightStats
		mockRecordsErr       error
		mockLatestErr        error
		mockStatsErr         error
		expectedStatus       int
		expectedRecordsCount int
	}{
		{
			name:   "正常に記録一覧を取得できるべき（week）",
			period: "week",
			userID: userID,
			mockRecords: []*repository.WeightRecord{
				{ID: uuid.New(), UserID: userID, Weight: 65.0, RecordedAt: "2024-01-14"},
				{ID: uuid.New(), UserID: userID, Weight: 65.5, RecordedAt: "2024-01-15"},
			},
			mockLatest:           &repository.WeightRecord{ID: uuid.New(), UserID: userID, Weight: 65.5, RecordedAt: "2024-01-15"},
			mockStats:            &repository.WeightStats{Min: 65.0, Max: 65.5, Average: 65.25},
			expectedStatus:       http.StatusOK,
			expectedRecordsCount: 2,
		},
		{
			name:           "未認証の場合は401を返すべき",
			period:         "week",
			userID:         uuid.Nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "記録取得エラーの場合は500を返すべき",
			period:         "week",
			userID:         userID,
			mockRecordsErr: errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "最新記録取得エラーの場合は500を返すべき",
			period: "week",
			userID: userID,
			mockRecords: []*repository.WeightRecord{
				{ID: uuid.New(), UserID: userID, Weight: 65.0, RecordedAt: "2024-01-14"},
			},
			mockLatestErr:  errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "統計取得エラーの場合は500を返すべき",
			period: "week",
			userID: userID,
			mockRecords: []*repository.WeightRecord{
				{ID: uuid.New(), UserID: userID, Weight: 65.0, RecordedAt: "2024-01-14"},
			},
			mockLatest:     &repository.WeightRecord{ID: uuid.New(), UserID: userID, Weight: 65.0, RecordedAt: "2024-01-14"},
			mockStatsErr:   errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "正常に記録一覧を取得できるべき（month）",
			period: "month",
			userID: userID,
			mockRecords: []*repository.WeightRecord{
				{ID: uuid.New(), UserID: userID, Weight: 65.0, RecordedAt: "2024-01-01"},
			},
			mockLatest:           &repository.WeightRecord{ID: uuid.New(), UserID: userID, Weight: 65.0, RecordedAt: "2024-01-01"},
			mockStats:            &repository.WeightStats{Min: 65.0, Max: 65.0, Average: 65.0},
			expectedStatus:       http.StatusOK,
			expectedRecordsCount: 1,
		},
		{
			name:   "正常に記録一覧を取得できるべき（3months）",
			period: "3months",
			userID: userID,
			mockRecords: []*repository.WeightRecord{
				{ID: uuid.New(), UserID: userID, Weight: 65.0, RecordedAt: "2024-01-01"},
			},
			mockLatest:           &repository.WeightRecord{ID: uuid.New(), UserID: userID, Weight: 65.0, RecordedAt: "2024-01-01"},
			mockStats:            &repository.WeightStats{Min: 65.0, Max: 65.0, Average: 65.0},
			expectedStatus:       http.StatusOK,
			expectedRecordsCount: 1,
		},
		{
			name:   "不正なperiodの場合はデフォルト（week）を使用すべき",
			period: "invalid-period",
			userID: userID,
			mockRecords: []*repository.WeightRecord{
				{ID: uuid.New(), UserID: userID, Weight: 65.0, RecordedAt: "2024-01-14"},
			},
			mockLatest:           &repository.WeightRecord{ID: uuid.New(), UserID: userID, Weight: 65.0, RecordedAt: "2024-01-14"},
			mockStats:            &repository.WeightStats{Min: 65.0, Max: 65.0, Average: 65.0},
			expectedStatus:       http.StatusOK,
			expectedRecordsCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockWeightRepository{
				GetRecordsByPeriodFunc: func(ctx context.Context, uid uuid.UUID, startDate, endDate string) ([]*repository.WeightRecord, error) {
					if tt.mockRecordsErr != nil {
						return nil, tt.mockRecordsErr
					}
					return tt.mockRecords, nil
				},
				GetLatestRecordFunc: func(ctx context.Context, uid uuid.UUID) (*repository.WeightRecord, error) {
					if tt.mockLatestErr != nil {
						return nil, tt.mockLatestErr
					}
					return tt.mockLatest, nil
				},
				GetStatsByPeriodFunc: func(ctx context.Context, uid uuid.UUID, startDate, endDate string) (*repository.WeightStats, error) {
					if tt.mockStatsErr != nil {
						return nil, tt.mockStatsErr
					}
					return tt.mockStats, nil
				},
			}
			handler := NewWeightHandler(mockRepo)

			url := "/api/weight-records?period=" + tt.period
			req := createRequestWithUserID(http.MethodGet, url, nil, tt.userID)
			rec := httptest.NewRecorder()

			handler.HandleGetRecords(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("期待されるステータス %d, 実際のステータス %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response WeightRecordsResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Errorf("レスポンスのパースに失敗: %v", err)
				}
				if len(response.Records) != tt.expectedRecordsCount {
					t.Errorf("期待される記録数 %d, 実際の記録数 %d", tt.expectedRecordsCount, len(response.Records))
				}
			}
		})
	}
}

func TestHandleGetGoal(t *testing.T) {
	userID := uuid.New()
	goalID := uuid.New()

	tests := []struct {
		name           string
		userID         uuid.UUID
		mockGoal       *repository.WeightGoal
		mockLatest     *repository.WeightRecord
		mockGoalErr    error
		mockLatestErr  error
		expectedStatus int
	}{
		{
			name:   "正常に目標を取得できるべき",
			userID: userID,
			mockGoal: &repository.WeightGoal{
				ID:           goalID,
				UserID:       userID,
				TargetWeight: 60.0,
				TargetDate:   time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
			},
			mockLatest:     &repository.WeightRecord{Weight: 65.0},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "目標が設定されていない場合はnullを返すべき",
			userID:         userID,
			mockGoal:       nil,
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
			mockGoalErr:    errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "最新記録取得エラーの場合は500を返すべき",
			userID: userID,
			mockGoal: &repository.WeightGoal{
				ID:           goalID,
				UserID:       userID,
				TargetWeight: 60.0,
				TargetDate:   time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
			},
			mockLatestErr:  errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "目標日の形式が不正な場合は500を返すべき",
			userID: userID,
			mockGoal: &repository.WeightGoal{
				ID:           goalID,
				UserID:       userID,
				TargetWeight: 60.0,
				TargetDate:   "invalid-date-format",
			},
			mockLatest:     &repository.WeightRecord{Weight: 65.0},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockWeightRepository{
				GetGoalFunc: func(ctx context.Context, uid uuid.UUID) (*repository.WeightGoal, error) {
					if tt.mockGoalErr != nil {
						return nil, tt.mockGoalErr
					}
					return tt.mockGoal, nil
				},
				GetLatestRecordFunc: func(ctx context.Context, uid uuid.UUID) (*repository.WeightRecord, error) {
					if tt.mockLatestErr != nil {
						return nil, tt.mockLatestErr
					}
					return tt.mockLatest, nil
				},
			}
			handler := NewWeightHandler(mockRepo)

			req := createRequestWithUserID(http.MethodGet, "/api/weight-goal", nil, tt.userID)
			rec := httptest.NewRecorder()

			handler.HandleGetGoal(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("期待されるステータス %d, 実際のステータス %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestHandleUpdateGoal(t *testing.T) {
	userID := uuid.New()
	goalID := uuid.New()

	tests := []struct {
		name           string
		requestBody    UpdateWeightGoalRequest
		userID         uuid.UUID
		mockFunc       func(ctx context.Context, userID uuid.UUID, targetWeight float64, targetDate string) (*repository.WeightGoal, error)
		expectedStatus int
	}{
		{
			name: "正常に目標を設定できるべき",
			requestBody: UpdateWeightGoalRequest{
				TargetWeight: 60.0,
				TargetDate:   "2024-06-01",
			},
			userID: userID,
			mockFunc: func(ctx context.Context, uid uuid.UUID, targetWeight float64, targetDate string) (*repository.WeightGoal, error) {
				return &repository.WeightGoal{
					ID:           goalID,
					UserID:       uid,
					TargetWeight: targetWeight,
					TargetDate:   targetDate,
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "未認証の場合は401を返すべき",
			requestBody: UpdateWeightGoalRequest{
				TargetWeight: 60.0,
				TargetDate:   "2024-06-01",
			},
			userID:         uuid.Nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "目標体重が範囲外の場合は400を返すべき",
			requestBody: UpdateWeightGoalRequest{
				TargetWeight: 0.05,
				TargetDate:   "2024-06-01",
			},
			userID:         userID,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "目標日が空の場合は400を返すべき",
			requestBody: UpdateWeightGoalRequest{
				TargetWeight: 60.0,
				TargetDate:   "",
			},
			userID:         userID,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "目標日の形式が不正の場合は400を返すべき",
			requestBody: UpdateWeightGoalRequest{
				TargetWeight: 60.0,
				TargetDate:   "invalid-date",
			},
			userID:         userID,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "リポジトリエラーの場合は500を返すべき",
			requestBody: UpdateWeightGoalRequest{
				TargetWeight: 60.0,
				TargetDate:   "2024-06-01",
			},
			userID: userID,
			mockFunc: func(ctx context.Context, uid uuid.UUID, targetWeight float64, targetDate string) (*repository.WeightGoal, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockWeightRepository{
				CreateOrUpdateGoalFunc: tt.mockFunc,
			}
			handler := NewWeightHandler(mockRepo)

			body, _ := json.Marshal(tt.requestBody)
			req := createRequestWithUserID(http.MethodPut, "/api/weight-goal", body, tt.userID)
			rec := httptest.NewRecorder()

			handler.HandleUpdateGoal(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("期待されるステータス %d, 実際のステータス %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}
