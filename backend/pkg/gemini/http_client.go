package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL  = "https://generativelanguage.googleapis.com"
	modelName       = "gemini-3-flash-preview"
	maxResponseSize = 10 << 20 // 10MB: JSONレスポンスの上限（通常100KB未満）
	maxImageSize    = 20 << 20 // 20MB: 画像ファイルサイズの上限
)

// GeminiHTTPClient はGemini APIのHTTPクライアントを表すインターフェース
type GeminiHTTPClient interface {
	Execute(ctx context.Context, prompt string, schema *Schema) (*Response, error)
	ExecuteWithImage(ctx context.Context, prompt string, imageData []byte, mimeType string, schema *Schema) (*Response, error)
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

// ErrEmptyPrompt はプロンプトが空の場合に返されるエラー
var ErrEmptyPrompt = errors.New("プロンプトが空です")

// ErrInvalidTimeout はタイムアウト値が正でない場合に返されるエラー
var ErrInvalidTimeout = errors.New("タイムアウトは正の値を指定してください")

// ErrImageTooLarge は画像サイズが上限を超えた場合に返されるエラー
var ErrImageTooLarge = errors.New("画像サイズが上限(20MB)を超えています")

// NewHTTPClient は新しいGemini HTTP APIクライアントを作成
// APIキーが空の場合、またはタイムアウトが正でない場合はエラーを返す
func NewHTTPClient(apiKey string, timeout time.Duration) (*HTTPClient, error) {
	if apiKey == "" {
		return nil, ErrEmptyAPIKey
	}
	if timeout <= 0 {
		return nil, ErrInvalidTimeout
	}
	return &HTTPClient{
		apiKey:     apiKey,
		timeout:    timeout,
		baseURL:    defaultBaseURL,
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
	ResponseMimeType string  `json:"responseMimeType,omitempty"`
	ResponseSchema   *Schema `json:"responseSchema,omitempty"`
}

// Schema はGemini APIのレスポンススキーマ定義
// responseSchemaを指定すると、モデルの出力が指定したJSON Schemaに準拠することが保証される
type Schema struct {
	Type       SchemaType         `json:"type"`
	Items      *Schema            `json:"items,omitempty"`
	Properties map[string]*Schema `json:"properties,omitempty"`
	Required   []string           `json:"required,omitempty"`
	Enum       []string           `json:"enum,omitempty"`
}

// SchemaType はGemini APIのスキーマタイプ
type SchemaType string

const (
	SchemaTypeString  SchemaType = "STRING"
	SchemaTypeNumber  SchemaType = "NUMBER"
	SchemaTypeInteger SchemaType = "INTEGER"
	SchemaTypeBoolean SchemaType = "BOOLEAN"
	SchemaTypeArray   SchemaType = "ARRAY"
	SchemaTypeObject  SchemaType = "OBJECT"
)

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
// schemaがnilでない場合、responseSchemaを設定してモデル出力を制約する
func (c *HTTPClient) Execute(ctx context.Context, prompt string, schema *Schema) (*Response, error) {
	if strings.TrimSpace(prompt) == "" {
		return nil, ErrEmptyPrompt
	}

	log.Printf("Gemini API呼び出し開始: テキストプロンプト (長さ: %d文字)", len(prompt))

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
			ResponseSchema:   schema,
		},
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		log.Printf("Gemini API呼び出しエラー: テキストプロンプト: %v", err)
		return nil, err
	}

	log.Printf("Gemini API呼び出し完了: テキストプロンプト (レスポンス長: %d文字)", len(resp.Response))
	return resp, nil
}

// ExecuteWithImage は画像付きプロンプトを実行してレスポンスを返す
// schemaがnilでない場合、responseSchemaを設定してモデル出力を制約する
func (c *HTTPClient) ExecuteWithImage(ctx context.Context, prompt string, imageData []byte, mimeType string, schema *Schema) (*Response, error) {
	if strings.TrimSpace(prompt) == "" {
		return nil, ErrEmptyPrompt
	}
	if len(imageData) == 0 {
		return nil, fmt.Errorf("画像データが空です")
	}
	if len(imageData) > maxImageSize {
		return nil, ErrImageTooLarge
	}

	log.Printf("Gemini API呼び出し開始: 画像付きプロンプト (画像サイズ: %d bytes, MIME: %s)", len(imageData), mimeType)

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
			ResponseSchema:   schema,
		},
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		log.Printf("Gemini API呼び出しエラー: 画像付きプロンプト: %v", err)
		return nil, err
	}

	log.Printf("Gemini API呼び出し完了: 画像付きプロンプト (レスポンス長: %d文字)", len(resp.Response))
	return resp, nil
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
		log.Printf("Gemini APIレスポンスのJSONパースエラー: %v", err)
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
	if err := json.Unmarshal(body, &apiErr); err != nil {
		log.Printf("Gemini APIエラーレスポンスのJSONパース失敗 (status %d): %v", statusCode, err)
	}

	switch statusCode {
	case http.StatusUnauthorized:
		log.Printf("Gemini API認証エラー: APIキーが無効です")
		return fmt.Errorf("認証エラー: APIキーが無効です")
	case http.StatusForbidden:
		log.Printf("Gemini APIアクセス拒否: 権限がありません")
		return fmt.Errorf("アクセス拒否: APIへのアクセス権限がありません")
	case http.StatusTooManyRequests:
		log.Printf("Gemini APIレート制限に到達")
		return fmt.Errorf("レート制限: しばらく待ってから再試行してください")
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		log.Printf("Gemini APIサービスエラー (status %d)", statusCode)
		return fmt.Errorf("Geminiサービスエラー: サービスが一時的に利用できません")
	default:
		if apiErr.Error.Message != "" {
			log.Printf("Gemini APIエラー (status %d): %s", statusCode, apiErr.Error.Message)
			return fmt.Errorf("APIエラー (status %d): %s", statusCode, apiErr.Error.Message)
		}
		log.Printf("Gemini APIエラー (status %d): メッセージなし", statusCode)
		return fmt.Errorf("APIエラー (status %d)", statusCode)
	}
}
