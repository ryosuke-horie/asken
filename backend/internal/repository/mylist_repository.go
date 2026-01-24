package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
)

type MylistItem struct {
	ID            uuid.UUID              `json:"id"`
	UserID        uuid.UUID              `json:"user_id"`
	Name          string                 `json:"name"`
	BaseAmount    string                 `json:"base_amount"`
	Unit          string                 `json:"unit"`
	Calories      float64                `json:"calories"`
	Protein       float64                `json:"protein"`
	Fat           float64                `json:"fat"`
	Carbohydrates float64                `json:"carbohydrates"`
	Foods         []gemini.NutritionInfo `json:"foods"`
	SortOrder     int                    `json:"sort_order"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type MylistRepository interface {
	GetAll(ctx context.Context, userID uuid.UUID) ([]*MylistItem, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*MylistItem, error)
	Create(ctx context.Context, item *MylistItem) (*MylistItem, error)
	Update(ctx context.Context, item *MylistItem) (*MylistItem, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	Reorder(ctx context.Context, userID uuid.UUID, itemIDs []uuid.UUID) error
}

type postgresMylistRepository struct {
	db *sql.DB
}

func NewMylistRepository(db *sql.DB) MylistRepository {
	return &postgresMylistRepository{db: db}
}

func (r *postgresMylistRepository) GetAll(ctx context.Context, userID uuid.UUID) ([]*MylistItem, error) {
	query := `
		SELECT id, user_id, name, base_amount, unit, calories, protein, fat, carbohydrates, foods, sort_order, created_at, updated_at
		FROM mylist_items
		WHERE user_id = $1
		ORDER BY sort_order ASC, created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("マイリストの取得に失敗: %w", err)
	}
	defer rows.Close()

	var items []*MylistItem
	for rows.Next() {
		item, err := scanMylistItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("マイリストの読み取りに失敗: %w", err)
	}

	return items, nil
}

func (r *postgresMylistRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*MylistItem, error) {
	query := `
		SELECT id, user_id, name, base_amount, unit, calories, protein, fat, carbohydrates, foods, sort_order, created_at, updated_at
		FROM mylist_items
		WHERE id = $1 AND user_id = $2
	`

	var item MylistItem
	var foodsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.Name,
		&item.BaseAmount,
		&item.Unit,
		&item.Calories,
		&item.Protein,
		&item.Fat,
		&item.Carbohydrates,
		&foodsJSON,
		&item.SortOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("マイリストアイテムの取得に失敗: %w", err)
	}

	if err := json.Unmarshal(foodsJSON, &item.Foods); err != nil {
		return nil, fmt.Errorf("foods のデシリアライズに失敗: %w", err)
	}

	return &item, nil
}

func (r *postgresMylistRepository) Create(ctx context.Context, item *MylistItem) (*MylistItem, error) {
	foodsJSON, err := json.Marshal(item.Foods)
	if err != nil {
		return nil, fmt.Errorf("foods のJSON変換に失敗: %w", err)
	}

	maxSortOrderQuery := `SELECT COALESCE(MAX(sort_order), -1) + 1 FROM mylist_items WHERE user_id = $1`
	var nextSortOrder int
	if err := r.db.QueryRowContext(ctx, maxSortOrderQuery, item.UserID).Scan(&nextSortOrder); err != nil {
		return nil, fmt.Errorf("sort_orderの取得に失敗: %w", err)
	}

	query := `
		INSERT INTO mylist_items (user_id, name, base_amount, unit, calories, protein, fat, carbohydrates, foods, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, user_id, name, base_amount, unit, calories, protein, fat, carbohydrates, foods, sort_order, created_at, updated_at
	`

	var created MylistItem
	var returnedFoodsJSON []byte

	err = r.db.QueryRowContext(ctx, query,
		item.UserID,
		item.Name,
		item.BaseAmount,
		item.Unit,
		item.Calories,
		item.Protein,
		item.Fat,
		item.Carbohydrates,
		foodsJSON,
		nextSortOrder,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.Name,
		&created.BaseAmount,
		&created.Unit,
		&created.Calories,
		&created.Protein,
		&created.Fat,
		&created.Carbohydrates,
		&returnedFoodsJSON,
		&created.SortOrder,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("マイリストアイテムの作成に失敗: %w", err)
	}

	if err := json.Unmarshal(returnedFoodsJSON, &created.Foods); err != nil {
		return nil, fmt.Errorf("foods のデシリアライズに失敗: %w", err)
	}

	return &created, nil
}

func (r *postgresMylistRepository) Update(ctx context.Context, item *MylistItem) (*MylistItem, error) {
	foodsJSON, err := json.Marshal(item.Foods)
	if err != nil {
		return nil, fmt.Errorf("foods のJSON変換に失敗: %w", err)
	}

	query := `
		UPDATE mylist_items
		SET name = $1, base_amount = $2, unit = $3, calories = $4, protein = $5, fat = $6, carbohydrates = $7, foods = $8
		WHERE id = $9 AND user_id = $10
		RETURNING id, user_id, name, base_amount, unit, calories, protein, fat, carbohydrates, foods, sort_order, created_at, updated_at
	`

	var updated MylistItem
	var returnedFoodsJSON []byte

	err = r.db.QueryRowContext(ctx, query,
		item.Name,
		item.BaseAmount,
		item.Unit,
		item.Calories,
		item.Protein,
		item.Fat,
		item.Carbohydrates,
		foodsJSON,
		item.ID,
		item.UserID,
	).Scan(
		&updated.ID,
		&updated.UserID,
		&updated.Name,
		&updated.BaseAmount,
		&updated.Unit,
		&updated.Calories,
		&updated.Protein,
		&updated.Fat,
		&updated.Carbohydrates,
		&returnedFoodsJSON,
		&updated.SortOrder,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("マイリストアイテムが見つかりません: %s", item.ID)
	}
	if err != nil {
		return nil, fmt.Errorf("マイリストアイテムの更新に失敗: %w", err)
	}

	if err := json.Unmarshal(returnedFoodsJSON, &updated.Foods); err != nil {
		return nil, fmt.Errorf("foods のデシリアライズに失敗: %w", err)
	}

	return &updated, nil
}

func (r *postgresMylistRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	query := `DELETE FROM mylist_items WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("マイリストアイテムの削除に失敗: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("影響を受けた行数の取得に失敗: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("マイリストアイテムが見つかりません: %s", id)
	}

	return nil
}

func (r *postgresMylistRepository) Reorder(ctx context.Context, userID uuid.UUID, itemIDs []uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("トランザクションの開始に失敗: %w", err)
	}
	defer tx.Rollback()

	query := `UPDATE mylist_items SET sort_order = $1 WHERE id = $2 AND user_id = $3`

	for i, id := range itemIDs {
		result, err := tx.ExecContext(ctx, query, i, id, userID)
		if err != nil {
			return fmt.Errorf("sort_orderの更新に失敗: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("影響を受けた行数の取得に失敗: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("マイリストアイテムが見つかりません: %s", id)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("トランザクションのコミットに失敗: %w", err)
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanMylistItem(rows rowScanner) (*MylistItem, error) {
	var item MylistItem
	var foodsJSON []byte

	err := rows.Scan(
		&item.ID,
		&item.UserID,
		&item.Name,
		&item.BaseAmount,
		&item.Unit,
		&item.Calories,
		&item.Protein,
		&item.Fat,
		&item.Carbohydrates,
		&foodsJSON,
		&item.SortOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("マイリストアイテムのスキャンに失敗: %w", err)
	}

	if err := json.Unmarshal(foodsJSON, &item.Foods); err != nil {
		return nil, fmt.Errorf("foods のデシリアライズに失敗: %w", err)
	}

	return &item, nil
}
