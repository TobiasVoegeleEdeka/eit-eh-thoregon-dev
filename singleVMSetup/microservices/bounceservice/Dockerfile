FROM golang:1.24-alpine AS builder
WORKDIR /app


COPY go.mod ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bounceservice ./bounceservice/cmd/server

FROM alpine:latest
WORKDIR /app

COPY --from=builder /bounceservice /app/bounceservice
COPY --from=builder /app/data /app/data

EXPOSE 8081
CMD ["/app/bounceservice"]