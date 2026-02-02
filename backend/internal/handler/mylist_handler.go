package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
)

type MylistHandler struct {
	repository         repository.MylistRepository
	analysisRepository repository.AnalysisRepository
	foodService        *service.FoodService
}

func NewMylistHandler(repository repository.MylistRepository, analysisRepository repository.AnalysisRepository, foodService *service.FoodService) *MylistHandler {
	return &MylistHandler{
		repository:         repository,
		analysisRepository: analysisRepository,
		foodService:        foodService,
	}
}

type CreateMylistItemRequest struct {
	Name          string                 `json:"name"`
	BaseAmount    string                 `json:"base_amount"`
	Unit          string                 `json:"unit"`
	Calories      float64                `json:"calories"`
	Protein       float64                `json:"protein"`
	Fat           float64                `json:"fat"`
	Carbohydrates float64                `json:"carbohydrates"`
	Foods         []gemini.NutritionInfo `json:"foods"`
	ImagePath     string                 `json:"image_path,omitempty"`
}

type UpdateMylistItemRequest struct {
	Name          string                 `json:"name"`
	BaseAmount    string                 `json:"base_amount"`
	Unit          string                 `json:"unit"`
	Calories      float64                `json:"calories"`
	Protein       float64                `json:"protein"`
	Fat           float64                `json:"fat"`
	Carbohydrates float64                `json:"carbohydrates"`
	Foods         []gemini.NutritionInfo `json:"foods"`
	ImagePath     string                 `json:"image_path,omitempty"`
}

type ReorderMylistRequest struct {
	ItemIDs []string `json:"item_ids"`
}

type AnalyzeMylistRequest struct {
	InputText string `json:"input_text"`
}

type AnalyzeMylistResponse struct {
	Foods              []gemini.NutritionInfo `json:"foods"`
	TotalCalories      float64                `json:"total_calories"`
	TotalProtein       float64                `json:"total_protein"`
	TotalFat           float64                `json:"total_fat"`
	TotalCarbohydrates float64                `json:"total_carbohydrates"`
}

type RecordFromMylistRequest struct {
	MylistItemID string  `json:"mylist_item_id"`
	Quantity     float64 `json:"quantity"`
	MealType     string  `json:"meal_type"`
	MealDate     string  `json:"meal_date"`
}

type RecordFromMylistResponse struct {
	ID string `json:"id"`
}

func (h *MylistHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	items, err := h.repository.GetAll(r.Context(), userID)
	if err != nil {
		log.Printf("マイリストの取得に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "マイリストの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	if items == nil {
		items = []*repository.MylistItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		http.Error(w, "レスポンスのエンコードに失敗しました", http.StatusInternalServerError)
		return
	}
}

func (h *MylistHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	var req CreateMylistItemRequest
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
	if req.BaseAmount == "" {
		http.Error(w, "基準量は必須です", http.StatusBadRequest)
		return
	}
	if req.Unit == "" {
		http.Error(w, "単位は必須です", http.StatusBadRequest)
		return
	}

	item := &repository.MylistItem{
		UserID:        userID,
		Name:          req.Name,
		BaseAmount:    req.BaseAmount,
		Unit:          req.Unit,
		Calories:      req.Calories,
		Protein:       req.Protein,
		Fat:           req.Fat,
		Carbohydrates: req.Carbohydrates,
		Foods:         req.Foods,
		ImagePath:     req.ImagePath,
	}

	created, err := h.repository.Create(r.Context(), item)
	if err != nil {
		log.Printf("マイリストアイテムの作成に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "マイリストアイテムの作成に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(created); err != nil {
		http.Error(w, "レスポンスのエンコードに失敗しました", http.StatusInternalServerError)
		return
	}
}

func (h *MylistHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	idStr := extractIDFromPath(r.URL.Path, "/api/mylist/")
	if idStr == "" {
		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}

	itemID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "無効なID形式です", http.StatusBadRequest)
		return
	}

	var req UpdateMylistItemRequest
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
	if req.BaseAmount == "" {
		http.Error(w, "基準量は必須です", http.StatusBadRequest)
		return
	}
	if req.Unit == "" {
		http.Error(w, "単位は必須です", http.StatusBadRequest)
		return
	}

	item := &repository.MylistItem{
		ID:            itemID,
		UserID:        userID,
		Name:          req.Name,
		BaseAmount:    req.BaseAmount,
		Unit:          req.Unit,
		Calories:      req.Calories,
		Protein:       req.Protein,
		Fat:           req.Fat,
		Carbohydrates: req.Carbohydrates,
		Foods:         req.Foods,
		ImagePath:     req.ImagePath,
	}

	updated, err := h.repository.Update(r.Context(), item)
	if err != nil {
		log.Printf("マイリストアイテムの更新に失敗 (user_id=%s, id=%s): %v", userID, itemID, err)
		if strings.Contains(err.Error(), "見つかりません") {
			http.Error(w, "マイリストアイテムが見つかりません", http.StatusNotFound)
			return
		}
		http.Error(w, "マイリストアイテムの更新に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updated); err != nil {
		http.Error(w, "レスポンスのエンコードに失敗しました", http.StatusInternalServerError)
		return
	}
}

func (h *MylistHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	idStr := extractIDFromPath(r.URL.Path, "/api/mylist/")
	if idStr == "" {
		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}

	itemID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "無効なID形式です", http.StatusBadRequest)
		return
	}

	if err := h.repository.Delete(r.Context(), itemID.String(), userID); err != nil {
		log.Printf("マイリストアイテムの削除に失敗 (user_id=%s, id=%s): %v", userID, itemID, err)
		if strings.Contains(err.Error(), "見つかりません") {
			http.Error(w, "マイリストアイテムが見つかりません", http.StatusNotFound)
			return
		}
		http.Error(w, "マイリストアイテムの削除に失敗しました", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MylistHandler) HandleReorder(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	var req ReorderMylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエスト形式です", http.StatusBadRequest)
		return
	}

	if len(req.ItemIDs) == 0 {
		http.Error(w, "item_idsは必須です", http.StatusBadRequest)
		return
	}

	itemIDs := make([]uuid.UUID, len(req.ItemIDs))
	for i, idStr := range req.ItemIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "無効なID形式です", http.StatusBadRequest)
			return
		}
		itemIDs[i] = id
	}

	if err := h.repository.Reorder(r.Context(), userID, itemIDs); err != nil {
		log.Printf("マイリストの並び替えに失敗 (user_id=%s): %v", userID, err)
		if strings.Contains(err.Error(), "見つかりません") {
			http.Error(w, "マイリストアイテムが見つかりません", http.StatusNotFound)
			return
		}
		http.Error(w, "マイリストの並び替えに失敗しました", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MylistHandler) HandleAnalyze(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	var req AnalyzeMylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエスト形式です", http.StatusBadRequest)
		return
	}

	if req.InputText == "" {
		http.Error(w, "input_textは必須です", http.StatusBadRequest)
		return
	}

	result, err := h.foodService.AnalyzeFoodText(r.Context(), req.InputText)
	if err != nil {
		log.Printf("AI分析に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "AI分析に失敗しました", http.StatusInternalServerError)
		return
	}

	response := AnalyzeMylistResponse{
		Foods:              result.Foods,
		TotalCalories:      result.TotalCalories,
		TotalProtein:       result.TotalProtein,
		TotalFat:           result.TotalFat,
		TotalCarbohydrates: result.TotalCarbohydrates,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "レスポンスのエンコードに失敗しました", http.StatusInternalServerError)
		return
	}
}

func (h *MylistHandler) HandleGetByID(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	idStr := extractIDFromPath(r.URL.Path, "/api/mylist/")
	if idStr == "" {
		http.Error(w, "IDが指定されていません", http.StatusBadRequest)
		return
	}

	itemID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "無効なID形式です", http.StatusBadRequest)
		return
	}

	item, err := h.repository.GetByID(r.Context(), itemID.String(), userID)
	if err != nil {
		log.Printf("マイリストアイテムの取得に失敗 (user_id=%s, id=%s): %v", userID, itemID, err)
		http.Error(w, "マイリストアイテムの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	if item == nil {
		http.Error(w, "マイリストアイテムが見つかりません", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(item); err != nil {
		http.Error(w, "レスポンスのエンコードに失敗しました", http.StatusInternalServerError)
		return
	}
}

func extractIDFromPath(path, prefix string) string {
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

func (h *MylistHandler) HandleRecordFromMylist(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	var req RecordFromMylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無効なリクエスト形式です", http.StatusBadRequest)
		return
	}

	if req.MylistItemID == "" {
		http.Error(w, "mylist_item_idは必須です", http.StatusBadRequest)
		return
	}
	if req.Quantity <= 0 {
		http.Error(w, "quantityは0より大きい値が必要です", http.StatusBadRequest)
		return
	}
	if req.MealType == "" {
		http.Error(w, "meal_typeは必須です", http.StatusBadRequest)
		return
	}
	validMealTypes := map[string]bool{"breakfast": true, "lunch": true, "dinner": true, "snack": true}
	if !validMealTypes[req.MealType] {
		http.Error(w, "無効なmeal_typeです", http.StatusBadRequest)
		return
	}
	if req.MealDate == "" {
		http.Error(w, "meal_dateは必須です", http.StatusBadRequest)
		return
	}

	itemID, err := uuid.Parse(req.MylistItemID)
	if err != nil {
		http.Error(w, "無効なmylist_item_id形式です", http.StatusBadRequest)
		return
	}

	// マイリストアイテムを取得
	item, err := h.repository.GetByID(r.Context(), itemID.String(), userID)
	if err != nil {
		log.Printf("マイリストアイテムの取得に失敗 (user_id=%s, id=%s): %v", userID, itemID, err)
		http.Error(w, "マイリストアイテムの取得に失敗しました", http.StatusInternalServerError)
		return
	}
	if item == nil {
		http.Error(w, "マイリストアイテムが見つかりません", http.StatusNotFound)
		return
	}

	// 数量を適用して栄養素を計算
	scaledFoods := make([]gemini.NutritionInfo, len(item.Foods))
	for i, food := range item.Foods {
		scaledFoods[i] = gemini.NutritionInfo{
			Name:            food.Name,
			EstimatedAmount: food.EstimatedAmount,
			Calories:        food.Calories * req.Quantity,
			Protein:         food.Protein * req.Quantity,
			Fat:             food.Fat * req.Quantity,
			Carbohydrates:   food.Carbohydrates * req.Quantity,
		}
	}

	result := &service.AnalysisResult{
		Foods:              scaledFoods,
		TotalCalories:      item.Calories * req.Quantity,
		TotalProtein:       item.Protein * req.Quantity,
		TotalFat:           item.Fat * req.Quantity,
		TotalCarbohydrates: item.Carbohydrates * req.Quantity,
	}

	// 入力テキストを生成
	inputText := item.Name
	if req.Quantity != 1.0 {
		inputText = fmt.Sprintf("%s (x%s)", item.Name, formatQuantity(req.Quantity))
	}

	// 食事記録として保存
	recordID, err := h.analysisRepository.CreateRequestFromMylist(r.Context(), inputText, req.MealType, req.MealDate, &userID, result)
	if err != nil {
		log.Printf("食事記録の保存に失敗 (user_id=%s): %v", userID, err)
		http.Error(w, "食事記録の保存に失敗しました", http.StatusInternalServerError)
		return
	}

	response := RecordFromMylistResponse{
		ID: recordID.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "レスポンスのエンコードに失敗しました", http.StatusInternalServerError)
		return
	}
}

func formatQuantity(q float64) string {
	if q == float64(int(q)) {
		return fmt.Sprintf("%d", int(q))
	}
	s := fmt.Sprintf("%.2f", q)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}
