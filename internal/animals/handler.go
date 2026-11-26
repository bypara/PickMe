package animals

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/bypara/PickMe/internal/auth"
	"github.com/bypara/PickMe/internal/httpapi"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Handler обслуживает каталог и карточки животных через общий пул PostgreSQL.
type Handler struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

type animal struct {
	ID            int64     `json:"id,string"`
	Name          string    `json:"name"`
	FoundLocation string    `json:"found_location"`
	Breed         string    `json:"breed"`
	CreatedAt     time.Time `json:"created_at"`
	PhotoURL      *string   `json:"photo_url"`
}

// New создаёт обработчик животных с общим пулом соединений приложения.
func New(db *pgxpool.Pool, logger *slog.Logger) *Handler {
	return &Handler{db: db, logger: logger}
}

// RegisterRoutes открывает каталог для гостей и проверяет сессию перед добавлением.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, accounts *auth.Handler) {
	mux.HandleFunc("GET /api/animals", h.list)
	mux.HandleFunc("GET /api/animals/{id}", h.get)
	mux.HandleFunc("GET /api/animals/{id}/photo", h.getPhoto)
	mux.Handle("POST /api/animals", accounts.WithSession(auth.RequireUser(http.HandlerFunc(h.create))))
}

func (h *Handler) internalError(w http.ResponseWriter, operation string, err error) {
	h.logger.Error("animal request failed", "operation", operation, "error", err)
	httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "Не удалось выполнить запрос. Попробуйте позже.")
}

func readAnimalID(r *http.Request) (int64, bool) {
	value := r.PathValue("id")
	id, err := strconv.ParseInt(value, 10, 64)
	return id, err == nil && id > 0 && value == strconv.FormatInt(id, 10)
}

func photoURL(id int64) *string {
	url := "/api/animals/" + strconv.FormatInt(id, 10) + "/photo"
	return &url
}

func animalNotFound(w http.ResponseWriter) {
	httpapi.WriteError(w, http.StatusNotFound, "animal_not_found", "Животное не найдено.")
}
