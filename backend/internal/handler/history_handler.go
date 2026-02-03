package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
)

// HistoryHandler は履歴取得エンドポイントのハンドラー
type HistoryHandler struct {
	repository  repository.AnalysisRepository
	foodService service.FoodServiceInterface
}

// NewHistoryHandler は新しいHistoryHandlerを作成
func NewHistoryHandler(repository repository.AnalysisRepository, foodService service.FoodServiceInterface) *HistoryHandler {
	return &HistoryHandler{
		repository:  repository,
		foodService: foodService,
	}
}

// HandleList はGET /api/historyリクエストを処理（履歴一覧取得）
func (h *HistoryHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received history list request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

	// GETメソッドのみ許可
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// contextからユーザーIDを取得
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	// クエリパラメータからpage, limitを取得
	page := 1
	limit := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	log.Printf("Fetching history list: userID=%s, page=%d, limit=%d", userID, page, limit)

	// リポジトリから履歴一覧を取得（userIDでスコープ）
	items, total, err := h.repository.GetHistoryList(r.Context(), userID, page, limit)
	if err != nil {
		log.Printf("Error getting history list: %v", err)
		http.Error(w, "Failed to get history list", http.StatusInternalServerError)
		return
	}

	log.Printf("Retrieved %d items, total=%d", len(items), total)

	// レスポンスを生成
	response := map[string]interface{}{
		"items": items,
		"total": total,
		"page":  page,
		"limit": limit,
	}

	// JSONレスポンスを返却
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("History list response sent successfully")
}

// HandleDetail はGET /api/history/:idリクエストを処理（履歴詳細取得）
func (h *HistoryHandler) HandleDetail(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received history detail request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

	// GETメソッドのみ許可
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// contextからユーザーIDを取得
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	// URLからhistory_idを抽出
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		log.Printf("Invalid URL path: %s", r.URL.Path)
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	historyIDStr := pathParts[3]
	historyID, err := uuid.Parse(historyIDStr)
	if err != nil {
		log.Printf("Invalid UUID: %s, error: %v", historyIDStr, err)
		http.Error(w, "Invalid history ID", http.StatusBadRequest)
		return
	}

	log.Printf("Getting history detail for ID: %s, userID: %s", historyID, userID)

	// リポジトリから履歴詳細を取得（userIDでスコープ）
	detail, err := h.repository.GetHistoryDetail(r.Context(), userID, historyID)
	if err != nil {
		log.Printf("Error getting history detail: %v", err)
		if strings.Contains(err.Error(), "見つかりません") {
			http.Error(w, "History not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get history detail", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("History detail retrieved successfully")

	// JSONレスポンスを返却
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(detail); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("History detail response sent successfully for ID: %s", historyID)
}

// UpdateFoodItem は更新リクエストの食材アイテム
type UpdateFoodItem struct {
	Name            string  `json:"name"`
	EstimatedAmount string  `json:"estimated_amount"`
	Calories        float64 `json:"calories_kcal"`
	Protein         float64 `json:"protein_g"`
	Fat             float64 `json:"fat_g"`
	Carbohydrates   float64 `json:"carbohydrates_g"`
}

// UpdateHistoryRequest は履歴更新リクエストの構造体
type UpdateHistoryRequest struct {
	Foods []UpdateFoodItem `json:"foods"`
}

// HandleUpdate はPUT /api/history/:idリクエストを処理（履歴更新）
func (h *HistoryHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received history update request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

	// PUTメソッドのみ許可
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// contextからユーザーIDを取得
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	// URLからhistory_idを抽出
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		log.Printf("Invalid URL path: %s", r.URL.Path)
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	historyIDStr := pathParts[3]
	historyID, err := uuid.Parse(historyIDStr)
	if err != nil {
		log.Printf("Invalid UUID: %s, error: %v", historyIDStr, err)
		http.Error(w, "Invalid history ID", http.StatusBadRequest)
		return
	}

	// リクエストボディをパース
	var req UpdateHistoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Updating history for ID: %s, userID: %s, with %d foods", historyID, userID, len(req.Foods))

	// リクエストをNutritionInfo形式に変換（現在の値をそのまま保存）
	foods := make([]gemini.NutritionInfo, len(req.Foods))
	for i, f := range req.Foods {
		foods[i] = gemini.NutritionInfo{
			Name:            f.Name,
			EstimatedAmount: f.EstimatedAmount,
			Calories:        f.Calories,
			Protein:         f.Protein,
			Fat:             f.Fat,
			Carbohydrates:   f.Carbohydrates,
		}
	}

	// リポジトリで更新（userIDでスコープ、まず現在の値で保存）
	if err := h.repository.UpdateResult(r.Context(), userID, historyID, foods); err != nil {
		log.Printf("Error updating history: %v", err)
		if strings.Contains(err.Error(), "見つかりません") {
			http.Error(w, "History not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update history", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("History updated successfully for ID: %s", historyID)

	// 非同期で栄養素を再計算（userIDを渡す）
	if h.foodService != nil {
		go h.recalculateNutritionAsync(userID, historyID, req.Foods)
	}

	// 更新後の詳細を取得して返却（userIDでスコープ）
	detail, err := h.repository.GetHistoryDetail(r.Context(), userID, historyID)
	if err != nil {
		log.Printf("Error getting updated history detail: %v", err)
		http.Error(w, "Failed to get updated history", http.StatusInternalServerError)
		return
	}

	// JSONレスポンスを返却
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(detail); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("History update response sent successfully for ID: %s (nutrition recalculation started in background)", historyID)
}

// recalculateNutritionAsync は非同期で栄養素を再計算する
// タイムアウト: 60秒（Gemini API呼び出し + DB更新）
func (h *HistoryHandler) recalculateNutritionAsync(userID string, historyID uuid.UUID, foods []UpdateFoodItem) {
	log.Printf("Starting async nutrition recalculation for history ID: %s, userID: %s", historyID, userID)

	// タイムアウト付きのcontextを作成（60秒）
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 食材リストをテキスト形式に変換
	var foodDescriptions []string
	for _, f := range foods {
		if f.Name != "" {
			desc := f.Name
			if f.EstimatedAmount != "" {
				desc += " " + f.EstimatedAmount
			}
			foodDescriptions = append(foodDescriptions, desc)
		}
	}

	if len(foodDescriptions) == 0 {
		log.Printf("No valid foods to recalculate for history ID: %s", historyID)
		return
	}

	// 食材リストを1つのテキストに結合
	inputText := strings.Join(foodDescriptions, ", ")
	log.Printf("Recalculating nutrition for: %s", inputText)

	// Geminiで栄養素を計算
	result, err := h.foodService.AnalyzeFoodText(ctx, inputText)
	if err != nil {
		log.Printf("Error recalculating nutrition for history ID %s: %v", historyID, err)
		return
	}

	// 計算結果をNutritionInfo形式に変換
	nutritionInfos := make([]gemini.NutritionInfo, len(result.Foods))
	for i, f := range result.Foods {
		nutritionInfos[i] = gemini.NutritionInfo{
			Name:            f.Name,
			EstimatedAmount: f.EstimatedAmount,
			Calories:        f.Calories,
			Protein:         f.Protein,
			Fat:             f.Fat,
			Carbohydrates:   f.Carbohydrates,
		}
	}

	// データベースを更新（userIDでスコープ）
	if err := h.repository.UpdateResult(ctx, userID, historyID, nutritionInfos); err != nil {
		log.Printf("Error saving recalculated nutrition for history ID %s: %v", historyID, err)
		return
	}

	log.Printf("Nutrition recalculation completed for history ID: %s (%d foods, total %.0f kcal)",
		historyID, len(nutritionInfos), result.TotalCalories)
}
