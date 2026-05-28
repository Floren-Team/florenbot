package engine

import (
	"database/sql"
	"fmt"
	"log"
	"os"



	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// Функция для отладки (дебаггер)
func debug(format string, v ...interface{}) {
	if os.Getenv("DEBUG") == "true" {
		log.Printf("[DEBUG] "+format, v...)
	}
}





func InitDB() {
	var err error
	dbMode := getEnv("DB_MODE", "mysql")
	debug("Режим БД: %s", dbMode)

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
			os.Mkdir("db", 0755)
		}
		DB, err = sql.Open("sqlite3", "bot.db")

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

	// SQL-запросы для создания таблиц
	var queries []string
	if dbMode == "mysql" {
		queries = []string{
			// 1. Спочатку створюємо незалежні таблиці
			`CREATE TABLE IF NOT EXISTS clans (
				id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, 
				name VARCHAR(255) NOT NULL, 
				owner_id BIGINT NOT NULL, 
				invite_code VARCHAR(32) NULL UNIQUE,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			) ENGINE=InnoDB;`,

			`CREATE TABLE IF NOT EXISTS promocodes (
				id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, 
				code VARCHAR(255) NOT NULL UNIQUE, 
				amount INT NOT NULL, 
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			) ENGINE=InnoDB;`,

			// 2. Таблиця users з FOREIGN KEY
			`CREATE TABLE IF NOT EXISTS users (
				id BIGINT PRIMARY KEY, 
				username VARCHAR(255), 
				balance FLOAT DEFAULT 1000, 
				promocode VARCHAR(32), 
				clan_id INT DEFAULT NULL, 
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (clan_id) REFERENCES clans(id) ON DELETE SET NULL,
				FOREIGN KEY (promocode) REFERENCES promocodes(code) ON DELETE SET NULL
			) ENGINE=InnoDB;`,

			// 3. Таблиця, що залежить від users
			`CREATE TABLE IF NOT EXISTS blacklists (
				id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, 
				user_id BIGINT NOT NULL, 
				reason TEXT NOT NULL, 
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
				CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			) ENGINE=InnoDB;`,

			`CREATE TABLE IF NOT EXISTS clans_members (
				clan_id INT NOT NULL,
				user_id BIGINT NOT NULL,
				role VARCHAR(20) DEFAULT 'member',
				joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (clan_id, user_id),
				FOREIGN KEY (clan_id) REFERENCES clans(id) ON DELETE CASCADE,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			) ENGINE=InnoDB;`,
		}
	} else {
		// SQLite (порядок створення критично важливий)
		queries = []string{
			`CREATE TABLE IF NOT EXISTS clans (
				id INTEGER PRIMARY KEY AUTOINCREMENT, 
				name VARCHAR(255) NOT NULL, 
				owner_id BIGINT NOT NULL,
				invite_code VARCHAR(32) NULL UNIQUE,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`,

			`CREATE TABLE IF NOT EXISTS promocodes (
				id INTEGER PRIMARY KEY AUTOINCREMENT, 
				code VARCHAR(255) NOT NULL UNIQUE, 
				amount INT NOT NULL, 
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`,

			`CREATE TABLE IF NOT EXISTS users (
				id BIGINT PRIMARY KEY, 
				username VARCHAR(255), 
				balance FLOAT DEFAULT 1000, 
				promocode VARCHAR(32), 
				clan_id INTEGER, 
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY(clan_id) REFERENCES clans(id) ON DELETE SET NULL,
				FOREIGN KEY(promocode) REFERENCES promocodes(code) ON DELETE SET NULL
			);`,

			`CREATE TABLE IF NOT EXISTS blacklists (
				id INTEGER PRIMARY KEY AUTOINCREMENT, 
				user_id BIGINT NOT NULL, 
				reason TEXT NOT NULL, 
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
				FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
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
		}
	}
	for _, query := range queries {
		debug("Выполнение SQL: %s", query)
		if _, err = DB.Exec(query); err != nil {
			log.Fatalf("❌ Ошибка создания таблицы: %v", err)
		}
	}

	log.Println("✅ Подключение к БД успешно")
}




func GetUserBalanceSQL(id int64, username string) (int, error) {
	debug("GetUserBalanceSQL: id=%d", id)
	var balance int
	err := DB.QueryRow("SELECT balance FROM users WHERE id = ?", id).Scan(&balance)

	if err == sql.ErrNoRows {
		debug("Пользователь не найден, создаем: %d", id)
		_, err = DB.Exec("INSERT INTO users (id, username, balance) VALUES (?, ?, ?)", id, username, 1000)
		return 1000, err
	}
	return balance, err
}




// Вспомогательные функции...
func GetUser(id int64) (int64, error) {
	var telegram_id int64
	err := DB.QueryRow("SELECT id FROM users WHERE id = ?", id).Scan(&telegram_id)
	return telegram_id, err
}




func UpdateBalanceSQL(id int64, amount int) error {
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
