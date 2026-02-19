package gemini

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatQuantity(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		unit     string
		expected string
	}{
		{"整数値_g", 150, "g", "150g"},
		{"整数値_杯", 1, "杯", "1杯"},
		{"整数値_人前", 2, "人前", "2人前"},
		{"整数値_個", 3, "個", "3個"},
		{"小数値_g", 1.5, "g", "1.5g"},
		{"小数値_杯", 0.5, "杯", "0.5杯"},
		{"整数相当の浮動小数点", 1.0, "皿", "1皿"},
		{"整数相当の浮動小数点_大きい値", 200.0, "ml", "200ml"},
		{"ゼロ", 0, "g", "0g"},
		{"NaN", math.NaN(), "g", "0g"},
		{"正の無限大", math.Inf(1), "ml", "0ml"},
		{"負の無限大", math.Inf(-1), "杯", "0杯"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatQuantity(tt.value, tt.unit)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClassifierResponseItem_toFoodItem(t *testing.T) {
	item := classifierResponseItem{
		Name:          "味噌ラーメン",
		QuantityValue: 1,
		QuantityUnit:  "杯",
	}

	food := item.toFoodItem()

	assert.Equal(t, "味噌ラーメン", food.Name)
	assert.Equal(t, "1杯", food.EstimatedAmount)
}

func TestNutritionResponseItem_toNutritionInfo(t *testing.T) {
	item := nutritionResponseItem{
		Name:          "カレーライス",
		QuantityValue: 1,
		QuantityUnit:  "皿",
		Calories:      600,
		Protein:       15.5,
		Fat:           20.3,
		Carbohydrates: 85.2,
		Iron:          3.5,
		Calcium:       120.0,
		VitaminC:      15.0,
	}

	info := item.toNutritionInfo()

	assert.Equal(t, "カレーライス", info.Name)
	assert.Equal(t, "1皿", info.EstimatedAmount)
	assert.Equal(t, 600.0, info.Calories)
	assert.Equal(t, 15.5, info.Protein)
	assert.Equal(t, 20.3, info.Fat)
	assert.Equal(t, 85.2, info.Carbohydrates)
	assert.Equal(t, 3.5, info.Micronutrients["iron_mg"])
	assert.Equal(t, 120.0, info.Micronutrients["calcium_mg"])
	assert.Equal(t, 15.0, info.Micronutrients["vitamin_c_mg"])
	assert.Len(t, info.Micronutrients, len(AllMicronutrients))
}

func TestFoodItemSchema(t *testing.T) {
	schema := FoodItemSchema()

	assert.Equal(t, SchemaTypeArray, schema.Type)
	assert.NotNil(t, schema.Items)
	assert.Equal(t, SchemaTypeObject, schema.Items.Type)
	assert.Contains(t, schema.Items.Properties, "name")
	assert.Contains(t, schema.Items.Properties, "quantity_value")
	assert.Contains(t, schema.Items.Properties, "quantity_unit")
	assert.Equal(t, SchemaTypeString, schema.Items.Properties["quantity_unit"].Type)
	assert.NotEmpty(t, schema.Items.Properties["quantity_unit"].Enum)
	assert.Contains(t, schema.Items.Properties["quantity_unit"].Enum, "g")
	assert.Contains(t, schema.Items.Properties["quantity_unit"].Enum, "杯")
}

func TestNutritionInfoSchema(t *testing.T) {
	schema := NutritionInfoSchema()

	assert.Equal(t, SchemaTypeArray, schema.Type)
	assert.NotNil(t, schema.Items)

	// 基本プロパティ
	baseProps := []string{"name", "quantity_value", "quantity_unit", "calories_kcal", "protein_g", "fat_g", "carbohydrates_g"}
	for _, prop := range baseProps {
		assert.Contains(t, schema.Items.Properties, prop)
	}

	// マイクロニュートリエント（レジストリから検証）
	for _, key := range AllMicronutrientKeys() {
		assert.Contains(t, schema.Items.Properties, key)
		assert.Equal(t, SchemaTypeNumber, schema.Items.Properties[key].Type)
	}

	assert.Len(t, schema.Items.Required, len(baseProps)+len(AllMicronutrients))
}
