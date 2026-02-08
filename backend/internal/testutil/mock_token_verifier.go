package testutil

import "context"

// MockTokenVerifier はTokenVerifierのモック実装
type MockTokenVerifier struct {
	VerifyFunc func(token string) (string, error)
}

// VerifyAndGetUID はトークン検証のモック
func (m *MockTokenVerifier) VerifyAndGetUID(_ context.Context, idToken string) (string, error) {
	if m.VerifyFunc != nil {
		return m.VerifyFunc(idToken)
	}
	return "mock-user-id", nil
}
