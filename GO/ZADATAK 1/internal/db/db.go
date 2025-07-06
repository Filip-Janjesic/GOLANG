package db

import (
	"fmt"
	"log"
	"time"

	"zadatak-filip-janjesic/internal/models" // Import the models package

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitGormDB opens the SQLite database using GORM and performs any necessary migrations.
func InitGormDB() (*gorm.DB, error) {
	// Open the SQLite database using the GORM SQLite driver
	db, err := gorm.Open(sqlite.Open("./notes.db"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err) // Return an error if the database connection fails
	}

	// Perform database migrations for the User and Note models
	if err := db.AutoMigrate(&models.User{}, &models.Note{}); err != nil {
		return nil, fmt.Errorf("error migrating database: %w", err)
	}

	log.Println("Database successfully initialized.") // Log successful initialization
	return db, nil                                    // Return the GORM database connection
}

// GetUserByID retrieves a user's full details by their ID using GORM.
func GetUserByID(dbConn *gorm.DB, userID int) (models.User, error) {
	var user models.User
	// Retrieve the user by ID, ensuring it’s not deleted (soft delete handled with deleted_at)
	if err := dbConn.First(&user, "id = ? AND deleted_at IS NULL", userID).Error; err != nil {
		if err.Error() == "record not found" {
			return user, fmt.Errorf("user not found")
		}
		return user, err
	}
	return user, nil
}

// LoadNotes retrieves all active (non-deleted) notes from the database using GORM.
func LoadNotes(db *gorm.DB) ([]models.Note, error) {
	var notes []models.Note
	// Filter out deleted notes using soft delete logic (`deleted_at IS NULL`)
	if err := db.Where("deleted_at IS NULL").Find(&notes).Error; err != nil {
		return nil, fmt.Errorf("error loading notes: %w", err)
	}
	return notes, nil
}

// InsertNote adds a new note to the database for a given user using GORM.
func InsertNote(db *gorm.DB, note *models.Note) error {
	// GORM will automatically populate the `CreatedAt` and `UpdatedAt` timestamps if they are defined in the model.
	if err := db.Create(note).Error; err != nil {
		return fmt.Errorf("error inserting note: %w", err)
	}
	return nil
}

// UpdateNoteInDB modifies an existing note in the database using GORM.
func UpdateNoteInDB(db *gorm.DB, note *models.Note) error {
	// Update the `UpdatedAt` field manually if needed
	note.UpdatedAt = time.Now()
	if err := db.Save(note).Error; err != nil {
		return fmt.Errorf("error updating note: %w", err)
	}
	return nil
}

// SoftDeleteNoteInDB sets the `deleted_at` timestamp to mark a note as deleted using GORM.
func SoftDeleteNoteInDB(db *gorm.DB, noteID int) error {
	// Perform a soft delete by setting `deleted_at` to the current time
	if err := db.Model(&models.Note{}).Where("id = ?", noteID).Update("deleted_at", time.Now()).Error; err != nil {
		return fmt.Errorf("error soft deleting note: %w", err)
	}
	return nil
}

// NoteExists checks if a specific note exists for the given user and is active (not deleted) using GORM.
func NoteExists(db *gorm.DB, noteID, userID int) (bool, error) {
	var count int64
	// Count the notes with the given user ID and ensure it's not deleted
	err := db.Model(&models.Note{}).Where("id = ? AND user_id = ? AND deleted_at IS NULL", noteID, userID).Count(&count).Error
	return count > 0, err
}
