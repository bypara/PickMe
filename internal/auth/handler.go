package auth

import (
	"log/slog"
	"net/http"

	"github.com/bypara/PickMe/internal/httpapi"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Handler связывает маршруты аккаунта с PostgreSQL.
type Handler struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

type account struct {
	ID       int64  `json:"id,string"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

// New создаёт обработчик аккаунтов с общим пулом соединений приложения.
func New(db *pgxpool.Pool, logger *slog.Logger) *Handler {
	return &Handler{db: db, logger: logger}
}

// RegisterRoutes добавляет маршруты аккаунтов к общему маршрутизатору API.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/register", h.register)
}

func (h *Handler) internalError(w http.ResponseWriter, r *http.Request, err error) {
	h.logger.Error("account request failed", "path", r.URL.Path, "error", err)
	httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "Не удалось выполнить запрос. Попробуйте позже.")
}
