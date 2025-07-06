package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// Cache represents a cached entry in the database for user notes.
type Cache struct {
	gorm.Model           // Embeds ID, CreatedAt, UpdatedAt, and DeletedAt fields
	UserID     int       `json:"user_id" gorm:"not null;index"` // The user ID associated with the cache
	Notes      string    `json:"notes" gorm:"not null"`         // Cached notes data stored as JSON string
	ExpiryDate time.Time `json:"expiry_date" gorm:"type:date"`  // Expiry date for the cache entry
}

// In-memory cache map for user notes, indexed by user ID.
var notesCache = make(map[int][]Note)

// Mutex to ensure thread-safe access to the in-memory cache.
var cacheMutex = sync.RWMutex{}

// AutoMigrateCache migrates the cache table in the database.
func AutoMigrateCache(db *gorm.DB) error {
	return db.AutoMigrate(&Cache{})
}

// SaveNotes saves the provided notes to both the in-memory cache and the database.
func SaveNotes(db *gorm.DB, userID int, notes []Note) error {
	// Lock for writing to ensure thread-safe cache access
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Save notes to in-memory cache
	notesCache[userID] = notes

	// Marshal notes to JSON for database storage
	notesData, err := json.Marshal(notes)
	if err != nil {
		return fmt.Errorf("error serializing notes for cache: %w", err)
	}

	// Use a transaction to create or update the cache in the database
	return db.Transaction(func(tx *gorm.DB) error {
		var cache Cache
		if err := tx.Where("user_id = ?", userID).First(&cache).Error; err == nil {
			// Update existing cache entry
			cache.Notes = string(notesData)
			cache.ExpiryDate = time.Now().Add(24 * time.Hour)
			return tx.Save(&cache).Error
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new cache entry if none exists
			cache = Cache{
				UserID:     userID,
				Notes:      string(notesData),
				ExpiryDate: time.Now().Add(24 * time.Hour),
			}
			return tx.Create(&cache).Error
		} else {
			return err
		}
	})
}

// LoadNotes retrieves notes from in-memory cache or the database if not found.
func LoadNotes(db *gorm.DB, userID int) ([]Note, error) {
	// Lock for reading to safely access the cache
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	// Check the in-memory cache
	if notes, found := notesCache[userID]; found {
		return notes, nil
	}

	// Load notes from the database cache if not in memory
	var cache Cache
	if err := db.Where("user_id = ?", userID).First(&cache).Error; err != nil {
		return nil, fmt.Errorf("error loading notes from database cache: %w", err)
	}

	// Unmarshal JSON data from the cache
	var notes []Note
	if err := json.Unmarshal([]byte(cache.Notes), &notes); err != nil {
		return nil, fmt.Errorf("error unmarshalling notes from cache: %w", err)
	}

	// Store the notes in-memory cache for future use
	notesCache[userID] = notes
	return notes, nil
}

// ClearNotesCache clears the in-memory cache and deletes cache entries from the database.
func ClearNotesCache(db *gorm.DB) error {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Clear in-memory cache
	notesCache = make(map[int][]Note)

	// Delete all entries from the database cache
	if err := db.Delete(&Cache{}).Error; err != nil {
		log.Printf("error clearing cache entries from database: %v", err)
		return err
	}
	return nil
}
