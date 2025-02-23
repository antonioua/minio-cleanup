# Stage 1: Build the Go application
FROM golang:1.22.5 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules and the source files
COPY go.mod go.sum ./
COPY main.go main.go
COPY cmd/ cmd/

# Download dependencies
RUN go mod download

# Build the Go application for the target platform
RUN CGO_ENABLED=0 GO111MODULE=on \
    go build \
    -a \
    -o /minio_cleanup

RUN chmod +x /minio_cleanup

# Stage 2: Create the final image
FROM alpine:latest

WORKDIR /

COPY --from=builder /minio_cleanup /minio_cleanup

ENTRYPOINT ["/minio_cleanup"]
CMD ["-h"]
