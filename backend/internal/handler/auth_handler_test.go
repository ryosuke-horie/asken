package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserRepository struct {
	findByEmailFn        func(ctx context.Context, email string) (*repository.User, error)
	createWithPasswordFn func(ctx context.Context, email, name, passwordHash string) (*repository.User, error)
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*repository.User, error) {
	if m.findByEmailFn != nil {
		return m.findByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *mockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*repository.User, error) {
	return nil, nil
}

func (m *mockUserRepository) CreateWithPassword(ctx context.Context, email, name, passwordHash string) (*repository.User, error) {
	if m.createWithPasswordFn != nil {
		return m.createWithPasswordFn(ctx, email, name, passwordHash)
	}
	return nil, nil
}

func TestAuthHandler_Register(t *testing.T) {
	authService := service.NewAuthService("test-secret", 24*time.Hour)

	t.Run("新規ユーザーを登録すべき", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			findByEmailFn: func(ctx context.Context, email string) (*repository.User, error) {
				return nil, nil
			},
			createWithPasswordFn: func(ctx context.Context, email, name, passwordHash string) (*repository.User, error) {
				return &repository.User{
					ID:        uuid.New(),
					Email:     email,
					Name:      name,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil
			},
		}
		handler := NewAuthHandler(authService, mockRepo)

		reqBody := `{"email": "test@example.com", "password": "password123", "name": "Test User"}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleRegister(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.NotEmpty(t, response["token"])
		assert.NotNil(t, response["user"])
	})

	t.Run("既存メールアドレスで登録を拒否すべき", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			findByEmailFn: func(ctx context.Context, email string) (*repository.User, error) {
				return &repository.User{
					ID:    uuid.New(),
					Email: email,
				}, nil
			},
		}
		handler := NewAuthHandler(authService, mockRepo)

		reqBody := `{"email": "existing@example.com", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleRegister(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("短いパスワードを拒否すべき", func(t *testing.T) {
		mockRepo := &mockUserRepository{}
		handler := NewAuthHandler(authService, mockRepo)

		reqBody := `{"email": "test@example.com", "password": "short"}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleRegister(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("無効なメールアドレスを拒否すべき", func(t *testing.T) {
		mockRepo := &mockUserRepository{}
		handler := NewAuthHandler(authService, mockRepo)

		reqBody := `{"email": "invalid-email", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleRegister(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthHandler_Register_DBErrors(t *testing.T) {
	authService := service.NewAuthService("test-secret", 24*time.Hour)

	t.Run("FindByEmailでDBエラー時に500を返すべき", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			findByEmailFn: func(ctx context.Context, email string) (*repository.User, error) {
				return nil, assert.AnError
			},
		}
		handler := NewAuthHandler(authService, mockRepo)

		reqBody := `{"email": "test@example.com", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleRegister(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("CreateWithPasswordでDBエラー時に500を返すべき", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			findByEmailFn: func(ctx context.Context, email string) (*repository.User, error) {
				return nil, nil
			},
			createWithPasswordFn: func(ctx context.Context, email, name, passwordHash string) (*repository.User, error) {
				return nil, assert.AnError
			},
		}
		handler := NewAuthHandler(authService, mockRepo)

		reqBody := `{"email": "test@example.com", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleRegister(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("不正なJSONで400を返すべき", func(t *testing.T) {
		mockRepo := &mockUserRepository{}
		handler := NewAuthHandler(authService, mockRepo)

		reqBody := `{invalid json}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleRegister(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("空のメールアドレスで400を返すべき", func(t *testing.T) {
		mockRepo := &mockUserRepository{}
		handler := NewAuthHandler(authService, mockRepo)

		reqBody := `{"email": "", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleRegister(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("空のパスワードで400を返すべき", func(t *testing.T) {
		mockRepo := &mockUserRepository{}
		handler := NewAuthHandler(authService, mockRepo)

		reqBody := `{"email": "test@example.com", "password": ""}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleRegister(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	authService := service.NewAuthService("test-secret", 24*time.Hour)

	t.Run("正しい認証情報でログインすべき", func(t *testing.T) {
		passwordHash, _ := authService.HashPassword("password123")
		mockRepo := &mockUserRepository{
			findByEmailFn: func(ctx context.Context, email string) (*repository.User, error) {
				return &repository.User{
					ID:           uuid.New(),
					Email:        email,
					Name:         "Test User",
					PasswordHash: passwordHash,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}, nil
			},
		}
		handler := NewAuthHandler(authService, mockRepo)

		reqBody := `{"email": "test@example.com", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleLogin(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.NotEmpty(t, response["token"])
		assert.NotNil(t, response["user"])
	})

	t.Run("存在しないユーザーでログインを拒否すべき", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			findByEmailFn: func(ctx context.Context, email string) (*repository.User, error) {
				return nil, nil
			},
		}
		handler := NewAuthHandler(authService, mockRepo)

		reqBody := `{"email": "notexist@example.com", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleLogin(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("誤ったパスワードでログインを拒否すべき", func(t *testing.T) {
		passwordHash, _ := authService.HashPassword("correctPassword")
		mockRepo := &mockUserRepository{
			findByEmailFn: func(ctx context.Context, email string) (*repository.User, error) {
				return &repository.User{
					ID:           uuid.New(),
					Email:        email,
					PasswordHash: passwordHash,
				}, nil
			},
		}
		handler := NewAuthHandler(authService, mockRepo)

		reqBody := `{"email": "test@example.com", "password": "wrongPassword"}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleLogin(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("無効なメールアドレス形式でログインを拒否すべき", func(t *testing.T) {
		mockRepo := &mockUserRepository{}
		handler := NewAuthHandler(authService, mockRepo)

		reqBody := `{"email": "invalid-email", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleLogin(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthHandler_Login_DBErrors(t *testing.T) {
	authService := service.NewAuthService("test-secret", 24*time.Hour)

	t.Run("FindByEmailでDBエラー時に500を返すべき", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			findByEmailFn: func(ctx context.Context, email string) (*repository.User, error) {
				return nil, assert.AnError
			},
		}
		handler := NewAuthHandler(authService, mockRepo)

		reqBody := `{"email": "test@example.com", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleLogin(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("不正なJSONで400を返すべき", func(t *testing.T) {
		mockRepo := &mockUserRepository{}
		handler := NewAuthHandler(authService, mockRepo)

		reqBody := `{invalid json}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleLogin(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("空のメールアドレスで400を返すべき", func(t *testing.T) {
		mockRepo := &mockUserRepository{}
		handler := NewAuthHandler(authService, mockRepo)

		reqBody := `{"email": "", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleLogin(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("空のパスワードで400を返すべき", func(t *testing.T) {
		mockRepo := &mockUserRepository{}
		handler := NewAuthHandler(authService, mockRepo)

		reqBody := `{"email": "test@example.com", "password": ""}`
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandleLogin(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
