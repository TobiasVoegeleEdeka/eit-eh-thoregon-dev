# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project (mit der neuen Struktur)
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-s -w" -o /main ./cmd/server

# Stage 2: Create the minimal runtime image
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /main .

# Copy configuration files if any
# COPY configs/ ./configs/

EXPOSE 8080

CMD ["./main"]