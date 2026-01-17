package entity

import "time"

type Incident struct {
	ID        int
	Name      string
	Descr     string
	Latitude  float64
	Longitude float64
	Radius    float64
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
