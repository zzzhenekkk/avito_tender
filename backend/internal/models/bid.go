package models

import (
	"time"

	"github.com/google/uuid"
)

type BidStatus string

const (
	BidStatusCreated   BidStatus = "Created"
	BidStatusPublished BidStatus = "Published"
	BidStatusCanceled  BidStatus = "Canceled"
	BidStatusApproved  BidStatus = "Approved"
	BidStatusRejected  BidStatus = "Rejected"
)

type BidDecision string

const (
	BidDecisionApproved BidDecision = "Approved"
	BidDecisionRejected BidDecision = "Rejected"
)

type BidAuthorType string

const (
	BidAuthorTypeOrganization BidAuthorType = "Organization"
	BidAuthorTypeUser         BidAuthorType = "User"
)

type Bid struct {
	ID          uuid.UUID     `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name        string        `gorm:"type:varchar(100);not null"`
	Description string        `gorm:"type:varchar(500)"`
	Status      BidStatus     `gorm:"type:varchar(20)"`
	TenderID    uuid.UUID     `gorm:"type:uuid;not null"`
	AuthorType  BidAuthorType `gorm:"type:varchar(20)"`
	AuthorID    uuid.UUID     `gorm:"type:uuid;not null"`
	Version     int           `gorm:"default:1"`
	CreatedAt   time.Time     `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time     `gorm:"default:CURRENT_TIMESTAMP"`
}

type BidVersion struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	BidID       uuid.UUID `gorm:"type:uuid;not null"`
	Version     int       `gorm:"not null"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Description string    `gorm:"type:varchar(500)"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type BidFeedback struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	BidID     uuid.UUID `gorm:"type:uuid;not null"`
	Feedback  string    `gorm:"type:varchar(1000)"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}
