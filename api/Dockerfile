FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=1 GOOS=linux go build -o ung-api cmd/server/main.go

# Runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/ung-api .

# Create data directory
RUN mkdir -p /root/.ung/users

# Expose port
EXPOSE 8080

# Run
CMD ["./ung-api"]
