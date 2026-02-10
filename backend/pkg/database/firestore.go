package database

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

// NewFirestoreClient は Firestore クライアントを作成する
// credentialsPath が空の場合は GOOGLE_APPLICATION_CREDENTIALS 環境変数を使用
func NewFirestoreClient(ctx context.Context, credentialsPath string) (*firestore.Client, error) {
	var app *firebase.App
	var err error

	if credentialsPath != "" {
		opt := option.WithAuthCredentialsFile(option.ServiceAccount, credentialsPath)
		app, err = firebase.NewApp(ctx, nil, opt)
	} else {
		// GOOGLE_APPLICATION_CREDENTIALS 環境変数を使用
		app, err = firebase.NewApp(ctx, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("Firebase App 初期化失敗: %w", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("Firestore Client 作成失敗: %w", err)
	}

	return client, nil
}
