package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
)

// DailyMealsHandler は日次食事データ取得エンドポイントのハンドラー
type DailyMealsHandler struct {
	repository repository.AnalysisRepository
}

// NewDailyMealsHandler は新しいDailyMealsHandlerを作成
func NewDailyMealsHandler(repository repository.AnalysisRepository) *DailyMealsHandler {
	return &DailyMealsHandler{repository: repository}
}

// DailyMealsResponse は日次食事データのレスポンス
type DailyMealsResponse struct {
	Date       string                                `json:"date"`
	Meals      map[string][]repository.HistoryDetail `json:"meals"`
	DailyTotal repository.DailyTotal                 `json:"daily_total"`
}

// Handle はGET /api/meals/dailyリクエストを処理
func (h *DailyMealsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// contextからユーザーIDを取得
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		log.Printf("Authentication failed for %s: %s %s - no Firebase UID in context", r.RemoteAddr, r.Method, r.URL.Path)
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	// date, tz パラメータ取得
	date := r.URL.Query().Get("date")
	tz := r.URL.Query().Get("tz")

	// tzが未指定の場合はUTCとして処理（後方互換性）
	if tz == "" {
		tz = "UTC"
	}

	// 日付が未指定の場合、指定タイムゾーンでの現在日付を使用
	if date == "" {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			loc = time.UTC
		}
		date = time.Now().In(loc).Format("2006-01-02")
	}

	log.Printf("Getting daily meals for userID: %s, date: %s, tz: %s", userID, date, tz)

	// リポジトリから取得（userIDでスコープ）
	meals, total, err := h.repository.GetDailyMeals(r.Context(), userID, date, tz)
	if err != nil {
		log.Printf("Error getting daily meals: %v", err)
		http.Error(w, "Failed to get daily meals", http.StatusInternalServerError)
		return
	}

	response := DailyMealsResponse{
		Date:       date,
		Meals:      meals,
		DailyTotal: total,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
