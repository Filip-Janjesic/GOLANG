package handlers

import (
	"fmt"
	"log"
	"os"
	"time"
	"zadatak-filip-janjesic/internal/models" // Import your models

	"github.com/go-playground/validator/v10" // Validator for input validation
	"github.com/gofiber/fiber/v2"            // Fiber for HTTP handling
	"github.com/golang-jwt/jwt/v4"           // JWT for token creation
	"github.com/joho/godotenv"               // Load environment variables
	"golang.org/x/crypto/bcrypt"             // Password hashing
	"gorm.io/gorm"                           // Database operations
)

// Initialize JWT secret key from environment variables
var secretKey []byte

// Initialize environment variables
func init() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file") // Log error if .env fails to load
	}
	secretKey = []byte(os.Getenv("JWT_SECRET"))
}

// Register handles user registration by validating and creating new user entries
func Register(database *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user models.User // User struct to hold incoming data

		// Parse JSON request body into user struct; return error if parsing fails
		if err := c.BodyParser(&user); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid input")
		}

		// Validate the user data
		if err := models.ValidateUser(user); err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				validationErrorsMap := make(map[string]string)
				for _, validationErr := range validationErrors {
					validationErrorsMap[validationErr.Field()] = validationErr.Tag()
				}
				return c.Status(fiber.StatusBadRequest).JSON(validationErrorsMap)
			}
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		// Log the decoded user data for debugging purposes
		log.Printf("Decoded user: %+v", user)

		// Check for existing username in the database
		var existingUser models.User
		err := database.Where("username = ?", user.Username).First(&existingUser).Error
		if err == nil {
			return c.Status(fiber.StatusBadRequest).SendString("Username already exists")
		}
		if err != gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusInternalServerError).SendString("Database error")
		}

		// Hash password before saving user data
		hashedPassword, err := hashPassword(user.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error while hashing password")
		}
		user.Password = hashedPassword

		// Create new user in the database
		if err := database.Create(&user).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Unable to register user")
		}

		// Return the created user data as JSON
		return c.Status(fiber.StatusCreated).JSON(user)
	}
}

// Login handles user login requests by verifying credentials and generating a JWT if successful
func Login(database *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var loginData models.User

		// Parse JSON request body into loginData struct
		if err := c.BodyParser(&loginData); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
		}

		// Retrieve user data from database based on provided username
		var user models.User
		if err := database.Where("username = ?", loginData.Username).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid username or password"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error retrieving user data"})
		}

		// Compare hashed password from DB with provided password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid username or password"})
		}

		// Generate JWT token upon successful login
		token, err := generateToken(user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error generating token"})
		}

		// Send the generated token in the response
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"token": token})
	}
}

// AuthMiddleware validates JWT tokens for protected routes
func AuthMiddleware(database *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Retrieve token from Authorization header
		tokenString := c.Get("Authorization")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authorization token required"})
		}

		// Parse and validate JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})

		// Return error if parsing or validation fails
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}

		// Extract user ID from the token claims
		claims := token.Claims.(jwt.MapClaims)
		userID := int(claims["user_id"].(float64))

		// Check if the user exists in the database
		var user models.User
		if err := database.First(&user, userID).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
		}

		// Store user data in the context for use in handlers
		c.Locals("user", user)

		return c.Next()
	}
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}
	return string(hashedPassword), nil
}

// generateToken generates a JWT token for the given user
func generateToken(user models.User) (string, error) {
	// Define JWT claims, including the user ID and expiration time
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	// Create the JWT token using the claims and secret key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("error generating token: %w", err)
	}

	return tokenString, nil
}
