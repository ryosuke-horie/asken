package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name,omitempty"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	CreateWithPassword(ctx context.Context, email, name, passwordHash string) (*User, error)
}

type postgresUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

// scanUser は*sql.Rowからユーザーデータをスキャンするヘルパー関数
// ErrNoRowsの場合は(nil, nil)を返す
func scanUser(row *sql.Row) (*User, error) {
	var user User
	var name sql.NullString

	err := row.Scan(
		&user.ID,
		&user.Email,
		&name,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if name.Valid {
		user.Name = name.String
	}

	return &user, nil
}

func (r *postgresUserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user, err := scanUser(r.db.QueryRowContext(ctx, query, email))
	if err != nil {
		return nil, fmt.Errorf("ユーザーの取得に失敗: %w", err)
	}

	return user, nil
}

func (r *postgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user, err := scanUser(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		return nil, fmt.Errorf("ユーザーの取得に失敗: %w", err)
	}

	return user, nil
}

func (r *postgresUserRepository) CreateWithPassword(ctx context.Context, email, name, passwordHash string) (*User, error) {
	query := `
		INSERT INTO users (email, name, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, name, password_hash, created_at, updated_at
	`

	var nameParam sql.NullString
	if name != "" {
		nameParam = sql.NullString{String: name, Valid: true}
	}

	var user User
	var returnedName sql.NullString

	err := r.db.QueryRowContext(ctx, query, email, nameParam, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&returnedName,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("ユーザーの作成に失敗: %w", err)
	}

	if returnedName.Valid {
		user.Name = returnedName.String
	}

	return &user, nil
}
