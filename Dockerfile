# ЕТАП 1: Сборка (Build)
FROM golang:1.25-alpine AS builder

# Встановлюємо залежності для збірки
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Копіюємо залежності окремо для ефективного кешування шарів
COPY go.mod go.sum ./
RUN go mod download

# Копіюємо решту вихідного коду
COPY . .


# Збірка бінарника
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bot ./bot

# ЕТАП 2: Фінальний образ (Run)
FROM alpine:latest

# Встановлюємо сертифікати для HTTPS (Telegram API)
RUN apk --no-cache add ca-certificates bash gnupg

# Встановлюємо робочу директорію
WORKDIR /app

# Копіюємо тільки скомпілований бінарний файл з етапу builder
COPY --from=builder /app/bot .

# Якщо вашому додатку потрібні статичні файли (шаблони, конфіги тощо),
# копіюйте їх тут. Але НЕ копіюйте .env
# COPY static/ ./static/

# Запускаємо додаток
CMD ["./bot", "--skip"]
