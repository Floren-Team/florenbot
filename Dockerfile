FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bot ./bot


FROM alpine:latest

RUN apk --no-cache add ca-certificates bash gnupg

WORKDIR /app

COPY --from=builder /app/bot /app/bot

CMD ["./bot", "--skip"]