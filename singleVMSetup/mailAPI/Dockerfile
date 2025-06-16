# Stage 1: Build the Go application
FROM golang:1.22-alpine AS builder
# Uses golang:1.22-alpine for a lightweight build environment.

WORKDIR /app

# Copy go.mod to download dependencies
# go.sum is not copied as it does not exist in this setup
COPY go.mod ./
RUN go mod download

# Copy Go source files from the build context root
# Assumes all .go files (e.g., main.go) are in the root directory
COPY *.go ./
# Alternative: Use `COPY . .` if code is in subdirectories

# Build the application
# CGO_ENABLED=0 ensures a statically linked binary
# -ldflags="-s -w" strips debug symbols for a smaller binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-s -w" -o /main .

# Stage 2: Create the minimal runtime image
FROM alpine:latest
# Uses alpine:latest for a very small runtime image

WORKDIR /app
# Sets working directory to /app for consistency

# Copy the binary from the builder stage
COPY --from=builder /main .

# Expose port 8080 (matches LISTEN_PORT in docker-compose.yaml)
EXPOSE 8080

# Run the application
CMD ["./main"]