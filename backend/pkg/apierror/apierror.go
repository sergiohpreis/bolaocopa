package apierror

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Write(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(Error{Code: code, Message: message}); err != nil {
		http.Error(w, `{"code":"internal_error","message":"failed to encode error response"}`, http.StatusInternalServerError)
	}
}

func BadRequest(w http.ResponseWriter, message string) {
	Write(w, http.StatusBadRequest, "bad_request", message)
}

func NotFound(w http.ResponseWriter, message string) {
	Write(w, http.StatusNotFound, "not_found", message)
}

func Unauthorized(w http.ResponseWriter, message string) {
	Write(w, http.StatusUnauthorized, "unauthorized", message)
}

func Forbidden(w http.ResponseWriter, message string) {
	Write(w, http.StatusForbidden, "forbidden", message)
}

func Conflict(w http.ResponseWriter, message string) {
	Write(w, http.StatusConflict, "conflict", message)
}

func Internal(w http.ResponseWriter, err error) {
	slog.Error("internal server error", "error", err)
	Write(w, http.StatusInternalServerError, "internal_error", "internal server error")
}
