package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const CookieName = "user_id"

type ctxKey struct{}

func UserIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

func withUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func sign(payload string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return payload + "." + sig
}

func verify(value string, secret []byte) (string, bool) {
	i := strings.LastIndex(value, ".")
	if i < 0 {
		return "", false
	}
	payload, sig := value[:i], value[i+1:]

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return "", false
	}
	return payload, true
}

func Auth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c, err := r.Cookie(CookieName); err == nil {
				if userID, ok := verify(c.Value, secret); ok {
					next.ServeHTTP(w, r.WithContext(withUserID(r.Context(), userID)))
					return
				}
			}

			userID := uuid.NewString()
			http.SetCookie(w, &http.Cookie{
				Name:     CookieName,
				Value:    sign(userID, secret),
				Path:     "/",
				HttpOnly: true,
			})
			next.ServeHTTP(w, r.WithContext(withUserID(r.Context(), userID)))
		})
	}
}
