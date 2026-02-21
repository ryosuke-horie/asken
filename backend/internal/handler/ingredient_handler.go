package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
)

// IngredientHandler は食材管理 CRUD エンドポイントのハンドラー
type IngredientHandler struct {
	repository repository.IngredientRepository
}

// NewIngredientHandler は新しいIngredientHandlerを作成
func NewIngredientHandler(repo repository.IngredientRepository) *IngredientHandler {
	if repo == nil {
		panic("ingredient handler: repository must not be nil")
	}
	return &IngredientHandler{repository: repo}
}

// CreateIngredientRequest は食材作成リクエスト
type CreateIngredientRequest struct {
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	PurchaseDate string  `json:"purchase_date"` // YYYY-MM-DD or empty
	ExpiryDate   string  `json:"expiry_date"`   // YYYY-MM-DD or empty
	Source       string  `json:"source"`
}

// UpdateIngredientRequest は食材更新リクエスト
type UpdateIngredientRequest struct {
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	PurchaseDate string  `json:"purchase_date"` // YYYY-MM-DD or empty
	ExpiryDate   string  `json:"expiry_date"`   // YYYY-MM-DD or empty
}

// IngredientResponse は食材のレスポンス
type IngredientResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	PurchaseDate string  `json:"purchase_date,omitempty"`
	ExpiryDate   string  `json:"expiry_date,omitempty"`
	Source       string  `json:"source"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// IngredientsListResponse は食材一覧のレスポンス
type IngredientsListResponse struct {
	Ingredients []IngredientResponse `json:"ingredients"`
}

// HandleList はGET /api/ingredients リクエストを処理
func (h *IngredientHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	category := r.URL.Query().Get("category")
	if category != "" {
		if err := validateIngredientCategory(category); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	items, err := h.repository.List(r.Context(), userID, category)
	if err != nil {
		log.Printf("Error listing ingredients: userID=%s, error=%v", userID, err)
		http.Error(w, "食材の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	response := IngredientsListResponse{
		Ingredients: make([]IngredientResponse, 0, len(items)),
	}
	for _, item := range items {
		response.Ingredients = append(response.Ingredients, toIngredientResponse(item))
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

// HandleCreate はPOST /api/ingredients リクエストを処理
func (h *IngredientHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4*1024) // 4KB

	var req CreateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: userID=%s, error=%v", userID, err)
		http.Error(w, "リクエストのパースに失敗しました", http.StatusBadRequest)
		return
	}

	input, err := validateAndBuildCreateInput(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	item, err := h.repository.Create(r.Context(), userID, *input)
	if err != nil {
		log.Printf("Error creating ingredient: userID=%s, error=%v", userID, err)
		http.Error(w, "食材の作成に失敗しました", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(toIngredientResponse(*item))
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

// HandleUpdate はPUT /api/ingredients/{id} リクエストを処理
func (h *IngredientHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	ingredientID, err := extractIngredientID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4*1024)

	var req UpdateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: userID=%s, error=%v", userID, err)
		http.Error(w, "リクエストのパースに失敗しました", http.StatusBadRequest)
		return
	}

	input, err := validateAndBuildUpdateInput(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	item, err := h.repository.Update(r.Context(), userID, ingredientID, *input)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "食材が見つかりません", http.StatusNotFound)
			return
		}
		log.Printf("Error updating ingredient: userID=%s, ingredientID=%s, error=%v", userID, ingredientID, err)
		http.Error(w, "食材の更新に失敗しました", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(toIngredientResponse(*item))
	if err != nil {
		log.Printf("Error encoding response: userID=%s, ingredientID=%s, error=%v", userID, ingredientID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("Error writing response: userID=%s, ingredientID=%s, error=%v", userID, ingredientID, err)
	}
}

// HandleDelete はDELETE /api/ingredients/{id} リクエストを処理
func (h *IngredientHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	ingredientID, err := extractIngredientID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.repository.Delete(r.Context(), userID, ingredientID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "食材が見つかりません", http.StatusNotFound)
			return
		}
		log.Printf("Error deleting ingredient: userID=%s, ingredientID=%s, error=%v", userID, ingredientID, err)
		http.Error(w, "食材の削除に失敗しました", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// validateAndBuildCreateInput はCreateIngredientRequestをバリデーションしてCreateIngredientInputを返す
func validateAndBuildCreateInput(req CreateIngredientRequest) (*repository.CreateIngredientInput, error) {
	if err := validateIngredientName(req.Name); err != nil {
		return nil, err
	}
	if err := validateIngredientCategory(req.Category); err != nil {
		return nil, err
	}
	if err := validateIngredientQuantity(req.Quantity); err != nil {
		return nil, err
	}
	if err := validateIngredientUnit(req.Unit); err != nil {
		return nil, err
	}
	if err := validateIngredientSource(req.Source); err != nil {
		return nil, err
	}

	purchaseDate, err := parseOptionalDate(req.PurchaseDate, "purchaseDate")
	if err != nil {
		return nil, err
	}
	expiryDate, err := parseOptionalDate(req.ExpiryDate, "expiryDate")
	if err != nil {
		return nil, err
	}

	return &repository.CreateIngredientInput{
		Name:         req.Name,
		Category:     req.Category,
		Quantity:     req.Quantity,
		Unit:         req.Unit,
		PurchaseDate: purchaseDate,
		ExpiryDate:   expiryDate,
		Source:       req.Source,
	}, nil
}

// validateAndBuildUpdateInput はUpdateIngredientRequestをバリデーションしてUpdateIngredientInputを返す
func validateAndBuildUpdateInput(req UpdateIngredientRequest) (*repository.UpdateIngredientInput, error) {
	if err := validateIngredientName(req.Name); err != nil {
		return nil, err
	}
	if err := validateIngredientCategory(req.Category); err != nil {
		return nil, err
	}
	if err := validateIngredientQuantity(req.Quantity); err != nil {
		return nil, err
	}
	if err := validateIngredientUnit(req.Unit); err != nil {
		return nil, err
	}

	purchaseDate, err := parseOptionalDate(req.PurchaseDate, "purchaseDate")
	if err != nil {
		return nil, err
	}
	expiryDate, err := parseOptionalDate(req.ExpiryDate, "expiryDate")
	if err != nil {
		return nil, err
	}

	return &repository.UpdateIngredientInput{
		Name:         req.Name,
		Category:     req.Category,
		Quantity:     req.Quantity,
		Unit:         req.Unit,
		PurchaseDate: purchaseDate,
		ExpiryDate:   expiryDate,
	}, nil
}

// validateIngredientName は食材名のバリデーション
func validateIngredientName(name string) error {
	length := utf8.RuneCountInString(name)
	if length == 0 {
		return fmt.Errorf("nameは必須です")
	}
	if length > 100 {
		return fmt.Errorf("nameは100文字以内で指定してください")
	}
	return nil
}

// validateIngredientCategory は食材カテゴリのバリデーション
func validateIngredientCategory(category string) error {
	validCategories := map[repository.IngredientCategory]bool{
		repository.CategoryMeat:      true,
		repository.CategoryFish:      true,
		repository.CategoryVegetable: true,
		repository.CategoryFruit:     true,
		repository.CategoryDairy:     true,
		repository.CategoryGrain:     true,
		repository.CategorySeasoning: true,
		repository.CategoryBeverage:  true,
		repository.CategoryOther:     true,
	}
	if !validCategories[repository.IngredientCategory(category)] {
		return fmt.Errorf("categoryはmeat, fish, vegetable, fruit, dairy, grain, seasoning, beverage, otherのいずれかである必要があります")
	}
	return nil
}

// validateIngredientQuantity は数量のバリデーション
func validateIngredientQuantity(quantity float64) error {
	if math.IsNaN(quantity) || math.IsInf(quantity, 0) {
		return fmt.Errorf("quantityが不正な値です")
	}
	if quantity <= 0 {
		return fmt.Errorf("quantityは0より大きい値である必要があります")
	}
	if quantity > 999999 {
		return fmt.Errorf("quantityが大きすぎます（999999以下で指定してください）")
	}
	return nil
}

// validateIngredientUnit は単位のバリデーション
func validateIngredientUnit(unit string) error {
	validUnits := make(map[string]bool)
	for _, u := range gemini.SupportedUnits() {
		validUnits[u] = true
	}
	if !validUnits[unit] {
		return fmt.Errorf("unitはサポートされていない単位です")
	}
	return nil
}

// validateIngredientSource は入力元のバリデーション
func validateIngredientSource(source string) error {
	s := repository.IngredientSource(source)
	if s != repository.SourceReceipt && s != repository.SourceManual {
		return fmt.Errorf("sourceはreceipt またはmanualである必要があります")
	}
	return nil
}

// parseOptionalDate はオプショナルな日付文字列（YYYY-MM-DD）を*time.Timeに変換する
func parseOptionalDate(dateStr string, fieldName string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("%sはYYYY-MM-DD形式で指定してください", fieldName)
	}
	return &t, nil
}

// extractIngredientID はURLパスから食材IDを抽出しUUID形式を検証する
func extractIngredientID(path string) (string, error) {
	// /api/ingredients/{id} からIDを抽出
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(parts) < 3 || parts[len(parts)-1] == "" {
		return "", fmt.Errorf("食材IDが指定されていません")
	}
	id := parts[len(parts)-1]
	if _, err := uuid.Parse(id); err != nil {
		return "", fmt.Errorf("食材IDの形式が不正です")
	}
	return id, nil
}

// toIngredientResponse はIngredientをレスポンスに変換
func toIngredientResponse(item repository.Ingredient) IngredientResponse {
	resp := IngredientResponse{
		ID:        item.ID,
		Name:      item.Name,
		Category:  item.Category,
		Quantity:  item.Quantity,
		Unit:      item.Unit,
		Source:    item.Source,
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.Format(time.RFC3339),
	}
	if item.PurchaseDate != nil {
		resp.PurchaseDate = item.PurchaseDate.Format("2006-01-02")
	}
	if item.ExpiryDate != nil {
		resp.ExpiryDate = item.ExpiryDate.Format("2006-01-02")
	}
	return resp
}
