package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// ErrTrainingNotFound はリソースが見つからない場合のエラー
var ErrTrainingNotFound = errors.New("training resource not found")

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

// TrainingMenu は練習メニューを表す構造体
type TrainingMenu struct {
	ID        uuid.UUID  `json:"id"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	Name      string     `json:"name"`
	IsDefault bool       `json:"is_default"`
	SortOrder int        `json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// TrainingRecord は練習記録を表す構造体
type TrainingRecord struct {
	ID           uuid.UUID       `json:"id"`
	UserID       uuid.UUID       `json:"user_id"`
	LocationID   *uuid.UUID      `json:"location_id,omitempty"`
	LocationName *string         `json:"location_name,omitempty"`
	RecordedAt   time.Time       `json:"recorded_at"`
	Completed    bool            `json:"completed"`
	Duration     *int            `json:"duration,omitempty"`
	Intensity    *int            `json:"intensity,omitempty"`
	Satisfaction *int            `json:"satisfaction,omitempty"`
	Notes        *string         `json:"notes,omitempty"`
	Menus        []*TrainingMenu `json:"menus,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
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

	// Menu関連
	GetMenus(ctx context.Context, userID uuid.UUID) ([]*TrainingMenu, error)
	CreateMenu(ctx context.Context, menu *TrainingMenu) (*TrainingMenu, error)
	DeleteMenu(ctx context.Context, id, userID uuid.UUID) error

	// Record関連
	GetRecords(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*TrainingRecord, error)
	GetRecordByDate(ctx context.Context, userID uuid.UUID, date time.Time) (*TrainingRecord, error)
	GetRecordByID(ctx context.Context, id, userID uuid.UUID) (*TrainingRecord, error)
	CreateRecord(ctx context.Context, record *TrainingRecord, menuIDs []uuid.UUID) (*TrainingRecord, error)
	UpdateRecord(ctx context.Context, record *TrainingRecord, menuIDs []uuid.UUID) (*TrainingRecord, error)
	DeleteRecord(ctx context.Context, id, userID uuid.UUID) error
	// 後方互換性のため残す
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
		return nil, fmt.Errorf("%w: location %s", ErrTrainingNotFound, location.ID)
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
		return fmt.Errorf("%w: location %s", ErrTrainingNotFound, id)
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
		return nil, fmt.Errorf("%w: equipment %s", ErrTrainingNotFound, equipment.ID)
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
		return fmt.Errorf("%w: equipment %s", ErrTrainingNotFound, id)
	}

	return nil
}

// Menu関連の実装

func (r *postgresTrainingRepository) GetMenus(ctx context.Context, userID uuid.UUID) ([]*TrainingMenu, error) {
	query := `
		SELECT id, user_id, name, is_default, sort_order, created_at, updated_at
		FROM training_menus
		WHERE user_id IS NULL OR user_id = $1
		ORDER BY is_default DESC, sort_order ASC, created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("メニューの取得に失敗: %w", err)
	}
	defer rows.Close()

	var menus []*TrainingMenu
	for rows.Next() {
		var menu TrainingMenu
		var userIDNull sql.NullString
		if err := rows.Scan(
			&menu.ID,
			&userIDNull,
			&menu.Name,
			&menu.IsDefault,
			&menu.SortOrder,
			&menu.CreatedAt,
			&menu.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("メニューのスキャンに失敗: %w", err)
		}
		if userIDNull.Valid {
			id, err := uuid.Parse(userIDNull.String)
			if err == nil {
				menu.UserID = &id
			}
		}
		menus = append(menus, &menu)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("メニューの読み取りに失敗: %w", err)
	}

	return menus, nil
}

func (r *postgresTrainingRepository) CreateMenu(ctx context.Context, menu *TrainingMenu) (*TrainingMenu, error) {
	maxSortOrderQuery := `SELECT COALESCE(MAX(sort_order), 99) + 1 FROM training_menus WHERE user_id = $1`
	var nextSortOrder int
	if err := r.db.QueryRowContext(ctx, maxSortOrderQuery, menu.UserID).Scan(&nextSortOrder); err != nil {
		return nil, fmt.Errorf("sort_orderの取得に失敗: %w", err)
	}

	query := `
		INSERT INTO training_menus (user_id, name, is_default, sort_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, name, is_default, sort_order, created_at, updated_at
	`

	var created TrainingMenu
	var userIDNull sql.NullString

	err := r.db.QueryRowContext(ctx, query,
		menu.UserID,
		menu.Name,
		false,
		nextSortOrder,
	).Scan(
		&created.ID,
		&userIDNull,
		&created.Name,
		&created.IsDefault,
		&created.SortOrder,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("メニューの作成に失敗: %w", err)
	}

	if userIDNull.Valid {
		id, err := uuid.Parse(userIDNull.String)
		if err == nil {
			created.UserID = &id
		}
	}

	return &created, nil
}

func (r *postgresTrainingRepository) DeleteMenu(ctx context.Context, id, userID uuid.UUID) error {
	// 固定メニュー（user_id IS NULL）は削除不可
	query := `DELETE FROM training_menus WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("メニューの削除に失敗: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("影響を受けた行数の取得に失敗: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: menu %s", ErrTrainingNotFound, id)
	}

	return nil
}

// Record関連の実装

func (r *postgresTrainingRepository) GetRecords(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*TrainingRecord, error) {
	query := `
		SELECT tr.id, tr.user_id, tr.location_id, tl.name, tr.recorded_at, tr.completed,
		       tr.duration, tr.intensity, tr.satisfaction, tr.notes,
		       tr.created_at, tr.updated_at
		FROM training_records tr
		LEFT JOIN training_locations tl ON tr.location_id = tl.id
		WHERE tr.user_id = $1 AND tr.recorded_at >= $2 AND tr.recorded_at <= $3
		ORDER BY tr.recorded_at DESC, tr.created_at DESC
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
		var duration sql.NullInt64
		var intensity sql.NullInt64
		var satisfaction sql.NullInt64
		var notes sql.NullString

		if err := rows.Scan(
			&rec.ID,
			&rec.UserID,
			&locationID,
			&locationName,
			&rec.RecordedAt,
			&rec.Completed,
			&duration,
			&intensity,
			&satisfaction,
			&notes,
			&rec.CreatedAt,
			&rec.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("練習記録のスキャンに失敗: %w", err)
		}
		if locationID.Valid {
			id, err := uuid.Parse(locationID.String)
			if err != nil {
				log.Printf("警告: location_idのパースに失敗 (value=%s): %v", locationID.String, err)
			} else {
				rec.LocationID = &id
			}
		}
		if locationName.Valid {
			rec.LocationName = &locationName.String
		}
		if duration.Valid {
			d := int(duration.Int64)
			rec.Duration = &d
		}
		if intensity.Valid {
			i := int(intensity.Int64)
			rec.Intensity = &i
		}
		if satisfaction.Valid {
			s := int(satisfaction.Int64)
			rec.Satisfaction = &s
		}
		if notes.Valid {
			rec.Notes = &notes.String
		}

		// メニュー情報を取得
		menus, err := r.getMenusByRecordID(ctx, rec.ID)
		if err != nil {
			log.Printf("警告: メニューの取得に失敗 (record_id=%s): %v", rec.ID, err)
		} else {
			rec.Menus = menus
		}

		records = append(records, &rec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("練習記録の読み取りに失敗: %w", err)
	}

	return records, nil
}

func (r *postgresTrainingRepository) getMenusByRecordID(ctx context.Context, recordID uuid.UUID) ([]*TrainingMenu, error) {
	query := `
		SELECT tm.id, tm.user_id, tm.name, tm.is_default, tm.sort_order, tm.created_at, tm.updated_at
		FROM training_menus tm
		INNER JOIN training_record_menus trm ON tm.id = trm.menu_id
		WHERE trm.record_id = $1
		ORDER BY tm.sort_order ASC
	`

	rows, err := r.db.QueryContext(ctx, query, recordID)
	if err != nil {
		return nil, fmt.Errorf("メニューの取得に失敗: %w", err)
	}
	defer rows.Close()

	var menus []*TrainingMenu
	for rows.Next() {
		var menu TrainingMenu
		var userIDNull sql.NullString
		if err := rows.Scan(
			&menu.ID,
			&userIDNull,
			&menu.Name,
			&menu.IsDefault,
			&menu.SortOrder,
			&menu.CreatedAt,
			&menu.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("メニューのスキャンに失敗: %w", err)
		}
		if userIDNull.Valid {
			id, err := uuid.Parse(userIDNull.String)
			if err == nil {
				menu.UserID = &id
			}
		}
		menus = append(menus, &menu)
	}

	return menus, nil
}

func (r *postgresTrainingRepository) GetRecordByDate(ctx context.Context, userID uuid.UUID, date time.Time) (*TrainingRecord, error) {
	query := `
		SELECT tr.id, tr.user_id, tr.location_id, tl.name, tr.recorded_at, tr.completed,
		       tr.duration, tr.intensity, tr.satisfaction, tr.notes,
		       tr.created_at, tr.updated_at
		FROM training_records tr
		LEFT JOIN training_locations tl ON tr.location_id = tl.id
		WHERE tr.user_id = $1 AND tr.recorded_at = $2
		LIMIT 1
	`

	var rec TrainingRecord
	var locationID sql.NullString
	var locationName sql.NullString
	var duration sql.NullInt64
	var intensity sql.NullInt64
	var satisfaction sql.NullInt64
	var notes sql.NullString

	err := r.db.QueryRowContext(ctx, query, userID, date).Scan(
		&rec.ID,
		&rec.UserID,
		&locationID,
		&locationName,
		&rec.RecordedAt,
		&rec.Completed,
		&duration,
		&intensity,
		&satisfaction,
		&notes,
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
		id, err := uuid.Parse(locationID.String)
		if err != nil {
			log.Printf("警告: location_idのパースに失敗 (value=%s): %v", locationID.String, err)
		} else {
			rec.LocationID = &id
		}
	}
	if locationName.Valid {
		rec.LocationName = &locationName.String
	}
	if duration.Valid {
		d := int(duration.Int64)
		rec.Duration = &d
	}
	if intensity.Valid {
		i := int(intensity.Int64)
		rec.Intensity = &i
	}
	if satisfaction.Valid {
		s := int(satisfaction.Int64)
		rec.Satisfaction = &s
	}
	if notes.Valid {
		rec.Notes = &notes.String
	}

	// メニュー情報を取得
	menus, err := r.getMenusByRecordID(ctx, rec.ID)
	if err != nil {
		log.Printf("警告: メニューの取得に失敗 (record_id=%s): %v", rec.ID, err)
	} else {
		rec.Menus = menus
	}

	return &rec, nil
}

func (r *postgresTrainingRepository) GetRecordByID(ctx context.Context, id, userID uuid.UUID) (*TrainingRecord, error) {
	query := `
		SELECT tr.id, tr.user_id, tr.location_id, tl.name, tr.recorded_at, tr.completed,
		       tr.duration, tr.intensity, tr.satisfaction, tr.notes,
		       tr.created_at, tr.updated_at
		FROM training_records tr
		LEFT JOIN training_locations tl ON tr.location_id = tl.id
		WHERE tr.id = $1 AND tr.user_id = $2
	`

	var rec TrainingRecord
	var locationID sql.NullString
	var locationName sql.NullString
	var duration sql.NullInt64
	var intensity sql.NullInt64
	var satisfaction sql.NullInt64
	var notes sql.NullString

	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&rec.ID,
		&rec.UserID,
		&locationID,
		&locationName,
		&rec.RecordedAt,
		&rec.Completed,
		&duration,
		&intensity,
		&satisfaction,
		&notes,
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
		id, err := uuid.Parse(locationID.String)
		if err != nil {
			log.Printf("警告: location_idのパースに失敗 (value=%s): %v", locationID.String, err)
		} else {
			rec.LocationID = &id
		}
	}
	if locationName.Valid {
		rec.LocationName = &locationName.String
	}
	if duration.Valid {
		d := int(duration.Int64)
		rec.Duration = &d
	}
	if intensity.Valid {
		i := int(intensity.Int64)
		rec.Intensity = &i
	}
	if satisfaction.Valid {
		s := int(satisfaction.Int64)
		rec.Satisfaction = &s
	}
	if notes.Valid {
		rec.Notes = &notes.String
	}

	// メニュー情報を取得
	menus, err := r.getMenusByRecordID(ctx, rec.ID)
	if err != nil {
		log.Printf("警告: メニューの取得に失敗 (record_id=%s): %v", rec.ID, err)
	} else {
		rec.Menus = menus
	}

	return &rec, nil
}

func (r *postgresTrainingRepository) CreateRecord(ctx context.Context, record *TrainingRecord, menuIDs []uuid.UUID) (*TrainingRecord, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		INSERT INTO training_records (user_id, location_id, recorded_at, completed, duration, intensity, satisfaction, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, location_id, recorded_at, completed, duration, intensity, satisfaction, notes, created_at, updated_at
	`

	var created TrainingRecord
	var locationID sql.NullString
	var duration sql.NullInt64
	var intensity sql.NullInt64
	var satisfaction sql.NullInt64
	var notes sql.NullString

	var locationIDParam interface{}
	if record.LocationID != nil {
		locationIDParam = *record.LocationID
	}

	err = tx.QueryRowContext(ctx, query,
		record.UserID,
		locationIDParam,
		record.RecordedAt,
		record.Completed,
		record.Duration,
		record.Intensity,
		record.Satisfaction,
		record.Notes,
	).Scan(
		&created.ID,
		&created.UserID,
		&locationID,
		&created.RecordedAt,
		&created.Completed,
		&duration,
		&intensity,
		&satisfaction,
		&notes,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("練習記録の作成に失敗: %w", err)
	}

	if locationID.Valid {
		id, err := uuid.Parse(locationID.String)
		if err == nil {
			created.LocationID = &id
		}
	}
	if duration.Valid {
		d := int(duration.Int64)
		created.Duration = &d
	}
	if intensity.Valid {
		i := int(intensity.Int64)
		created.Intensity = &i
	}
	if satisfaction.Valid {
		s := int(satisfaction.Int64)
		created.Satisfaction = &s
	}
	if notes.Valid {
		created.Notes = &notes.String
	}

	// メニュー紐付け
	if len(menuIDs) > 0 {
		menuQuery := `INSERT INTO training_record_menus (record_id, menu_id) VALUES ($1, $2)`
		for _, menuID := range menuIDs {
			if _, err := tx.ExecContext(ctx, menuQuery, created.ID, menuID); err != nil {
				return nil, fmt.Errorf("メニュー紐付けに失敗: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションコミットに失敗: %w", err)
	}

	// 場所名を取得
	if created.LocationID != nil {
		nameQuery := `SELECT name FROM training_locations WHERE id = $1`
		var name string
		if err := r.db.QueryRowContext(ctx, nameQuery, *created.LocationID).Scan(&name); err != nil {
			if err != sql.ErrNoRows {
				log.Printf("警告: 場所名の取得に失敗 (location_id=%s): %v", *created.LocationID, err)
			}
		} else {
			created.LocationName = &name
		}
	}

	// メニュー情報を取得
	menus, err := r.getMenusByRecordID(ctx, created.ID)
	if err != nil {
		log.Printf("警告: メニューの取得に失敗 (record_id=%s): %v", created.ID, err)
	} else {
		created.Menus = menus
	}

	return &created, nil
}

func (r *postgresTrainingRepository) UpdateRecord(ctx context.Context, record *TrainingRecord, menuIDs []uuid.UUID) (*TrainingRecord, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		UPDATE training_records
		SET location_id = $1, recorded_at = $2, completed = $3, duration = $4, intensity = $5, satisfaction = $6, notes = $7
		WHERE id = $8 AND user_id = $9
		RETURNING id, user_id, location_id, recorded_at, completed, duration, intensity, satisfaction, notes, created_at, updated_at
	`

	var updated TrainingRecord
	var locationID sql.NullString
	var duration sql.NullInt64
	var intensity sql.NullInt64
	var satisfaction sql.NullInt64
	var notes sql.NullString

	var locationIDParam interface{}
	if record.LocationID != nil {
		locationIDParam = *record.LocationID
	}

	err = tx.QueryRowContext(ctx, query,
		locationIDParam,
		record.RecordedAt,
		record.Completed,
		record.Duration,
		record.Intensity,
		record.Satisfaction,
		record.Notes,
		record.ID,
		record.UserID,
	).Scan(
		&updated.ID,
		&updated.UserID,
		&locationID,
		&updated.RecordedAt,
		&updated.Completed,
		&duration,
		&intensity,
		&satisfaction,
		&notes,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w: record %s", ErrTrainingNotFound, record.ID)
	}
	if err != nil {
		return nil, fmt.Errorf("練習記録の更新に失敗: %w", err)
	}

	if locationID.Valid {
		id, err := uuid.Parse(locationID.String)
		if err == nil {
			updated.LocationID = &id
		}
	}
	if duration.Valid {
		d := int(duration.Int64)
		updated.Duration = &d
	}
	if intensity.Valid {
		i := int(intensity.Int64)
		updated.Intensity = &i
	}
	if satisfaction.Valid {
		s := int(satisfaction.Int64)
		updated.Satisfaction = &s
	}
	if notes.Valid {
		updated.Notes = &notes.String
	}

	// 既存のメニュー紐付けを削除
	deleteMenusQuery := `DELETE FROM training_record_menus WHERE record_id = $1`
	if _, err := tx.ExecContext(ctx, deleteMenusQuery, record.ID); err != nil {
		return nil, fmt.Errorf("メニュー紐付けの削除に失敗: %w", err)
	}

	// 新しいメニュー紐付け
	if len(menuIDs) > 0 {
		menuQuery := `INSERT INTO training_record_menus (record_id, menu_id) VALUES ($1, $2)`
		for _, menuID := range menuIDs {
			if _, err := tx.ExecContext(ctx, menuQuery, updated.ID, menuID); err != nil {
				return nil, fmt.Errorf("メニュー紐付けに失敗: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションコミットに失敗: %w", err)
	}

	// 場所名を取得
	if updated.LocationID != nil {
		nameQuery := `SELECT name FROM training_locations WHERE id = $1`
		var name string
		if err := r.db.QueryRowContext(ctx, nameQuery, *updated.LocationID).Scan(&name); err != nil {
			if err != sql.ErrNoRows {
				log.Printf("警告: 場所名の取得に失敗 (location_id=%s): %v", *updated.LocationID, err)
			}
		} else {
			updated.LocationName = &name
		}
	}

	// メニュー情報を取得
	menus, err := r.getMenusByRecordID(ctx, updated.ID)
	if err != nil {
		log.Printf("警告: メニューの取得に失敗 (record_id=%s): %v", updated.ID, err)
	} else {
		updated.Menus = menus
	}

	return &updated, nil
}

func (r *postgresTrainingRepository) DeleteRecord(ctx context.Context, id, userID uuid.UUID) error {
	query := `DELETE FROM training_records WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("練習記録の削除に失敗: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("影響を受けた行数の取得に失敗: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: record %s", ErrTrainingNotFound, id)
	}

	return nil
}

// UpsertRecord は後方互換性のために残す（1日複数記録対応後は非推奨）
func (r *postgresTrainingRepository) UpsertRecord(ctx context.Context, record *TrainingRecord) (*TrainingRecord, error) {
	// 既存のレコードを検索
	existing, err := r.GetRecordByDate(ctx, record.UserID, record.RecordedAt)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		// 既存レコードを更新
		record.ID = existing.ID
		return r.UpdateRecord(ctx, record, nil)
	}

	// 新規作成
	return r.CreateRecord(ctx, record, nil)
}
