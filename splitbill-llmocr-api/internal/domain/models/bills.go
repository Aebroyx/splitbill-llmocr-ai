package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Bills represents the bills table
type Bills struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string         `json:"name" gorm:"size:255"`
	Status    string         `json:"status" gorm:"size:20;not null;default:'active'"`
	TaxAmount float64        `json:"tax_amount" gorm:"type:numeric(10,2);default:0.00"`
	TipAmount float64        `json:"tip_amount" gorm:"type:numeric(10,2);default:0.00"`
	CreatedAt time.Time      `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Items        []Items        `json:"items,omitempty" gorm:"foreignKey:BillID"`
	Participants []Participants `json:"participants,omitempty" gorm:"foreignKey:BillID"`
}

// Items represents the items table
type Items struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	BillID    uuid.UUID `json:"bill_id" gorm:"type:uuid;not null"`
	Name      string    `json:"name" gorm:"size:255;not null"`
	Price     float64   `json:"price" gorm:"type:numeric(10,2);not null"`
	Quantity  int       `json:"quantity" gorm:"not null;default:1"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Bill            Bills             `json:"bill,omitempty" gorm:"foreignKey:BillID"`
	ItemAssignments []ItemAssignments `json:"item_assignments,omitempty" gorm:"foreignKey:ItemID"`
}

// Participants represents the participants table
type Participants struct {
	ID                 uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	BillID             uuid.UUID `json:"bill_id" gorm:"type:uuid;not null"`
	Name               string    `json:"name" gorm:"size:255;not null"`
	PaymentStatus      string    `json:"payment_status" gorm:"size:20;not null;default:'unpaid'"`
	ShareOfCommonCosts float64   `json:"share_of_common_costs" gorm:"type:numeric(10,2);default:0.00"`
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Bill            Bills             `json:"bill,omitempty" gorm:"foreignKey:BillID"`
	ItemAssignments []ItemAssignments `json:"item_assignments,omitempty" gorm:"foreignKey:ParticipantID"`
}

// ItemAssignments represents the item_assignments table (join table)
type ItemAssignments struct {
	ItemID        uint      `json:"item_id" gorm:"primaryKey"`
	ParticipantID uint      `json:"participant_id" gorm:"primaryKey"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Item        Items        `json:"item,omitempty" gorm:"foreignKey:ItemID"`
	Participant Participants `json:"participant,omitempty" gorm:"foreignKey:ParticipantID"`
}

// BillRequest represents the request payload for creating/updating a bill
type BillRequest struct {
	Name      string  `json:"name" validate:"required,max=255"`
	TaxAmount float64 `json:"tax_amount" validate:"gte=0"`
	TipAmount float64 `json:"tip_amount" validate:"gte=0"`
}

// BillResponse represents the response payload for a bill
type BillResponse struct {
	ID           uuid.UUID             `json:"id"`
	Name         string                `json:"name"`
	Status       string                `json:"status"`
	TaxAmount    float64               `json:"tax_amount"`
	TipAmount    float64               `json:"tip_amount"`
	CreatedAt    time.Time             `json:"created_at"`
	Items        []ItemResponse        `json:"items,omitempty"`
	Participants []ParticipantResponse `json:"participants,omitempty"`
}

// ItemRequest represents the request payload for creating/updating an item
type ItemRequest struct {
	Name     string  `json:"name" validate:"required,max=255"`
	Price    float64 `json:"price" validate:"required,gt=0"`
	Quantity int     `json:"quantity" validate:"required,gt=0"`
}

// ItemResponse represents the response payload for an item
type ItemResponse struct {
	ID        uint      `json:"id"`
	BillID    uuid.UUID `json:"bill_id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
}

// ParticipantRequest represents the request payload for creating/updating a participant
type ParticipantRequest struct {
	Name               string  `json:"name" validate:"required,max=255"`
	ShareOfCommonCosts float64 `json:"share_of_common_costs" validate:"gte=0"`
}

// ParticipantResponse represents the response payload for a participant
type ParticipantResponse struct {
	ID                 uint      `json:"id"`
	BillID             uuid.UUID `json:"bill_id"`
	Name               string    `json:"name"`
	PaymentStatus      string    `json:"payment_status"`
	ShareOfCommonCosts float64   `json:"share_of_common_costs"`
	CreatedAt          time.Time `json:"created_at"`
}

// ItemAssignmentRequest represents the request payload for assigning items to participants
type ItemAssignmentRequest struct {
	ItemID        uint `json:"item_id" validate:"required"`
	ParticipantID uint `json:"participant_id" validate:"required"`
}

// BillSummary represents a summary of bill calculations
type BillSummary struct {
	BillID            uuid.UUID          `json:"bill_id"`
	TotalItems        float64            `json:"total_items"`
	TaxAmount         float64            `json:"tax_amount"`
	TipAmount         float64            `json:"tip_amount"`
	TotalBill         float64            `json:"total_bill"`
	ParticipantShares map[string]float64 `json:"participant_shares"`
}

// ExtractedItemData represents the structure of extracted item data from LLM
type ExtractedItemData struct {
	Items []ExtractedItem `json:"items"`
	Tax   float64         `json:"tax"`
	Tip   float64         `json:"tip"`
	Total float64         `json:"total"`
}

// ExtractedItem represents a single item extracted from the bill
type ExtractedItem struct {
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}
