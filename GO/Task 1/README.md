
## Note-Taking Application API

### Task Description

**Objective**: Create a RESTful API in Go for a simple note-taking application with user authentication.

### Endpoints

1. **POST /register**: Register a new user.
2. **POST /login**: Authenticate a user and return a token.
3. **GET /notes**: Retrieve all notes for the authenticated user.
4. **POST /notes**: Create a new note for the authenticated user.
5. **PUT /notes/{id}**: Update an existing note by ID for the authenticated user.
6. **DELETE /notes/{id}**: Delete a note by ID for the authenticated user.
7. **GET /me**: Retrieve information about the authenticated user.

### Data Model

**User**:
```go
type User struct {
    ID       int        `json:"id"`
    Username string     `json:"username"`
    Password string     `json:"password"`
    FirstName string    `json:"first_name,omitempty"`
    LastName  string    `json:"last_name,omitempty"`
    Email     string    `json:"email,omitempty"`
    Phone     string    `json:"phone,omitempty"`
    DateOfBirth string  `json:"date_of_birth,omitempty"`
    City      string    `json:"city,omitempty"`
    Country   string    `json:"country,omitempty"`
}
```

**Note**:
```go
type Note struct {
    ID        int           `json:"id"`
    UserID    int           `json:"user_id"`
    Title     string        `json:"title"`
    Body      string        `json:"body"`
    CreatedAt time.Time     `json:"created_at"`
    UpdatedAt time.Time     `json:"updated_at"`
    DeletedAt *time.Time    `json:"deleted_at,omitempty"`
}
```

### Project Structure

1. **cmd/**: Contains the main application entry point (`main.go`), where HTTP routes are set up for handling registration, login, and note management. Authentication is managed through JWT middleware.
   
2. **internal/handlers/**: Contains handlers for each API endpoint:
   - **Register**: Handles user registration.
   - **Login**: Handles user authentication and JWT token generation.
   - **NotesHandler**: Handles CRUD operations for notes (create, read, update, delete).
   - **ContextHandler**: Handles the new `/me` endpoint to retrieve user information from context.

3. **internal/models/**: Contains data models (User and Note) used across the application.

4. **internal/cache/**:  Includes caching functions like `SaveNotesToCache` and `LoadNotesFromCache` to optimize performance.

5. **.env**: Stores sensitive configuration variables such as database connection strings, JWT secrets, etc.

### How the Project Works

1. The API is designed to register users, authenticate them via JWT tokens, and manage note-taking functionalities.

2. The application uses the Fiber framework for handling HTTP requests. User passwords are hashed using bcrypt, and JWT tokens are used for authentication, ensuring that only authorized users can access their notes.
   
3. Data is stored in an SQLite database, which is managed via GORM, a Go ORM library. Auto-migration is used to automatically update the database schema according to the GORM models.

4. The application now uses GORM queries for all database interactions, removing the need for raw SQL. This makes the application easier to maintain and scale.

5. Transactions have been added for both update and delete operations to ensure atomicity and consistency. These transactions are used to ensure that operations are either fully completed or fully rolled back in case of an error.

6. Local caching is implemented externally, allowing for improved performance without relying on Docker for storage.

7. The application provides clear error messages to inform users about any issues that arise during API interactions, including problems with registration, authentication, and note management.

8. Update and delete operations include thorough checks to verify the existence of the specified ID before proceeding, preventing unnecessary errors.

9. Comments are added to functions throughout the codebase to enhance readability and maintainability.

### Additional Implementation Guidelines

1. **Use `.env`**: Ensure sensitive configuration is stored in an `.env` file.
2. **User Registration**: Update the registration API to accept personal and address information.
3. **Database Timestamps**: Implement `created_at`, `updated_at`, and `deleted_at` fields in the database.
4. **GORM Models**: All models now use GORM for database interactions, making queries more intuitive and manageable.
5. **Auto-Migration**: GORM’s `AutoMigrate` function ensures that the database schema is automatically synchronized with the Go models.
6. **SQL Queries**: Separate SQL queries into a dedicated file for better organization.
7. **Transactions**: For the update and delete operations, transactions have been implemented to ensure that changes are applied atomically.
8. **Caching**: Implement cache functions to improve performance.
9. **Parameter Validation**: Implement validation for parameters to ensure proper data types (e.g., integer IDs and text lengths).
10. **Indexing**: Consider indexing `deleted_at` for improved performance.

### Developer Notes

1. **Investigate how the timestamp is automatically set for `created_at` and `updated_at`.**
    In GORM, these fields are typically set automatically when using `Create` or `Save`. 
    You can use GORM’s `BeforeCreate` and `BeforeUpdate` hooks or rely on GORM’s auto timestamp behavior to handle `created_at` and `updated_at`.

2. **Put a constraint for the necessary parameters in the DB via GORM.**
    Ensure that constraints like NOT NULL, UNIQUE, and others are applied to fields where necessary (e.g., `Username` or `Email`).

3. **Set the foreign key where necessary.**
    When defining models, ensure that the `UserID` in the `Note` model is linked to the `User` model via a foreign key. GORM allows you to define relationships easily, and you can use `gorm:"foreignKey:UserID"` in your `Note` model.

4. **ORM has the ability to filter items that are not searchable (without having to do it manually).**
    In GORM, you can use the `Scopes` feature or apply filters to exclude non-searchable fields from queries.

5. **Insert github.com/go-playground/validator/v10 for all validations.**
    Use the `validator` package to validate the input data in your handlers.
    This ensures that all fields are validated before interacting with the database.

### CURL Commands

1. **Register a New User**
```bash
curl -X POST http://localhost:8080/register \
-d '{"username": "exampleUser", "password": "examplePassword"}' \
-H "Content-Type: application/json" | json_pp
```

2. **Login and Get Token**
```bash
curl -X POST http://localhost:8080/login \
-d '{"username": "exampleUser", "password": "examplePassword"}' \
-H "Content-Type: application/json" | json_pp
```

3. **Get All Notes (requires token)**
```bash
curl -X GET http://localhost:8080/notes \
-H "Authorization: Bearer <token>" | json_pp
```

4. **Create a New Note (requires token)**
```bash
curl -X POST http://localhost:8080/notes \
-d '{"title": "New Note", "body": "Note content"}' \
-H "Content-Type: application/json" \
-H "Authorization: Bearer <token>" | json_pp
```

5. **Update a Note (requires token)**
```bash
curl -X PUT http://localhost:8080/notes/1 \
-d '{"title": "Updated Title", "body": "Updated body"}' \
-H "Content-Type: application/json" \
-H "Authorization: Bearer <token>" | json_pp
```

6. **Delete a Note (requires token)**
```bash
curl -X DELETE http://localhost:8080/notes/1 \
-H "Authorization: Bearer <token>"
```

7. **GET /me**
    - Retrieve information about the authenticated user from context.