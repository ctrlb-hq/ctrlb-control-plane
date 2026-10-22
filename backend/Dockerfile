FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=1
RUN go build -trimpath -ldflags="-s -w" -o /out/backend ./cmd/backend

FROM alpine:3.24

LABEL org.opencontainers.image.source="https://github.com/ctrlb-hq/ctrlb-control-plane" \
      org.opencontainers.image.licenses="AGPL-3.0-only"

RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 ctrlb \
    && adduser -S -D -H -u 10001 -G ctrlb ctrlb \
    && mkdir /data \
    && chown ctrlb:ctrlb /data

COPY --from=builder /out/backend /usr/local/bin/backend

# The application stores backend.db and app.log in its working directory.
WORKDIR /data
USER 10001:10001

ENV PORT=8096

EXPOSE 8096

ENTRYPOINT ["/usr/local/bin/backend"]
