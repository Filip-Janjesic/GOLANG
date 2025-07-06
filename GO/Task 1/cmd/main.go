package main

import (
	"log"
	"os"

	"zadatak-filip-janjesic/internal/db"       // Importing the db package where DB connection and functions are defined
	"zadatak-filip-janjesic/internal/handlers" // Importing handlers to manage the routes and logic for the API

	"github.com/gofiber/fiber/v2" // Importing the Fiber framework for routing and web server functionality
	"github.com/joho/godotenv"    // Importing godotenv for loading .env variables
	"gorm.io/gorm"                // Importing GORM
)

var database *gorm.DB // Global variable to hold the GORM database connection

func main() {
	// Load the environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file") // Log and exit if the .env file fails to load
	}

	// Retrieve the port value from the environment variables to start the server on a specified port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default to port 8080 if no port is set in .env
	}

	// Initialize the database connection using GORM
	var errDb error
	database, errDb = db.InitGormDB() // Initialize database connection
	if errDb != nil {
		log.Fatalf("Error connecting to the database: %v", errDb) // Log and exit if there’s an error with the DB connection
	}

	// Create a new instance of the Fiber app to set up the API routes
	app := fiber.New()

	// Middleware for validation: apply to POST and PUT routes
	app.Use("/register", handlers.ValidateUser)
	app.Use("/login", handlers.ValidateLogin)
	app.Use("/notes", handlers.ValidateNote)

	// Define the API routes by calling the defineRoutes function
	defineRoutes(app)

	// Start the HTTP server and listen on the specified port
	log.Printf("Server started on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting the server: %v", err) // Log and exit if there’s an error while starting the server
	}
}

// This function defines all the routes for the API
func defineRoutes(app *fiber.App) {
	// Set up the route for user registration (POST request to /register)
	app.Post("/register", handlers.Register(database))

	// Set up the route for user login (POST request to /login)
	app.Post("/login", handlers.Login(database)) // Pass the `database` here

	// Set up the route for retrieving user information (GET request to /me) - Protected with AuthMiddleware
	// Pass the `database` to AuthMiddleware
	app.Get("/me", handlers.AuthMiddleware(database), handlers.GetMe)

	// Set up routes for notes management
	app.Get("/notes", handlers.NotesHandler(database))  // GET request to /notes retrieves the list of notes
	app.Post("/notes", handlers.NotesHandler(database)) // POST request to /notes creates a new note

	// Set up routes for updating and deleting notes by ID
	app.Put("/notes/:id", handlers.NotesHandler(database))    // PUT request to /notes/:id updates a specific note
	app.Delete("/notes/:id", handlers.NotesHandler(database)) // DELETE request to /notes/:id deletes a specific note
}
