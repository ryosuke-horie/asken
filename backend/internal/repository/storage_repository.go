package repository

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

// StorageRepository はファイルストレージ操作を担当するインターフェース
type StorageRepository interface {
	// Upload はファイルをストレージにアップロードし、オブジェクト名を返す
	Upload(ctx context.Context, file io.Reader, filename string, contentType string) (string, error)

	// GetSignedURL は指定されたオブジェクトの署名付きURLを生成
	GetSignedURL(ctx context.Context, objectName string, expiration time.Duration) (string, error)

	// Delete は指定されたオブジェクトを削除
	Delete(ctx context.Context, objectName string) error
}

// cloudStorageRepository はCloud Storageを使用したStorageRepositoryの実装
type cloudStorageRepository struct {
	client     *storage.Client
	bucketName string
}

// NewStorageRepositoryCloudStorage は新しいCloud StorageベースのStorageRepositoryを作成します
func NewStorageRepositoryCloudStorage(client *storage.Client, bucketName string) StorageRepository {
	return &cloudStorageRepository{
		client:     client,
		bucketName: bucketName,
	}
}

// Upload はファイルをCloud Storageにアップロードし、オブジェクト名を返す
func (r *cloudStorageRepository) Upload(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
	// UUIDを生成してオブジェクト名を作成
	fileID := uuid.New().String()
	ext := filepath.Ext(filename)
	objectName := fmt.Sprintf("uploads/%s%s", fileID, ext)

	obj := r.client.Bucket(r.bucketName).Object(objectName)
	writer := obj.NewWriter(ctx)
	writer.ContentType = contentType

	if _, err := io.Copy(writer, file); err != nil {
		return "", fmt.Errorf("Cloud Storageへのアップロードに失敗: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("Cloud Storageへのアップロード完了に失敗: %w", err)
	}

	return objectName, nil
}

// GetSignedURL は指定されたオブジェクトの署名付きURLを生成
func (r *cloudStorageRepository) GetSignedURL(ctx context.Context, objectName string, expiration time.Duration) (string, error) {
	opts := &storage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(expiration),
	}

	url, err := r.client.Bucket(r.bucketName).SignedURL(objectName, opts)
	if err != nil {
		return "", fmt.Errorf("署名付きURLの生成に失敗: %w", err)
	}

	return url, nil
}

// Delete は指定されたオブジェクトをCloud Storageから削除
func (r *cloudStorageRepository) Delete(ctx context.Context, objectName string) error {
	obj := r.client.Bucket(r.bucketName).Object(objectName)

	if err := obj.Delete(ctx); err != nil {
		// オブジェクトが存在しない場合は無視（既に削除済み）
		if err == storage.ErrObjectNotExist {
			return nil
		}
		return fmt.Errorf("Cloud Storageからの削除に失敗: %w", err)
	}

	return nil
}
