package gemini

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// skipIfNoGeminiAPIKey skips the test if GEMINI_API_KEY is not set
func skipIfNoGeminiAPIKey(t *testing.T) {
	t.Helper()
	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("GEMINI_API_KEY not set, skipping integration test")
	}
}

func TestNewHTTPClient(t *testing.T) {
	t.Run("有効なAPI Keyで作成できる", func(t *testing.T) {
		client := NewHTTPClient("test-api-key", 30*time.Second)

		assert.NotNil(t, client)
	})

	t.Run("空のAPI Keyでもクライアント自体は作成できる", func(t *testing.T) {
		client := NewHTTPClient("", 30*time.Second)

		assert.NotNil(t, client)
	})
}

func TestHTTPClient_Execute_TextOnly(t *testing.T) {
	t.Run("テキストプロンプトで正常なレスポンスを受け取る", func(t *testing.T) {
		// モックサーバーを作成
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// リクエストの検証
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, r.URL.Path, "/v1beta/models/gemini-2.5-flash:generateContent")
			assert.Equal(t, "test-api-key", r.Header.Get("x-goog-api-key")) // APIキーはヘッダーで送信

			// レスポンスを返す
			response := GenerateContentResponse{
				Candidates: []Candidate{
					{
						Content: ContentResponse{
							Parts: []PartResponse{
								{Text: `[{"name": "ラーメン", "estimated_amount": "1杯"}]`},
							},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewHTTPClient("test-api-key", 30*time.Second)
		client.baseURL = server.URL // テスト用にbaseURLを上書き

		resp, err := client.Execute(context.Background(), "テスト用プロンプト")

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Response, "ラーメン")
	})

	t.Run("APIエラーレスポンスを適切に処理する", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
		}))
		defer server.Close()

		client := NewHTTPClient("invalid-key", 30*time.Second)
		client.baseURL = server.URL

		_, err := client.Execute(context.Background(), "テスト")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "APIエラー")
	})

	t.Run("空のCandidatesでエラーを返す", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := GenerateContentResponse{
				Candidates: []Candidate{},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewHTTPClient("test-api-key", 30*time.Second)
		client.baseURL = server.URL

		_, err := client.Execute(context.Background(), "テスト")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "レスポンスが空です")
	})
}

func TestHTTPClient_Execute_Timeout(t *testing.T) {
	t.Run("タイムアウトでエラーを返す", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// タイムアウトを引き起こすために遅延
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewHTTPClient("test-api-key", 10*time.Millisecond)
		client.baseURL = server.URL

		_, err := client.Execute(context.Background(), "テスト")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "タイムアウト")
	})
}

func TestHTTPClient_Execute_Cancel(t *testing.T) {
	t.Run("コンテキストキャンセルでエラーを返す", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewHTTPClient("test-api-key", 30*time.Second)
		client.baseURL = server.URL

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		_, err := client.Execute(ctx, "テスト")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "キャンセル")
	})
}

func TestHTTPClient_ExecuteWithImage(t *testing.T) {
	t.Run("画像付きリクエストで正常なレスポンスを受け取る", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// リクエストボディを検証
			var req GenerateContentRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)

			// 画像データとテキストが含まれていることを確認
			assert.Len(t, req.Contents, 1)
			assert.Len(t, req.Contents[0].Parts, 2) // 画像 + テキスト

			// レスポンスを返す
			response := GenerateContentResponse{
				Candidates: []Candidate{
					{
						Content: ContentResponse{
							Parts: []PartResponse{
								{Text: `[{"name": "味噌ラーメン", "estimated_amount": "1杯"}]`},
							},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewHTTPClient("test-api-key", 30*time.Second)
		client.baseURL = server.URL

		// テスト用の画像データ（小さなダミーデータ）
		imageData := []byte{0xFF, 0xD8, 0xFF, 0xE0} // JPEG magic bytes

		resp, err := client.ExecuteWithImage(context.Background(), "テスト用プロンプト", imageData, "image/jpeg")

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Contains(t, resp.Response, "味噌ラーメン")
	})

	t.Run("空の画像データでエラーを返す", func(t *testing.T) {
		client := NewHTTPClient("test-api-key", 30*time.Second)

		_, err := client.ExecuteWithImage(context.Background(), "テスト", nil, "image/jpeg")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "画像データが空です")
	})
}

func TestHTTPClient_Execute_Integration(t *testing.T) {
	skipIfNoGeminiAPIKey(t)

	apiKey := os.Getenv("GEMINI_API_KEY")
	client := NewHTTPClient(apiKey, 60*time.Second)

	t.Run("実際のAPIでテキストプロンプトを実行できる", func(t *testing.T) {
		resp, err := client.Execute(context.Background(), "こんにちは。1+1は？")

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Response)
	})
}

func TestRemoveCodeBlock_HTTPClient(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "JSONコードブロック",
			input:    "```json\n[{\"name\":\"test\"}]\n```",
			expected: `[{"name":"test"}]`,
		},
		{
			name:     "コードブロックなし",
			input:    `[{"name":"test"}]`,
			expected: `[{"name":"test"}]`,
		},
		{
			name:     "大文字JSON",
			input:    "```JSON\n[{\"name\":\"test\"}]\n```",
			expected: `[{"name":"test"}]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := removeCodeBlock(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestHTTPClient_Execute_ErrorCases(t *testing.T) {
	t.Run("不正なJSONレスポンスでエラーを返す", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{invalid json`))
		}))
		defer server.Close()

		client := NewHTTPClient("test-api-key", 30*time.Second)
		client.baseURL = server.URL

		_, err := client.Execute(context.Background(), "テスト")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "パースエラー")
	})

	t.Run("空のPartsでエラーを返す", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := GenerateContentResponse{
				Candidates: []Candidate{
					{
						Content: ContentResponse{
							Parts: []PartResponse{}, // 空のParts
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewHTTPClient("test-api-key", 30*time.Second)
		client.baseURL = server.URL

		_, err := client.Execute(context.Background(), "テスト")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "コンテンツが空です")
	})

	t.Run("空のレスポンステキストでエラーを返す", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := GenerateContentResponse{
				Candidates: []Candidate{
					{
						Content: ContentResponse{
							Parts: []PartResponse{
								{Text: ""}, // 空のテキスト
							},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewHTTPClient("test-api-key", 30*time.Second)
		client.baseURL = server.URL

		_, err := client.Execute(context.Background(), "テスト")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "レスポンステキストが空です")
	})
}

func TestHTTPClient_HandleAPIError(t *testing.T) {
	t.Run("認証エラー_401", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": {"message": "API key not valid"}}`))
		}))
		defer server.Close()

		client := NewHTTPClient("invalid-key", 30*time.Second)
		client.baseURL = server.URL

		_, err := client.Execute(context.Background(), "テスト")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "認証エラー")
		assert.Contains(t, err.Error(), "APIキーが無効")
	})

	t.Run("レート制限_429", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": {"message": "Rate limit exceeded"}}`))
		}))
		defer server.Close()

		client := NewHTTPClient("test-api-key", 30*time.Second)
		client.baseURL = server.URL

		_, err := client.Execute(context.Background(), "テスト")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "レート制限")
	})

	t.Run("サーバーエラー_500", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": {"message": "Internal error"}}`))
		}))
		defer server.Close()

		client := NewHTTPClient("test-api-key", 30*time.Second)
		client.baseURL = server.URL

		_, err := client.Execute(context.Background(), "テスト")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "Geminiサービスエラー")
	})

	t.Run("エラーメッセージ付きの一般エラー", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": {"message": "Invalid request format"}}`))
		}))
		defer server.Close()

		client := NewHTTPClient("test-api-key", 30*time.Second)
		client.baseURL = server.URL

		_, err := client.Execute(context.Background(), "テスト")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid request format")
	})

	t.Run("エラーメッセージなしの一般エラー", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{}`)) // メッセージなし
		}))
		defer server.Close()

		client := NewHTTPClient("test-api-key", 30*time.Second)
		client.baseURL = server.URL

		_, err := client.Execute(context.Background(), "テスト")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "APIエラー (status 400)")
	})

	t.Run("アクセス拒否_403", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": {"message": "Forbidden"}}`))
		}))
		defer server.Close()

		client := NewHTTPClient("test-api-key", 30*time.Second)
		client.baseURL = server.URL

		_, err := client.Execute(context.Background(), "テスト")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "アクセス拒否")
	})

	t.Run("BadGateway_502", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer server.Close()

		client := NewHTTPClient("test-api-key", 30*time.Second)
		client.baseURL = server.URL

		_, err := client.Execute(context.Background(), "テスト")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "Geminiサービスエラー")
	})

	t.Run("ServiceUnavailable_503", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		client := NewHTTPClient("test-api-key", 30*time.Second)
		client.baseURL = server.URL

		_, err := client.Execute(context.Background(), "テスト")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "Geminiサービスエラー")
	})
}
