package service

import (
	"context"
	"errors"
	"testing"

	"firebase.google.com/go/v4/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAuthClient はテスト用のauthClient実装
type mockAuthClient struct {
	verifyIDTokenFunc func(ctx context.Context, idToken string) (*auth.Token, error)
}

func (m *mockAuthClient) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	return m.verifyIDTokenFunc(ctx, idToken)
}

func TestFirebaseAuthService_VerifyIDToken_Success(t *testing.T) {
	expectedToken := &auth.Token{UID: "test-uid-123"}
	mock := &mockAuthClient{
		verifyIDTokenFunc: func(ctx context.Context, idToken string) (*auth.Token, error) {
			assert.Equal(t, "valid-token", idToken)
			return expectedToken, nil
		},
	}

	svc := newFirebaseAuthServiceWithClient(mock)
	token, err := svc.VerifyIDToken(context.Background(), "valid-token")

	require.NoError(t, err)
	assert.Equal(t, "test-uid-123", token.UID)
}

func TestFirebaseAuthService_VerifyIDToken_Error(t *testing.T) {
	mock := &mockAuthClient{
		verifyIDTokenFunc: func(ctx context.Context, idToken string) (*auth.Token, error) {
			return nil, errors.New("token expired")
		},
	}

	svc := newFirebaseAuthServiceWithClient(mock)
	token, err := svc.VerifyIDToken(context.Background(), "expired-token")

	assert.Nil(t, token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID トークン検証失敗")
	assert.Contains(t, err.Error(), "token expired")
}

func TestFirebaseAuthService_GetUserUID(t *testing.T) {
	mock := &mockAuthClient{}
	svc := newFirebaseAuthServiceWithClient(mock)

	token := &auth.Token{UID: "user-abc-456"}
	uid := svc.GetUserUID(token)

	assert.Equal(t, "user-abc-456", uid)
}

func TestFirebaseAuthService_VerifyAndGetUID_Success(t *testing.T) {
	mock := &mockAuthClient{
		verifyIDTokenFunc: func(ctx context.Context, idToken string) (*auth.Token, error) {
			return &auth.Token{UID: "verified-uid"}, nil
		},
	}

	svc := newFirebaseAuthServiceWithClient(mock)
	uid, err := svc.VerifyAndGetUID(context.Background(), "valid-token")

	require.NoError(t, err)
	assert.Equal(t, "verified-uid", uid)
}

func TestFirebaseAuthService_VerifyAndGetUID_Error(t *testing.T) {
	mock := &mockAuthClient{
		verifyIDTokenFunc: func(ctx context.Context, idToken string) (*auth.Token, error) {
			return nil, errors.New("invalid token")
		},
	}

	svc := newFirebaseAuthServiceWithClient(mock)
	uid, err := svc.VerifyAndGetUID(context.Background(), "bad-token")

	assert.Empty(t, uid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID トークン検証失敗")
}
