package repository

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
)

const (
	MinTargetCalories = 800.0
	MaxTargetCalories = 5000.0
)

// ValidateTargetCalories は目標カロリーの範囲バリデーション
func ValidateTargetCalories(calories float64) error {
	if math.IsNaN(calories) || math.IsInf(calories, 0) {
		return fmt.Errorf("target_caloriesに無効な値が指定されています")
	}
	if calories < MinTargetCalories || calories > MaxTargetCalories {
		return fmt.Errorf("target_caloriesは%.0f〜%.0fの範囲で指定してください", MinTargetCalories, MaxTargetCalories)
	}
	return nil
}

// PFCRatio はPFCバランス比率を表す構造体
type PFCRatio struct {
	Protein       float64 `json:"protein"`       // たんぱく質比率 (0.0-1.0)
	Fat           float64 `json:"fat"`           // 脂質比率 (0.0-1.0)
	Carbohydrates float64 `json:"carbohydrates"` // 炭水化物比率 (0.0-1.0)
}

// Validate はPFC比率のバリデーション
func (p PFCRatio) Validate() error {
	const tolerance = 0.001

	if p.Protein < 0 || p.Protein > 1 {
		return fmt.Errorf("たんぱく質比率は0.0〜1.0の範囲で指定してください")
	}
	if p.Fat < 0 || p.Fat > 1 {
		return fmt.Errorf("脂質比率は0.0〜1.0の範囲で指定してください")
	}
	if p.Carbohydrates < 0 || p.Carbohydrates > 1 {
		return fmt.Errorf("炭水化物比率は0.0〜1.0の範囲で指定してください")
	}

	total := p.Protein + p.Fat + p.Carbohydrates
	if math.Abs(total-1.0) > tolerance {
		return fmt.Errorf("PFC比率の合計は1.0（100%%）である必要があります。現在: %.2f", total)
	}

	return nil
}

// NutritionPhase は栄養フェーズを表す
type NutritionPhase string

const (
	NutritionPhaseWeightLoss  NutritionPhase = "weight_loss" // 減量期
	NutritionPhaseMaintenance NutritionPhase = "maintenance" // 維持期
	NutritionPhaseWeightGain  NutritionPhase = "weight_gain" // 増量期
)

// GetDefaultPFCRatio はフェーズに応じたデフォルトPFC比率を返す
func GetDefaultPFCRatio(phase NutritionPhase) PFCRatio {
	switch phase {
	case NutritionPhaseWeightLoss:
		// 減量期: 高たんぱく質
		return PFCRatio{Protein: 0.30, Fat: 0.20, Carbohydrates: 0.50}
	case NutritionPhaseWeightGain:
		// 増量期: 高脂質
		return PFCRatio{Protein: 0.20, Fat: 0.30, Carbohydrates: 0.50}
	case NutritionPhaseMaintenance:
		// 維持期: バランス型
		fallthrough
	default:
		return PFCRatio{Protein: 0.20, Fat: 0.25, Carbohydrates: 0.55}
	}
}

// NutritionGoal は栄養目標を表す構造体
type NutritionGoal struct {
	TargetCalories       float64            `json:"target_calories"`                 // 目標カロリー（kcal）- ユーザー設定
	TargetProtein        float64            `json:"target_protein"`                  // 目標たんぱく質（g）- 自動計算
	TargetFat            float64            `json:"target_fat"`                      // 目標脂質（g）- 自動計算
	TargetCarbohydrates  float64            `json:"target_carbohydrates"`            // 目標炭水化物（g）- 自動計算
	Phase                NutritionPhase     `json:"phase"`                           // 栄養フェーズ
	MicronutrientTargets map[string]float64 `json:"micronutrient_targets,omitempty"` // マイクロニュートリエント目標値
	UpdatedAt            time.Time          `json:"updated_at"`
}

// NutritionGoalRepository は栄養目標の永続化を担当するインターフェース
type NutritionGoalRepository interface {
	// GetGoal は栄養目標を取得します
	// 目標が未設定の場合は (nil, nil) を返します
	GetGoal(ctx context.Context, userID string, currentWeightKg *float64, targetWeightKg *float64) (*NutritionGoal, error)

	// SetGoal は目標カロリーを設定・更新します
	// PFC値は現在体重と目標体重から自動計算されます
	SetGoal(ctx context.Context, userID string, targetCalories float64) (*NutritionGoal, error)
}

// CalculateNutritionGoal は目標カロリーとフェーズから栄養目標を計算します
func CalculateNutritionGoal(targetCalories float64, phase NutritionPhase) *NutritionGoal {
	ratio := GetDefaultPFCRatio(phase)

	// たんぱく質: 1g = 4kcal
	proteinG := (targetCalories * ratio.Protein) / 4
	// 脂質: 1g = 9kcal
	fatG := (targetCalories * ratio.Fat) / 9
	// 炭水化物: 1g = 4kcal
	carbsG := (targetCalories * ratio.Carbohydrates) / 4

	return &NutritionGoal{
		TargetCalories:       targetCalories,
		TargetProtein:        roundToOneDecimal(proteinG),
		TargetFat:            roundToOneDecimal(fatG),
		TargetCarbohydrates:  roundToOneDecimal(carbsG),
		Phase:                phase,
		MicronutrientTargets: gemini.DefaultMicronutrientTargets(),
		UpdatedAt:            time.Now(),
	}
}

// DeterminePhase は現在体重と目標体重からフェーズを判定します
// currentWeightKg または targetWeightKg が nil の場合は維持期を返します
func DeterminePhase(currentWeightKg *float64, targetWeightKg *float64) NutritionPhase {
	if currentWeightKg == nil || targetWeightKg == nil {
		return NutritionPhaseMaintenance
	}
	diff := *currentWeightKg - *targetWeightKg
	// 1kg以上の差分があれば減量期または増量期と判定
	if diff > 1.0 {
		return NutritionPhaseWeightLoss
	} else if diff < -1.0 {
		return NutritionPhaseWeightGain
	}
	return NutritionPhaseMaintenance
}
