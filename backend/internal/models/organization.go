package models

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationType string

const (
	IE  OrganizationType = "IE"
	LLC OrganizationType = "LLC"
	JSC OrganizationType = "JSC"
)

type Organization struct {
	ID          uuid.UUID        `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name        string           `gorm:"type:varchar(100);not null"`
	Description string           `gorm:"type:text"`
	Type        OrganizationType `gorm:"type:varchar(10)"`
	CreatedAt   time.Time        `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time        `gorm:"default:CURRENT_TIMESTAMP"`
}

type OrganizationResponsible struct {
	ID             uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null"`
	UserID         uuid.UUID `gorm:"type:uuid;not null"`
}
