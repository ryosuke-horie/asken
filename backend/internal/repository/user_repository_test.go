package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var now = time.Now()

func TestFindByEmail_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	expectedID := uuid.New()
	email := "test@example.com"

	rows := sqlmock.NewRows([]string{"id", "email", "name", "password_hash", "created_at", "updated_at"}).
		AddRow(expectedID, email, nil, "hashed_password", now, now)

	mock.ExpectQuery(`SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE email`).
		WithArgs(email).
		WillReturnRows(rows)

	user, err := repo.FindByEmail(ctx, email)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedID, user.ID)
	assert.Equal(t, email, user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByEmail_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	email := "notfound@example.com"

	mock.ExpectQuery(`SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE email`).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.FindByEmail(ctx, email)

	assert.NoError(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByEmail_DBError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	email := "test@example.com"

	mock.ExpectQuery(`SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE email`).
		WithArgs(email).
		WillReturnError(sql.ErrConnDone)

	user, err := repo.FindByEmail(ctx, email)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "ユーザーの取得に失敗")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	email := "test@example.com"

	rows := sqlmock.NewRows([]string{"id", "email", "name", "password_hash", "created_at", "updated_at"}).
		AddRow(userID, email, "Test User", "hashed_password", now, now)

	mock.ExpectQuery(`SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE id`).
		WithArgs(userID).
		WillReturnRows(rows)

	user, err := repo.FindByID(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, "Test User", user.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByID_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	mock.ExpectQuery(`SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE id`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.FindByID(ctx, userID)

	assert.NoError(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateWithPassword_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	email := "new@example.com"
	name := "New User"
	passwordHash := "hashed_password_123"
	expectedID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "email", "name", "password_hash", "created_at", "updated_at"}).
		AddRow(expectedID, email, name, passwordHash, now, now)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(email, sql.NullString{String: name, Valid: true}, passwordHash).
		WillReturnRows(rows)

	user, err := repo.CreateWithPassword(ctx, email, name, passwordHash)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedID, user.ID)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, name, user.Name)
	assert.Equal(t, passwordHash, user.PasswordHash)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateWithPassword_WithoutName(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	email := "new@example.com"
	passwordHash := "hashed_password_123"
	expectedID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "email", "name", "password_hash", "created_at", "updated_at"}).
		AddRow(expectedID, email, nil, passwordHash, now, now)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(email, sql.NullString{}, passwordHash).
		WillReturnRows(rows)

	user, err := repo.CreateWithPassword(ctx, email, "", passwordHash)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedID, user.ID)
	assert.Equal(t, email, user.Email)
	assert.Empty(t, user.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateWithPassword_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	email := "duplicate@example.com"
	passwordHash := "hashed_password_123"

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(email, sql.NullString{}, passwordHash).
		WillReturnError(sql.ErrConnDone)

	user, err := repo.CreateWithPassword(ctx, email, "", passwordHash)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "ユーザーの作成に失敗")
	assert.NoError(t, mock.ExpectationsWereMet())
}
