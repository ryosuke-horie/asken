package repository

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateTargetCalories(t *testing.T) {
	tests := []struct {
		name     string
		calories float64
		wantErr  bool
	}{
		{"最小値", MinTargetCalories, false},
		{"最大値", MaxTargetCalories, false},
		{"通常値", 2000.0, false},
		{"最小値未満", 799.0, true},
		{"最大値超過", 5001.0, true},
		{"NaN", math.NaN(), true},
		{"正の無限大", math.Inf(1), true},
		{"負の無限大", math.Inf(-1), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTargetCalories(tt.calories)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPFCRatio_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ratio   PFCRatio
		wantErr bool
	}{
		{
			"有効な比率",
			PFCRatio{Protein: 0.30, Fat: 0.20, Carbohydrates: 0.50},
			false,
		},
		{
			"合計が1.0でない",
			PFCRatio{Protein: 0.30, Fat: 0.30, Carbohydrates: 0.50},
			true,
		},
		{
			"たんぱく質が範囲外",
			PFCRatio{Protein: -0.1, Fat: 0.50, Carbohydrates: 0.60},
			true,
		},
		{
			"脂質が範囲外",
			PFCRatio{Protein: 0.20, Fat: 1.1, Carbohydrates: -0.30},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ratio.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetDefaultPFCRatio(t *testing.T) {
	tests := []struct {
		name  string
		phase NutritionPhase
		want  PFCRatio
	}{
		{"減量期", NutritionPhaseWeightLoss, PFCRatio{Protein: 0.30, Fat: 0.20, Carbohydrates: 0.50}},
		{"維持期", NutritionPhaseMaintenance, PFCRatio{Protein: 0.20, Fat: 0.25, Carbohydrates: 0.55}},
		{"増量期", NutritionPhaseWeightGain, PFCRatio{Protein: 0.20, Fat: 0.30, Carbohydrates: 0.50}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDefaultPFCRatio(tt.phase)
			assert.Equal(t, tt.want, got)
			assert.NoError(t, got.Validate(), "デフォルト比率は常に有効であるべき")
		})
	}
}

func TestDeterminePhase(t *testing.T) {
	floatPtr := func(v float64) *float64 { return &v }

	tests := []struct {
		name          string
		currentWeight *float64
		targetWeight  *float64
		want          NutritionPhase
	}{
		{"currentWeightがnilの場合は維持期", nil, floatPtr(70.0), NutritionPhaseMaintenance},
		{"targetWeightがnilの場合は維持期", floatPtr(70.0), nil, NutritionPhaseMaintenance},
		{"両方nilの場合は維持期", nil, nil, NutritionPhaseMaintenance},
		{"差が1kg以内は維持期", floatPtr(70.0), floatPtr(70.5), NutritionPhaseMaintenance},
		{"現在>目標で1kg超は減量期", floatPtr(75.0), floatPtr(70.0), NutritionPhaseWeightLoss},
		{"目標>現在で1kg超は増量期", floatPtr(60.0), floatPtr(70.0), NutritionPhaseWeightGain},
		{"ちょうど1kgの差は維持期", floatPtr(71.0), floatPtr(70.0), NutritionPhaseMaintenance},
		{"1.01kgの差は減量期", floatPtr(71.01), floatPtr(70.0), NutritionPhaseWeightLoss},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeterminePhase(tt.currentWeight, tt.targetWeight)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCalculateNutritionGoal(t *testing.T) {
	tests := []struct {
		name     string
		calories float64
		phase    NutritionPhase
	}{
		{"減量期_2000kcal", 2000.0, NutritionPhaseWeightLoss},
		{"維持期_2000kcal", 2000.0, NutritionPhaseMaintenance},
		{"増量期_3000kcal", 3000.0, NutritionPhaseWeightGain},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateNutritionGoal(tt.calories, tt.phase)
			require.NotNil(t, got)

			assert.Equal(t, tt.calories, got.TargetCalories)
			assert.Equal(t, tt.phase, got.Phase)
			assert.Greater(t, got.TargetProtein, 0.0)
			assert.Greater(t, got.TargetFat, 0.0)
			assert.Greater(t, got.TargetCarbohydrates, 0.0)

			// PFC合計カロリーが目標カロリーと近似することを確認
			totalCalories := got.TargetProtein*4 + got.TargetFat*9 + got.TargetCarbohydrates*4
			assert.InDelta(t, tt.calories, totalCalories, 5.0, "PFC合計カロリーは目標カロリーと近似すべき")
		})
	}
}

func TestCalculateNutritionGoal_PFCValues(t *testing.T) {
	// 2000kcal + 維持期 (P:20%, F:25%, C:55%) の具体的な値を検証
	goal := CalculateNutritionGoal(2000.0, NutritionPhaseMaintenance)

	// P: 2000 * 0.20 / 4 = 100.0g
	assert.Equal(t, 100.0, goal.TargetProtein)
	// F: 2000 * 0.25 / 9 = 55.555... → 55.6g
	assert.Equal(t, 55.6, goal.TargetFat)
	// C: 2000 * 0.55 / 4 = 275.0g
	assert.Equal(t, 275.0, goal.TargetCarbohydrates)
}
