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
	testUID := os.Getenv("E2E_TEST_UID")
	if testUID == "" {
		testUID = "e2e-test-user"
	}

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = os.Getenv("GCP_PROJECT")
	}

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to create Firestore client: %w", err)
	}
	defer client.Close()

	return deleteCollection(ctx, client, fmt.Sprintf("users/%s/analysisRequests", testUID))
}

// deleteCollection は指定コレクション内の全ドキュメントを削除する
func deleteCollection(ctx context.Context, client *firestore.Client, path string) error {
	iter := client.Collection(path).Documents(ctx)
	defer iter.Stop()

	batch := client.BulkWriter(ctx)
	count := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate documents: %w", err)
		}

		batch.Delete(doc.Ref)
		count++
	}

	batch.End()
	fmt.Fprintf(os.Stderr, "Cleaned up %d test documents from %s\n", count, path)
	return nil
}
