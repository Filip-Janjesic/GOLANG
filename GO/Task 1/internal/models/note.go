package models

import (
	"time"

	"gorm.io/gorm"
)

// Note represents a single note in the system with fields for tracking ownership,
// content, creation and update times, and an optional soft delete timestamp.
type Note struct {
	gorm.Model            // Embeds ID, CreatedAt, UpdatedAt, and DeletedAt (for soft delete support)
	UserID     int        `json:"user_id" gorm:"not null;index" validate:"required"`               // Foreign key for user
	User       User       `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // Foreign key reference with cascading delete and update
	Title      string     `json:"title" gorm:"not null" validate:"required"`                       // Title of the note
	Body       string     `json:"body" gorm:"not null" validate:"required"`                        // Content of the note
	DeletedAt  *time.Time `json:"deleted_at,omitempty" gorm:"index" validate:"omitempty"`          // Nullable timestamp for soft delete; indexed for performance
}
