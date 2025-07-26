ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder
# Update package lists and install ca-certificates

WORKDIR /usr/src/app
RUN go env -w GOMODCACHE=/root/.cache/go-build
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build go mod download && go mod verify
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build go build -v -o /brease .

FROM debian:bookworm
# Update package lists and install ca-certificates and Infisical CLI
RUN apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    apt-get update --fix-missing && apt-get install -y \
    ca-certificates \
    curl \
    bash \
    && curl -1sLf 'https://artifacts-cli.infisical.com/setup.deb.sh' | bash \
    && apt-get update && apt-get install -y infisical \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /brease /usr/local/bin/
COPY start.sh /usr/local/bin/start.sh
RUN chmod +x /usr/local/bin/start.sh

CMD ["/usr/local/bin/start.sh"]
