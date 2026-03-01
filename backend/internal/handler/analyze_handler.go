package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/internal/util"
)

// FoodService は食品分析サービスのインターフェース
type FoodService interface {
	AnalyzeFoodImage(ctx context.Context, imagePath string) (*repository.AnalysisResult, error)
	AnalyzeFoodText(ctx context.Context, inputText string) (*repository.AnalysisResult, error)
}

// AnalyzeHandler は画像分析エンドポイントのハンドラー
type AnalyzeHandler struct {
	foodService FoodService
	repository  repository.AnalysisRepository
	storageRepo repository.StorageRepository
}

// NewAnalyzeHandler は新しいAnalyzeHandlerを作成
func NewAnalyzeHandler(foodService FoodService, repository repository.AnalysisRepository, storageRepo repository.StorageRepository) *AnalyzeHandler {
	return &AnalyzeHandler{
		foodService: foodService,
		repository:  repository,
		storageRepo: storageRepo,
	}
}

// Handle はPOST /api/analyzeリクエストを処理
func (h *AnalyzeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Content-Typeで処理を分岐
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		h.handleTextInput(w, r)
		return
	}

	// 既存の画像アップロード処理
	h.handleImageUpload(w, r)
}

// handleTextInput はテキスト入力を処理
func (h *AnalyzeHandler) handleTextInput(w http.ResponseWriter, r *http.Request) {
	log.Printf("Processing text input request")

	var req struct {
		InputText string `json:"input_text"`
		MealType  string `json:"meal_type"`
		MealDate  string `json:"meal_date"`
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4096) // 4KB: テキスト入力(len()で最大1000バイト) + JSONオーバーヘッド

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			log.Printf("Request body too large: limit=%d", maxBytesErr.Limit)
			http.Error(w, "リクエストボディが大きすぎます", http.StatusRequestEntityTooLarge)
			return
		}
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "リクエストの解析に失敗しました", http.StatusBadRequest)
		return
	}

	// バリデーション
	if req.InputText == "" {
		http.Error(w, "テキストを入力してください", http.StatusBadRequest)
		return
	}
	if len(req.InputText) > 1000 {
		http.Error(w, "テキストは1000文字以内で入力してください", http.StatusBadRequest)
		return
	}

	// meal_type のバリデーション
	if !isValidMealType(req.MealType) {
		log.Printf("Invalid meal_type: %s", req.MealType)
		http.Error(w, "無効な食事タイプです（breakfast, lunch, dinner, snackのいずれか）", http.StatusBadRequest)
		return
	}

	// meal_date が空の場合は今日の日付、指定時はフォーマット検証
	if req.MealDate == "" {
		req.MealDate = time.Now().Format("2006-01-02")
	} else if _, err := time.Parse("2006-01-02", req.MealDate); err != nil {
		http.Error(w, "meal_dateはYYYY-MM-DD形式で指定してください", http.StatusBadRequest)
		return
	}

	// contextからユーザーIDを取得
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	log.Printf("Text input: %s, Meal type: %s, Meal date: %s, UserID: %v", util.TruncateForLog(req.InputText, 50), req.MealType, req.MealDate, userID)

	// リポジトリにテキスト分析リクエストを登録
	analysisID, err := h.repository.CreateRequestWithText(r.Context(), req.InputText, req.MealType, req.MealDate, userIDPtr)
	if err != nil {
		log.Printf("Error creating text analysis request: %v", err)
		http.Error(w, "分析リクエストの作成に失敗しました", http.StatusInternalServerError)
		return
	}

	log.Printf("Text analysis request created with ID: %s", analysisID)

	// 202 Accepted レスポンスを返却
	if err := writeAnalysisResponse(w, analysisID, ""); err != nil {
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
	}
}

// handleImageUpload は画像アップロードを処理
func (h *AnalyzeHandler) handleImageUpload(w http.ResponseWriter, r *http.Request) {
	// 1. multipart/form-data パース
	err := r.ParseMultipartForm(10 << 20) // 10MB制限
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		http.Error(w, "ファイルのパースに失敗しました", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		log.Printf("Error getting form file: %v", err)
		http.Error(w, "画像ファイルが見つかりません", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Received file: %s (size: %d bytes)", header.Filename, header.Size)

	// 2. ファイルバリデーション（JPEG, PNG, HEIC、最大10MB）
	if err := validateImageFile(file, header); err != nil {
		log.Printf("File validation error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 3. meal_type, meal_date を取得・バリデーション（アップロード前に検証）
	mealType := r.FormValue("meal_type")
	mealDate := r.FormValue("meal_date")
	tz := r.FormValue("tz")

	// meal_type のバリデーション
	if !isValidMealType(mealType) {
		log.Printf("Invalid meal_type: %s", mealType)
		http.Error(w, "無効な食事タイプです（breakfast, lunch, dinner, snackのいずれか）", http.StatusBadRequest)
		return
	}

	// meal_date が空の場合は指定タイムゾーンでの今日の日付、指定時はフォーマット検証
	if mealDate == "" {
		loc := time.UTC
		if tz != "" {
			if parsedLoc, err := time.LoadLocation(tz); err == nil {
				loc = parsedLoc
			}
		}
		mealDate = time.Now().In(loc).Format("2006-01-02")
	} else if _, err := time.Parse("2006-01-02", mealDate); err != nil {
		http.Error(w, "meal_dateはYYYY-MM-DD形式で指定してください", http.StatusBadRequest)
		return
	}

	// 4. Cloud Storageにアップロード
	contentType := getContentType(header)
	objectName, err := h.storageRepo.Upload(r.Context(), file, header.Filename, contentType)
	if err != nil {
		log.Printf("Error uploading file to Cloud Storage: %v", err)
		http.Error(w, "ファイルの保存に失敗しました", http.StatusInternalServerError)
		return
	}

	log.Printf("File uploaded to Cloud Storage: %s", objectName)

	// contextからユーザーIDを取得
	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	log.Printf("Meal type: %s, Meal date: %s, UserID: %v", mealType, mealDate, userID)

	// 5. リポジトリに分析リクエストを登録（imagePath = Cloud Storageのオブジェクト名）
	analysisID, err := h.repository.CreateRequest(r.Context(), objectName, mealType, mealDate, userIDPtr)
	if err != nil {
		log.Printf("Error creating analysis request: %v", err)
		// Cloud Storageからファイルを削除
		if delErr := h.storageRepo.Delete(r.Context(), objectName); delErr != nil {
			log.Printf("Error deleting file from Cloud Storage: %v", delErr)
		}
		http.Error(w, "分析リクエストの作成に失敗しました", http.StatusInternalServerError)
		return
	}

	log.Printf("Analysis request created with ID: %s", analysisID)

	// 6. 202 Accepted レスポンスを返却
	if err := writeAnalysisResponse(w, analysisID, objectName); err != nil {
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
	}
}

// HandleUploadImage は画像のみをアップロードし、パスを返す（分析なし）
func (h *AnalyzeHandler) HandleUploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. multipart/form-data パース
	err := r.ParseMultipartForm(10 << 20) // 10MB制限
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		http.Error(w, "ファイルのパースに失敗しました", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		log.Printf("Error getting form file: %v", err)
		http.Error(w, "画像ファイルが見つかりません", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Received file for upload: %s (size: %d bytes)", header.Filename, header.Size)

	// 2. ファイルバリデーション
	if err := validateImageFile(file, header); err != nil {
		log.Printf("File validation error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 3. Cloud Storageにアップロード
	contentType := getContentType(header)
	objectName, err := h.storageRepo.Upload(r.Context(), file, header.Filename, contentType)
	if err != nil {
		log.Printf("Error uploading file to Cloud Storage: %v", err)
		http.Error(w, "ファイルの保存に失敗しました", http.StatusInternalServerError)
		return
	}

	log.Printf("File uploaded to Cloud Storage: %s", objectName)

	// 4. レスポンスを返す
	response := map[string]string{
		"image_path": objectName,
	}

	// バッファに先にエンコードしてエラーを検出（ヘッダー書き込み前）
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		// レスポンス生成失敗時はCloud Storageからファイルを削除
		if delErr := h.storageRepo.Delete(r.Context(), objectName); delErr != nil {
			log.Printf("Error deleting file from Cloud Storage: %v", delErr)
		}
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		log.Printf("Error writing response: %v", err)
		// 既にヘッダーを書き込んでいるのでエラーレスポンスは返せないが、
		// ファイルは削除しない（クライアントがレスポンスを受信した可能性がある）
	}
}

// validateImageFile はファイルのバリデーションを行う
func validateImageFile(file multipart.File, header *multipart.FileHeader) error {
	// ファイルサイズチェック（最大10MB）
	if header.Size > 10<<20 {
		return fmt.Errorf("ファイルサイズが大きすぎます（最大10MB）")
	}

	// 拡張子チェック
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".heic": true,
	}

	if !validExtensions[ext] {
		return fmt.Errorf("サポートされていないファイル形式です（JPEG, PNG, HEICのみ）")
	}

	// MIMEタイプチェック
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("ファイルの読み込みに失敗しました")
	}

	// ファイルポインタを先頭に戻す
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("ファイルポインタのリセットに失敗しました: %w", err)
		}
	}

	contentType := http.DetectContentType(buffer)
	validContentTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
	}

	// HEICはDetectContentTypeで検出できないため、マジックバイトで検証
	if !validContentTypes[contentType] {
		if ext == ".heic" && isHEICMagicBytes(buffer) {
			return nil
		}
		return fmt.Errorf("画像ファイルではありません")
	}

	return nil
}

// validHEICBrands はHEIC/HEIF形式のブランド識別子の集合（パッケージ初期化時に一度だけ作成）
var validHEICBrands = map[string]bool{
	"heic": true,
	"heix": true,
	"hevc": true,
	"hevx": true,
	"heim": true,
	"heis": true,
	"hevm": true,
	"hevs": true,
	"mif1": true,
}

// isHEICMagicBytes はバッファがHEIC/HEIF形式のマジックバイトを持つか検証する。
// ISOBMFF形式: オフセット4-7が"ftyp"、オフセット8-11がブランド識別子。
func isHEICMagicBytes(buf []byte) bool {
	if len(buf) < 12 {
		return false
	}

	// オフセット4-7: "ftyp" マーカー
	if string(buf[4:8]) != "ftyp" {
		return false
	}

	// オフセット8-11: HEIC/HEIF ブランド識別子
	return validHEICBrands[string(buf[8:12])]
}

// getContentType はファイルヘッダーからContent-Typeを取得
func getContentType(header *multipart.FileHeader) string {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".heic":
		return "image/heic"
	default:
		return "application/octet-stream"
	}
}

// validMealTypes は有効な食事タイプの集合（パッケージ初期化時に一度だけ作成）
var validMealTypes = map[string]bool{
	"breakfast": true,
	"lunch":     true,
	"dinner":    true,
	"snack":     true,
}

// isValidMealType は meal_type が有効な値かチェックします
func isValidMealType(mealType string) bool {
	return validMealTypes[mealType]
}

// writeAnalysisResponse は分析IDを含む202 Acceptedレスポンスを書き込みます
func writeAnalysisResponse(w http.ResponseWriter, analysisID uuid.UUID, imagePath string) error {
	response := map[string]string{
		"id": analysisID.String(),
	}
	if imagePath != "" {
		response["image_path"] = imagePath
	}

	// バッファに先にエンコードしてエラーを検出（ヘッダー書き込み前）
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	if _, err := w.Write(data); err != nil {
		log.Printf("Error writing response: %v", err)
		return err
	}

	log.Printf("Response sent successfully: id=%s, image_path=%s", analysisID, imagePath)
	return nil
}
