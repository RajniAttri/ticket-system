package middleware

import (
	"context"
	"net/http"
	"strings"

	"ticket-system/internal/auth"
	"ticket-system/internal/httpx"
)


type contextKey string

const userIDKey contextKey = "userID"

type Authenticator struct {
	jwt *auth.JWTManager
}

// NewAuthenticator is the constructor.
func NewAuthenticator(jwtManager *auth.JWTManager) *Authenticator {
	return &Authenticator{jwt: jwtManager}
}

func (a *Authenticator) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		scheme, token, found := strings.Cut(header, " ")
		if !found || !strings.EqualFold(scheme, "Bearer") || token == "" {
			httpx.Error(w, http.StatusUnauthorized, "missing or malformed Authorization header")
			return
		}

		claims, err := a.jwt.Parse(token)
		if err != nil {
			httpx.Error(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}
