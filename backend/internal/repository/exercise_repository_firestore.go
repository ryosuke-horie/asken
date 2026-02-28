package repository

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// firestoreExerciseDocument はFirestoreに保存する運動記録ドキュメント構造
type firestoreExerciseDocument struct {
	ID                 string    `firestore:"id"`
	ExerciseName       string    `firestore:"exerciseName"`
	DurationMinutes    int       `firestore:"durationMinutes"`
	BurnedCaloriesKcal float64   `firestore:"burnedCaloriesKcal"`
	EstimationMethod   string    `firestore:"estimationMethod"`
	RecordedDate       string    `firestore:"recordedDate"`
	CreatedAt          time.Time `firestore:"createdAt"`
	UpdatedAt          time.Time `firestore:"updatedAt"`
}

// firestoreExerciseRepository はFirestoreを使用したExerciseRepositoryの実装
type firestoreExerciseRepository struct {
	client *firestore.Client
}

// NewExerciseRepository は新しいFirestoreベースのExerciseRepositoryを作成します
func NewExerciseRepository(client *firestore.Client) (ExerciseRepository, error) {
	if client == nil {
		return nil, fmt.Errorf("firestore client is required")
	}
	return &firestoreExerciseRepository{client: client}, nil
}

// getCollection はユーザーのexerciseRecordsコレクション参照を返す
func (r *firestoreExerciseRepository) getCollection(userID string) *firestore.CollectionRef {
	return r.client.Collection("users").Doc(userID).Collection("exerciseRecords")
}

func (r *firestoreExerciseRepository) Create(ctx context.Context, userID string, input CreateExerciseInput) (*ExerciseRecord, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	now := time.Now()
	id := uuid.New().String()

	doc := firestoreExerciseDocument{
		ID:                 id,
		ExerciseName:       input.ExerciseName,
		DurationMinutes:    input.DurationMinutes,
		BurnedCaloriesKcal: input.BurnedCaloriesKcal,
		EstimationMethod:   string(input.EstimationMethod),
		RecordedDate:       input.RecordedDate,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	_, err := r.getCollection(userID).Doc(id).Set(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("運動記録の作成に失敗: %w", err)
	}

	return toExerciseRecord(&doc), nil
}

func (r *firestoreExerciseRepository) ListByDate(ctx context.Context, userID string, recordedDate string) ([]ExerciseRecord, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	iter := r.getCollection(userID).
		Where("recordedDate", "==", recordedDate).
		OrderBy("createdAt", firestore.Asc).
		Documents(ctx)
	defer iter.Stop()

	var records []ExerciseRecord
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("運動記録の取得に失敗: %w", err)
		}

		var fsDoc firestoreExerciseDocument
		if err := doc.DataTo(&fsDoc); err != nil {
			return nil, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
		}

		records = append(records, *toExerciseRecord(&fsDoc))
	}

	return records, nil
}

func (r *firestoreExerciseRepository) Delete(ctx context.Context, userID string, recordID string) error {
	if userID == "" {
		return fmt.Errorf("userIDが必要です")
	}

	docRef := r.getCollection(userID).Doc(recordID)

	_, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return fmt.Errorf("運動記録が見つかりません: %s: %w", recordID, ErrNotFound)
		}
		return fmt.Errorf("運動記録の取得に失敗: %w", err)
	}

	_, err = docRef.Delete(ctx)
	if err != nil {
		return fmt.Errorf("運動記録の削除に失敗: %w", err)
	}

	return nil
}

// toExerciseRecord はFirestoreドキュメントをExerciseRecordに変換する
func toExerciseRecord(doc *firestoreExerciseDocument) *ExerciseRecord {
	return &ExerciseRecord{
		ID:                 doc.ID,
		ExerciseName:       doc.ExerciseName,
		DurationMinutes:    doc.DurationMinutes,
		BurnedCaloriesKcal: doc.BurnedCaloriesKcal,
		EstimationMethod:   EstimationMethod(doc.EstimationMethod),
		RecordedDate:       doc.RecordedDate,
		CreatedAt:          doc.CreatedAt,
		UpdatedAt:          doc.UpdatedAt,
	}
}
