# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

# Install git for go mod download (some modules need it)
RUN apk --no-cache add git

WORKDIR /app

# Copy dependency files first — Docker layer cache means this only re-runs
# when go.mod or go.sum changes, not on every code change.
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
    -trimpath \
    -o /ecg-backend \
    ./cmd/api

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM alpine:3.19

# ca-certificates: for HTTPS calls (LDAP over TLS etc.)
# tzdata: so time.LoadLocation works correctly in Ghana timezone
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S ecg && adduser -S ecg -G ecg

WORKDIR /app

COPY --from=builder /ecg-backend .

# Run as non-root
USER ecg

EXPOSE 9400

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:9400/health || exit 1

ENTRYPOINT ["/app/ecg-backend"]