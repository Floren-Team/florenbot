package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

// Функция для отладки (дебаггер)
func Debug(format string, v ...interface{}) {
	if os.Getenv("DEBUG") == "true" {
		log.Printf("[DEBUG] "+format, v...)
	}
}

func InitDB() {
	var err error
	dbMode := getEnv("DB_MODE", "mysql")
	Debug("Режим БД: %s", dbMode)

	if dbMode == "mysql" {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			getEnv("DB_USER", "root"),
			getEnv("DB_PASSWORD", ""),
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "3306"),
			getEnv("DB_NAME", "game_db"),
		)
		DB, err = sql.Open("mysql", dsn)
	} else {
		log.Print("Использую локальную БД (SQLite)...")

		if _, err := os.Stat("db"); os.IsNotExist(err) {
			log.Println("Директория 'db' не найдена, создаю её...")
			if err := os.Mkdir("db", 0755); err != nil && !os.IsExist(err) {
				log.Printf("Ошибка при создании db: %v", err)
			}
		}
		var err error
		DB, err = sql.Open("sqlite", "bot.db")
		if err != nil {
			log.Fatalf("Ошибка при открытии SQLite: %v", err)
		}

		_, err = DB.Exec("PRAGMA journal_mode=WAL;")
		if err != nil {
			log.Printf("⚠️ Не удалось включить WAL-режим: %v", err)
		} else {
			log.Println("✅ WAL-режим включен")
		}
	}

	if err != nil {
		log.Fatalf("❌ Ошибка при открытии БД: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("❌ Не удалось подключиться к БД: %v", err)
	}


	var queries []string


	if dbMode == "mysql" {
		queries = []string{
			// 1. Таблицы без внешних ключей
			`CREATE TABLE IF NOT EXISTS clans (
				id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, 
				name VARCHAR(255) NOT NULL, 
				owner_id BIGINT NOT NULL, 
				owner_name VARCHAR(32) NOT NULL, 
				invite_code VARCHAR(32) NULL UNIQUE,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			) ENGINE=InnoDB;`,

			`CREATE TABLE IF NOT EXISTS users (
				id BIGINT PRIMARY KEY, 
				username VARCHAR(255), 
				balance FLOAT DEFAULT 1000,
				first_name VARCHAR(70), 
				promocode VARCHAR(32),
				role VARCHAR(32) DEFAULT 'user', 
				floren_coin FLOAT DEFAULT 300000,
				negative_reputation INT DEFAULT 0, 
				positive_reputation INT DEFAULT 0, 
				euro FLOAT DEFAULT 17800,
				wins INT DEFAULT 0,
				losses INT DEFAULT 0,
				clan_id INT DEFAULT NULL, 
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			) ENGINE=InnoDB;`,

			`CREATE TABLE IF NOT EXISTS promocodes (
				id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, 
				code VARCHAR(255) NOT NULL UNIQUE, 
				owner_id BIGINT NOT NULL UNIQUE,
				amount INT NOT NULL, 
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			) ENGINE=InnoDB;`,

			`CREATE TABLE IF NOT EXISTS chats (
				id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
				chat_name VARCHAR(255) NOT NULL,
				user_id BIGINT NOT NULL UNIQUE,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			) ENGINE=InnoDB;`,

			`CREATE TABLE IF NOT EXISTS reports (
				id INT AUTO_INCREMENT PRIMARY KEY,
				user_id BIGINT NULL,
				text TEXT NOT NULL,
				active BOOLEAN DEFAULT TRUE,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			) ENGINE=InnoDB;`,

			`CREATE TABLE IF NOT EXISTS blacklists (
				id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, 
				user_id BIGINT NOT NULL, 
				reason TEXT NOT NULL, 
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			) ENGINE=InnoDB;`,

			`CREATE TABLE IF NOT EXISTS clans_members (
			clan_id INT NOT NULL,
			user_id BIGINT NOT NULL,
			role VARCHAR(20) DEFAULT 'member',
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (clan_id, user_id)
		) ENGINE=InnoDB;`,

		`CREATE TABLE IF NOT EXISTS clans_blacklist (
			clan_id INT NOT NULL,
			user_id BIGINT NOT NULL,
			reason TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (clan_id, user_id)
		) ENGINE=InnoDB;`,

		// Додаємо зв'язки для нових таблиць
		`ALTER TABLE clans_members ADD FOREIGN KEY (clan_id) REFERENCES clans(id) ON DELETE CASCADE;`,
		`ALTER TABLE clans_members ADD FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;`,
		`ALTER TABLE clans_blacklist ADD FOREIGN KEY (clan_id) REFERENCES clans(id) ON DELETE CASCADE;`,
		`ALTER TABLE clans_blacklist ADD FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;`,

			// 2. Добавление связей через ALTER TABLE
			`ALTER TABLE promocodes ADD FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE;`,
			`ALTER TABLE chats ADD FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;`,
			`ALTER TABLE users ADD FOREIGN KEY (clan_id) REFERENCES clans(id) ON DELETE SET NULL;`,
			`ALTER TABLE reports ADD FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;`,
			`ALTER TABLE blacklists ADD FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;`,
		}
	} else {
		// SQLite версия
		queries = []string{
			`CREATE TABLE IF NOT EXISTS clans (
				id INTEGER PRIMARY KEY AUTOINCREMENT, 
				name VARCHAR(255) NOT NULL, 
				owner_id BIGINT NOT NULL,
				invite_code VARCHAR(32) UNIQUE,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`,

			`CREATE TABLE IF NOT EXISTS users (
				id BIGINT PRIMARY KEY, 
				username VARCHAR(255), 
				balance INT DEFAULT 1000, 
				promocode VARCHAR(32), 
				floren_coin FLOAT DEFAULT 300000,
				euro FLOAT DEFAULT 17800,
				wins INT DEFAULT 0,
				losses INT DEFAULT 0,
				role VARCHAR(32) DEFAULT 'user',
				first_name VARCHAR(70),
				clan_id INTEGER, 
				negative_reputation INT DEFAULT 0, 
				positive_reputation INT DEFAULT 0, 
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY(clan_id) REFERENCES clans(id) ON DELETE SET NULL
			);`,

			`CREATE TABLE IF NOT EXISTS promocodes (
				id INTEGER PRIMARY KEY AUTOINCREMENT, 
				code VARCHAR(255) NOT NULL UNIQUE, 
				owner_id BIGINT NOT NULL UNIQUE,
				amount INT NOT NULL, 
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
			);`,

			`CREATE TABLE IF NOT EXISTS chats (
				id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
				chat_name VARCHAR(255) NOT NULL,
				user_id BIGINT NOT NULL UNIQUE,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);`,

			`CREATE TABLE IF NOT EXISTS clans_members (
				clan_id INTEGER NOT NULL,
				user_id BIGINT NOT NULL,
				role VARCHAR(20) DEFAULT 'member',
				joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (clan_id, user_id),
				FOREIGN KEY (clan_id) REFERENCES clans(id) ON DELETE CASCADE,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);`,

			`CREATE TABLE IF NOT EXISTS clans_blacklist (
				clan_id INTEGER NOT NULL,
				user_id BIGINT NOT NULL,
				reason TEXT NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (clan_id, user_id),
				FOREIGN KEY (clan_id) REFERENCES clans(id) ON DELETE CASCADE,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);`,
		}
	}

	for _, query := range queries {
		Debug("Выполнение SQL: %s", query)
		if _, err = DB.Exec(query); err != nil {
			log.Fatalf("❌ Ошибка создания таблицы: %v", err)
		}
	}

	log.Println("✅ Подключение к БД успешно")
}

func GetUserBalanceSQL(id uint64, username string) (float64, error) {
	Debug("GetUserBalanceSQL: id=%d", id)
	var balance float64
	err := DB.QueryRow("SELECT balance FROM users WHERE id = ?", id).Scan(&balance)

	if err == sql.ErrNoRows {
		Debug("Пользователь не найден, создаем: %d", id)
		_, err = DB.Exec("INSERT INTO users (id, username, balance) VALUES (?, ?, ?)", id, username, 1000)
		return 1000, err
	}
	return balance, err
}

// Вспомогательные функции...
// func GetUser(id uint64) (int64, error) {
// 	var telegram_id int64
// 	err := DB.QueryRow("SELECT id FROM users WHERE id = ?", id).Scan(&telegram_id)
// 	return telegram_id, err
// }

func UpdateBalanceSQL(id uint64, amount float64) error {
	_, err := DB.Exec("UPDATE users SET balance = balance + ? WHERE id = ?", amount, id)
	return err
}

func IsUserBanned(id int64) (bool, error) {
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM blacklists WHERE user_id = ?)", id).Scan(&exists)
	return exists, err
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func CloseDB() {
	if DB != nil {
		log.Println("💾 Закрытие подключения к БД...")
		DB.Close()
	}
}
