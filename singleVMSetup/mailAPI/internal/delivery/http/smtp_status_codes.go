package http

import (
	"net/http"
)

// SMTPStatusCodes enthält die Zuordnung von SMTP-Codes zu HTTP-Statuscodes
var SMTPStatusCodes = map[string]int{
	"2":   http.StatusOK,
	"4":   http.StatusTooManyRequests,
	"5":   http.StatusFailedDependency,
	"421": http.StatusServiceUnavailable,
	"450": http.StatusServiceUnavailable,
	"451": http.StatusServiceUnavailable,
	"452": http.StatusPaymentRequired,
	"500": http.StatusBadRequest,
	"501": http.StatusBadRequest,
	"502": http.StatusBadGateway,
	"503": http.StatusServiceUnavailable,
	"504": http.StatusGatewayTimeout,
	"550": http.StatusNotFound,
	"551": http.StatusNotFound,
	"552": http.StatusPaymentRequired,
	"553": http.StatusForbidden,
	"554": http.StatusFailedDependency,
}
