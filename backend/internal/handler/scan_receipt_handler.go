package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ryosuke-horie/uchikomi/backend/internal/middleware"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
)

// ReceiptParserClient はレシート解析クライアントのインターフェース
type ReceiptParserClient interface {
	ParseReceiptImage(ctx context.Context, imageData []byte, mimeType string) ([]gemini.ReceiptIngredient, error)
}

// ScanReceiptHandler はPOST /api/ingredients/scan-receipt エンドポイントのハンドラー
type ScanReceiptHandler struct {
	receiptParser ReceiptParserClient
}

// NewScanReceiptHandler は新しいScanReceiptHandlerを作成
func NewScanReceiptHandler(receiptParser ReceiptParserClient) *ScanReceiptHandler {
	if receiptParser == nil {
		panic("scan receipt handler: receiptParser must not be nil")
	}
	return &ScanReceiptHandler{receiptParser: receiptParser}
}

// ScanReceiptIngredientItem はスキャン結果の食材アイテム
type ScanReceiptIngredientItem struct {
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
	Source   string  `json:"source"`
}

// ScanReceiptResponse はレシートスキャンのレスポンス
type ScanReceiptResponse struct {
	Ingredients []ScanReceiptIngredientItem `json:"ingredients"`
}

// Handle はPOST /api/ingredients/scan-receipt リクエストを処理
func (h *ScanReceiptHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetFirebaseUIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	imageData, mimeType, ok := parseAndValidateImageUpload(w, r, userID)
	if !ok {
		return
	}

	// Gemini APIでレシート解析
	ingredients, err := h.receiptParser.ParseReceiptImage(r.Context(), imageData, mimeType)
	if err != nil {
		handleReceiptParseError(w, userID, err)
		return
	}

	// レスポンス生成
	items := make([]ScanReceiptIngredientItem, len(ingredients))
	for i, ing := range ingredients {
		items[i] = ScanReceiptIngredientItem{
			Name:     ing.Name,
			Category: ing.Category,
			Quantity: ing.Quantity,
			Unit:     ing.Unit,
			Source:   "receipt",
		}
	}

	response := ScanReceiptResponse{Ingredients: items}
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("ScanReceiptHandler: レスポンス生成エラー: userID=%s, error=%v", userID, err)
		http.Error(w, "レスポンスの生成に失敗しました", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		// この時点でヘッダーは送信済みのため、エラーレスポンスへの変更は不可能
		// クライアントには不完全なJSONが届く可能性があり、クライアント側で再試行が必要
		log.Printf("ScanReceiptHandler: レスポンス書き込みエラー（クライアントに不完全なデータが送信された可能性あり）: userID=%s, error=%v", userID, err)
	}
}

// parseAndValidateImageUpload はmultipartリクエストから画像を取り出してバリデーションする
// 返値: imageData, mimeType, ok（falseのとき既にエラーレスポンス送信済み）
func parseAndValidateImageUpload(w http.ResponseWriter, r *http.Request, userID string) ([]byte, string, bool) {
	// リクエスト全体を10MBに制限（超過はParseMultipartFormで検出される）
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			log.Printf("ScanReceiptHandler: リクエストサイズ超過: userID=%s, limit=%d", userID, maxBytesErr.Limit)
			http.Error(w, "ファイルサイズが大きすぎます（最大10MB）", http.StatusBadRequest)
			return nil, "", false
		}
		log.Printf("ScanReceiptHandler: multipartパースエラー: userID=%s, error=%v", userID, err)
		http.Error(w, "ファイルのパースに失敗しました", http.StatusBadRequest)
		return nil, "", false
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		log.Printf("ScanReceiptHandler: 画像取得エラー: userID=%s, error=%v", userID, err)
		http.Error(w, "画像ファイルが見つかりません", http.StatusBadRequest)
		return nil, "", false
	}
	defer file.Close()

	log.Printf("ScanReceiptHandler: ファイル受信: name=%s, size=%d bytes, userID=%s", header.Filename, header.Size, userID)

	// ファイル名が空の場合はエラー
	if header.Filename == "" {
		log.Printf("ScanReceiptHandler: ファイル名が空: userID=%s", userID)
		http.Error(w, "ファイル名が指定されていません", http.StatusBadRequest)
		return nil, "", false
	}

	// 拡張子チェック（JPEG, PNGのみ。HEICはGemini API非対応）
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExtensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	if !validExtensions[ext] {
		log.Printf("ScanReceiptHandler: 非対応拡張子: userID=%s, filename=%s, ext=%s", userID, header.Filename, ext)
		http.Error(w, "サポートされていないファイル形式です（JPEG, PNGのみ）", http.StatusBadRequest)
		return nil, "", false
	}

	// 画像データを全て読み込む（10MB+1バイトで上限を検知）
	imageData, err := io.ReadAll(io.LimitReader(file, 10<<20+1))
	if err != nil {
		log.Printf("ScanReceiptHandler: 画像読み込みエラー: userID=%s, error=%v", userID, err)
		http.Error(w, "ファイルの読み込みに失敗しました", http.StatusInternalServerError)
		return nil, "", false
	}
	if int64(len(imageData)) > 10<<20 {
		log.Printf("ScanReceiptHandler: ファイルサイズ超過: userID=%s, filename=%s, size=%d bytes", userID, header.Filename, len(imageData))
		http.Error(w, "ファイルサイズが大きすぎます（最大10MB）", http.StatusBadRequest)
		return nil, "", false
	}

	// MIMEタイプをマジックバイトから判定
	mimeType, err := detectReceiptImageMimeType(imageData)
	if err != nil {
		log.Printf("ScanReceiptHandler: MIMEタイプ判定エラー（拡張子とマジックバイト不一致の可能性）: userID=%s, filename=%s, ext=%s, error=%v", userID, header.Filename, ext, err)
		http.Error(w, "ファイルの形式が正しくありません（JPEG, PNGのみ）", http.StatusBadRequest)
		return nil, "", false
	}

	return imageData, mimeType, true
}

// handleReceiptParseError はレシート解析エラーを種別に応じて処理する
func handleReceiptParseError(w http.ResponseWriter, userID string, err error) {
	if errors.Is(err, context.Canceled) {
		log.Printf("ScanReceiptHandler: クライアント切断によるキャンセル: userID=%s", userID)
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		log.Printf("ScanReceiptHandler: レシート解析タイムアウト: userID=%s, error=%v", userID, err)
		http.Error(w, "レシートの解析がタイムアウトしました。しばらく待ってから再試行してください", http.StatusGatewayTimeout)
		return
	}
	log.Printf("ScanReceiptHandler: レシート解析エラー: userID=%s, error=%v", userID, err)
	http.Error(w, "レシートの解析に失敗しました", http.StatusInternalServerError)
}

// detectReceiptImageMimeType はレシート画像のMIMEタイプをマジックバイトで判定する
// JPEG と PNG のみサポート。拡張子フォールバックは行わない
func detectReceiptImageMimeType(data []byte) (string, error) {
	if len(data) < 3 {
		return "", fmt.Errorf("画像データが不正です")
	}

	// JPEGマジックバイト: FF D8 FF
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg", nil
	}

	// PNGマジックバイト: 89 50 4E 47
	if len(data) >= 4 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png", nil
	}

	return "", fmt.Errorf("画像ファイルではありません（JPEG, PNGのみ対応）")
}
