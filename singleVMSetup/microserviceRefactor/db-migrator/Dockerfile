FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o migrator .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/migrator /app/migrator
COPY --from=builder /app/migrations /migrations

RUN apk add --no-cache postgresql-client

ENTRYPOINT ["/app/migrator"]