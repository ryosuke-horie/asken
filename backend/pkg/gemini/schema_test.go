package gemini

import (
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
	}

	info := item.toNutritionInfo()

	assert.Equal(t, "カレーライス", info.Name)
	assert.Equal(t, "1皿", info.EstimatedAmount)
	assert.Equal(t, 600.0, info.Calories)
	assert.Equal(t, 15.5, info.Protein)
	assert.Equal(t, 20.3, info.Fat)
	assert.Equal(t, 85.2, info.Carbohydrates)
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
	assert.Contains(t, schema.Items.Properties, "name")
	assert.Contains(t, schema.Items.Properties, "quantity_value")
	assert.Contains(t, schema.Items.Properties, "quantity_unit")
	assert.Contains(t, schema.Items.Properties, "calories_kcal")
	assert.Contains(t, schema.Items.Properties, "protein_g")
	assert.Contains(t, schema.Items.Properties, "fat_g")
	assert.Contains(t, schema.Items.Properties, "carbohydrates_g")
	assert.Len(t, schema.Items.Required, 7)
}
