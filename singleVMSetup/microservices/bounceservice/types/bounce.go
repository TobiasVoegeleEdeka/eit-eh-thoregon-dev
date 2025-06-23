package types

type Bounce struct {
	Timestamp string `json:"timestamp"`
	QueueID   string `json:"queue_id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Status    string `json:"status"`
	Reason    string `json:"reason"`
	Raw       string `json:"raw,omitempty"`
}
