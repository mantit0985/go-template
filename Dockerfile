# syntax=docker/dockerfile:1

# STAGE 1: Builder
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main src/main.go

# --- RUNTIME TARGETS ---

# Target: runtime-slim (The "Debug" image)
FROM alpine:3.22 AS runtime-slim
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/main .
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot
USER nonroot
CMD ["./main"]

# Target: runtime-alpine (The "IoT/Edge" image)
FROM alpine:3.22 AS runtime-alpine
WORKDIR /app
COPY --from=builder /app/main .
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot
USER nonroot
CMD ["./main"]

# Target: runtime-distroless (The "Hardened" image)
FROM gcr.io/distroless/static-debian12 AS runtime-distroless
WORKDIR /app
COPY --from=builder /app/main .
USER nonroot
CMD ["./main"]

# fixed groupadd change
# Default target for smoke tests
FROM alpine:3.22 AS default
WORKDIR /app
COPY --from=builder /app/main .
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot
USER nonroot
CMD ["./main"]
