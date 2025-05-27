package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userClaimsKey = contextKey("userClaims")

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Authorization header format must be Bearer {token}", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Парсим публичный ключ из PEM
		pubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1fRWRUud7ol3gi3n+0/H
aFz5CThaKPS6DSKRAFqZ5+JLu0zxESB7ENNSucfAlEUQcO5HEyqPXZ/AN+0xf/Vg
QmaSpkfXugJ4dJquglbce4K2gWQP3WW4swcY3AtCvWmSaSeKsg+3eirqjbx741H2
lhoV+AM9OeNIYre0oeqbBxGpCd9BBESqFZMeQn8uOQeUIiiaLPvb4hzI7UDP3hRo
mZfJIZpz21kbIspwgexbly818cNbt/KQ2ChqD4m0jmfTUvx77ufwubMGM2mIhyWE
Bsc5aY6N81fUaERFgarbxnVrzx7ccgwwWlY65w8m4ZerrbJv4bzveoE12rCKFEi/
fwIDAQAB
-----END PUBLIC KEY-----`))
		if err != nil {
			http.Error(w, "Failed to parse public key", http.StatusInternalServerError)
			return
		}

		// Парсим токен с проверкой подписи
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return pubKey, nil
		})

		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Проверка срока действия
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				http.Error(w, "Token expired", http.StatusUnauthorized)
				return
			}
		}

		ctx := context.WithValue(r.Context(), userClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func FromContext(ctx context.Context) (jwt.MapClaims, bool) {
	claims, ok := ctx.Value(userClaimsKey).(jwt.MapClaims)
	return claims, ok
}
