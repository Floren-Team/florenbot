#!/bin/bash

# Определение цветов
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color (сброс цвета)

# 1. Проверка Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: docker is not installed.${NC}" >&2
    exit 1
fi
echo -e "${GREEN}Docker найден.${NC}"

# 2. Проверка Docker Compose
if command -v docker-compose &> /dev/null; then
    COMPOSE_CMD="docker-compose"
    echo -e "${GREEN}Docker Compose v1 найден ($COMPOSE_CMD).${NC}"
elif docker compose version &> /dev/null; then
    COMPOSE_CMD="docker compose"
    echo -e "${GREEN}Docker Compose v2 (плагин) найден ($COMPOSE_CMD).${NC}"
else
    echo -e "${RED}Error: docker-compose is not installed.${NC}" >&2
    exit 1
fi

# Проверка наличия файла docker-compose.yml
if [ -f "docker-compose.yml" ]; then
    echo -e "${GREEN}docker-compose.yml найден!${NC}"
else
    echo -e "${RED}Error: docker-compose.yml not found.${NC}" >&2
    exit 1
fi

# Проверка Dockerfile
if [ -f "Dockerfile" ]; then
    echo -e "${GREEN}Dockerfile найден!${NC}"
else
    echo -e "${RED}Error: Dockerfile not found.${NC}" >&2
    exit 1
fi

sleep 2
echo -e "${GREEN}Запуск сборки проекта..."
sleep 3
$COMPOSE_CMD up -d --build

if [ $? -eq 0 ]; then
    echo -e "${GREEN}Сборка успешно завершена!${NC}"
else
    echo -e "${RED}Ошибка во время сборки.${NC}" >&2
    exit 1
fi