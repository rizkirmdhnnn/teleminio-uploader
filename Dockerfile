# Build stage
FROM golang:alpine AS builder

# Set the working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o teleminio-uploader ./cmd/bot

# Final stage
FROM alpine:3.19

# Set the working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/teleminio-uploader .

# Create necessary directories
RUN mkdir -p /app/session

# Run the application
CMD ["./teleminio-uploader"]
