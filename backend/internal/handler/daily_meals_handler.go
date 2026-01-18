package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ryosuke-horie/asken/backend/internal/repository"
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
	Date       string                                       `json:"date"`
	Meals      map[string][]repository.HistoryDetail        `json:"meals"`
	DailyTotal repository.DailyTotal                        `json:"daily_total"`
}

// Handle はGET /api/meals/dailyリクエストを処理
func (h *DailyMealsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// date パラメータ取得
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	// リポジトリから取得
	meals, total, err := h.repository.GetDailyMeals(r.Context(), date)
	if err != nil {
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
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
