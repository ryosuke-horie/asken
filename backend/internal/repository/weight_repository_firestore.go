package repository

import (
	"context"
	"fmt"
	"math"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// firestoreWeightRecordDocument はFirestoreに保存する体重記録ドキュメント構造
type firestoreWeightRecordDocument struct {
	ID         string    `firestore:"id"`
	WeightKg   float64   `firestore:"weightKg"`
	RecordedAt time.Time `firestore:"recordedAt"`
	Note       string    `firestore:"note,omitempty"`
	CreatedAt  time.Time `firestore:"createdAt"`
	UpdatedAt  time.Time `firestore:"updatedAt"`
}

// firestoreWeightGoalDocument はFirestoreに保存する目標体重ドキュメント構造
type firestoreWeightGoalDocument struct {
	TargetWeightKg float64   `firestore:"targetWeightKg"`
	UpdatedAt      time.Time `firestore:"updatedAt"`
}

// firestoreWeightRecordRepository はFirestoreを使用したWeightRecordRepositoryの実装
type firestoreWeightRecordRepository struct {
	client *firestore.Client
}

// NewWeightRecordRepositoryFirestore は新しいFirestoreベースのWeightRecordRepositoryを作成します
func NewWeightRecordRepositoryFirestore(client *firestore.Client) (WeightRecordRepository, error) {
	if client == nil {
		return nil, fmt.Errorf("firestore client is required")
	}
	return &firestoreWeightRecordRepository{client: client}, nil
}

// getUserWeightRecordsCollection はユーザーのweightRecordsコレクション参照を取得
func (r *firestoreWeightRecordRepository) getUserWeightRecordsCollection(userID string) *firestore.CollectionRef {
	return r.client.Collection("users").Doc(userID).Collection("weightRecords")
}

// getUserWeightGoalDoc はユーザーのweightGoalドキュメント参照を取得
func (r *firestoreWeightRecordRepository) getUserWeightGoalDoc(userID string) *firestore.DocumentRef {
	return r.client.Collection("users").Doc(userID).Collection("weightGoal").Doc("current")
}

// roundToOneDecimal は小数点1桁に丸める
func roundToOneDecimal(v float64) float64 {
	return math.Round(v*10) / 10
}

func (r *firestoreWeightRecordRepository) CreateRecord(ctx context.Context, userID string, weightKg float64, recordedAt time.Time, note string) (*WeightRecord, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	now := time.Now()
	id := uuid.New().String()
	rounded := roundToOneDecimal(weightKg)

	doc := firestoreWeightRecordDocument{
		ID:         id,
		WeightKg:   rounded,
		RecordedAt: recordedAt.UTC(),
		Note:       note,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	_, err := r.getUserWeightRecordsCollection(userID).Doc(id).Set(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("体重記録の作成に失敗: %w", err)
	}

	return r.toWeightRecord(&doc), nil
}

func (r *firestoreWeightRecordRepository) GetRecord(ctx context.Context, userID string, recordID string) (*WeightRecord, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	doc, err := r.getUserWeightRecordsCollection(userID).Doc(recordID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("体重記録が見つかりません: %s: %w", recordID, ErrNotFound)
		}
		return nil, fmt.Errorf("体重記録の取得に失敗: %w", err)
	}

	var fsDoc firestoreWeightRecordDocument
	if err := doc.DataTo(&fsDoc); err != nil {
		return nil, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
	}

	return r.toWeightRecord(&fsDoc), nil
}

func (r *firestoreWeightRecordRepository) UpdateRecord(ctx context.Context, userID string, recordID string, weightKg float64, note string) (*WeightRecord, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	docRef := r.getUserWeightRecordsCollection(userID).Doc(recordID)

	// ドキュメントの存在確認
	existingDoc, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("体重記録が見つかりません: %s: %w", recordID, ErrNotFound)
		}
		return nil, fmt.Errorf("体重記録の取得に失敗: %w", err)
	}

	var fsDoc firestoreWeightRecordDocument
	if err := existingDoc.DataTo(&fsDoc); err != nil {
		return nil, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
	}

	now := time.Now()
	rounded := roundToOneDecimal(weightKg)

	_, err = docRef.Update(ctx, []firestore.Update{
		{Path: "weightKg", Value: rounded},
		{Path: "note", Value: note},
		{Path: "updatedAt", Value: now},
	})
	if err != nil {
		return nil, fmt.Errorf("体重記録の更新に失敗: %w", err)
	}

	fsDoc.WeightKg = rounded
	fsDoc.Note = note
	fsDoc.UpdatedAt = now

	return r.toWeightRecord(&fsDoc), nil
}

func (r *firestoreWeightRecordRepository) DeleteRecord(ctx context.Context, userID string, recordID string) error {
	if userID == "" {
		return fmt.Errorf("userIDが必要です")
	}

	docRef := r.getUserWeightRecordsCollection(userID).Doc(recordID)

	// 存在確認
	_, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return fmt.Errorf("体重記録が見つかりません: %s: %w", recordID, ErrNotFound)
		}
		return fmt.Errorf("体重記録の取得に失敗: %w", err)
	}

	_, err = docRef.Delete(ctx)
	if err != nil {
		return fmt.Errorf("体重記録の削除に失敗: %w", err)
	}

	return nil
}

func (r *firestoreWeightRecordRepository) ListRecordsByDateRange(ctx context.Context, userID string, from time.Time, to time.Time) ([]WeightRecord, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	iter := r.getUserWeightRecordsCollection(userID).
		Where("recordedAt", ">=", from.UTC()).
		Where("recordedAt", "<=", to.UTC()).
		OrderBy("recordedAt", firestore.Asc).
		Documents(ctx)
	defer iter.Stop()

	var records []WeightRecord
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("体重記録の取得に失敗: %w", err)
		}

		var fsDoc firestoreWeightRecordDocument
		if err := doc.DataTo(&fsDoc); err != nil {
			return nil, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
		}

		records = append(records, *r.toWeightRecord(&fsDoc))
	}

	return records, nil
}

func (r *firestoreWeightRecordRepository) GetGoal(ctx context.Context, userID string) (*WeightGoal, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	doc, err := r.getUserWeightGoalDoc(userID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("目標体重の取得に失敗: %w", err)
	}

	var fsDoc firestoreWeightGoalDocument
	if err := doc.DataTo(&fsDoc); err != nil {
		return nil, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
	}

	return &WeightGoal{
		TargetWeightKg: fsDoc.TargetWeightKg,
		UpdatedAt:      fsDoc.UpdatedAt,
	}, nil
}

func (r *firestoreWeightRecordRepository) SetGoal(ctx context.Context, userID string, targetWeightKg float64) (*WeightGoal, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	now := time.Now()
	rounded := roundToOneDecimal(targetWeightKg)

	doc := firestoreWeightGoalDocument{
		TargetWeightKg: rounded,
		UpdatedAt:      now,
	}

	_, err := r.getUserWeightGoalDoc(userID).Set(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("目標体重の設定に失敗: %w", err)
	}

	return &WeightGoal{
		TargetWeightKg: rounded,
		UpdatedAt:      now,
	}, nil
}

// toWeightRecord はFirestoreドキュメントをWeightRecordに変換
func (r *firestoreWeightRecordRepository) toWeightRecord(doc *firestoreWeightRecordDocument) *WeightRecord {
	return &WeightRecord{
		ID:         doc.ID,
		WeightKg:   doc.WeightKg,
		RecordedAt: doc.RecordedAt,
		Note:       doc.Note,
		CreatedAt:  doc.CreatedAt,
		UpdatedAt:  doc.UpdatedAt,
	}
}
