# Stage 1: Build the Go application
FROM golang:1.22.5 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o minio_cleanup ./cmd

# Stage 2: Create a small image with the built binary
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/minio_cleanup .

# Expose the port the app runs on (if applicable)
# EXPOSE 8080

# Command to run the application
ENTRYPOINT ["./minio_cleanup"]
