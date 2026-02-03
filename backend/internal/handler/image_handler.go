package handler

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
)

// 署名付きURLの有効期限
const signedURLExpiration = 15 * time.Minute

// ImageHandler は画像配信エンドポイントのハンドラー
type ImageHandler struct {
	storageRepo repository.StorageRepository
}

// NewImageHandler は新しいImageHandlerを作成
func NewImageHandler(storageRepo repository.StorageRepository) *ImageHandler {
	return &ImageHandler{
		storageRepo: storageRepo,
	}
}

// Handle はGET /api/images/:filenameリクエストを処理
// Cloud Storageの署名付きURLにリダイレクトする
func (h *ImageHandler) Handle(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received image request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

	// GETメソッドのみ許可
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// URLからファイル名を抽出
	// パス形式: /api/images/{filename}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		log.Printf("Invalid URL path: %s", r.URL.Path)
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	filename := pathParts[3]
	if filename == "" {
		log.Printf("Empty filename")
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	// セキュリティチェック: パストラバーサル防止
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		log.Printf("Path traversal attempt detected: %s", filename)
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Cloud Storageのオブジェクト名を構築
	objectName := "uploads/" + filename

	log.Printf("Generating signed URL for: %s", objectName)

	// 署名付きURLを生成
	signedURL, err := h.storageRepo.GetSignedURL(r.Context(), objectName, signedURLExpiration)
	if err != nil {
		log.Printf("Error generating signed URL: %v", err)
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	log.Printf("Redirecting to signed URL for: %s", objectName)

	// 署名付きURLにリダイレクト
	http.Redirect(w, r, signedURL, http.StatusFound)
}
