//go:build e2e

package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var testClient *Client
var authHelper *AuthHelper

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	testClient, err = NewClient()
	if err != nil {
		os.Stderr.WriteString("Failed to create test client: " + err.Error() + "\n")
		os.Exit(1)
	}

	// 認証ヘルパーの初期化
	if os.Getenv("E2E_FIREBASE_API_KEY") != "" {
		authHelper, err = NewAuthHelper(ctx)
		if err != nil {
			os.Stderr.WriteString("Failed to create auth helper: " + err.Error() + "\n")
			os.Exit(1)
		}
	} else if os.Getenv("CI") != "" {
		// CI環境ではE2E_FIREBASE_API_KEYが必須（設定漏れによるサイレントスキップを防止）
		os.Stderr.WriteString("E2E_FIREBASE_API_KEY is required in CI environment\n")
		os.Exit(1)
	}

	exitCode := m.Run()

	// テストデータのクリーンアップ
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := CleanupTestData(cleanupCtx); err != nil {
		os.Stderr.WriteString("Failed to cleanup test data: " + err.Error() + "\n")
	}
	cleanupCancel()

	os.Exit(exitCode)
}

// authenticatedClient は認証済みクライアントとコンテキストを返すヘルパー
// 認証が利用できない場合はテストをスキップする
func authenticatedClient(t *testing.T, timeout time.Duration) (*Client, context.Context) {
	t.Helper()
	if authHelper == nil {
		t.Skip("E2E_FIREBASE_API_KEY is not set, skipping authenticated tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)

	token, err := authHelper.GetTestToken(ctx)
	require.NoError(t, err, "Failed to get test token")

	return testClient.WithAuthToken(token), ctx
}
