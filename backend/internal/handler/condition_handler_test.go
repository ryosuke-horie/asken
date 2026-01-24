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

type MockConditionRepository struct {
	CreateOrUpdateRecordFunc func(ctx context.Context, userID uuid.UUID, condition, fatigue int, recordedAt string) (*repository.ConditionRecord, error)
	GetRecordByDateFunc      func(ctx context.Context, userID uuid.UUID, date string) (*repository.ConditionRecord, error)
}

func (m *MockConditionRepository) CreateOrUpdateRecord(ctx context.Context, userID uuid.UUID, condition, fatigue int, recordedAt string) (*repository.ConditionRecord, error) {
	if m.CreateOrUpdateRecordFunc != nil {
		return m.CreateOrUpdateRecordFunc(ctx, userID, condition, fatigue, recordedAt)
	}
	return nil, nil
}

func (m *MockConditionRepository) GetRecordByDate(ctx context.Context, userID uuid.UUID, date string) (*repository.ConditionRecord, error) {
	if m.GetRecordByDateFunc != nil {
		return m.GetRecordByDateFunc(ctx, userID, date)
	}
	return nil, nil
}

func createConditionRequestWithUserID(method, url string, body []byte, userID uuid.UUID) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserIDToContext(req.Context(), userID)
	return req.WithContext(ctx)
}

func TestConditionHandler_HandleCreateRecord(t *testing.T) {
	userID := uuid.New()
	recordID := uuid.New()

	tests := []struct {
		name           string
		requestBody    CreateConditionRecordRequest
		userID         uuid.UUID
		mockFunc       func(ctx context.Context, userID uuid.UUID, condition, fatigue int, recordedAt string) (*repository.ConditionRecord, error)
		expectedStatus int
	}{
		{
			name: "正常に体調を記録できるべき",
			requestBody: CreateConditionRecordRequest{
				Condition:  3,
				Fatigue:    2,
				RecordedAt: "2024-01-15",
			},
			userID: userID,
			mockFunc: func(ctx context.Context, uid uuid.UUID, condition, fatigue int, recordedAt string) (*repository.ConditionRecord, error) {
				return &repository.ConditionRecord{
					ID:         recordID,
					UserID:     uid,
					Condition:  condition,
					Fatigue:    fatigue,
					RecordedAt: recordedAt,
					CreatedAt:  time.Now(),
				}, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "未認証の場合は401を返すべき",
			requestBody: CreateConditionRecordRequest{
				Condition:  3,
				Fatigue:    2,
				RecordedAt: "2024-01-15",
			},
			userID:         uuid.Nil,
			mockFunc:       nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "体調が範囲外（0）の場合は400を返すべき",
			requestBody: CreateConditionRecordRequest{
				Condition:  0,
				Fatigue:    2,
				RecordedAt: "2024-01-15",
			},
			userID:         userID,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "体調が範囲外（4）の場合は400を返すべき",
			requestBody: CreateConditionRecordRequest{
				Condition:  4,
				Fatigue:    2,
				RecordedAt: "2024-01-15",
			},
			userID:         userID,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "疲労度が範囲外（0）の場合は400を返すべき",
			requestBody: CreateConditionRecordRequest{
				Condition:  2,
				Fatigue:    0,
				RecordedAt: "2024-01-15",
			},
			userID:         userID,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "疲労度が範囲外（4）の場合は400を返すべき",
			requestBody: CreateConditionRecordRequest{
				Condition:  2,
				Fatigue:    4,
				RecordedAt: "2024-01-15",
			},
			userID:         userID,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "日付形式が不正の場合は400を返すべき",
			requestBody: CreateConditionRecordRequest{
				Condition:  3,
				Fatigue:    2,
				RecordedAt: "invalid-date",
			},
			userID:         userID,
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "リポジトリエラーの場合は500を返すべき",
			requestBody: CreateConditionRecordRequest{
				Condition:  3,
				Fatigue:    2,
				RecordedAt: "2024-01-15",
			},
			userID: userID,
			mockFunc: func(ctx context.Context, uid uuid.UUID, condition, fatigue int, recordedAt string) (*repository.ConditionRecord, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockConditionRepository{
				CreateOrUpdateRecordFunc: tt.mockFunc,
			}
			handler := NewConditionHandler(mockRepo)

			body, _ := json.Marshal(tt.requestBody)
			req := createConditionRequestWithUserID(http.MethodPost, "/api/condition-records", body, tt.userID)
			rec := httptest.NewRecorder()

			handler.HandleCreateRecord(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("期待されるステータス %d, 実際のステータス %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestConditionHandler_HandleCreateRecord_EmptyRecordedAt(t *testing.T) {
	userID := uuid.New()
	recordID := uuid.New()
	today := time.Now().Format("2006-01-02")

	var capturedRecordedAt string

	mockRepo := &MockConditionRepository{
		CreateOrUpdateRecordFunc: func(ctx context.Context, uid uuid.UUID, condition, fatigue int, recordedAt string) (*repository.ConditionRecord, error) {
			capturedRecordedAt = recordedAt
			return &repository.ConditionRecord{
				ID:         recordID,
				UserID:     uid,
				Condition:  condition,
				Fatigue:    fatigue,
				RecordedAt: recordedAt,
				CreatedAt:  time.Now(),
			}, nil
		},
	}
	handler := NewConditionHandler(mockRepo)

	requestBody := CreateConditionRecordRequest{
		Condition: 3,
		Fatigue:   2,
		// RecordedAt is intentionally empty
	}
	body, _ := json.Marshal(requestBody)
	req := createConditionRequestWithUserID(http.MethodPost, "/api/condition-records", body, userID)
	rec := httptest.NewRecorder()

	handler.HandleCreateRecord(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("期待されるステータス %d, 実際のステータス %d", http.StatusCreated, rec.Code)
	}

	if capturedRecordedAt != today {
		t.Errorf("recorded_atが空の場合は本日の日付が使用されるべき。期待値: %s, 実際の値: %s", today, capturedRecordedAt)
	}
}

func TestConditionHandler_HandleCreateRecord_InvalidJSON(t *testing.T) {
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
			mockRepo := &MockConditionRepository{}
			handler := NewConditionHandler(mockRepo)

			req := createConditionRequestWithUserID(http.MethodPost, "/api/condition-records", tt.body, tt.userID)
			rec := httptest.NewRecorder()

			handler.HandleCreateRecord(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("期待されるステータス %d, 実際のステータス %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestConditionHandler_HandleGetRecord(t *testing.T) {
	userID := uuid.New()
	recordID := uuid.New()

	tests := []struct {
		name           string
		date           string
		userID         uuid.UUID
		mockRecord     *repository.ConditionRecord
		mockErr        error
		expectedStatus int
	}{
		{
			name:   "正常に記録を取得できるべき",
			date:   "2024-01-15",
			userID: userID,
			mockRecord: &repository.ConditionRecord{
				ID:         recordID,
				UserID:     userID,
				Condition:  3,
				Fatigue:    2,
				RecordedAt: "2024-01-15",
				CreatedAt:  time.Now(),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "記録が存在しない場合はnullを返すべき",
			date:           "2024-01-15",
			userID:         userID,
			mockRecord:     nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "未認証の場合は401を返すべき",
			date:           "2024-01-15",
			userID:         uuid.Nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "日付が空の場合は400を返すべき",
			date:           "",
			userID:         userID,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "日付形式が不正の場合は400を返すべき",
			date:           "invalid-date",
			userID:         userID,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "リポジトリエラーの場合は500を返すべき",
			date:           "2024-01-15",
			userID:         userID,
			mockErr:        errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockConditionRepository{
				GetRecordByDateFunc: func(ctx context.Context, uid uuid.UUID, date string) (*repository.ConditionRecord, error) {
					if tt.mockErr != nil {
						return nil, tt.mockErr
					}
					return tt.mockRecord, nil
				},
			}
			handler := NewConditionHandler(mockRepo)

			url := "/api/condition-records?date=" + tt.date
			req := createConditionRequestWithUserID(http.MethodGet, url, nil, tt.userID)
			rec := httptest.NewRecorder()

			handler.HandleGetRecord(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("期待されるステータス %d, 実際のステータス %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}
