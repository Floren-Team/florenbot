# ЭТАП 1: Сборка (Build)
FROM golang:1.25-alpine AS builder

# Устанавливаем необходимые зависимости для сборки (например, gcc для cgo)
RUN apk add --no-cache git gcc musl-dev

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и ***REMOVED*** для кэширования зависимостей
COPY go.mod ***REMOVED*** ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарный файл
# -ldflags="-s -w" убирает отладочную информацию (уменьшает размер)
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o bot .

# ЭТАП 2: Финальный образ (Run)
FROM alpine:latest

# Устанавливаем сертификаты (нужны для связи с Telegram API)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем только скомпилированный бинарный файл из этапа builder
COPY --from=builder /app/bot .
# Если есть файлы .env, их тоже нужно скопировать
COPY .env .

# Запускаем приложение
CMD ["./bot"]