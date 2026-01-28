FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bot ./cmd/bot


FROM alpine:3.21

RUN apk add --no-cache tzdata ca-certificates

ENV TZ=America/Lima
ENV DISCORD_SCHEDULE_PATH=/app/config/schedules.json

WORKDIR /app

COPY --from=builder /app/bot .
COPY --from=builder /app/config ./config

ENTRYPOINT ["./bot"]
