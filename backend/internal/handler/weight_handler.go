package handler

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
)

type WeightHandler struct {
	repository repository.WeightRepository
}

func NewWeightHandler(repository repository.WeightRepository) *WeightHandler {
	return &WeightHandler{repository: repository}
}

type CreateWeightRecordRequest struct {
	Weight     float64 `json:"weight"`
	RecordedAt string  `json:"recorded_at"`
}

type WeightRecordsResponse struct {
	Records []*repository.WeightRecord `json:"records"`
	Latest  *repository.WeightRecord   `json:"latest"`
	Stats   *repository.WeightStats    `json:"stats"`
}

type WeightGoalResponse struct {
	TargetWeight  float64 `json:"target_weight"`
	TargetDate    string  `json:"target_date"`
	DaysRemaining int     `json:"days_remaining"`
	WeightToLose  float64 `json:"weight_to_lose"`
}

type UpdateWeightGoalRequest struct {
	TargetWeight float64 `json:"target_weight"`
	TargetDate   string  `json:"target_date"`
}

func (h *WeightHandler) HandleCreateRecord(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	var req CreateWeightRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエスト形式です", http.StatusBadRequest)
		return
	}

	if req.Weight < 0.1 || req.Weight > 999.9 {
		http.Error(w, "体重は0.1〜999.9kgの範囲で入力してください", http.StatusBadRequest)
		return
	}

	if req.RecordedAt == "" {
		req.RecordedAt = time.Now().Format("2006-01-02")
	} else {
		if _, err := time.Parse("2006-01-02", req.RecordedAt); err != nil {
			http.Error(w, "日付形式が不正です（YYYY-MM-DD形式で入力してください）", http.StatusBadRequest)
			return
		}
	}

	record, err := h.repository.CreateOrUpdateRecord(r.Context(), userID, req.Weight, req.RecordedAt)
	if err != nil {
		log.Printf("体重記録の作成/更新に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "体重の記録に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(record); err != nil {
		http.Error(w, "レスポンスのエンコードに失敗しました", http.StatusInternalServerError)
		return
	}
}

func (h *WeightHandler) HandleGetRecords(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "week"
	}

	endDate := time.Now()
	var startDate time.Time

	switch period {
	case "week":
		startDate = endDate.AddDate(0, 0, -7)
	case "month":
		startDate = endDate.AddDate(0, -1, 0)
	case "3months":
		startDate = endDate.AddDate(0, -3, 0)
	default:
		startDate = endDate.AddDate(0, 0, -7)
	}

	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	records, err := h.repository.GetRecordsByPeriod(r.Context(), userID, startDateStr, endDateStr)
	if err != nil {
		log.Printf("期間内の体重記録取得に失敗 (user_id=%s, period=%s-%s): %v", userID, startDateStr, endDateStr, err)
		http.Error(w, "体重記録の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	latest, err := h.repository.GetLatestRecord(r.Context(), userID)
	if err != nil {
		log.Printf("最新体重記録の取得に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "最新の体重記録の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	stats, err := h.repository.GetStatsByPeriod(r.Context(), userID, startDateStr, endDateStr)
	if err != nil {
		log.Printf("体重統計の取得に失敗 (user_id=%s, period=%s-%s): %v", userID, startDateStr, endDateStr, err)
		http.Error(w, "体重統計の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	if records == nil {
		records = []*repository.WeightRecord{}
	}

	response := WeightRecordsResponse{
		Records: records,
		Latest:  latest,
		Stats:   stats,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "レスポンスのエンコードに失敗しました", http.StatusInternalServerError)
		return
	}
}

func (h *WeightHandler) HandleGetGoal(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	goal, err := h.repository.GetGoal(r.Context(), userID)
	if err != nil {
		log.Printf("目標体重の取得に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "目標体重の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	if goal == nil {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte("null")); err != nil {
			log.Printf("レスポンス書き込みに失敗: %v", err)
		}
		return
	}

	latest, err := h.repository.GetLatestRecord(r.Context(), userID)
	if err != nil {
		log.Printf("最新体重記録の取得に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "最新の体重記録の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	targetDate, err := time.Parse("2006-01-02", goal.TargetDate)
	if err != nil {
		log.Printf("目標日のパースに失敗 (user_id=%s, target_date=%s): %v", userID, goal.TargetDate, err)
		http.Error(w, "目標日のデータが不正です。目標を再設定してください", http.StatusInternalServerError)
		return
	}
	daysRemaining := int(math.Ceil(time.Until(targetDate).Hours() / 24))
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	var weightToLose float64
	if latest != nil {
		weightToLose = math.Round((latest.Weight-goal.TargetWeight)*10) / 10
	}

	response := WeightGoalResponse{
		TargetWeight:  goal.TargetWeight,
		TargetDate:    goal.TargetDate,
		DaysRemaining: daysRemaining,
		WeightToLose:  weightToLose,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "レスポンスのエンコードに失敗しました", http.StatusInternalServerError)
		return
	}
}

func (h *WeightHandler) HandleUpdateGoal(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	var req UpdateWeightGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエスト形式です", http.StatusBadRequest)
		return
	}

	if req.TargetWeight < 0.1 || req.TargetWeight > 999.9 {
		http.Error(w, "目標体重は0.1〜999.9kgの範囲で入力してください", http.StatusBadRequest)
		return
	}

	if req.TargetDate == "" {
		http.Error(w, "目標日は必須です", http.StatusBadRequest)
		return
	}

	if _, err := time.Parse("2006-01-02", req.TargetDate); err != nil {
		http.Error(w, "目標日の形式が不正です（YYYY-MM-DD形式で入力してください）", http.StatusBadRequest)
		return
	}

	goal, err := h.repository.CreateOrUpdateGoal(r.Context(), userID, req.TargetWeight, req.TargetDate)
	if err != nil {
		log.Printf("目標体重の保存に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "目標体重の保存に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(goal); err != nil {
		http.Error(w, "レスポンスのエンコードに失敗しました", http.StatusInternalServerError)
		return
	}
}
