package req

type IncidentCreateRequest struct {
	Name      string  `json:"name"`
	Descr     string  `json:"descr"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Radius    float64 `json:"radius_m"`
}

type IncidentUpdateRequest struct {
	Name      string  `json:"name"`
	Descr     string  `json:"descr"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Radius    float64 `json:"radius_m"`
	IsActive  bool    `json:"is_active"`
}
