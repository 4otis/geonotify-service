package resp

type LocationCheckResponse struct {
	HasAlert  bool               `json:"has_alert"`
	Incidents []IncidentResponse `json:"incidents,omitempty"`
}
