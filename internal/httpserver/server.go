package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/bypara/PickMe/internal/auth"
	"github.com/bypara/PickMe/internal/httpapi"
	"github.com/jackc/pgx/v5/pgxpool"
)

// New создаёт маршруты API и подключает общую проверку источника небезопасных HTTP-методов.
func New(db *pgxpool.Pool, logger *slog.Logger, cookieSecure bool) http.Handler {
	mux := http.NewServeMux()
	accounts := auth.New(db, logger, cookieSecure)
	accounts.RegisterRoutes(mux)
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		// Проверяем доступность БД с коротким таймаутом, чтобы Compose быстро увидел сбой.
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		w.Header().Set("Cache-Control", "no-store")
		if err := db.Ping(ctx); err != nil {
			logger.Error("database health check failed", "error", err)
			httpapi.WriteError(w, http.StatusServiceUnavailable, "database_unavailable", "Database unavailable")
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	originProtection := http.NewCrossOriginProtection()
	originProtection.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpapi.WriteError(w, http.StatusForbidden, "cross_origin_request", "Cross-origin request rejected")
	}))
	return originProtection.Handler(mux)
}
