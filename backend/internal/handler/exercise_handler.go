package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
)

// ExerciseService は運動記録サービスのインターフェース
type ExerciseService interface {
	CreateExerciseRecord(ctx context.Context, userID string, input service.CreateExerciseInput, recordedDate string) (*repository.ExerciseRecord, error)
	GetDailyExercise(ctx context.Context, userID string, recordedDate string) (*repository.ExerciseDailyResult, error)
	DeleteExerciseRecord(ctx context.Context, userID string, recordID string) error
}

// ExerciseHandler は運動記録 CRUD エンドポイントのハンドラー
type ExerciseHandler struct {
	service ExerciseService
}

// NewExerciseHandler は新しいExerciseHandlerを作成
func NewExerciseHandler(svc ExerciseService) *ExerciseHandler {
	if svc == nil {
		panic("exercise handler: service must not be nil")
	}
	return &ExerciseHandler{service: svc}
}

// CreateExerciseRecordRequest はPOST /api/exercise/records のリクエスト
type CreateExerciseRecordRequest struct {
	ExerciseName    string `json:"exercise_name"`
	DurationMinutes int    `json:"duration_minutes"`
	RecordedDate    string `json:"recorded_date"`
}

// ExerciseRecordResponse は運動記録のレスポンス
type ExerciseRecordResponse struct {
	ID                 string  `json:"id"`
	ExerciseName       string  `json:"exercise_name"`
	DurationMinutes    int     `json:"duration_minutes"`
	BurnedCaloriesKcal float64 `json:"burned_calories_kcal"`
	EstimationMethod   string  `json:"estimation_method"`
	RecordedDate       string  `json:"recorded_date"`
	CreatedAt          string  `json:"created_at"`
}

// DailyExerciseResponse はGET /api/exercise/daily のレスポンス
type DailyExerciseResponse struct {
	Records                 []ExerciseRecordResponse `json:"records"`
	TotalBurnedCaloriesKcal float64                  `json:"total_burned_calories_kcal"`
}

// HandleCreate はPOST /api/exercise/records リクエストを処理
func (h *ExerciseHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
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
	var req CreateExerciseRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ExerciseHandler.HandleCreate: decode error userID=%s error=%v", userID, err)
		http.Error(w, "リクエストのパースに失敗しました", http.StatusBadRequest)
		return
	}

	if req.RecordedDate == "" {
		http.Error(w, "recorded_dateは必須です", http.StatusBadRequest)
		return
	}

	record, err := h.service.CreateExerciseRecord(r.Context(), userID, service.CreateExerciseInput{
		ExerciseName:    req.ExerciseName,
		DurationMinutes: req.DurationMinutes,
	}, req.RecordedDate)
	if err != nil {
		var validErr *service.ValidationError
		if errors.As(err, &validErr) {
			http.Error(w, validErr.Message, http.StatusBadRequest)
		} else {
			log.Printf("ExerciseHandler.HandleCreate: service error userID=%s error=%v", userID, err)
			http.Error(w, "運動記録の作成に失敗しました", http.StatusInternalServerError)
		}
		return
	}

	data, err := json.Marshal(toExerciseRecordResponse(*record))
	if err != nil {
		log.Printf("ExerciseHandler.HandleCreate: marshal error userID=%s error=%v", userID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(data); err != nil {
		log.Printf("ExerciseHandler.HandleCreate: write error userID=%s error=%v", userID, err)
	}
}

// HandleListByDate はGET /api/exercise/daily?date=YYYY-MM-DD リクエストを処理
func (h *ExerciseHandler) HandleListByDate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, "dateパラメータは必須です", http.StatusBadRequest)
		return
	}

	result, err := h.service.GetDailyExercise(r.Context(), userID, date)
	if err != nil {
		log.Printf("ExerciseHandler.HandleListByDate: service error userID=%s date=%s error=%v", userID, date, err)
		http.Error(w, "運動記録の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	records := make([]ExerciseRecordResponse, 0, len(result.Records))
	for _, rec := range result.Records {
		records = append(records, toExerciseRecordResponse(rec))
	}

	response := DailyExerciseResponse{
		Records:                 records,
		TotalBurnedCaloriesKcal: result.TotalBurnedCaloriesKcal,
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("ExerciseHandler.HandleListByDate: marshal error userID=%s error=%v", userID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("ExerciseHandler.HandleListByDate: write error userID=%s error=%v", userID, err)
	}
}

// HandleDelete はDELETE /api/exercise/records/{id} リクエストを処理
func (h *ExerciseHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	recordID := extractExerciseRecordID(r.URL.Path)
	if recordID == "" {
		http.Error(w, "記録IDが指定されていません", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteExerciseRecord(r.Context(), userID, recordID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "運動記録が見つかりません", http.StatusNotFound)
			return
		}
		log.Printf("ExerciseHandler.HandleDelete: service error userID=%s recordID=%s error=%v", userID, recordID, err)
		http.Error(w, "運動記録の削除に失敗しました", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// extractExerciseRecordID はURLパスから記録IDを抽出する
// /api/exercise/records/{id} の形式を想定
func extractExerciseRecordID(path string) string {
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

// toExerciseRecordResponse はExerciseRecordをレスポンスに変換する
func toExerciseRecordResponse(record repository.ExerciseRecord) ExerciseRecordResponse {
	return ExerciseRecordResponse{
		ID:                 record.ID,
		ExerciseName:       record.ExerciseName,
		DurationMinutes:    record.DurationMinutes,
		BurnedCaloriesKcal: record.BurnedCaloriesKcal,
		EstimationMethod:   string(record.EstimationMethod),
		RecordedDate:       record.RecordedDate,
		CreatedAt:          record.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
