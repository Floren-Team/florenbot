# ЕТАП 1: Сборка (Build)
FROM golang:1.25-alpine AS builder

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Копируем зависимости отдельно для эффективного кэширования слоев
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной исходный код
COPY . .

# Сборка бинарника (с включенным CGO, так как ранее были добавлены gcc/musl-dev)
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bot ./bot


# ЭТАП 2: Финальный образ (Run)
FROM alpine:latest

# Устанавливаем сертификаты для HTTPS (Telegram API) и другие нужные утилиты
RUN apk --no-cache add ca-certificates bash gnupg

# Устанавливаем рабочую директорию
WORKDIR /app

# 1. Сначала копируем все файлы (конфиги, статику и т.д.), кроме того, что исключено через .dockerignore
COPY --from=builder /app/ ./

# 2. Затем поверх копируем готовый скомпилированный бинарник
COPY --from=builder /app/bot .

# Запускаем приложение
CMD ["./bot", "--skip"]