//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// CleanupTestData はE2Eテストで作成されたFirestoreのテストデータを削除する
func CleanupTestData(ctx context.Context) error {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = os.Getenv("GCP_PROJECT")
	}
	if projectID == "" {
		return fmt.Errorf("neither GOOGLE_CLOUD_PROJECT nor GCP_PROJECT environment variable is set")
	}

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to create Firestore client: %w", err)
	}
	defer client.Close()

	// テストデータのコレクションをクリーンアップ
	collections := []string{
		fmt.Sprintf("users/%s/analysisRequests", testUID()),
		fmt.Sprintf("users/%s/weightRecords", testUID()),
		fmt.Sprintf("users/%s/weightGoal", testUID()),
		fmt.Sprintf("users/%s/ingredients", testUID()),
		fmt.Sprintf("users/%s/menuSuggestions", testUID()),
		fmt.Sprintf("users/%s/myMenu", testUID()),
		fmt.Sprintf("users/%s/exerciseRecords", testUID()),
	}

	for _, collPath := range collections {
		if err := deleteCollection(ctx, client, collPath); err != nil {
			return fmt.Errorf("failed to cleanup %s: %w", collPath, err)
		}
	}
	return nil
}

// deleteCollection は指定コレクション内の全ドキュメントを削除する
func deleteCollection(ctx context.Context, client *firestore.Client, path string) error {
	iter := client.Collection(path).Documents(ctx)
	defer iter.Stop()

	bw := client.BulkWriter(ctx)
	defer func() {
		bw.Flush()
		bw.End()
	}()

	var deleteJobs []*firestore.BulkWriterJob
	count := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate documents: %w", err)
		}

		job, err := bw.Delete(doc.Ref)
		if err != nil {
			return fmt.Errorf("failed to add delete job for %s: %w", doc.Ref.Path, err)
		}
		deleteJobs = append(deleteJobs, job)
		count++
	}

	// 削除結果を確認
	for _, job := range deleteJobs {
		if _, err := job.Results(); err != nil {
			return fmt.Errorf("failed to delete document: %w", err)
		}
	}

	fmt.Fprintf(os.Stderr, "Cleaned up %d test documents from %s\n", count, path)
	return nil
}
