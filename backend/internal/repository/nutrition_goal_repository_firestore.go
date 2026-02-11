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
	TargetCalories float64   `firestore:"targetCalories"`
	UpdatedAt      time.Time `firestore:"updatedAt"`
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

// calculatePFCFromCalories はカロリーとPFC比率からグラム数を計算
func calculatePFCFromCalories(calories float64, ratio PFCRatio) (protein, fat, carbs float64) {
	// たんぱく質: 4kcal/g
	protein = (calories * ratio.Protein) / 4.0
	// 脂質: 9kcal/g
	fat = (calories * ratio.Fat) / 9.0
	// 炭水化物: 4kcal/g
	carbs = (calories * ratio.Carbohydrates) / 4.0
	return
}

// determineNutritionPhase は現在体重と目標体重から栄養フェーズを判定
func determineNutritionPhase(currentWeightKg, targetWeightKg float64) NutritionPhase {
	const threshold = 1.0 // 1kgの差

	if currentWeightKg-targetWeightKg > threshold {
		return NutritionPhaseWeightLoss
	} else if targetWeightKg-currentWeightKg > threshold {
		return NutritionPhaseWeightGain
	}
	return NutritionPhaseMaintenance
}

func (r *firestoreNutritionGoalRepository) GetGoal(ctx context.Context, userID string, currentWeightKg float64, targetWeightKg float64) (*NutritionGoal, error) {
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

	// フェーズ判定とPFC比率の計算
	phase := determineNutritionPhase(currentWeightKg, targetWeightKg)
	ratio := GetDefaultPFCRatio(phase)

	protein, fat, carbs := calculatePFCFromCalories(fsDoc.TargetCalories, ratio)

	return &NutritionGoal{
		TargetCalories:      fsDoc.TargetCalories,
		TargetProtein:       roundToOneDecimal(protein),
		TargetFat:           roundToOneDecimal(fat),
		TargetCarbohydrates: roundToOneDecimal(carbs),
		Phase:               phase,
		UpdatedAt:           fsDoc.UpdatedAt,
	}, nil
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
	ratio := GetDefaultPFCRatio(phase)

	protein, fat, carbs := calculatePFCFromCalories(targetCalories, ratio)

	return &NutritionGoal{
		TargetCalories:      targetCalories,
		TargetProtein:       roundToOneDecimal(protein),
		TargetFat:           roundToOneDecimal(fat),
		TargetCarbohydrates: roundToOneDecimal(carbs),
		Phase:               phase,
		UpdatedAt:           now,
	}, nil
}
