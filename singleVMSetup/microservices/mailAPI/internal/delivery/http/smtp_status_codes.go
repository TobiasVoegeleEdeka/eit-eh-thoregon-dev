package http

import (
	"net/http"
)

var SMTPStatusCodes = map[string]int{

	"2": http.StatusOK, // 250, 251, etc.

	// Temporäre Fehler (Retry later)
	"4":   http.StatusTooManyRequests, // 450, 452
	"421": http.StatusServiceUnavailable,

	// Permanente Fehler (Client muss Änderungen vornehmen)
	"5":   http.StatusBadRequest, // 500, 501, etc.
	"502": http.StatusBadGateway,
	"503": http.StatusServiceUnavailable,
	"550": http.StatusForbidden,           // "Mailbox unavailable"
	"553": http.StatusForbidden,           // "Sender not allowed"
	"554": http.StatusUnprocessableEntity, // Policy rejection (besser als 424)
}
