package handler

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ImageHandler は画像配信エンドポイントのハンドラー
type ImageHandler struct {
	uploadsDir string
}

// NewImageHandler は新しいImageHandlerを作成
func NewImageHandler(uploadsDir string) *ImageHandler {
	return &ImageHandler{
		uploadsDir: uploadsDir,
	}
}

// Handle はGET /api/images/:filenameリクエストを処理
func (h *ImageHandler) Handle(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received image request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

	// GETメソッドのみ許可
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// URLからファイル名を抽出
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

	// ディレクトリトラバーサル対策
	// 1. filepath.Clean でパスを正規化
	cleanPath := filepath.Clean(filepath.Join(h.uploadsDir, filename))

	// 2. uploadsDir配下であることを確認
	absUploadsDir, err := filepath.Abs(h.uploadsDir)
	if err != nil {
		log.Printf("Error getting absolute path for uploads directory: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	absCleanPath, err := filepath.Abs(cleanPath)
	if err != nil {
		log.Printf("Error getting absolute path for file: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// uploadsDir配下であることをチェック
	if !strings.HasPrefix(absCleanPath, absUploadsDir) {
		log.Printf("Path traversal attempt detected: %s", filename)
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	log.Printf("Serving image file: %s", absCleanPath)

	// ファイルの存在確認
	if _, err := os.Stat(absCleanPath); os.IsNotExist(err) {
		log.Printf("Image file not found: %s", absCleanPath)
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	// 画像ファイルを配信
	w.Header().Set("Content-Type", "image/jpeg")
	http.ServeFile(w, r, absCleanPath)

	log.Printf("Image served successfully: %s", absCleanPath)
}
