package handler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// RecordValidationResult はバリデーション結果を保持
type RecordValidationResult struct {
	RecordedAt time.Time
	LocationID *uuid.UUID
	MenuIDs    []uuid.UUID
}

// validateRecordDate は日付のバリデーションを行う
func validateRecordDate(recordedAtStr string) (time.Time, error) {
	if recordedAtStr == "" {
		return time.Time{}, fmt.Errorf("recorded_atは必須です")
	}
	recordedAt, err := time.Parse("2006-01-02", recordedAtStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("無効な日付形式です（YYYY-MM-DD）")
	}
	return recordedAt, nil
}

// validateRecordFields は練習記録のフィールドをバリデーションする
func validateRecordFields(intensity, satisfaction, duration *int) error {
	if intensity != nil && (*intensity < 1 || *intensity > 5) {
		return fmt.Errorf("強度は1-5の範囲で指定してください")
	}
	if satisfaction != nil && (*satisfaction < 1 || *satisfaction > 5) {
		return fmt.Errorf("満足度は1-5の範囲で指定してください")
	}
	if duration != nil && *duration < 0 {
		return fmt.Errorf("練習時間は0以上の値を指定してください")
	}
	return nil
}

// parseLocationID はlocation_idをパースする
func parseLocationID(locationIDStr *string) (*uuid.UUID, error) {
	if locationIDStr == nil {
		return nil, nil
	}
	locationID, err := uuid.Parse(*locationIDStr)
	if err != nil {
		return nil, fmt.Errorf("無効なlocation_id形式です")
	}
	return &locationID, nil
}

// parseMenuIDs はmenu_idsをパースする
func parseMenuIDs(menuIDStrs []string) ([]uuid.UUID, error) {
	menuIDs := make([]uuid.UUID, 0, len(menuIDStrs))
	for _, menuIDStr := range menuIDStrs {
		menuID, err := uuid.Parse(menuIDStr)
		if err != nil {
			return nil, fmt.Errorf("無効なmenu_id形式です")
		}
		menuIDs = append(menuIDs, menuID)
	}
	return menuIDs, nil
}

// validateAndParseLocationID はlocation_idをパースし、所有権を確認する
func (h *TrainingHandler) validateAndParseLocationID(
	ctx context.Context,
	userID string,
	locationIDStr *string,
) (*uuid.UUID, error) {
	locationID, err := parseLocationID(locationIDStr)
	if err != nil {
		return nil, err
	}
	if locationID == nil {
		return nil, nil
	}

	location, err := h.repository.GetLocationByID(ctx, (*locationID).String(), userID)
	if err != nil {
		log.Printf("場所の取得に失敗 (location_id=%s, user_id=%s): %v", *locationID, userID, err)
		return nil, fmt.Errorf("場所の取得に失敗しました")
	}
	if location == nil {
		return nil, fmt.Errorf("指定された場所が見つかりません")
	}
	return locationID, nil
}

// validateCreateRecordRequest はCreateRecordリクエストをバリデーションする
func (h *TrainingHandler) validateCreateRecordRequest(
	ctx context.Context,
	userID string,
	req *CreateRecordRequest,
) (*RecordValidationResult, error) {
	recordedAt, err := validateRecordDate(req.RecordedAt)
	if err != nil {
		return nil, err
	}

	if err := validateRecordFields(req.Intensity, req.Satisfaction, req.Duration); err != nil {
		return nil, err
	}

	locationID, err := h.validateAndParseLocationID(ctx, userID, req.LocationID)
	if err != nil {
		return nil, err
	}

	menuIDs, err := parseMenuIDs(req.MenuIDs)
	if err != nil {
		return nil, err
	}

	return &RecordValidationResult{
		RecordedAt: recordedAt,
		LocationID: locationID,
		MenuIDs:    menuIDs,
	}, nil
}

// validateUpdateRecordRequest はUpdateRecordリクエストをバリデーションする
func (h *TrainingHandler) validateUpdateRecordRequest(
	ctx context.Context,
	userID string,
	req *UpdateRecordRequest,
) (*RecordValidationResult, error) {
	recordedAt, err := validateRecordDate(req.RecordedAt)
	if err != nil {
		return nil, err
	}

	if err := validateRecordFields(req.Intensity, req.Satisfaction, req.Duration); err != nil {
		return nil, err
	}

	locationID, err := h.validateAndParseLocationID(ctx, userID, req.LocationID)
	if err != nil {
		return nil, err
	}

	menuIDs, err := parseMenuIDs(req.MenuIDs)
	if err != nil {
		return nil, err
	}

	return &RecordValidationResult{
		RecordedAt: recordedAt,
		LocationID: locationID,
		MenuIDs:    menuIDs,
	}, nil
}

// extractAndValidateRecordID はURLパスからIDを抽出してバリデーションする
func extractAndValidateRecordID(urlPath string) (uuid.UUID, error) {
	idStr := extractTrainingIDFromPath(urlPath, "/api/training/records/")
	if idStr == "" {
		return uuid.Nil, fmt.Errorf("IDが指定されていません")
	}

	recordID, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("無効なID形式です")
	}
	return recordID, nil
}
