package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/bypara/PickMe/internal/httpapi"
	"github.com/jackc/pgx/v5"
)

const (
	sessionCookieName = "pickme_session"
	sessionTokenSize  = 32
	sessionLifetime   = 24 * time.Hour
)

type accountContextKey struct{}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	input, ok := readCredentials(w, r)
	if !ok {
		return
	}
	if input.Email == "" || len(input.Email) > 254 || input.Password == "" || utf8.RuneCountInString(input.Password) > 128 {
		invalidCredentials(w)
		return
	}

	var user account
	var encodedHash string
	err := h.db.QueryRow(r.Context(), `
		SELECT id, email, COALESCE(full_name, ''), password_hash FROM users WHERE email = $1`, input.Email,
	).Scan(&user.ID, &user.Email, &user.FullName, &encodedHash)
	unknownEmail := errors.Is(err, pgx.ErrNoRows)
	if unknownEmail {
		// Для неизвестного email выполняем вычисление той же стоимости, что и при проверке пароля.
		encodedHash = dummyPasswordHash
	} else if err != nil {
		h.internalError(w, r, err)
		return
	}

	matches, err := verifyPassword(input.Password, encodedHash)
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	if unknownEmail || !matches {
		invalidCredentials(w)
		return
	}

	var token [sessionTokenSize]byte
	rand.Read(token[:])
	digest := sha256.Sum256(token[:])
	expires := time.Now().UTC().Truncate(time.Second).Add(sessionLifetime)
	if _, err := h.db.Exec(r.Context(), `
		INSERT INTO sessions (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`,
		digest[:], user.ID, expires,
	); err != nil {
		h.internalError(w, r, err)
		return
	}

	cookie := h.sessionCookie()
	cookie.Value = base64.RawURLEncoding.EncodeToString(token[:])
	cookie.MaxAge = int(sessionLifetime / time.Second)
	cookie.Expires = expires
	http.SetCookie(w, cookie)
	httpapi.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if digest, ok := sessionDigest(r); ok {
		if _, err := h.db.Exec(r.Context(), `DELETE FROM sessions WHERE token_hash = $1`, digest[:]); err != nil {
			h.internalError(w, r, err)
			return
		}
	}

	cookie := h.sessionCookie()
	cookie.MaxAge = -1
	cookie.Expires = time.Unix(0, 0).UTC()
	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(accountContextKey{}).(*account)
	httpapi.WriteJSON(w, http.StatusOK, struct {
		User *account `json:"user"`
	}{User: user})
}

// WithSession определяет пользователя по cookie. Гостевой запрос проходит дальше без пользователя.
func (h *Handler) WithSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		digest, hasCookie := sessionDigest(r)
		if !hasCookie {
			next.ServeHTTP(w, r)
			return
		}

		var user account
		err := h.db.QueryRow(r.Context(), `
			SELECT u.id, u.email, COALESCE(u.full_name, '')
			FROM sessions s JOIN users u ON u.id = s.user_id
			WHERE s.token_hash = $1 AND s.expires_at > now()`, digest[:],
		).Scan(&user.ID, &user.Email, &user.FullName)
		if errors.Is(err, pgx.ErrNoRows) {
			next.ServeHTTP(w, r)
			return
		}
		if err != nil {
			h.internalError(w, r, err)
			return
		}

		ctx := context.WithValue(r.Context(), accountContextKey{}, &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireUser пропускает только запросы с действующей сессией. Подключайте его внутри WithSession.
func RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := CurrentUser(r); !ok {
			httpapi.WriteError(w, http.StatusUnauthorized, "authentication_required", "Войдите в аккаунт, чтобы продолжить.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CurrentUser возвращает ID пользователя, найденного WithSession, и признак действующей сессии.
func CurrentUser(r *http.Request) (int64, bool) {
	user, ok := r.Context().Value(accountContextKey{}).(*account)
	if !ok {
		return 0, false
	}
	return user.ID, true
}

func (h *Handler) guestOnly(next http.HandlerFunc) http.Handler {
	return h.WithSession(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := CurrentUser(r); ok {
			httpapi.WriteError(w, http.StatusConflict, "already_authenticated", "Вы уже вошли в аккаунт.")
			return
		}
		next(w, r)
	}))
}

func sessionDigest(r *http.Request) ([sha256.Size]byte, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || len(cookie.Value) != base64.RawURLEncoding.EncodedLen(sessionTokenSize) {
		return [sha256.Size]byte{}, false
	}
	token, err := base64.RawURLEncoding.Strict().DecodeString(cookie.Value)
	if err != nil {
		return [sha256.Size]byte{}, false
	}
	return sha256.Sum256(token), true
}

func (h *Handler) sessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     sessionCookieName,
		Path:     "/api",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
}

func invalidCredentials(w http.ResponseWriter) {
	httpapi.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Неверный email или пароль")
}
