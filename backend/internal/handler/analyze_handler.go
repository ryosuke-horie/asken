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
	"strings"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/asken/backend/internal/repository"
	"github.com/ryosuke-horie/asken/backend/internal/service"
)

// FoodService は食品分析サービスのインターフェース
type FoodService interface {
	AnalyzeFoodImage(ctx context.Context, imagePath string) (*service.AnalysisResult, error)
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

	// 4. リポジトリに分析リクエストを登録
	analysisID, err := h.repository.CreateRequest(r.Context(), permanentPath)
	if err != nil {
		log.Printf("Error creating analysis request: %v", err)
		// ファイル削除
		os.Remove(permanentPath)
		http.Error(w, "分析リクエストの作成に失敗しました", http.StatusInternalServerError)
		return
	}

	log.Printf("Analysis request created with ID: %s", analysisID)

	// 5. 202 Accepted レスポンスを返却
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	response := map[string]string{
		"analysis_id": analysisID.String(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}

	log.Printf("Response sent successfully: analysis_id=%s", analysisID)
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
		seeker.Seek(0, io.SeekStart)
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
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}

	// ファイルパスを作成（保存先を /tmp/asken/uploads/ に制限）
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
