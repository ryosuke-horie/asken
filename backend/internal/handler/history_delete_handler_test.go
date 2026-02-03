package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestHistoryDeleteHandler_Handle_Success(t *testing.T) {
	historyID := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		DeleteHistoryFunc: func(ctx context.Context, userID string, id uuid.UUID) error {
			assert.Equal(t, testUserID, userID)
			assert.Equal(t, historyID, id)
			return nil
		},
	}

	handler := NewHistoryDeleteHandler(mockRepo)

	req := httptest.NewRequest(http.MethodDelete, "/api/history/"+historyID.String(), nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestHistoryDeleteHandler_Handle_NotFound(t *testing.T) {
	historyID := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		DeleteHistoryFunc: func(ctx context.Context, userID string, id uuid.UUID) error {
			return fmt.Errorf("履歴が見つかりません: %s: %w", id, repository.ErrNotFound)
		},
	}

	handler := NewHistoryDeleteHandler(mockRepo)

	req := httptest.NewRequest(http.MethodDelete, "/api/history/"+historyID.String(), nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHistoryDeleteHandler_Handle_RepositoryError(t *testing.T) {
	historyID := uuid.New()
	testUserID := "test-user-123"

	mockRepo := &MockAnalysisRepository{
		DeleteHistoryFunc: func(ctx context.Context, userID string, id uuid.UUID) error {
			return errors.New("database error")
		},
	}

	handler := NewHistoryDeleteHandler(mockRepo)

	req := httptest.NewRequest(http.MethodDelete, "/api/history/"+historyID.String(), nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHistoryDeleteHandler_Handle_InvalidUUID(t *testing.T) {
	testUserID := "test-user-123"
	mockRepo := &MockAnalysisRepository{}
	handler := NewHistoryDeleteHandler(mockRepo)

	req := httptest.NewRequest(http.MethodDelete, "/api/history/invalid-uuid", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHistoryDeleteHandler_Handle_MethodNotAllowed(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewHistoryDeleteHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/history/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHistoryDeleteHandler_Handle_InvalidURL(t *testing.T) {
	testUserID := "test-user-123"
	mockRepo := &MockAnalysisRepository{}
	handler := NewHistoryDeleteHandler(mockRepo)

	req := httptest.NewRequest(http.MethodDelete, "/api/history", nil)
	ctx := middleware.SetFirebaseUIDToContext(req.Context(), testUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHistoryDeleteHandler_Handle_Unauthorized(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewHistoryDeleteHandler(mockRepo)

	// コンテキストにユーザーIDを設定しない
	req := httptest.NewRequest(http.MethodDelete, "/api/history/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
