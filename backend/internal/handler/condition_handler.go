package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
)

type ConditionHandler struct {
	repository repository.ConditionRepository
}

func NewConditionHandler(repository repository.ConditionRepository) *ConditionHandler {
	return &ConditionHandler{repository: repository}
}

type CreateConditionRecordRequest struct {
	Condition  int    `json:"condition"`
	Fatigue    int    `json:"fatigue"`
	RecordedAt string `json:"recorded_at"`
}

func (h *ConditionHandler) HandleCreateRecord(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	var req CreateConditionRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエスト形式です", http.StatusBadRequest)
		return
	}

	if req.Condition < 1 || req.Condition > 3 {
		http.Error(w, "体調は1〜3の範囲で入力してください", http.StatusBadRequest)
		return
	}

	if req.Fatigue < 1 || req.Fatigue > 3 {
		http.Error(w, "疲労度は1〜3の範囲で入力してください", http.StatusBadRequest)
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

	record, err := h.repository.CreateOrUpdateRecord(r.Context(), userID, req.Condition, req.Fatigue, req.RecordedAt)
	if err != nil {
		log.Printf("体調記録の作成/更新に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "体調の記録に失敗しました", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(record)
	if err != nil {
		log.Printf("JSONエンコードに失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(jsonData); err != nil {
		log.Printf("レスポンス書き込みに失敗 (user_id=%s): %v", userID, err)
	}
}

func (h *ConditionHandler) HandleGetRecord(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, "日付パラメータは必須です", http.StatusBadRequest)
		return
	}

	if _, err := time.Parse("2006-01-02", date); err != nil {
		http.Error(w, "日付形式が不正です（YYYY-MM-DD形式で入力してください）", http.StatusBadRequest)
		return
	}

	record, err := h.repository.GetRecordByDate(r.Context(), userID, date)
	if err != nil {
		log.Printf("体調記録の取得に失敗 (user_id=%s, date=%s): %v", userID, date, err)
		http.Error(w, "体調記録の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if record == nil {
		if _, err := w.Write([]byte("null")); err != nil {
			log.Printf("レスポンス書き込みに失敗 (user_id=%s, date=%s): %v", userID, date, err)
		}
		return
	}

	jsonData, err := json.Marshal(record)
	if err != nil {
		log.Printf("JSONエンコードに失敗 (user_id=%s, date=%s): %v", userID, date, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(jsonData); err != nil {
		log.Printf("レスポンス書き込みに失敗 (user_id=%s, date=%s): %v", userID, date, err)
	}
}
