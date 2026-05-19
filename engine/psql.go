package engine

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq" // Драйвер для PostgreSQL
)

var DB *sql.DB

// InitDB инициализирует подключение к PostgreSQL и создает таблицы
func InitDB() {
	// Собираем строку подключения (DSN)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_NAME", "postgres"),
		getEnv("DB_SSLMODE", "disable"),
	)

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Ошибка конфигурации PostgreSQL: %v", err)
	}

	// Проверяем реальное соединение с базой
	err = DB.Ping()
	if err != nil {
		log.Fatalf("Не удалось подключиться к PostgreSQL (Ping): %v", err)
	}

	// 1. Создаем таблицу пользователей
	usersQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id BIGINT PRIMARY KEY,
		username TEXT,
		balance INTEGER DEFAULT 1000
	);`

	_, err = DB.Exec(usersQuery)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы пользователей: %v", err)
	}

	// 2. Создаем таблицу черного списка
	// Используем обратные кавычки для многострочного текста
	blacklistQuery := `
	CREATE TABLE IF NOT EXISTS blacklists (
		id SERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		reason TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT check_reason_not_empty CHECK (length(trim(reason)) > 0),
		CONSTRAINT check_user_id_positive CHECK (user_id > 0),
		CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);`
    
	_, err = DB.Exec(blacklistQuery)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы blacklists: %v", err)
	}
}

// GetUserBalanceSQL проверяет наличие пользователя в БД. Если его нет — создает.
func GetUserBalanceSQL(id int64, username string) (int, error) {
	var balance int
	err := DB.QueryRow("SELECT balance FROM users WHERE id = $1", id).Scan(&balance)

	if err == sql.ErrNoRows {
		initialBalance := 1000
		_, err = DB.Exec("INSERT INTO users (id, username, balance) VALUES ($1, $2, $3)", id, username, initialBalance)
		if err != nil {
			return 0, err
		}
		log.Printf("🆕 Создана новая запись в PostgreSQL для: %s (ID: %d)", username, id)
		return initialBalance, nil
	}

	if err != nil {
		return 0, err
	}

	return balance, nil
}

// UpdateBalanceSQL обновляет баланс пользователя
func UpdateBalanceSQL(id int64, amount int) error {
	_, err := DB.Exec("UPDATE users SET balance = balance + $1 WHERE id = $2", amount, id)
	return err
}

// getEnv вспомогательная функция для чтения переменных окружения
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}



// Blacklist

func IsUserBanned(id int64) (bool, error) {
    var exists bool
    // Используем EXISTS для быстрой проверки наличия записи
    query := "SELECT EXISTS(SELECT 1 FROM blacklists WHERE user_id = $1)"
    
    err := DB.QueryRow(query, id).Scan(&exists)
    if err != nil {
        return false, err
    }
    
    return exists, nil
}