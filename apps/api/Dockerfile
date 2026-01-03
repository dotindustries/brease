ARG GO_VERSION=1.24
FROM golang:${GO_VERSION}-bookworm AS builder

WORKDIR /usr/src/app

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies with proper cache mounts
# Separate cache directories for modules and build cache
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download && go mod verify

# Copy source code
COPY . .

# Build with cache mounts and optimizations
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -v -ldflags="-w -s" -o /brease .

# Runtime stage - using distroless for minimal size and better security
FROM gcr.io/distroless/static-debian12:nonroot AS runtime-distroless

COPY --from=builder /brease /usr/local/bin/brease

# Distroless doesn't support shell scripts, so we need a different approach
# If start.sh is essential, use the debian variant below instead
CMD ["/usr/local/bin/brease"]

# Alternative runtime with Infisical support (use this if you need start.sh)
FROM debian:bookworm-slim AS runtime-debian

# Install dependencies in a single layer with cleanup
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    bash \
    && curl -1sLf 'https://artifacts-cli.infisical.com/setup.deb.sh' | bash \
    && apt-get update && apt-get install -y --no-install-recommends infisical \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -r -u 65532 -g root nonroot

COPY --from=builder /brease /usr/local/bin/brease
COPY start.sh /usr/local/bin/start.sh
RUN chmod +x /usr/local/bin/start.sh /usr/local/bin/brease

USER nonroot

CMD ["/usr/local/bin/start.sh"]

# Final stage selector - uncomment the one you want to use
FROM runtime-debian AS final
