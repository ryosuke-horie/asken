package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestHistoryDeleteHandler_Handle_Success(t *testing.T) {
	historyID := uuid.New()

	mockRepo := &MockAnalysisRepository{
		DeleteHistoryFunc: func(ctx context.Context, id uuid.UUID) error {
			assert.Equal(t, historyID, id)
			return nil
		},
	}

	handler := NewHistoryDeleteHandler(mockRepo)

	req := httptest.NewRequest(http.MethodDelete, "/api/history/"+historyID.String(), nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestHistoryDeleteHandler_Handle_NotFound(t *testing.T) {
	historyID := uuid.New()

	mockRepo := &MockAnalysisRepository{
		DeleteHistoryFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("履歴が見つかりません")
		},
	}

	handler := NewHistoryDeleteHandler(mockRepo)

	req := httptest.NewRequest(http.MethodDelete, "/api/history/"+historyID.String(), nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHistoryDeleteHandler_Handle_RepositoryError(t *testing.T) {
	historyID := uuid.New()

	mockRepo := &MockAnalysisRepository{
		DeleteHistoryFunc: func(ctx context.Context, id uuid.UUID) error {
			return errors.New("database error")
		},
	}

	handler := NewHistoryDeleteHandler(mockRepo)

	req := httptest.NewRequest(http.MethodDelete, "/api/history/"+historyID.String(), nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHistoryDeleteHandler_Handle_InvalidUUID(t *testing.T) {
	mockRepo := &MockAnalysisRepository{}
	handler := NewHistoryDeleteHandler(mockRepo)

	req := httptest.NewRequest(http.MethodDelete, "/api/history/invalid-uuid", nil)
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
	mockRepo := &MockAnalysisRepository{}
	handler := NewHistoryDeleteHandler(mockRepo)

	req := httptest.NewRequest(http.MethodDelete, "/api/history", nil)
	w := httptest.NewRecorder()

	handler.Handle(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
