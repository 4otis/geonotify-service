package resp

import "time"

type IncidentCreateResponse struct {
	IncidentID int `json:"incident_id"`
}

type IncidentResponse struct {
	IncidentID string    `json:"incident_id"`
	Name       string    `json:"name"`
	Descr      string    `json:"descr"`
	Lattitude  float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Radius     float64   `json:"radius_m"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

type IncidentListResponse struct {
	Incidents []IncidentResponse `json:"incidents"`
	Page      int                `json:"page"`
	Limit     int                `json:"limit"`
	Total     int                `json:"total"`
}
