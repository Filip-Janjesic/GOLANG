# Stage 1: Build the Go application
FROM golang:1.23-alpine AS builder

# Install dependencies for Go C bindings
RUN apk add --no-cache gcc musl-dev

# Set working directory to root, so paths are simpler
WORKDIR /

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN CGO_ENABLED=1 go build -o zadatak-filip-janjesic-app ./cmd/main.go

# Stage 2: Run the Go application
FROM alpine:latest

# Copy the compiled binary and .env file from the builder stage
COPY --from=builder /zadatak-filip-janjesic-app .
COPY --from=builder /.env .env

# Make the binary executable
RUN chmod +x ./zadatak-filip-janjesic-app

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./zadatak-filip-janjesic-app"]
