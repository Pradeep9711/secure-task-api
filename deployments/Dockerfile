# Stage 1: build
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o main ./cmd/server


# Stage 2: runtime
FROM alpine:latest

RUN apk add --no-cache ca-certificates
WORKDIR /root

COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

EXPOSE 8080
CMD ["./main"]