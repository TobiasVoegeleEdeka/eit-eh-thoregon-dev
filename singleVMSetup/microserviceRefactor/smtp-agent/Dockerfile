FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /smtp-worker .


FROM alpine:latest

WORKDIR /root/


COPY --from=builder /smtp-worker .


CMD ["./smtp-worker"]
