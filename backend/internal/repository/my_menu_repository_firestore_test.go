package repository

import (
	"testing"

	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
	"github.com/stretchr/testify/assert"
)

func TestCalculateTotals_WithMicronutrients(t *testing.T) {
	foods := []gemini.NutritionInfo{
		{
			Calories:       100,
			Protein:        10,
			Fat:            5,
			Carbohydrates:  20,
			Micronutrients: map[string]float64{"iron_mg": 2.0, "calcium_mg": 100.0},
		},
		{
			Calories:       200,
			Protein:        20,
			Fat:            10,
			Carbohydrates:  30,
			Micronutrients: map[string]float64{"iron_mg": 3.0, "zinc_mg": 1.5},
		},
	}

	calories, protein, fat, carbs, micro := calculateTotals(foods)

	assert.Equal(t, 300.0, calories)
	assert.Equal(t, 30.0, protein)
	assert.Equal(t, 15.0, fat)
	assert.Equal(t, 50.0, carbs)
	assert.Equal(t, 5.0, micro["iron_mg"])
	assert.Equal(t, 100.0, micro["calcium_mg"])
	assert.Equal(t, 1.5, micro["zinc_mg"])
}

func TestCalculateTotals_WithNilMicronutrients(t *testing.T) {
	foods := []gemini.NutritionInfo{
		{
			Calories:       100,
			Micronutrients: map[string]float64{"iron_mg": 2.0},
		},
		{
			Calories:       200,
			Micronutrients: nil, // micronutrientsなしの食品
		},
	}

	_, _, _, _, micro := calculateTotals(foods)

	assert.Equal(t, 2.0, micro["iron_mg"])
	assert.NotNil(t, micro)
}

func TestCalculateTotals_AllNilMicronutrients_ReturnsNilMap(t *testing.T) {
	foods := []gemini.NutritionInfo{
		{Calories: 100, Micronutrients: nil},
		{Calories: 200, Micronutrients: nil},
	}

	_, _, _, _, micro := calculateTotals(foods)

	// micronutrientsがない場合はomitemptyが機能するようnilを返すべき
	assert.Nil(t, micro)
}

func TestCalculateTotals_EmptyMicronutrients_ReturnsNilMap(t *testing.T) {
	foods := []gemini.NutritionInfo{
		{Calories: 100, Micronutrients: map[string]float64{}},
	}

	_, _, _, _, micro := calculateTotals(foods)

	// 空mapの場合もomitemptyが機能するようnilを返すべき
	assert.Nil(t, micro)
}

func TestCalculateTotals_SingleFood(t *testing.T) {
	foods := []gemini.NutritionInfo{
		{
			Calories:       500,
			Protein:        30,
			Fat:            20,
			Carbohydrates:  50,
			Micronutrients: map[string]float64{"vitamin_c_mg": 50.0},
		},
	}

	calories, protein, fat, carbs, micro := calculateTotals(foods)

	assert.Equal(t, 500.0, calories)
	assert.Equal(t, 30.0, protein)
	assert.Equal(t, 20.0, fat)
	assert.Equal(t, 50.0, carbs)
	assert.Equal(t, 50.0, micro["vitamin_c_mg"])
}
