package models

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// User represents a user account in the system with detailed personal, address, and account information.
type User struct {
	gorm.Model            // Embeds fields like ID, CreatedAt, UpdatedAt, and DeletedAt (for soft delete support)
	Username    string    `json:"username" gorm:"unique;not null" validate:"required,min=3,max=32"`
	Password    string    `json:"password" gorm:"not null" validate:"required,min=8"`
	FirstName   string    `json:"first_name" gorm:"not null" validate:"required"`
	LastName    string    `json:"last_name" gorm:"not null" validate:"required"`
	Email       string    `json:"email" gorm:"unique;not null" validate:"required,email"`
	PhoneNumber string    `json:"phone_number,omitempty" gorm:"type:varchar(20)" validate:"omitempty"` // Changed: removed "gorm:\"-\""
	City        string    `json:"city,omitempty" gorm:"type:varchar(100)" validate:"omitempty"`        // Changed: removed "gorm:\"-\""
	Country     string    `json:"country,omitempty" gorm:"type:varchar(100)" validate:"omitempty"`     // Changed: removed "gorm:\"-\""
	DateOfBirth time.Time `json:"dateOfBirth" validate:"omitempty,datetime=2006-01-02" example:"2006-01-02"`
}

// ValidateUser validates user data before registration or login.
func ValidateUser(user User) error {
	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			fmt.Printf("Field %s failed validation with tag %s\n", e.Field(), e.Tag())
		}
		return err
	}
	return nil
}
