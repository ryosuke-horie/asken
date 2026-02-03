package util

import (
	"fmt"
	"time"
)

// GetDayRangeInTimezone は指定タイムゾーンでの日付範囲をUTCで返す
// dateStr: "2006-01-02"形式の日付文字列
// tz: IANAタイムゾーン名（例: "Asia/Tokyo", "UTC"）
// 戻り値: その日の開始時刻(UTC)と終了時刻(UTC)
func GetDayRangeInTimezone(dateStr, tz string) (start, end time.Time, err error) {
	if tz == "" {
		tz = "UTC"
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("無効なタイムゾーン: %w", err)
	}

	localStart, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("日付のパースに失敗: %w", err)
	}

	localEnd := localStart.Add(24 * time.Hour)

	return localStart.UTC(), localEnd.UTC(), nil
}

// ParseDateInTimezone は日付文字列を指定タイムゾーンの開始時刻としてパースしUTCで返す
func ParseDateInTimezone(dateStr, tz string) (time.Time, error) {
	if tz == "" {
		tz = "UTC"
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.Time{}, fmt.Errorf("無効なタイムゾーン: %w", err)
	}

	t, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("日付のパースに失敗: %w", err)
	}

	return t.UTC(), nil
}
