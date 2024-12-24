FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o bridge-app ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata postgresql-client

# Copy binary from builder
COPY --from=builder /app/bridge-app .
COPY --from=builder /app/database/migrations ./database/migrations

# Copy config files
COPY config/config.yaml ./config/
COPY scripts/wait-for-postgres.sh .
RUN chmod +x wait-for-postgres.sh

EXPOSE 8080

CMD ["./bridge-app"]