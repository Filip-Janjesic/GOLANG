package handlers

import (
	"fmt"
	"os"
	"strings"

	"zadatak-filip-janjesic/internal/models" // Import models for user struct

	"github.com/go-playground/validator/v10" // Importing the validator for data validation
	"github.com/gofiber/fiber/v2"            // Fiber framework for web server
	"github.com/golang-jwt/jwt/v4"           // JWT library for token handling
	"github.com/joho/godotenv"               // Load env variables from .env file
	"gorm.io/gorm"                           // GORM for ORM and database operations
)

// Load environment variables from .env when package is initialized.
func init() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file")
	}
}

// getUserIDFromToken extracts the user ID from the JWT token in the Authorization header.
func getUserIDFromToken(c *fiber.Ctx) (int, error) {
	// Get the token from the Authorization header
	tokenString := c.Get("Authorization")
	if tokenString == "" {
		return 0, fmt.Errorf("authorization header is missing")
	}

	// Remove "Bearer " prefix if present in the token string
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Parse the token and validate it using the JWT_SECRET_KEY
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	})
	if err != nil || !token.Valid {
		return 0, fmt.Errorf("invalid token: %v", err) // Return error if token is invalid
	}

	// Retrieve claims (data) from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid claims") // Error if claims aren't in the expected format
	}

	// Extract user ID from claims, assuming it's stored under "user_id"
	if userID, ok := claims["user_id"].(float64); ok {
		return int(userID), nil // Convert to int and return the user ID
	}

	return 0, fmt.Errorf("user ID not found in claims") // Error if user ID isn't found in the claims
}

// ValidateUser middleware to validate the user data.
func ValidateUser(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Validate user data using validator.v10
	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Validation failed: " + err.Error(),
		})
	}

	// If the user is valid, move to the next handler
	return c.Next()
}

// ValidateLogin middleware to validate the login data.
func ValidateLogin(c *fiber.Ctx) error {
	var loginData struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	}

	if err := c.BodyParser(&loginData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Validate login data using validator.v10
	validate := validator.New()
	if err := validate.Struct(loginData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Validation failed: " + err.Error(),
		})
	}

	// If the login data is valid, move to the next handler
	return c.Next()
}

// ValidateNote middleware to validate the note data.
func ValidateNote(c *fiber.Ctx) error {
	var note models.Note
	if err := c.BodyParser(&note); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Validate note data using validator.v10
	validate := validator.New()
	if err := validate.Struct(note); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Validation failed: " + err.Error(),
		})
	}

	// If the note data is valid, move to the next handler
	return c.Next()
}

// GetUser retrieves a user by ID from the database and returns it as JSON.
func GetUser(c *fiber.Ctx, db *gorm.DB) error {
	// Retrieve user ID from token
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Find the user in the database with the extracted userID
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to retrieve user",
		})
	}

	// Return the user data as JSON
	return c.JSON(user)
}

// UpdateUser allows a user to update their information in the database.
func UpdateUser(c *fiber.Ctx, db *gorm.DB) error {
	// Retrieve user ID from the JWT token
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Parse the request body data into updatedUser for updating the user
	var updatedUser models.User
	if err := c.BodyParser(&updatedUser); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Validate user input for any required fields
	if err := models.ValidateUser(updatedUser); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user details",
		})
	}

	// Update the user data in the database
	if err := db.Model(&models.User{}).Where("id = ?", userID).Updates(updatedUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to update user",
		})
	}

	// Return the updated user data as JSON
	return c.JSON(updatedUser)
}

// GetMe handles the /me route to retrieve user data from the context.
func GetMe(c *fiber.Ctx) error {
	// Retrieve user data from the context
	userData, ok := c.Locals("user").(models.User)

	if !ok {
		// Return error if user data is not found in context
		return c.Status(fiber.StatusInternalServerError).SendString("User data not found in context")
	}

	// Return the user data as a JSON response
	return c.JSON(userData)
}
