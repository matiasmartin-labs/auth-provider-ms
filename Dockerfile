# syntax=docker/dockerfile:1

# ---------------------------------------------------------------------------
# Stage 1: builder
# ---------------------------------------------------------------------------
FROM golang:1.25-alpine AS builder

ARG GOARCH
ARG GOOS=linux

WORKDIR /build

# Copy dependency manifests first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy full source tree
COPY . .

# Build a static binary with debug info stripped
RUN CGO_ENABLED=0 GOOS=${GOOS} go build \
  -ldflags="-s -w" \
  -o /app/provider-auth-ms \
  ./cmd/provider-auth-ms

# ---------------------------------------------------------------------------
# Stage 2: runtime
# ---------------------------------------------------------------------------
FROM alpine:3.21 AS runtime

# CA certificates for HTTPS calls to Google APIs
RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/provider-auth-ms /app/provider-auth-ms
COPY cmd/provider-auth-ms/config.yaml /app/config.yaml

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/provider-auth-ms"]
