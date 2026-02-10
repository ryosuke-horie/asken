package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
)

// MyMenuHandler はマイメニュー CRUD エンドポイントのハンドラー
type MyMenuHandler struct {
	myMenuRepo   repository.MyMenuRepository
	analysisRepo repository.AnalysisRepository
}

// NewMyMenuHandler は新しいMyMenuHandlerを作成
func NewMyMenuHandler(myMenuRepo repository.MyMenuRepository, analysisRepo repository.AnalysisRepository) *MyMenuHandler {
	if myMenuRepo == nil || analysisRepo == nil {
		panic("my menu handler: repositories must not be nil")
	}
	return &MyMenuHandler{
		myMenuRepo:   myMenuRepo,
		analysisRepo: analysisRepo,
	}
}

// CreateMyMenuRequest はマイメニュー作成リクエスト
type CreateMyMenuRequest struct {
	Name  string                `json:"name"`
	Foods []gemini.NutritionInfo `json:"foods"`
}

// UpdateMyMenuRequest はマイメニュー更新リクエスト
type UpdateMyMenuRequest struct {
	Name  string                `json:"name"`
	Foods []gemini.NutritionInfo `json:"foods"`
}

// RecordMyMenuRequest はマイメニューから食事記録するリクエスト
type RecordMyMenuRequest struct {
	MealType string `json:"meal_type"`
	MealDate string `json:"meal_date"`
}

// MyMenuResponse はマイメニューのレスポンス
type MyMenuResponse struct {
	ID                 string                    `json:"id"`
	Name               string                    `json:"name"`
	Foods              []gemini.NutritionInfo    `json:"foods"`
	TotalCalories      float64                   `json:"totalCalories"`
	TotalProtein       float64                   `json:"totalProtein"`
	TotalFat           float64                   `json:"totalFat"`
	TotalCarbohydrates float64                   `json:"totalCarbohydrates"`
	CreatedAt          string                    `json:"createdAt"`
	UpdatedAt          string                    `json:"updatedAt"`
}

// AnalysisIDResponse は分析IDのレスポンス
type AnalysisIDResponse struct {
	ID string `json:"id"`
}

// HandleList はGET /api/my-menu リクエストを処理
func (h *MyMenuHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	items, err := h.myMenuRepo.List(r.Context(), userID)
	if err != nil {
		log.Printf("Error listing my menu: userID=%s, error=%v", userID, err)
		http.Error(w, "マイメニューの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	response := make([]MyMenuResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toMyMenuResponse(item))
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error encoding response: userID=%s, error=%v", userID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("Error writing response: userID=%s, error=%v", userID, err)
	}
}

// HandleCreate はPOST /api/my-menu リクエストを処理
func (h *MyMenuHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 16*1024) // 16KB

	var req CreateMyMenuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: userID=%s, error=%v", userID, err)
		http.Error(w, "リクエストのパースに失敗しました", http.StatusBadRequest)
		return
	}

	if err := validateMyMenuName(req.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateFoods(req.Foods); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	item, err := h.myMenuRepo.Create(r.Context(), userID, req.Name, req.Foods)
	if err != nil {
		log.Printf("Error creating my menu: userID=%s, error=%v", userID, err)
		http.Error(w, "マイメニューの作成に失敗しました", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(toMyMenuResponse(*item))
	if err != nil {
		log.Printf("Error encoding response: userID=%s, error=%v", userID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(data); err != nil {
		log.Printf("Error writing response: userID=%s, error=%v", userID, err)
	}
}

// HandleGet はGET /api/my-menu/{id} リクエストを処理
func (h *MyMenuHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	menuID, err := extractMenuID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	item, err := h.myMenuRepo.Get(r.Context(), userID, menuID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "マイメニューが見つかりません", http.StatusNotFound)
			return
		}
		log.Printf("Error getting my menu: userID=%s, menuID=%s, error=%v", userID, menuID, err)
		http.Error(w, "マイメニューの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(toMyMenuResponse(*item))
	if err != nil {
		log.Printf("Error encoding response: userID=%s, menuID=%s, error=%v", userID, menuID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("Error writing response: userID=%s, menuID=%s, error=%v", userID, menuID, err)
	}
}

// HandleUpdate はPUT /api/my-menu/{id} リクエストを処理
func (h *MyMenuHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	menuID, err := extractMenuID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)

	var req UpdateMyMenuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: userID=%s, error=%v", userID, err)
		http.Error(w, "リクエストのパースに失敗しました", http.StatusBadRequest)
		return
	}

	if err := validateMyMenuName(req.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateFoods(req.Foods); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	item, err := h.myMenuRepo.Update(r.Context(), userID, menuID, req.Name, req.Foods)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "マイメニューが見つかりません", http.StatusNotFound)
			return
		}
		log.Printf("Error updating my menu: userID=%s, menuID=%s, error=%v", userID, menuID, err)
		http.Error(w, "マイメニューの更新に失敗しました", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(toMyMenuResponse(*item))
	if err != nil {
		log.Printf("Error encoding response: userID=%s, menuID=%s, error=%v", userID, menuID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("Error writing response: userID=%s, menuID=%s, error=%v", userID, menuID, err)
	}
}

// HandleDelete はDELETE /api/my-menu/{id} リクエストを処理
func (h *MyMenuHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	menuID, err := extractMenuID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.myMenuRepo.Delete(r.Context(), userID, menuID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "マイメニューが見つかりません", http.StatusNotFound)
			return
		}
		log.Printf("Error deleting my menu: userID=%s, menuID=%s, error=%v", userID, menuID, err)
		http.Error(w, "マイメニューの削除に失敗しました", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleRecord はPOST /api/my-menu/{id}/record リクエストを処理
func (h *MyMenuHandler) HandleRecord(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	menuID, err := extractMenuIDForRecord(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// マイメニューを取得
	menu, err := h.myMenuRepo.Get(r.Context(), userID, menuID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "マイメニューが見つかりません", http.StatusNotFound)
			return
		}
		log.Printf("Error getting my menu: userID=%s, menuID=%s, error=%v", userID, menuID, err)
		http.Error(w, "マイメニューの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1024)

	var req RecordMyMenuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: userID=%s, error=%v", userID, err)
		http.Error(w, "リクエストのパースに失敗しました", http.StatusBadRequest)
		return
	}

	// バリデーション
	if err := validateMealType(req.MealType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.MealDate == "" {
		// デフォルトは今日の日付
		req.MealDate = time.Now().Format("2006-01-02")
	}

	if _, err := time.Parse("2006-01-02", req.MealDate); err != nil {
		http.Error(w, "meal_dateはYYYY-MM-DD形式で指定してください", http.StatusBadRequest)
		return
	}

	// 既存のCreateRequestFromMylistメソッドを使用
	analysisResult := &service.AnalysisResult{
		Foods:              menu.Foods,
		TotalCalories:      menu.TotalCalories,
		TotalProtein:       menu.TotalProtein,
		TotalFat:           menu.TotalFat,
		TotalCarbohydrates: menu.TotalCarbohydrates,
	}

	inputText := menu.Name // マイメニュー名をinputTextに使用

	analysisID, err := h.analysisRepo.CreateRequestFromMylist(
		r.Context(),
		inputText,
		req.MealType,
		req.MealDate,
		&userID,
		analysisResult,
	)
	if err != nil {
		log.Printf("Error creating meal record from my menu: userID=%s, menuID=%s, error=%v", userID, menuID, err)
		http.Error(w, "食事記録の作成に失敗しました", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(AnalysisIDResponse{ID: analysisID.String()})
	if err != nil {
		log.Printf("Error encoding response: userID=%s, menuID=%s, error=%v", userID, menuID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(data); err != nil {
		log.Printf("Error writing response: userID=%s, menuID=%s, error=%v", userID, menuID, err)
	}
}

// validateMyMenuName はマイメニュー名のバリデーション
func validateMyMenuName(name string) error {
	if name == "" {
		return fmt.Errorf("メニュー名は必須です")
	}
	if len(name) > 50 {
		return fmt.Errorf("メニュー名は50文字以内である必要があります")
	}
	return nil
}

// validateFoods は食品リストのバリデーション
func validateFoods(foods []gemini.NutritionInfo) error {
	if len(foods) == 0 {
		return fmt.Errorf("少なくとも1つの食品が必要です")
	}
	if len(foods) > 100 {
		return fmt.Errorf("食品は100個以内である必要があります")
	}
	return nil
}

// validateMealType は食事タイプのバリデーション
func validateMealType(mealType string) error {
	validTypes := map[string]bool{
		"breakfast": true,
		"lunch":     true,
		"dinner":    true,
		"snack":     true,
	}
	if !validTypes[mealType] {
		return fmt.Errorf("meal_typeはbreakfast, lunch, dinner, snackのいずれかである必要があります")
	}
	return nil
}

// extractMenuID はURLパスからメニューIDを抽出
func extractMenuID(path string) (string, error) {
	// /api/my-menu/{id} からIDを抽出
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(parts) < 3 || parts[len(parts)-1] == "" {
		return "", fmt.Errorf("メニューIDが指定されていません")
	}
	id := parts[len(parts)-1]
	if _, err := uuid.Parse(id); err != nil {
		return "", fmt.Errorf("メニューIDの形式が不正です")
	}
	return id, nil
}

// extractMenuIDForRecord はURLパスからメニューIDを抽出（/record用）
func extractMenuIDForRecord(path string) (string, error) {
	// /api/my-menu/{id}/record からIDを抽出
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(parts) < 4 {
		return "", fmt.Errorf("メニューIDが指定されていません")
	}
	id := parts[3]
	if _, err := uuid.Parse(id); err != nil {
		return "", fmt.Errorf("メニューIDの形式が不正です")
	}
	return id, nil
}

// toMyMenuResponse はMyMenuItemをレスポンスに変換
func toMyMenuResponse(item repository.MyMenuItem) MyMenuResponse {
	return MyMenuResponse{
		ID:                 item.ID,
		Name:               item.Name,
		Foods:              item.Foods,
		TotalCalories:      item.TotalCalories,
		TotalProtein:       item.TotalProtein,
		TotalFat:           item.TotalFat,
		TotalCarbohydrates: item.TotalCarbohydrates,
		CreatedAt:          item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          item.UpdatedAt.Format(time.RFC3339),
	}
}
