package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
)

const minPasswordLength = 8

type AuthHandler struct {
	authService    *service.AuthService
	userRepository repository.UserRepository
}

func NewAuthHandler(authService *service.AuthService, userRepository repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		authService:    authService,
		userRepository: userRepository,
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	log.Printf("Register request from %s", r.RemoteAddr)

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding register request: %v", err)
		http.Error(w, "リクエストの解析に失敗しました", http.StatusBadRequest)
		return
	}

	if !emailRegex.MatchString(req.Email) {
		http.Error(w, "無効なメールアドレス形式です", http.StatusBadRequest)
		return
	}

	if len(req.Password) < minPasswordLength {
		http.Error(w, "パスワードは8文字以上で入力してください", http.StatusBadRequest)
		return
	}

	existingUser, err := h.userRepository.FindByEmail(r.Context(), req.Email)
	if err != nil {
		log.Printf("Error checking existing user: %v", err)
		http.Error(w, "サーバーエラーが発生しました", http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		http.Error(w, "このメールアドレスは既に登録されています", http.StatusConflict)
		return
	}

	passwordHash, err := h.authService.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "サーバーエラーが発生しました", http.StatusInternalServerError)
		return
	}

	user, err := h.userRepository.CreateWithPassword(r.Context(), req.Email, req.Name, passwordHash)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "ユーザーの作成に失敗しました", http.StatusInternalServerError)
		return
	}

	token, err := h.authService.GenerateToken(user.ID)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		http.Error(w, "サーバーエラーが発生しました", http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Token: token,
		User: UserResponse{
			ID:    user.ID.String(),
			Email: user.Email,
			Name:  user.Name,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	log.Printf("Login request from %s", r.RemoteAddr)

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding login request: %v", err)
		http.Error(w, "リクエストの解析に失敗しました", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "メールアドレスとパスワードを入力してください", http.StatusBadRequest)
		return
	}

	// email形式の検証
	if !emailRegex.MatchString(req.Email) {
		http.Error(w, "無効なメールアドレス形式です", http.StatusBadRequest)
		return
	}

	user, err := h.userRepository.FindByEmail(r.Context(), req.Email)
	if err != nil {
		log.Printf("Error finding user: %v", err)
		http.Error(w, "サーバーエラーが発生しました", http.StatusInternalServerError)
		return
	}

	// タイミング攻撃対策としてユーザーが存在しない場合もパスワード検証を実行する。
	// ダミーハッシュを使用して一定時間を消費させる。
	passwordHash := ""
	if user != nil {
		passwordHash = user.PasswordHash
	} else {
		// ダミーハッシュ（bcryptの有効なハッシュ形式）
		passwordHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	}

	passwordValid := h.authService.VerifyPassword(passwordHash, req.Password)

	// ユーザーが存在しない、またはパスワードが不正の場合は認証失敗
	if user == nil || !passwordValid {
		http.Error(w, "メールアドレスまたはパスワードが正しくありません", http.StatusUnauthorized)
		return
	}

	token, err := h.authService.GenerateToken(user.ID)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		http.Error(w, "サーバーエラーが発生しました", http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Token: token,
		User: UserResponse{
			ID:    user.ID.String(),
			Email: user.Email,
			Name:  user.Name,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
