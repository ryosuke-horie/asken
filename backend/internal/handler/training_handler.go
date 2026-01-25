package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
)

// TrainingHandler はトレーニング関連のハンドラー
type TrainingHandler struct {
	repository          repository.TrainingRepository
	menuSuggester       *gemini.MenuSuggester
	equipmentNormalizer *gemini.EquipmentNormalizer
}

// NewTrainingHandler は新しいTrainingHandlerを作成
func NewTrainingHandler(repository repository.TrainingRepository, menuSuggester *gemini.MenuSuggester, equipmentNormalizer *gemini.EquipmentNormalizer) *TrainingHandler {
	return &TrainingHandler{
		repository:          repository,
		menuSuggester:       menuSuggester,
		equipmentNormalizer: equipmentNormalizer,
	}
}

// Location関連のリクエスト/レスポンス構造体

type CreateLocationRequest struct {
	Name string `json:"name"`
}

type UpdateLocationRequest struct {
	Name string `json:"name"`
}

// Equipment関連のリクエスト/レスポンス構造体

type CreateEquipmentRequest struct {
	Name         string  `json:"name"`
	OriginalName *string `json:"original_name,omitempty"`
}

type UpdateEquipmentRequest struct {
	Name         string  `json:"name"`
	OriginalName *string `json:"original_name,omitempty"`
}

type NormalizeEquipmentRequest struct {
	Names []string `json:"names"`
}

type NormalizeEquipmentResponse struct {
	NormalizedNames []gemini.NormalizedEquipment `json:"normalized_names"`
}

// Menu関連のリクエスト/レスポンス構造体

type CreateMenuRequest struct {
	Name string `json:"name"`
}

// Record関連のリクエスト/レスポンス構造体

type CreateRecordRequest struct {
	RecordedAt   string   `json:"recorded_at"`
	LocationID   *string  `json:"location_id,omitempty"`
	Completed    bool     `json:"completed"`
	Duration     *int     `json:"duration,omitempty"`
	Intensity    *int     `json:"intensity,omitempty"`
	Satisfaction *int     `json:"satisfaction,omitempty"`
	Notes        *string  `json:"notes,omitempty"`
	MenuIDs      []string `json:"menu_ids,omitempty"`
}

type UpdateRecordRequest struct {
	RecordedAt   string   `json:"recorded_at"`
	LocationID   *string  `json:"location_id,omitempty"`
	Completed    bool     `json:"completed"`
	Duration     *int     `json:"duration,omitempty"`
	Intensity    *int     `json:"intensity,omitempty"`
	Satisfaction *int     `json:"satisfaction,omitempty"`
	Notes        *string  `json:"notes,omitempty"`
	MenuIDs      []string `json:"menu_ids,omitempty"`
}

// 後方互換性のためのリクエスト構造体
type UpsertRecordRequest struct {
	RecordedAt string  `json:"recorded_at"`
	LocationID *string `json:"location_id,omitempty"`
	Completed  bool    `json:"completed"`
}

// Gemini連携リクエスト/レスポンス構造体

type SuggestMenuRequest struct {
	Equipment []string `json:"equipment"`
	Duration  int      `json:"duration"` // 分
	Goals     []string `json:"goals,omitempty"`
}

type SuggestMenuResponse struct {
	Menu []gemini.MenuItem `json:"menu"`
}

// Location ハンドラー

func (h *TrainingHandler) HandleListLocations(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	locations, err := h.repository.GetAllLocations(r.Context(), userID)
	if err != nil {
		log.Printf("トレーニング場所の取得に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "トレーニング場所の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	if locations == nil {
		locations = []*repository.TrainingLocation{}
	}

	data, err := json.Marshal(locations)
	if err != nil {
		log.Printf("レスポンスのエンコードに失敗: %v", err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("レスポンス書き込みに失敗: %v", err)
		return
	}
}

func (h *TrainingHandler) HandleCreateLocation(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	var req CreateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエスト形式です", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "名前は必須です", http.StatusBadRequest)
		return
	}
	if len(req.Name) > 100 {
		http.Error(w, "名前は100文字以内で入力してください", http.StatusBadRequest)
		return
	}

	location := &repository.TrainingLocation{
		UserID: userID,
		Name:   req.Name,
	}

	created, err := h.repository.CreateLocation(r.Context(), location)
	if err != nil {
		log.Printf("トレーニング場所の作成に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "トレーニング場所の作成に失敗しました", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(created)
	if err != nil {
		log.Printf("レスポンスのエンコードに失敗: %v", err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(data); err != nil {
		log.Printf("レスポンス書き込みに失敗: %v", err)
		return
	}
}

func (h *TrainingHandler) HandleUpdateLocation(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	idStr := extractTrainingIDFromPath(r.URL.Path, "/api/training/locations/")
	if idStr == "" {
		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}

	locationID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "無効なID形式です", http.StatusBadRequest)
		return
	}

	var req UpdateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエスト形式です", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "名前は必須です", http.StatusBadRequest)
		return
	}
	if len(req.Name) > 100 {
		http.Error(w, "名前は100文字以内で入力してください", http.StatusBadRequest)
		return
	}

	location := &repository.TrainingLocation{
		ID:     locationID,
		UserID: userID,
		Name:   req.Name,
	}

	updated, err := h.repository.UpdateLocation(r.Context(), location)
	if err != nil {
		log.Printf("トレーニング場所の更新に失敗 (user_id=%s, id=%s): %v", userID, locationID, err)
		if errors.Is(err, repository.ErrTrainingNotFound) {
			http.Error(w, "トレーニング場所が見つかりません", http.StatusNotFound)
			return
		}
		http.Error(w, "トレーニング場所の更新に失敗しました", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(updated)
	if err != nil {
		log.Printf("レスポンスのエンコードに失敗: %v", err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("レスポンス書き込みに失敗: %v", err)
		return
	}
}

func (h *TrainingHandler) HandleDeleteLocation(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	idStr := extractTrainingIDFromPath(r.URL.Path, "/api/training/locations/")
	if idStr == "" {
		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}

	locationID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "無効なID形式です", http.StatusBadRequest)
		return
	}

	if err := h.repository.DeleteLocation(r.Context(), locationID, userID); err != nil {
		log.Printf("トレーニング場所の削除に失敗 (user_id=%s, id=%s): %v", userID, locationID, err)
		if errors.Is(err, repository.ErrTrainingNotFound) {
			http.Error(w, "トレーニング場所が見つかりません", http.StatusNotFound)
			return
		}
		http.Error(w, "トレーニング場所の削除に失敗しました", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Equipment ハンドラー

func (h *TrainingHandler) HandleListEquipment(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	// パスから location_id を抽出: /api/training/locations/{id}/equipment
	path := r.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/api/training/locations/"), "/")
	if len(parts) < 2 || parts[1] != "equipment" {
		http.Error(w, "無効なパスです", http.StatusBadRequest)
		return
	}

	locationID, err := uuid.Parse(parts[0])
	if err != nil {
		http.Error(w, "無効なlocation_id形式です", http.StatusBadRequest)
		return
	}

	// 場所の所有権確認
	location, err := h.repository.GetLocationByID(r.Context(), locationID, userID)
	if err != nil {
		log.Printf("トレーニング場所の取得に失敗 (user_id=%s, id=%s): %v", userID, locationID, err)
		http.Error(w, "トレーニング場所の取得に失敗しました", http.StatusInternalServerError)
		return
	}
	if location == nil {
		http.Error(w, "トレーニング場所が見つかりません", http.StatusNotFound)
		return
	}

	equipment, err := h.repository.GetEquipmentByLocation(r.Context(), locationID)
	if err != nil {
		log.Printf("器具の取得に失敗 (location_id=%s): %v", locationID, err)
		http.Error(w, "器具の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	if equipment == nil {
		equipment = []*repository.TrainingEquipment{}
	}

	data, err := json.Marshal(equipment)
	if err != nil {
		log.Printf("レスポンスのエンコードに失敗: %v", err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("レスポンス書き込みに失敗: %v", err)
		return
	}
}

func (h *TrainingHandler) HandleCreateEquipment(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	// パスから location_id を抽出
	path := r.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/api/training/locations/"), "/")
	if len(parts) < 2 || parts[1] != "equipment" {
		http.Error(w, "無効なパスです", http.StatusBadRequest)
		return
	}

	locationID, err := uuid.Parse(parts[0])
	if err != nil {
		http.Error(w, "無効なlocation_id形式です", http.StatusBadRequest)
		return
	}

	// 場所の所有権確認
	location, err := h.repository.GetLocationByID(r.Context(), locationID, userID)
	if err != nil {
		log.Printf("トレーニング場所の取得に失敗 (user_id=%s, id=%s): %v", userID, locationID, err)
		http.Error(w, "トレーニング場所の取得に失敗しました", http.StatusInternalServerError)
		return
	}
	if location == nil {
		http.Error(w, "トレーニング場所が見つかりません", http.StatusNotFound)
		return
	}

	var req CreateEquipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエスト形式です", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "名前は必須です", http.StatusBadRequest)
		return
	}
	if len(req.Name) > 100 {
		http.Error(w, "名前は100文字以内で入力してください", http.StatusBadRequest)
		return
	}

	equipment := &repository.TrainingEquipment{
		LocationID:   locationID,
		Name:         req.Name,
		OriginalName: req.OriginalName,
	}

	created, err := h.repository.CreateEquipment(r.Context(), equipment)
	if err != nil {
		log.Printf("器具の作成に失敗 (location_id=%s): %v", locationID, err)
		http.Error(w, "器具の作成に失敗しました", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(created)
	if err != nil {
		log.Printf("レスポンスのエンコードに失敗: %v", err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(data); err != nil {
		log.Printf("レスポンス書き込みに失敗: %v", err)
		return
	}
}

func (h *TrainingHandler) HandleUpdateEquipment(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	idStr := extractTrainingIDFromPath(r.URL.Path, "/api/training/equipment/")
	if idStr == "" {
		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}

	equipmentID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "無効なID形式です", http.StatusBadRequest)
		return
	}

	// 器具の存在確認と所有権確認
	existing, err := h.repository.GetEquipmentByID(r.Context(), equipmentID)
	if err != nil {
		log.Printf("器具の取得に失敗 (id=%s): %v", equipmentID, err)
		http.Error(w, "器具の取得に失敗しました", http.StatusInternalServerError)
		return
	}
	if existing == nil {
		http.Error(w, "器具が見つかりません", http.StatusNotFound)
		return
	}

	// 場所の所有権確認
	location, err := h.repository.GetLocationByID(r.Context(), existing.LocationID, userID)
	if err != nil {
		log.Printf("場所の所有権確認に失敗 (location_id=%s, user_id=%s): %v", existing.LocationID, userID, err)
		http.Error(w, "場所の所有権確認に失敗しました", http.StatusInternalServerError)
		return
	}
	if location == nil {
		http.Error(w, "アクセス権限がありません", http.StatusForbidden)
		return
	}

	var req UpdateEquipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエスト形式です", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "名前は必須です", http.StatusBadRequest)
		return
	}
	if len(req.Name) > 100 {
		http.Error(w, "名前は100文字以内で入力してください", http.StatusBadRequest)
		return
	}

	equipment := &repository.TrainingEquipment{
		ID:           equipmentID,
		Name:         req.Name,
		OriginalName: req.OriginalName,
	}

	updated, err := h.repository.UpdateEquipment(r.Context(), equipment)
	if err != nil {
		log.Printf("器具の更新に失敗 (id=%s): %v", equipmentID, err)
		if errors.Is(err, repository.ErrTrainingNotFound) {
			http.Error(w, "器具が見つかりません", http.StatusNotFound)
			return
		}
		http.Error(w, "器具の更新に失敗しました", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(updated)
	if err != nil {
		log.Printf("レスポンスのエンコードに失敗: %v", err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("レスポンス書き込みに失敗: %v", err)
		return
	}
}

func (h *TrainingHandler) HandleDeleteEquipment(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	idStr := extractTrainingIDFromPath(r.URL.Path, "/api/training/equipment/")
	if idStr == "" {
		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}

	equipmentID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "無効なID形式です", http.StatusBadRequest)
		return
	}

	// 器具の存在確認と所有権確認
	existing, err := h.repository.GetEquipmentByID(r.Context(), equipmentID)
	if err != nil {
		log.Printf("器具の取得に失敗 (id=%s): %v", equipmentID, err)
		http.Error(w, "器具の取得に失敗しました", http.StatusInternalServerError)
		return
	}
	if existing == nil {
		http.Error(w, "器具が見つかりません", http.StatusNotFound)
		return
	}

	// 場所の所有権確認
	location, err := h.repository.GetLocationByID(r.Context(), existing.LocationID, userID)
	if err != nil {
		log.Printf("場所の所有権確認に失敗 (location_id=%s, user_id=%s): %v", existing.LocationID, userID, err)
		http.Error(w, "場所の所有権確認に失敗しました", http.StatusInternalServerError)
		return
	}
	if location == nil {
		http.Error(w, "アクセス権限がありません", http.StatusForbidden)
		return
	}

	if err := h.repository.DeleteEquipment(r.Context(), equipmentID); err != nil {
		log.Printf("器具の削除に失敗 (id=%s): %v", equipmentID, err)
		if errors.Is(err, repository.ErrTrainingNotFound) {
			http.Error(w, "器具が見つかりません", http.StatusNotFound)
			return
		}
		http.Error(w, "器具の削除に失敗しました", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Menu ハンドラー

func (h *TrainingHandler) HandleListMenus(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	menus, err := h.repository.GetMenus(r.Context(), userID)
	if err != nil {
		log.Printf("メニューの取得に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "メニューの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	if menus == nil {
		menus = []*repository.TrainingMenu{}
	}

	data, err := json.Marshal(menus)
	if err != nil {
		log.Printf("レスポンスのエンコードに失敗: %v", err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("レスポンス書き込みに失敗: %v", err)
		return
	}
}

func (h *TrainingHandler) HandleCreateMenu(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	var req CreateMenuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエスト形式です", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "名前は必須です", http.StatusBadRequest)
		return
	}
	if len(req.Name) > 100 {
		http.Error(w, "名前は100文字以内で入力してください", http.StatusBadRequest)
		return
	}

	menu := &repository.TrainingMenu{
		UserID: &userID,
		Name:   req.Name,
	}

	created, err := h.repository.CreateMenu(r.Context(), menu)
	if err != nil {
		log.Printf("メニューの作成に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "メニューの作成に失敗しました", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(created)
	if err != nil {
		log.Printf("レスポンスのエンコードに失敗: %v", err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(data); err != nil {
		log.Printf("レスポンス書き込みに失敗: %v", err)
		return
	}
}

func (h *TrainingHandler) HandleDeleteMenu(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	idStr := extractTrainingIDFromPath(r.URL.Path, "/api/training/menus/")
	if idStr == "" {
		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}

	menuID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "無効なID形式です", http.StatusBadRequest)
		return
	}

	if err := h.repository.DeleteMenu(r.Context(), menuID, userID); err != nil {
		log.Printf("メニューの削除に失敗 (user_id=%s, id=%s): %v", userID, menuID, err)
		if errors.Is(err, repository.ErrTrainingNotFound) {
			http.Error(w, "メニューが見つかりません（固定メニューは削除できません）", http.StatusNotFound)
			return
		}
		http.Error(w, "メニューの削除に失敗しました", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Record ハンドラー

func (h *TrainingHandler) HandleListRecords(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	// クエリパラメータから期間を取得
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var startDate, endDate time.Time
	var err error

	if startStr != "" {
		startDate, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			http.Error(w, "無効なstart日付形式です（YYYY-MM-DD）", http.StatusBadRequest)
			return
		}
	} else {
		// デフォルト: 30日前
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if endStr != "" {
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			http.Error(w, "無効なend日付形式です（YYYY-MM-DD）", http.StatusBadRequest)
			return
		}
	} else {
		// デフォルト: 今日
		endDate = time.Now()
	}

	records, err := h.repository.GetRecords(r.Context(), userID, startDate, endDate)
	if err != nil {
		log.Printf("練習記録の取得に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "練習記録の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	if records == nil {
		records = []*repository.TrainingRecord{}
	}

	data, err := json.Marshal(records)
	if err != nil {
		log.Printf("レスポンスのエンコードに失敗: %v", err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("レスポンス書き込みに失敗: %v", err)
		return
	}
}

func (h *TrainingHandler) HandleCreateRecord(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	var req CreateRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエスト形式です", http.StatusBadRequest)
		return
	}

	if req.RecordedAt == "" {
		http.Error(w, "recorded_atは必須です", http.StatusBadRequest)
		return
	}

	recordedAt, err := time.Parse("2006-01-02", req.RecordedAt)
	if err != nil {
		http.Error(w, "無効な日付形式です（YYYY-MM-DD）", http.StatusBadRequest)
		return
	}

	// バリデーション
	if req.Intensity != nil && (*req.Intensity < 1 || *req.Intensity > 5) {
		http.Error(w, "強度は1-5の範囲で指定してください", http.StatusBadRequest)
		return
	}
	if req.Satisfaction != nil && (*req.Satisfaction < 1 || *req.Satisfaction > 5) {
		http.Error(w, "満足度は1-5の範囲で指定してください", http.StatusBadRequest)
		return
	}
	if req.Duration != nil && *req.Duration < 0 {
		http.Error(w, "練習時間は0以上の値を指定してください", http.StatusBadRequest)
		return
	}

	record := &repository.TrainingRecord{
		UserID:       userID,
		RecordedAt:   recordedAt,
		Completed:    req.Completed,
		Duration:     req.Duration,
		Intensity:    req.Intensity,
		Satisfaction: req.Satisfaction,
		Notes:        req.Notes,
	}

	if req.LocationID != nil {
		locationID, err := uuid.Parse(*req.LocationID)
		if err != nil {
			http.Error(w, "無効なlocation_id形式です", http.StatusBadRequest)
			return
		}
		// 場所の所有権確認
		location, err := h.repository.GetLocationByID(r.Context(), locationID, userID)
		if err != nil {
			log.Printf("場所の取得に失敗 (location_id=%s, user_id=%s): %v", locationID, userID, err)
			http.Error(w, "場所の取得に失敗しました", http.StatusInternalServerError)
			return
		}
		if location == nil {
			http.Error(w, "指定された場所が見つかりません", http.StatusBadRequest)
			return
		}
		record.LocationID = &locationID
	}

	// メニューIDをパース
	var menuIDs []uuid.UUID
	for _, menuIDStr := range req.MenuIDs {
		menuID, err := uuid.Parse(menuIDStr)
		if err != nil {
			http.Error(w, "無効なmenu_id形式です", http.StatusBadRequest)
			return
		}
		menuIDs = append(menuIDs, menuID)
	}

	created, err := h.repository.CreateRecord(r.Context(), record, menuIDs)
	if err != nil {
		log.Printf("練習記録の作成に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "練習記録の作成に失敗しました", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(created)
	if err != nil {
		log.Printf("レスポンスのエンコードに失敗: %v", err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(data); err != nil {
		log.Printf("レスポンス書き込みに失敗: %v", err)
		return
	}
}

func (h *TrainingHandler) HandleUpdateRecord(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	idStr := extractTrainingIDFromPath(r.URL.Path, "/api/training/records/")
	if idStr == "" {
		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}

	recordID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "無効なID形式です", http.StatusBadRequest)
		return
	}

	// 記録の存在確認
	existing, err := h.repository.GetRecordByID(r.Context(), recordID, userID)
	if err != nil {
		log.Printf("練習記録の取得に失敗 (id=%s, user_id=%s): %v", recordID, userID, err)
		http.Error(w, "練習記録の取得に失敗しました", http.StatusInternalServerError)
		return
	}
	if existing == nil {
		http.Error(w, "練習記録が見つかりません", http.StatusNotFound)
		return
	}

	var req UpdateRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエスト形式です", http.StatusBadRequest)
		return
	}

	if req.RecordedAt == "" {
		http.Error(w, "recorded_atは必須です", http.StatusBadRequest)
		return
	}

	recordedAt, err := time.Parse("2006-01-02", req.RecordedAt)
	if err != nil {
		http.Error(w, "無効な日付形式です（YYYY-MM-DD）", http.StatusBadRequest)
		return
	}

	// バリデーション
	if req.Intensity != nil && (*req.Intensity < 1 || *req.Intensity > 5) {
		http.Error(w, "強度は1-5の範囲で指定してください", http.StatusBadRequest)
		return
	}
	if req.Satisfaction != nil && (*req.Satisfaction < 1 || *req.Satisfaction > 5) {
		http.Error(w, "満足度は1-5の範囲で指定してください", http.StatusBadRequest)
		return
	}
	if req.Duration != nil && *req.Duration < 0 {
		http.Error(w, "練習時間は0以上の値を指定してください", http.StatusBadRequest)
		return
	}

	record := &repository.TrainingRecord{
		ID:           recordID,
		UserID:       userID,
		RecordedAt:   recordedAt,
		Completed:    req.Completed,
		Duration:     req.Duration,
		Intensity:    req.Intensity,
		Satisfaction: req.Satisfaction,
		Notes:        req.Notes,
	}

	if req.LocationID != nil {
		locationID, err := uuid.Parse(*req.LocationID)
		if err != nil {
			http.Error(w, "無効なlocation_id形式です", http.StatusBadRequest)
			return
		}
		// 場所の所有権確認
		location, err := h.repository.GetLocationByID(r.Context(), locationID, userID)
		if err != nil {
			log.Printf("場所の取得に失敗 (location_id=%s, user_id=%s): %v", locationID, userID, err)
			http.Error(w, "場所の取得に失敗しました", http.StatusInternalServerError)
			return
		}
		if location == nil {
			http.Error(w, "指定された場所が見つかりません", http.StatusBadRequest)
			return
		}
		record.LocationID = &locationID
	}

	// メニューIDをパース
	var menuIDs []uuid.UUID
	for _, menuIDStr := range req.MenuIDs {
		menuID, err := uuid.Parse(menuIDStr)
		if err != nil {
			http.Error(w, "無効なmenu_id形式です", http.StatusBadRequest)
			return
		}
		menuIDs = append(menuIDs, menuID)
	}

	updated, err := h.repository.UpdateRecord(r.Context(), record, menuIDs)
	if err != nil {
		log.Printf("練習記録の更新に失敗 (id=%s, user_id=%s): %v", recordID, userID, err)
		if errors.Is(err, repository.ErrTrainingNotFound) {
			http.Error(w, "練習記録が見つかりません", http.StatusNotFound)
			return
		}
		http.Error(w, "練習記録の更新に失敗しました", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(updated)
	if err != nil {
		log.Printf("レスポンスのエンコードに失敗: %v", err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("レスポンス書き込みに失敗: %v", err)
		return
	}
}

func (h *TrainingHandler) HandleDeleteRecord(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	idStr := extractTrainingIDFromPath(r.URL.Path, "/api/training/records/")
	if idStr == "" {
		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}

	recordID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "無効なID形式です", http.StatusBadRequest)
		return
	}

	if err := h.repository.DeleteRecord(r.Context(), recordID, userID); err != nil {
		log.Printf("練習記録の削除に失敗 (id=%s, user_id=%s): %v", recordID, userID, err)
		if errors.Is(err, repository.ErrTrainingNotFound) {
			http.Error(w, "練習記録が見つかりません", http.StatusNotFound)
			return
		}
		http.Error(w, "練習記録の削除に失敗しました", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// 後方互換性のためのハンドラー
func (h *TrainingHandler) HandleUpsertRecord(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	var req UpsertRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエスト形式です", http.StatusBadRequest)
		return
	}

	if req.RecordedAt == "" {
		http.Error(w, "recorded_atは必須です", http.StatusBadRequest)
		return
	}

	recordedAt, err := time.Parse("2006-01-02", req.RecordedAt)
	if err != nil {
		http.Error(w, "無効な日付形式です（YYYY-MM-DD）", http.StatusBadRequest)
		return
	}

	record := &repository.TrainingRecord{
		UserID:     userID,
		RecordedAt: recordedAt,
		Completed:  req.Completed,
	}

	if req.LocationID != nil {
		locationID, err := uuid.Parse(*req.LocationID)
		if err != nil {
			http.Error(w, "無効なlocation_id形式です", http.StatusBadRequest)
			return
		}
		// 場所の所有権確認
		location, err := h.repository.GetLocationByID(r.Context(), locationID, userID)
		if err != nil {
			log.Printf("場所の取得に失敗 (location_id=%s, user_id=%s): %v", locationID, userID, err)
			http.Error(w, "場所の取得に失敗しました", http.StatusInternalServerError)
			return
		}
		if location == nil {
			http.Error(w, "指定された場所が見つかりません", http.StatusBadRequest)
			return
		}
		record.LocationID = &locationID
	}

	upserted, err := h.repository.UpsertRecord(r.Context(), record)
	if err != nil {
		log.Printf("練習記録の作成/更新に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "練習記録の作成/更新に失敗しました", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(upserted)
	if err != nil {
		log.Printf("レスポンスのエンコードに失敗: %v", err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(data); err != nil {
		log.Printf("レスポンス書き込みに失敗: %v", err)
		return
	}
}

// Gemini連携ハンドラー

func (h *TrainingHandler) HandleSuggestMenu(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	var req SuggestMenuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエスト形式です", http.StatusBadRequest)
		return
	}

	if len(req.Equipment) == 0 {
		http.Error(w, "equipmentは必須です", http.StatusBadRequest)
		return
	}
	if req.Duration <= 0 {
		http.Error(w, "durationは0より大きい値が必要です", http.StatusBadRequest)
		return
	}

	if h.menuSuggester == nil {
		http.Error(w, "メニュー提案機能は利用できません", http.StatusServiceUnavailable)
		return
	}

	menu, err := h.menuSuggester.SuggestMenu(r.Context(), req.Equipment, req.Duration, req.Goals)
	if err != nil {
		log.Printf("メニュー提案に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "メニュー提案に失敗しました", http.StatusInternalServerError)
		return
	}

	response := SuggestMenuResponse{
		Menu: menu,
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("レスポンスのエンコードに失敗: %v", err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("レスポンス書き込みに失敗: %v", err)
		return
	}
}

func (h *TrainingHandler) HandleNormalizeEquipment(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == uuid.Nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	var req NormalizeEquipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエスト形式です", http.StatusBadRequest)
		return
	}

	if len(req.Names) == 0 {
		http.Error(w, "namesは必須です", http.StatusBadRequest)
		return
	}

	if h.equipmentNormalizer == nil {
		http.Error(w, "器具名正規化機能は利用できません", http.StatusServiceUnavailable)
		return
	}

	normalizedNames, err := h.equipmentNormalizer.NormalizeEquipmentNames(r.Context(), req.Names)
	if err != nil {
		log.Printf("器具名正規化に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "器具名正規化に失敗しました", http.StatusInternalServerError)
		return
	}

	response := NormalizeEquipmentResponse{
		NormalizedNames: normalizedNames,
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("レスポンスのエンコードに失敗: %v", err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("レスポンス書き込みに失敗: %v", err)
		return
	}
}

func extractTrainingIDFromPath(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	remaining := strings.TrimPrefix(path, prefix)
	parts := strings.Split(remaining, "/")
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}
	return parts[0]
}
