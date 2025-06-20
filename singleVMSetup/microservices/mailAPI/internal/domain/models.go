package domain

type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type EmailResponse struct {
	Status string `json:"status"`
}
