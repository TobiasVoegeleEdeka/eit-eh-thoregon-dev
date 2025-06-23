# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Kopiere alle notwendigen Dateien
COPY mailservice/go.mod ./
RUN go mod download

COPY mailservice/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /main ./cmd/server

# Stage 2: Runtime
FROM alpine:latest
COPY --from=builder /main .
EXPOSE 8080
CMD ["./main"]