package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
)

// emailRegex はメールアドレスバリデーション用の正規表現（パッケージ初期化時に一度だけコンパイル）
var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// FoodService は食品分析サービスのインターフェース
type FoodService interface {
	AnalyzeFoodImage(ctx context.Context, imagePath string) (*service.AnalysisResult, error)
	AnalyzeFoodText(ctx context.Context, inputText string) (*service.AnalysisResult, error)
}

// AnalyzeHandler は画像分析エンドポイントのハンドラー
type AnalyzeHandler struct {
	foodService FoodService
	repository  repository.AnalysisRepository
}

// NewAnalyzeHandler は新しいAnalyzeHandlerを作成
func NewAnalyzeHandler(foodService FoodService, repository repository.AnalysisRepository) *AnalyzeHandler {
	return &AnalyzeHandler{
		foodService: foodService,
		repository:  repository,
	}
}

// Handle はPOST /api/analyzeリクエストを処理
func (h *AnalyzeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

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

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

	// meal_date が空の場合は今日の日付
	if req.MealDate == "" {
		req.MealDate = time.Now().Format("2006-01-02")
	}

	// contextからユーザーIDを取得
	userID := middleware.GetUserIDFromContext(r.Context())
	var userIDPtr *uuid.UUID
	if userID != uuid.Nil {
		userIDPtr = &userID
	}

	log.Printf("Text input: %s, Meal type: %s, Meal date: %s, UserID: %v", req.InputText, req.MealType, req.MealDate, userID)

	// リポジトリにテキスト分析リクエストを登録
	analysisID, err := h.repository.CreateRequestWithText(r.Context(), req.InputText, req.MealType, req.MealDate, userIDPtr)
	if err != nil {
		log.Printf("Error creating text analysis request: %v", err)
		http.Error(w, "分析リクエストの作成に失敗しました", http.StatusInternalServerError)
		return
	}

	log.Printf("Text analysis request created with ID: %s", analysisID)

	// 202 Accepted レスポンスを返却
	if err := writeAnalysisResponse(w, analysisID); err != nil {
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

	// 3. 永続保存: uploads/{uuid}.{ext}（ワーカーが処理後に削除）
	permanentPath, err := savePermanentFile(file, header)
	if err != nil {
		log.Printf("Error saving file: %v", err)
		http.Error(w, "ファイルの保存に失敗しました", http.StatusInternalServerError)
		return
	}

	log.Printf("File saved permanently to: %s", permanentPath)

	// 4. meal_type, meal_date を取得・バリデーション
	mealType := r.FormValue("meal_type")
	mealDate := r.FormValue("meal_date")

	// meal_type のバリデーション
	if !isValidMealType(mealType) {
		log.Printf("Invalid meal_type: %s", mealType)
		os.Remove(permanentPath)
		http.Error(w, "無効な食事タイプです（breakfast, lunch, dinner, snackのいずれか）", http.StatusBadRequest)
		return
	}

	// meal_date が空の場合は今日の日付
	if mealDate == "" {
		mealDate = time.Now().Format("2006-01-02")
	}

	// contextからユーザーIDを取得
	userID := middleware.GetUserIDFromContext(r.Context())
	var userIDPtr *uuid.UUID
	if userID != uuid.Nil {
		userIDPtr = &userID
	}

	log.Printf("Meal type: %s, Meal date: %s, UserID: %v", mealType, mealDate, userID)

	// 5. リポジトリに分析リクエストを登録
	analysisID, err := h.repository.CreateRequest(r.Context(), permanentPath, mealType, mealDate, userIDPtr)
	if err != nil {
		log.Printf("Error creating analysis request: %v", err)
		// ファイル削除
		os.Remove(permanentPath)
		http.Error(w, "分析リクエストの作成に失敗しました", http.StatusInternalServerError)
		return
	}

	log.Printf("Analysis request created with ID: %s", analysisID)

	// 5. 202 Accepted レスポンスを返却
	if err := writeAnalysisResponse(w, analysisID); err != nil {
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
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
		_, _ = seeker.Seek(0, io.SeekStart)
	}

	contentType := http.DetectContentType(buffer)
	validContentTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
	}

	// HEICはDetectContentTypeで検出できないため、拡張子でのみチェック
	if !validContentTypes[contentType] && ext != ".heic" {
		return fmt.Errorf("画像ファイルではありません")
	}

	return nil
}

// savePermanentFile はファイルを永続的に保存（ワーカーが処理後に削除）
func savePermanentFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	// UUIDを生成してファイル名を作成（ディレクトリトラバーサル対策）
	fileID := uuid.New().String()
	ext := filepath.Ext(header.Filename)
	filename := fileID + ext

	// 保存先ディレクトリを作成（バックエンドディレクトリ内のuploads/）
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return "", fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}

	// ファイルパスを作成（保存先を /tmp/uchikomi/uploads/ に制限）
	destPath := filepath.Join(uploadDir, filename)

	// ディレクトリトラバーサル対策: destPathがuploadDir配下であることを確認
	if !strings.HasPrefix(filepath.Clean(destPath), filepath.Clean(uploadDir)) {
		return "", fmt.Errorf("不正なファイルパスです")
	}

	// ファイルを保存
	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("ファイルの作成に失敗しました: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, file); err != nil {
		return "", fmt.Errorf("ファイルのコピーに失敗しました: %w", err)
	}

	return destPath, nil
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
func writeAnalysisResponse(w http.ResponseWriter, analysisID uuid.UUID) error {
	response := map[string]string{
		"analysis_id": analysisID.String(),
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

	log.Printf("Response sent successfully: analysis_id=%s", analysisID)
	return nil
}
