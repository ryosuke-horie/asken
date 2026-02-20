package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
)

// MenuSuggestionHandler はメニューサジェスト関連のエンドポイントを処理するハンドラー
type MenuSuggestionHandler struct {
	menuRepo          repository.MenuSuggestionRepository
	ingredientRepo    repository.IngredientRepository
	nutritionGoalRepo repository.NutritionGoalRepository
	weightRecordRepo  repository.WeightRecordRepository
	analysisRepo      repository.AnalysisRepository
	suggester         *gemini.MenuSuggester
}

// NewMenuSuggestionHandler は新しいMenuSuggestionHandlerを作成する
func NewMenuSuggestionHandler(
	menuRepo repository.MenuSuggestionRepository,
	ingredientRepo repository.IngredientRepository,
	nutritionGoalRepo repository.NutritionGoalRepository,
	weightRecordRepo repository.WeightRecordRepository,
	analysisRepo repository.AnalysisRepository,
	suggester *gemini.MenuSuggester,
) *MenuSuggestionHandler {
	if menuRepo == nil {
		panic("menu suggestion handler: menuRepo must not be nil")
	}
	if ingredientRepo == nil {
		panic("menu suggestion handler: ingredientRepo must not be nil")
	}
	if nutritionGoalRepo == nil {
		panic("menu suggestion handler: nutritionGoalRepo must not be nil")
	}
	if weightRecordRepo == nil {
		panic("menu suggestion handler: weightRecordRepo must not be nil")
	}
	if analysisRepo == nil {
		panic("menu suggestion handler: analysisRepo must not be nil")
	}
	if suggester == nil {
		panic("menu suggestion handler: suggester must not be nil")
	}
	return &MenuSuggestionHandler{
		menuRepo:          menuRepo,
		ingredientRepo:    ingredientRepo,
		nutritionGoalRepo: nutritionGoalRepo,
		weightRecordRepo:  weightRecordRepo,
		analysisRepo:      analysisRepo,
		suggester:         suggester,
	}
}

// suggestRequest はPOST /api/menu/suggest のリクエスト
type suggestRequest struct {
	MealType string `json:"mealType"`
	Count    int    `json:"count"`
}

// menuSuggestionResponse はメニューサジェストのレスポンス
type menuSuggestionResponse struct {
	ID                 string                             `json:"id"`
	Title              string                             `json:"title"`
	Description        string                             `json:"description"`
	Reason             string                             `json:"reason"`
	IngredientsUsed    []menuSuggestionIngredientResponse `json:"ingredientsUsed"`
	Recipe             string                             `json:"recipe,omitempty"`
	EstimatedNutrition estimatedNutritionResponse         `json:"estimatedNutrition"`
	MealType           string                             `json:"mealType"`
	Status             string                             `json:"status"`
	CreatedAt          string                             `json:"createdAt"`
}

// menuSuggestionIngredientResponse はサジェスト食材のレスポンス
type menuSuggestionIngredientResponse struct {
	IngredientID string  `json:"ingredientId"`
	Name         string  `json:"name"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
}

// estimatedNutritionResponse は推定栄養素のレスポンス
type estimatedNutritionResponse struct {
	Calories      float64 `json:"calories"`
	Protein       float64 `json:"protein"`
	Fat           float64 `json:"fat"`
	Carbohydrates float64 `json:"carbohydrates"`
}

// suggestionsListResponse はサジェスト一覧のレスポンス
type suggestionsListResponse struct {
	Suggestions []menuSuggestionResponse `json:"suggestions"`
}

// acceptResponse はサジェスト採用のレスポンス
type acceptResponse struct {
	AnalysisRequestID   string                   `json:"analysisRequestId"`
	DeductedIngredients []deductedIngredientResp `json:"deductedIngredients"`
}

// deductedIngredientResp は控除食材のレスポンス
type deductedIngredientResp struct {
	IngredientID string  `json:"ingredientId"`
	Name         string  `json:"name"`
	Deducted     float64 `json:"deducted"`
	Remaining    float64 `json:"remaining"`
}

// HandleSuggest はPOST /api/menu/suggest を処理する
func (h *MenuSuggestionHandler) HandleSuggest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4*1024)
	var req suggestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "リクエストのパースに失敗しました", http.StatusBadRequest)
		return
	}

	if err := validateMenuSuggestionMealType(req.MealType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Count <= 0 {
		req.Count = 3
	}
	if req.Count > 5 {
		req.Count = 5
	}

	// コンテキスト収集
	input, err := h.buildSuggestionInput(r, userID, req.MealType, req.Count)
	if err != nil {
		log.Printf("Error building suggestion input: userID=%s, error=%v", userID, err)
		http.Error(w, "コンテキストの収集に失敗しました", http.StatusInternalServerError)
		return
	}

	// Gemini API でメニュー提案を生成
	geminiSuggestions, err := h.suggester.SuggestMenus(r.Context(), *input)
	if err != nil {
		log.Printf("Error generating menu suggestions: userID=%s, error=%v", userID, err)
		http.Error(w, "メニュー提案の生成に失敗しました", http.StatusInternalServerError)
		return
	}

	// 提案結果を Firestore に保存
	savedSuggestions := make([]menuSuggestionResponse, 0, len(geminiSuggestions))
	for _, gs := range geminiSuggestions {
		createInput := h.buildCreateInput(gs, req.MealType, input.Ingredients)
		saved, err := h.menuRepo.Create(r.Context(), userID, createInput)
		if err != nil {
			log.Printf("Error saving menu suggestion: userID=%s, title=%s, error=%v", userID, gs.Title, err)
			http.Error(w, "メニューサジェストの保存に失敗しました", http.StatusInternalServerError)
			return
		}
		savedSuggestions = append(savedSuggestions, toMenuSuggestionResponse(*saved))
	}

	resp := suggestionsListResponse{Suggestions: savedSuggestions}
	data, err := json.Marshal(resp)
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

// HandleList はGET /api/menu/suggestions を処理する
func (h *MenuSuggestionHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	statusParam := r.URL.Query().Get("status")
	if statusParam == "" {
		statusParam = string(repository.MenuStatusSuggested)
	}
	if statusParam != "all" {
		if err := validateSuggestionStatus(statusParam); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	filterStatus := statusParam
	if statusParam == "all" {
		filterStatus = ""
	}

	items, err := h.menuRepo.List(r.Context(), userID, filterStatus, limit)
	if err != nil {
		log.Printf("Error listing menu suggestions: userID=%s, error=%v", userID, err)
		http.Error(w, "メニューサジェストの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	responses := make([]menuSuggestionResponse, len(items))
	for i, item := range items {
		responses[i] = toMenuSuggestionResponse(item)
	}

	data, err := json.Marshal(suggestionsListResponse{Suggestions: responses})
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

// HandleGet はGET /api/menu/suggestions/{id} を処理する（レシピ遅延生成付き）
func (h *MenuSuggestionHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	suggestionID, err := extractMenuSuggestionID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	suggestion, err := h.menuRepo.GetByID(r.Context(), userID, suggestionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "メニューサジェストが見つかりません", http.StatusNotFound)
			return
		}
		log.Printf("Error getting menu suggestion: userID=%s, id=%s, error=%v", userID, suggestionID, err)
		http.Error(w, "メニューサジェストの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// レシピが未生成の場合、Gemini で生成して保存する（遅延生成）
	if suggestion.Recipe == "" {
		recipe, err := h.generateAndSaveRecipe(r, userID, suggestion)
		if err != nil {
			// レシピ生成失敗はノンクリティカル - エラーをログに記録してレシピなしで返す
			log.Printf("Error generating recipe: userID=%s, suggestionID=%s, error=%v", userID, suggestionID, err)
		} else {
			suggestion.Recipe = recipe
		}
	}

	data, err := json.Marshal(toMenuSuggestionResponse(*suggestion))
	if err != nil {
		log.Printf("Error encoding response: userID=%s, id=%s, error=%v", userID, suggestionID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("Error writing response: userID=%s, id=%s, error=%v", userID, suggestionID, err)
	}
}

// HandleAccept はPOST /api/menu/suggestions/{id}/accept を処理する
func (h *MenuSuggestionHandler) HandleAccept(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	suggestionID, err := extractMenuSuggestionIDFromSubPath(r.URL.Path, "accept")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.menuRepo.Accept(r.Context(), userID, suggestionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "メニューサジェストが見つかりません", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "既に処理済み") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		log.Printf("Error accepting menu suggestion: userID=%s, id=%s, error=%v", userID, suggestionID, err)
		http.Error(w, "サジェストの採用に失敗しました", http.StatusInternalServerError)
		return
	}

	deducted := make([]deductedIngredientResp, len(result.DeductedIngredients))
	for i, d := range result.DeductedIngredients {
		deducted[i] = deductedIngredientResp{
			IngredientID: d.IngredientID,
			Name:         d.Name,
			Deducted:     d.Deducted,
			Remaining:    d.Remaining,
		}
	}

	data, err := json.Marshal(acceptResponse{
		AnalysisRequestID:   result.AnalysisRequestID,
		DeductedIngredients: deducted,
	})
	if err != nil {
		log.Printf("Error encoding response: userID=%s, id=%s, error=%v", userID, suggestionID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("Error writing response: userID=%s, id=%s, error=%v", userID, suggestionID, err)
	}
}

// HandleDismiss はPOST /api/menu/suggestions/{id}/dismiss を処理する
func (h *MenuSuggestionHandler) HandleDismiss(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	suggestionID, err := extractMenuSuggestionIDFromSubPath(r.URL.Path, "dismiss")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.menuRepo.Dismiss(r.Context(), userID, suggestionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "メニューサジェストが見つかりません", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "既に処理済み") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		log.Printf("Error dismissing menu suggestion: userID=%s, id=%s, error=%v", userID, suggestionID, err)
		http.Error(w, "サジェストの却下に失敗しました", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// buildSuggestionInput はGemini に渡すコンテキストを収集する
func (h *MenuSuggestionHandler) buildSuggestionInput(r *http.Request, userID, mealType string, count int) (*gemini.MenuSuggestionInput, error) {
	ctx := r.Context()
	input := &gemini.MenuSuggestionInput{
		MealType: mealType,
		Count:    count,
	}

	// 食材一覧を取得
	ingredients, err := h.ingredientRepo.List(ctx, userID, "")
	if err != nil {
		return nil, fmt.Errorf("食材一覧の取得に失敗: %w", err)
	}
	input.Ingredients = make([]gemini.IngredientContext, len(ingredients))
	for i, ing := range ingredients {
		input.Ingredients[i] = gemini.IngredientContext{
			ID:         ing.ID,
			Name:       ing.Name,
			Quantity:   ing.Quantity,
			Unit:       ing.Unit,
			ExpiryDate: ing.ExpiryDate,
		}
	}

	// 栄養目標を取得（未設定でも処理を継続）
	nutritionGoal, err := h.nutritionGoalRepo.GetGoal(ctx, userID, nil, nil)
	if err == nil && nutritionGoal != nil {
		input.NutritionGoal = &gemini.NutritionGoalContext{
			TargetCalories:      nutritionGoal.TargetCalories,
			TargetProtein:       nutritionGoal.TargetProtein,
			TargetFat:           nutritionGoal.TargetFat,
			TargetCarbohydrates: nutritionGoal.TargetCarbohydrates,
			Phase:               string(nutritionGoal.Phase),
		}
	}

	// 直近7日間の食事履歴を取得
	historyItems, _, err := h.analysisRepo.GetHistoryList(ctx, userID, 1, 100)
	if err == nil {
		sevenDaysAgo := time.Now().AddDate(0, 0, -7)
		input.RecentMeals = make([]gemini.RecentMealContext, 0)
		for _, item := range historyItems {
			if item.MealDate.Before(sevenDaysAgo) {
				continue
			}
			name := item.InputText
			if name == "" {
				name = string(item.InputType)
			}
			input.RecentMeals = append(input.RecentMeals, gemini.RecentMealContext{
				Date:          item.MealDate.Format("2006-01-02"),
				MealType:      item.MealType,
				Name:          name,
				TotalCalories: item.TotalCalories,
				TotalProtein:  item.TotalProtein,
				TotalFat:      item.TotalFat,
				TotalCarbs:    item.TotalCarbohydrates,
			})
		}
	}

	// 直近30日間の体重推移を取得
	to := time.Now()
	from := to.AddDate(0, 0, -30)
	weightRecords, err := h.weightRecordRepo.ListRecordsByDateRange(ctx, userID, from, to)
	if err == nil {
		input.WeightTrend = make([]gemini.WeightTrendContext, len(weightRecords))
		for i, w := range weightRecords {
			input.WeightTrend[i] = gemini.WeightTrendContext{
				Date:   w.RecordedAt.Format("2006-01-02"),
				Weight: w.WeightKg,
			}
		}
	}

	return input, nil
}

// buildCreateInput はGeminiレスポンスからCreateMenuSuggestionInputを構築する
// 食材名マッチングによりingredientIDを補完する
func (h *MenuSuggestionHandler) buildCreateInput(
	gs gemini.GeminiMenuSuggestion,
	mealType string,
	availableIngredients []gemini.IngredientContext,
) repository.CreateMenuSuggestionInput {
	// 食材名 → IDのマッピングを構築
	nameToID := make(map[string]string, len(availableIngredients))
	for _, ing := range availableIngredients {
		nameToID[ing.Name] = ing.ID
	}

	ings := make([]repository.MenuSuggestionIngredient, len(gs.Ingredients))
	for i, gIng := range gs.Ingredients {
		ings[i] = repository.MenuSuggestionIngredient{
			IngredientID: nameToID[gIng.Name], // マッチしない場合は空文字
			Name:         gIng.Name,
			Quantity:     gIng.Quantity,
			Unit:         gIng.Unit,
		}
	}

	return repository.CreateMenuSuggestionInput{
		Title:           gs.Title,
		Description:     gs.Description,
		Reason:          gs.Reason,
		IngredientsUsed: ings,
		EstimatedNutrition: repository.EstimatedNutrition{
			Calories:      gs.EstimatedNutrition.Calories,
			Protein:       gs.EstimatedNutrition.Protein,
			Fat:           gs.EstimatedNutrition.Fat,
			Carbohydrates: gs.EstimatedNutrition.Carbohydrates,
		},
		MealType: mealType,
	}
}

// generateAndSaveRecipe はレシピを生成してFirestoreに保存する
func (h *MenuSuggestionHandler) generateAndSaveRecipe(r *http.Request, userID string, suggestion *repository.MenuSuggestion) (string, error) {
	geminiIngs := make([]gemini.GeminiIngredient, len(suggestion.IngredientsUsed))
	for i, ing := range suggestion.IngredientsUsed {
		geminiIngs[i] = gemini.GeminiIngredient{
			Name:     ing.Name,
			Quantity: ing.Quantity,
			Unit:     ing.Unit,
		}
	}

	recipe, err := h.suggester.GenerateRecipe(r.Context(), suggestion.Title, geminiIngs)
	if err != nil {
		return "", fmt.Errorf("レシピ生成に失敗: %w", err)
	}

	if err := h.menuRepo.UpdateRecipe(r.Context(), userID, suggestion.ID, recipe); err != nil {
		return "", fmt.Errorf("レシピの保存に失敗: %w", err)
	}

	return recipe, nil
}

// validateMenuSuggestionMealType は食事タイプのバリデーション
func validateMenuSuggestionMealType(mealType string) error {
	valid := map[string]bool{
		"breakfast": true,
		"lunch":     true,
		"dinner":    true,
		"snack":     true,
	}
	if !valid[mealType] {
		return fmt.Errorf("mealTypeはbreakfast, lunch, dinner, snackのいずれかである必要があります")
	}
	return nil
}

// validateSuggestionStatus はサジェストステータスのバリデーション
func validateSuggestionStatus(s string) error {
	valid := map[string]bool{
		string(repository.MenuStatusSuggested): true,
		string(repository.MenuStatusAccepted):  true,
		string(repository.MenuStatusDismissed): true,
	}
	if !valid[s] {
		return fmt.Errorf("statusはsuggested, accepted, dismissedのいずれかである必要があります")
	}
	return nil
}

// extractMenuSuggestionID はURLパスからサジェストIDを抽出する
// /api/menu/suggestions/{id} の形式を想定
func extractMenuSuggestionID(path string) (string, error) {
	normalized := strings.TrimSuffix(path, "/")
	parts := strings.Split(normalized, "/")
	if len(parts) < 5 || parts[len(parts)-1] == "" {
		return "", fmt.Errorf("サジェストIDが指定されていません")
	}
	id := parts[len(parts)-1]
	if id == "" {
		return "", fmt.Errorf("サジェストIDが指定されていません")
	}
	return id, nil
}

// extractMenuSuggestionIDFromSubPath はサブパス付きURLからサジェストIDを抽出する
// /api/menu/suggestions/{id}/accept や /api/menu/suggestions/{id}/dismiss の形式を想定
func extractMenuSuggestionIDFromSubPath(path string, subPath string) (string, error) {
	normalized := strings.TrimSuffix(path, "/")
	suffix := "/" + subPath
	if !strings.HasSuffix(normalized, suffix) {
		return "", fmt.Errorf("不正なパスです")
	}
	withoutSuffix := strings.TrimSuffix(normalized, suffix)
	parts := strings.Split(withoutSuffix, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("サジェストIDが指定されていません")
	}
	id := parts[len(parts)-1]
	if id == "" {
		return "", fmt.Errorf("サジェストIDが指定されていません")
	}
	return id, nil
}

// toMenuSuggestionResponse はMenuSuggestionをレスポンスに変換する
func toMenuSuggestionResponse(s repository.MenuSuggestion) menuSuggestionResponse {
	ings := make([]menuSuggestionIngredientResponse, len(s.IngredientsUsed))
	for i, ing := range s.IngredientsUsed {
		ings[i] = menuSuggestionIngredientResponse{
			IngredientID: ing.IngredientID,
			Name:         ing.Name,
			Quantity:     ing.Quantity,
			Unit:         ing.Unit,
		}
	}
	return menuSuggestionResponse{
		ID:              s.ID,
		Title:           s.Title,
		Description:     s.Description,
		Reason:          s.Reason,
		IngredientsUsed: ings,
		Recipe:          s.Recipe,
		EstimatedNutrition: estimatedNutritionResponse{
			Calories:      s.EstimatedNutrition.Calories,
			Protein:       s.EstimatedNutrition.Protein,
			Fat:           s.EstimatedNutrition.Fat,
			Carbohydrates: s.EstimatedNutrition.Carbohydrates,
		},
		MealType:  s.MealType,
		Status:    s.Status,
		CreatedAt: s.CreatedAt.Format(time.RFC3339),
	}
}
