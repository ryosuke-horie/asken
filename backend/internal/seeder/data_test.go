package seeder

import (
	"testing"
	"time"
)

func TestGeneratePastDates(t *testing.T) {
	t.Run("過去N日間の日付を生成すべき", func(t *testing.T) {
		days := 7
		dates := GeneratePastDates(days)

		if len(dates) != days {
			t.Errorf("期待: %d件, 実際: %d件", days, len(dates))
		}

		// 最初の日付は今日であるべき
		today := time.Now()
		if dates[0].Day() != today.Day() {
			t.Errorf("最初の日付は今日であるべき")
		}
	})

	t.Run("0日の場合は空のスライスを返すべき", func(t *testing.T) {
		dates := GeneratePastDates(0)
		if len(dates) != 0 {
			t.Errorf("期待: 0件, 実際: %d件", len(dates))
		}
	})
}

func TestFormatDateForDB(t *testing.T) {
	t.Run("日付をYYYY-MM-DD形式でフォーマットすべき", func(t *testing.T) {
		date := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
		result := FormatDateForDB(date)

		expected := "2024-01-15"
		if result != expected {
			t.Errorf("期待: %s, 実際: %s", expected, result)
		}
	})
}

func TestCalculateTotals(t *testing.T) {
	t.Run("栄養素の合計を計算すべき", func(t *testing.T) {
		foods := SampleNutritionData["breakfast"]
		calories, protein, fat, carbs := CalculateTotals(foods)

		// 合計値が0より大きいことを確認
		if calories <= 0 {
			t.Errorf("カロリーの合計が0以下: %f", calories)
		}
		if protein <= 0 {
			t.Errorf("タンパク質の合計が0以下: %f", protein)
		}
		if fat < 0 {
			t.Errorf("脂質の合計が負: %f", fat)
		}
		if carbs <= 0 {
			t.Errorf("炭水化物の合計が0以下: %f", carbs)
		}
	})

	t.Run("空のスライスの場合は全て0を返すべき", func(t *testing.T) {
		calories, protein, fat, carbs := CalculateTotals(nil)

		if calories != 0 || protein != 0 || fat != 0 || carbs != 0 {
			t.Errorf("空のスライスの合計は全て0であるべき")
		}
	})
}

func TestDefaultTestUsers(t *testing.T) {
	t.Run("デフォルトのテストユーザーが存在すべき", func(t *testing.T) {
		if len(DefaultTestUsers) == 0 {
			t.Error("デフォルトのテストユーザーが空")
		}

		for i, user := range DefaultTestUsers {
			if user.Email == "" {
				t.Errorf("ユーザー%dのメールアドレスが空", i)
			}
			if user.Password == "" {
				t.Errorf("ユーザー%dのパスワードが空", i)
			}
			if user.Name == "" {
				t.Errorf("ユーザー%dの名前が空", i)
			}
		}
	})
}

func TestSampleNutritionData(t *testing.T) {
	t.Run("全ての食事タイプにサンプルデータが存在すべき", func(t *testing.T) {
		for _, mealType := range MealTypes {
			foods, ok := SampleNutritionData[mealType]
			if !ok {
				t.Errorf("食事タイプ '%s' のサンプルデータが存在しない", mealType)
				continue
			}
			if len(foods) == 0 {
				t.Errorf("食事タイプ '%s' のサンプルデータが空", mealType)
			}
		}
	})
}

func TestGetSampleNutritionInfo(t *testing.T) {
	t.Run("存在する食事タイプのデータを返すべき", func(t *testing.T) {
		foods := GetSampleNutritionInfo("breakfast")
		if len(foods) == 0 {
			t.Error("breakfastのデータが空")
		}
	})

	t.Run("存在しない食事タイプはlunchのデータを返すべき", func(t *testing.T) {
		foods := GetSampleNutritionInfo("invalid")
		expected := SampleNutritionData["lunch"]
		if len(foods) != len(expected) {
			t.Errorf("期待: %d件, 実際: %d件", len(expected), len(foods))
		}
	})
}
