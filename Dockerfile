FROM golang:1.24.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /random-coffee-bot

FROM alpine:latest

WORKDIR /app

COPY --from=builder /random-coffee-bot /app/random-coffee-bot
COPY .env /app/.env

CMD ["/app/random-coffee-bot"]