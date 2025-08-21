FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o bin/server cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/bin/server .

# Create data directory
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Set environment variables
ENV DB_PATH=/data/continuity.db

# Run the binary
CMD ["./server"]