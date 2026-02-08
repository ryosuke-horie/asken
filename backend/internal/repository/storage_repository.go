package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

// ErrObjectNotFound はオブジェクトが存在しない場合のエラー
var ErrObjectNotFound = errors.New("object not found")

const maxDownloadSize = 15 << 20 // 15MB: アップロード制限10MBに安全マージンを追加

// StorageRepository はファイルストレージ操作を担当するインターフェース
type StorageRepository interface {
	// Upload はファイルをストレージにアップロードし、オブジェクト名を返す
	Upload(ctx context.Context, file io.Reader, filename string, contentType string) (string, error)

	// Download は指定されたオブジェクトをダウンロードしてバイトデータを返す
	Download(ctx context.Context, objectName string) ([]byte, error)

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
func NewStorageRepositoryCloudStorage(client *storage.Client, bucketName string) (StorageRepository, error) {
	if client == nil {
		return nil, fmt.Errorf("storage client is required")
	}
	if bucketName == "" {
		return nil, fmt.Errorf("bucket name is required")
	}
	return &cloudStorageRepository{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// generateObjectName はファイル名からUUID付きのオブジェクト名を生成する
func generateObjectName(filename string) string {
	fileID := uuid.New().String()
	ext := filepath.Ext(filename)
	return fmt.Sprintf("uploads/%s%s", fileID, ext)
}

// convertStorageError はCloud Storageのエラーをドメインエラーに変換する
func convertStorageError(err error) error {
	if errors.Is(err, storage.ErrObjectNotExist) {
		return ErrObjectNotFound
	}
	return err
}

// Upload はファイルをCloud Storageにアップロードし、オブジェクト名を返す
func (r *cloudStorageRepository) Upload(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
	objectName := generateObjectName(filename)

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

// Download は指定されたオブジェクトをCloud Storageからダウンロードしてバイトデータを返す
func (r *cloudStorageRepository) Download(ctx context.Context, objectName string) ([]byte, error) {
	obj := r.client.Bucket(r.bucketName).Object(objectName)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, ErrObjectNotFound
		}
		return nil, fmt.Errorf("Cloud Storageからの読み取りに失敗: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(io.LimitReader(reader, maxDownloadSize))
	if err != nil {
		return nil, fmt.Errorf("Cloud Storageからのデータ読み取りに失敗: %w", err)
	}
	if int64(len(data)) >= maxDownloadSize {
		return nil, fmt.Errorf("ファイルサイズが上限(%dMB)を超えています", maxDownloadSize>>20)
	}

	return data, nil
}

// GetSignedURL は指定されたオブジェクトの署名付きURLを生成
func (r *cloudStorageRepository) GetSignedURL(ctx context.Context, objectName string, expiration time.Duration) (string, error) {
	// オブジェクトの存在確認
	obj := r.client.Bucket(r.bucketName).Object(objectName)
	_, err := obj.Attrs(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return "", ErrObjectNotFound
		}
		return "", fmt.Errorf("オブジェクト情報の取得に失敗: %w", err)
	}

	// BucketHandle.SignedURL()は認証情報を自動検出する
	// GCP環境ではデフォルト認証情報からSAメールを取得し、IAM signBlobで署名する
	// 前提条件: サービスアカウントにroles/iam.serviceAccountTokenCreator権限が必要
	opts := &storage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(expiration),
		Scheme:  storage.SigningSchemeV4,
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
		if errors.Is(err, storage.ErrObjectNotExist) {
			log.Printf("Debug: Cloud Storage object already deleted or not found: %s", objectName)
			return nil
		}
		return fmt.Errorf("Cloud Storageからの削除に失敗: %w", err)
	}

	return nil
}
