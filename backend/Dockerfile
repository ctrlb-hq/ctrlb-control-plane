# Stage 1: Build
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Enable CGO explicitly
ENV CGO_ENABLED=1
RUN go build -o backend cmd/backend/main.go

# Stage 2: Minimal image with sqlite support
FROM alpine:latest

RUN apk --no-cache add ca-certificates libc6-compat sqlite

COPY --from=builder /app/backend /backend

EXPOSE 8096

ENTRYPOINT ["/backend"]
