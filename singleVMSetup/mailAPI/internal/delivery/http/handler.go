package http

import (
	"encoding/json"
	"mailservice/internal/domain"
	"mailservice/internal/usecase"
	"net/http"
)

type Handler struct {
	mailer *usecase.Mailer
	logger Logger
}

type Logger interface {
	Printf(format string, v ...interface{})
}

func NewHandler(mailer *usecase.Mailer, logger Logger) *Handler {
	return &Handler{mailer: mailer, logger: logger}
}

func (h *Handler) SendEmail(w http.ResponseWriter, r *http.Request) {
	var req domain.EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if err := h.mailer.SendEmail(&req); err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, domain.EmailResponse{Status: "success"})
}

func (h *Handler) respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, message string, code int) {
	w.WriteHeader(code)
	h.respondJSON(w, map[string]string{"error": message})
}
