package usecase

import (
	"time"

	"github.com/google/uuid"
)

type Asset struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	OwnerType string
	Kind      string
	Path      string
	IsPrimary bool
	Position  int
	Colors    []byte
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
