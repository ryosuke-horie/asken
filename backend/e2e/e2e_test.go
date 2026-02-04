//go:build e2e

package e2e

import (
	"os"
	"testing"
)

var testClient *Client
var authHelper *AuthHelper

func TestMain(m *testing.M) {
	var err error
	testClient, err = NewClient()
	if err != nil {
		os.Stderr.WriteString("Failed to create test client: " + err.Error() + "\n")
		os.Exit(1)
	}

	// 認証ヘルパーの初期化（E2E_FIREBASE_API_KEYが設定されている場合のみ）
	if os.Getenv("E2E_FIREBASE_API_KEY") != "" {
		authHelper, err = NewAuthHelper()
		if err != nil {
			os.Stderr.WriteString("Failed to create auth helper: " + err.Error() + "\n")
			os.Exit(1)
		}
	}

	os.Exit(m.Run())
}
