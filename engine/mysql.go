package engine

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql" // Драйвер для MySQL
)

var DB *sql.DB

// InitDB инициализирует подключение к MySQL и создает таблицы
func InitDB() {
	// DSN формат: username:password@tcp(host:port)/dbname
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		getEnv("DB_USER", "root"),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "3306"),
		getEnv("DB_NAME", "game_db"),
	)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Ошибка конфигурации MySQL: %v", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("Не удалось подключиться к MySQL (Ping): %v", err)
	}

	// 1. Таблица пользователей
	usersQuery := `
    CREATE TABLE IF NOT EXISTS users (
        id BIGINT PRIMARY KEY,
        username VARCHAR(255),
        balance INTEGER DEFAULT 1000
    ) ENGINE=InnoDB;`

	_, err = DB.Exec(usersQuery)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы пользователей: %v", err)
	}

	// 2. Таблица черного списка
	blacklistQuery := `
    CREATE TABLE IF NOT EXISTS blacklists (
        id INT AUTO_INCREMENT PRIMARY KEY,
        user_id BIGINT NOT NULL,
        reason TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
    ) ENGINE=InnoDB;`

	_, err = DB.Exec(blacklistQuery)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы blacklists: %v", err)
	}
}

// GetUserBalanceSQL проверяет наличие пользователя в БД
func GetUserBalanceSQL(id int64, username string) (int, error) {
	var balance int
	err := DB.QueryRow("SELECT balance FROM users WHERE id = ?", id).Scan(&balance)

	if err == sql.ErrNoRows {
		initialBalance := 1000
		_, err = DB.Exec("INSERT INTO users (id, username, balance) VALUES (?, ?, ?)", id, username, initialBalance)
		if err != nil {
			return 0, err
		}
		log.Printf("🆕 Создана новая запись в MySQL для: %s (ID: %d)", username, id)
		return initialBalance, nil
	}

	if err != nil {
		return 0, err
	}

	return balance, nil
}

// UpdateBalanceSQL обновляет баланс пользователя
func UpdateBalanceSQL(id int64, amount int) error {
	_, err := DB.Exec("UPDATE users SET balance = balance + ? WHERE id = ?", amount, id)
	return err
}

// IsUserBanned проверяет наличие пользователя в черном списке
func IsUserBanned(id int64) (bool, error) {
	var exists bool
	// MySQL использует COUNT или EXISTS
	query := "SELECT EXISTS(SELECT 1 FROM blacklists WHERE user_id = ?)"

	err := DB.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}