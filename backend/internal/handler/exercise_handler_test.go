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

	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockExerciseService はテスト用のExerciseServiceモック
type MockExerciseService struct {
	CreateExerciseRecordFunc func(ctx context.Context, userID string, input service.CreateExerciseInput, recordedDate string) (*repository.ExerciseRecord, error)
	GetDailyExerciseFunc     func(ctx context.Context, userID string, recordedDate string) (*repository.ExerciseDailyResult, error)
	DeleteExerciseRecordFunc func(ctx context.Context, userID string, recordID string) error
}

func (m *MockExerciseService) CreateExerciseRecord(ctx context.Context, userID string, input service.CreateExerciseInput, recordedDate string) (*repository.ExerciseRecord, error) {
	if m.CreateExerciseRecordFunc != nil {
		return m.CreateExerciseRecordFunc(ctx, userID, input, recordedDate)
	}
	return nil, nil
}

func (m *MockExerciseService) GetDailyExercise(ctx context.Context, userID string, recordedDate string) (*repository.ExerciseDailyResult, error) {
	if m.GetDailyExerciseFunc != nil {
		return m.GetDailyExerciseFunc(ctx, userID, recordedDate)
	}
	return &repository.ExerciseDailyResult{Records: []repository.ExerciseRecord{}}, nil
}

func (m *MockExerciseService) DeleteExerciseRecord(ctx context.Context, userID string, recordID string) error {
	if m.DeleteExerciseRecordFunc != nil {
		return m.DeleteExerciseRecordFunc(ctx, userID, recordID)
	}
	return nil
}

func TestExerciseHandler_HandleCreate_Success(t *testing.T) {
	testUserID := "test-user-123"
	now := time.Now()

	mockSvc := &MockExerciseService{
		CreateExerciseRecordFunc: func(ctx context.Context, userID string, input service.CreateExerciseInput, recordedDate string) (*repository.ExerciseRecord, error) {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, "柔術", input.ExerciseName)
			assert.Equal(t, 90, input.DurationMinutes)
			assert.Equal(t, "2026-02-28", recordedDate)
			return &repository.ExerciseRecord{
				ID:                 "record-001",
				ExerciseName:       "柔術",
				DurationMinutes:    90,
				BurnedCaloriesKcal: 1102.5,
				EstimationMethod:   repository.EstimationMethodMET,
				RecordedDate:       "2026-02-28",
				CreatedAt:          now,
			}, nil
		},
	}

	handler := NewExerciseHandler(mockSvc)

	body := CreateExerciseRecordRequest{
		ExerciseName:    "柔術",
		DurationMinutes: 90,
		RecordedDate:    "2026-02-28",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/exercise/records", bytes.NewReader(bodyBytes))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleCreate(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response ExerciseRecordResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "record-001", response.ID)
	assert.Equal(t, "柔術", response.ExerciseName)
	assert.Equal(t, 90, response.DurationMinutes)
	assert.InDelta(t, 1102.5, response.BurnedCaloriesKcal, 0.01)
	assert.Equal(t, "met", response.EstimationMethod)
	assert.Equal(t, "2026-02-28", response.RecordedDate)
}

func TestExerciseHandler_HandleCreate_Unauthorized(t *testing.T) {
	handler := NewExerciseHandler(&MockExerciseService{})

	body := CreateExerciseRecordRequest{
		ExerciseName:    "柔術",
		DurationMinutes: 90,
		RecordedDate:    "2026-02-28",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/exercise/records", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.HandleCreate(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestExerciseHandler_HandleCreate_MethodNotAllowed(t *testing.T) {
	handler := NewExerciseHandler(&MockExerciseService{})

	req := httptest.NewRequest(http.MethodGet, "/api/exercise/records", nil)
	w := httptest.NewRecorder()

	handler.HandleCreate(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestExerciseHandler_HandleCreate_InvalidJSON(t *testing.T) {
	testUserID := "test-user-123"
	handler := NewExerciseHandler(&MockExerciseService{})

	req := httptest.NewRequest(http.MethodPost, "/api/exercise/records", bytes.NewReader([]byte("{invalid json}")))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleCreate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExerciseHandler_HandleCreate_MissingRecordedDate(t *testing.T) {
	testUserID := "test-user-123"
	handler := NewExerciseHandler(&MockExerciseService{})

	body := CreateExerciseRecordRequest{
		ExerciseName:    "柔術",
		DurationMinutes: 90,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/exercise/records", bytes.NewReader(bodyBytes))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleCreate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExerciseHandler_HandleCreate_ServiceError(t *testing.T) {
	testUserID := "test-user-123"

	mockSvc := &MockExerciseService{
		CreateExerciseRecordFunc: func(ctx context.Context, userID string, input service.CreateExerciseInput, recordedDate string) (*repository.ExerciseRecord, error) {
			return nil, &service.ValidationError{Message: "exercise_nameは必須です"}
		},
	}
	handler := NewExerciseHandler(mockSvc)

	body := CreateExerciseRecordRequest{
		ExerciseName:    "",
		DurationMinutes: 90,
		RecordedDate:    "2026-02-28",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/exercise/records", bytes.NewReader(bodyBytes))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleCreate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExerciseHandler_HandleCreate_InternalServerError(t *testing.T) {
	testUserID := "test-user-123"

	mockSvc := &MockExerciseService{
		CreateExerciseRecordFunc: func(ctx context.Context, userID string, input service.CreateExerciseInput, recordedDate string) (*repository.ExerciseRecord, error) {
			return nil, errors.New("database connection error")
		},
	}
	handler := NewExerciseHandler(mockSvc)

	body := CreateExerciseRecordRequest{
		ExerciseName:    "柔術",
		DurationMinutes: 90,
		RecordedDate:    "2026-02-28",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/exercise/records", bytes.NewReader(bodyBytes))
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleCreate(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestExerciseHandler_HandleListByDate_Success(t *testing.T) {
	testUserID := "test-user-123"
	now := time.Now()

	mockSvc := &MockExerciseService{
		GetDailyExerciseFunc: func(ctx context.Context, userID string, recordedDate string) (*repository.ExerciseDailyResult, error) {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, "2026-02-28", recordedDate)
			return &repository.ExerciseDailyResult{
				Records: []repository.ExerciseRecord{
					{
						ID:                 "record-001",
						ExerciseName:       "柔術",
						DurationMinutes:    90,
						BurnedCaloriesKcal: 1102.5,
						EstimationMethod:   repository.EstimationMethodMET,
						RecordedDate:       "2026-02-28",
						CreatedAt:          now,
					},
					{
						ID:                 "record-002",
						ExerciseName:       "ランニング",
						DurationMinutes:    60,
						BurnedCaloriesKcal: 720.3,
						EstimationMethod:   repository.EstimationMethodMET,
						RecordedDate:       "2026-02-28",
						CreatedAt:          now,
					},
				},
				TotalBurnedCaloriesKcal: 1822.8,
			}, nil
		},
	}

	handler := NewExerciseHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/exercise/daily?date=2026-02-28", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleListByDate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DailyExerciseResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Len(t, response.Records, 2)
	assert.InDelta(t, 1822.8, response.TotalBurnedCaloriesKcal, 0.01)
	assert.Equal(t, "柔術", response.Records[0].ExerciseName)
	assert.Equal(t, "ランニング", response.Records[1].ExerciseName)
}

func TestExerciseHandler_HandleListByDate_Empty(t *testing.T) {
	testUserID := "test-user-123"

	mockSvc := &MockExerciseService{
		GetDailyExerciseFunc: func(ctx context.Context, userID string, recordedDate string) (*repository.ExerciseDailyResult, error) {
			return &repository.ExerciseDailyResult{
				Records:                 []repository.ExerciseRecord{},
				TotalBurnedCaloriesKcal: 0,
			}, nil
		},
	}

	handler := NewExerciseHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/exercise/daily?date=2026-02-28", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleListByDate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DailyExerciseResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Empty(t, response.Records)
	assert.Equal(t, 0.0, response.TotalBurnedCaloriesKcal)
}

func TestExerciseHandler_HandleListByDate_Unauthorized(t *testing.T) {
	handler := NewExerciseHandler(&MockExerciseService{})

	req := httptest.NewRequest(http.MethodGet, "/api/exercise/daily?date=2026-02-28", nil)
	w := httptest.NewRecorder()

	handler.HandleListByDate(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestExerciseHandler_HandleListByDate_MethodNotAllowed(t *testing.T) {
	handler := NewExerciseHandler(&MockExerciseService{})

	req := httptest.NewRequest(http.MethodPost, "/api/exercise/daily", nil)
	w := httptest.NewRecorder()

	handler.HandleListByDate(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestExerciseHandler_HandleListByDate_MissingDate(t *testing.T) {
	testUserID := "test-user-123"
	handler := NewExerciseHandler(&MockExerciseService{})

	req := httptest.NewRequest(http.MethodGet, "/api/exercise/daily", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleListByDate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExerciseHandler_HandleListByDate_InvalidDateFormat(t *testing.T) {
	testUserID := "test-user-123"
	handler := NewExerciseHandler(&MockExerciseService{})

	tests := []struct {
		name string
		date string
	}{
		{"スラッシュ区切り", "2026/02/28"},
		{"フォーマット不正", "26-02-28"},
		{"日付なし文字列", "not-a-date"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/exercise/daily?date="+tt.date, nil)
			ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.HandleListByDate(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestExerciseHandler_HandleListByDate_ServiceError(t *testing.T) {
	testUserID := "test-user-123"

	mockSvc := &MockExerciseService{
		GetDailyExerciseFunc: func(ctx context.Context, userID string, recordedDate string) (*repository.ExerciseDailyResult, error) {
			return nil, errors.New("database error")
		},
	}
	handler := NewExerciseHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/exercise/daily?date=2026-02-28", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleListByDate(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestExerciseHandler_HandleDelete_Success(t *testing.T) {
	testUserID := "test-user-123"
	validID := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

	mockSvc := &MockExerciseService{
		DeleteExerciseRecordFunc: func(ctx context.Context, userID string, recordID string) error {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, validID, recordID)
			return nil
		},
	}

	handler := NewExerciseHandler(mockSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/exercise/records/"+validID, nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleDelete(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestExerciseHandler_HandleDelete_Unauthorized(t *testing.T) {
	handler := NewExerciseHandler(&MockExerciseService{})

	req := httptest.NewRequest(http.MethodDelete, "/api/exercise/records/record-001", nil)
	w := httptest.NewRecorder()

	handler.HandleDelete(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestExerciseHandler_HandleDelete_MethodNotAllowed(t *testing.T) {
	handler := NewExerciseHandler(&MockExerciseService{})

	req := httptest.NewRequest(http.MethodGet, "/api/exercise/records/record-001", nil)
	w := httptest.NewRecorder()

	handler.HandleDelete(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestExerciseHandler_HandleDelete_NotFound(t *testing.T) {
	testUserID := "test-user-123"
	validID := "00000000-0000-0000-0000-000000000001"

	mockSvc := &MockExerciseService{
		DeleteExerciseRecordFunc: func(ctx context.Context, userID string, recordID string) error {
			return repository.ErrNotFound
		},
	}
	handler := NewExerciseHandler(mockSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/exercise/records/"+validID, nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleDelete(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestExerciseHandler_HandleDelete_InvalidID(t *testing.T) {
	testUserID := "test-user-123"
	handler := NewExerciseHandler(&MockExerciseService{})

	req := httptest.NewRequest(http.MethodDelete, "/api/exercise/records/not-a-uuid", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleDelete(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExerciseHandler_HandleDelete_ServiceError(t *testing.T) {
	testUserID := "test-user-123"
	validID := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

	mockSvc := &MockExerciseService{
		DeleteExerciseRecordFunc: func(ctx context.Context, userID string, recordID string) error {
			return errors.New("database error")
		},
	}
	handler := NewExerciseHandler(mockSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/exercise/records/"+validID, nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.HandleDelete(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNewExerciseHandler_NilServicePanic(t *testing.T) {
	assert.Panics(t, func() {
		NewExerciseHandler(nil)
	})
}
