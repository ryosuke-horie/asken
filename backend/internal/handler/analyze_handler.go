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
	"github.com/ryosuke-horie/asken/backend/internal/service"
)

// FoodService は食品分析サービスのインターフェース
type FoodService interface {
	AnalyzeFoodImage(ctx context.Context, imagePath string) (*service.AnalysisResult, error)
}

// AnalyzeHandler は画像分析エンドポイントのハンドラー
type AnalyzeHandler struct {
	foodService FoodService
}

// NewAnalyzeHandler は新しいAnalyzeHandlerを作成
func NewAnalyzeHandler(foodService FoodService) *AnalyzeHandler {
	return &AnalyzeHandler{
		foodService: foodService,
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

	// 3. 一時保存: /tmp/asken/uploads/{uuid}.{ext}
	tempPath, err := saveTemporaryFile(file, header)
	if err != nil {
		log.Printf("Error saving file: %v", err)
		http.Error(w, "ファイルの保存に失敗しました", http.StatusInternalServerError)
		return
	}

	log.Printf("File saved to: %s", tempPath)

	// 6. defer でアップロードファイル削除
	defer func() {
		if err := os.Remove(tempPath); err != nil {
			log.Printf("Error removing temp file: %v", err)
		} else {
			log.Printf("Temp file removed: %s", tempPath)
		}
	}()

	// 4. FoodService.AnalyzeFoodImage() 呼び出し
	log.Printf("Starting food analysis for: %s", tempPath)
	result, err := h.foodService.AnalyzeFoodImage(r.Context(), tempPath)
	if err != nil {
		log.Printf("Analysis error: %v", err)
		http.Error(w, fmt.Sprintf("分析に失敗しました: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Analysis completed successfully. Found %d foods", len(result.Foods))

	// 5. JSON レスポンス返却
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}

	log.Printf("Response sent successfully")
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

// saveTemporaryFile はファイルを一時ディレクトリに保存
func saveTemporaryFile(file multipart.File, header *multipart.FileHeader) (string, error) {
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
