# ── Build stage ──────────────────────────────────────────────────────────────
FROM docker.io/library/golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Download dependencies first (layer caching)
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o crowdspeech-backend .

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM docker.io/library/alpine:3.19

# Install ffmpeg for audio processing
RUN apk add --no-cache ffmpeg ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/crowdspeech-backend .

# PocketBase data directory
RUN mkdir -p /app/pb_data

VOLUME ["/app/pb_data"]

EXPOSE 10001

ENV PB_PORT=10001
ENV PB_DATA_DIR=/app/pb_data

CMD ["./crowdspeech-backend", "serve", "--http=0.0.0.0:10001", "--dir=/app/pb_data"]
