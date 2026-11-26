package httpapi

import (
	"encoding/json"
	"net/http"
)

type responseError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// WriteJSON отправляет JSON-ответ с заданным HTTP-статусом.
func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

// WriteError отправляет ошибку в общем формате API.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	writeError(w, status, responseError{Code: code, Message: message})
}

// WriteFieldErrors отправляет ошибки валидации, привязанные к полям формы.
func WriteFieldErrors(w http.ResponseWriter, fields map[string]string) {
	writeError(w, http.StatusUnprocessableEntity, responseError{
		Code: "validation_failed", Message: "Проверьте поля формы.", Fields: fields,
	})
}

func writeError(w http.ResponseWriter, status int, details responseError) {
	WriteJSON(w, status, struct {
		Error responseError `json:"error"`
	}{Error: details})
}
