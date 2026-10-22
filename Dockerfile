# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build static binary
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/emby-scraper-plus ./cmd/server

# Final runtime stage
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/emby-scraper-plus /app/emby-scraper-plus

EXPOSE 8080

ENTRYPOINT ["/app/emby-scraper-plus"]
