package middleware

import (
	"os"
	"strconv"
	"time"
)

// RateLimitConfig はレート制限の設定を保持する
type RateLimitConfig struct {
	// IP単位のレート制限（認証前に適用）
	IPRateLimit float64
	IPBurstSize int

	// ユーザー単位のレート制限（認証後に適用）
	UserRateLimit float64
	UserBurstSize int

	// Gemini API関連エンドポイントの厳しい制限
	GeminiRateLimit float64
	GeminiBurstSize int

	// クリーンアップ間隔
	CleanupInterval time.Duration
	// エントリの有効期限（最終アクセスからの経過時間）
	EntryTTL time.Duration
}

// LoadRateLimitConfig は環境変数からレート制限設定をロードする
func LoadRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		IPRateLimit:     getEnvFloat("RATE_LIMIT_IP_RPS", 10),
		IPBurstSize:     getEnvInt("RATE_LIMIT_IP_BURST", 20),
		UserRateLimit:   getEnvFloat("RATE_LIMIT_USER_RPS", 5),
		UserBurstSize:   getEnvInt("RATE_LIMIT_USER_BURST", 10),
		GeminiRateLimit: getEnvFloat("RATE_LIMIT_GEMINI_RPS", 0.5),
		GeminiBurstSize: getEnvInt("RATE_LIMIT_GEMINI_BURST", 3),
		CleanupInterval: time.Duration(getEnvInt("RATE_LIMIT_CLEANUP_INTERVAL", 300)) * time.Second,
		EntryTTL:        time.Duration(getEnvInt("RATE_LIMIT_ENTRY_TTL", 600)) * time.Second,
	}
}

func getEnvFloat(key string, defaultVal float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	parsed, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return defaultVal
	}
	return parsed
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return parsed
}
