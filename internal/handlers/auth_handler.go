package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"ticket-system/internal/auth"
	"ticket-system/internal/httpx"
	"ticket-system/internal/models"
	"ticket-system/internal/store"
)

type AuthHandler struct {
	store store.UserStore
	jwt   *auth.JWTManager
}

func NewAuthHandler(s store.UserStore, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{store: s, jwt: jwtManager}
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	Token string `json:"token"`
}


func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body credentials
	if err := decodeJSON(r, &body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if msg, ok := validateCredentials(body); !ok {
		httpx.Error(w, http.StatusBadRequest, msg)
		return
	}

	hash, err := auth.HashPassword(body.Password)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not hash password")
		return
	}

	user := &models.User{
		ID:           uuid.NewString(),
		Email:        strings.ToLower(strings.TrimSpace(body.Email)),
		PasswordHash: hash,
		CreatedAt:    time.Now(),
	}

	if err := h.store.CreateUser(user); err != nil {
		if errors.Is(err, store.ErrEmailExists) {
			httpx.Error(w, http.StatusConflict, "email already registered")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "could not create user")
		return
	}
	httpx.JSON(w, http.StatusCreated, user)
}


func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body credentials
	if err := decodeJSON(r, &body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	email := strings.ToLower(strings.TrimSpace(body.Email))
	user, err := h.store.GetUserByEmail(email)
	if err != nil || !auth.CheckPassword(user.PasswordHash, body.Password) {
		httpx.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := h.jwt.Generate(user.ID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "could not generate token")
		return
	}

	httpx.JSON(w, http.StatusOK, tokenResponse{Token: token})
}

// --- small shared helpers used by handlers ---


func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

// validateCredentials enforces basic input rules and returns a message + ok.
func validateCredentials(c credentials) (string, bool) {
	if strings.TrimSpace(c.Email) == "" || !strings.Contains(c.Email, "@") {
		return "a valid email is required", false
	}
	if len(c.Password) < 6 {
		return "password must be at least 6 characters", false
	}
	return "", true
}
