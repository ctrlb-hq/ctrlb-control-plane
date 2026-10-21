FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ctrlb-agent ./cmd/ctrlb_collector

FROM alpine:3.23
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/ctrlb-agent /usr/local/bin/ctrlb-agent
COPY internal/config/config.yaml /app/config.yaml
EXPOSE 3421
ENTRYPOINT ["/usr/local/bin/ctrlb-agent"]
