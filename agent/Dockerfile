FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG GO_BUILD_PARALLELISM=2
ARG GO_MEMORY_LIMIT=4GiB
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOMEMLIMIT=${GO_MEMORY_LIMIT} GOMAXPROCS=${GO_BUILD_PARALLELISM} CGO_ENABLED=0 \
    go build -p ${GO_BUILD_PARALLELISM} -trimpath -ldflags="-s -w" -o /out/ctrlb-agent ./cmd/ctrlb_collector

FROM alpine:3.23
LABEL org.opencontainers.image.source="https://github.com/ctrlb-hq/ctrlb-control-plane" \
      org.opencontainers.image.licenses="AGPL-3.0-or-later"
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/ctrlb-agent /usr/local/bin/ctrlb-agent
COPY internal/config/config.yaml /app/config.yaml
EXPOSE 3421
ENTRYPOINT ["/usr/local/bin/ctrlb-agent"]
