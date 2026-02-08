//go:build e2e

package e2e

import (
	"context"
	"os"
	"testing"
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

	// 認証ヘルパーの初期化（E2E_FIREBASE_API_KEYが設定されている場合のみ）
	if os.Getenv("E2E_FIREBASE_API_KEY") != "" {
		authHelper, err = NewAuthHelper(ctx)
		if err != nil {
			os.Stderr.WriteString("Failed to create auth helper: " + err.Error() + "\n")
			os.Exit(1)
		}
	}

	exitCode := m.Run()

	// テストデータのクリーンアップ
	if err := CleanupTestData(ctx); err != nil {
		os.Stderr.WriteString("Failed to cleanup test data: " + err.Error() + "\n")
	}

	os.Exit(exitCode)
}
