package auth

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/bypara/PickMe/internal/httpapi"
	"github.com/jackc/pgx/v5"
)

const maxCredentialsBody = 8 * 1024

type credentials struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	input, ok := readCredentials(w, r)
	if !ok {
		return
	}
	if fields := registrationErrors(input); len(fields) > 0 {
		httpapi.WriteFieldErrors(w, fields)
		return
	}

	var user account
	// Уникальный индекс проверяет email атомарно, в том числе при двух одновременных регистрациях.
	err := h.db.QueryRow(r.Context(), `
		INSERT INTO users (email, password_hash, full_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO NOTHING
		RETURNING id, email, full_name`, input.Email, hashPassword(input.Password), input.FullName,
	).Scan(&user.ID, &user.Email, &user.FullName)
	if errors.Is(err, pgx.ErrNoRows) {
		httpapi.WriteError(w, http.StatusConflict, "email_taken", "Этот email уже зарегистрирован.")
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}

	httpapi.WriteJSON(w, http.StatusCreated, user)
}

func readCredentials(w http.ResponseWriter, r *http.Request) (credentials, bool) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_body", "Отправьте данные в формате JSON.")
		return credentials{}, false
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxCredentialsBody))
	if err != nil {
		var limit *http.MaxBytesError
		if errors.As(err, &limit) {
			httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "Размер запроса превышает допустимый.")
		} else {
			httpapi.WriteError(w, http.StatusBadRequest, "invalid_body", "Не удалось прочитать данные запроса.")
		}
		return credentials{}, false
	}

	// Указатель отличает JSON null от объекта с пустыми полями.
	var input *credentials
	if err := json.Unmarshal(body, &input); err != nil || input == nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_body", "Отправьте данные в виде JSON-объекта.")
		return credentials{}, false
	}
	input.FullName = strings.TrimSpace(input.FullName)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	return *input, true
}

func registrationErrors(input credentials) map[string]string {
	fields := make(map[string]string)
	if length := utf8.RuneCountInString(input.FullName); length < 2 || length > 100 {
		fields["full_name"] = "Введите имя длиной от 2 до 100 символов."
	} else if strings.ContainsFunc(input.FullName, unicode.IsControl) {
		fields["full_name"] = "Введите имя и фамилию в одной строке."
	}
	address, err := mail.ParseAddress(input.Email)
	if err != nil || len(input.Email) > 254 || address.Name != "" || address.Address != input.Email {
		fields["email"] = "Введите почту в формате name@example.com."
	}
	if length := utf8.RuneCountInString(input.Password); length < 12 || length > 128 {
		fields["password"] = "Пароль должен содержать от 12 до 128 символов."
	}
	return fields
}
