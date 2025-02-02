FROM golang:1.23 AS builder

RUN apt-get update && apt-get install -y \
    cmake \
    make \
    gcc \
    git \
    curl \
    unzip \
    flex \
    bison \
    libyaml-dev \
    libssl-dev \
    libc-ares-dev

WORKDIR /app

RUN git clone https://github.com/ctrlb-hq/ctrlb-fluent-bit
WORKDIR /app/ctrlb-fluent-bit/build
RUN cmake -DFLB_RELEASE=On -DFLB_TRACE=Off -DFLB_CONFIG_YAML=ON -DFLB_JEMALLOC=Off -DFLB_TLS=On -DFLB_SHARED_LIB=On -DFLB_EXAMPLES=Off -DFLB_HTTP_SERVER=On -DFLB_IN_SYSTEMD=On -DFLB_OUT_KAFKA=On -DFLB_ALL=On ../
RUN make

WORKDIR /app

ENV PKG_CONFIG_PATH=/app/ctrlb-fluent-bit/build/lib/pkgconfig
ENV CGO_CFLAGS="-I/app/ctrlb-fluent-bit/lib/monkey/include -I/app/ctrlb-fluent-bit/include"
ENV LD_LIBRARY_PATH=/app/ctrlb-fluent-bit/build/lib

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o ctrlb-collector cmd/ctrlb_collector/main.go
ENTRYPOINT ["/app/ctrlb-collector"]
