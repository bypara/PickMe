package animals

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bypara/PickMe/internal/httpapi"
	"github.com/jackc/pgx/v5"
)

// В списке и карточке нужен только признак наличия фото. Байты читаются отдельным запросом.
const animalColumns = `a.id, a.name, a.found_location, a.breed, a.created_at,
	EXISTS (SELECT 1 FROM animal_photos p WHERE p.animal_id = a.id)`

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(), `SELECT `+animalColumns+`
		FROM animals a ORDER BY a.created_at DESC, a.id DESC`)
	if err != nil {
		h.internalError(w, "list", err)
		return
	}
	defer rows.Close()

	items := make([]animal, 0)
	for rows.Next() {
		item, err := scanAnimal(rows)
		if err != nil {
			h.internalError(w, "list scan", err)
			return
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		h.internalError(w, "list rows", err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := readAnimalID(r)
	if !ok {
		animalNotFound(w)
		return
	}

	item, err := scanAnimal(h.db.QueryRow(r.Context(), `SELECT `+animalColumns+`
		FROM animals a WHERE a.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		animalNotFound(w)
		return
	}
	if err != nil {
		h.internalError(w, "get", err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) getPhoto(w http.ResponseWriter, r *http.Request) {
	id, ok := readAnimalID(r)
	if !ok {
		animalNotFound(w)
		return
	}

	var contentType string
	var data []byte
	err := h.db.QueryRow(r.Context(), `SELECT content_type, data FROM animal_photos WHERE animal_id = $1`, id).Scan(&contentType, &data)
	if errors.Is(err, pgx.ErrNoRows) {
		httpapi.WriteError(w, http.StatusNotFound, "photo_not_found", "Фотография не найдена.")
		return
	}
	if err != nil {
		h.internalError(w, "get photo", err)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(data)
}

func scanAnimal(row pgx.Row) (animal, error) {
	var item animal
	var hasPhoto bool
	if err := row.Scan(&item.ID, &item.Name, &item.FoundLocation, &item.Breed, &item.CreatedAt, &hasPhoto); err != nil {
		return animal{}, err
	}
	item.CreatedAt = item.CreatedAt.UTC()
	if hasPhoto {
		item.PhotoURL = photoURL(item.ID)
	}
	return item, nil
}
