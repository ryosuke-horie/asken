package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

// --- TestLoadRateLimitConfig ---

func TestLoadRateLimitConfig(t *testing.T) {
	t.Run("環境変数未設定時にデフォルト値が使用されるべき", func(t *testing.T) {
		config := LoadRateLimitConfig()

		assert.Equal(t, float64(10), config.IPRateLimit)
		assert.Equal(t, 20, config.IPBurstSize)
		assert.Equal(t, float64(5), config.UserRateLimit)
		assert.Equal(t, 10, config.UserBurstSize)
		assert.Equal(t, 0.5, config.GeminiRateLimit)
		assert.Equal(t, 3, config.GeminiBurstSize)
		assert.Equal(t, 300*time.Second, config.CleanupInterval)
		assert.Equal(t, 600*time.Second, config.EntryTTL)
	})

	t.Run("環境変数設定時にその値が使用されるべき", func(t *testing.T) {
		t.Setenv("RATE_LIMIT_IP_RPS", "20")
		t.Setenv("RATE_LIMIT_IP_BURST", "40")
		t.Setenv("RATE_LIMIT_USER_RPS", "10")
		t.Setenv("RATE_LIMIT_USER_BURST", "20")
		t.Setenv("RATE_LIMIT_GEMINI_RPS", "1.0")
		t.Setenv("RATE_LIMIT_GEMINI_BURST", "5")
		t.Setenv("RATE_LIMIT_CLEANUP_INTERVAL", "60")
		t.Setenv("RATE_LIMIT_ENTRY_TTL", "120")

		config := LoadRateLimitConfig()

		assert.Equal(t, float64(20), config.IPRateLimit)
		assert.Equal(t, 40, config.IPBurstSize)
		assert.Equal(t, float64(10), config.UserRateLimit)
		assert.Equal(t, 20, config.UserBurstSize)
		assert.Equal(t, 1.0, config.GeminiRateLimit)
		assert.Equal(t, 5, config.GeminiBurstSize)
		assert.Equal(t, 60*time.Second, config.CleanupInterval)
		assert.Equal(t, 120*time.Second, config.EntryTTL)
	})

	t.Run("不正な値の場合にデフォルト値にフォールバックすべき", func(t *testing.T) {
		t.Setenv("RATE_LIMIT_IP_RPS", "invalid")
		t.Setenv("RATE_LIMIT_IP_BURST", "not-a-number")

		config := LoadRateLimitConfig()

		assert.Equal(t, float64(10), config.IPRateLimit)
		assert.Equal(t, 20, config.IPBurstSize)
	})
}

// --- TestInMemoryRateLimiterStore ---

func TestInMemoryRateLimiterStore_GetOrCreate(t *testing.T) {
	t.Run("同じキーで同じLimiterが返されるべき", func(t *testing.T) {
		store := NewInMemoryRateLimiterStore()

		limiter1 := store.GetOrCreate("key1", rate.Limit(10), 20)
		limiter2 := store.GetOrCreate("key1", rate.Limit(10), 20)

		assert.Same(t, limiter1, limiter2)
	})

	t.Run("異なるキーで異なるLimiterが返されるべき", func(t *testing.T) {
		store := NewInMemoryRateLimiterStore()

		limiter1 := store.GetOrCreate("key1", rate.Limit(10), 20)
		limiter2 := store.GetOrCreate("key2", rate.Limit(10), 20)

		assert.NotSame(t, limiter1, limiter2)
	})
}

func TestInMemoryRateLimiterStore_Cleanup(t *testing.T) {
	t.Run("TTL超過したエントリが削除されるべき", func(t *testing.T) {
		store := NewInMemoryRateLimiterStore()

		limiterBefore := store.GetOrCreate("old-key", rate.Limit(10), 20)
		assert.Equal(t, 1, store.Len())

		// 短いTTLで即座にexpire
		time.Sleep(2 * time.Millisecond)
		store.Cleanup(1 * time.Millisecond)

		// 削除されたことを確認
		assert.Equal(t, 0, store.Len())

		// 再度GetOrCreateすると新しいLimiterが作成される
		limiterAfter := store.GetOrCreate("old-key", rate.Limit(10), 20)
		assert.NotSame(t, limiterBefore, limiterAfter)
		assert.Equal(t, 1, store.Len())
	})

	t.Run("TTL内のエントリは保持されるべき", func(t *testing.T) {
		store := NewInMemoryRateLimiterStore()

		store.GetOrCreate("fresh-key", rate.Limit(10), 20)

		store.Cleanup(1 * time.Hour)

		assert.Equal(t, 1, store.Len())
	})

	t.Run("空のストアでもエラーにならないべき", func(t *testing.T) {
		store := NewInMemoryRateLimiterStore()

		assert.NotPanics(t, func() {
			store.Cleanup(1 * time.Millisecond)
		})
	})
}

// --- TestGetClientIP ---

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		xForwardedFor  string
		remoteAddr     string
		expectedIP     string
	}{
		{
			name:          "X-Forwarded-Forが単一IPの場合",
			xForwardedFor: "192.168.1.1",
			remoteAddr:    "10.0.0.1:12345",
			expectedIP:    "192.168.1.1",
		},
		{
			name:          "X-Forwarded-Forが複数IPの場合は最初のIPを返すべき",
			xForwardedFor: "192.168.1.1, 10.0.0.1, 172.16.0.1",
			remoteAddr:    "10.0.0.1:12345",
			expectedIP:    "192.168.1.1",
		},
		{
			name:          "X-Forwarded-Forがない場合はRemoteAddrを使用すべき",
			xForwardedFor: "",
			remoteAddr:    "192.168.1.100:54321",
			expectedIP:    "192.168.1.100",
		},
		{
			name:          "RemoteAddrがhost:port形式でない場合",
			xForwardedFor: "",
			remoteAddr:    "192.168.1.100",
			expectedIP:    "192.168.1.100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}

			ip := getClientIP(req)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}

// --- TestRateLimitMiddleware_LimitByIP ---

func TestRateLimitMiddleware_LimitByIP(t *testing.T) {
	t.Run("制限内のリクエストが通過すべき", func(t *testing.T) {
		config := RateLimitConfig{
			IPRateLimit:     100,
			IPBurstSize:     100,
			CleanupInterval: 1 * time.Hour,
			EntryTTL:        1 * time.Hour,
		}
		rl := NewRateLimitMiddleware(config)
		defer rl.Stop()

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/meals/daily", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		rl.LimitByIP(next).ServeHTTP(w, req)

		assert.True(t, nextCalled)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("制限超過で429を返すべき", func(t *testing.T) {
		config := RateLimitConfig{
			IPRateLimit:     1,
			IPBurstSize:     1,
			CleanupInterval: 1 * time.Hour,
			EntryTTL:        1 * time.Hour,
		}
		rl := NewRateLimitMiddleware(config)
		defer rl.Stop()

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := rl.LimitByIP(next)

		// 1回目: 通過
		req1 := httptest.NewRequest(http.MethodGet, "/api/meals/daily", nil)
		req1.RemoteAddr = "192.168.1.1:12345"
		w1 := httptest.NewRecorder()
		handler.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// 2回目: 制限超過
		req2 := httptest.NewRequest(http.MethodGet, "/api/meals/daily", nil)
		req2.RemoteAddr = "192.168.1.1:12345"
		w2 := httptest.NewRecorder()
		handler.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusTooManyRequests, w2.Code)
		assert.Equal(t, "1", w2.Header().Get("Retry-After"))
	})

	t.Run("/api/healthはレート制限の対象外であるべき", func(t *testing.T) {
		config := RateLimitConfig{
			IPRateLimit:     1,
			IPBurstSize:     1,
			CleanupInterval: 1 * time.Hour,
			EntryTTL:        1 * time.Hour,
		}
		rl := NewRateLimitMiddleware(config)
		defer rl.Stop()

		nextCallCount := 0
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCallCount++
			w.WriteHeader(http.StatusOK)
		})

		handler := rl.LimitByIP(next)

		// ヘルスチェックは何度でも通過
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}

		assert.Equal(t, 5, nextCallCount)
	})

	t.Run("OPTIONSリクエストはレート制限の対象外であるべき", func(t *testing.T) {
		config := RateLimitConfig{
			IPRateLimit:     1,
			IPBurstSize:     1,
			CleanupInterval: 1 * time.Hour,
			EntryTTL:        1 * time.Hour,
		}
		rl := NewRateLimitMiddleware(config)
		defer rl.Stop()

		nextCallCount := 0
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCallCount++
			w.WriteHeader(http.StatusOK)
		})

		handler := rl.LimitByIP(next)

		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodOptions, "/api/meals/daily", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}

		assert.Equal(t, 5, nextCallCount)
	})

	t.Run("異なるIPアドレスは独立してカウントされるべき", func(t *testing.T) {
		config := RateLimitConfig{
			IPRateLimit:     1,
			IPBurstSize:     1,
			CleanupInterval: 1 * time.Hour,
			EntryTTL:        1 * time.Hour,
		}
		rl := NewRateLimitMiddleware(config)
		defer rl.Stop()

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := rl.LimitByIP(next)

		// IP1: 1回目通過
		req1 := httptest.NewRequest(http.MethodGet, "/api/meals/daily", nil)
		req1.RemoteAddr = "192.168.1.1:12345"
		w1 := httptest.NewRecorder()
		handler.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// IP2: 1回目通過（別IP）
		req2 := httptest.NewRequest(http.MethodGet, "/api/meals/daily", nil)
		req2.RemoteAddr = "192.168.1.2:12345"
		w2 := httptest.NewRecorder()
		handler.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)
	})
}

// --- TestIsGeminiEndpoint ---

func TestIsGeminiEndpoint(t *testing.T) {
	rl := &RateLimitMiddleware{}

	tests := []struct {
		name     string
		method   string
		path     string
		expected bool
	}{
		{"POST /api/analyze はGemini対象", http.MethodPost, "/api/analyze", true},
		{"GET /api/analyze はGemini対象外", http.MethodGet, "/api/analyze", false},
		{"GET /api/history/123 はGemini対象", http.MethodGet, "/api/history/123", true},
		{"PUT /api/history/123 はGemini対象", http.MethodPut, "/api/history/123", true},
		{"DELETE /api/history/123 はGemini対象外", http.MethodDelete, "/api/history/123", false},
		{"GET /api/history はGemini対象外", http.MethodGet, "/api/history", false},
		{"GET /api/meals/daily はGemini対象外", http.MethodGet, "/api/meals/daily", false},
		{"POST /api/meals/skip はGemini対象外", http.MethodPost, "/api/meals/skip", false},
		{"POST /api/upload-image はGemini対象外", http.MethodPost, "/api/upload-image", false},
		{"GET /api/weight/records はGemini対象外", http.MethodGet, "/api/weight/records", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			result := rl.isGeminiEndpoint(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// --- TestRateLimitMiddleware_LimitByUser ---

func TestRateLimitMiddleware_LimitByUser(t *testing.T) {
	t.Run("通常エンドポイントで制限内のリクエストが通過すべき", func(t *testing.T) {
		config := RateLimitConfig{
			UserRateLimit:   100,
			UserBurstSize:   100,
			GeminiRateLimit: 100,
			GeminiBurstSize: 100,
			CleanupInterval: 1 * time.Hour,
			EntryTTL:        1 * time.Hour,
		}
		rl := NewRateLimitMiddleware(config)
		defer rl.Stop()

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/meals/daily", nil)
		ctx := SetFirebaseUIDToContext(req.Context(), "user-123")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		rl.LimitByUser(next).ServeHTTP(w, req)

		assert.True(t, nextCalled)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("通常エンドポイントで制限超過時に429を返すべき", func(t *testing.T) {
		config := RateLimitConfig{
			UserRateLimit:   1,
			UserBurstSize:   1,
			GeminiRateLimit: 100,
			GeminiBurstSize: 100,
			CleanupInterval: 1 * time.Hour,
			EntryTTL:        1 * time.Hour,
		}
		rl := NewRateLimitMiddleware(config)
		defer rl.Stop()

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := rl.LimitByUser(next)

		// 1回目: 通過
		req1 := httptest.NewRequest(http.MethodGet, "/api/meals/daily", nil)
		ctx1 := SetFirebaseUIDToContext(req1.Context(), "user-123")
		req1 = req1.WithContext(ctx1)
		w1 := httptest.NewRecorder()
		handler.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// 2回目: 制限超過
		req2 := httptest.NewRequest(http.MethodGet, "/api/meals/daily", nil)
		ctx2 := SetFirebaseUIDToContext(req2.Context(), "user-123")
		req2 = req2.WithContext(ctx2)
		w2 := httptest.NewRecorder()
		handler.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusTooManyRequests, w2.Code)
		assert.Equal(t, "1", w2.Header().Get("Retry-After"))
	})

	t.Run("Geminiエンドポイントでより厳しい制限が適用されるべき", func(t *testing.T) {
		config := RateLimitConfig{
			UserRateLimit:   100,
			UserBurstSize:   100,
			GeminiRateLimit: 1,
			GeminiBurstSize: 1,
			CleanupInterval: 1 * time.Hour,
			EntryTTL:        1 * time.Hour,
		}
		rl := NewRateLimitMiddleware(config)
		defer rl.Stop()

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := rl.LimitByUser(next)

		// Geminiエンドポイント1回目: 通過
		req1 := httptest.NewRequest(http.MethodPost, "/api/analyze", nil)
		ctx1 := SetFirebaseUIDToContext(req1.Context(), "user-123")
		req1 = req1.WithContext(ctx1)
		w1 := httptest.NewRecorder()
		handler.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Geminiエンドポイント2回目: 制限超過
		req2 := httptest.NewRequest(http.MethodPost, "/api/analyze", nil)
		ctx2 := SetFirebaseUIDToContext(req2.Context(), "user-123")
		req2 = req2.WithContext(ctx2)
		w2 := httptest.NewRecorder()
		handler.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusTooManyRequests, w2.Code)
		assert.Equal(t, "2", w2.Header().Get("Retry-After"))

		// 通常エンドポイントはまだ通過するべき
		req3 := httptest.NewRequest(http.MethodGet, "/api/meals/daily", nil)
		ctx3 := SetFirebaseUIDToContext(req3.Context(), "user-123")
		req3 = req3.WithContext(ctx3)
		w3 := httptest.NewRecorder()
		handler.ServeHTTP(w3, req3)
		assert.Equal(t, http.StatusOK, w3.Code)
	})

	t.Run("UID未設定時にリクエストが通過すべき", func(t *testing.T) {
		config := RateLimitConfig{
			UserRateLimit:   1,
			UserBurstSize:   1,
			GeminiRateLimit: 1,
			GeminiBurstSize: 1,
			CleanupInterval: 1 * time.Hour,
			EntryTTL:        1 * time.Hour,
		}
		rl := NewRateLimitMiddleware(config)
		defer rl.Stop()

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/meals/daily", nil)
		w := httptest.NewRecorder()

		rl.LimitByUser(next).ServeHTTP(w, req)

		assert.True(t, nextCalled)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("異なるユーザーは独立してカウントされるべき", func(t *testing.T) {
		config := RateLimitConfig{
			UserRateLimit:   1,
			UserBurstSize:   1,
			GeminiRateLimit: 100,
			GeminiBurstSize: 100,
			CleanupInterval: 1 * time.Hour,
			EntryTTL:        1 * time.Hour,
		}
		rl := NewRateLimitMiddleware(config)
		defer rl.Stop()

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := rl.LimitByUser(next)

		// ユーザー1: 通過
		req1 := httptest.NewRequest(http.MethodGet, "/api/meals/daily", nil)
		ctx1 := SetFirebaseUIDToContext(req1.Context(), "user-1")
		req1 = req1.WithContext(ctx1)
		w1 := httptest.NewRecorder()
		handler.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// ユーザー2: 通過（別ユーザー）
		req2 := httptest.NewRequest(http.MethodGet, "/api/meals/daily", nil)
		ctx2 := SetFirebaseUIDToContext(req2.Context(), "user-2")
		req2 = req2.WithContext(ctx2)
		w2 := httptest.NewRecorder()
		handler.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)
	})
}

// --- TestRateLimitMiddleware_Stop ---

func TestRateLimitMiddleware_Stop(t *testing.T) {
	t.Run("Stop呼び出し後にパニックが発生しないべき", func(t *testing.T) {
		config := RateLimitConfig{
			IPRateLimit:     10,
			IPBurstSize:     20,
			CleanupInterval: 10 * time.Millisecond,
			EntryTTL:        1 * time.Hour,
		}
		rl := NewRateLimitMiddleware(config)

		// クリーンアップが少なくとも1回実行される時間を待つ
		time.Sleep(20 * time.Millisecond)

		require.NotPanics(t, func() {
			rl.Stop()
		})
	})

	t.Run("Stop後もLimitByIPが正常に動作すべき", func(t *testing.T) {
		config := RateLimitConfig{
			IPRateLimit:     100,
			IPBurstSize:     100,
			CleanupInterval: 1 * time.Hour,
			EntryTTL:        1 * time.Hour,
		}
		rl := NewRateLimitMiddleware(config)
		rl.Stop()

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/meals/daily", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		rl.LimitByIP(next).ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

