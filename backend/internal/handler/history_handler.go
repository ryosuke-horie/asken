package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
)

const recalculateTimeout = 120 * time.Second

// リトライ間隔（初回失敗時に最大1回再試行、計2回まで）:
// Gemini APIはレート制限を考慮して長め、Firestoreは一時的障害向けに短め。
// 持続的なレート制限や長時間障害には対応していない。
const geminiRetryDelay = 2 * time.Second
const firestoreRetryDelay = 1 * time.Second

// NutritionRecalculator は栄養素再計算のインターフェース
type NutritionRecalculator interface {
	CalculateNutrition(ctx context.Context, foods []gemini.FoodItem) ([]gemini.NutritionInfo, error)
}

// HistoryHandler は履歴の取得・更新エンドポイントのハンドラー
type HistoryHandler struct {
	repository   repository.AnalysisRepository
	recalculator NutritionRecalculator
}

// NewHistoryHandler は新しいHistoryHandlerを作成
func NewHistoryHandler(repository repository.AnalysisRepository, recalculator NutritionRecalculator) *HistoryHandler {
	return &HistoryHandler{
		repository:   repository,
		recalculator: recalculator,
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
		log.Printf("Authentication failed for %s: %s %s - no Firebase UID in context", r.RemoteAddr, r.Method, r.URL.Path)
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
		log.Printf("Authentication failed for %s: %s %s - no Firebase UID in context", r.RemoteAddr, r.Method, r.URL.Path)
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
		if errors.Is(err, repository.ErrNotFound) {
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

// Validate はUpdateFoodItemのバリデーションを行う
func (f UpdateFoodItem) Validate() error {
	if strings.TrimSpace(f.Name) == "" {
		return fmt.Errorf("food name is required")
	}
	if strings.TrimSpace(f.EstimatedAmount) == "" {
		return fmt.Errorf("estimated amount is required")
	}
	if f.Calories < 0 || f.Protein < 0 || f.Fat < 0 || f.Carbohydrates < 0 {
		return fmt.Errorf("nutrition values must be non-negative")
	}
	return nil
}

// maxFoodsPerUpdate は1回の更新で許可される食材の最大数
const maxFoodsPerUpdate = 50

// UpdateHistoryRequest は履歴更新リクエストの構造体
type UpdateHistoryRequest struct {
	Foods []UpdateFoodItem `json:"foods"`
}

// Validate はUpdateHistoryRequestのバリデーションを行う
func (r UpdateHistoryRequest) Validate() error {
	if len(r.Foods) == 0 {
		return fmt.Errorf("at least one food item is required")
	}
	if len(r.Foods) > maxFoodsPerUpdate {
		return fmt.Errorf("too many food items: maximum %d allowed", maxFoodsPerUpdate)
	}
	for i, f := range r.Foods {
		if err := f.Validate(); err != nil {
			return fmt.Errorf("foods[%d]: %w", i, err)
		}
	}
	return nil
}

// decodeUpdateRequest はリクエストボディをデコード・検証し、UpdateHistoryRequestを返す。
// エラー時はHTTPレスポンスを書き込みfalseを返す。
func decodeUpdateRequest(w http.ResponseWriter, r *http.Request) (*UpdateHistoryRequest, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB: 食材リストを含むため余裕を持たせる

	var req UpdateHistoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			log.Printf("Request body too large: limit=%d", maxBytesErr.Limit)
			http.Error(w, "リクエストボディが大きすぎます", http.StatusRequestEntityTooLarge)
			return nil, false
		}
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return nil, false
	}

	if err := req.Validate(); err != nil {
		log.Printf("Validation error: %v", err)
		http.Error(w, fmt.Sprintf("Validation error: %s", err), http.StatusBadRequest)
		return nil, false
	}

	return &req, true
}

// toNutritionInfoSlice はUpdateFoodItemスライスをNutritionInfoスライスに変換する
func toNutritionInfoSlice(foods []UpdateFoodItem) []gemini.NutritionInfo {
	result := make([]gemini.NutritionInfo, len(foods))
	for i, f := range foods {
		result[i] = gemini.NutritionInfo{
			Name:            f.Name,
			EstimatedAmount: f.EstimatedAmount,
			Calories:        f.Calories,
			Protein:         f.Protein,
			Fat:             f.Fat,
			Carbohydrates:   f.Carbohydrates,
		}
	}
	return result
}

// triggerRecalculationIfNeeded はメニュー名変更を検知し、必要に応じて非同期で栄養素を再計算する
func (h *HistoryHandler) triggerRecalculationIfNeeded(userID string, historyID uuid.UUID, oldFoods, newFoods []gemini.NutritionInfo) {
	if h.recalculator == nil {
		return
	}

	// detectNameChangesは要素数が異なる場合nilを返す（インデックスベース比較が不正確なため）
	changedFoods := detectNameChanges(oldFoods, newFoods)
	if changedFoods == nil {
		if len(oldFoods) != len(newFoods) {
			log.Printf("Skipping async recalculation for history %s: food count changed (old=%d, new=%d)", historyID, len(oldFoods), len(newFoods))
		}
		return
	}

	if len(changedFoods) > 0 {
		log.Printf("Detected %d food name changes, triggering async recalculation for history %s", len(changedFoods), historyID)
		// goroutineに渡す前にスライスをコピー（呼び出し元との共有を防止）
		foodsCopy := make([]gemini.NutritionInfo, len(newFoods))
		copy(foodsCopy, newFoods)
		go h.recalculateAsync(userID, historyID, foodsCopy)
	}
}

// HandleUpdate はPUT /api/history/:idリクエストを処理（履歴更新）
func (h *HistoryHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received history update request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		log.Printf("Authentication failed for %s: %s %s - no Firebase UID in context", r.RemoteAddr, r.Method, r.URL.Path)
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		log.Printf("Invalid URL path: %s", r.URL.Path)
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	historyID, err := uuid.Parse(pathParts[3])
	if err != nil {
		log.Printf("Invalid UUID: %s, error: %v", pathParts[3], err)
		http.Error(w, "Invalid history ID", http.StatusBadRequest)
		return
	}

	req, ok := decodeUpdateRequest(w, r)
	if !ok {
		return
	}

	log.Printf("Updating history for ID: %s, userID: %s, with %d foods", historyID, userID, len(req.Foods))

	oldDetail, err := h.repository.GetHistoryDetail(r.Context(), userID, historyID)
	if err != nil {
		log.Printf("Error getting old history detail for comparison: %v", err)
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "History not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get history detail", http.StatusInternalServerError)
		}
		return
	}

	foods := toNutritionInfoSlice(req.Foods)

	if err := h.repository.UpdateResult(r.Context(), userID, historyID, foods); err != nil {
		log.Printf("Error updating history: %v", err)
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "History not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update history", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("History updated successfully for ID: %s", historyID)

	h.triggerRecalculationIfNeeded(userID, historyID, oldDetail.Foods, foods)

	detail, err := h.repository.GetHistoryDetail(r.Context(), userID, historyID)
	if err != nil {
		log.Printf("Error getting updated history detail: %v", err)
		http.Error(w, "Failed to get updated history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(detail); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("History update response sent successfully for ID: %s", historyID)
}

// detectNameChanges は旧foodsと新foodsをインデックスベースで比較し、名前が変わったインデックスを返す。
// iOSクライアントはfoodsの順序を維持して送信するため、インデックスベースの比較で十分である。
// 要素数が異なる場合（食材の追加・削除時）はインデックスベース比較が不正確になるため、
// nilを返す（比較不能を示し、呼び出し側で適切に処理する）。
func detectNameChanges(oldFoods []gemini.NutritionInfo, newFoods []gemini.NutritionInfo) []int {
	if len(oldFoods) != len(newFoods) {
		return nil
	}

	var changed []int
	for i := 0; i < len(oldFoods); i++ {
		if oldFoods[i].Name != newFoods[i].Name {
			changed = append(changed, i)
		}
	}

	return changed
}

// retryWithDelay は操作を1回リトライする。リトライ不要なエラー（context系）はリトライせず即座にエラーを返す。
// nonRetryableCheck はリトライ不要な追加エラー条件を指定する（nilの場合は context 系のみ判定）。
func retryWithDelay(ctx context.Context, delay time.Duration, nonRetryableCheck func(error) bool, op func() error) error {
	err := op()
	if err == nil {
		return nil
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if nonRetryableCheck != nil && nonRetryableCheck(err) {
		return err
	}

	log.Printf("WARN: Operation failed, retrying after %v: %v", delay, err)

	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return ctx.Err()
	}

	return op()
}

// recalculateAsync はGemini APIで非同期に全食材の栄養素を一括再計算し、結果をFirestoreに保存する。
// 名前が変更された食材だけでなく、全食材をGemini APIに渡して再計算する
// （食材の組み合わせにより栄養素の推定が変わる可能性があるため）。
// goroutineとして起動されるため、panicリカバリを含む。
// 書き込み前に鮮度チェックを行い、再計算中にユーザーが再保存した場合は書き込みをスキップする。
// 注意: 鮮度チェック（GetHistoryDetail）と書き込み（UpdateResult）の間には小さな競合ウィンドウが
// 残る（TOCTOU）。競合が発生した場合、Geminiの再計算結果がユーザーの最新保存を上書きし、
// ユーザーが再度保存するまで不正確な栄養素値が表示される可能性がある。
// 競合時はUpdateResultが全フィールド（名前・量・栄養素）を上書きするため、
// ユーザーの最新編集がすべて失われる可能性がある。ただし鮮度チェックにより
// 競合ウィンドウは非常に小さく、発生確率は低いため、許容範囲とする。
// Firestoreトランザクションの導入はリポジトリインターフェースの変更を伴うため、現時点では見送る。
func (h *HistoryHandler) recalculateAsync(userID string, historyID uuid.UUID, currentFoods []gemini.NutritionInfo) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic recovered in recalculateAsync for history %s: %v\nStack trace:\n%s", historyID, r, debug.Stack())
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), recalculateTimeout)
	defer cancel()

	// 現在の食材リストをFoodItemに変換
	foodItems := make([]gemini.FoodItem, len(currentFoods))
	for i, f := range currentFoods {
		foodItems[i] = gemini.FoodItem{
			Name:            f.Name,
			EstimatedAmount: f.EstimatedAmount,
		}
	}

	log.Printf("Starting async nutrition recalculation for history %s with %d foods", historyID, len(foodItems))

	// Gemini APIで栄養素を再計算（1回リトライ）
	var recalculated []gemini.NutritionInfo
	err := retryWithDelay(ctx, geminiRetryDelay, nil, func() error {
		var calcErr error
		recalculated, calcErr = h.recalculator.CalculateNutrition(ctx, foodItems)
		return calcErr
	})
	if err != nil {
		log.Printf("ERROR: Async nutrition recalculation failed for history %s, userID=%s, foodCount=%d: %v", historyID, userID, len(foodItems), err)
		return
	}

	// 鮮度チェック: 再計算中にユーザーが再保存していないか確認
	if stale, checkErr := h.isFoodsStale(ctx, userID, historyID, currentFoods); checkErr != nil || stale {
		return
	}

	// 再計算結果をFirestoreに保存（1回リトライ、リトライ前に鮮度を再チェック）
	if err := h.saveRecalculatedResult(ctx, userID, historyID, currentFoods, recalculated); err != nil {
		log.Printf("ERROR: Failed to save recalculated nutrition for history %s, userID=%s: %v", historyID, userID, err)
		return
	}

	log.Printf("Async nutrition recalculation completed successfully for history %s", historyID)
}

// saveRecalculatedResult は再計算結果をFirestoreに保存する（1回リトライ）。
// リトライ前に鮮度を再チェックし、ユーザーが再保存した場合はスキップする。
func (h *HistoryHandler) saveRecalculatedResult(ctx context.Context, userID string, historyID uuid.UUID, currentFoods, recalculated []gemini.NutritionInfo) error {
	err := h.repository.UpdateResult(ctx, userID, historyID, recalculated)
	if err == nil {
		return nil
	}
	if errors.Is(err, repository.ErrNotFound) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}

	select {
	case <-time.After(firestoreRetryDelay):
	case <-ctx.Done():
		return ctx.Err()
	}

	// リトライ前に鮮度を再チェック（sleep中にユーザーが再保存した可能性がある）
	if stale, checkErr := h.isFoodsStale(ctx, userID, historyID, currentFoods); checkErr != nil {
		return checkErr
	} else if stale {
		return nil
	}

	return h.repository.UpdateResult(ctx, userID, historyID, recalculated)
}

// isFoodsStale は再計算中にユーザーがデータを変更したか確認する。
// エラーまたはデータが変更されている場合はtrue（古い）を返す。
func (h *HistoryHandler) isFoodsStale(ctx context.Context, userID string, historyID uuid.UUID, expectedFoods []gemini.NutritionInfo) (bool, error) {
	currentDetail, err := h.repository.GetHistoryDetail(ctx, userID, historyID)
	if err != nil {
		log.Printf("ERROR: Staleness check failed for history %s, userID=%s: %v", historyID, userID, err)
		return true, err
	}
	if !foodsMatch(expectedFoods, currentDetail.Foods) {
		log.Printf("Skipping async recalculation for history %s: data modified during recalculation", historyID)
		return true, nil
	}
	return false, nil
}

// foodsMatch は2つの食材リストのName・EstimatedAmountが一致するか確認する。
// 鮮度チェックで使用し、再計算中にデータが変更されていないか検知する。
// 栄養素値（Calories, Protein, Fat, Carbohydrates）は比較対象外。再計算前の値（ユーザー送信値）と
// Firestoreの現在値では、先行する再計算の反映や並行するユーザー編集により
// 栄養素が異なり得る。ユーザーの意図した食材構成が変わっていなければ
// 再計算結果の上書きは安全である。
func foodsMatch(a, b []gemini.NutritionInfo) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name || a[i].EstimatedAmount != b[i].EstimatedAmount {
			return false
		}
	}
	return true
}
