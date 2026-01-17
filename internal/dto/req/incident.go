package req

type IncidentCreateRequest struct {
	Name      string  `json:"name"`
	Descr     string  `json:"descr"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Radius    float64 `json:"radius_m"`
}

type IncidentUpdateRequest struct {
	Name      *string  `json:"name,omitempty"`
	Descr     *string  `json:"descr,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	Radius    *float64 `json:"radius_m,omitempty"`
	IsActive  *bool    `json:"is_active,omitempty"`
}
