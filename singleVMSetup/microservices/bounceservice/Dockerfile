FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY bounceservice/go.mod  ./
RUN go mod download

COPY bounceservice/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-s -w" -o /bounce-service-app .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /bounce-service-app .

CMD ["./bounce-service-app"]
