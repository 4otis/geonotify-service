package resp

import "time"

type HealthResponse struct {
	Status          string    `json:"status"`
	Timestamp       time.Time `json:"timestamp"`
	ActiveIncidents int       `json:"active_incidents"`
	PendingWebhooks int       `json:"pending_webhooks"`
}
