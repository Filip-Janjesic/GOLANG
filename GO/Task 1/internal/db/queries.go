package db

import (
	"fmt"
	"time"

	"zadatak-filip-janjesic/internal/models" // Import the models package

	"gorm.io/gorm"
)

// CheckExistingUser checks if a user with the given username already exists in the database.
// It returns true if the user exists, and false otherwise, along with any error encountered.
func CheckExistingUser(db *gorm.DB, username string) (bool, error) {
	var count int64
	// Query the database to check if a record exists for the given username
	if err := db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, fmt.Errorf("error checking if user exists: %w", err)
	}
	return count > 0, nil
}

// InsertNewUser inserts a new user into the database with the current timestamp for created_at and updated_at fields.
// The function receives a 'user' object and returns an error if something goes wrong.
func InsertNewUser(db *gorm.DB, user models.User) error {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	// Insert the user into the database
	if err := db.Create(&user).Error; err != nil {
		return fmt.Errorf("error inserting new user: %w", err)
	}
	return nil
}

// GetUserID retrieves the user ID for a given username from the database.
// The function returns the user ID and any error encountered during the process.
func GetUserID(db *gorm.DB, username string) (uint, error) {
	var user models.User
	// Query the database for the user ID based on the provided username
	if err := db.Model(&models.User{}).Where("username = ?", username).First(&user).Error; err != nil {
		return 0, fmt.Errorf("error retrieving user ID: %w", err)
	}
	return user.ID, nil
}

// GetUserByUsername retrieves the user's details (ID, username, password) based on the provided username.
// The function returns the user struct and any error encountered.
func GetUserByUsername(db *gorm.DB, username string) (models.User, error) {
	var user models.User
	// Query the database to retrieve the user's details
	if err := db.Model(&models.User{}).Where("username = ?", username).First(&user).Error; err != nil {
		return user, fmt.Errorf("error retrieving user by username: %w", err)
	}
	return user, nil
}

// UpdateUserTimestamp updates the 'updated_at' timestamp for a user whenever their information is modified.
// The function receives the user's ID and returns an error if the update fails.
func UpdateUserTimestamp(db *gorm.DB, userID uint) error {
	now := time.Now()
	// Update the 'updated_at' field for the user with the specified ID
	if err := db.Model(&models.User{}).Where("id = ?", userID).Update("updated_at", now).Error; err != nil {
		return fmt.Errorf("error updating user timestamp: %w", err)
	}
	return nil
}
