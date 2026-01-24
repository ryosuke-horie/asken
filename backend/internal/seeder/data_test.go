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

func TestGenerateWeightRecords(t *testing.T) {
	t.Run("指定日数分の体重記録を生成すべき", func(t *testing.T) {
		days := 30
		config := WeightSeedConfig{
			StartWeight:  75.0,
			TargetWeight: 68.0,
		}
		records := GenerateWeightRecords(days, config)

		// 週末のスキップがあるため、日数より少なくなる可能性がある
		if len(records) == 0 {
			t.Error("体重記録が生成されていない")
		}
		if len(records) > days {
			t.Errorf("体重記録が指定日数より多い: 期待最大%d件, 実際%d件", days, len(records))
		}
	})

	t.Run("0日の場合はnilを返すべき", func(t *testing.T) {
		records := GenerateWeightRecords(0, DefaultWeightSeedConfig)
		if records != nil {
			t.Errorf("0日の場合はnilであるべき: 実際%d件", len(records))
		}
	})

	t.Run("負の日数の場合はnilを返すべき", func(t *testing.T) {
		records := GenerateWeightRecords(-1, DefaultWeightSeedConfig)
		if records != nil {
			t.Errorf("負の日数の場合はnilであるべき: 実際%d件", len(records))
		}
	})

	t.Run("体重が減量トレンドを示すべき", func(t *testing.T) {
		days := 90
		config := WeightSeedConfig{
			StartWeight:  75.0,
			TargetWeight: 68.0,
		}
		records := GenerateWeightRecords(days, config)

		if len(records) < 2 {
			t.Skip("記録が2件未満のためスキップ")
		}

		// 最初と最後の記録を比較
		firstWeight := records[0].Weight
		lastWeight := records[len(records)-1].Weight

		// 日々の変動はあるが、全体的に減少しているはず
		// 変動幅を考慮して緩い条件でチェック
		if lastWeight > firstWeight+1.0 {
			t.Errorf("体重が増加している: 開始%.1fkg → 終了%.1fkg", firstWeight, lastWeight)
		}
	})

	t.Run("体重が妥当な範囲であるべき", func(t *testing.T) {
		days := 30
		config := WeightSeedConfig{
			StartWeight:  75.0,
			TargetWeight: 68.0,
		}
		records := GenerateWeightRecords(days, config)

		for _, record := range records {
			if record.Weight < 65.0 || record.Weight > 80.0 {
				t.Errorf("体重が範囲外: %.1fkg", record.Weight)
			}
		}
	})

	t.Run("日付がYYYY-MM-DD形式であるべき", func(t *testing.T) {
		records := GenerateWeightRecords(7, DefaultWeightSeedConfig)
		for _, record := range records {
			if len(record.RecordedAt) != 10 {
				t.Errorf("日付の形式が不正: %s", record.RecordedAt)
			}
		}
	})
}

func TestDefaultWeightSeedConfig(t *testing.T) {
	t.Run("デフォルト設定が妥当な値であるべき", func(t *testing.T) {
		if DefaultWeightSeedConfig.StartWeight <= 0 {
			t.Errorf("開始体重が0以下: %.1f", DefaultWeightSeedConfig.StartWeight)
		}
		if DefaultWeightSeedConfig.TargetWeight <= 0 {
			t.Errorf("目標体重が0以下: %.1f", DefaultWeightSeedConfig.TargetWeight)
		}
		if DefaultWeightSeedConfig.StartWeight <= DefaultWeightSeedConfig.TargetWeight {
			t.Errorf("開始体重が目標体重以下: 開始%.1fkg <= 目標%.1fkg",
				DefaultWeightSeedConfig.StartWeight, DefaultWeightSeedConfig.TargetWeight)
		}
	})
}

func TestGetDefaultTargetDate(t *testing.T) {
	t.Run("1ヶ月後の日付を返すべき", func(t *testing.T) {
		result := GetDefaultTargetDate()
		if len(result) != 10 {
			t.Errorf("日付の形式が不正: %s", result)
		}

		// パースして1ヶ月後であることを確認
		parsed, err := time.Parse("2006-01-02", result)
		if err != nil {
			t.Errorf("日付のパースに失敗: %v", err)
		}

		now := time.Now()
		expected := now.AddDate(0, 1, 0)
		if parsed.Year() != expected.Year() || parsed.Month() != expected.Month() || parsed.Day() != expected.Day() {
			t.Errorf("日付が1ヶ月後でない: 期待%s, 実際%s",
				expected.Format("2006-01-02"), result)
		}
	})
}

func TestCalculateMylistTotals(t *testing.T) {
	t.Run("マイリストアイテムの栄養素合計を計算すべき", func(t *testing.T) {
		foods := DefaultMylistItems[0].Foods // 鶏むね肉定食
		calories, protein, fat, carbs := CalculateMylistTotals(foods)

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
		calories, protein, fat, carbs := CalculateMylistTotals(nil)

		if calories != 0 || protein != 0 || fat != 0 || carbs != 0 {
			t.Errorf("空のスライスの合計は全て0であるべき")
		}
	})
}

func TestDefaultMylistItems(t *testing.T) {
	t.Run("デフォルトのマイリストアイテムが存在すべき", func(t *testing.T) {
		if len(DefaultMylistItems) == 0 {
			t.Error("デフォルトのマイリストアイテムが空")
		}
	})

	t.Run("全てのアイテムが必須フィールドを持つべき", func(t *testing.T) {
		for i, item := range DefaultMylistItems {
			if item.Name == "" {
				t.Errorf("アイテム%dの名前が空", i)
			}
			if item.BaseAmount == "" {
				t.Errorf("アイテム%dのBaseAmountが空", i)
			}
			if item.Unit == "" {
				t.Errorf("アイテム%dのUnitが空", i)
			}
			if len(item.Foods) == 0 {
				t.Errorf("アイテム%dのFoodsが空", i)
			}
			if item.SeedImageSource == "" {
				t.Errorf("アイテム%dのSeedImageSourceが空", i)
			}
		}
	})

	t.Run("各アイテムの栄養素合計が妥当であるべき", func(t *testing.T) {
		for i, item := range DefaultMylistItems {
			calories, protein, _, _ := CalculateMylistTotals(item.Foods)
			if calories <= 0 {
				t.Errorf("アイテム%d (%s) のカロリーが0以下", i, item.Name)
			}
			if protein <= 0 {
				t.Errorf("アイテム%d (%s) のタンパク質が0以下", i, item.Name)
			}
		}
	})
}
