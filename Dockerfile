# ЭТАП 1: Сборка (Build)
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Собираем бинарник в специальную поддиректорию, чтобы избежать конфликта с папкой bot/
RUN mkdir -p /build && CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /build/bot ./bot


# ЭТАП 2: Финальный образ (Run)
FROM alpine:latest

RUN apk --no-cache add ca-certificates bash gnupg

WORKDIR /app

# Четко копируем скомпилированный бинарник из временной папки сборки
COPY --from=builder /build/bot /app/bot

# Запускаем приложение
CMD ["/app/bot", "--skip"]