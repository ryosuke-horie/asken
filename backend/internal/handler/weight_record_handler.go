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
	"github.com/ryosuke-horie/uchikomi/backend/internal/util"
)

// WeightRecordHandler は体重記録 CRUD エンドポイントのハンドラー
type WeightRecordHandler struct {
	repository     repository.WeightRecordRepository
	goalRepository repository.WeightGoalRepository
}

// NewWeightRecordHandler は新しいWeightRecordHandlerを作成
func NewWeightRecordHandler(repo repository.WeightRecordRepository, goalRepo repository.WeightGoalRepository) *WeightRecordHandler {
	if repo == nil || goalRepo == nil {
		panic("weight record handler: repositories must not be nil")
	}
	return &WeightRecordHandler{repository: repo, goalRepository: goalRepo}
}

// CreateWeightRecordRequest は体重記録作成リクエスト
type CreateWeightRecordRequest struct {
	WeightKg   float64 `json:"weight_kg"`
	RecordedAt string  `json:"recorded_at"`
	Note       string  `json:"note"`
}

// WeightRecordResponse は体重記録のレスポンス
type WeightRecordResponse struct {
	ID         string  `json:"id"`
	WeightKg   float64 `json:"weight_kg"`
	RecordedAt string  `json:"recorded_at"`
	Note       string  `json:"note,omitempty"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// DailySummary は日別サマリー
type DailySummary struct {
	LatestWeight float64 `json:"latest_weight"`
	Count        int     `json:"count"`
}

// WeightRecordsListResponse は体重記録一覧のレスポンス
type WeightRecordsListResponse struct {
	Records      []WeightRecordResponse  `json:"records"`
	DailySummary map[string]DailySummary `json:"daily_summary"`
	Goal         *WeightGoalResponse     `json:"goal"`
}

// UpdateWeightRecordRequest は体重記録更新リクエスト
type UpdateWeightRecordRequest struct {
	WeightKg float64 `json:"weight_kg"`
	Note     string  `json:"note"`
}

// dateRangeParams は日付範囲クエリパラメータの解析結果
type dateRangeParams struct {
	from time.Time
	to   time.Time
	loc  *time.Location
}

// parseDateRangeParams はクエリパラメータの日付範囲を解析・検証する。
// エラー時はHTTPレスポンスを書き込みエラーを返す。
func parseDateRangeParams(w http.ResponseWriter, r *http.Request, userID string) (*dateRangeParams, error) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	tz := r.URL.Query().Get("tz")
	if tz == "" {
		tz = "UTC"
	}

	if fromStr == "" || toStr == "" {
		http.Error(w, "fromとtoパラメータが必要です", http.StatusBadRequest)
		return nil, fmt.Errorf("missing from/to parameters")
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		http.Error(w, fmt.Sprintf("タイムゾーンが不正です: %s", tz), http.StatusBadRequest)
		return nil, err
	}

	from, err := util.ParseDateInTimezone(fromStr, tz)
	if err != nil {
		log.Printf("Error parsing from parameter: userID=%s, from=%s, error=%v", userID, fromStr, err)
		http.Error(w, "fromパラメータが不正です。YYYY-MM-DD形式で指定してください", http.StatusBadRequest)
		return nil, err
	}

	_, toEnd, err := util.GetDayRangeInTimezone(toStr, tz)
	if err != nil {
		log.Printf("Error parsing to parameter: userID=%s, to=%s, error=%v", userID, toStr, err)
		http.Error(w, "toパラメータが不正です。YYYY-MM-DD形式で指定してください", http.StatusBadRequest)
		return nil, err
	}

	if toEnd.Sub(from) > 366*24*time.Hour {
		http.Error(w, "取得期間は最大366日です", http.StatusBadRequest)
		return nil, fmt.Errorf("date range exceeds 366 days")
	}

	return &dateRangeParams{from: from, to: toEnd, loc: loc}, nil
}

// HandleList はGET /api/weight/recordsリクエストを処理
func (h *WeightRecordHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	params, err := parseDateRangeParams(w, r, userID)
	if err != nil {
		return
	}

	records, err := h.repository.ListRecordsByDateRange(r.Context(), userID, params.from, params.to)
	if err != nil {
		log.Printf("Error listing weight records: userID=%s, error=%v", userID, err)
		http.Error(w, "体重記録の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// 目標体重を取得（取得失敗時はnilのまま続行）
	goal, err := h.goalRepository.GetGoal(r.Context(), userID)
	if err != nil {
		log.Printf("Warning: failed to get weight goal, continuing without it: userID=%s, error=%v", userID, err)
	}

	// レスポンスを構築
	responseRecords := make([]WeightRecordResponse, 0, len(records))
	dailySummary := map[string]DailySummary{}

	for _, record := range records {
		responseRecords = append(responseRecords, toWeightRecordResponse(record))

		dateKey := record.RecordedAt.In(params.loc).Format("2006-01-02")
		summary := dailySummary[dateKey]
		summary.Count++
		summary.LatestWeight = record.WeightKg // recordedAtで昇順ソート済みなので最後が最新
		dailySummary[dateKey] = summary
	}

	var goalResponse *WeightGoalResponse
	if goal != nil {
		goalResponse = &WeightGoalResponse{
			TargetWeightKg: goal.TargetWeightKg,
			UpdatedAt:      goal.UpdatedAt.Format(time.RFC3339),
		}
	}

	response := WeightRecordsListResponse{
		Records:      responseRecords,
		DailySummary: dailySummary,
		Goal:         goalResponse,
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

// HandleCreate はPOST /api/weight/recordsリクエストを処理
func (h *WeightRecordHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1024)

	var req CreateWeightRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: userID=%s, error=%v", userID, err)
		http.Error(w, "リクエストのパースに失敗しました", http.StatusBadRequest)
		return
	}

	// バリデーション
	if err := validateWeightKg(req.WeightKg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.RecordedAt == "" {
		http.Error(w, "recorded_atは必須です", http.StatusBadRequest)
		return
	}

	recordedAt, err := time.Parse(time.RFC3339, req.RecordedAt)
	if err != nil {
		http.Error(w, "recorded_atはRFC3339形式で指定してください", http.StatusBadRequest)
		return
	}

	// 未来日時チェック（5分まで許容）
	if recordedAt.After(time.Now().Add(5 * time.Minute)) {
		http.Error(w, "未来の日時は指定できません", http.StatusBadRequest)
		return
	}

	if err := validateNote(req.Note); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	record, err := h.repository.CreateRecord(r.Context(), userID, req.WeightKg, recordedAt, req.Note)
	if err != nil {
		log.Printf("Error creating weight record: userID=%s, error=%v", userID, err)
		http.Error(w, "体重記録の作成に失敗しました", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(toWeightRecordResponse(*record))
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

// HandleGet はGET /api/weight/records/{id}リクエストを処理
func (h *WeightRecordHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	recordID, err := extractRecordID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	record, err := h.repository.GetRecord(r.Context(), userID, recordID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "体重記録が見つかりません", http.StatusNotFound)
			return
		}
		log.Printf("Error getting weight record: userID=%s, recordID=%s, error=%v", userID, recordID, err)
		http.Error(w, "体重記録の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(toWeightRecordResponse(*record))
	if err != nil {
		log.Printf("Error encoding response: userID=%s, recordID=%s, error=%v", userID, recordID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("Error writing response: userID=%s, recordID=%s, error=%v", userID, recordID, err)
	}
}

// HandleUpdate はPUT /api/weight/records/{id}リクエストを処理
func (h *WeightRecordHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	recordID, err := extractRecordID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1024)

	var req UpdateWeightRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: userID=%s, error=%v", userID, err)
		http.Error(w, "リクエストのパースに失敗しました", http.StatusBadRequest)
		return
	}

	if err := validateWeightKg(req.WeightKg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateNote(req.Note); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	record, err := h.repository.UpdateRecord(r.Context(), userID, recordID, req.WeightKg, req.Note)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "体重記録が見つかりません", http.StatusNotFound)
			return
		}
		log.Printf("Error updating weight record: userID=%s, recordID=%s, error=%v", userID, recordID, err)
		http.Error(w, "体重記録の更新に失敗しました", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(toWeightRecordResponse(*record))
	if err != nil {
		log.Printf("Error encoding response: userID=%s, recordID=%s, error=%v", userID, recordID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("Error writing response: userID=%s, recordID=%s, error=%v", userID, recordID, err)
	}
}

// HandleDelete はDELETE /api/weight/records/{id}リクエストを処理
func (h *WeightRecordHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	recordID, err := extractRecordID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.repository.DeleteRecord(r.Context(), userID, recordID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "体重記録が見つかりません", http.StatusNotFound)
			return
		}
		log.Printf("Error deleting weight record: userID=%s, recordID=%s, error=%v", userID, recordID, err)
		http.Error(w, "体重記録の削除に失敗しました", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// validateWeightKg は体重値のバリデーション
func validateWeightKg(weightKg float64) error {
	return repository.ValidateWeightKg(weightKg)
}

// validateNote はメモのバリデーション
func validateNote(note string) error {
	if utf8.RuneCountInString(note) > 200 {
		return fmt.Errorf("noteは200文字以内で指定してください")
	}
	return nil
}

// extractRecordID はURLパスから記録IDを抽出しUUID形式を検証する
func extractRecordID(path string) (string, error) {
	// /api/weight/records/{id} からIDを抽出
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(parts) < 4 || parts[len(parts)-1] == "" {
		return "", fmt.Errorf("記録IDが指定されていません")
	}
	id := parts[len(parts)-1]
	if _, err := uuid.Parse(id); err != nil {
		return "", fmt.Errorf("記録IDの形式が不正です")
	}
	return id, nil
}

// roundToOneDecimalForJSON はレスポンス生成用に小数点1桁に丸める
// Firestoreの浮動小数点誤差に対する防御的な丸め処理
// （保存時の丸めはrepository層のroundToOneDecimalが担当）
func roundToOneDecimalForJSON(v float64) float64 {
	return math.Round(v*10) / 10
}

// toWeightRecordResponse はWeightRecordをレスポンスに変換
func toWeightRecordResponse(record repository.WeightRecord) WeightRecordResponse {
	return WeightRecordResponse{
		ID:         record.ID,
		WeightKg:   roundToOneDecimalForJSON(record.WeightKg),
		RecordedAt: record.RecordedAt.Format(time.RFC3339),
		Note:       record.Note,
		CreatedAt:  record.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  record.UpdatedAt.Format(time.RFC3339),
	}
}
