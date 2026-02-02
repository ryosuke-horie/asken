package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetFirebaseUIDToContext(t *testing.T) {
	t.Run("Firebase UIDをコンテキストに設定して取得できるべき", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		ctx := SetFirebaseUIDToContext(req.Context(), "test-firebase-uid")
		req = req.WithContext(ctx)

		uid := GetFirebaseUIDFromContext(req.Context())
		assert.Equal(t, "test-firebase-uid", uid)
	})

	t.Run("空のUIDを設定できるべき", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		ctx := SetFirebaseUIDToContext(req.Context(), "")
		req = req.WithContext(ctx)

		uid := GetFirebaseUIDFromContext(req.Context())
		assert.Equal(t, "", uid)
	})
}

func TestGetFirebaseUIDFromContext(t *testing.T) {
	t.Run("UIDが設定されていない場合は空文字を返すべき", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)

		uid := GetFirebaseUIDFromContext(req.Context())
		assert.Equal(t, "", uid)
	})
}
