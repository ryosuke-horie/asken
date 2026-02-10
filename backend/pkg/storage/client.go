package storage

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// NewStorageClient は Cloud Storage クライアントを作成する
// credentialsPath が空の場合は GOOGLE_APPLICATION_CREDENTIALS 環境変数を使用
func NewStorageClient(ctx context.Context, credentialsPath string) (*storage.Client, error) {
	var client *storage.Client
	var err error

	if credentialsPath != "" {
		opt := option.WithAuthCredentialsFile(option.ServiceAccount, credentialsPath)
		client, err = storage.NewClient(ctx, opt)
	} else {
		// GOOGLE_APPLICATION_CREDENTIALS 環境変数を使用
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("Cloud Storage Client 作成失敗: %w", err)
	}

	return client, nil
}
