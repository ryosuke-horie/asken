package repository

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// firestoreNutritionGoalDocument はFirestoreに保存する栄養目標ドキュメント構造
type firestoreNutritionGoalDocument struct {
	TargetCalories       float64            `firestore:"targetCalories"`
	MicronutrientTargets map[string]float64 `firestore:"micronutrientTargets,omitempty"`
	UpdatedAt            time.Time          `firestore:"updatedAt"`
}

// firestoreNutritionGoalRepository はFirestoreを使用したNutritionGoalRepositoryの実装
type firestoreNutritionGoalRepository struct {
	client *firestore.Client
}

// NewNutritionGoalRepository は新しいFirestoreベースの栄養目標リポジトリを作成します
func NewNutritionGoalRepository(client *firestore.Client) (NutritionGoalRepository, error) {
	if client == nil {
		return nil, fmt.Errorf("firestore client is required")
	}
	return &firestoreNutritionGoalRepository{client: client}, nil
}

// getUserNutritionGoalDoc はユーザーのnutritionGoalドキュメント参照を取得
func (r *firestoreNutritionGoalRepository) getUserNutritionGoalDoc(userID string) *firestore.DocumentRef {
	return r.client.Collection("users").Doc(userID).Collection("nutritionGoal").Doc("current")
}

func (r *firestoreNutritionGoalRepository) GetGoal(ctx context.Context, userID string, currentWeightKg *float64, targetWeightKg *float64) (*NutritionGoal, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	doc, err := r.getUserNutritionGoalDoc(userID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("栄養目標の取得に失敗: %w", err)
	}

	var fsDoc firestoreNutritionGoalDocument
	if err := doc.DataTo(&fsDoc); err != nil {
		return nil, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
	}

	phase := DeterminePhase(currentWeightKg, targetWeightKg)
	calculated := CalculateNutritionGoal(fsDoc.TargetCalories, phase)
	calculated.UpdatedAt = fsDoc.UpdatedAt

	// ユーザーがカスタマイズしたマイクロニュートリエント目標がある場合は上書き
	if fsDoc.MicronutrientTargets != nil {
		for k, v := range fsDoc.MicronutrientTargets {
			calculated.MicronutrientTargets[k] = v
		}
	}

	return calculated, nil
}

func (r *firestoreNutritionGoalRepository) SetGoal(ctx context.Context, userID string, targetCalories float64) (*NutritionGoal, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	if err := ValidateTargetCalories(targetCalories); err != nil {
		return nil, err
	}

	now := time.Now()

	doc := firestoreNutritionGoalDocument{
		TargetCalories: targetCalories,
		UpdatedAt:      now,
	}

	_, err := r.getUserNutritionGoalDoc(userID).Set(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("栄養目標の設定に失敗: %w", err)
	}

	// デフォルトは維持期で返す（現在体重が不明なため）
	phase := NutritionPhaseMaintenance
	calculated := CalculateNutritionGoal(targetCalories, phase)
	calculated.UpdatedAt = now
	return calculated, nil
}
