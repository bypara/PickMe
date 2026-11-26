package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireUser(t *testing.T) {
	const userID = 42
	protected := RequireUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, ok := CurrentUser(r); !ok || id != userID {
			t.Errorf("CurrentUser = (%d, %t), want (%d, true)", id, ok, userID)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/animals", nil)
	guest := httptest.NewRecorder()
	protected.ServeHTTP(guest, req)
	if guest.Code != http.StatusUnauthorized {
		t.Errorf("guest status = %d, want 401", guest.Code)
	}

	authorized := httptest.NewRecorder()
	ctx := context.WithValue(req.Context(), accountContextKey{}, &account{ID: userID})
	protected.ServeHTTP(authorized, req.WithContext(ctx))
	if authorized.Code != http.StatusNoContent {
		t.Errorf("authorized status = %d, want 204", authorized.Code)
	}
}
