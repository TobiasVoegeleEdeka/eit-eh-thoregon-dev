package http

import (
	"encoding/json"
	"mailservice/internal/domain"
	"mailservice/internal/usecase"
	"net/http"
	"strings"
)

type Handler struct {
	mailer *usecase.Mailer
	logger Logger
}

type Logger interface {
	Printf(format string, v ...interface{})
}

// EnhancedResponse erweitert die Standardantwort
type EnhancedResponse struct {
	Status       string `json:"status"`
	SMTPCode     string `json:"smtp_code,omitempty"`
	Message      string `json:"message,omitempty"`
	BounceReason string `json:"bounce_reason,omitempty"`
	Error        string `json:"error,omitempty"`
}

func NewHandler(mailer *usecase.Mailer, logger Logger) *Handler {
	return &Handler{mailer: mailer, logger: logger}
}

func (h *Handler) SendEmail(w http.ResponseWriter, r *http.Request) {
	var req domain.EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, &EnhancedResponse{
			Status: "error",
			Error:  "Invalid request payload",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.mailer.SendEmail(&req)
	if err != nil {
		response := &EnhancedResponse{
			Status:       "error",
			SMTPCode:     result.SMTPCode,
			Message:      result.Message,
			BounceReason: result.BounceReason,
			Error:        err.Error(),
		}

		// Spezifischer HTTP-Statuscode basierend auf SMTP-Code
		statusCode := http.StatusInternalServerError
		if code, ok := SMTPStatusCodes[result.SMTPCode]; ok {
			statusCode = code
		} else if strings.HasPrefix(result.SMTPCode, "2") {
			statusCode = http.StatusOK
		} else if strings.HasPrefix(result.SMTPCode, "4") {
			statusCode = http.StatusTooManyRequests
		} else if strings.HasPrefix(result.SMTPCode, "5") {
			statusCode = http.StatusFailedDependency
		}

		h.respondError(w, response, statusCode)
		return
	}

	h.respondJSON(w, &EnhancedResponse{
		Status:   "success",
		SMTPCode: result.SMTPCode,
		Message:  result.Message,
	}, http.StatusOK)
}

func (h *Handler) respondJSON(w http.ResponseWriter, data *EnhancedResponse, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, response *EnhancedResponse, statusCode int) {
	h.logger.Printf("Request failed: %+v", response)
	h.respondJSON(w, response, statusCode)
}
