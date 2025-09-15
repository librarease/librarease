package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Asset struct {
	ID        uuid.UUID       `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	OwnerID   uuid.UUID       `gorm:"column:owner_id;type:uuid;not null;index"`
	OwnerType string          `gorm:"column:owner_type;type:varchar(50);not null;index"`
	Kind      string          `gorm:"column:kind;type:varchar(50);not null;index"`
	Path      string          `gorm:"column:path;type:varchar(255);not null"`
	IsPrimary bool            `gorm:"column:is_primary;type:boolean;default:false"`
	Position  int             `gorm:"column:position;type:int;default:0"`
	Colors    datatypes.JSON  `gorm:"column:colors"`
	CreatedAt time.Time       `gorm:"column:created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Asset) TableName() string {
	return "assets"
}
