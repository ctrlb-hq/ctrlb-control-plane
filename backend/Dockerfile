# Start from the official Go image
FROM golang:1.23

# Set the working directory
WORKDIR /app

# Copy the Go modules files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN go build -o backend cmd/backend/main.go 

# Expose port 8096
EXPOSE 8096

# Run the Go application
CMD ["./backend"]
