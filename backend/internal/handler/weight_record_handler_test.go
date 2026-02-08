package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockWeightRecordRepository はテスト用のモック
type MockWeightRecordRepository struct {
	CreateRecordFunc          func(ctx context.Context, userID string, weightKg float64, recordedAt time.Time, note string) (*repository.WeightRecord, error)
	GetRecordFunc             func(ctx context.Context, userID string, recordID string) (*repository.WeightRecord, error)
	UpdateRecordFunc          func(ctx context.Context, userID string, recordID string, weightKg float64, note string) (*repository.WeightRecord, error)
	DeleteRecordFunc          func(ctx context.Context, userID string, recordID string) error
	ListRecordsByDateRangeFunc func(ctx context.Context, userID string, from time.Time, to time.Time) ([]repository.WeightRecord, error)
	GetGoalFunc               func(ctx context.Context, userID string) (*repository.WeightGoal, error)
	SetGoalFunc               func(ctx context.Context, userID string, targetWeightKg float64) (*repository.WeightGoal, error)
}

func (m *MockWeightRecordRepository) CreateRecord(ctx context.Context, userID string, weightKg float64, recordedAt time.Time, note string) (*repository.WeightRecord, error) {
	if m.CreateRecordFunc != nil {
		return m.CreateRecordFunc(ctx, userID, weightKg, recordedAt, note)
	}
	return nil, nil
}

func (m *MockWeightRecordRepository) GetRecord(ctx context.Context, userID string, recordID string) (*repository.WeightRecord, error) {
	if m.GetRecordFunc != nil {
		return m.GetRecordFunc(ctx, userID, recordID)
	}
	return nil, nil
}

func (m *MockWeightRecordRepository) UpdateRecord(ctx context.Context, userID string, recordID string, weightKg float64, note string) (*repository.WeightRecord, error) {
	if m.UpdateRecordFunc != nil {
		return m.UpdateRecordFunc(ctx, userID, recordID, weightKg, note)
	}
	return nil, nil
}

func (m *MockWeightRecordRepository) DeleteRecord(ctx context.Context, userID string, recordID string) error {
	if m.DeleteRecordFunc != nil {
		return m.DeleteRecordFunc(ctx, userID, recordID)
	}
	return nil
}

func (m *MockWeightRecordRepository) ListRecordsByDateRange(ctx context.Context, userID string, from time.Time, to time.Time) ([]repository.WeightRecord, error) {
	if m.ListRecordsByDateRangeFunc != nil {
		return m.ListRecordsByDateRangeFunc(ctx, userID, from, to)
	}
	return nil, nil
}

func (m *MockWeightRecordRepository) GetGoal(ctx context.Context, userID string) (*repository.WeightGoal, error) {
	if m.GetGoalFunc != nil {
		return m.GetGoalFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockWeightRecordRepository) SetGoal(ctx context.Context, userID string, targetWeightKg float64) (*repository.WeightGoal, error) {
	if m.SetGoalFunc != nil {
		return m.SetGoalFunc(ctx, userID, targetWeightKg)
	}
	return nil, nil
}

func TestWeightRecordHandler_HandleCreate_Success(t *testing.T) {
	testUserID := "test-user-123"
	now := time.Now()

	mockRepo := &MockWeightRecordRepository{
		CreateRecordFunc: func(ctx context.Context, userID string, weightKg float64, recordedAt time.Time, note string) (*repository.WeightRecord, error) {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, 65.3, weightKg)
			assert.Equal(t, "朝食前", note)
			return &repository.WeightRecord{
				ID:         "record-1",
				WeightKg:   65.3,
				RecordedAt: recordedAt,
				Note:       "朝食前",
				CreatedAt:  now,
				UpdatedAt:  now,
			}, nil
		},
	}

	handler := NewWeightRecordHandler(mockRepo)

	body := CreateWeightRecordRequest{
		WeightKg:   65.3,
		RecordedAt: now.Format(time.RFC3339),
		Note:       "朝食前",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/weight/records", bytes.NewReader(bodyBytes))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleCreate(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response WeightRecordResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "record-1", response.ID)
	assert.Equal(t, 65.3, response.WeightKg)
	assert.Equal(t, "朝食前", response.Note)
}

func TestWeightRecordHandler_HandleCreate_ValidationErrors(t *testing.T) {
	mockRepo := &MockWeightRecordRepository{}
	handler := NewWeightRecordHandler(mockRepo)
	testUserID := "test-user-123"

	tests := []struct {
		name     string
		body     CreateWeightRecordRequest
		wantCode int
	}{
		{
			name:     "体重が範囲外（低すぎ）",
			body:     CreateWeightRecordRequest{WeightKg: 10.0, RecordedAt: time.Now().Format(time.RFC3339)},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "体重が範囲外（高すぎ）",
			body:     CreateWeightRecordRequest{WeightKg: 500.0, RecordedAt: time.Now().Format(time.RFC3339)},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "recorded_atが未指定",
			body:     CreateWeightRecordRequest{WeightKg: 65.0},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "未来の日時",
			body:     CreateWeightRecordRequest{WeightKg: 65.0, RecordedAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339)},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/weight/records", bytes.NewReader(bodyBytes))
			ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.HandleCreate(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestWeightRecordHandler_HandleCreate_Unauthorized(t *testing.T) {
	mockRepo := &MockWeightRecordRepository{}
	handler := NewWeightRecordHandler(mockRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/weight/records", nil)
	w := httptest.NewRecorder()

	handler.HandleCreate(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWeightRecordHandler_HandleList_Success(t *testing.T) {
	testUserID := "test-user-123"
	now := time.Now()

	mockRepo := &MockWeightRecordRepository{
		ListRecordsByDateRangeFunc: func(ctx context.Context, userID string, from time.Time, to time.Time) ([]repository.WeightRecord, error) {
			assert.Equal(t, testUserID, userID)
			return []repository.WeightRecord{
				{
					ID:         "record-1",
					WeightKg:   65.3,
					RecordedAt: time.Date(2026, 2, 8, 7, 30, 0, 0, time.UTC),
					Note:       "朝",
					CreatedAt:  now,
					UpdatedAt:  now,
				},
				{
					ID:         "record-2",
					WeightKg:   65.1,
					RecordedAt: time.Date(2026, 2, 8, 20, 0, 0, 0, time.UTC),
					Note:       "夜",
					CreatedAt:  now,
					UpdatedAt:  now,
				},
			}, nil
		},
		GetGoalFunc: func(ctx context.Context, userID string) (*repository.WeightGoal, error) {
			return &repository.WeightGoal{
				TargetWeightKg: 63.0,
				UpdatedAt:      now,
			}, nil
		},
	}

	handler := NewWeightRecordHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/weight/records?from=2026-02-01&to=2026-02-28", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response WeightRecordsListResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Len(t, response.Records, 2)
	assert.NotNil(t, response.Goal)
	assert.Equal(t, 63.0, response.Goal.TargetWeightKg)
	// 同じ日なのでdaily_summaryは1エントリ
	assert.Len(t, response.DailySummary, 1)
	summary, exists := response.DailySummary["2026-02-08"]
	assert.True(t, exists)
	assert.Equal(t, 2, summary.Count)
	assert.Equal(t, 65.1, summary.LatestWeight) // 最後のレコードの体重
}

func TestWeightRecordHandler_HandleList_MissingParams(t *testing.T) {
	mockRepo := &MockWeightRecordRepository{}
	handler := NewWeightRecordHandler(mockRepo)
	testUserID := "test-user-123"

	req := httptest.NewRequest(http.MethodGet, "/api/weight/records", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleList(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWeightRecordHandler_HandleUpdate_Success(t *testing.T) {
	testUserID := "test-user-123"
	now := time.Now()

	mockRepo := &MockWeightRecordRepository{
		UpdateRecordFunc: func(ctx context.Context, userID string, recordID string, weightKg float64, note string) (*repository.WeightRecord, error) {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, "record-1", recordID)
			assert.Equal(t, 64.8, weightKg)
			return &repository.WeightRecord{
				ID:         "record-1",
				WeightKg:   64.8,
				RecordedAt: now,
				Note:       "更新後",
				CreatedAt:  now,
				UpdatedAt:  now,
			}, nil
		},
	}

	handler := NewWeightRecordHandler(mockRepo)

	body := UpdateWeightRecordRequest{WeightKg: 64.8, Note: "更新後"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/weight/records/record-1", bytes.NewReader(bodyBytes))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWeightRecordHandler_HandleUpdate_NotFound(t *testing.T) {
	testUserID := "test-user-123"

	mockRepo := &MockWeightRecordRepository{
		UpdateRecordFunc: func(ctx context.Context, userID string, recordID string, weightKg float64, note string) (*repository.WeightRecord, error) {
			return nil, fmt.Errorf("体重記録が見つかりません: %w", repository.ErrNotFound)
		},
	}

	handler := NewWeightRecordHandler(mockRepo)

	body := UpdateWeightRecordRequest{WeightKg: 64.8}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/weight/records/nonexistent", bytes.NewReader(bodyBytes))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWeightRecordHandler_HandleDelete_Success(t *testing.T) {
	testUserID := "test-user-123"

	mockRepo := &MockWeightRecordRepository{
		DeleteRecordFunc: func(ctx context.Context, userID string, recordID string) error {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, "record-1", recordID)
			return nil
		},
	}

	handler := NewWeightRecordHandler(mockRepo)

	req := httptest.NewRequest(http.MethodDelete, "/api/weight/records/record-1", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleDelete(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestWeightRecordHandler_HandleDelete_NotFound(t *testing.T) {
	testUserID := "test-user-123"

	mockRepo := &MockWeightRecordRepository{
		DeleteRecordFunc: func(ctx context.Context, userID string, recordID string) error {
			return fmt.Errorf("体重記録が見つかりません: %w", repository.ErrNotFound)
		},
	}

	handler := NewWeightRecordHandler(mockRepo)

	req := httptest.NewRequest(http.MethodDelete, "/api/weight/records/nonexistent", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleDelete(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWeightRecordHandler_HandleGet_Success(t *testing.T) {
	testUserID := "test-user-123"
	now := time.Now()

	mockRepo := &MockWeightRecordRepository{
		GetRecordFunc: func(ctx context.Context, userID string, recordID string) (*repository.WeightRecord, error) {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, "record-1", recordID)
			return &repository.WeightRecord{
				ID:         "record-1",
				WeightKg:   65.3,
				RecordedAt: now,
				Note:       "朝食前",
				CreatedAt:  now,
				UpdatedAt:  now,
			}, nil
		},
	}

	handler := NewWeightRecordHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/weight/records/record-1", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response WeightRecordResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "record-1", response.ID)
	assert.Equal(t, 65.3, response.WeightKg)
	assert.Equal(t, "朝食前", response.Note)
}

func TestWeightRecordHandler_HandleGet_NotFound(t *testing.T) {
	testUserID := "test-user-123"

	mockRepo := &MockWeightRecordRepository{
		GetRecordFunc: func(ctx context.Context, userID string, recordID string) (*repository.WeightRecord, error) {
			return nil, fmt.Errorf("体重記録が見つかりません: %w", repository.ErrNotFound)
		},
	}

	handler := NewWeightRecordHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/weight/records/nonexistent", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWeightRecordHandler_HandleGet_Unauthorized(t *testing.T) {
	mockRepo := &MockWeightRecordRepository{}
	handler := NewWeightRecordHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/weight/records/record-1", nil)
	w := httptest.NewRecorder()

	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWeightRecordHandler_HandleCreate_RepositoryError(t *testing.T) {
	testUserID := "test-user-123"

	mockRepo := &MockWeightRecordRepository{
		CreateRecordFunc: func(ctx context.Context, userID string, weightKg float64, recordedAt time.Time, note string) (*repository.WeightRecord, error) {
			return nil, fmt.Errorf("database error")
		},
	}

	handler := NewWeightRecordHandler(mockRepo)

	body := CreateWeightRecordRequest{
		WeightKg:   65.3,
		RecordedAt: time.Now().Format(time.RFC3339),
		Note:       "テスト",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/weight/records", bytes.NewReader(bodyBytes))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleCreate(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWeightRecordHandler_HandleList_RepositoryError(t *testing.T) {
	testUserID := "test-user-123"

	mockRepo := &MockWeightRecordRepository{
		ListRecordsByDateRangeFunc: func(ctx context.Context, userID string, from time.Time, to time.Time) ([]repository.WeightRecord, error) {
			return nil, fmt.Errorf("database error")
		},
	}

	handler := NewWeightRecordHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/weight/records?from=2026-02-01&to=2026-02-28", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleList(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWeightRecordHandler_HandleList_GoalError(t *testing.T) {
	testUserID := "test-user-123"

	mockRepo := &MockWeightRecordRepository{
		ListRecordsByDateRangeFunc: func(ctx context.Context, userID string, from time.Time, to time.Time) ([]repository.WeightRecord, error) {
			return []repository.WeightRecord{}, nil
		},
		GetGoalFunc: func(ctx context.Context, userID string) (*repository.WeightGoal, error) {
			return nil, fmt.Errorf("database error")
		},
	}

	handler := NewWeightRecordHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/weight/records?from=2026-02-01&to=2026-02-28", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleList(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
