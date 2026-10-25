package core_http_response

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	core_errors "github.com/dyingvoid/shorturl/gateway/internal/core/errors"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

type Handler struct {
	log *slog.Logger
	w   http.ResponseWriter
}

func NewHandler(log *slog.Logger, w http.ResponseWriter) *Handler {
	return &Handler{log: log, w: w}
}

func (h *Handler) JSON(body any, status int) {
	h.w.Header().Set("Content-Type", "application/json; charset=utf-8")
	h.w.WriteHeader(status)
	if err := json.NewEncoder(h.w).Encode(body); err != nil {
		h.log.Error("encode JSON response", "error", err)
	}
}

func (h *Handler) NoContent() {
	h.w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Redirect(location string, status int) {
	h.w.Header().Set("Location", location)
	h.w.WriteHeader(status)
}

func (h *Handler) Error(err error, msg string) {
	status, logFn := http.StatusInternalServerError, h.log.Error
	switch {
	case errors.Is(err, core_errors.ErrInvalidArgument):
		status, logFn = http.StatusBadRequest, h.log.Warn
	case errors.Is(err, core_errors.ErrUnauthenticated):
		status, logFn = http.StatusUnauthorized, h.log.Warn
	case errors.Is(err, core_errors.ErrPermissionDenied):
		status, logFn = http.StatusForbidden, h.log.Warn
	case errors.Is(err, core_errors.ErrNotFound):
		status, logFn = http.StatusNotFound, h.log.Debug
	case errors.Is(err, core_errors.ErrConflict):
		status, logFn = http.StatusConflict, h.log.Warn
	case errors.Is(err, core_errors.ErrResourceExhausted):
		status, logFn = http.StatusTooManyRequests, h.log.Warn
	case errors.Is(err, core_errors.ErrDeadlineExceeded):
		status, logFn = http.StatusGatewayTimeout, h.log.Warn
	case errors.Is(err, core_errors.ErrUnavailable):
		status, logFn = http.StatusServiceUnavailable, h.log.Warn
	}

	logFn(msg, "error", err)

	errorText := err.Error()
	if status == http.StatusInternalServerError {
		errorText = "internal error"
	}

	h.JSON(ErrorResponse{Message: msg, Error: errorText}, status)
}

func (h *Handler) Panic(value any, msg string) {
	h.log.Error(msg, "panic", value, "stack", string(debug.Stack()))
	h.JSON(ErrorResponse{Message: msg, Error: msg}, http.StatusInternalServerError)
}
