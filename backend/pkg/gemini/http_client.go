package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL     = "https://generativelanguage.googleapis.com"
	modelName          = "gemini-3-flash-preview"
	maxResponseSize    = 10 << 20 // 10MB: JSONレスポンスの上限（通常100KB未満）
)

// GeminiHTTPClient はGemini APIのHTTPクライアントを表すインターフェース
type GeminiHTTPClient interface {
	Execute(ctx context.Context, prompt string) (*Response, error)
	ExecuteWithImage(ctx context.Context, prompt string, imageData []byte, mimeType string) (*Response, error)
}

// HTTPClient はGemini HTTP APIクライアント
type HTTPClient struct {
	apiKey     string
	timeout    time.Duration
	baseURL    string
	httpClient *http.Client
}

// ErrEmptyAPIKey はAPIキーが空の場合に返されるエラー
var ErrEmptyAPIKey = errors.New("APIキーが指定されていません")

// NewHTTPClient は新しいGemini HTTP APIクライアントを作成
// APIキーが空の場合はエラーを返す
func NewHTTPClient(apiKey string, timeout time.Duration) (*HTTPClient, error) {
	if apiKey == "" {
		return nil, ErrEmptyAPIKey
	}
	return &HTTPClient{
		apiKey:  apiKey,
		timeout: timeout,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{}, // タイムアウトはcontextで制御
	}, nil
}

// GenerateContentRequest はGemini APIのリクエスト構造体
type GenerateContentRequest struct {
	Contents         []Content        `json:"contents"`
	GenerationConfig GenerationConfig `json:"generationConfig,omitempty"`
}

// Content はリクエスト内のコンテンツ
type Content struct {
	Parts []Part `json:"parts"`
}

// Part はコンテンツ内のパーツ（テキストまたは画像）
type Part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inlineData,omitempty"`
}

// InlineData は画像データを含む構造体
type InlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // base64エンコードされた画像データ
}

// GenerationConfig は生成設定
type GenerationConfig struct {
	ResponseMimeType string `json:"responseMimeType,omitempty"`
}

// GenerateContentResponse はGemini APIのレスポンス構造体
type GenerateContentResponse struct {
	Candidates []Candidate `json:"candidates"`
}

// Candidate はレスポンス内の候補
type Candidate struct {
	Content ContentResponse `json:"content"`
}

// ContentResponse はレスポンス内のコンテンツ
type ContentResponse struct {
	Parts []PartResponse `json:"parts"`
}

// PartResponse はレスポンス内のパーツ
type PartResponse struct {
	Text string `json:"text"`
}

// Execute はテキストプロンプトを実行してレスポンスを返す
func (c *HTTPClient) Execute(ctx context.Context, prompt string) (*Response, error) {
	req := GenerateContentRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: GenerationConfig{
			ResponseMimeType: "application/json",
		},
	}

	return c.doRequest(ctx, req)
}

// ExecuteWithImage は画像付きプロンプトを実行してレスポンスを返す
func (c *HTTPClient) ExecuteWithImage(ctx context.Context, prompt string, imageData []byte, mimeType string) (*Response, error) {
	if len(imageData) == 0 {
		return nil, fmt.Errorf("画像データが空です")
	}

	// 画像をbase64エンコード
	encodedImage := base64.StdEncoding.EncodeToString(imageData)

	req := GenerateContentRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{
						InlineData: &InlineData{
							MimeType: mimeType,
							Data:     encodedImage,
						},
					},
					{Text: prompt},
				},
			},
		},
		GenerationConfig: GenerationConfig{
			ResponseMimeType: "application/json",
		},
	}

	return c.doRequest(ctx, req)
}

// doRequest は実際のHTTPリクエストを実行する
func (c *HTTPClient) doRequest(ctx context.Context, reqBody GenerateContentRequest) (*Response, error) {
	// タイムアウト付きコンテキストを作成
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// リクエストボディをJSONにエンコード
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("リクエストのJSON化エラー: %w", err)
	}

	// URLを構築
	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent", c.baseURL, modelName)

	// HTTPリクエストを作成
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("リクエスト作成エラー: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey) // APIキーをヘッダーで送信（セキュリティ向上）

	// リクエストを実行
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// コンテキストのエラーチェック
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("タイムアウト: Gemini APIの実行が%v以内に完了しませんでした", c.timeout)
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil, fmt.Errorf("キャンセル: Gemini APIの実行がキャンセルされました")
		}
		return nil, fmt.Errorf("HTTPリクエストエラー: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスボディを読み取り（サイズ制限付き）
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("レスポンス読み取りエラー: %w", err)
	}
	if int64(len(body)) > maxResponseSize {
		return nil, fmt.Errorf("レスポンスサイズが上限(%dMB)を超えています", maxResponseSize>>20)
	}

	// ステータスコードをチェック
	if resp.StatusCode != http.StatusOK {
		return nil, c.handleAPIError(resp.StatusCode, body)
	}

	// レスポンスをパース
	var apiResp GenerateContentResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("レスポンスのパースエラー: %w\n生データ: %s", err, string(body))
	}

	// Candidatesが空かチェック
	if len(apiResp.Candidates) == 0 {
		return nil, fmt.Errorf("レスポンスが空です")
	}

	// Partsが空かチェック
	if len(apiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("レスポンスのコンテンツが空です")
	}

	// レスポンステキストを抽出
	responseText := apiResp.Candidates[0].Content.Parts[0].Text
	if responseText == "" {
		return nil, fmt.Errorf("レスポンステキストが空です")
	}

	// 既存のResponse型に変換（互換性のため）
	return &Response{
		Response: responseText,
	}, nil
}

// handleAPIError はAPIエラーレスポンスを安全に処理する
func (c *HTTPClient) handleAPIError(statusCode int, body []byte) error {
	// エラーレスポンスをパース試行（APIキー等の機密情報を含む可能性のある生データは出力しない）
	var apiErr struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &apiErr)

	switch statusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("認証エラー: APIキーが無効です")
	case http.StatusForbidden:
		return fmt.Errorf("アクセス拒否: APIへのアクセス権限がありません")
	case http.StatusTooManyRequests:
		return fmt.Errorf("レート制限: しばらく待ってから再試行してください")
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return fmt.Errorf("Geminiサービスエラー: サービスが一時的に利用できません")
	default:
		if apiErr.Error.Message != "" {
			return fmt.Errorf("APIエラー (status %d): %s", statusCode, apiErr.Error.Message)
		}
		return fmt.Errorf("APIエラー (status %d)", statusCode)
	}
}
