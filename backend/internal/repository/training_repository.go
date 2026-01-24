package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TrainingLocation はトレーニング場所を表す構造体
type TrainingLocation struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TrainingEquipment は器具を表す構造体
type TrainingEquipment struct {
	ID           uuid.UUID `json:"id"`
	LocationID   uuid.UUID `json:"location_id"`
	Name         string    `json:"name"`
	OriginalName *string   `json:"original_name,omitempty"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TrainingRecord は練習記録を表す構造体
type TrainingRecord struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	LocationID   *uuid.UUID `json:"location_id,omitempty"`
	LocationName *string    `json:"location_name,omitempty"`
	RecordedAt   time.Time  `json:"recorded_at"`
	Completed    bool       `json:"completed"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// TrainingRepository はトレーニング関連のデータアクセスインターフェース
type TrainingRepository interface {
	// Location関連
	GetAllLocations(ctx context.Context, userID uuid.UUID) ([]*TrainingLocation, error)
	GetLocationByID(ctx context.Context, id, userID uuid.UUID) (*TrainingLocation, error)
	CreateLocation(ctx context.Context, location *TrainingLocation) (*TrainingLocation, error)
	UpdateLocation(ctx context.Context, location *TrainingLocation) (*TrainingLocation, error)
	DeleteLocation(ctx context.Context, id, userID uuid.UUID) error

	// Equipment関連
	GetEquipmentByLocation(ctx context.Context, locationID uuid.UUID) ([]*TrainingEquipment, error)
	GetEquipmentByID(ctx context.Context, id uuid.UUID) (*TrainingEquipment, error)
	CreateEquipment(ctx context.Context, equipment *TrainingEquipment) (*TrainingEquipment, error)
	UpdateEquipment(ctx context.Context, equipment *TrainingEquipment) (*TrainingEquipment, error)
	DeleteEquipment(ctx context.Context, id uuid.UUID) error

	// Record関連
	GetRecords(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*TrainingRecord, error)
	GetRecordByDate(ctx context.Context, userID uuid.UUID, date time.Time) (*TrainingRecord, error)
	UpsertRecord(ctx context.Context, record *TrainingRecord) (*TrainingRecord, error)
}

type postgresTrainingRepository struct {
	db *sql.DB
}

// NewTrainingRepository は新しいTrainingRepositoryを作成
func NewTrainingRepository(db *sql.DB) TrainingRepository {
	return &postgresTrainingRepository{db: db}
}

// Location関連の実装

func (r *postgresTrainingRepository) GetAllLocations(ctx context.Context, userID uuid.UUID) ([]*TrainingLocation, error) {
	query := `
		SELECT id, user_id, name, sort_order, created_at, updated_at
		FROM training_locations
		WHERE user_id = $1
		ORDER BY sort_order ASC, created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("トレーニング場所の取得に失敗: %w", err)
	}
	defer rows.Close()

	var locations []*TrainingLocation
	for rows.Next() {
		var loc TrainingLocation
		if err := rows.Scan(
			&loc.ID,
			&loc.UserID,
			&loc.Name,
			&loc.SortOrder,
			&loc.CreatedAt,
			&loc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("トレーニング場所のスキャンに失敗: %w", err)
		}
		locations = append(locations, &loc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("トレーニング場所の読み取りに失敗: %w", err)
	}

	return locations, nil
}

func (r *postgresTrainingRepository) GetLocationByID(ctx context.Context, id, userID uuid.UUID) (*TrainingLocation, error) {
	query := `
		SELECT id, user_id, name, sort_order, created_at, updated_at
		FROM training_locations
		WHERE id = $1 AND user_id = $2
	`

	var loc TrainingLocation
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&loc.ID,
		&loc.UserID,
		&loc.Name,
		&loc.SortOrder,
		&loc.CreatedAt,
		&loc.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("トレーニング場所の取得に失敗: %w", err)
	}

	return &loc, nil
}

func (r *postgresTrainingRepository) CreateLocation(ctx context.Context, location *TrainingLocation) (*TrainingLocation, error) {
	maxSortOrderQuery := `SELECT COALESCE(MAX(sort_order), -1) + 1 FROM training_locations WHERE user_id = $1`
	var nextSortOrder int
	if err := r.db.QueryRowContext(ctx, maxSortOrderQuery, location.UserID).Scan(&nextSortOrder); err != nil {
		return nil, fmt.Errorf("sort_orderの取得に失敗: %w", err)
	}

	query := `
		INSERT INTO training_locations (user_id, name, sort_order)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, name, sort_order, created_at, updated_at
	`

	var created TrainingLocation
	err := r.db.QueryRowContext(ctx, query,
		location.UserID,
		location.Name,
		nextSortOrder,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.Name,
		&created.SortOrder,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("トレーニング場所の作成に失敗: %w", err)
	}

	return &created, nil
}

func (r *postgresTrainingRepository) UpdateLocation(ctx context.Context, location *TrainingLocation) (*TrainingLocation, error) {
	query := `
		UPDATE training_locations
		SET name = $1
		WHERE id = $2 AND user_id = $3
		RETURNING id, user_id, name, sort_order, created_at, updated_at
	`

	var updated TrainingLocation
	err := r.db.QueryRowContext(ctx, query,
		location.Name,
		location.ID,
		location.UserID,
	).Scan(
		&updated.ID,
		&updated.UserID,
		&updated.Name,
		&updated.SortOrder,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("トレーニング場所が見つかりません: %s", location.ID)
	}
	if err != nil {
		return nil, fmt.Errorf("トレーニング場所の更新に失敗: %w", err)
	}

	return &updated, nil
}

func (r *postgresTrainingRepository) DeleteLocation(ctx context.Context, id, userID uuid.UUID) error {
	query := `DELETE FROM training_locations WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("トレーニング場所の削除に失敗: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("影響を受けた行数の取得に失敗: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("トレーニング場所が見つかりません: %s", id)
	}

	return nil
}

// Equipment関連の実装

func (r *postgresTrainingRepository) GetEquipmentByLocation(ctx context.Context, locationID uuid.UUID) ([]*TrainingEquipment, error) {
	query := `
		SELECT id, location_id, name, original_name, sort_order, created_at, updated_at
		FROM training_equipment
		WHERE location_id = $1
		ORDER BY sort_order ASC, created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, locationID)
	if err != nil {
		return nil, fmt.Errorf("器具の取得に失敗: %w", err)
	}
	defer rows.Close()

	var equipment []*TrainingEquipment
	for rows.Next() {
		var eq TrainingEquipment
		var originalName sql.NullString
		if err := rows.Scan(
			&eq.ID,
			&eq.LocationID,
			&eq.Name,
			&originalName,
			&eq.SortOrder,
			&eq.CreatedAt,
			&eq.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("器具のスキャンに失敗: %w", err)
		}
		if originalName.Valid {
			eq.OriginalName = &originalName.String
		}
		equipment = append(equipment, &eq)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("器具の読み取りに失敗: %w", err)
	}

	return equipment, nil
}

func (r *postgresTrainingRepository) GetEquipmentByID(ctx context.Context, id uuid.UUID) (*TrainingEquipment, error) {
	query := `
		SELECT id, location_id, name, original_name, sort_order, created_at, updated_at
		FROM training_equipment
		WHERE id = $1
	`

	var eq TrainingEquipment
	var originalName sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&eq.ID,
		&eq.LocationID,
		&eq.Name,
		&originalName,
		&eq.SortOrder,
		&eq.CreatedAt,
		&eq.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("器具の取得に失敗: %w", err)
	}

	if originalName.Valid {
		eq.OriginalName = &originalName.String
	}

	return &eq, nil
}

func (r *postgresTrainingRepository) CreateEquipment(ctx context.Context, equipment *TrainingEquipment) (*TrainingEquipment, error) {
	maxSortOrderQuery := `SELECT COALESCE(MAX(sort_order), -1) + 1 FROM training_equipment WHERE location_id = $1`
	var nextSortOrder int
	if err := r.db.QueryRowContext(ctx, maxSortOrderQuery, equipment.LocationID).Scan(&nextSortOrder); err != nil {
		return nil, fmt.Errorf("sort_orderの取得に失敗: %w", err)
	}

	query := `
		INSERT INTO training_equipment (location_id, name, original_name, sort_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id, location_id, name, original_name, sort_order, created_at, updated_at
	`

	var created TrainingEquipment
	var originalName sql.NullString

	var originalNameParam interface{}
	if equipment.OriginalName != nil {
		originalNameParam = *equipment.OriginalName
	} else {
		originalNameParam = nil
	}

	err := r.db.QueryRowContext(ctx, query,
		equipment.LocationID,
		equipment.Name,
		originalNameParam,
		nextSortOrder,
	).Scan(
		&created.ID,
		&created.LocationID,
		&created.Name,
		&originalName,
		&created.SortOrder,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("器具の作成に失敗: %w", err)
	}

	if originalName.Valid {
		created.OriginalName = &originalName.String
	}

	return &created, nil
}

func (r *postgresTrainingRepository) UpdateEquipment(ctx context.Context, equipment *TrainingEquipment) (*TrainingEquipment, error) {
	query := `
		UPDATE training_equipment
		SET name = $1, original_name = $2
		WHERE id = $3
		RETURNING id, location_id, name, original_name, sort_order, created_at, updated_at
	`

	var updated TrainingEquipment
	var originalName sql.NullString

	var originalNameParam interface{}
	if equipment.OriginalName != nil {
		originalNameParam = *equipment.OriginalName
	} else {
		originalNameParam = nil
	}

	err := r.db.QueryRowContext(ctx, query,
		equipment.Name,
		originalNameParam,
		equipment.ID,
	).Scan(
		&updated.ID,
		&updated.LocationID,
		&updated.Name,
		&originalName,
		&updated.SortOrder,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("器具が見つかりません: %s", equipment.ID)
	}
	if err != nil {
		return nil, fmt.Errorf("器具の更新に失敗: %w", err)
	}

	if originalName.Valid {
		updated.OriginalName = &originalName.String
	}

	return &updated, nil
}

func (r *postgresTrainingRepository) DeleteEquipment(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM training_equipment WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("器具の削除に失敗: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("影響を受けた行数の取得に失敗: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("器具が見つかりません: %s", id)
	}

	return nil
}

// Record関連の実装

func (r *postgresTrainingRepository) GetRecords(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*TrainingRecord, error) {
	query := `
		SELECT tr.id, tr.user_id, tr.location_id, tl.name, tr.recorded_at, tr.completed, tr.created_at, tr.updated_at
		FROM training_records tr
		LEFT JOIN training_locations tl ON tr.location_id = tl.id
		WHERE tr.user_id = $1 AND tr.recorded_at >= $2 AND tr.recorded_at <= $3
		ORDER BY tr.recorded_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("練習記録の取得に失敗: %w", err)
	}
	defer rows.Close()

	var records []*TrainingRecord
	for rows.Next() {
		var rec TrainingRecord
		var locationID sql.NullString
		var locationName sql.NullString
		if err := rows.Scan(
			&rec.ID,
			&rec.UserID,
			&locationID,
			&locationName,
			&rec.RecordedAt,
			&rec.Completed,
			&rec.CreatedAt,
			&rec.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("練習記録のスキャンに失敗: %w", err)
		}
		if locationID.Valid {
			id, _ := uuid.Parse(locationID.String)
			rec.LocationID = &id
		}
		if locationName.Valid {
			rec.LocationName = &locationName.String
		}
		records = append(records, &rec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("練習記録の読み取りに失敗: %w", err)
	}

	return records, nil
}

func (r *postgresTrainingRepository) GetRecordByDate(ctx context.Context, userID uuid.UUID, date time.Time) (*TrainingRecord, error) {
	query := `
		SELECT tr.id, tr.user_id, tr.location_id, tl.name, tr.recorded_at, tr.completed, tr.created_at, tr.updated_at
		FROM training_records tr
		LEFT JOIN training_locations tl ON tr.location_id = tl.id
		WHERE tr.user_id = $1 AND tr.recorded_at = $2
	`

	var rec TrainingRecord
	var locationID sql.NullString
	var locationName sql.NullString
	err := r.db.QueryRowContext(ctx, query, userID, date).Scan(
		&rec.ID,
		&rec.UserID,
		&locationID,
		&locationName,
		&rec.RecordedAt,
		&rec.Completed,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("練習記録の取得に失敗: %w", err)
	}

	if locationID.Valid {
		id, _ := uuid.Parse(locationID.String)
		rec.LocationID = &id
	}
	if locationName.Valid {
		rec.LocationName = &locationName.String
	}

	return &rec, nil
}

func (r *postgresTrainingRepository) UpsertRecord(ctx context.Context, record *TrainingRecord) (*TrainingRecord, error) {
	query := `
		INSERT INTO training_records (user_id, location_id, recorded_at, completed)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, recorded_at)
		DO UPDATE SET location_id = EXCLUDED.location_id, completed = EXCLUDED.completed
		RETURNING id, user_id, location_id, recorded_at, completed, created_at, updated_at
	`

	var upserted TrainingRecord
	var locationID sql.NullString

	var locationIDParam interface{}
	if record.LocationID != nil {
		locationIDParam = *record.LocationID
	} else {
		locationIDParam = nil
	}

	err := r.db.QueryRowContext(ctx, query,
		record.UserID,
		locationIDParam,
		record.RecordedAt,
		record.Completed,
	).Scan(
		&upserted.ID,
		&upserted.UserID,
		&locationID,
		&upserted.RecordedAt,
		&upserted.Completed,
		&upserted.CreatedAt,
		&upserted.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("練習記録の作成/更新に失敗: %w", err)
	}

	if locationID.Valid {
		id, _ := uuid.Parse(locationID.String)
		upserted.LocationID = &id
	}

	// 場所名を取得
	if upserted.LocationID != nil {
		nameQuery := `SELECT name FROM training_locations WHERE id = $1`
		var name string
		if err := r.db.QueryRowContext(ctx, nameQuery, *upserted.LocationID).Scan(&name); err == nil {
			upserted.LocationName = &name
		}
	}

	return &upserted, nil
}
