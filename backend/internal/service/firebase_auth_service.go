package service

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// FirebaseAuthService は Firebase Authentication のトークン検証を行うサービス
type FirebaseAuthService struct {
	client *auth.Client
}

// NewFirebaseAuthService は FirebaseAuthService を作成する
// credentialsPath が空の場合は GOOGLE_APPLICATION_CREDENTIALS 環境変数を使用
func NewFirebaseAuthService(ctx context.Context, credentialsPath string) (*FirebaseAuthService, error) {
	var app *firebase.App
	var err error

	if credentialsPath != "" {
		opt := option.WithCredentialsFile(credentialsPath)
		app, err = firebase.NewApp(ctx, nil, opt)
	} else {
		// GOOGLE_APPLICATION_CREDENTIALS 環境変数を使用
		app, err = firebase.NewApp(ctx, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("Firebase App 初期化失敗: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("Firebase Auth Client 作成失敗: %w", err)
	}

	return &FirebaseAuthService{client: client}, nil
}

// VerifyIDToken は Firebase ID トークンを検証し、ユーザー情報を返す
func (s *FirebaseAuthService) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	token, err := s.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("ID トークン検証失敗: %w", err)
	}
	return token, nil
}

// GetUserUID は検証済みトークンからユーザー UID を取得する
func (s *FirebaseAuthService) GetUserUID(token *auth.Token) string {
	return token.UID
}

// VerifyAndGetUID は Firebase ID トークンを検証し、UID を直接返す
func (s *FirebaseAuthService) VerifyAndGetUID(ctx context.Context, idToken string) (string, error) {
	token, err := s.VerifyIDToken(ctx, idToken)
	if err != nil {
		return "", err
	}
	return token.UID, nil
}
