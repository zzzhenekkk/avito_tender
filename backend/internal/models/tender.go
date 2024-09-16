package models

import (
	"time"

	"github.com/google/uuid"
)

type TenderStatus string

const (
	TenderStatusCreated   TenderStatus = "Created"
	TenderStatusPublished TenderStatus = "Published"
	TenderStatusClosed    TenderStatus = "Closed"
)

type ServiceType string

const (
	ServiceTypeConstruction ServiceType = "Construction"
	ServiceTypeDelivery     ServiceType = "Delivery"
	ServiceTypeManufacture  ServiceType = "Manufacture"
)

type Tender struct {
	ID             uuid.UUID    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name           string       `gorm:"type:varchar(100);not null"`
	Description    string       `gorm:"type:varchar(500)"`
	ServiceType    ServiceType  `gorm:"type:varchar(20)"`
	Status         TenderStatus `gorm:"type:varchar(20)"`
	OrganizationID uuid.UUID    `gorm:"type:uuid;not null"`
	Version        int          `gorm:"default:1"`
	CreatedAt      time.Time    `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time    `gorm:"default:CURRENT_TIMESTAMP"`
}

type TenderVersion struct {
	ID          uuid.UUID   `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	TenderID    uuid.UUID   `gorm:"type:uuid;not null"`
	Version     int         `gorm:"not null"`
	Name        string      `gorm:"type:varchar(100);not null"`
	Description string      `gorm:"type:varchar(500)"`
	ServiceType ServiceType `gorm:"type:varchar(20)"`
	CreatedAt   time.Time   `gorm:"default:CURRENT_TIMESTAMP"`
}
