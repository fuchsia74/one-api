# * for amd64: docker build -t ppcelery/one-api:arm64-latest .
# * for arm64: DOCKER_BUILDKIT=1 docker build --platform linux/arm64 --build-arg TARGETARCH=arm64 -t ppcelery/one-api:arm64-latest .
FROM --platform=$TARGETPLATFORM node:24-bookworm AS builder

RUN npm install -g npm react-scripts

WORKDIR /web
COPY ./VERSION .
COPY ./web .

# Install dependencies for each project
# do not build parallel to avoid OOM on github actions
RUN cd /web/default && yarn install --network-timeout 600000
RUN cd /web/berry && yarn install --network-timeout 600000
RUN cd /web/air && yarn install --network-timeout 600000
RUN cd /web/modern && yarn install --network-timeout 600000

RUN mkdir -p /web/build

# Build the web projects
# do not build parallel to avoid OOM on github actions
RUN DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat ./VERSION) npm run build --prefix /web/default
RUN DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat ./VERSION) npm run build --prefix /web/berry
RUN DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat ./VERSION) npm run build --prefix /web/air
RUN DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat ./VERSION) npm run build --prefix /web/modern

FROM golang:1.24.6-bookworm AS builder2

# Make sure to use ARG with a default value
ARG TARGETARCH=amd64

# Set proper environment variables based on TARGETARCH
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=${TARGETARCH}

# Print architecture information for debugging
RUN echo "Building for TARGETARCH=${TARGETARCH}" && \
    echo "Current architecture: $(uname -m)"

# For ARM64 builds
RUN apt-get update && \
    if [ "${TARGETARCH}" = "arm64" ]; then \
        apt-get install -y gcc-aarch64-linux-gnu && \
        export CC=aarch64-linux-gnu-gcc && \
        export GOARCH=arm64 && \
        export CGO_ENABLED=1 && \
        # This is critical for ARM64 cross-compilation
        export CGO_CFLAGS="-g -O2 -fPIC"; \
    else \
        apt-get install -y build-essential; \
    fi

# Common dependencies
RUN apt-get install -y --no-install-recommends \
    sqlite3 libsqlite3-dev && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=builder /web/build ./web/build

# Simplified build command that handles both architectures
RUN if [ "${TARGETARCH}" = "arm64" ]; then \
        CC=aarch64-linux-gnu-gcc \
        CGO_ENABLED=1 \
        GOARCH=arm64 \
        CGO_CFLAGS="-g -O2 -fPIC" \
        go build -trimpath -ldflags "-s -w -X github.com/songquanpeng/one-api/common.Version=$(cat VERSION)" -o one-api; \
    else \
        go build -trimpath -ldflags "-s -w -X github.com/songquanpeng/one-api/common.Version=$(cat VERSION)" -o one-api; \
    fi

# Use a pre-built image that already has ffmpeg for ARM64/AMD64
FROM --platform=$TARGETPLATFORM jrottenberg/ffmpeg:6.1.2-ubuntu2404 AS ffmpeg

# Use Ubuntu as the base image which has better ARM64 support
FROM --platform=$TARGETPLATFORM ubuntu:24.04

ARG TARGETARCH=amd64
ENV DEBIAN_FRONTEND=noninteractive

# Install basic requirements without triggering libc-bin reconfiguration
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata bash haveged && \
    rm -rf /var/lib/apt/lists/*

# Copy ffmpeg binaries from the ffmpeg image
COPY --from=ffmpeg /usr/local/bin/ffmpeg /usr/local/bin/
COPY --from=ffmpeg /usr/local/bin/ffprobe /usr/local/bin/

COPY --from=builder2 /build/one-api /

# RUN if [ "${TARGETARCH}" = "arm64" ]; then \
#     else \
#         rm -rf /web/build \
#     fi

EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/one-api"]
