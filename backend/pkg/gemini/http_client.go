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
	defaultBaseURL = "https://generativelanguage.googleapis.com"
	modelName      = "gemini-3.0-flash"
)

// HTTPClient はGemini HTTP APIクライアント
type HTTPClient struct {
	apiKey     string
	timeout    time.Duration
	baseURL    string
	httpClient *http.Client
}

// NewHTTPClient は新しいGemini HTTP APIクライアントを作成
func NewHTTPClient(apiKey string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		apiKey:  apiKey,
		timeout: timeout,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
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

	// レスポンスボディを読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンス読み取りエラー: %w", err)
	}

	// ステータスコードをチェック
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("APIエラー (status %d): %s", resp.StatusCode, string(body))
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

	// レスポンステキストを抽出
	var responseText string
	if len(apiResp.Candidates) > 0 && len(apiResp.Candidates[0].Content.Parts) > 0 {
		responseText = apiResp.Candidates[0].Content.Parts[0].Text
	}

	// 既存のResponse型に変換（互換性のため）
	return &Response{
		Response: responseText,
	}, nil
}
