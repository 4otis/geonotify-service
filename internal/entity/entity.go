package entity

import (
	"errors"
	"time"
)

var (
	ErrIncidentNotFound   = errors.New("incident not found")
	ErrInvalidCoordinates = errors.New("invalid coordinates")
	ErrUserIDRequired     = errors.New("user_id is required")
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
	Payload     []byte
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ScheduledAt time.Time
}

type Check struct {
	ID        int
	UserID    string
	Latitude  float64
	Longitude float64
	HasAlert  bool
	CreatedAt time.Time
}
