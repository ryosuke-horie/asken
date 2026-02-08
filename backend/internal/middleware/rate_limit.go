package middleware

import (
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type limiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// InMemoryRateLimiterStore はインメモリのレートリミッターストア
type InMemoryRateLimiterStore struct {
	mu       sync.Mutex
	limiters map[string]*limiterEntry
}

// NewInMemoryRateLimiterStore は新しいインメモリストアを作成する
func NewInMemoryRateLimiterStore() *InMemoryRateLimiterStore {
	return &InMemoryRateLimiterStore{
		limiters: make(map[string]*limiterEntry),
	}
}

// GetOrCreate は指定キーのLimiterを取得する。存在しない場合は新規作成する
func (s *InMemoryRateLimiterStore) GetOrCreate(key string, limit rate.Limit, burst int) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry, ok := s.limiters[key]; ok {
		entry.lastAccess = time.Now()
		return entry.limiter
	}

	limiter := rate.NewLimiter(limit, burst)
	s.limiters[key] = &limiterEntry{
		limiter:    limiter,
		lastAccess: time.Now(),
	}
	return limiter
}

// Cleanup はTTL超過したエントリを削除する
func (s *InMemoryRateLimiterStore) Cleanup(ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, entry := range s.limiters {
		if now.Sub(entry.lastAccess) > ttl {
			delete(s.limiters, key)
		}
	}
}

// Len はストア内のエントリ数を返す（テスト用）
func (s *InMemoryRateLimiterStore) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.limiters)
}

// RateLimitMiddleware はレート制限ミドルウェア
type RateLimitMiddleware struct {
	ipStore     *InMemoryRateLimiterStore
	userStore   *InMemoryRateLimiterStore
	config      RateLimitConfig
	stopCleanup chan struct{}
	stopOnce    sync.Once
}

// NewRateLimitMiddleware は新しいレート制限ミドルウェアを作成し、クリーンアップゴルーチンを起動する
func NewRateLimitMiddleware(config RateLimitConfig) *RateLimitMiddleware {
	m := &RateLimitMiddleware{
		ipStore:     NewInMemoryRateLimiterStore(),
		userStore:   NewInMemoryRateLimiterStore(),
		config:      config,
		stopCleanup: make(chan struct{}),
	}
	go m.startCleanup()
	return m
}

// Stop はクリーンアップゴルーチンを停止する（複数回呼び出し可能）
func (m *RateLimitMiddleware) Stop() {
	m.stopOnce.Do(func() {
		close(m.stopCleanup)
	})
}

func (m *RateLimitMiddleware) startCleanup() {
	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.ipStore.Cleanup(m.config.EntryTTL)
			m.userStore.Cleanup(m.config.EntryTTL)
		case <-m.stopCleanup:
			return
		}
	}
}

// LimitByIP はIPアドレス単位のレート制限を適用する（認証前）
func (m *RateLimitMiddleware) LimitByIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/health" || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		ip := getClientIP(r)
		limiter := m.ipStore.GetOrCreate(ip, rate.Limit(m.config.IPRateLimit), m.config.IPBurstSize)

		if !limiter.Allow() {
			log.Printf("Rate limit exceeded: ip=%s path=%s", ip, r.URL.Path)
			w.Header().Set("Retry-After", "1")
			http.Error(w, "リクエスト数が制限を超えました。しばらくしてから再試行してください", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// LimitByUser はFirebase UID単位のレート制限を適用する（認証後）
func (m *RateLimitMiddleware) LimitByUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := GetFirebaseUIDFromContext(r.Context())
		if uid == "" {
			next.ServeHTTP(w, r)
			return
		}

		if m.isGeminiEndpoint(r) {
			key := "gemini:" + uid
			limiter := m.userStore.GetOrCreate(key, rate.Limit(m.config.GeminiRateLimit), m.config.GeminiBurstSize)
			if !limiter.Allow() {
				log.Printf("Gemini rate limit exceeded: uid=%s path=%s", uid, r.URL.Path)
				w.Header().Set("Retry-After", "2")
				http.Error(w, "分析リクエスト数が制限を超えました。しばらくしてから再試行してください", http.StatusTooManyRequests)
				return
			}
		}

		key := "user:" + uid
		limiter := m.userStore.GetOrCreate(key, rate.Limit(m.config.UserRateLimit), m.config.UserBurstSize)
		if !limiter.Allow() {
			log.Printf("User rate limit exceeded: uid=%s path=%s", uid, r.URL.Path)
			w.Header().Set("Retry-After", "1")
			http.Error(w, "リクエスト数が制限を超えました。しばらくしてから再試行してください", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isGeminiEndpoint はGemini API呼び出しを伴うエンドポイントかどうかを判定する
func (m *RateLimitMiddleware) isGeminiEndpoint(r *http.Request) bool {
	path := r.URL.Path
	method := r.Method

	if path == "/api/analyze" && method == http.MethodPost {
		return true
	}

	if strings.HasPrefix(path, "/api/history/") && (method == http.MethodGet || method == http.MethodPut) {
		return true
	}

	return false
}

// getClientIP はリクエストからクライアントIPアドレスを取得する
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
