//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
)

// AuthHelper はE2Eテスト用の認証ヘルパー
type AuthHelper struct {
	firebaseAPIKey string
	httpClient     *http.Client
	authClient     *auth.Client
}

// NewAuthHelper は新しいAuthHelperを作成する
// Firebase Appとauth.Clientを初期化時にキャッシュする
func NewAuthHelper(ctx context.Context) (*AuthHelper, error) {
	apiKey := os.Getenv("E2E_FIREBASE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("E2E_FIREBASE_API_KEY environment variable is required")
	}

	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Firebase app: %w", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Auth client: %w", err)
	}

	return &AuthHelper{
		firebaseAPIKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		authClient: authClient,
	}, nil
}

// GetTestToken はテスト用のFirebase IDトークンを取得する
func (h *AuthHelper) GetTestToken(ctx context.Context) (string, error) {
	customToken, err := h.authClient.CustomToken(ctx, testUID())
	if err != nil {
		return "", fmt.Errorf("failed to create custom token: %w", err)
	}

	idToken, err := h.exchangeCustomTokenForIDToken(ctx, customToken)
	if err != nil {
		return "", fmt.Errorf("failed to exchange token: %w", err)
	}

	return idToken, nil
}

// exchangeCustomTokenForIDToken はカスタムトークンをIDトークンに交換する
func (h *AuthHelper) exchangeCustomTokenForIDToken(ctx context.Context, customToken string) (string, error) {
	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithCustomToken?key=%s", h.firebaseAPIKey)

	reqBody := map[string]any{
		"token":             customToken,
		"returnSecureToken": true,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Firebase API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return "", fmt.Errorf("Firebase API returned status %d (failed to read body: %w)", resp.StatusCode, readErr)
		}
		return "", fmt.Errorf("Firebase API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		IDToken string `json:"idToken"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if result.IDToken == "" {
		return "", fmt.Errorf("Firebase API returned empty idToken")
	}

	return result.IDToken, nil
}
