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

	rows := sqlmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
		AddRow(expectedID, email, nil, now, now)

	mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE email`).
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

	mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE email`).
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

	mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE email`).
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

	rows := sqlmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
		AddRow(userID, email, "Test User", now, now)

	mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id`).
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

	mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.FindByID(ctx, userID)

	assert.NoError(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &User{
		Email: "new@example.com",
		Name:  "New User",
	}

	expectedID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow(expectedID, now, now)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Email, sql.NullString{String: user.Name, Valid: true}).
		WillReturnRows(rows)

	err := repo.Create(ctx, user)

	assert.NoError(t, err)
	assert.Equal(t, expectedID, user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &User{
		Email: "duplicate@example.com",
	}

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Email, sql.NullString{}).
		WillReturnError(sql.ErrConnDone)

	err := repo.Create(ctx, user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ユーザーの作成に失敗")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindOrCreate_NewUser(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	email := "new@example.com"
	expectedID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
		AddRow(expectedID, email, nil, now, now)

	mock.ExpectQuery(`INSERT INTO users .* ON CONFLICT`).
		WithArgs(email, sql.NullString{}).
		WillReturnRows(rows)

	user, err := repo.FindOrCreate(ctx, email, "")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedID, user.ID)
	assert.Equal(t, email, user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindOrCreate_ExistingUser(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	email := "existing@example.com"
	expectedID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
		AddRow(expectedID, email, "Existing User", now, now)

	mock.ExpectQuery(`INSERT INTO users .* ON CONFLICT`).
		WithArgs(email, sql.NullString{}).
		WillReturnRows(rows)

	user, err := repo.FindOrCreate(ctx, email, "")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedID, user.ID)
	assert.Equal(t, "Existing User", user.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindOrCreate_Error(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	email := "error@example.com"

	mock.ExpectQuery(`INSERT INTO users .* ON CONFLICT`).
		WithArgs(email, sql.NullString{}).
		WillReturnError(sql.ErrConnDone)

	user, err := repo.FindOrCreate(ctx, email, "")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "ユーザーの取得または作成に失敗")
	assert.NoError(t, mock.ExpectationsWereMet())
}
