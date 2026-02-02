package repository

import (
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

func TestRecordScanResult_ApplyToRecord_AllFieldsValid(t *testing.T) {
	validUUID := uuid.New()

	scanResult := recordScanResult{
		LocationID:   sql.NullString{String: validUUID.String(), Valid: true},
		Duration:     sql.NullInt64{Int64: 60, Valid: true},
		Intensity:    sql.NullInt64{Int64: 3, Valid: true},
		Satisfaction: sql.NullInt64{Int64: 4, Valid: true},
		Notes:        sql.NullString{String: "テストメモ", Valid: true},
	}

	record := &TrainingRecord{
		ID:     uuid.New(),
		UserID: "test-firebase-uid",
	}

	scanResult.applyToRecord(record)

	if record.LocationID == nil {
		t.Error("LocationIDが設定されていません")
	}
	if record.Duration == nil || *record.Duration != 60 {
		t.Error("Durationが正しく設定されていません")
	}
	if record.Intensity == nil || *record.Intensity != 3 {
		t.Error("Intensityが正しく設定されていません")
	}
	if record.Satisfaction == nil || *record.Satisfaction != 4 {
		t.Error("Satisfactionが正しく設定されていません")
	}
	if record.Notes == nil || *record.Notes != "テストメモ" {
		t.Error("Notesが正しく設定されていません")
	}
}

func TestRecordScanResult_ApplyToRecord_AllFieldsNull(t *testing.T) {
	scanResult := recordScanResult{
		LocationID:   sql.NullString{Valid: false},
		Duration:     sql.NullInt64{Valid: false},
		Intensity:    sql.NullInt64{Valid: false},
		Satisfaction: sql.NullInt64{Valid: false},
		Notes:        sql.NullString{Valid: false},
	}

	record := &TrainingRecord{
		ID:     uuid.New(),
		UserID: "test-firebase-uid",
	}

	scanResult.applyToRecord(record)

	if record.LocationID != nil {
		t.Error("LocationIDがnilであるべきです")
	}
	if record.Duration != nil {
		t.Error("Durationがnilであるべきです")
	}
	if record.Intensity != nil {
		t.Error("Intensityがnilであるべきです")
	}
	if record.Satisfaction != nil {
		t.Error("Satisfactionがnilであるべきです")
	}
	if record.Notes != nil {
		t.Error("Notesがnilであるべきです")
	}
}

func TestRecordScanResult_ApplyToRecord_InvalidUUID(t *testing.T) {
	// 無効なUUID形式の場合、LocationIDはnilになりログが出力される（エラーにはならない）
	scanResult := recordScanResult{
		LocationID:   sql.NullString{String: "invalid-uuid", Valid: true},
		Duration:     sql.NullInt64{Int64: 60, Valid: true},
		Intensity:    sql.NullInt64{Valid: false},
		Satisfaction: sql.NullInt64{Valid: false},
		Notes:        sql.NullString{Valid: false},
	}

	record := &TrainingRecord{
		ID:     uuid.New(),
		UserID: "test-firebase-uid",
	}

	// パニックしないことを確認
	scanResult.applyToRecord(record)

	// LocationIDはnilであるべき（パースエラーのため）
	if record.LocationID != nil {
		t.Error("無効なUUIDの場合、LocationIDはnilであるべきです")
	}
	// 他のフィールドは正常に設定される
	if record.Duration == nil || *record.Duration != 60 {
		t.Error("Durationが正しく設定されていません")
	}
}

func TestRecordScanResult_ApplyToRecord_PartialFields(t *testing.T) {
	scanResult := recordScanResult{
		LocationID:   sql.NullString{Valid: false},
		Duration:     sql.NullInt64{Int64: 45, Valid: true},
		Intensity:    sql.NullInt64{Int64: 5, Valid: true},
		Satisfaction: sql.NullInt64{Valid: false},
		Notes:        sql.NullString{String: "部分的なメモ", Valid: true},
	}

	record := &TrainingRecord{
		ID:     uuid.New(),
		UserID: "test-firebase-uid",
	}

	scanResult.applyToRecord(record)

	if record.LocationID != nil {
		t.Error("LocationIDがnilであるべきです")
	}
	if record.Duration == nil || *record.Duration != 45 {
		t.Error("Durationが正しく設定されていません")
	}
	if record.Intensity == nil || *record.Intensity != 5 {
		t.Error("Intensityが正しく設定されていません")
	}
	if record.Satisfaction != nil {
		t.Error("Satisfactionがnilであるべきです")
	}
	if record.Notes == nil || *record.Notes != "部分的なメモ" {
		t.Error("Notesが正しく設定されていません")
	}
}
