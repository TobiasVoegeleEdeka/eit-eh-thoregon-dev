# Stage 1: Build the Go application
FROM golang:1.22-alpine AS builder
# Using golang:alpine for a smaller build image. Adjust Go version if needed.

WORKDIR /app

# Copy go.mod to download dependencies
# (go.sum wird nicht kopiert, da es in Ihrem Fall nicht existiert)
COPY go.mod ./
RUN go mod download

# Copy the source code
# Dies kopiert alle .go-Dateien aus dem Hauptverzeichnis des Build-Kontexts
COPY *.go ./
# Wenn Ihr Code in Unterverzeichnissen liegt, verwenden Sie stattdessen:
# COPY . .

# Build the application.
# CGO_ENABLED=0 produces a statically linked binary.
# -ldflags="-s -w"` strips debug symbols and makes the binary smaller.
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-s -w" -o /main .

# Stage 2: Create the final lightweight image
FROM alpine:latest
# Using alpine:latest for a very small runtime image.

WORKDIR /root/ # Oder ein anderes Verzeichnis wie /app, wenn Sie es bevorzugen

# Copy the pre-built binary from the builder stage
COPY --from=builder /main .

# Expose port 8080 (oder den Port, den Ihre Anwendung verwendet)
EXPOSE 8080

# Command to run the executable
CMD ["./main"]