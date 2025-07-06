package handlers

import (
	"strconv"
	"time"
	"zadatak-filip-janjesic/internal/models"

	"github.com/go-playground/validator/v10" // Import the validator package
	"github.com/gofiber/fiber/v2"            // Import Fiber package
	"gorm.io/gorm"                           // Import GORM for database handling
)

// NotesHandler handles HTTP requests for notes, with caching functionality.
func NotesHandler(database *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		switch c.Method() {
		case "GET":
			return getNotes(database, c) // Get notes with cache check
		case "POST":
			return createNote(database, c)
		case "PUT":
			return updateNote(database, c)
		case "DELETE":
			return deleteNote(database, c)
		default:
			return c.Status(fiber.StatusMethodNotAllowed).SendString("Method not allowed")
		}
	}
}

// sendJSONResponse sends a JSON response with the specified status code.
func sendJSONResponse(c *fiber.Ctx, data interface{}, statusCode int) error {
	c.Status(statusCode)
	return c.JSON(data)
}

// getNotes retrieves all active (non-deleted) notes for the authenticated user.
func getNotes(database *gorm.DB, c *fiber.Ctx) error {
	// Extract user ID from JWT token
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString(err.Error())
	}

	// Try to load notes from cache first
	notes, err := models.LoadNotes(database, userID) // Pass database as first argument
	if err != nil || len(notes) == 0 {               // Cache miss, fetch from DB
		notes, err = fetchUserNotes(database, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Database error")
		}
		// Save notes to cache for future requests
		models.SaveNotes(database, userID, notes) // Pass database as first argument
	}

	// Return the notes as JSON
	return sendJSONResponse(c, notes, fiber.StatusOK)
}

// fetchUserNotes retrieves active notes from the database for the given user.
func fetchUserNotes(database *gorm.DB, userID int) ([]models.Note, error) {
	var notes []models.Note
	// Use GORM to fetch notes where deleted_at is NULL (not deleted)
	if err := database.Where("user_id = ? AND deleted_at IS NULL", userID).Find(&notes).Error; err != nil {
		return nil, err
	}
	return notes, nil
}

// createNote creates a new note for the authenticated user.
func createNote(database *gorm.DB, c *fiber.Ctx) error {
	var note models.Note
	// Parse the incoming request body to get note data
	if err := c.BodyParser(&note); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid input")
	}

	// Validate the note fields
	validate := validator.New()
	if err := validate.Struct(note); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Validation failed")
	}

	// Extract user ID from JWT token to associate the note with the user
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString(err.Error())
	}
	note.UserID = userID
	// Timestamps are managed by GORM automatically.

	// Insert note into the database using GORM
	if err := database.Create(&note).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Unable to create note")
	}

	// Clear cache for the user to ensure we fetch updated data
	models.ClearNotesCache(database) // Pass database as first argument

	// Return the newly created note as JSON
	return sendJSONResponse(c, note, fiber.StatusCreated)
}

// updateNote updates an existing note.
func updateNote(database *gorm.DB, c *fiber.Ctx) error {
	// Get note ID from the URL (path parameter)
	noteIDStr := c.Params("id")

	// Convert the note ID from string to uint
	noteID, err := strconv.ParseUint(noteIDStr, 10, 32) // Parse as uint
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid note ID")
	}

	var note models.Note
	// Parse the incoming request body to get the updated note data
	if err := c.BodyParser(&note); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid input")
	}

	// Validate the note fields
	validate := validator.New()
	if err := validate.Struct(note); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Validation failed")
	}

	// Extract user ID from JWT token to ensure they are updating their own note
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString(err.Error())
	}

	// Check if the note exists and belongs to the user using GORM
	var existingNote models.Note
	if err := database.Where("id = ? AND user_id = ?", noteID, userID).First(&existingNote).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).SendString("Note not found")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Database error")
	}

	// Update the note fields using GORM
	note.ID = uint(noteID) // Convert noteID (uint) to match the ID field type
	note.UserID = userID
	if err := database.Save(&note).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Unable to update note")
	}

	// Clear cache for the user to ensure we fetch updated data
	models.ClearNotesCache(database) // Pass database as first argument

	// Return the updated note as JSON
	return sendJSONResponse(c, note, fiber.StatusOK)
}

// deleteNote marks a note as deleted by setting the deleted_at timestamp.
func deleteNote(database *gorm.DB, c *fiber.Ctx) error {
	// Get note ID from the URL
	noteIDStr := c.Params("id")

	// Convert the note ID from string to uint
	noteID, err := strconv.ParseUint(noteIDStr, 10, 32) // Parse as uint
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid note ID")
	}

	// Extract user ID from JWT token to ensure they are deleting their own note
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString(err.Error())
	}

	// Check if the note exists and belongs to the user using GORM
	var existingNote models.Note
	if err := database.Where("id = ? AND user_id = ?", noteID, userID).First(&existingNote).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).SendString("Note not found")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Database error")
	}

	// Soft delete the note by setting the deleted_at timestamp
	currentTime := time.Now()
	existingNote.DeletedAt = &currentTime // Set the DeletedAt field to the current time

	if err := database.Save(&existingNote).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Unable to delete note")
	}

	// Clear cache for the user to ensure we fetch updated data
	models.ClearNotesCache(database) // Pass database as first argument

	// Return no content response as the note is deleted
	return c.SendStatus(fiber.StatusNoContent)
}
