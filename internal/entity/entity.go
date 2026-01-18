package entity

import (
	"errors"
	"time"
)

var (
	ErrIncidentNotFound = errors.New("incident not found")
)

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

type Webhook struct {
	ID          int
	CheckID     int
	State       string
	RetryCnt    int
	CreatedAt   time.Time
	DeliveredAt time.Time
	Payload     []byte
}

type Check struct {
	ID        int
	UserID    int
	Latitude  float64
	Longitude float64
	HasAlert  bool
	CreatedAt time.Time
}
