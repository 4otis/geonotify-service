package resp

import "time"

type IncidentResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Lattitude float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Radius    float64   `json:"radius_m"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type IncidentListResponse struct {
	Incidents []IncidentResponse `json:"incidents"`
	Page      int                `json:"page"`
	Limit     int                `json:"limit"`
	Total     int                `json:"total"`
}
