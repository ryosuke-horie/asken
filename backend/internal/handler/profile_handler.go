package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
)

// 許可された競技種別
var allowedSportTypes = map[string]bool{
	"柔術":       true,
	"キックボクシング": true,
	"MMA":      true,
	"ボクシング":    true,
	"レスリング":    true,
	"ムエタイ":     true,
	"空手":       true,
	"テコンドー":    true,
	"総合格闘技":    true,
	"その他":      true,
}

// 許可されたトレーニング目標
var allowedTrainingGoals = map[string]bool{
	"減量":     true,
	"スタミナ強化": true,
	"筋力強化":   true,
	"技術向上":   true,
	"柔軟性向上":  true,
	"維持":     true,
}

type ProfileHandler struct {
	repository repository.ProfileRepository
}

func NewProfileHandler(repository repository.ProfileRepository) *ProfileHandler {
	return &ProfileHandler{repository: repository}
}

type UpdateProfileRequest struct {
	SportType     *string  `json:"sport_type"`
	TrainingGoals []string `json:"training_goals"`
	WeightClass   *int     `json:"weight_class"`
}

func (h *ProfileHandler) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	profile, err := h.repository.GetByUserID(r.Context(), userID)
	if err != nil {
		log.Printf("プロフィールの取得に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "プロフィールの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if profile == nil {
		if _, err := w.Write([]byte("null")); err != nil {
			log.Printf("レスポンス書き込みに失敗 (user_id=%s): %v", userID, err)
		}
		return
	}

	jsonData, err := json.Marshal(profile)
	if err != nil {
		log.Printf("JSONエンコードに失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(jsonData); err != nil {
		log.Printf("レスポンス書き込みに失敗 (user_id=%s): %v", userID, err)
	}
}

func (h *ProfileHandler) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエスト形式です", http.StatusBadRequest)
		return
	}

	if req.WeightClass != nil && (*req.WeightClass < 1 || *req.WeightClass > 200) {
		http.Error(w, "体重階級は1〜200kgの範囲で入力してください", http.StatusBadRequest)
		return
	}

	// SportTypeのバリデーション
	if req.SportType != nil && *req.SportType != "" && !allowedSportTypes[*req.SportType] {
		http.Error(w, "無効な競技種別です", http.StatusBadRequest)
		return
	}

	// TrainingGoalsのバリデーション
	for _, goal := range req.TrainingGoals {
		if !allowedTrainingGoals[goal] {
			http.Error(w, "無効なトレーニング目標が含まれています", http.StatusBadRequest)
			return
		}
	}

	profile := &repository.UserProfile{
		UserID:        userID,
		SportType:     req.SportType,
		TrainingGoals: req.TrainingGoals,
		WeightClass:   req.WeightClass,
	}

	result, err := h.repository.CreateOrUpdate(r.Context(), profile)
	if err != nil {
		log.Printf("プロフィールの保存に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "プロフィールの保存に失敗しました", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		log.Printf("JSONエンコードに失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(jsonData); err != nil {
		log.Printf("レスポンス書き込みに失敗 (user_id=%s): %v", userID, err)
	}
}
