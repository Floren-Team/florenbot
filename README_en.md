![License](https://img.shields.io/badge/license-GPL--3.0-blue)
![Language](https://img.shields.io/badge/language-Go-00ADD8)
![Telegram](https://img.shields.io/badge/platform-Telegram-26A5E4)

### Game Bot in Go

What is this bot?

This is an interactive game bot designed for group chat entertainment. It is fully Open Source and can be used anywhere.

Why are we different from other bots?

1. Regular updates: We release updates more frequently to keep things fresh and exciting.
2. Comfortable and fun bot :)

## About the Bot

This project is a high-performance game bot written in Go, utilizing MySQL for persistent data storage and Redis for caching and real-time state management.

## Architecture:

    Language: Go (Golang) — ensures high request processing speed and efficient concurrency.

    Database (MySQL): Stores player profiles.

    Cache (Redis): Stores sessions.

## Key Features

    Authorization System: Fast data verification via Redis with subsequent synchronization to MySQL.

    Game Loop: Real-time event processing.

    Scalability: Uses goroutines to handle thousands of concurrent connections.

    Optimization: Minimizes SQL load by caching frequent queries in Redis.

## System Requirements

    Go 1.26+

    MySQL 8.0/MariaDB 11.0+

    Redis 7.0+

    Docker

## Installation and Setup

1. Clone the repository

 ```bash
   git clone <repository-link> florenbot
   cd florenbot
 ```

2. **Environment Configuration:**
   Copy `.env.docker.example` to `.env` and specify your database connection parameters:
   ```env
    REDIS_ADDR=localhost:6379
    REDIS_PASSWORD=
    REDIS_DB=0
    
    DB_HOST=localhost
    DB_PORT=3306
    DB_USER=
    DB_PASSWORD=
    DB_NAME=
    DB_ROOT_PASSWORD=
   ```

## 3. Run the Project

```bash
    docker compose up -d
```

## Important!

Don't forget about the **BOT_TOKEN** — you need to get it from @BotFather, which issues the bot token, and paste it there.

## Build Instructions for Debian 13:

https://github.com/FlorenBot/florenbot/blob/main/BUILDING.md
