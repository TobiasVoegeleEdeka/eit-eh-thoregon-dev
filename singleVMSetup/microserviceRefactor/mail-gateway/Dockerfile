FROM golang:1.24-alpine AS builder

WORKDIR /app


COPY go.mod ./
COPY go.sum ./
RUN go mod download


COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /mail-gateway ./main.go



FROM alpine:latest


WORKDIR /root/


COPY --from=builder /mail-gateway .


EXPOSE 8080


CMD ["./mail-gateway"]