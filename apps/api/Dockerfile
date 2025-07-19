ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder
# Update package lists and install ca-certificates

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /brease .


FROM debian:bookworm
# Update package lists and install ca-certificates
RUN apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    apt-get update --fix-missing && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=builder /brease /usr/local/bin/
CMD ["brease"]
