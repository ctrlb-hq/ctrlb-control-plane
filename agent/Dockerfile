# Use the official Golang image to build the binary.
FROM golang:1.23-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code.
COPY . .

# Build the Go app. Replace "myapp" with your binary name.
RUN go build -o ctrlb-agent ./cmd/ctrlb_collector/main.go

# Use a minimal base image for deployment.
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage.
COPY --from=builder /app/ctrlb-agent .
COPY --from=builder /app/internal/config/config.yaml .

# Command to run the executable
CMD ["./ctrlb-agent"]