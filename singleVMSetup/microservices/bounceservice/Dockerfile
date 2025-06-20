FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY bounceservice/go.mod ./
RUN go mod download

COPY bounceservice/ ./

RUN go build -o /bounce-service .


FROM alpine:latest

WORKDIR /app


COPY --from=builder /bounce-service .

CMD ["./bounce-service"]