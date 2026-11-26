package animals

import (
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/bypara/PickMe/internal/auth"
	"github.com/bypara/PickMe/internal/httpapi"
	"github.com/jackc/pgx/v5"
)

const maxAnimalFormSize = 6 * 1024 * 1024

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	createdBy, _ := auth.CurrentUser(r)
	r.Body = http.MaxBytesReader(w, r.Body, maxAnimalFormSize)
	// Лимит тела ограничивает память. Контейнер API может обработать форму без записи на диск.
	err := r.ParseMultipartForm(maxAnimalFormSize)
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}
	if err != nil {
		var limit *http.MaxBytesError
		if errors.As(err, &limit) || errors.Is(err, multipart.ErrMessageTooLarge) {
			httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "Размер формы превышает 6 МиБ.")
			return
		}
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_body", "Отправьте данные в формате multipart/form-data.")
		return
	}

	item, fields := readAnimalFields(r.MultipartForm.Value)
	if len(fields) > 0 {
		httpapi.WriteFieldErrors(w, fields)
		return
	}
	photo, err := readPhoto(r.MultipartForm)
	switch {
	case errors.Is(err, errPhotoTooLarge):
		httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "Размер фотографии превышает 5 МиБ.")
		return
	case errors.Is(err, errInvalidPhoto), errors.Is(err, errPhotoDimensions), errors.Is(err, errMultiplePhotos):
		httpapi.WriteFieldErrors(w, map[string]string{"photo": err.Error()})
		return
	case err != nil:
		h.internalError(w, "read upload", err)
		return
	}

	ctx := r.Context()
	// Фото и карточка становятся видны вместе после подтверждения транзакции.
	err = pgx.BeginFunc(ctx, h.db, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `
			INSERT INTO animals (name, found_location, breed, created_by)
			VALUES ($1, $2, $3, $4) RETURNING id, created_at`,
			item.Name, item.FoundLocation, item.Breed, createdBy,
		).Scan(&item.ID, &item.CreatedAt)
		if err != nil || photo == nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO animal_photos (animal_id, content_type, data) VALUES ($1, $2, $3)`,
			item.ID, photo.contentType, photo.data)
		return err
	})
	if err != nil {
		h.internalError(w, "create", err)
		return
	}

	item.CreatedAt = item.CreatedAt.UTC()
	if photo != nil {
		item.PhotoURL = photoURL(item.ID)
	}
	w.Header().Set("Location", "/api/animals/"+strconv.FormatInt(item.ID, 10))
	httpapi.WriteJSON(w, http.StatusCreated, item)
}

func readAnimalFields(values map[string][]string) (animal, map[string]string) {
	var item animal
	fields := make(map[string]string)
	for _, field := range []struct {
		key     string
		max     int
		message string
		target  *string
	}{
		{"name", 100, "Введите кличку длиной от 1 до 100 символов.", &item.Name},
		{"found_location", 255, "Укажите место обнаружения длиной от 1 до 255 символов.", &item.FoundLocation},
		{"breed", 100, "Введите породу длиной от 1 до 100 символов.", &item.Breed},
	} {
		parts := values[field.key]
		if len(parts) != 1 {
			fields[field.key] = field.message
			continue
		}
		value := strings.TrimSpace(parts[0])
		if value == "" || !utf8.ValidString(value) || utf8.RuneCountInString(value) > field.max || strings.ContainsFunc(value, unicode.IsControl) {
			fields[field.key] = field.message
			continue
		}
		*field.target = value
	}
	return item, fields
}
