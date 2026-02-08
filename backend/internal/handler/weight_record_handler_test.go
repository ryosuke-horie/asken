package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeightRecordHandler_HandleCreate_Success(t *testing.T) {
	testUserID := "test-user-123"
	now := time.Now()

	mockRepo := &MockWeightRecordRepository{
		CreateRecordFunc: func(ctx context.Context, userID string, weightKg float64, recordedAt time.Time, note string) (*repository.WeightRecord, error) {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, 65.3, weightKg)
			assert.Equal(t, "朝食前", note)
			return &repository.WeightRecord{
				ID:         testRecordUUID,
				WeightKg:   65.3,
				RecordedAt: recordedAt,
				Note:       "朝食前",
				CreatedAt:  now,
				UpdatedAt:  now,
			}, nil
		},
	}

	handler := NewWeightRecordHandler(mockRepo, &MockWeightGoalRepository{})

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
	assert.Equal(t, testRecordUUID, response.ID)
	assert.Equal(t, 65.3, response.WeightKg)
	assert.Equal(t, "朝食前", response.Note)
}

func TestWeightRecordHandler_HandleCreate_ValidationErrors(t *testing.T) {
	mockRepo := &MockWeightRecordRepository{}
	handler := NewWeightRecordHandler(mockRepo, &MockWeightGoalRepository{})
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
	handler := NewWeightRecordHandler(mockRepo, &MockWeightGoalRepository{})

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
					ID:         testRecordUUID,
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
	}

	mockGoalRepo := &MockWeightGoalRepository{
		GetGoalFunc: func(ctx context.Context, userID string) (*repository.WeightGoal, error) {
			return &repository.WeightGoal{
				TargetWeightKg: 63.0,
				UpdatedAt:      now,
			}, nil
		},
	}

	handler := NewWeightRecordHandler(mockRepo, mockGoalRepo)

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
	handler := NewWeightRecordHandler(mockRepo, &MockWeightGoalRepository{})
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
			assert.Equal(t, testRecordUUID, recordID)
			assert.Equal(t, 64.8, weightKg)
			return &repository.WeightRecord{
				ID:         testRecordUUID,
				WeightKg:   64.8,
				RecordedAt: now,
				Note:       "更新後",
				CreatedAt:  now,
				UpdatedAt:  now,
			}, nil
		},
	}

	handler := NewWeightRecordHandler(mockRepo, &MockWeightGoalRepository{})

	body := UpdateWeightRecordRequest{WeightKg: 64.8, Note: "更新後"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/weight/records/550e8400-e29b-41d4-a716-446655440000", bytes.NewReader(bodyBytes))
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

	handler := NewWeightRecordHandler(mockRepo, &MockWeightGoalRepository{})

	body := UpdateWeightRecordRequest{WeightKg: 64.8}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/weight/records/550e8400-e29b-41d4-a716-446655440001", bytes.NewReader(bodyBytes))
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
			assert.Equal(t, testRecordUUID, recordID)
			return nil
		},
	}

	handler := NewWeightRecordHandler(mockRepo, &MockWeightGoalRepository{})

	req := httptest.NewRequest(http.MethodDelete, "/api/weight/records/550e8400-e29b-41d4-a716-446655440000", nil)
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

	handler := NewWeightRecordHandler(mockRepo, &MockWeightGoalRepository{})

	req := httptest.NewRequest(http.MethodDelete, "/api/weight/records/550e8400-e29b-41d4-a716-446655440001", nil)
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
			assert.Equal(t, testRecordUUID, recordID)
			return &repository.WeightRecord{
				ID:         testRecordUUID,
				WeightKg:   65.3,
				RecordedAt: now,
				Note:       "朝食前",
				CreatedAt:  now,
				UpdatedAt:  now,
			}, nil
		},
	}

	handler := NewWeightRecordHandler(mockRepo, &MockWeightGoalRepository{})

	req := httptest.NewRequest(http.MethodGet, "/api/weight/records/550e8400-e29b-41d4-a716-446655440000", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response WeightRecordResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, testRecordUUID, response.ID)
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

	handler := NewWeightRecordHandler(mockRepo, &MockWeightGoalRepository{})

	req := httptest.NewRequest(http.MethodGet, "/api/weight/records/550e8400-e29b-41d4-a716-446655440001", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWeightRecordHandler_HandleGet_Unauthorized(t *testing.T) {
	mockRepo := &MockWeightRecordRepository{}
	handler := NewWeightRecordHandler(mockRepo, &MockWeightGoalRepository{})

	req := httptest.NewRequest(http.MethodGet, "/api/weight/records/550e8400-e29b-41d4-a716-446655440000", nil)
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

	handler := NewWeightRecordHandler(mockRepo, &MockWeightGoalRepository{})

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

	handler := NewWeightRecordHandler(mockRepo, &MockWeightGoalRepository{})

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
	}

	mockGoalRepo := &MockWeightGoalRepository{
		GetGoalFunc: func(ctx context.Context, userID string) (*repository.WeightGoal, error) {
			return nil, fmt.Errorf("database error")
		},
	}

	handler := NewWeightRecordHandler(mockRepo, mockGoalRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/weight/records?from=2026-02-01&to=2026-02-28", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleList(w, req)

	// Goal取得失敗時は非致命的（200 OKを返す）
	assert.Equal(t, http.StatusOK, w.Code)

	var response WeightRecordsListResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Nil(t, response.Goal)
}

func TestWeightRecordHandler_HandleList_DateRangeLimit(t *testing.T) {
	testUserID := "test-user-123"

	mockRepo := &MockWeightRecordRepository{
		ListRecordsByDateRangeFunc: func(ctx context.Context, userID string, from time.Time, to time.Time) ([]repository.WeightRecord, error) {
			return []repository.WeightRecord{}, nil
		},
	}

	mockGoalRepo := &MockWeightGoalRepository{
		GetGoalFunc: func(ctx context.Context, userID string) (*repository.WeightGoal, error) {
			return nil, nil
		},
	}

	handler := NewWeightRecordHandler(mockRepo, mockGoalRepo)

	tests := []struct {
		name     string
		from     string
		to       string
		wantCode int
	}{
		{
			name:     "367日間は拒否される",
			from:     "2025-01-01",
			to:       "2026-01-03",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "366日間は許可される",
			from:     "2025-01-01",
			to:       "2026-01-01",
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/weight/records?from=%s&to=%s", tt.from, tt.to)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.HandleList(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestWeightRecordHandler_HandleCreate_NoteTooLong(t *testing.T) {
	testUserID := "test-user-123"

	mockRepo := &MockWeightRecordRepository{}
	handler := NewWeightRecordHandler(mockRepo, &MockWeightGoalRepository{})

	// 201文字の日本語メモ（マルチバイト）
	longNote := ""
	for i := 0; i < 201; i++ {
		longNote += "あ"
	}

	body := CreateWeightRecordRequest{
		WeightKg:   65.0,
		RecordedAt: time.Now().Format(time.RFC3339),
		Note:       longNote,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/weight/records", bytes.NewReader(bodyBytes))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleCreate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWeightRecordHandler_HandleUpdate_ValidationErrors(t *testing.T) {
	mockRepo := &MockWeightRecordRepository{}
	handler := NewWeightRecordHandler(mockRepo, &MockWeightGoalRepository{})
	testUserID := "test-user-123"

	tests := []struct {
		name     string
		body     string
		wantCode int
	}{
		{
			name:     "体重が範囲外（低すぎ）",
			body:     `{"weight_kg": 10.0, "note": ""}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "体重が範囲外（高すぎ）",
			body:     `{"weight_kg": 500.0, "note": ""}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "noteが200文字超",
			body:     `{"weight_kg": 65.0, "note": "` + strings.Repeat("あ", 201) + `"}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "不正なJSON",
			body:     `invalid json`,
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/api/weight/records/550e8400-e29b-41d4-a716-446655440000", bytes.NewReader([]byte(tt.body)))
			ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.HandleUpdate(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestWeightRecordHandler_HandleUpdate_Unauthorized(t *testing.T) {
	handler := NewWeightRecordHandler(&MockWeightRecordRepository{}, &MockWeightGoalRepository{})

	body := UpdateWeightRecordRequest{WeightKg: 65.0}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/weight/records/550e8400-e29b-41d4-a716-446655440000", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWeightRecordHandler_HandleUpdate_RepositoryError(t *testing.T) {
	testUserID := "test-user-123"

	mockRepo := &MockWeightRecordRepository{
		UpdateRecordFunc: func(ctx context.Context, userID string, recordID string, weightKg float64, note string) (*repository.WeightRecord, error) {
			return nil, fmt.Errorf("database error")
		},
	}

	handler := NewWeightRecordHandler(mockRepo, &MockWeightGoalRepository{})

	body := UpdateWeightRecordRequest{WeightKg: 64.8}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/weight/records/550e8400-e29b-41d4-a716-446655440000", bytes.NewReader(bodyBytes))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleUpdate(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWeightRecordHandler_HandleDelete_Unauthorized(t *testing.T) {
	handler := NewWeightRecordHandler(&MockWeightRecordRepository{}, &MockWeightGoalRepository{})

	req := httptest.NewRequest(http.MethodDelete, "/api/weight/records/550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()

	handler.HandleDelete(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWeightRecordHandler_HandleDelete_RepositoryError(t *testing.T) {
	testUserID := "test-user-123"

	mockRepo := &MockWeightRecordRepository{
		DeleteRecordFunc: func(ctx context.Context, userID string, recordID string) error {
			return fmt.Errorf("database error")
		},
	}

	handler := NewWeightRecordHandler(mockRepo, &MockWeightGoalRepository{})

	req := httptest.NewRequest(http.MethodDelete, "/api/weight/records/550e8400-e29b-41d4-a716-446655440000", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleDelete(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWeightRecordHandler_HandleList_InvalidTimezone(t *testing.T) {
	handler := NewWeightRecordHandler(&MockWeightRecordRepository{}, &MockWeightGoalRepository{})
	testUserID := "test-user-123"

	req := httptest.NewRequest(http.MethodGet, "/api/weight/records?from=2026-02-01&to=2026-02-28&tz=Invalid/Zone", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleList(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWeightRecordHandler_HandleList_InvalidDates(t *testing.T) {
	handler := NewWeightRecordHandler(&MockWeightRecordRepository{}, &MockWeightGoalRepository{})
	testUserID := "test-user-123"

	tests := []struct {
		name     string
		url      string
		wantCode int
	}{
		{
			name:     "無効なfrom日付",
			url:      "/api/weight/records?from=not-a-date&to=2026-02-28",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "無効なto日付",
			url:      "/api/weight/records?from=2026-02-01&to=invalid",
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.HandleList(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestWeightRecordHandler_HandleCreate_InvalidJSON(t *testing.T) {
	handler := NewWeightRecordHandler(&MockWeightRecordRepository{}, &MockWeightGoalRepository{})
	testUserID := "test-user-123"

	req := httptest.NewRequest(http.MethodPost, "/api/weight/records", bytes.NewReader([]byte("{invalid json}")))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleCreate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestValidateWeightKg_BoundaryValues(t *testing.T) {
	tests := []struct {
		name    string
		weight  float64
		wantErr bool
	}{
		{"19.9は拒否", 19.9, true},
		{"20.0は許可", 20.0, false},
		{"300.0は許可", 300.0, false},
		{"300.1は拒否", 300.1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWeightKg(tt.weight)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWeightRecordHandler_ExtractRecordID_InvalidUUID(t *testing.T) {
	mockRepo := &MockWeightRecordRepository{}
	handler := NewWeightRecordHandler(mockRepo, &MockWeightGoalRepository{})
	testUserID := "test-user-123"

	// 不正なUUID形式のID
	req := httptest.NewRequest(http.MethodGet, "/api/weight/records/not-a-valid-uuid", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleGet(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
