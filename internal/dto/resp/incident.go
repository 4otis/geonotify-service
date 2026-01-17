package resp

import "time"

type IncidentCreateResponse struct {
	IncidentID int `json:"incident_id"`
}

type IncidentResponse struct {
	IncidentID int       `json:"incident_id"`
	Name       string    `json:"name"`
	Descr      string    `json:"descr"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Radius     float64   `json:"radius_m"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type IncidentsListResponse struct {
	Incidents  []IncidentResponse `json:"incidents"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
}
